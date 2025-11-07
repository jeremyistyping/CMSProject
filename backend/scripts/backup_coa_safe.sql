-- Backup Chart of Accounts (COA) Script - Safe Version
-- Script ini akan membuat backup dari tabel accounts sebelum reset data transaksi
-- Versi yang lebih aman dan handle table yang tidak ada

-- Buat tabel backup untuk accounts
CREATE TABLE IF NOT EXISTS accounts_backup AS 
SELECT 
    id, code, name, description, type, category, parent_id, 
    level, is_header, is_active, balance, 
    created_at, updated_at
FROM accounts 
WHERE deleted_at IS NULL;

-- Backup struktur hierarki accounts
CREATE TABLE IF NOT EXISTS accounts_hierarchy_backup AS
WITH RECURSIVE account_tree AS (
    -- Base case: root accounts (no parent)
    SELECT 
        id, code, name, parent_id, level, 
        ARRAY[id] as path,
        code::text as full_path
    FROM accounts 
    WHERE parent_id IS NULL AND deleted_at IS NULL
    
    UNION ALL
    
    -- Recursive case: child accounts
    SELECT 
        a.id, a.code, a.name, a.parent_id, a.level,
        at.path || a.id,
        at.full_path || ' > ' || a.code
    FROM accounts a
    INNER JOIN account_tree at ON a.parent_id = at.id
    WHERE a.deleted_at IS NULL
)
SELECT * FROM account_tree;

-- Backup account balances (reset ke 0 nanti)
CREATE TABLE IF NOT EXISTS accounts_original_balances AS
SELECT 
    id, code, name, balance, 
    'ORIGINAL' as balance_type,
    CURRENT_TIMESTAMP as backup_date
FROM accounts 
WHERE deleted_at IS NULL AND balance != 0;

-- Summary informasi backup
SELECT 
    'BACKUP_SUMMARY' as status,
    (SELECT COUNT(*) FROM accounts_backup) as total_accounts_backed_up,
    (SELECT COUNT(*) FROM accounts_hierarchy_backup) as hierarchy_records,
    (SELECT COUNT(*) FROM accounts_original_balances) as accounts_with_balance,
    CURRENT_TIMESTAMP as backup_timestamp;
