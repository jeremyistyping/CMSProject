-- ==========================================
-- CHECK ACCOUNT BALANCE SYNCHRONIZATION
-- Bandingkan balance di accounts table vs calculated from journal
-- ==========================================

-- 1. Check accounts yang balance-nya tidak sync dengan journal
SELECT 
    a.code,
    a.name,
    a.type,
    a.balance as stored_balance,
    COALESCE((
        SELECT 
            CASE 
                WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                    SUM(ujl.debit_amount - ujl.credit_amount)
                ELSE 
                    SUM(ujl.credit_amount - ujl.debit_amount)
            END
        FROM unified_journal_lines ujl
        JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
        WHERE ujl.account_id = a.id 
          AND uje.status = 'POSTED'
          AND uje.deleted_at IS NULL
    ), 0) as calculated_balance,
    COALESCE((
        SELECT 
            CASE 
                WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                    SUM(ujl.debit_amount - ujl.credit_amount)
                ELSE 
                    SUM(ujl.credit_amount - ujl.debit_amount)
            END
        FROM unified_journal_lines ujl
        JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
        WHERE ujl.account_id = a.id 
          AND uje.status = 'POSTED'
          AND uje.deleted_at IS NULL
    ), 0) - a.balance as difference
FROM accounts a
WHERE a.deleted_at IS NULL
  AND a.is_header = false
  AND (
    -- Balance tidak sama dengan calculated
    COALESCE((
        SELECT 
            CASE 
                WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                    SUM(ujl.debit_amount - ujl.credit_amount)
                ELSE 
                    SUM(ujl.credit_amount - ujl.debit_amount)
            END
        FROM unified_journal_lines ujl
        JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
        WHERE ujl.account_id = a.id 
          AND uje.status = 'POSTED'
          AND uje.deleted_at IS NULL
    ), 0) != a.balance
  )
ORDER BY ABS(COALESCE((
    SELECT 
        CASE 
            WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                SUM(ujl.debit_amount - ujl.credit_amount)
            ELSE 
                SUM(ujl.credit_amount - ujl.debit_amount)
        END
    FROM unified_journal_lines ujl
    JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
    WHERE ujl.account_id = a.id 
      AND uje.status = 'POSTED'
      AND uje.deleted_at IS NULL
), 0) - a.balance) DESC;

-- 2. Summary: Total accounts yang out of sync
SELECT 
    'Total Accounts' as metric,
    COUNT(*) as count
FROM accounts
WHERE deleted_at IS NULL AND is_header = false
UNION ALL
SELECT 
    'Out of Sync' as metric,
    COUNT(*) as count
FROM accounts a
WHERE a.deleted_at IS NULL
  AND a.is_header = false
  AND (
    COALESCE((
        SELECT 
            CASE 
                WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                    SUM(ujl.debit_amount - ujl.credit_amount)
                ELSE 
                    SUM(ujl.credit_amount - ujl.debit_amount)
            END
        FROM unified_journal_lines ujl
        JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
        WHERE ujl.account_id = a.id 
          AND uje.status = 'POSTED'
          AND uje.deleted_at IS NULL
    ), 0) != a.balance
  );

-- 3. Check specific accounts dari screenshot
SELECT 
    a.code,
    a.name,
    a.type,
    a.balance as stored_balance,
    COALESCE((
        SELECT 
            CASE 
                WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                    SUM(ujl.debit_amount - ujl.credit_amount)
                ELSE 
                    SUM(ujl.credit_amount - ujl.debit_amount)
            END
        FROM unified_journal_lines ujl
        JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
        WHERE ujl.account_id = a.id 
          AND uje.status = 'POSTED'
          AND uje.deleted_at IS NULL
    ), 0) as calculated_balance,
    (SELECT COUNT(*) 
     FROM unified_journal_lines ujl
     JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
     WHERE ujl.account_id = a.id 
       AND uje.status = 'POSTED'
       AND uje.deleted_at IS NULL) as journal_entry_count
FROM accounts a
WHERE a.deleted_at IS NULL
  AND a.code IN (
    '4101',  -- PENDAPATAN PENJUALAN
    '5101',  -- HARGA POKOK PENJUALAN
    '1201',  -- PIUTANG USAHA
    '1301',  -- PERSEDIAAN BARANG
    '2103',  -- PPN KELUARAN
    '1240',  -- PPN MASUKAN
    '2101',  -- HUTANG USAHA
    '1101',  -- KAS
    '1102'   -- BANK
  )
ORDER BY a.code;

-- 4. Check journal entries untuk account tertentu
SELECT 
    uje.entry_date,
    uje.reference,
    uje.description,
    a.code,
    a.name,
    ujl.debit_amount,
    ujl.credit_amount,
    CASE 
        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
            ujl.debit_amount - ujl.credit_amount
        ELSE 
            ujl.credit_amount - ujl.debit_amount
    END as net_effect
FROM unified_journal_ledger uje
JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
JOIN accounts a ON a.id = ujl.account_id
WHERE a.code = '4101'  -- PENDAPATAN PENJUALAN
  AND uje.status = 'POSTED'
  AND uje.deleted_at IS NULL
ORDER BY uje.entry_date DESC, uje.id DESC
LIMIT 10;
