-- Performance optimization indexes for journal entries and related tables
-- Created: 2024-01-18
-- Purpose: Improve query performance for journal entry operations and reporting

-- Journal Entries Indexes
-- 1. Composite index for reference lookups (most common query pattern)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_entries_reference 
ON journal_entries(reference_type, reference_id) WHERE reference_type IS NOT NULL;

-- 2. Composite index for date and status filtering (used in reports)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_entries_date_status 
ON journal_entries(entry_date, status) WHERE deleted_at IS NULL;

-- 3. Index for user-based queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_entries_user_date 
ON journal_entries(user_id, entry_date DESC) WHERE deleted_at IS NULL;

-- 4. Index for posting operations
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_entries_posting 
ON journal_entries(status, posted_by, posting_date) 
WHERE status IN ('DRAFT', 'POSTED') AND deleted_at IS NULL;

-- 5. Index for balance calculations
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_entries_balance 
ON journal_entries(is_balanced, total_debit, total_credit) 
WHERE deleted_at IS NULL;

-- 6. Index for auto-generated entries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_entries_auto_generated 
ON journal_entries(is_auto_generated, reference_type, entry_date) 
WHERE deleted_at IS NULL;

-- Journal Lines Indexes
-- 7. Composite index for account and date (used in account balance calculations)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_lines_account_date 
ON journal_lines(account_id, created_at) WHERE deleted_at IS NULL;

-- 8. Index for journal entry lines lookup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_lines_entry_line 
ON journal_lines(journal_entry_id, line_number) WHERE deleted_at IS NULL;

-- 9. Index for amount-based queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_lines_amounts 
ON journal_lines(account_id, debit_amount, credit_amount) 
WHERE (debit_amount > 0 OR credit_amount > 0) AND deleted_at IS NULL;

-- Accounts Indexes (if not already exist)
-- 10. Composite index for account type and active status
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_type_active 
ON accounts(type, is_active) WHERE deleted_at IS NULL;

-- 11. Index for parent-child relationships
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_parent_level 
ON accounts(parent_id, level) WHERE deleted_at IS NULL;

-- 12. Index for account balance queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_balance 
ON accounts(type, balance) WHERE is_active = true AND deleted_at IS NULL;

-- Sales Related Indexes
-- 13. Index for sales reference in journal entries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sales_outstanding 
ON sales(outstanding_amount, status, due_date) 
WHERE outstanding_amount > 0 AND deleted_at IS NULL;

-- 14. Index for customer outstanding calculations
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sales_customer_outstanding 
ON sales(customer_id, outstanding_amount, status) 
WHERE outstanding_amount > 0 AND deleted_at IS NULL;

-- Purchase Related Indexes
-- 15. Index for purchase reference in journal entries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_purchases_outstanding 
ON purchases(outstanding_amount, status, due_date) 
WHERE outstanding_amount > 0 AND deleted_at IS NULL;

-- 16. Index for vendor outstanding calculations
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_purchases_vendor_outstanding 
ON purchases(vendor_id, outstanding_amount, status) 
WHERE outstanding_amount > 0 AND deleted_at IS NULL;

-- Cash Bank Indexes
-- 17. Index for cash bank account linking
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cashbank_account 
ON cash_banks(account_id) WHERE account_id IS NOT NULL AND deleted_at IS NULL;

-- 18. Index for cash bank transactions
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cashbank_transactions_date 
ON cash_bank_transactions(cash_bank_id, transaction_date DESC) 
WHERE deleted_at IS NULL;

-- Audit Logs Indexes
-- 19. Index for audit trail queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_table_record 
ON audit_logs(table_name, record_id, created_at DESC);

-- 20. Index for user audit trail
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_date 
ON audit_logs(user_id, created_at DESC) WHERE user_id IS NOT NULL;

-- Accounting Periods Indexes (for new period locking feature)
-- 21. Index for period lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounting_periods_year_month 
ON accounting_periods(year, month) WHERE deleted_at IS NULL;

-- 22. Index for period status
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounting_periods_status 
ON accounting_periods(is_closed, is_locked, year, month) 
WHERE deleted_at IS NULL;

-- Partial Indexes for Common Queries
-- 23. Draft journal entries only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_entries_draft_only 
ON journal_entries(user_id, created_at DESC) 
WHERE status = 'DRAFT' AND deleted_at IS NULL;

-- 24. Posted journal entries only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_entries_posted_only 
ON journal_entries(posted_by, posting_date DESC) 
WHERE status = 'POSTED' AND deleted_at IS NULL;

-- 25. Unbalanced entries for validation
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_entries_unbalanced 
ON journal_entries(id, code, total_debit, total_credit) 
WHERE is_balanced = false AND deleted_at IS NULL;

-- Covering Indexes for Common Report Queries
-- 26. Financial report covering index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_accounts_financial_report 
ON accounts(type, category, balance, is_active, is_header) 
WHERE deleted_at IS NULL;

-- 27. Journal lines report covering index
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_journal_lines_report 
ON journal_lines(journal_entry_id, account_id, debit_amount, credit_amount, line_number) 
WHERE deleted_at IS NULL;

-- Statistics Update (PostgreSQL specific)
-- Update table statistics for better query planning
ANALYZE journal_entries;
ANALYZE journal_lines;
ANALYZE accounts;
ANALYZE sales;
ANALYZE purchases;
ANALYZE cash_banks;
ANALYZE audit_logs;

-- Create maintenance script comment
COMMENT ON INDEX idx_journal_entries_reference IS 'Optimizes journal entry lookup by reference type and ID';
COMMENT ON INDEX idx_journal_entries_date_status IS 'Optimizes date range and status filtering for reports';
COMMENT ON INDEX idx_journal_lines_account_date IS 'Optimizes account balance calculations by date range';
COMMENT ON INDEX idx_accounts_type_active IS 'Optimizes account filtering by type and status';

-- Performance Monitoring Views
-- Create view for index usage monitoring
CREATE OR REPLACE VIEW v_journal_index_usage AS
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes 
WHERE tablename IN ('journal_entries', 'journal_lines', 'accounts')
ORDER BY idx_scan DESC;

-- Create view for table size monitoring
CREATE OR REPLACE VIEW v_journal_table_sizes AS
SELECT 
    tablename,
    pg_size_pretty(pg_total_relation_size(tablename::regclass)) as total_size,
    pg_size_pretty(pg_relation_size(tablename::regclass)) as table_size,
    pg_size_pretty(pg_total_relation_size(tablename::regclass) - pg_relation_size(tablename::regclass)) as index_size
FROM (VALUES 
    ('journal_entries'::text),
    ('journal_lines'::text),
    ('accounts'::text),
    ('sales'::text),
    ('purchases'::text)
) t(tablename);

COMMIT;