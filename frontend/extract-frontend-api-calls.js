// Script untuk menganalisis semua API calls di frontend services
const fs = require('fs');
const path = require('path');

console.log('=== FRONTEND API CALLS ANALYSIS ===\n');

// API calls berdasarkan analisis dari service files
const frontendApiCalls = {
  // salesService.ts - Sudah diupdate
  salesService: [
    'GET /api/v1/sales',
    'GET /api/v1/sales/:id',  
    'POST /api/v1/sales',
    'PUT /api/v1/sales/:id',
    'DELETE /api/v1/sales/:id',
    'POST /api/v1/sales/:id/confirm',
    'POST /api/v1/sales/:id/invoice',
    'POST /api/v1/sales/:id/cancel',
    'GET /api/v1/sales/:id/payments',
    'POST /api/v1/sales/:id/payments',
    'POST /api/v1/sales/:id/integrated-payment',
    'POST /api/v1/sales/:id/returns',
    'GET /api/v1/sales/returns',
    'GET /api/v1/sales/summary',
    'GET /api/v1/sales/analytics',
    'GET /api/v1/sales/receivables',
    'GET /api/v1/sales/customer/:customer_id',
    'GET /api/v1/sales/customer/:customer_id/invoices',
    'GET /api/v1/sales/:id/invoice/pdf',
    'GET /api/v1/sales/report/pdf'
  ],

  // purchaseService.ts - Perlu diupdate
  purchaseService: [
    'GET /purchases', // ❌ Missing /api/v1
    'POST /purchases', // ❌ Missing /api/v1  
    'GET /purchases/:id', // ❌ Missing /api/v1
    'PUT /purchases/:id', // ❌ Missing /api/v1
    'DELETE /purchases/:id', // ❌ Missing /api/v1
    'POST /purchases/:id/submit-approval', // ❌ Missing /api/v1
    'GET /purchases/pending-approvals', // ❌ Missing /api/v1
    'POST /purchases/:id/approve', // ❌ Missing /api/v1
    'POST /purchases/:id/reject' // ❌ Missing /api/v1
  ],

  // userService.ts - Perlu diupdate
  userService: [
    'GET /users', // ❌ Missing /api/v1
    'GET /users/:id', // ❌ Missing /api/v1
    'POST /users', // ❌ Missing /api/v1
    'PUT /users/:id', // ❌ Missing /api/v1
    'DELETE /users/:id', // ❌ Missing /api/v1
    'GET /permissions/users', // ❌ Missing /api/v1
    'GET /permissions/users/:userId', // ❌ Missing /api/v1
    'PUT /permissions/users/:userId' // ❌ Missing /api/v1
  ],

  // paymentService.ts - Perlu diupdate  
  paymentService: [
    'GET /payments', // ❌ Wrong endpoint - should be /api/payments
    'POST /payments', // ❌ Wrong endpoint - should be /api/payments
    'GET /payments/:id', // ❌ Wrong endpoint - should be /api/payments
    'PUT /payments/:id', // ❌ Wrong endpoint - should be /api/payments
    'DELETE /payments/:id', // ❌ Wrong endpoint - should be /api/payments
    'GET /payments/summary', // ❌ Wrong endpoint - should be /api/payments
    'GET /payments/analytics', // ❌ Wrong endpoint - should be /api/payments
    'POST /payments/enhanced-with-journal', // ❌ Wrong endpoint - should be /api/payments
    'GET /payments/account-balances', // ❌ Wrong endpoint - should be /api/payments
    'POST /payments/account-balances/refresh', // ❌ Wrong endpoint - should be /api/payments
    'GET /cashbank/accounts', // ❌ Missing /api
    'POST /cashbank/deposit', // ❌ Missing /api
    'POST /cashbank/withdrawal', // ❌ Missing /api
    'POST /cashbank/transfer' // ❌ Missing /api
  ],

  // productService.ts - Perlu diupdate
  productService: [
    'GET /products', // ❌ Missing /api/v1
    'GET /products/:id', // ❌ Missing /api/v1
    'POST /products', // ❌ Missing /api/v1
    'PUT /products/:id', // ❌ Missing /api/v1
    'DELETE /products/:id', // ❌ Missing /api/v1
    'POST /products/adjust-stock', // ❌ Missing /api/v1
    'POST /products/opname', // ❌ Missing /api/v1
    'POST /products/upload-image', // ❌ Missing /api/v1
    'GET /categories', // ❌ Missing /api/v1
    'POST /categories', // ❌ Missing /api/v1
    'GET /categories/:id', // ❌ Missing /api/v1
    'PUT /categories/:id', // ❌ Missing /api/v1
    'DELETE /categories/:id', // ❌ Missing /api/v1
    'GET /product-units', // ❌ Missing /api/v1
    'POST /product-units', // ❌ Missing /api/v1
    'GET /inventory/movements', // ❌ Missing /api/v1
    'GET /inventory/low-stock', // ❌ Missing /api/v1
    'POST /inventory/bulk-price-update' // ❌ Missing /api/v1
  ],

  // assetService.ts - Perlu diupdate
  assetService: [
    'GET /assets', // ❌ Missing /api/v1
    'POST /assets', // ❌ Missing /api/v1
    'GET /assets/:id', // ❌ Missing /api/v1
    'PUT /assets/:id', // ❌ Missing /api/v1
    'DELETE /assets/:id', // ❌ Missing /api/v1
    'POST /assets/upload-image', // ❌ Missing /api/v1
    'GET /assets/categories', // ❌ Missing /api/v1
    'POST /assets/categories', // ❌ Missing /api/v1
    'GET /assets/summary', // ❌ Missing /api/v1
    'GET /assets/depreciation-report', // ❌ Missing /api/v1
    'GET /assets/:id/depreciation-schedule', // ❌ Missing /api/v1
    'GET /assets/:id/calculate-depreciation' // ❌ Missing /api/v1
  ],

  // contactService.ts - Perlu diupdate
  contactService: [
    'GET /contacts', // ❌ Missing /api/v1
    'GET /contacts/:id', // ❌ Missing /api/v1
    'POST /contacts', // ❌ Missing /api/v1
    'PUT /contacts/:id', // ❌ Missing /api/v1
    'DELETE /contacts/:id', // ❌ Missing /api/v1
    'GET /contacts/type/:type', // ❌ Missing /api/v1
    'GET /contacts/search' // ❌ Missing /api/v1
  ],

  // accountService.ts - Perlu diupdate (jika ada)
  accountService: [
    'GET /accounts', // ❌ Missing /api/v1
    'POST /accounts', // ❌ Missing /api/v1
    'GET /accounts/hierarchy', // ❌ Missing /api/v1
    'GET /accounts/balance-summary' // ❌ Missing /api/v1
  ],

  // cashbankService.ts - Perlu diupdate
  cashbankService: [
    'GET /cashbank/accounts', // ❌ Missing /api
    'POST /cashbank/accounts', // ❌ Missing /api
    'GET /cashbank/accounts/:id', // ❌ Missing /api
    'PUT /cashbank/accounts/:id', // ❌ Missing /api
    'DELETE /cashbank/accounts/:id', // ❌ Missing /api
    'GET /cashbank/balance-summary', // ❌ Missing /api
    'POST /cashbank/deposit', // ❌ Missing /api
    'POST /cashbank/withdrawal', // ❌ Missing /api
    'POST /cashbank/transfer', // ❌ Missing /api
    'GET /cashbank/payment-accounts' // ❌ Missing /api
  ],

  // approvalService.ts - Mixed, beberapa perlu update
  approvalService: [
    'GET /purchases/pending-approval', // ❌ Missing /api/v1
    'GET /purchases/:id/approval-history', // ❌ Missing /api/v1
    'POST /purchases/:id/approve', // ❌ Missing /api/v1
    'POST /purchases/:id/reject', // ❌ Missing /api/v1
    'GET /notifications/approvals', // ✅ Correct - using API_ENDPOINTS
    'PUT /notifications/:id/read', // ✅ Correct - using API_ENDPOINTS
    'GET /sales/:id/payments', // ❌ Missing /api/v1
    'POST /sales/:id/integrated-payment' // ❌ Missing /api/v1
  ],

  // financialReportService.ts - Mixed
  financialReportService: [
    'GET /reports/profit-loss', // ❌ Missing /api/v1
    'GET /reports/balance-sheet', // ❌ Missing /api/v1
    'GET /reports/trial-balance', // ❌ Missing /api/v1
    'GET /reports/general-ledger', // ❌ Missing /api/v1
    'GET /reports/cash-flow', // ❌ Missing /api/v1
    'GET /enhanced-reports/profit-loss', // ❌ Wrong endpoint
    'GET /enhanced-reports/balance-sheet', // ❌ Wrong endpoint
    'POST /enhanced-reports/refresh', // ❌ Wrong endpoint
    'GET /ssot-reports/trial-balance', // ✅ Alias routes exist
    'GET /ssot-reports/general-ledger', // ✅ Alias routes exist
    'GET /accounts/hierarchy' // ❌ Missing /api/v1
  ],

  // SSOT Services - Correct (menggunakan alias routes)
  ssotServices: [
    'GET /ssot-reports/trial-balance', // ✅ Correct - alias route exists
    'GET /ssot-reports/general-ledger', // ✅ Correct - alias route exists  
    'GET /ssot-reports/journal-analysis', // ✅ Correct - alias route exists
    'GET /ssot-reports/purchase-report', // ✅ Correct - alias route exists
    'GET /reports/ssot-profit-loss', // ✅ Correct - direct route exists
    'GET /reports/ssot/balance-sheet', // ✅ Correct - direct route exists
    'GET /reports/ssot/cash-flow' // ✅ Correct - direct route exists
  ]
};

// Hitung total API calls dan identifikasi masalah
let totalCalls = 0;
let incorrectCalls = 0;
let correctCalls = 0;

Object.keys(frontendApiCalls).forEach(service => {
  console.log(`${service.toUpperCase()}:`);
  frontendApiCalls[service].forEach(call => {
    totalCalls++;
    if (call.includes('❌')) {
      incorrectCalls++;
      console.log(`  ${call}`);
    } else if (call.includes('✅')) {
      correctCalls++;
      console.log(`  ${call}`);
    } else {
      console.log(`  ${call}`);
    }
  });
  console.log('');
});

console.log('=== SUMMARY ===');
console.log(`Total API calls: ${totalCalls}`);
console.log(`Correct calls: ${correctCalls}`);
console.log(`Incorrect calls: ${incorrectCalls}`);
console.log(`Accuracy: ${Math.round((correctCalls/totalCalls)*100)}%`);
console.log('\n=== ANALYSIS COMPLETE ===');