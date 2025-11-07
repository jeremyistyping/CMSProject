-- ================================================================
-- Migration: Add validation to prevent GL sharing
-- ================================================================
-- This ensures each cash/bank account has unique GL account
-- Prevents the root cause bug we just fixed
-- ================================================================

BEGIN;

-- Add unique constraint to prevent multiple cash/banks sharing same GL
-- This enforces 1:1 mapping between cash_banks and GL accounts
-- Note: PostgreSQL doesn't support WHERE in UNIQUE constraint, use partial index instead
CREATE UNIQUE INDEX IF NOT EXISTS unique_cash_bank_account_id 
ON cash_banks(account_id) 
WHERE deleted_at IS NULL;

-- Add index for better performance on account_id lookups
CREATE INDEX IF NOT EXISTS idx_cash_banks_account_id 
ON cash_banks(account_id) 
WHERE deleted_at IS NULL AND is_active = true;

-- Add comment for documentation
COMMENT ON INDEX unique_cash_bank_account_id IS 
'Ensures each cash/bank account has its own unique GL account for proper balance tracking';

-- Log migration
INSERT INTO migration_log (version, description, applied_at)
VALUES (1, 'Add validation to prevent GL sharing between cash/bank accounts', NOW())
ON CONFLICT DO NOTHING;

COMMIT;

-- ================================================================
-- What this fixes:
-- - Prevents BANK BCA and BANK ABC from sharing GL 1102
-- - Each bank gets unique GL: 1102-001, 1102-002, etc.
-- - Enforces design decision at database level
-- ================================================================
