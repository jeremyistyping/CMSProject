-- Check if accounting_periods table exists and has data
SELECT 
    'Table accounting_periods' as info,
    COUNT(*) as record_count
FROM accounting_periods;

-- Show all closed periods
SELECT 
    id,
    start_date,
    end_date,
    description,
    is_closed,
    is_locked,
    total_revenue,
    total_expense,
    net_income,
    closed_at,
    created_at
FROM accounting_periods
WHERE is_closed = true
ORDER BY end_date DESC
LIMIT 10;

-- Check if table exists
SELECT EXISTS (
    SELECT FROM information_schema.tables 
    WHERE table_name = 'accounting_periods'
) as table_exists;
