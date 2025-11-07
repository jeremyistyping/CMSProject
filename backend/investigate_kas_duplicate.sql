-- INVESTIGASI DUPLIKASI AKUN KAS
-- ================================

-- 1. Tampilkan semua akun yang mengandung kata "Kas"
SELECT 
    id,
    code,
    name,
    type,
    balance,
    is_active,
    is_header,
    parent_id,
    level,
    created_at
FROM accounts 
WHERE name LIKE '%Kas%'
ORDER BY code;

-- 2. Cari duplikasi berdasarkan nama yang sama
SELECT 
    name, 
    COUNT(*) as count,
    GROUP_CONCAT(CONCAT(code, ' (ID:', id, ')') ORDER BY code) as accounts
FROM accounts 
GROUP BY name 
HAVING COUNT(*) > 1
ORDER BY count DESC;

-- 3. Detail akun Kas dengan informasi lengkap
SELECT 
    a.id,
    a.code,
    a.name,
    a.type,
    a.balance,
    a.is_active,
    a.is_header,
    a.parent_id,
    CASE 
        WHEN a.parent_id IS NOT NULL 
        THEN CONCAT(p.code, ' - ', p.name)
        ELSE 'ROOT'
    END as parent_account
FROM accounts a
LEFT JOIN accounts p ON a.parent_id = p.id
WHERE a.name LIKE '%Kas%'
ORDER BY a.code;

-- 4. Cek penggunaan akun Kas dalam journal entries
SELECT 
    a.id,
    a.code,
    a.name,
    COUNT(jl.id) as journal_entries_count,
    COALESCE(SUM(jl.debit_amount), 0) as total_debit,
    COALESCE(SUM(jl.credit_amount), 0) as total_credit,
    COALESCE(SUM(jl.debit_amount), 0) - COALESCE(SUM(jl.credit_amount), 0) as net_balance
FROM accounts a
LEFT JOIN journal_lines jl ON a.id = jl.account_id
WHERE a.name LIKE '%Kas%'
GROUP BY a.id, a.code, a.name
ORDER BY a.code;

-- 5. Tampilkan struktur hierarki untuk akun current assets
WITH RECURSIVE account_tree AS (
    -- Base case: accounts with no parent (root accounts)
    SELECT 
        id, 
        code, 
        name, 
        parent_id, 
        level,
        CAST(name AS CHAR(1000)) as path,
        0 as depth
    FROM accounts 
    WHERE parent_id IS NULL AND code LIKE '11%'
    
    UNION ALL
    
    -- Recursive case: child accounts
    SELECT 
        a.id, 
        a.code, 
        a.name, 
        a.parent_id, 
        a.level,
        CONCAT(at.path, ' > ', a.name),
        at.depth + 1
    FROM accounts a
    INNER JOIN account_tree at ON a.parent_id = at.id
    WHERE at.depth < 10 -- Prevent infinite recursion
)
SELECT 
    REPEAT('  ', depth) as indent,
    code,
    name,
    depth
FROM account_tree
ORDER BY code;

-- 6. Identifikasi akun yang berpotensi duplikat
SELECT 
    'DUPLICATE_CHECK' as issue_type,
    a1.id as account1_id,
    a1.code as account1_code,
    a1.name as account1_name,
    a1.is_active as account1_active,
    a2.id as account2_id,
    a2.code as account2_code,
    a2.name as account2_name,
    a2.is_active as account2_active,
    'Same name, different code' as description
FROM accounts a1
JOIN accounts a2 ON a1.name = a2.name AND a1.id != a2.id
WHERE a1.name LIKE '%Kas%'
ORDER BY a1.name, a1.code;

-- 7. Rekomendasi perbaikan
SELECT 
    'RECOMMENDATION' as action_type,
    id,
    code,
    name,
    CASE 
        WHEN code = '1101' AND name = 'Kas' AND is_active = 1 THEN 'KEEP - Standard cash account'
        WHEN code LIKE '1100-075' AND name = 'Kas' THEN 'REVIEW - Non-standard code'
        WHEN is_active = 0 THEN 'SAFE_TO_DELETE - Inactive account'
        WHEN is_header = 1 THEN 'KEEP - Header account'
        ELSE 'REVIEW - Check if needed'
    END as recommendation,
    CASE
        WHEN EXISTS(SELECT 1 FROM journal_lines jl WHERE jl.account_id = accounts.id) 
        THEN 'HAS_TRANSACTIONS'
        ELSE 'NO_TRANSACTIONS'
    END as transaction_status
FROM accounts 
WHERE name LIKE '%Kas%'
ORDER BY 
    CASE 
        WHEN code = '1101' THEN 1
        WHEN is_active = 1 AND is_header = 0 THEN 2
        WHEN is_header = 1 THEN 3
        ELSE 4
    END,
    code;