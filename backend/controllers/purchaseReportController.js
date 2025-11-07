const db = require('../config/db');

const purchaseReportController = {
  // Generate Purchase Report from SSOT Journal Data
  async generatePurchaseReport(req, res) {
    try {
      const { start_date, end_date, format = 'json' } = req.query;

      if (!start_date || !end_date) {
        return res.status(400).json({
          success: false,
          message: 'Start date and end date are required'
        });
      }

      console.log(`Generating Purchase Report from ${start_date} to ${end_date}`);

      // 1. Get company information
      const companyQuery = `
        SELECT name, address, city, phone, email 
        FROM company_profiles 
        LIMIT 1
      `;
      const [companyResult] = await db.execute(companyQuery);
      const company = companyResult[0] || {
        name: 'PT. Default Company',
        address: 'Jakarta',
        city: 'Jakarta',
        phone: '+62 21-1234567',
        email: 'info@defaultcompany.com'
      };

      // 2. Get purchase transactions from SSOT journal entries
      const purchaseQuery = `
        SELECT 
          je.id as journal_id,
          je.transaction_date,
          je.reference_number,
          je.description,
          je.total_amount,
          jed.account_code,
          jed.account_name,
          jed.debit_amount,
          jed.credit_amount,
          je.vendor_name,
          je.vendor_id,
          je.status
        FROM journal_entries je
        LEFT JOIN journal_entry_details jed ON je.id = jed.journal_entry_id
        WHERE je.transaction_date BETWEEN ? AND ?
          AND (je.transaction_type = 'PURCHASE' OR jed.account_code LIKE '2%' OR je.vendor_name IS NOT NULL)
          AND je.status = 'POSTED'
        ORDER BY je.transaction_date DESC, je.id ASC
      `;

      const [purchaseTransactions] = await db.execute(purchaseQuery, [start_date, end_date]);

      // 3. Get payment transactions from SSOT journal entries
      const paymentQuery = `
        SELECT 
          je.id as journal_id,
          je.transaction_date,
          je.reference_number,
          je.description,
          je.total_amount,
          jed.account_code,
          jed.debit_amount,
          jed.credit_amount,
          je.vendor_name,
          je.vendor_id
        FROM journal_entries je
        LEFT JOIN journal_entry_details jed ON je.id = jed.journal_entry_id
        WHERE je.transaction_date BETWEEN ? AND ?
          AND (je.transaction_type = 'PAYMENT' OR jed.account_code LIKE '1%')
          AND je.vendor_name IS NOT NULL
          AND je.status = 'POSTED'
        ORDER BY je.transaction_date DESC
      `;

      const [paymentTransactions] = await db.execute(paymentQuery, [start_date, end_date]);

      // 4. Process data and calculate metrics
      const vendorMap = new Map();
      let totalPurchases = 0;
      let totalPayments = 0;
      let activeVendors = new Set();

      // Process purchase transactions
      purchaseTransactions.forEach(transaction => {
        if (transaction.vendor_name) {
          const vendorKey = transaction.vendor_name;
          activeVendors.add(vendorKey);

          if (!vendorMap.has(vendorKey)) {
            vendorMap.set(vendorKey, {
              vendor_name: transaction.vendor_name,
              vendor_id: transaction.vendor_id,
              total_purchases: 0,
              total_payments: 0,
              outstanding: 0,
              transaction_count: 0,
              last_purchase_date: null,
              payment_days: [],
              rating: 'Good',
              payment_score: 100
            });
          }

          const vendor = vendorMap.get(vendorKey);
          
          // Count purchases (credit entries to payable accounts or debit to expense/asset accounts)
          if (transaction.account_code && transaction.account_code.startsWith('2')) {
            // Accounts Payable increase
            if (transaction.credit_amount > 0) {
              vendor.total_purchases += transaction.credit_amount;
              totalPurchases += transaction.credit_amount;
              vendor.transaction_count++;
              vendor.last_purchase_date = transaction.transaction_date;
            }
          } else if (transaction.debit_amount > 0 && transaction.account_code && 
                     (transaction.account_code.startsWith('5') || transaction.account_code.startsWith('6'))) {
            // Expense accounts increase
            vendor.total_purchases += transaction.debit_amount;
            totalPurchases += transaction.debit_amount;
            vendor.transaction_count++;
            vendor.last_purchase_date = transaction.transaction_date;
          }
        }
      });

      // Process payment transactions
      paymentTransactions.forEach(transaction => {
        if (transaction.vendor_name) {
          const vendorKey = transaction.vendor_name;
          
          if (vendorMap.has(vendorKey)) {
            const vendor = vendorMap.get(vendorKey);
            
            // Count payments (debit to payable accounts or credit to cash/bank accounts)
            if (transaction.account_code && transaction.account_code.startsWith('2')) {
              // Accounts Payable decrease
              if (transaction.debit_amount > 0) {
                vendor.total_payments += transaction.debit_amount;
                totalPayments += transaction.debit_amount;
              }
            } else if (transaction.credit_amount > 0 && transaction.account_code && 
                       transaction.account_code.startsWith('1')) {
              // Cash/Bank accounts decrease
              vendor.total_payments += transaction.credit_amount;
              totalPayments += transaction.credit_amount;
            }
          }
        }
      });

      // Calculate outstanding amounts and performance metrics
      const vendorsData = Array.from(vendorMap.values()).map(vendor => {
        vendor.outstanding = vendor.total_purchases - vendor.total_payments;
        
        // Calculate average payment days (simplified)
        vendor.average_payment_days = vendor.payment_days.length > 0 
          ? Math.round(vendor.payment_days.reduce((sum, days) => sum + days, 0) / vendor.payment_days.length)
          : 0;
          
        // Calculate payment score based on outstanding ratio
        const paymentRatio = vendor.total_purchases > 0 ? (vendor.total_payments / vendor.total_purchases) : 1;
        vendor.payment_score = Math.min(100, Math.max(0, Math.round(paymentRatio * 100)));
        
        // Determine rating
        if (vendor.payment_score >= 90) {
          vendor.rating = 'Excellent';
        } else if (vendor.payment_score >= 75) {
          vendor.rating = 'Good';
        } else if (vendor.payment_score >= 60) {
          vendor.rating = 'Fair';
        } else {
          vendor.rating = 'Poor';
        }
        
        return vendor;
      });

      // Sort vendors by purchase amount
      vendorsData.sort((a, b) => b.total_purchases - a.total_purchases);

      // Get top vendors by spend
      const topVendorsBySpend = vendorsData.slice(0, 5).map((vendor, index) => ({
        vendor_name: vendor.vendor_name,
        total_amount: vendor.total_purchases,
        percentage: totalPurchases > 0 ? (vendor.total_purchases / totalPurchases) * 100 : 0,
        rank: index + 1
      }));

      // Calculate payment analysis
      const onTimePayments = vendorsData.filter(v => v.payment_score >= 90).length;
      const latePayments = vendorsData.filter(v => v.payment_score >= 60 && v.payment_score < 90).length;
      const overduePayments = vendorsData.filter(v => v.payment_score < 60).length;
      
      const avgPaymentDays = vendorsData.length > 0 
        ? Math.round(vendorsData.reduce((sum, v) => sum + v.average_payment_days, 0) / vendorsData.length)
        : 0;

      const paymentEfficiency = totalPurchases > 0 ? Math.round((totalPayments / totalPurchases) * 100) : 0;

      // Calculate outstanding payables
      const outstandingPayables = totalPurchases - totalPayments;

      // Prepare response data
      const reportData = {
        company: company,
        currency: 'IDR',
        generated_at: new Date().toISOString(),
        period: {
          start_date: start_date,
          end_date: end_date
        },
        
        // Summary metrics
        total_vendors: vendorMap.size,
        active_vendors: activeVendors.size,
        total_purchases: totalPurchases,
        total_payments: totalPayments,
        outstanding_payables: outstandingPayables,
        
        // Vendor performance data
        vendors_by_performance: vendorsData,
        
        // Top vendors
        top_vendors_by_spend: topVendorsBySpend,
        
        // Payment analysis
        payment_analysis: {
          on_time_payments: onTimePayments,
          late_payments: latePayments,
          overdue_payments: overduePayments,
          payment_efficiency: paymentEfficiency,
          average_payment_days: avgPaymentDays
        }
      };

      console.log('Purchase Report generated successfully:', {
        totalVendors: vendorMap.size,
        totalPurchases: totalPurchases,
        totalPayments: totalPayments,
        outstandingPayables: outstandingPayables
      });

      res.json({
        success: true,
        message: 'Purchase Report generated successfully',
        data: reportData
      });

    } catch (error) {
      console.error('Error generating Purchase Report:', error);
      res.status(500).json({
        success: false,
        message: 'Failed to generate Purchase Report',
        error: error.message
      });
    }
  }
};

module.exports = purchaseReportController;