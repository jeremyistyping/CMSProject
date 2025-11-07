-- ================================================================
-- Rollback: Drop migration log table
-- ================================================================

BEGIN;

DROP TABLE IF EXISTS migration_log CASCADE;

COMMIT;
