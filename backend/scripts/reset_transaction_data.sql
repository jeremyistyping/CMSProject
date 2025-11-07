-- RESET TRANSACTION DATA SCRIPT
-- Script ini akan mengosongkan SEMUA data transaksi namun mempertahankan Chart of Accounts (COA)
-- PERINGATAN: Script ini akan menghapus semua data transaksi secara permanen!

-- =============================================================================
-- LANGKAH 1: DISABLE FOREIGN KEY CHECKS (untuk PostgreSQL gunakan transaction)
-- =============================================================================
BEGIN;

-- =============================================================================
-- LANGKAH 2: HAPUS DATA TRANSAKSI (Urutan berdasarkan foreign key dependencies)
-- =============================================================================

-- 1. HAPUS APPROVAL RELATED DATA
DELETE FROM approval_history;
DELETE FROM approval_actions;
DELETE FROM approval_requests;

-- 2. HAPUS RECONCILIATION DATA  
DELETE FROM reconciliation_items;
DELETE FROM bank_reconciliations;

-- 3. HAPUS PAYMENT ALLOCATIONS
DELETE FROM payment_allocations;

-- 4. HAPUS CASH BANK TRANSFERS
DELETE FROM cash_bank_transfers;

-- 5. HAPUS CASH BANK TRANSACTIONS
DELETE FROM cash_bank_transactions;

-- 6. HAPUS SALE RELATED DATA
DELETE FROM sale_return_items;
DELETE FROM sale_returns;
DELETE FROM sale_payments;
DELETE FROM sale_items;
DELETE FROM sales;

-- 7. HAPUS PURCHASE RELATED DATA  
DELETE FROM purchase_receipt_items;
DELETE FROM purchase_receipts;
DELETE FROM purchase_documents;
DELETE FROM purchase_items;
DELETE FROM purchases;

-- 8. HAPUS JOURNAL ENTRIES
DELETE FROM journal_entries;
DELETE FROM journals;

-- 9. HAPUS TRANSACTIONS
DELETE FROM transactions;

-- 10. HAPUS INVENTORY MOVEMENTS
DELETE FROM inventories;

-- 11. HAPUS EXPENSE DATA
DELETE FROM expenses;

-- 12. HAPUS ASSET TRANSACTIONS (tapi biarkan asset master data)
-- Asset tetap ada tapi reset accumulated depreciation
UPDATE assets SET accumulated_depreciation = 0 WHERE id > 0;

-- 13. HAPUS PAYMENT DATA
DELETE FROM payments;

-- 14. HAPUS BUDGET DATA
DELETE FROM budget_comparisons;
DELETE FROM budget_items;
DELETE FROM budgets;

-- 15. HAPUS REPORT DATA
DELETE FROM financial_ratios;
DELETE FROM account_balances;
DELETE FROM reports;

-- 16. HAPUS NOTIFICATION DATA
DELETE FROM notifications;

-- =============================================================================
-- LANGKAH 3: RESET BALANCES DI COA (ACCOUNTS) KE 0
-- =============================================================================

-- Reset semua balance di accounts ke 0 tapi pertahankan struktur COA
UPDATE accounts SET balance = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;

-- =============================================================================
-- LANGKAH 4: RESET CASH BANK BALANCES KE 0
-- =============================================================================

-- Reset balance cash bank ke 0 tapi pertahankan master data
UPDATE cash_banks SET balance = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;

-- =============================================================================
-- LANGKAH 5: RESET PRODUCT STOCK KE 0
-- =============================================================================

-- Reset stock produk ke 0 tapi pertahankan master data produk
UPDATE products SET stock = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;
UPDATE product_variants SET stock = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;

-- =============================================================================
-- LANGKAH 6: RESET AUTO INCREMENT SEQUENCES
-- =============================================================================

-- Reset sequence ID untuk transaksi tables agar mulai dari 1 lagi
ALTER SEQUENCE sales_id_seq RESTART WITH 1;
ALTER SEQUENCE sale_items_id_seq RESTART WITH 1;
ALTER SEQUENCE sale_payments_id_seq RESTART WITH 1;
ALTER SEQUENCE sale_returns_id_seq RESTART WITH 1;
ALTER SEQUENCE sale_return_items_id_seq RESTART WITH 1;

