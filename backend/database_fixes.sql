-- Database Performance & Fix Script
-- Run this with: psql -d your_database_name -f database_fixes.sql

\echo 'Starting database fixes...'

-- 1. Create missing account 1200 (ACCOUNTS RECEIVABLE)
\echo 'Adding missing account 1200...'
INSERT INTO accounts (code, name, type, category, level, is_header, is_active, balance, created_at, updated_at)
SELECT '1200', 'ACCOUNTS RECEIVABLE', 'ASSET', 'CURRENT_ASSET', 2, true, true, 0, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM accounts WHERE code = '1200');

-- 2. Update account 1201 to be child of 1200
\echo 'Updating account hierarchy...'
UPDATE accounts 
SET parent_id = (SELECT id FROM accounts WHERE code = '1200' AND deleted_at IS NULL)
WHERE code = '1201' 
AND parent_id != (SELECT id FROM accounts WHERE code = '1200' AND deleted_at IS NULL);

-- 3. Create performance indexes
\echo 'Creating performance indexes...'

-- Blacklisted tokens indexes
CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_token ON blacklisted_tokens(token);
CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_expires_at ON blacklisted_tokens(expires_at);

-- Notifications indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user_type ON notifications(user_id, type);

-- Sales and business data indexes
CREATE INDEX IF NOT EXISTS idx_sales_date ON sales(date);
CREATE INDEX IF NOT EXISTS idx_sales_customer_date ON sales(customer_id, date);
CREATE INDEX IF NOT EXISTS idx_purchases_date ON purchases(date);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Transaction indexes
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_transactions_account_date ON transactions(account_id, transaction_date);

\echo 'Indexes created successfully!'

-- 4. Check for orphaned records
\echo 'Checking for orphaned records...'

-- Orphaned sale items
SELECT 'Orphaned sale items: ' || COUNT(*) as result
FROM sale_items si 
LEFT JOIN sales s ON si.sale_id = s.id 
WHERE s.id IS NULL;

-- Orphaned sale payments
SELECT 'Orphaned sale payments: ' || COUNT(*) as result
FROM sale_payments sp 
LEFT JOIN sales s ON sp.sale_id = s.id 
WHERE s.id IS NULL;

-- 5. Update database statistics for better performance
\echo 'Updating database statistics...'
ANALYZE;

-- 6. Verify fixes
\echo 'Verifying fixes...'

-- Check account 1200
SELECT 'Account 1200 exists: ' || CASE WHEN COUNT(*) > 0 THEN 'YES' ELSE 'NO' END as result
FROM accounts WHERE code = '1200' AND deleted_at IS NULL;

-- Check account hierarchy
SELECT 'Account hierarchy:' as result;
SELECT '  ' || a.code || ' - ' || a.name || ' (Level: ' || a.level || ', Header: ' || a.is_header || ')' as result
FROM accounts a 
WHERE code IN ('1100', '1200', '1201', '1104')
ORDER BY code;

-- Check index count
SELECT 'Performance indexes created: ' || COUNT(*) as result
FROM pg_indexes 
WHERE indexname LIKE 'idx_blacklisted_%' 
   OR indexname LIKE 'idx_notifications_%'
   OR indexname LIKE 'idx_sales_%'
   OR indexname LIKE 'idx_transactions_%'
   OR indexname LIKE 'idx_audit_logs_%';

\echo 'Database fixes completed!'
\echo 'Note: Security models (security_incidents, etc.) will be created automatically when you restart the application.'
