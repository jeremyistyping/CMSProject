-- ================================================================
-- Rollback: Remove GL sharing validation
-- ================================================================
-- WARNING: Rolling back this migration allows GL sharing again
-- Only use if absolutely necessary
-- ================================================================

BEGIN;

-- Remove unique constraint
ALTER TABLE cash_banks 
DROP CONSTRAINT IF EXISTS unique_cash_bank_account_id;

-- Remove index
DROP INDEX IF EXISTS idx_cash_banks_account_id;

-- Remove log entry
DELETE FROM migration_log WHERE version = 1;

COMMIT;

-- ================================================================
-- After rollback:
-- - Multiple cash/banks CAN share same GL again (not recommended!)
-- - You may need to manually separate shared GLs
-- ================================================================
