-- SAFE RESET TRANSACTION DATA SCRIPT
-- Script ini menggunakan pendekatan yang lebih aman dengan IF EXISTS checks

-- =============================================================================
-- LANGKAH 1: START TRANSACTION
-- =============================================================================
BEGIN;

-- =============================================================================
-- LANGKAH 2: HAPUS DATA TRANSAKSI (dengan IF EXISTS checks)
-- =============================================================================

-- Delete transaction data safely (check if tables exist first)
DO $$
BEGIN
    -- 1. APPROVAL RELATED DATA (check and delete safely)
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'approval_history') THEN
        DELETE FROM approval_history;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'approval_actions') THEN
        DELETE FROM approval_actions;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'approval_requests') THEN
        DELETE FROM approval_requests;
    END IF;
    
    -- 2. RECONCILIATION DATA
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'reconciliation_items') THEN
        DELETE FROM reconciliation_items;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'bank_reconciliations') THEN
        DELETE FROM bank_reconciliations;
    END IF;
    
    -- 3. PAYMENT ALLOCATIONS
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'payment_allocations') THEN
        DELETE FROM payment_allocations;
    END IF;
    
    -- 4. CASH BANK TRANSFERS
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'cash_bank_transfers') THEN
        DELETE FROM cash_bank_transfers;
    END IF;
    
    -- 5. CASH BANK TRANSACTIONS
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'cash_bank_transactions') THEN
        DELETE FROM cash_bank_transactions;
    END IF;
    
    -- 6. SALE RELATED DATA
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sale_return_items') THEN
        DELETE FROM sale_return_items;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sale_returns') THEN
        DELETE FROM sale_returns;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sale_payments') THEN
        DELETE FROM sale_payments;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sale_items') THEN
        DELETE FROM sale_items;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sales') THEN
        DELETE FROM sales;
    END IF;
    
    -- 7. PURCHASE RELATED DATA
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'purchase_receipt_items') THEN
        DELETE FROM purchase_receipt_items;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'purchase_receipts') THEN
        DELETE FROM purchase_receipts;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'purchase_documents') THEN
        DELETE FROM purchase_documents;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'purchase_items') THEN
        DELETE FROM purchase_items;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'purchases') THEN
        DELETE FROM purchases;
    END IF;
    
    -- 8. JOURNAL ENTRIES
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'journal_entries') THEN
        DELETE FROM journal_entries;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'journals') THEN
        DELETE FROM journals;
    END IF;
    
    -- 9. TRANSACTIONS
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'transactions') THEN
        DELETE FROM transactions;
    END IF;
    
    -- 10. INVENTORY MOVEMENTS
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'inventories') THEN
        DELETE FROM inventories;
    END IF;
    
    -- 11. EXPENSE DATA
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'expenses') THEN
        DELETE FROM expenses;
    END IF;
    
    -- 12. PAYMENT DATA
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'payments') THEN
        DELETE FROM payments;
    END IF;
    
    -- 13. BUDGET DATA
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'budget_comparisons') THEN
        DELETE FROM budget_comparisons;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'budget_items') THEN
        DELETE FROM budget_items;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'budgets') THEN
        DELETE FROM budgets;
    END IF;
    
    -- 14. REPORT DATA
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'financial_ratios') THEN
        DELETE FROM financial_ratios;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'account_balances') THEN
        DELETE FROM account_balances;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'reports') THEN
        DELETE FROM reports;
    END IF;
    
    -- 15. NOTIFICATION DATA
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'notifications') THEN
        DELETE FROM notifications;
    END IF;
    
END $$;

-- =============================================================================
-- LANGKAH 3: RESET BALANCES (dengan checks)
-- =============================================================================

-- Reset accounts balance ke 0 (jika table exists)
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'accounts') THEN
        UPDATE accounts SET balance = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;
    END IF;
END $$;

-- Reset cash_banks balance ke 0 (jika table exists)
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'cash_banks') THEN
        UPDATE cash_banks SET balance = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;
    END IF;
END $$;

-- Reset product stock ke 0 (jika table exists)
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'products') THEN
        UPDATE products SET stock = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'product_variants') THEN
        UPDATE product_variants SET stock = 0, updated_at = CURRENT_TIMESTAMP WHERE id > 0;
    END IF;
    
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'assets') THEN
        UPDATE assets SET accumulated_depreciation = 0 WHERE id > 0;
    END IF;
END $$;

-- =============================================================================
-- LANGKAH 4: RESET SEQUENCES (dengan checks)
-- =============================================================================

-- Reset sequences safely
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
        'notifications_id_seq'
    ];
BEGIN
    FOREACH seq_name IN ARRAY sequences
    LOOP
        -- Check if sequence exists before trying to restart it
        IF EXISTS (SELECT 1 FROM pg_class WHERE relname = seq_name AND relkind = 'S') THEN
            EXECUTE 'ALTER SEQUENCE ' || seq_name || ' RESTART WITH 1';
        END IF;
    END LOOP;
END $$;

-- =============================================================================
-- LANGKAH 5: COMMIT TRANSACTION
-- =============================================================================
COMMIT;

-- =============================================================================
-- SUMMARY REPORT
-- =============================================================================
SELECT 
    'RESET_COMPLETE' as status,
    'Data transaksi telah dihapus, COA dan master data dipertahankan' as message,
    CURRENT_TIMESTAMP as reset_timestamp,
    CASE WHEN EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'accounts') 
         THEN (SELECT COUNT(*) FROM accounts WHERE deleted_at IS NULL) 
         ELSE 0 END as total_coa_accounts_preserved,
    CASE WHEN EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'products') 
         THEN (SELECT COUNT(*) FROM products WHERE deleted_at IS NULL) 
         ELSE 0 END as total_products_preserved,
    CASE WHEN EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'contacts') 
         THEN (SELECT COUNT(*) FROM contacts WHERE deleted_at IS NULL) 
         ELSE 0 END as total_contacts_preserved,
    CASE WHEN EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'cash_banks') 
         THEN (SELECT COUNT(*) FROM cash_banks WHERE deleted_at IS NULL) 
         ELSE 0 END as total_cashbank_preserved,
    CASE WHEN EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sales') 
         THEN (SELECT COUNT(*) FROM sales) 
         ELSE 0 END as remaining_sales,
    CASE WHEN EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'purchases') 
         THEN (SELECT COUNT(*) FROM purchases) 
         ELSE 0 END as remaining_purchases,
    CASE WHEN EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'transactions') 
         THEN (SELECT COUNT(*) FROM transactions) 
         ELSE 0 END as remaining_transactions,
    CASE WHEN EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'journals') 
         THEN (SELECT COUNT(*) FROM journals) 
         ELSE 0 END as remaining_journals;
