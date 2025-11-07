/**
 * Cash Flow Export Utilities
 * 
 * Provides enhanced CSV and PDF export functionality for Cash Flow reports
 * Supports SSOT cash flow data formats with detailed account information
 */

import jsPDF from 'jspdf';
import 'jspdf-autotable';
import { SSOTCashFlowData } from '../services/ssotCashFlowReportService';

// Extend jsPDF type to include autoTable functionality
declare module 'jspdf' {
  interface jsPDF {
    autoTable: any;
    lastAutoTable: {
      finalY: number;
    };
  }
}

/**
 * Enhanced CSV Export for Cash Flow Statement (already implemented in reports page)
 * This function is provided here for consistency and potential reuse
 */
export function exportCashFlowToCSV(
  data: SSOTCashFlowData,
  options: {
    includeAccountDetails?: boolean;
    companyName?: string;
  } = {}
): string {
  const { includeAccountDetails = true, companyName } = options;
  
  const csvLines: string[] = [];
  
  // Helper function to escape CSV values
  const escapeCSV = (value: string | number | null | undefined): string => {
    if (value === null || value === undefined) return '';
    const str = String(value);
    if (str.includes(',') || str.includes('"') || str.includes('\n')) {
      return `"${str.replace(/"/g, '""')}"`;
    }
    return str;
  };
  
  // Helper function to format currency
  const formatCurrency = (amount: number | null | undefined): string => {
    if (amount === null || amount === undefined || isNaN(Number(amount))) {
      return '0';
    }
    return Number(amount).toLocaleString('id-ID');
  };
  
  // Header
  csvLines.push(escapeCSV(companyName || data.company?.name || 'PT. Sistem Akuntansi'));
  csvLines.push('CASH FLOW STATEMENT');
  csvLines.push(`Period: ${data.start_date} to ${data.end_date}`);
  csvLines.push(`Generated on: ${new Date().toLocaleString('id-ID')}`);
  csvLines.push(''); // Empty line
  
  // Summary
  csvLines.push('CASH POSITION SUMMARY');
  csvLines.push('Metric,Amount');
  csvLines.push(`Net Cash Flow,${formatCurrency(data.net_cash_flow)}`);
  csvLines.push(`Cash at Beginning of Period,${formatCurrency(data.cash_at_beginning)}`);
  csvLines.push(`Cash at End of Period,${formatCurrency(data.cash_at_end)}`);
  csvLines.push(''); // Empty line
  
  // Operating Activities
  if (data.operating_activities) {
    csvLines.push('OPERATING ACTIVITIES BREAKDOWN');
    csvLines.push('Category,Amount,Type');
    csvLines.push(`Net Income,${formatCurrency(data.operating_activities.net_income)},Operating`);
    
    // Adjustments
    if (data.operating_activities.adjustments) {
      csvLines.push('ADJUSTMENTS FOR NON-CASH ITEMS,,');
      if (data.operating_activities.adjustments.items && data.operating_activities.adjustments.items.length > 0) {
        data.operating_activities.adjustments.items.forEach(item => {
          csvLines.push([
            escapeCSV(`${item.account_code} - ${item.account_name}`),
            formatCurrency(item.amount),
            escapeCSV(item.type || 'Adjustment')
          ].join(','));
        });
      } else {
        csvLines.push(`Depreciation,${formatCurrency(data.operating_activities.adjustments.depreciation)},Adjustment`);
        csvLines.push(`Amortization,${formatCurrency(data.operating_activities.adjustments.amortization)},Adjustment`);
        csvLines.push(`Bad Debt Expense,${formatCurrency(data.operating_activities.adjustments.bad_debt_expense)},Adjustment`);
        csvLines.push(`Gain/Loss on Asset Disposal,${formatCurrency(data.operating_activities.adjustments.gain_loss_on_asset_disposal)},Adjustment`);
        csvLines.push(`Other Non-Cash Items,${formatCurrency(data.operating_activities.adjustments.other_non_cash_items)},Adjustment`);
      }
      csvLines.push(`Total Adjustments,${formatCurrency(data.operating_activities.adjustments.total_adjustments)},Total`);
    }
    
    // Working Capital Changes
    if (data.operating_activities.working_capital_changes) {
      csvLines.push('WORKING CAPITAL CHANGES,,');
      if (data.operating_activities.working_capital_changes.items && data.operating_activities.working_capital_changes.items.length > 0) {
        data.operating_activities.working_capital_changes.items.forEach(item => {
          csvLines.push([
            escapeCSV(`${item.account_code} - ${item.account_name}`),
            formatCurrency(item.amount),
            escapeCSV(item.type || 'Working Capital')
          ].join(','));
        });
      } else {
        csvLines.push(`Accounts Receivable Change,${formatCurrency(data.operating_activities.working_capital_changes.accounts_receivable_change)},Working Capital`);
        csvLines.push(`Inventory Change,${formatCurrency(data.operating_activities.working_capital_changes.inventory_change)},Working Capital`);
        csvLines.push(`Prepaid Expenses Change,${formatCurrency(data.operating_activities.working_capital_changes.prepaid_expenses_change)},Working Capital`);
        csvLines.push(`Accounts Payable Change,${formatCurrency(data.operating_activities.working_capital_changes.accounts_payable_change)},Working Capital`);
        csvLines.push(`Accrued Liabilities Change,${formatCurrency(data.operating_activities.working_capital_changes.accrued_liabilities_change)},Working Capital`);
        csvLines.push(`Other Working Capital Change,${formatCurrency(data.operating_activities.working_capital_changes.other_working_capital_change)},Working Capital`);
      }
      csvLines.push(`Total Working Capital Changes,${formatCurrency(data.operating_activities.working_capital_changes.total_working_capital_changes)},Total`);
    }
    
    csvLines.push(`TOTAL OPERATING CASH FLOW,${formatCurrency(data.operating_activities.total_operating_cash_flow)},Total`);
    csvLines.push(''); // Empty line
  }
  
  // Investing Activities
  if (data.investing_activities) {
    csvLines.push('INVESTING ACTIVITIES BREAKDOWN');
    csvLines.push('Item,Amount,Type');
    if (data.investing_activities.items && data.investing_activities.items.length > 0) {
      data.investing_activities.items.forEach(item => {
        csvLines.push([
          escapeCSV(`${item.account_code} - ${item.account_name}`),
          formatCurrency(item.amount),
          escapeCSV(item.type || 'Investing')
        ].join(','));
      });
    } else {
      csvLines.push(`Purchase of Fixed Assets,${formatCurrency(data.investing_activities.purchase_of_fixed_assets)},Investing`);
      csvLines.push(`Sale of Fixed Assets,${formatCurrency(data.investing_activities.sale_of_fixed_assets)},Investing`);
      csvLines.push(`Purchase of Investments,${formatCurrency(data.investing_activities.purchase_of_investments)},Investing`);
      csvLines.push(`Sale of Investments,${formatCurrency(data.investing_activities.sale_of_investments)},Investing`);
      csvLines.push(`Intangible Asset Purchases,${formatCurrency(data.investing_activities.intangible_asset_purchases)},Investing`);
      csvLines.push(`Other Investing Activities,${formatCurrency(data.investing_activities.other_investing_activities)},Investing`);
    }
    csvLines.push(`TOTAL INVESTING CASH FLOW,${formatCurrency(data.investing_activities.total_investing_cash_flow)},Total`);
    csvLines.push(''); // Empty line
  }
  
  // Financing Activities
  if (data.financing_activities) {
    csvLines.push('FINANCING ACTIVITIES BREAKDOWN');
    csvLines.push('Item,Amount,Type');
    if (data.financing_activities.items && data.financing_activities.items.length > 0) {
      data.financing_activities.items.forEach(item => {
        csvLines.push([
          escapeCSV(`${item.account_code} - ${item.account_name}`),
          formatCurrency(item.amount),
          escapeCSV(item.type || 'Financing')
        ].join(','));
      });
    } else {
      csvLines.push(`Share Capital Increase,${formatCurrency(data.financing_activities.share_capital_increase)},Financing`);
      csvLines.push(`Share Capital Decrease,${formatCurrency(data.financing_activities.share_capital_decrease)},Financing`);
      csvLines.push(`Long Term Debt Increase,${formatCurrency(data.financing_activities.long_term_debt_increase)},Financing`);
      csvLines.push(`Long Term Debt Decrease,${formatCurrency(data.financing_activities.long_term_debt_decrease)},Financing`);
      csvLines.push(`Short Term Debt Increase,${formatCurrency(data.financing_activities.short_term_debt_increase)},Financing`);
      csvLines.push(`Short Term Debt Decrease,${formatCurrency(data.financing_activities.short_term_debt_decrease)},Financing`);
      csvLines.push(`Dividends Paid,${formatCurrency(data.financing_activities.dividends_paid)},Financing`);
      csvLines.push(`Other Financing Activities,${formatCurrency(data.financing_activities.other_financing_activities)},Financing`);
    }
    csvLines.push(`TOTAL FINANCING CASH FLOW,${formatCurrency(data.financing_activities.total_financing_cash_flow)},Total`);
    csvLines.push(''); // Empty line
  }
  
  // Detailed Account Breakdown
  if (data.account_details && data.account_details.length > 0) {
    csvLines.push('DETAILED ACCOUNT BREAKDOWN');
    csvLines.push('Account Code,Account Name,Account Type,Debits,Credits,Net Change');
    data.account_details.forEach(account => {
      csvLines.push([
        escapeCSV(account.account_code),
        escapeCSV(account.account_name),
        escapeCSV(account.account_type),
        formatCurrency(account.debit_total),
        formatCurrency(account.credit_total),
        formatCurrency(account.net_balance)
      ].join(','));
    });
    csvLines.push(''); // Empty line
  }
  
  // Ratios
  if (data.cash_flow_ratios) {
    csvLines.push('CASH FLOW RATIOS');
    csvLines.push('Ratio,Value');
    csvLines.push(`Operating Cash Flow Ratio,${data.cash_flow_ratios.operating_cash_flow_ratio}`);
    csvLines.push(`Cash Flow to Debt Ratio,${formatCurrency(data.cash_flow_ratios.cash_flow_to_debt_ratio)}`);
    csvLines.push(`Free Cash Flow,${formatCurrency(data.cash_flow_ratios.free_cash_flow)}`);
    if (data.cash_flow_ratios.cash_flow_per_share !== undefined) {
      csvLines.push(`Cash Flow per Share,${data.cash_flow_ratios.cash_flow_per_share}`);
    }
    csvLines.push(''); // Empty line
  }
  
  // Footer
  csvLines.push('Generated by Sistem Akuntansi');
  csvLines.push(`Report Date: ${new Date().toLocaleString('id-ID')}`);
  csvLines.push(`Data Source: ${data.enhanced ? 'SSOT Enhanced' : 'SSOT Standard'}`);
  
  return csvLines.join('\n');
}

