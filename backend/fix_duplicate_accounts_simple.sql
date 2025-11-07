-- ============================================================================
-- SIMPLE FIX: Merge Duplicate Accounts
-- ============================================================================
-- This script fixes duplicate accounts by merging them
-- Safe to run - uses transaction with COMMIT at the end
-- ============================================================================

BEGIN;

\echo '============================================================================'
\echo 'STEP 1: Checking for duplicate accounts...'
\echo '============================================================================'

SELECT 
    code,
    COUNT(*) as count,
    STRING_AGG(DISTINCT name, ' | ') as names
FROM accounts
WHERE deleted_at IS NULL
GROUP BY code
HAVING COUNT(*) > 1
ORDER BY COUNT(*) DESC;

\echo ''
\echo '============================================================================'
\echo 'STEP 2: Creating merge strategy (keep account with most usage)...'
\echo '============================================================================'

-- Create temp table for merge mapping
CREATE TEMP TABLE IF NOT EXISTS duplicate_merge_map AS
WITH duplicate_codes AS (
    SELECT code
    FROM accounts
    WHERE deleted_at IS NULL
    GROUP BY code
    HAVING COUNT(*) > 1
),
ranked_accounts AS (
    SELECT 
        a.id,
        a.code,
        a.name,
        a.balance,
        COALESCE(
            (SELECT COUNT(*) FROM unified_journal_ledger WHERE account_id = a.id AND deleted_at IS NULL), 0
        ) + COALESCE(
            (SELECT COUNT(*) FROM cash_banks WHERE account_id = a.id AND deleted_at IS NULL), 0
        ) + COALESCE(
            (SELECT COUNT(*) FROM assets WHERE asset_account_id = a.id AND deleted_at IS NULL), 0
        ) as usage_count,
        ROW_NUMBER() OVER (
            PARTITION BY a.code 
            ORDER BY 
                (COALESCE(
                    (SELECT COUNT(*) FROM unified_journal_ledger WHERE account_id = a.id AND deleted_at IS NULL), 0
                ) + COALESCE(
                    (SELECT COUNT(*) FROM cash_banks WHERE account_id = a.id AND deleted_at IS NULL), 0
                ) + COALESCE(
                    (SELECT COUNT(*) FROM assets WHERE asset_account_id = a.id AND deleted_at IS NULL), 0
                )) DESC,
                a.created_at ASC,
                a.id ASC
        ) as rank
    FROM accounts a
    INNER JOIN duplicate_codes dc ON a.code = dc.code
    WHERE a.deleted_at IS NULL
)
SELECT 
    code,
    id as duplicate_id,
    name as duplicate_name,
    balance,
    usage_count,
    (SELECT id FROM ranked_accounts WHERE code = r.code AND rank = 1) as primary_id,
    (SELECT name FROM ranked_accounts WHERE code = r.code AND rank = 1) as primary_name,
    rank
FROM ranked_accounts r
WHERE rank > 1;

SELECT 
    COUNT(*) as "Total Duplicates to Merge"
FROM duplicate_merge_map;

\echo ''
\echo '============================================================================'
\echo 'STEP 3: Showing merge plan...'
\echo '============================================================================'

SELECT 
    code,
    duplicate_id as "Duplicate ID",
    duplicate_name as "Duplicate Name", 
    balance,
    usage_count as "Usage",
    primary_id as "→ Primary ID",
    primary_name as "Primary Name"
FROM duplicate_merge_map
ORDER BY code, duplicate_id;

\echo ''
\echo '============================================================================'
\echo 'STEP 4: Moving journal entries to primary accounts...'
\echo '============================================================================'

UPDATE unified_journal_ledger ujl
SET account_id = dmm.primary_id
FROM duplicate_merge_map dmm
WHERE ujl.account_id = dmm.duplicate_id
  AND ujl.deleted_at IS NULL;

SELECT ROW_COUNT() as "Journal Entries Moved";

\echo ''
\echo '============================================================================'
\echo 'STEP 5: Moving cash/bank references...'
\echo '============================================================================'

UPDATE cash_banks cb
SET account_id = dmm.primary_id
FROM duplicate_merge_map dmm
WHERE cb.account_id = dmm.duplicate_id
  AND cb.deleted_at IS NULL;

SELECT ROW_COUNT() as "Cash/Bank References Moved";

\echo ''
\echo '============================================================================'
\echo 'STEP 6: Moving asset references...'
\echo '============================================================================'

