// Script untuk mengekstrak semua API routes dari backend
const fs = require('fs');
const path = require('path');

console.log('=== BACKEND ROUTES EXTRACTION ===\n');

// Routes from main routes.go (based on manual analysis)
const backendRoutes = {
  // Auth routes (under /api/v1/auth)
  auth: [
    'POST /api/v1/auth/login',
    'POST /api/v1/auth/register',
    'POST /api/v1/auth/refresh', 
    'GET /api/v1/auth/validate-token'
  ],

  // Profile (under /api/v1)
  profile: [
    'GET /api/v1/profile'
  ],

  // Journal Drilldown (under /api/v1/journal-drilldown)
  journalDrilldown: [
    'POST /api/v1/journal-drilldown',
    'GET /api/v1/journal-drilldown/entries',
    'GET /api/v1/journal-drilldown/entries/:id',
    'GET /api/v1/journal-drilldown/accounts'
  ],

  // Unified Journals (under /api/v1/journals)
  journals: [
    'POST /api/v1/journals',
    'GET /api/v1/journals',
    'GET /api/v1/journals/:id',
    'GET /api/v1/journals/account-balances',
    'POST /api/v1/journals/account-balances/refresh',
    'GET /api/v1/journals/summary'
  ],

  // Users (under /api/v1/users) 
  users: [
    'GET /api/v1/users',
    'GET /api/v1/users/:id',
    'POST /api/v1/users',
    'PUT /api/v1/users/:id',
    'DELETE /api/v1/users/:id'
  ],

  // Permissions (under /api/v1/permissions)
  permissions: [
    'GET /api/v1/permissions/users',
    'GET /api/v1/permissions/users/:userId', 
    'PUT /api/v1/permissions/users/:userId',
    'POST /api/v1/permissions/users/:userId/reset',
    'GET /api/v1/permissions/me',
    'GET /api/v1/permissions/check'
  ],

  // Dashboard (under /api/v1/dashboard)
  dashboard: [
    'GET /api/v1/dashboard/analytics',
    'GET /api/v1/dashboard/finance',
    'GET /api/v1/dashboard/stock-alerts',
    'POST /api/v1/dashboard/stock-alerts/:id/dismiss'
  ],

  // Products (under /api/v1/products)
  products: [
    'GET /api/v1/products',
    'GET /api/v1/products/:id',
    'POST /api/v1/products',
    'PUT /api/v1/products/:id',
    'DELETE /api/v1/products/:id',
    'POST /api/v1/products/adjust-stock',
    'POST /api/v1/products/opname',
    'POST /api/v1/products/upload-image'
  ],

  // Categories (under /api/v1/categories)
  categories: [
    'GET /api/v1/categories',
    'GET /api/v1/categories/tree',
    'GET /api/v1/categories/:id',
    'GET /api/v1/categories/:id/products',
    'POST /api/v1/categories',
    'PUT /api/v1/categories/:id', 
    'DELETE /api/v1/categories/:id'
  ],

  // Product Units (under /api/v1/product-units)
  productUnits: [
    'GET /api/v1/product-units',
    'GET /api/v1/product-units/:id',
    'POST /api/v1/product-units',
    'PUT /api/v1/product-units/:id',
    'DELETE /api/v1/product-units/:id'
  ],

  // Warehouse Locations (under /api/v1/warehouse-locations) 
  warehouseLocations: [
    'GET /api/v1/warehouse-locations',
    'GET /api/v1/warehouse-locations/:id',
    'POST /api/v1/warehouse-locations',
    'PUT /api/v1/warehouse-locations/:id',
    'DELETE /api/v1/warehouse-locations/:id'
  ],

  // Accounts (under /api/v1/accounts)
  accounts: [
    'GET /api/v1/accounts',
    'GET /api/v1/accounts/hierarchy',
    'GET /api/v1/accounts/balance-summary',
    'GET /api/v1/accounts/validate-code',
    'POST /api/v1/accounts/fix-header-status',
    'GET /api/v1/accounts/:code',
    'POST /api/v1/accounts',
    'PUT /api/v1/accounts/:code',
    'DELETE /api/v1/accounts/:code',
    'DELETE /api/v1/accounts/admin/:code',
    'POST /api/v1/accounts/import',
    'GET /api/v1/accounts/export/pdf',
    'GET /api/v1/accounts/export/excel'
  ],

  // Public Account Catalog (no auth required)
  publicAccounts: [
    'GET /api/v1/accounts/catalog',
    'GET /api/v1/accounts/credit'
  ],

  // Contacts (under /api/v1/contacts)
  contacts: [
    'GET /api/v1/contacts',
    'GET /api/v1/contacts/:id',
    'POST /api/v1/contacts',
    'PUT /api/v1/contacts/:id',
    'DELETE /api/v1/contacts/:id',
    'GET /api/v1/contacts/type/:type',
    'GET /api/v1/contacts/search',
    'POST /api/v1/contacts/import',
    'GET /api/v1/contacts/export',
    'POST /api/v1/contacts/:id/addresses',
    'PUT /api/v1/contacts/:id/addresses/:address_id',
    'DELETE /api/v1/contacts/:id/addresses/:address_id'
  ],

  // Notifications (under /api/v1/notifications)
  notifications: [
    'GET /api/v1/notifications',
    'GET /api/v1/notifications/unread-count',
    'PUT /api/v1/notifications/:id/read',
    'PUT /api/v1/notifications/read-all',
    'GET /api/v1/notifications/type/:type',
    'GET /api/v1/notifications/approvals'
  ],

  // Sales (under /api/v1/sales)
  sales: [
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
    'GET /api/v1/sales/:id/for-payment',
    'POST /api/v1/sales/:id/integrated-payment',
    'POST /api/v1/sales/:id/returns',
    'GET /api/v1/sales/returns',
    'GET /api/v1/sales/summary',
    'GET /api/v1/sales/analytics',
    'GET /api/v1/sales/receivables',
    'GET /api/v1/sales/:id/invoice/pdf',
    'GET /api/v1/sales/report/pdf',
    'GET /api/v1/sales/customer/:customer_id',
    'GET /api/v1/sales/customer/:customer_id/invoices'
  ],

  // Purchases (under /api/v1/purchases)
  purchases: [
    'GET /api/v1/purchases',
    'GET /api/v1/purchases/approval-stats',
    'GET /api/v1/purchases/:id',
    'POST /api/v1/purchases',
    'PUT /api/v1/purchases/:id',
    'DELETE /api/v1/purchases/:id',
    'POST /api/v1/purchases/:id/submit-approval',
    'POST /api/v1/purchases/:id/approve',
    'POST /api/v1/purchases/:id/reject',
    'GET /api/v1/purchases/:id/approval-history',
    'GET /api/v1/purchases/pending-approval',
    'POST /api/v1/purchases/:id/documents',
    'GET /api/v1/purchases/:id/documents',
    'DELETE /api/v1/purchases/documents/:document_id',
    'POST /api/v1/purchases/receipts',
    'GET /api/v1/purchases/:id/receipts',
    'GET /api/v1/purchases/receipts/:receipt_id/pdf',
    'GET /api/v1/purchases/:id/receipts/pdf',
    'GET /api/v1/purchases/summary',
    'GET /api/v1/purchases/pending-approvals',
    'GET /api/v1/purchases/dashboard',
    'GET /api/v1/purchases/vendor/:vendor_id/summary',
    'GET /api/v1/purchases/:id/payments',
    'POST /api/v1/purchases/:id/payments',
    'GET /api/v1/purchases/:id/for-payment',
    'POST /api/v1/purchases/:id/integrated-payment',
    'GET /api/v1/purchases/:id/matching',
    'POST /api/v1/purchases/:id/validate-matching',
    'GET /api/v1/purchases/:id/journal-entries'
  ],

  // Assets (under /api/v1/assets)
  assets: [
    'GET /api/v1/assets',
    'GET /api/v1/assets/:id',
    'POST /api/v1/assets',
    'PUT /api/v1/assets/:id',
    'DELETE /api/v1/assets/:id',
    'POST /api/v1/assets/upload-image',
    'GET /api/v1/assets/categories',
    'POST /api/v1/assets/categories',
    'GET /api/v1/assets/summary',
    'GET /api/v1/assets/depreciation-report',
    'GET /api/v1/assets/:id/depreciation-schedule',
    'GET /api/v1/assets/:id/calculate-depreciation'
  ],

  // Inventory (under /api/v1/inventory)
  inventory: [
    'GET /api/v1/inventory/movements',
    'GET /api/v1/inventory/low-stock',
    'GET /api/v1/inventory/valuation',
    'GET /api/v1/inventory/report',
    'POST /api/v1/inventory/bulk-price-update'
  ],

  // Approval Workflows (under /api/v1/approval-workflows)
  approvalWorkflows: [
    'GET /api/v1/approval-workflows',
    'POST /api/v1/approval-workflows'
  ]
};

// Count total endpoints
let totalEndpoints = 0;
Object.keys(backendRoutes).forEach(category => {
  console.log(`${category.toUpperCase()}:`);
  backendRoutes[category].forEach(route => {
    console.log(`  ${route}`);
    totalEndpoints++;
  });
  console.log('');
});

console.log(`TOTAL BACKEND ENDPOINTS: ${totalEndpoints}\n`);
console.log('=== EXTRACTION COMPLETE ===');