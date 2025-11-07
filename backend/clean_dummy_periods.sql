-- Script to clean dummy accounting_periods data
-- WARNING: This will delete ALL existing accounting periods

-- First, let's see what we have
SELECT 
    COUNT(*) as total_periods,
    MIN(start_date) as earliest_period,
    MAX(end_date) as latest_period,
    SUM(CASE WHEN is_closed THEN 1 ELSE 0 END) as closed_count
FROM accounting_periods;

-- Show all periods before deletion
SELECT 
    id,
    start_date::date as start_date,
    end_date::date as end_date,
    is_closed,
    is_locked,
    created_at::date as created
FROM accounting_periods
ORDER BY end_date DESC;

-- Uncomment the lines below to DELETE all dummy data
-- WARNING: This action cannot be undone!

BEGIN;

-- Delete all accounting periods
DELETE FROM accounting_periods;

-- Reset sequence (optional - makes IDs start from 1 again)
ALTER SEQUENCE accounting_periods_id_seq RESTART WITH 1;

-- Verify deletion
SELECT COUNT(*) as remaining_records FROM accounting_periods;

COMMIT;

-- After running this, the accounting_periods table will be empty
-- and ready for real period closing data
