-- =====================================================
-- Reset Migration Log for 037
-- =====================================================
-- This script resets the migration log for 037 so it can run again
-- Run this manually if migration 037 failed and needs to be retried

-- Check current status
SELECT migration_name, status, message, executed_at 
FROM migration_logs 
WHERE migration_name LIKE '%037%'
ORDER BY executed_at DESC;

-- Delete failed migration entries for 037
DELETE FROM migration_logs 
WHERE migration_name IN (
    '037_add_invoice_types_system.sql',
    '037_rollback_invoice_types_system.sql'
);

-- Verify deletion
SELECT 'Migration log entries for 037 have been reset' as status;
SELECT COUNT(*) as remaining_037_entries 
FROM migration_logs 
WHERE migration_name LIKE '%037%';