-- Fix sales performance index that uses CURRENT_DATE (not immutable)
-- Replace with a version without CURRENT_DATE in WHERE clause

DO $$
BEGIN
    -- Drop the problematic index if it exists
    DROP INDEX IF EXISTS idx_sales_recent_activity;
    
    RAISE NOTICE '🔧 Creating sales performance indices without CURRENT_DATE';
    
    -- Create index without date filter (more flexible)
    CREATE INDEX IF NOT EXISTS idx_sales_date_status_customer 
    ON sales(date DESC, status, customer_id, total_amount DESC) 
    WHERE deleted_at IS NULL;
    
    RAISE NOTICE '✅ Created idx_sales_date_status_customer index';
    
    -- Mark migration as success if it was failed
    UPDATE migration_logs 
    SET status = 'SUCCESS', 
        message = 'Fixed by replacing CURRENT_DATE with static index',
        updated_at = NOW()
    WHERE migration_name = '021_add_sales_performance_indices.sql'
      AND status = 'FAILED';
    
    RAISE NOTICE '🎯 Sales performance index fix completed';
END $$;
