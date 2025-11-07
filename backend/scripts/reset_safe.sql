-- SAFE RESET TRANSACTION DATA WITH FOREIGN KEY HANDLING
-- Script ini akan disable foreign key checks sementara untuk reset yang aman

BEGIN;

-- Disable foreign key checks sementara
SET session_replication_role = replica;

-- =============================================================================
-- LANGKAH 1: HAPUS DATA DALAM URUTAN YANG BENAR (child to parent)
-- =============================================================================

-- 1. Hapus data paling dependen terlebih dahulu
DELETE FROM purchase_receipt_items;
DELETE FROM purchase_receipts;
DELETE FROM purchase_documents;

DELETE FROM sale_return_items;
DELETE FROM sale_returns;
DELETE FROM sale_payments;

DELETE FROM payment_allocations;
DELETE FROM reconciliation_items;
DELETE FROM bank_reconciliations;
DELETE FROM cash_bank_transfers;
DELETE FROM cash_bank_transactions;

-- 2. Hapus approval data
DELETE FROM approval_history;
DELETE FROM approval_actions;
DELETE FROM approval_requests;

-- 3. Hapus transaction items
DELETE FROM sale_items;
DELETE FROM purchase_items;

-- 4. Hapus main transaction tables
DELETE FROM sales;
DELETE FROM purchases;

-- 5. Hapus journal dan payment data
DELETE FROM journal_entries;
DELETE FROM journals;
DELETE FROM payments;

-- 6. Hapus data lainnya
DELETE FROM transactions;
DELETE FROM inventories;
DELETE FROM expenses;
DELETE FROM notifications;

-- 7. Hapus budget dan report data
DELETE FROM budget_comparisons;
DELETE FROM budget_items;
DELETE FROM budgets;
DELETE FROM financial_ratios;
DELETE FROM account_balances;
DELETE FROM reports;

-- =============================================================================
-- LANGKAH 2: RESET BALANCES KE 0
-- =============================================================================

-- Reset account balances
UPDATE accounts SET balance = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;

-- Reset cash bank balances  
UPDATE cash_banks SET balance = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;

-- Reset product stock
UPDATE products SET stock = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;

-- Reset asset depreciation (if exists)
UPDATE assets SET accumulated_depreciation = 0 WHERE id > 0;

-- =============================================================================
-- LANGKAH 3: RESET SEQUENCES
-- =============================================================================

-- Reset auto increment sequences
DO $$
DECLARE
    seq_name text;
    sequences text[] := ARRAY[
        'sales_id_seq', 'sale_items_id_seq', 'sale_payments_id_seq', 
        'sale_returns_id_seq', 'sale_return_items_id_seq',
        'purchases_id_seq', 'purchase_items_id_seq', 'purchase_receipts_id_seq',
        'purchase_receipt_items_id_seq', 'purchase_documents_id_seq',
        'journals_id_seq', 'journal_entries_id_seq', 'transactions_id_seq',
        'cash_bank_transactions_id_seq', 'cash_bank_transfers_id_seq',
        'payments_id_seq', 'payment_allocations_id_seq',
        'expenses_id_seq', 'inventories_id_seq',
        'budgets_id_seq', 'budget_items_id_seq', 'budget_comparisons_id_seq',
        'reports_id_seq', 'financial_ratios_id_seq', 'account_balances_id_seq',
        'notifications_id_seq', 'approval_requests_id_seq', 'approval_actions_id_seq',
        'approval_history_id_seq'
    ];
BEGIN
    FOREACH seq_name IN ARRAY sequences
    LOOP
        IF EXISTS (SELECT 1 FROM pg_class WHERE relname = seq_name AND relkind = 'S') THEN
            EXECUTE 'ALTER SEQUENCE ' || seq_name || ' RESTART WITH 1';
        END IF;
    END LOOP;
END $$;

-- =============================================================================
-- LANGKAH 4: RE-ENABLE FOREIGN KEY CHECKS
-- =============================================================================

-- Re-enable foreign key checks
SET session_replication_role = DEFAULT;

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
    (SELECT COUNT(*) FROM journals) as remaining_journals,
    (SELECT COUNT(*) FROM payments) as remaining_payments;