/**
 * Enhanced PDF Export for Cash Flow Statement
 * Generates professional PDF report with tables for all sections
 */
export function exportCashFlowToPDF(
  data: SSOTCashFlowData,
  options: {
    companyName?: string;
    logoUrl?: string;
    includeAccountDetails?: boolean;
  } = {}
): jsPDF {
  const { companyName, includeAccountDetails = true } = options;
  
  // Initialize PDF
  const doc = new jsPDF('portrait', 'mm', 'a4');
  const pageWidth = doc.internal.pageSize.width;
  const pageHeight = doc.internal.pageSize.height;
  
  // Helper function to format currency for PDF
  const formatCurrencyForPDF = (amount: number | null | undefined): string => {
    if (amount === null || amount === undefined || isNaN(Number(amount))) {
      return 'Rp 0';
    }
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0
    }).format(Number(amount));
  };
  
  // Set fonts
  doc.setFont('helvetica');
  
  // Header section
  let yPosition = 20;
  
  // Company name
  doc.setFontSize(16);
  doc.setFont('helvetica', 'bold');
  const company = companyName || data.company?.name || 'PT. Sistem Akuntansi';
  doc.text(company, pageWidth / 2, yPosition, { align: 'center' });
  
  yPosition += 10;
  
  // Report title
  doc.setFontSize(14);
  doc.text('CASH FLOW STATEMENT', pageWidth / 2, yPosition, { align: 'center' });
  
  yPosition += 8;
  
  // Date range
  doc.setFontSize(12);
  doc.setFont('helvetica', 'normal');
  doc.text(`Period: ${new Date(data.start_date).toLocaleDateString('id-ID')} to ${new Date(data.end_date).toLocaleDateString('id-ID')}`, pageWidth / 2, yPosition, { align: 'center' });
  
  yPosition += 15;
  
  // Summary table
  doc.setFontSize(12);
  doc.setFont('helvetica', 'bold');
  doc.text('CASH POSITION SUMMARY', 20, yPosition);
  yPosition += 5;
  
  const summaryData = [
    ['Net Cash Flow', formatCurrencyForPDF(data.net_cash_flow)],
    ['Cash at Beginning of Period', formatCurrencyForPDF(data.cash_at_beginning)],
    ['Cash at End of Period', formatCurrencyForPDF(data.cash_at_end)]
  ];
  
  // Ensure plugin is properly initialized for Next.js compatibility
  if (typeof (doc as any).autoTable !== 'function') {
    console.error('autoTable function is not available on jsPDF instance');
    // Fallback: skip table and continue with text-only format
    yPosition += 20; // Add some space for the skipped table
  } else {
    (doc as any).autoTable({
      startY: yPosition,
      head: [['Metric', 'Amount']],
      body: summaryData,
      theme: 'grid',
      headStyles: { fillColor: [66, 139, 202], textColor: 255, fontStyle: 'bold' },
      bodyStyles: { fontSize: 11 },
      alternateRowStyles: { fillColor: [245, 245, 245] },
      margin: { left: 20, right: 20 },
      columnStyles: {
        1: { halign: 'right' }
      }
    });
    
    yPosition = (doc as any).lastAutoTable.finalY + 10;
  }
  
  // Operating Activities Section
  if (data.operating_activities) {
    doc.setFontSize(12);
    doc.setFont('helvetica', 'bold');
    doc.text('OPERATING ACTIVITIES', 20, yPosition);
    yPosition += 5;
    
    // Check if we need a new page
    if (yPosition > pageHeight - 100) {
      doc.addPage();
      yPosition = 20;
    }
    
    // Net Income
    const operatingSummaryData = [
      ['Net Income', formatCurrencyForPDF(data.operating_activities.net_income)]
    ];
    
    if (typeof (doc as any).autoTable === 'function') {
      (doc as any).autoTable({
        startY: yPosition,
        head: [['Description', 'Amount']],
        body: operatingSummaryData,
        theme: 'grid',
        headStyles: { fillColor: [40, 167, 69], textColor: 255, fontStyle: 'bold' },
        bodyStyles: { fontSize: 10 },
        margin: { left: 20, right: 20 },
        columnStyles: {
          1: { halign: 'right' }
        }
      });
      
      yPosition = (doc as any).lastAutoTable.finalY + 5;
    } else {
      yPosition += 15; // Add space if table can't be rendered
    }
    
    // Adjustments for Non-Cash Items
    if (data.operating_activities.adjustments) {
      // Prepare adjustment data
      const adjustmentItems: any[] = [];
      
      if (data.operating_activities.adjustments.items && data.operating_activities.adjustments.items.length > 0) {
        data.operating_activities.adjustments.items.forEach(item => {
          adjustmentItems.push([
            item.account_code || '',
            item.account_name || '',
            formatCurrencyForPDF(item.amount)
          ]);
        });
      } else {
        // Use standard adjustment items with more structured data if available
        if (data.operating_activities.adjustments.depreciation) {
          adjustmentItems.push(['', 'Depreciation', formatCurrencyForPDF(data.operating_activities.adjustments.depreciation)]);
        }
        if (data.operating_activities.adjustments.amortization) {
          adjustmentItems.push(['', 'Amortization', formatCurrencyForPDF(data.operating_activities.adjustments.amortization)]);
        }
        if (data.operating_activities.adjustments.bad_debt_expense) {
          adjustmentItems.push(['', 'Bad Debt Expense', formatCurrencyForPDF(data.operating_activities.adjustments.bad_debt_expense)]);
        }
        if (data.operating_activities.adjustments.gain_loss_on_asset_disposal) {
          adjustmentItems.push(['', 'Gain/Loss on Asset Disposal', formatCurrencyForPDF(data.operating_activities.adjustments.gain_loss_on_asset_disposal)]);
        }
        if (data.operating_activities.adjustments.other_non_cash_items) {
          adjustmentItems.push(['', 'Other Non-Cash Items', formatCurrencyForPDF(data.operating_activities.adjustments.other_non_cash_items)]);
        }
      }
      
      if (adjustmentItems.length > 0) {
        if (typeof (doc as any).autoTable === 'function') {
          (doc as any).autoTable({
            startY: yPosition,
            head: [['Account Code', 'Adjustment Item', 'Amount']],
            body: adjustmentItems,
            theme: 'grid',
            headStyles: { fillColor: [52, 58, 64], textColor: 255, fontStyle: 'bold' },
            bodyStyles: { fontSize: 9 },
            alternateRowStyles: { fillColor: [248, 249, 250] },
            margin: { left: 25, right: 20 },
            columnStyles: {
              2: { halign: 'right' }
            }
          });
          
          yPosition = (doc as any).lastAutoTable.finalY + 2;
        } else {
          yPosition += 10; // Add space if table can't be rendered
        }
        
        // Total adjustments
        if (typeof (doc as any).autoTable === 'function') {
          (doc as any).autoTable({
            startY: yPosition,
            head: [['', 'Total Adjustments', '']],
            body: [['', formatCurrencyForPDF(data.operating_activities.adjustments.total_adjustments), '']],
            theme: 'grid',
            headStyles: { fillColor: [52, 58, 64], textColor: 255, fontStyle: 'bold' },
            bodyStyles: { fontSize: 10, fontStyle: 'bold' },
            margin: { left: 25, right: 20 },
            columnStyles: {
              1: { halign: 'right', fillColor: [248, 249, 250] }
            }
          });
          
          yPosition = (doc as any).lastAutoTable.finalY + 2;
        } else {
          yPosition += 8; // Add space if table can't be rendered
        }
      }
    }
    
    // Working Capital Changes
    if (data.operating_activities.working_capital_changes) {
      // Prepare working capital change data
      const workingCapitalItems: any[] = [];
      
      if (data.operating_activities.working_capital_changes.items && data.operating_activities.working_capital_changes.items.length > 0) {
        data.operating_activities.working_capital_changes.items.forEach(item => {
          workingCapitalItems.push([
            item.account_code || '',
            item.account_name || '',
            formatCurrencyForPDF(item.amount)
          ]);
        });
      } else {
        // Use standard working capital items with more structured data if available
        if (data.operating_activities.working_capital_changes.accounts_receivable_change) {
          workingCapitalItems.push(['', 'Accounts Receivable Change', formatCurrencyForPDF(data.operating_activities.working_capital_changes.accounts_receivable_change)]);
        }
        if (data.operating_activities.working_capital_changes.inventory_change) {
          workingCapitalItems.push(['', 'Inventory Change', formatCurrencyForPDF(data.operating_activities.working_capital_changes.inventory_change)]);
        }
        if (data.operating_activities.working_capital_changes.prepaid_expenses_change) {
          workingCapitalItems.push(['', 'Prepaid Expenses Change', formatCurrencyForPDF(data.operating_activities.working_capital_changes.prepaid_expenses_change)]);
        }
        if (data.operating_activities.working_capital_changes.accounts_payable_change) {
          workingCapitalItems.push(['', 'Accounts Payable Change', formatCurrencyForPDF(data.operating_activities.working_capital_changes.accounts_payable_change)]);
        }
        if (data.operating_activities.working_capital_changes.accrued_liabilities_change) {
          workingCapitalItems.push(['', 'Accrued Liabilities Change', formatCurrencyForPDF(data.operating_activities.working_capital_changes.accrued_liabilities_change)]);
        }
        if (data.operating_activities.working_capital_changes.other_working_capital_change) {
          workingCapitalItems.push(['', 'Other Working Capital Change', formatCurrencyForPDF(data.operating_activities.working_capital_changes.other_working_capital_change)]);
        }
      }
      
      if (workingCapitalItems.length > 0) {
        if (typeof (doc as any).autoTable === 'function') {
          (doc as any).autoTable({
            startY: yPosition,
            head: [['Account Code', 'Working Capital Change', 'Amount']],
            body: workingCapitalItems,
            theme: 'grid',
            headStyles: { fillColor: [52, 58, 64], textColor: 255, fontStyle: 'bold' },
            bodyStyles: { fontSize: 9 },
            alternateRowStyles: { fillColor: [248, 249, 250] },
            margin: { left: 25, right: 20 },
            columnStyles: {
              2: { halign: 'right' }
            }
          });
          
          yPosition = (doc as any).lastAutoTable.finalY + 2;
        } else {
          yPosition += 10; // Add space if table can't be rendered
        }
        
        // Total working capital changes
        if (typeof (doc as any).autoTable === 'function') {
          (doc as any).autoTable({
            startY: yPosition,
            head: [['', 'Total Working Capital Changes', '']],
            body: [['', formatCurrencyForPDF(data.operating_activities.working_capital_changes.total_working_capital_changes), '']],
            theme: 'grid',
            headStyles: { fillColor: [52, 58, 64], textColor: 255, fontStyle: 'bold' },
            bodyStyles: { fontSize: 10, fontStyle: 'bold' },
            margin: { left: 25, right: 20 },
            columnStyles: {
              1: { halign: 'right', fillColor: [248, 249, 250] }
            }
          });
          
          yPosition = (doc as any).lastAutoTable.finalY + 2;
        } else {
          yPosition += 8; // Add space if table can't be rendered
        }
      }
    }
    
    // Total Operating Cash Flow
    if (typeof (doc as any).autoTable === 'function') {
      (doc as any).autoTable({
        startY: yPosition,
        head: [['TOTAL OPERATING CASH FLOW', '']],
        body: [[formatCurrencyForPDF(data.operating_activities.total_operating_cash_flow), '']],
        theme: 'grid',
        headStyles: { fillColor: [40, 167, 69], textColor: 255, fontStyle: 'bold' },
        bodyStyles: { fontSize: 11, fontStyle: 'bold' },
        margin: { left: 20, right: 20 },
        columnStyles: {
          0: { halign: 'right', fillColor: [240, 255, 240] }
        }
      });
      
      yPosition = (doc as any).lastAutoTable.finalY + 10;
    } else {
      yPosition += 15; // Add space if table can't be rendered
    }
  }
  
  // Investing Activities Section
  if (data.investing_activities) {
    doc.setFontSize(12);
    doc.setFont('helvetica', 'bold');
    doc.text('INVESTING ACTIVITIES', 20, yPosition);
    yPosition += 5;
    
    // Check if we need a new page
    if (yPosition > pageHeight - 100) {
      doc.addPage();
      yPosition = 20;
    }
    
    // Prepare investing activities data
    const investingItems: any[] = [];
    
    if (data.investing_activities.items && data.investing_activities.items.length > 0) {
      data.investing_activities.items.forEach(item => {
        investingItems.push([
          `${item.account_code} - ${item.account_name}`,
          formatCurrencyForPDF(item.amount)
        ]);
      });
    } else {
      // Use standard investing activity items
      if (data.investing_activities.purchase_of_fixed_assets) {
        investingItems.push(['Purchase of Fixed Assets', formatCurrencyForPDF(data.investing_activities.purchase_of_fixed_assets)]);
      }
      if (data.investing_activities.sale_of_fixed_assets) {
        investingItems.push(['Sale of Fixed Assets', formatCurrencyForPDF(data.investing_activities.sale_of_fixed_assets)]);
      }
      if (data.investing_activities.purchase_of_investments) {
        investingItems.push(['Purchase of Investments', formatCurrencyForPDF(data.investing_activities.purchase_of_investments)]);
      }
      if (data.investing_activities.sale_of_investments) {
        investingItems.push(['Sale of Investments', formatCurrencyForPDF(data.investing_activities.sale_of_investments)]);
      }
      if (data.investing_activities.intangible_asset_purchases) {
        investingItems.push(['Intangible Asset Purchases', formatCurrencyForPDF(data.investing_activities.intangible_asset_purchases)]);
      }
      if (data.investing_activities.other_investing_activities) {
        investingItems.push(['Other Investing Activities', formatCurrencyForPDF(data.investing_activities.other_investing_activities)]);
      }
    }
    
    if (investingItems.length > 0) {
      if (typeof (doc as any).autoTable === 'function') {
        (doc as any).autoTable({
          startY: yPosition,
          head: [['Investment Activity', 'Amount']],
          body: investingItems,
          theme: 'grid',
          headStyles: { fillColor: [138, 43, 226], textColor: 255, fontStyle: 'bold' },
          bodyStyles: { fontSize: 10 },
          alternateRowStyles: { fillColor: [248, 249, 250] },
          margin: { left: 20, right: 20 },
          columnStyles: {
            1: { halign: 'right' }
          }
        });
        
        yPosition = (doc as any).lastAutoTable.finalY + 2;
      } else {
        yPosition += 10; // Add space if table can't be rendered
      }
    }
    
    // Total Investing Cash Flow
    if (typeof (doc as any).autoTable === 'function') {
      (doc as any).autoTable({
        startY: yPosition,
        head: [['TOTAL INVESTING CASH FLOW', '']],
        body: [[formatCurrencyForPDF(data.investing_activities.total_investing_cash_flow), '']],
        theme: 'grid',
        headStyles: { fillColor: [138, 43, 226], textColor: 255, fontStyle: 'bold' },
        bodyStyles: { fontSize: 11, fontStyle: 'bold' },
        margin: { left: 20, right: 20 },
        columnStyles: {
          0: { halign: 'right', fillColor: [240, 240, 255] }
        }
      });
      
      yPosition = (doc as any).lastAutoTable.finalY + 10;
    } else {
      yPosition += 15; // Add space if table can't be rendered
    }
  }
  
  // Financing Activities Section
  if (data.financing_activities) {
    doc.setFontSize(12);
    doc.setFont('helvetica', 'bold');
    doc.text('FINANCING ACTIVITIES', 20, yPosition);
    yPosition += 5;
    
    // Check if we need a new page
    if (yPosition > pageHeight - 100) {
      doc.addPage();
      yPosition = 20;
    }
    
    // Prepare financing activities data
    const financingItems: any[] = [];
    
    if (data.financing_activities.items && data.financing_activities.items.length > 0) {
      data.financing_activities.items.forEach(item => {
        financingItems.push([
          `${item.account_code} - ${item.account_name}`,
          formatCurrencyForPDF(item.amount)
        ]);
      });
    } else {
      // Use standard financing activity items
      if (data.financing_activities.share_capital_increase) {
        financingItems.push(['Share Capital Increase', formatCurrencyForPDF(data.financing_activities.share_capital_increase)]);
      }
      if (data.financing_activities.share_capital_decrease) {
        financingItems.push(['Share Capital Decrease', formatCurrencyForPDF(data.financing_activities.share_capital_decrease)]);
      }
      if (data.financing_activities.long_term_debt_increase) {
        financingItems.push(['Long Term Debt Increase', formatCurrencyForPDF(data.financing_activities.long_term_debt_increase)]);
      }
      if (data.financing_activities.long_term_debt_decrease) {
        financingItems.push(['Long Term Debt Decrease', formatCurrencyForPDF(data.financing_activities.long_term_debt_decrease)]);
      }
      if (data.financing_activities.short_term_debt_increase) {
        financingItems.push(['Short Term Debt Increase', formatCurrencyForPDF(data.financing_activities.short_term_debt_increase)]);
      }
      if (data.financing_activities.short_term_debt_decrease) {
        financingItems.push(['Short Term Debt Decrease', formatCurrencyForPDF(data.financing_activities.short_term_debt_decrease)]);
      }
      if (data.financing_activities.dividends_paid) {
        financingItems.push(['Dividends Paid', formatCurrencyForPDF(data.financing_activities.dividends_paid)]);
      }
      if (data.financing_activities.other_financing_activities) {
        financingItems.push(['Other Financing Activities', formatCurrencyForPDF(data.financing_activities.other_financing_activities)]);
      }
    }
    
    if (financingItems.length > 0) {
      if (typeof (doc as any).autoTable === 'function') {
        (doc as any).autoTable({
          startY: yPosition,
          head: [['Financing Activity', 'Amount']],
          body: financingItems,
          theme: 'grid',
          headStyles: { fillColor: [255, 193, 7], textColor: [33, 37, 41], fontStyle: 'bold' },
          bodyStyles: { fontSize: 10 },
          alternateRowStyles: { fillColor: [248, 249, 250] },
          margin: { left: 20, right: 20 },
          columnStyles: {
            1: { halign: 'right' }
          }
        });
        
        yPosition = (doc as any).lastAutoTable.finalY + 2;
      } else {
        yPosition += 10; // Add space if table can't be rendered
      }
    }
    
    // Total Financing Cash Flow
    if (typeof (doc as any).autoTable === 'function') {
      (doc as any).autoTable({
        startY: yPosition,
        head: [['TOTAL FINANCING CASH FLOW', '']],
        body: [[formatCurrencyForPDF(data.financing_activities.total_financing_cash_flow), '']],
        theme: 'grid',
        headStyles: { fillColor: [255, 193, 7], textColor: [33, 37, 41], fontStyle: 'bold' },
        bodyStyles: { fontSize: 11, fontStyle: 'bold' },
        margin: { left: 20, right: 20 },
        columnStyles: {
          0: { halign: 'right', fillColor: [255, 255, 200] }
        }
      });
      
      yPosition = (doc as any).lastAutoTable.finalY + 10;
    } else {
      yPosition += 15; // Add space if table can't be rendered
    }
  }
  
  // Detailed Account Breakdown by Activity Type (if included)
  if (includeAccountDetails) {
    // Operating Activities Detail
    if (data.operating_activities && data.operating_activities.items && data.operating_activities.items.length > 0) {
      doc.setFontSize(12);
      doc.setFont('helvetica', 'bold');
      doc.text('OPERATING ACTIVITIES - DETAIL', 20, yPosition);
      yPosition += 5;
      
      // Check if we need a new page
      if (yPosition > pageHeight - 100) {
        doc.addPage();
        yPosition = 20;
      }
      
      // Prepare operating activities data
      const operatingData: any[] = [];
      data.operating_activities.items.forEach(item => {
        operatingData.push([
          item.account_code || '',
          item.account_name || '',
          formatCurrencyForPDF(item.amount)
        ]);
      });
      
      if (typeof (doc as any).autoTable === 'function') {
        (doc as any).autoTable({
          startY: yPosition,
          head: [['Account Code', 'Account Name', 'Amount']],
          body: operatingData,
          theme: 'grid',
          headStyles: { fillColor: [40, 167, 69], textColor: 255, fontStyle: 'bold' },
          bodyStyles: { fontSize: 9 },
          alternateRowStyles: { fillColor: [248, 249, 250] },
          margin: { left: 20, right: 20 },
          columnStyles: {
            2: { halign: 'right' }
          }
        });
        
        yPosition = (doc as any).lastAutoTable.finalY + 10;
      } else {
        yPosition += 15; // Add space if table can't be rendered
      }
    }
    
    // Investing Activities Detail
    if (data.investing_activities && data.investing_activities.items && data.investing_activities.items.length > 0) {
      doc.setFontSize(12);
      doc.setFont('helvetica', 'bold');
      doc.text('INVESTING ACTIVITIES - DETAIL', 20, yPosition);
      yPosition += 5;
      
      // Check if we need a new page
      if (yPosition > pageHeight - 100) {
        doc.addPage();
        yPosition = 20;
      }
      
      // Prepare investing activities data
      const investingData: any[] = [];
      data.investing_activities.items.forEach(item => {
        investingData.push([
          item.account_code || '',
          item.account_name || '',
          formatCurrencyForPDF(item.amount)
        ]);
      });
      
      if (typeof (doc as any).autoTable === 'function') {
        (doc as any).autoTable({
          startY: yPosition,
          head: [['Account Code', 'Account Name', 'Amount']],
          body: investingData,
          theme: 'grid',
          headStyles: { fillColor: [138, 43, 226], textColor: 255, fontStyle: 'bold' },
          bodyStyles: { fontSize: 9 },
          alternateRowStyles: { fillColor: [248, 249, 250] },
          margin: { left: 20, right: 20 },
          columnStyles: {
            2: { halign: 'right' }
          }
        });
        
        yPosition = (doc as any).lastAutoTable.finalY + 10;
      } else {
        yPosition += 15; // Add space if table can't be rendered
      }
    }
    
    // Financing Activities Detail
    if (data.financing_activities && data.financing_activities.items && data.financing_activities.items.length > 0) {
      doc.setFontSize(12);
      doc.setFont('helvetica', 'bold');
      doc.text('FINANCING ACTIVITIES - DETAIL', 20, yPosition);
      yPosition += 5;
      
      // Check if we need a new page
      if (yPosition > pageHeight - 100) {
        doc.addPage();
        yPosition = 20;
      }
      
      // Prepare financing activities data
      const financingData: any[] = [];
      data.financing_activities.items.forEach(item => {
        financingData.push([
          item.account_code || '',
          item.account_name || '',
          formatCurrencyForPDF(item.amount)
        ]);
      });
      
      if (typeof (doc as any).autoTable === 'function') {
        (doc as any).autoTable({
          startY: yPosition,
          head: [['Account Code', 'Account Name', 'Amount']],
          body: financingData,
          theme: 'grid',
          headStyles: { fillColor: [255, 193, 7], textColor: [33, 37, 41], fontStyle: 'bold' },
          bodyStyles: { fontSize: 9 },
          alternateRowStyles: { fillColor: [248, 249, 250] },
          margin: { left: 20, right: 20 },
          columnStyles: {
            2: { halign: 'right' }
          }
        });
        
        yPosition = (doc as any).lastAutoTable.finalY + 10;
      } else {
        yPosition += 15; // Add space if table can't be rendered
      }
    }
    
    // Overall Account Details (if available)
    if (data.account_details && data.account_details.length > 0) {
      doc.setFontSize(12);
      doc.setFont('helvetica', 'bold');
      doc.text('ALL ACCOUNTS SUMMARY', 20, yPosition);
      yPosition += 5;
      
      // Check if we need a new page
      if (yPosition > pageHeight - 100) {
        doc.addPage();
        yPosition = 20;
      }
      
      // Prepare account details data
      const accountDetailsData: any[] = [];
      data.account_details.forEach(account => {
        accountDetailsData.push([
          account.account_code || '',
          account.account_name || '',
          account.account_type || '',
          formatCurrencyForPDF(account.debit_total),
          formatCurrencyForPDF(account.credit_total),
          formatCurrencyForPDF(account.net_balance)
        ]);
      });
      
      if (typeof (doc as any).autoTable === 'function') {
        (doc as any).autoTable({
          startY: yPosition,
          head: [['Code', 'Account Name', 'Type', 'Debits', 'Credits', 'Net Change']],
          body: accountDetailsData,
          theme: 'grid',
          headStyles: { fillColor: [108, 117, 125], textColor: 255, fontStyle: 'bold' },
          bodyStyles: { fontSize: 8 },
          alternateRowStyles: { fillColor: [248, 249, 250] },
          margin: { left: 20, right: 20 },
          columnStyles: {
            3: { halign: 'right' },
            4: { halign: 'right' },
            5: { halign: 'right' }
          }
        });
        
        yPosition = (doc as any).lastAutoTable.finalY + 10;
      } else {
        yPosition += 15; // Add space if table can't be rendered
      }
    }
  }
  
  // Cash Flow Ratios (if available)
  if (data.cash_flow_ratios) {
    doc.setFontSize(12);
    doc.setFont('helvetica', 'bold');
    doc.text('CASH FLOW RATIOS', 20, yPosition);
    yPosition += 5;
    
    // Check if we need a new page
    if (yPosition > pageHeight - 100) {
      doc.addPage();
      yPosition = 20;
    }
    
    const ratiosData: any[] = [];
    if (data.cash_flow_ratios.operating_cash_flow_ratio !== undefined) {
      ratiosData.push(['Operating Cash Flow Ratio', data.cash_flow_ratios.operating_cash_flow_ratio.toString()]);
    }
    if (data.cash_flow_ratios.cash_flow_to_debt_ratio !== undefined) {
      ratiosData.push(['Cash Flow to Debt Ratio', formatCurrencyForPDF(data.cash_flow_ratios.cash_flow_to_debt_ratio)]);
    }
    if (data.cash_flow_ratios.free_cash_flow !== undefined) {
      ratiosData.push(['Free Cash Flow', formatCurrencyForPDF(data.cash_flow_ratios.free_cash_flow)]);
    }
    if (data.cash_flow_ratios.cash_flow_per_share !== undefined) {
      ratiosData.push(['Cash Flow per Share', data.cash_flow_ratios.cash_flow_per_share.toString()]);
    }
    
    if (ratiosData.length > 0) {
      if (typeof (doc as any).autoTable === 'function') {
        (doc as any).autoTable({
          startY: yPosition,
          head: [['Ratio', 'Value']],
          body: ratiosData,
          theme: 'grid',
          headStyles: { fillColor: [52, 58, 64], textColor: 255, fontStyle: 'bold' },
          bodyStyles: { fontSize: 10 },
          alternateRowStyles: { fillColor: [248, 249, 250] },
          margin: { left: 20, right: 20 },
          columnStyles: {
            1: { halign: 'right' }
          }
        });
        
        yPosition = (doc as any).lastAutoTable.finalY + 10;
      } else {
        yPosition += 15; // Add space if table can't be rendered
      }
    }
  }
  
  // Footer
  const footerY = pageHeight - 20;
  doc.setFontSize(8);
  doc.setFont('helvetica', 'normal');
  doc.text('Generated by Sistem Akuntansi', 20, footerY);
  doc.text(`Generated on: ${new Date().toLocaleString('id-ID')}`, pageWidth - 20, footerY, { align: 'right' });
  doc.text(`Data Source: ${data.enhanced ? 'SSOT Enhanced' : 'SSOT Standard'}`, pageWidth / 2, footerY, { align: 'center' });
  
  // Add page numbers if multiple pages
  const pageCount = doc.getNumberOfPages();
  if (pageCount > 1) {
    for (let i = 1; i <= pageCount; i++) {
      doc.setPage(i);
      doc.setFontSize(8);
      doc.text(`Page ${i} of ${pageCount}`, pageWidth - 20, footerY - 5, { align: 'right' });
    }
  }
  
  return doc;
}

/**
 * Download PDF file
 */
export function downloadCashFlowPDF(doc: jsPDF, filename?: string): void {
  const pdfName = filename || `cash_flow_${new Date().toISOString().split('T')[0]}.pdf`;
  doc.save(pdfName);
}

/**
 * Export Cash Flow as PDF and trigger download
 */
export function exportAndDownloadCashFlowPDF(
  data: SSOTCashFlowData,
  options?: {
    companyName?: string;
    logoUrl?: string;
    includeAccountDetails?: boolean;
    filename?: string;
  }
): void {
  const doc = exportCashFlowToPDF(data, options);
  downloadCashFlowPDF(doc, options?.filename);
}