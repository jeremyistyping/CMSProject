# Ikhtisar API

Dokumen ini merangkum endpoint penting dan pola penggunaannya. Rujuk README dan folder backend/controllers untuk detail lengkap.

## Autentikasi & User
- POST /api/v1/auth/login, /auth/register, /auth/refresh
- GET/PUT /api/v1/profile

## Sales
- GET/POST /api/v1/sales
- GET/PUT /api/v1/sales/{id}
- POST /api/v1/sales/{id}/confirm
- POST /api/v1/sales/{id}/invoice
- POST /api/v1/sales/{id}/payments
- GET /api/v1/sales/analytics

## Purchase
- GET/POST /api/v1/purchases
- POST /api/v1/purchases/{id}/submit-approval
- POST /api/v1/purchases/{id}/approve
- POST /api/v1/purchases/receipts
- GET /api/v1/purchases/pending-approval

## Inventory & Product
- GET/POST /api/v1/products
- POST /api/v1/products/adjust-stock
- POST /api/v1/products/opname
- GET /api/v1/inventory/movements
- GET /api/v1/inventory/low-stock

## Financial & Journals
- GET /api/v1/accounts, /cash-banks
- POST /api/v1/payments, /cash-banks/transfer, /journal-entries
- Enhanced Reports & Unified Reporting (lihat 08_REPORTING.md)

## Monitoring & Security
- GET /api/v1/monitoring/status, /monitoring/audit-logs, /notifications
- GET /api/v1/security/dashboard, /admin/security/incidents

Tips:
- Gunakan token Bearer di Authorization header
- Manfaatkan filter query (tanggal, status, pagination) untuk efisiensi
