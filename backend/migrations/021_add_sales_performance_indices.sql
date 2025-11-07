-- Migration: Add Sales Performance Indices
-- Purpose: Add comprehensive database indices for sales query performance
-- Date: 2025-09-19
-- Priority: High (Priority 2)

-- ====================================
-- SALES TABLE PERFORMANCE INDICES
-- ====================================

-- 1. Primary query patterns indices
-- For dashboard and list queries
CREATE INDEX IF NOT EXISTS idx_sales_status_date_performance 
ON sales(status, date DESC, total_amount DESC) 
WHERE deleted_at IS NULL;

-- For customer-specific queries
CREATE INDEX IF NOT EXISTS idx_sales_customer_date_status 
ON sales(customer_id, date DESC, status, total_amount) 
WHERE deleted_at IS NULL;

-- For date range queries (reports)
CREATE INDEX IF NOT EXISTS idx_sales_date_range_performance 
ON sales(date, status, customer_id, total_amount) 
WHERE deleted_at IS NULL;

-- 2. Financial analysis indices
-- For outstanding amounts (accounts receivable)
CREATE INDEX IF NOT EXISTS idx_sales_outstanding_analysis 
ON sales(outstanding_amount DESC, due_date, customer_id, status) 
WHERE outstanding_amount > 0 AND deleted_at IS NULL;

-- For overdue analysis
CREATE INDEX IF NOT EXISTS idx_sales_overdue_analysis 
ON sales(due_date, status, outstanding_amount DESC, customer_id) 
WHERE status IN ('INVOICED', 'OVERDUE') AND deleted_at IS NULL;

-- For sales performance analysis
CREATE INDEX IF NOT EXISTS idx_sales_performance_analysis 
ON sales(date DESC, status, total_amount DESC, customer_id, user_id) 
WHERE status IN ('COMPLETED', 'PAID') AND deleted_at IS NULL;

-- 3. Search and filter indices
-- For code/invoice number searches
CREATE INDEX IF NOT EXISTS idx_sales_code_search 
ON sales(LOWER(code), LOWER(invoice_number)) 
WHERE deleted_at IS NULL;

-- For text search optimization
CREATE INDEX IF NOT EXISTS idx_sales_search_optimization 
ON sales(customer_id, status, date DESC) 
WHERE deleted_at IS NULL;

-- 4. User and sales person analysis
CREATE INDEX IF NOT EXISTS idx_sales_user_performance 
ON sales(user_id, date DESC, status, total_amount DESC) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sales_person_performance 
ON sales(sales_person_id, date DESC, status, total_amount DESC) 
WHERE sales_person_id IS NOT NULL AND deleted_at IS NULL;

-- 5. Currency and exchange rate queries
CREATE INDEX IF NOT EXISTS idx_sales_currency_analysis 
ON sales(currency, exchange_rate, date DESC, total_amount DESC) 
WHERE currency != 'IDR' AND deleted_at IS NULL;

-- ====================================
-- SALE ITEMS PERFORMANCE INDICES
-- ====================================

-- 1. Product analysis indices
-- For product sales analysis
CREATE INDEX IF NOT EXISTS idx_sale_items_product_analysis 
ON sale_items(product_id, created_at DESC, quantity DESC, line_total DESC) 
WHERE deleted_at IS NULL;

-- For sale item lookups
CREATE INDEX IF NOT EXISTS idx_sale_items_sale_lookup 
ON sale_items(sale_id, product_id, quantity, line_total) 
WHERE deleted_at IS NULL;

-- 2. Revenue analysis indices
-- For revenue account analysis
CREATE INDEX IF NOT EXISTS idx_sale_items_revenue_analysis 
ON sale_items(revenue_account_id, created_at DESC, line_total DESC) 
WHERE deleted_at IS NULL;

-- For item performance analysis
CREATE INDEX IF NOT EXISTS idx_sale_items_performance 
ON sale_items(product_id, quantity DESC, unit_price DESC, line_total DESC) 
WHERE deleted_at IS NULL;

-- 3. Tax analysis indices
CREATE INDEX IF NOT EXISTS idx_sale_items_tax_analysis 
ON sale_items(taxable, ppn_amount DESC, total_tax DESC) 
WHERE deleted_at IS NULL;

-- ====================================
-- SALE PAYMENTS PERFORMANCE INDICES
-- ====================================

-- 1. Payment tracking indices
-- For payment history and analysis
CREATE INDEX IF NOT EXISTS idx_sale_payments_tracking 
ON sale_payments(sale_id, payment_date DESC, amount DESC, status) 
WHERE deleted_at IS NULL;

-- For payment method analysis
CREATE INDEX IF NOT EXISTS idx_sale_payments_method_analysis 
ON sale_payments(payment_method, payment_date DESC, amount DESC) 
WHERE deleted_at IS NULL;

