-- ==========================================
-- IMMEDIATE CLEANUP DUPLICATE ACCOUNTS
-- ==========================================

-- Step 1: Show current duplicates
SELECT 
    'BEFORE CLEANUP - Duplicate Accounts' as status,
    code,
    COUNT(*) as count,
    STRING_AGG(id::text || ':' || name, ' | ') as accounts
FROM accounts
WHERE deleted_at IS NULL
GROUP BY code
HAVING COUNT(*) > 1
ORDER BY code;

-- Step 2: Merge account 4101 duplicates
DO $$
DECLARE
    primary_id INTEGER;
    duplicate_count INTEGER;
BEGIN
    -- Get the account with actual balance (ID 22 based on logs)
    SELECT id INTO primary_id
    FROM accounts
    WHERE code = '4101' 
      AND deleted_at IS NULL
      AND balance > 0
    ORDER BY id
    LIMIT 1;
    
    IF primary_id IS NULL THEN
        -- If no account has balance, use the oldest one
        SELECT id INTO primary_id
        FROM accounts
        WHERE code = '4101' AND deleted_at IS NULL
        ORDER BY id
        LIMIT 1;
    END IF;
    
    RAISE NOTICE 'Primary account 4101 selected: ID = %', primary_id;
    
    -- Update unified journal lines to use primary account
    UPDATE unified_journal_lines
    SET account_id = primary_id
    WHERE account_id IN (
        SELECT id FROM accounts 
        WHERE code = '4101' 
          AND id != primary_id 
          AND deleted_at IS NULL
    );
    
    GET DIAGNOSTICS duplicate_count = ROW_COUNT;
    RAISE NOTICE 'Updated % unified journal lines', duplicate_count;
    
    -- Update legacy journal lines
    UPDATE journal_lines
    SET account_id = primary_id
    WHERE account_id IN (
        SELECT id FROM accounts 
        WHERE code = '4101' 
          AND id != primary_id 
          AND deleted_at IS NULL
    );
    
    GET DIAGNOSTICS duplicate_count = ROW_COUNT;
    RAISE NOTICE 'Updated % legacy journal lines', duplicate_count;
    
    -- Soft delete duplicate accounts
    UPDATE accounts
    SET deleted_at = NOW(),
        name = name || ' (MERGED)',
        is_active = false
    WHERE code = '4101' 
      AND id != primary_id 
      AND deleted_at IS NULL;
    
    GET DIAGNOSTICS duplicate_count = ROW_COUNT;
    RAISE NOTICE 'Soft deleted % duplicate accounts', duplicate_count;
    
    -- Standardize primary account name
    UPDATE accounts
    SET name = 'Pendapatan Penjualan'
    WHERE id = primary_id;
    
    RAISE NOTICE '✅ Account 4101 cleanup completed';
END $$;

-- Step 3: Cleanup other duplicate accounts (4201, 5101, 5201, 5202, 5203, 5204, 5900)
DO $$
DECLARE
    rec RECORD;
    primary_id INTEGER;
    merged_count INTEGER;
BEGIN
    FOR rec IN 
        SELECT code, COUNT(*) as dup_count
        FROM accounts
        WHERE deleted_at IS NULL
          AND is_header = false
        GROUP BY code
        HAVING COUNT(*) > 1
    LOOP
        -- Get primary account (with balance if exists, otherwise oldest)
        SELECT id INTO primary_id
        FROM accounts
        WHERE code = rec.code 
          AND deleted_at IS NULL
        ORDER BY 
            CASE WHEN balance != 0 THEN 0 ELSE 1 END,
            id
        LIMIT 1;
        
        -- Update unified journal lines
        UPDATE unified_journal_lines
        SET account_id = primary_id
        WHERE account_id IN (
            SELECT id FROM accounts 
            WHERE code = rec.code 
              AND id != primary_id 
              AND deleted_at IS NULL
        );
        
        -- Update legacy journal lines
        UPDATE journal_lines
        SET account_id = primary_id
        WHERE account_id IN (
            SELECT id FROM accounts 
            WHERE code = rec.code 
              AND id != primary_id 
              AND deleted_at IS NULL
        );
        
        -- Soft delete duplicates
        UPDATE accounts
        SET deleted_at = NOW(),
            name = name || ' (MERGED)',
            is_active = false
        WHERE code = rec.code 
          AND id != primary_id 
          AND deleted_at IS NULL;
        
        GET DIAGNOSTICS merged_count = ROW_COUNT;
        RAISE NOTICE 'Merged % duplicate(s) for account code %', merged_count, rec.code;
    END LOOP;
END $$;

-- Step 4: Verify cleanup
SELECT 
    '✅ AFTER CLEANUP - No Duplicates' as status,
    code,
    COUNT(*) as count
FROM accounts
WHERE deleted_at IS NULL
GROUP BY code
HAVING COUNT(*) > 1;

-- Step 5: Show final statistics
SELECT 
    'Final Statistics' as report,
    COUNT(*) as total_active_accounts,
    COUNT(DISTINCT code) as unique_codes,
    COUNT(*) - COUNT(DISTINCT code) as remaining_duplicates
FROM accounts
WHERE deleted_at IS NULL;

-- Step 6: Show account 4101 final state
SELECT 
    'Account 4101 Final State' as report,
    id,
    code,
    name,
    balance,
    is_active,
    deleted_at
FROM accounts
WHERE code = '4101'
ORDER BY id;

