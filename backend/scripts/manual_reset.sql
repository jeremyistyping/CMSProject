-- MANUAL RESET DATA TRANSAKSI - SIMPLE VERSION
-- Script sederhana untuk reset data transaksi yang ada di database

-- Backup COA terlebih dahulu
CREATE TABLE IF NOT EXISTS accounts_backup_manual AS 
SELECT * FROM accounts WHERE deleted_at IS NULL;

-- Reset data transaksi yang ada (hanya yang table-nya ada)

-- Sales
DELETE FROM sale_items WHERE id > 0;
DELETE FROM sales WHERE id > 0;

-- Purchases  
DELETE FROM purchase_items WHERE id > 0;
DELETE FROM purchases WHERE id > 0;

-- Journals
DELETE FROM journal_entries WHERE id > 0;
DELETE FROM journals WHERE id > 0;

-- Payments
DELETE FROM payments WHERE id > 0;

-- Reset balances
UPDATE accounts SET balance = 0 WHERE id > 0;
UPDATE cash_banks SET balance = 0 WHERE id > 0;
UPDATE products SET stock = 0 WHERE id > 0;

-- Reset sequences yang ada
DO $$
DECLARE
    seq_name text;
    sequences text[] := ARRAY['sales_id_seq', 'sale_items_id_seq', 'purchases_id_seq', 'purchase_items_id_seq', 'journals_id_seq', 'journal_entries_id_seq', 'payments_id_seq'];
BEGIN
    FOREACH seq_name IN ARRAY sequences
    LOOP
        IF EXISTS (SELECT 1 FROM pg_class WHERE relname = seq_name AND relkind = 'S') THEN
            EXECUTE 'ALTER SEQUENCE ' || seq_name || ' RESTART WITH 1';
        END IF;
    END LOOP;
END $$;

-- Summary
SELECT 
    'MANUAL_RESET_COMPLETE' as status,
    (SELECT COUNT(*) FROM accounts WHERE deleted_at IS NULL) as coa_preserved,
    (SELECT COUNT(*) FROM products WHERE deleted_at IS NULL) as products_preserved,
    (SELECT COUNT(*) FROM contacts WHERE deleted_at IS NULL) as contacts_preserved,
    (SELECT COUNT(*) FROM cash_banks WHERE deleted_at IS NULL) as cashbanks_preserved,
    (SELECT COUNT(*) FROM sales) as remaining_sales,
    (SELECT COUNT(*) FROM purchases) as remaining_purchases,
    (SELECT COUNT(*) FROM journals) as remaining_journals,
    (SELECT COUNT(*) FROM payments) as remaining_payments,
    CURRENT_TIMESTAMP as reset_time;