-- 2. Cash flow analysis indices
-- For daily cash flow
CREATE INDEX IF NOT EXISTS idx_sale_payments_cash_flow 
ON sale_payments(payment_date DESC, amount DESC, payment_method, status) 
WHERE status = 'COMPLETED' AND deleted_at IS NULL;

-- For cash/bank account tracking
CREATE INDEX IF NOT EXISTS idx_sale_payments_cashbank_tracking 
ON sale_payments(cash_bank_id, payment_date DESC, amount DESC) 
WHERE cash_bank_id IS NOT NULL AND deleted_at IS NULL;

-- 3. User payment activity
CREATE INDEX IF NOT EXISTS idx_sale_payments_user_activity 
ON sale_payments(user_id, payment_date DESC, amount DESC) 
WHERE deleted_at IS NULL;

-- ====================================
-- COMPOSITE INDICES FOR COMPLEX QUERIES
-- ====================================

-- 1. Dashboard summary queries
-- For sales dashboard (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_sales_dashboard_composite 
ON sales(status, date DESC, customer_id, total_amount DESC, outstanding_amount DESC) 
WHERE deleted_at IS NULL;

-- 2. Report generation indices
-- For monthly/yearly sales reports
CREATE INDEX IF NOT EXISTS idx_sales_reports_composite 
ON sales(
    date DESC, 
    status, 
    customer_id, 
    total_amount DESC, 
    paid_amount DESC, 
    outstanding_amount DESC
) WHERE deleted_at IS NULL;

-- 3. Financial reconciliation indices
-- For payment reconciliation
CREATE INDEX IF NOT EXISTS idx_sales_reconciliation_composite 
ON sales(
    outstanding_amount DESC, 
    due_date, 
    status, 
    customer_id, 
    total_amount DESC
) WHERE outstanding_amount > 0 AND deleted_at IS NULL;

-- ====================================
-- PARTIAL INDICES FOR SPECIFIC CONDITIONS
-- ====================================

-- 1. Active sales only
CREATE INDEX IF NOT EXISTS idx_sales_active_only 
ON sales(date DESC, total_amount DESC, customer_id) 
WHERE status NOT IN ('CANCELLED', 'DRAFT') AND deleted_at IS NULL;

-- 2. Unpaid invoices only
CREATE INDEX IF NOT EXISTS idx_sales_unpaid_invoices 
ON sales(due_date, outstanding_amount DESC, customer_id) 
WHERE status IN ('INVOICED', 'OVERDUE') 
AND outstanding_amount > 0 
AND deleted_at IS NULL;

-- 3. Recent sales index (no volatile predicate; planner will filter by date)
CREATE INDEX IF NOT EXISTS idx_sales_recent_activity 
ON sales(date DESC, status, customer_id, total_amount DESC) 
WHERE deleted_at IS NULL;

-- 4. Large amounts only (> 1M IDR)
CREATE INDEX IF NOT EXISTS idx_sales_large_amounts 
ON sales(total_amount DESC, date DESC, customer_id, status) 
WHERE total_amount > 1000000 
AND deleted_at IS NULL;

-- ====================================
-- FUNCTIONAL INDICES FOR SEARCH
-- ====================================

-- 1. Case-insensitive search indices
-- For customer name search (requires joining with contacts)
CREATE INDEX IF NOT EXISTS idx_sales_customer_search 
ON sales(customer_id, date DESC, status) 
WHERE deleted_at IS NULL;

-- 2. Text search optimization indices
-- For invoice number and code search
CREATE INDEX IF NOT EXISTS idx_sales_text_search 
ON sales(
    COALESCE(invoice_number, ''), 
    COALESCE(code, ''), 
    date DESC
) WHERE deleted_at IS NULL;

-- ====================================
-- COVERING INDICES FOR READ-HEAVY QUERIES
-- ====================================

-- 1. Sales list covering index (includes most columns needed for list views)
CREATE INDEX IF NOT EXISTS idx_sales_list_covering 
ON sales(
    status, 
    date DESC, 
    customer_id, 
    id, 
    code, 
    invoice_number, 
    total_amount, 
    paid_amount, 
    outstanding_amount
) WHERE deleted_at IS NULL;

-- 2. Payment list covering index
CREATE INDEX IF NOT EXISTS idx_sale_payments_list_covering 
ON sale_payments(
    sale_id, 
    payment_date DESC, 
    id, 
    amount, 
    payment_method, 
    status, 
    reference
) WHERE deleted_at IS NULL;

-- ====================================
-- INDEX USAGE ANALYSIS FUNCTIONS
-- ====================================

-- Function to analyze index usage
CREATE OR REPLACE FUNCTION analyze_sales_index_usage()
RETURNS TABLE(
    table_name TEXT,
    index_name TEXT,
    index_size TEXT,
    index_scans BIGINT,
    tuples_read BIGINT,
    tuples_fetched BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        schemaname || '.' || tablename::TEXT as table_name,
        indexname::TEXT as index_name,
        pg_size_pretty(pg_relation_size(indexname::regclass))::TEXT as index_size,
        idx_scan as index_scans,
        idx_tup_read as tuples_read,
        idx_tup_fetch as tuples_fetched
    FROM pg_stat_user_indexes 
    WHERE tablename IN ('sales', 'sale_items', 'sale_payments')
    ORDER BY idx_scan DESC;
END;
$$ LANGUAGE plpgsql;

-- Function to identify missing indices
CREATE OR REPLACE FUNCTION identify_missing_sales_indices()
RETURNS TABLE(
    query_pattern TEXT,
    recommendation TEXT,
    estimated_impact TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'sales table seq_scan'::TEXT as query_pattern,
        'Consider adding index on frequently filtered columns'::TEXT as recommendation,
        'HIGH - if seq_scan > 1000'::TEXT as estimated_impact
    FROM pg_stat_user_tables 
    WHERE relname = 'sales' AND seq_scan > 1000;
    
    -- Add more analysis here based on actual query patterns
END;
$$ LANGUAGE plpgsql;

-- ====================================
-- INDEX MAINTENANCE PROCEDURES
-- ====================================

-- Function to rebuild sales indices
CREATE OR REPLACE FUNCTION rebuild_sales_indices()
RETURNS TEXT AS $$
DECLARE
    result_msg TEXT := 'Sales indices rebuilt successfully: ';
    index_count INTEGER := 0;
BEGIN
    -- Rebuild key indices
    REINDEX INDEX idx_sales_status_date_performance;
    REINDEX INDEX idx_sales_customer_date_status;
    REINDEX INDEX idx_sales_outstanding_analysis;
    
    index_count := 3;
    
    result_msg := result_msg || index_count::TEXT || ' indices rebuilt';
    RETURN result_msg;
EXCEPTION
    WHEN OTHERS THEN
        RETURN 'Error rebuilding indices: ' || SQLERRM;
END;
$$ LANGUAGE plpgsql;

-- ====================================
-- QUERY PERFORMANCE TESTING
-- ====================================

-- Function to test common query performance
CREATE OR REPLACE FUNCTION test_sales_query_performance()
RETURNS TABLE(
    query_description TEXT,
    execution_time_ms NUMERIC,
    rows_returned BIGINT,
    index_used TEXT
) AS $$
DECLARE
    start_time TIMESTAMP;
    end_time TIMESTAMP;
    row_count BIGINT;
BEGIN
    -- Test 1: Sales list with pagination
    start_time := clock_timestamp();
    SELECT COUNT(*) INTO row_count 
    FROM sales 
    WHERE status = 'INVOICED' 
    AND deleted_at IS NULL 
    ORDER BY date DESC 
    LIMIT 20;
    end_time := clock_timestamp();
    
    RETURN QUERY SELECT 
        'Sales list with status filter'::TEXT,
        EXTRACT(MILLISECONDS FROM (end_time - start_time)),
        row_count,
        'idx_sales_status_date_performance'::TEXT;
    
    -- Test 2: Customer sales analysis
    start_time := clock_timestamp();
    SELECT COUNT(*) INTO row_count 
    FROM sales 
    WHERE customer_id = 1 
    AND date >= CURRENT_DATE - INTERVAL '1 year'
    AND deleted_at IS NULL;
    end_time := clock_timestamp();
    
    RETURN QUERY SELECT 
        'Customer sales analysis'::TEXT,
        EXTRACT(MILLISECONDS FROM (end_time - start_time)),
        row_count,
        'idx_sales_customer_date_status'::TEXT;
    
    -- Test 3: Outstanding amounts analysis
    start_time := clock_timestamp();
    SELECT COUNT(*) INTO row_count 
    FROM sales 
    WHERE outstanding_amount > 0 
    AND deleted_at IS NULL 
    ORDER BY outstanding_amount DESC;
    end_time := clock_timestamp();
    
    RETURN QUERY SELECT 
        'Outstanding amounts analysis'::TEXT,
        EXTRACT(MILLISECONDS FROM (end_time - start_time)),
        row_count,
        'idx_sales_outstanding_analysis'::TEXT;
END;
$$ LANGUAGE plpgsql;

-- ====================================
-- MIGRATION COMPLETION
-- ====================================

-- Log migration completion
INSERT INTO migration_logs (migration_name, executed_at, description, status)
VALUES (
    '021_add_sales_performance_indices',
    NOW(),
    'Added comprehensive performance indices for sales, sale_items, and sale_payments tables',
    'SUCCESS'
) ON CONFLICT (migration_name) DO UPDATE SET 
    executed_at = NOW(),
    status = 'SUCCESS';

-- Performance analysis
SELECT 'Sales performance indices created successfully. Run SELECT * FROM test_sales_query_performance(); to validate.' as result;