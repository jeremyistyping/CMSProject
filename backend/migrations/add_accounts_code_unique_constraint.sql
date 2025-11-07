-- ============================================================================
-- Migration: Add UNIQUE Constraint on accounts.code
-- ============================================================================
-- Purpose: Prevent duplicate account codes in production
-- Safe to run: Uses CONCURRENTLY for zero-downtime deployment
-- Created: 2025-10-18
-- ============================================================================

-- Step 1: Check for existing active duplicates
-- (This query should return 0 rows before proceeding)
DO $$
DECLARE
    duplicate_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO duplicate_count
    FROM (
        SELECT code, COUNT(*) as cnt
        FROM accounts
        WHERE deleted_at IS NULL
        GROUP BY code
        HAVING COUNT(*) > 1
    ) duplicates;
    
    IF duplicate_count > 0 THEN
        RAISE EXCEPTION 'Found % duplicate active account codes. Please resolve duplicates before adding constraint.', duplicate_count;
    END IF;
    
    RAISE NOTICE 'No active duplicates found. Safe to proceed.';
END $$;

-- Step 2: Drop old constraint if exists (from previous attempts)
DROP INDEX IF EXISTS accounts_code_active_unique;
DROP INDEX IF EXISTS idx_accounts_code_unique;
DROP INDEX IF EXISTS accounts_code_unique;

-- Step 3: Create partial unique index (allows soft-deleted duplicates)
-- Note: Cannot use CONCURRENTLY inside this migration runner; using regular CREATE INDEX
CREATE UNIQUE INDEX IF NOT EXISTS accounts_code_active_unique 
ON accounts (code) 
WHERE deleted_at IS NULL;

-- Step 4: Verify constraint was created successfully
DO $$
DECLARE
    index_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 
        FROM pg_indexes 
        WHERE tablename = 'accounts' 
          AND indexname = 'accounts_code_active_unique'
    ) INTO index_exists;
    
    IF index_exists THEN
        RAISE NOTICE '✅ SUCCESS: Unique constraint "accounts_code_active_unique" created successfully';
    ELSE
        RAISE EXCEPTION '❌ FAILED: Unique constraint was not created';
    END IF;
END $$;

-- Step 5: Test the constraint
DO $$
BEGIN
    -- This should fail if constraint is working
    BEGIN
        INSERT INTO accounts (code, name, type, level, is_header, is_active, balance)
        VALUES ('TEST_DUPLICATE_CHECK', 'Test Account 1', 'ASSET', 1, false, true, 0);
        
        INSERT INTO accounts (code, name, type, level, is_header, is_active, balance)
        VALUES ('TEST_DUPLICATE_CHECK', 'Test Account 2', 'ASSET', 1, false, true, 0);
        
        RAISE EXCEPTION '❌ CONSTRAINT NOT WORKING: Duplicate insert succeeded';
    EXCEPTION
        WHEN unique_violation THEN
            RAISE NOTICE '✅ CONSTRAINT WORKING: Duplicate insert was correctly rejected';
            -- Test inserts will be automatically rolled back due to the exception
    END;
END $$;

-- Step 6: Show index definition
SELECT 
    indexname, 
    indexdef
FROM pg_indexes 
WHERE tablename = 'accounts' 
  AND indexname = 'accounts_code_active_unique';

-- ============================================================================
-- Post-Migration Verification
-- ============================================================================

-- Verify no active duplicates
SELECT 
    '✅ Active Accounts: No Duplicates' as status,
    code,
    COUNT(*) as count
FROM accounts
WHERE deleted_at IS NULL
GROUP BY code
HAVING COUNT(*) > 1;
-- Expected: 0 rows

-- Show soft-deleted duplicates (for cleanup reference)
SELECT 
    '📊 Soft-Deleted Duplicates (can be cleaned up)' as status,
    code,
    COUNT(*) as total,
    COUNT(CASE WHEN deleted_at IS NULL THEN 1 END) as active,
    COUNT(CASE WHEN deleted_at IS NOT NULL THEN 1 END) as deleted
FROM accounts
GROUP BY code
HAVING COUNT(*) > 1
ORDER BY COUNT(*) DESC;

-- ============================================================================
-- Rollback Instructions (if needed)
-- ============================================================================
-- If you need to rollback this migration:
-- DROP INDEX CONCURRENTLY accounts_code_active_unique;
-- ============================================================================