ALTER SEQUENCE purchases_id_seq RESTART WITH 1;
ALTER SEQUENCE purchase_items_id_seq RESTART WITH 1;
ALTER SEQUENCE purchase_receipts_id_seq RESTART WITH 1;
ALTER SEQUENCE purchase_receipt_items_id_seq RESTART WITH 1;
ALTER SEQUENCE purchase_documents_id_seq RESTART WITH 1;

ALTER SEQUENCE journals_id_seq RESTART WITH 1;
ALTER SEQUENCE journal_entries_id_seq RESTART WITH 1;
ALTER SEQUENCE transactions_id_seq RESTART WITH 1;

ALTER SEQUENCE cash_bank_transactions_id_seq RESTART WITH 1;
ALTER SEQUENCE cash_bank_transfers_id_seq RESTART WITH 1;
ALTER SEQUENCE payments_id_seq RESTART WITH 1;
ALTER SEQUENCE payment_allocations_id_seq RESTART WITH 1;

ALTER SEQUENCE expenses_id_seq RESTART WITH 1;
ALTER SEQUENCE inventories_id_seq RESTART WITH 1;

ALTER SEQUENCE budgets_id_seq RESTART WITH 1;
ALTER SEQUENCE budget_items_id_seq RESTART WITH 1;
ALTER SEQUENCE budget_comparisons_id_seq RESTART WITH 1;

ALTER SEQUENCE reports_id_seq RESTART WITH 1;
ALTER SEQUENCE financial_ratios_id_seq RESTART WITH 1;
ALTER SEQUENCE account_balances_id_seq RESTART WITH 1;

ALTER SEQUENCE approval_workflows_id_seq RESTART WITH 1;
ALTER SEQUENCE approval_steps_id_seq RESTART WITH 1;
ALTER SEQUENCE approval_requests_id_seq RESTART WITH 1;
ALTER SEQUENCE approval_actions_id_seq RESTART WITH 1;
ALTER SEQUENCE approval_history_id_seq RESTART WITH 1;

ALTER SEQUENCE bank_reconciliations_id_seq RESTART WITH 1;
ALTER SEQUENCE reconciliation_items_id_seq RESTART WITH 1;

ALTER SEQUENCE notifications_id_seq RESTART WITH 1;

-- =============================================================================
-- LANGKAH 7: LOG RESET ACTIVITY
-- =============================================================================

-- Catat aktivitas reset dalam audit log (sesuai struktur audit_logs yang ada)
INSERT INTO audit_logs (action, table_name, record_id, old_values, new_values, user_id, created_at)
VALUES ('DELETE', 'ALL_TRANSACTION_TABLES', 0, '', 
        'Reset semua data transaksi, mempertahankan COA dan master data', 
        1, CURRENT_TIMESTAMP);

-- =============================================================================
-- LANGKAH 8: COMMIT TRANSACTION
-- =============================================================================
COMMIT;

-- =============================================================================
-- SUMMARY REPORT
-- =============================================================================
SELECT 
    'RESET_COMPLETE' as status,
    'Data transaksi telah dihapus, COA dan master data dipertahankan' as message,
    CURRENT_TIMESTAMP as reset_timestamp,
    (SELECT COUNT(*) FROM accounts WHERE deleted_at IS NULL) as total_coa_accounts_preserved,
    (SELECT COUNT(*) FROM products WHERE deleted_at IS NULL) as total_products_preserved,
    (SELECT COUNT(*) FROM contacts WHERE deleted_at IS NULL) as total_contacts_preserved,
    (SELECT COUNT(*) FROM cash_banks WHERE deleted_at IS NULL) as total_cashbank_preserved,
    (SELECT COUNT(*) FROM sales) as remaining_sales,
    (SELECT COUNT(*) FROM purchases) as remaining_purchases,
    (SELECT COUNT(*) FROM transactions) as remaining_transactions,
    (SELECT COUNT(*) FROM journals) as remaining_journals;
