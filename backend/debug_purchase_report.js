const db = require('./config/db');

async function analyzePurchaseReport() {
    try {
        console.log('🔍 ANALYZING PURCHASE REPORT DATA ISSUES');
        console.log('=========================================');

        // 1. Check available tables
        console.log('\n1. CHECKING AVAILABLE TABLES:');
        const [journalTables] = await db.execute("SHOW TABLES LIKE '%journal%'");
        console.log('Journal tables:', journalTables.map(r => Object.values(r)[0]));
        
        const [purchaseTables] = await db.execute("SHOW TABLES LIKE '%purchase%'");
        console.log('Purchase tables:', purchaseTables.map(r => Object.values(r)[0]));

        // 2. Check journal_entries table structure and data
        console.log('\n2. JOURNAL ENTRIES ANALYSIS:');
        try {
            const [journalCount] = await db.execute("SELECT COUNT(*) as count FROM journal_entries WHERE transaction_type = 'PURCHASE' AND status = 'POSTED'");
            console.log('Posted purchase journal entries:', journalCount[0].count);
            
            const [allJournalCount] = await db.execute("SELECT COUNT(*) as count FROM journal_entries");
            console.log('Total journal entries:', allJournalCount[0].count);
            
            if (journalCount[0].count > 0) {
                const [sampleJournal] = await db.execute("SELECT * FROM journal_entries WHERE transaction_type = 'PURCHASE' LIMIT 3");
                console.log('Sample purchase journal entries:', sampleJournal.length);
                if (sampleJournal.length > 0) {
                    console.log('First entry:', {
                        id: sampleJournal[0].id,
                        transaction_date: sampleJournal[0].transaction_date,
                        vendor_name: sampleJournal[0].vendor_name,
                        total_amount: sampleJournal[0].total_amount
                    });
                }
            }
        } catch (err) {
            console.log('journal_entries table issue:', err.message);
        }

        // 3. Check purchases table
        console.log('\n3. PURCHASES TABLE ANALYSIS:');
        try {
            const [purchaseCount] = await db.execute("SELECT COUNT(*) as count FROM purchases WHERE deleted_at IS NULL");
            console.log('Active purchases:', purchaseCount[0].count);
            
            if (purchaseCount[0].count > 0) {
                const [samplePurchases] = await db.execute("SELECT * FROM purchases WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT 3");
                console.log('Sample purchases:', samplePurchases.length);
                if (samplePurchases.length > 0) {
                    console.log('First purchase:', {
                        id: samplePurchases[0].id,
                        code: samplePurchases[0].code,
                        purchase_date: samplePurchases[0].purchase_date,
                        vendor_id: samplePurchases[0].vendor_id,
                        total_amount: samplePurchases[0].total_amount,
                        status: samplePurchases[0].status
                    });
                }
            }
        } catch (err) {
            console.log('purchases table issue:', err.message);
        }

        // 4. Check journal_entry_details table
        console.log('\n4. JOURNAL ENTRY DETAILS ANALYSIS:');
        try {
            const [detailsCount] = await db.execute("SELECT COUNT(*) as count FROM journal_entry_details");
            console.log('Total journal entry details:', detailsCount[0].count);
        } catch (err) {
            console.log('journal_entry_details table issue:', err.message);
        }

        // 5. Test the exact query from Purchase Report Controller
        console.log('\n5. TESTING PURCHASE REPORT QUERY:');
        const start_date = '2025-01-01';
        const end_date = '2025-12-31';
        
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
            LIMIT 5
        `;

        try {
            const [queryResult] = await db.execute(purchaseQuery, [start_date, end_date]);
            console.log('Purchase query result count:', queryResult.length);
            if (queryResult.length > 0) {
                console.log('Sample result:', {
                    vendor_name: queryResult[0].vendor_name,
                    total_amount: queryResult[0].total_amount,
                    transaction_date: queryResult[0].transaction_date,
                    account_code: queryResult[0].account_code
                });
            }
        } catch (err) {
            console.log('Purchase query failed:', err.message);
        }

        // 6. Check for missing vendor_name field in journal_entries
        console.log('\n6. CHECKING JOURNAL_ENTRIES SCHEMA:');
        try {
            const [schema] = await db.execute("DESCRIBE journal_entries");
            const hasVendorName = schema.find(col => col.Field === 'vendor_name');
            const hasVendorId = schema.find(col => col.Field === 'vendor_id');
            console.log('Has vendor_name field:', !!hasVendorName);
            console.log('Has vendor_id field:', !!hasVendorId);
            console.log('All columns:', schema.map(col => col.Field));
        } catch (err) {
            console.log('Schema check failed:', err.message);
        }

        // 7. Check date range issue
        console.log('\n7. CHECKING DATE RANGES:');
        try {
            const [dateRange] = await db.execute(`
                SELECT 
                    MIN(purchase_date) as earliest_purchase,
                    MAX(purchase_date) as latest_purchase,
                    COUNT(*) as total_purchases
                FROM purchases 
                WHERE deleted_at IS NULL
            `);
            console.log('Purchase date range:', dateRange[0]);
        } catch (err) {
            console.log('Date range check failed:', err.message);
        }

        console.log('\n🔍 ANALYSIS COMPLETE');
        process.exit(0);

    } catch (error) {
        console.error('❌ Analysis failed:', error.message);
        process.exit(1);
    }
}

analyzePurchaseReport();