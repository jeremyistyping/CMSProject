-- Debug Sales Summary Report Issue
-- Check if there's any sales data in the database

-- 1. Check total count of sales records
SELECT 
    'Total Sales Records' as check_type,
    COUNT(*) as count
FROM sales;

-- 2. Check sales by date range (around September 2025)
SELECT 
    'Sales in 2025' as check_type,
    COUNT(*) as count,
    SUM(total_amount) as total_amount
FROM sales 
WHERE date >= '2025-01-01' AND date <= '2025-12-31';

-- 3. Check sales specifically in September 2025
SELECT 
    'Sales in Sep 2025' as check_type,
    COUNT(*) as count,
    SUM(total_amount) as total_amount
FROM sales 
WHERE date >= '2025-09-01' AND date <= '2025-09-30';

-- 4. Check recent sales records with details
SELECT 
    id,
    code,
    date,
    total_amount,
    status,
    customer_id,
    created_at
FROM sales 
ORDER BY created_at DESC 
LIMIT 10;

-- 5. Check sales date distribution
SELECT 
    DATE_FORMAT(date, '%Y-%m') as month,
    COUNT(*) as sales_count,
    SUM(total_amount) as total_amount
FROM sales
GROUP BY DATE_FORMAT(date, '%Y-%m')
ORDER BY month DESC
LIMIT 12;

-- 6. Check sales table structure
DESCRIBE sales;