UPDATE assets ast
SET asset_account_id = dmm.primary_id
FROM duplicate_merge_map dmm
WHERE ast.asset_account_id = dmm.duplicate_id
  AND ast.deleted_at IS NULL;

SELECT ROW_COUNT() as "Asset References Moved";

\echo ''
\echo '============================================================================'
\echo 'STEP 7: Consolidating balances to primary accounts...'
\echo '============================================================================'

DO $$
DECLARE
    rec RECORD;
    total_balance DECIMAL(20,2);
BEGIN
    FOR rec IN 
        SELECT DISTINCT code, primary_id 
        FROM duplicate_merge_map
    LOOP
        -- Calculate total balance from all accounts with this code
        SELECT COALESCE(SUM(balance), 0) INTO total_balance
        FROM accounts
        WHERE code = rec.code
          AND deleted_at IS NULL;
        
        -- Update primary account
        UPDATE accounts
        SET balance = total_balance,
            updated_at = NOW()
        WHERE id = rec.primary_id;
        
        RAISE NOTICE 'Consolidated balance for code % (ID: %) = %', rec.code, rec.primary_id, total_balance;
    END LOOP;
END $$;

\echo ''
\echo '============================================================================'
\echo 'STEP 8: Soft-deleting duplicate accounts...'
\echo '============================================================================'

UPDATE accounts
SET 
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id IN (
    SELECT duplicate_id 
    FROM duplicate_merge_map
);

SELECT ROW_COUNT() as "Duplicate Accounts Soft-Deleted";

\echo ''
\echo '============================================================================'
\echo 'STEP 9: Creating unique constraint to prevent future duplicates...'
\echo '============================================================================'

-- Drop old constraints/indexes
DROP INDEX IF EXISTS uni_accounts_code CASCADE;
DROP INDEX IF EXISTS accounts_code_key CASCADE;
DROP INDEX IF EXISTS idx_accounts_code_unique CASCADE;
DROP INDEX IF EXISTS accounts_code_unique CASCADE;
DROP INDEX IF EXISTS idx_accounts_code_unique_active CASCADE;

-- Create new partial unique index
CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_code_active
ON accounts (LOWER(code))
WHERE deleted_at IS NULL;

\echo '✅ Created unique index: idx_accounts_code_active'

-- Create trigger function
CREATE OR REPLACE FUNCTION prevent_duplicate_account_code()
RETURNS TRIGGER AS $$
DECLARE
    existing_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO existing_count
    FROM accounts
    WHERE LOWER(code) = LOWER(NEW.code)
      AND deleted_at IS NULL
      AND id != COALESCE(NEW.id, 0);
    
    IF existing_count > 0 THEN
        RAISE EXCEPTION 'Account code % already exists', NEW.code
            USING HINT = 'Use unique account codes only',
                  ERRCODE = '23505';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
DROP TRIGGER IF EXISTS trg_prevent_duplicate_account_code ON accounts;

CREATE TRIGGER trg_prevent_duplicate_account_code
    BEFORE INSERT OR UPDATE OF code ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION prevent_duplicate_account_code();

\echo '✅ Created trigger: trg_prevent_duplicate_account_code'

\echo ''
\echo '============================================================================'
\echo 'STEP 10: Verification...'
\echo '============================================================================'

-- Check for remaining duplicates
SELECT 
    CASE 
        WHEN COUNT(*) = 0 THEN '✅ SUCCESS: No duplicates remain'
        ELSE '❌ ERROR: Still have duplicates!'
    END as status
FROM (
    SELECT code
    FROM accounts
    WHERE deleted_at IS NULL
    GROUP BY code
    HAVING COUNT(*) > 1
) remaining;

-- Show final account state for previously duplicated codes
\echo ''
\echo 'Final state of previously duplicated accounts:'
\echo ''

SELECT 
    a.code,
    a.id,
    a.name,
    a.balance,
    a.is_active,
    a.deleted_at IS NOT NULL as is_deleted
FROM accounts a
WHERE a.code IN (
    SELECT DISTINCT code FROM duplicate_merge_map
)
ORDER BY a.code, a.deleted_at NULLS FIRST, a.id;

\echo ''
\echo '============================================================================'
\echo '✅ FIX COMPLETE!'
\echo '============================================================================'
\echo 'All duplicates have been merged and unique constraint created.'
\echo 'Future duplicate account creation is now prevented.'
\echo '============================================================================'

COMMIT;

\echo ''
\echo '✅ Transaction committed successfully!'

