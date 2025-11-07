-- =====================================================
-- FINANCIAL REPORT ANALYSIS & VALIDATION SCRIPT
-- Script untuk analisis mendalam financial reports
-- dan validasi data journal entries
-- =====================================================

-- Informasi Database
SELECT 
    'DATABASE INFO' as section,
    current_database() as database_name,
    version() as postgres_version,
    now() as analysis_time;

-- =====================================================
-- 1. CHART OF ACCOUNTS ANALYSIS
-- =====================================================

-- Analisis struktur Chart of Accounts
SELECT 
    '==== CHART OF ACCOUNTS ANALYSIS ====' as analysis_section;

SELECT 
    'Account Distribution by Type' as metric,
    type as account_type,
    COUNT(*) as total_accounts,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_accounts,
    COUNT(CASE WHEN is_header = true THEN 1 END) as header_accounts,
    COUNT(CASE WHEN balance != 0 THEN 1 END) as accounts_with_balance,
    SUM(balance) as total_balance
FROM accounts 
WHERE deleted_at IS NULL
GROUP BY type
ORDER BY type;

-- Detail account balances by category
SELECT 
    'Account Balances by Category' as metric,
    type,
    category,
    COUNT(*) as account_count,
    SUM(balance) as total_balance,
    AVG(balance) as average_balance,
    MIN(balance) as min_balance,
    MAX(balance) as max_balance
FROM accounts 
WHERE deleted_at IS NULL AND is_active = true
GROUP BY type, category
ORDER BY type, category;

-- Account structure validation
SELECT 
    'Account Structure Issues' as metric,
    COUNT(CASE WHEN code IS NULL OR code = '' THEN 1 END) as missing_codes,
    COUNT(CASE WHEN name IS NULL OR name = '' THEN 1 END) as missing_names,
    COUNT(CASE WHEN type IS NULL OR type = '' THEN 1 END) as missing_types,
    COUNT(CASE WHEN LENGTH(code) < 3 THEN 1 END) as short_codes,
    COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_accounts
FROM accounts 
WHERE deleted_at IS NULL;

-- =====================================================
-- 2. JOURNAL ENTRIES ANALYSIS
-- =====================================================

SELECT 
    '==== JOURNAL ENTRIES ANALYSIS ====' as analysis_section;

-- Journal entries overview
SELECT 
    'Journal Entries Overview' as metric,
    COUNT(*) as total_entries,
    COUNT(CASE WHEN is_balanced = true THEN 1 END) as balanced_entries,
    COUNT(CASE WHEN is_balanced = false THEN 1 END) as unbalanced_entries,
    COUNT(CASE WHEN status = 'POSTED' THEN 1 END) as posted_entries,
    COUNT(CASE WHEN status = 'DRAFT' THEN 1 END) as draft_entries,
    SUM(total_debit) as total_debits,
    SUM(total_credit) as total_credits,
    SUM(total_debit) - SUM(total_credit) as debit_credit_difference
FROM journal_entries 
WHERE deleted_at IS NULL;

-- Journal entries by reference type
SELECT 
    'Entries by Reference Type' as metric,
    COALESCE(reference_type, 'NO_TYPE') as ref_type,
    COUNT(*) as entry_count,
    SUM(total_debit) as total_debits,
    SUM(total_credit) as total_credits,
    COUNT(CASE WHEN is_balanced = true THEN 1 END) as balanced_count,
    ROUND((COUNT(CASE WHEN is_balanced = true THEN 1 END)::decimal / COUNT(*)) * 100, 2) as balance_percentage
FROM journal_entries 
WHERE deleted_at IS NULL
GROUP BY reference_type
ORDER BY entry_count DESC;

-- Monthly journal entry activity
SELECT 
    'Monthly Journal Activity' as metric,
    DATE_TRUNC('month', entry_date) as month,
    COUNT(*) as entry_count,
    SUM(total_debit) as total_debits,
    SUM(total_credit) as total_credits,
    COUNT(CASE WHEN status = 'POSTED' THEN 1 END) as posted_count,
    COUNT(CASE WHEN is_balanced = true THEN 1 END) as balanced_count
FROM journal_entries 
WHERE deleted_at IS NULL
GROUP BY DATE_TRUNC('month', entry_date)
ORDER BY month DESC
LIMIT 12;

-- =====================================================
-- 3. TRIAL BALANCE VALIDATION
-- =====================================================

SELECT 
    '==== TRIAL BALANCE VALIDATION ====' as analysis_section;

-- Calculated trial balance from journal entries
WITH account_movements AS (
    SELECT 
        jl.account_id,
        a.code as account_code,
        a.name as account_name,
        a.type as account_type,
        SUM(jl.debit_amount) as total_debits,
        SUM(jl.credit_amount) as total_credits,
        SUM(jl.debit_amount) - SUM(jl.credit_amount) as net_movement
    FROM journal_lines jl
    JOIN accounts a ON jl.account_id = a.id
    JOIN journal_entries je ON jl.journal_entry_id = je.id
    WHERE je.status = 'POSTED' 
    AND je.deleted_at IS NULL 
    AND a.deleted_at IS NULL
    GROUP BY jl.account_id, a.code, a.name, a.type
),
trial_balance AS (
    SELECT 
        account_type,
        account_code,
        account_name,
        total_debits,
        total_credits,
        net_movement,
        CASE 
            WHEN account_type IN ('ASSET', 'EXPENSE') THEN net_movement
            ELSE 0
        END as debit_balance,
        CASE 
            WHEN account_type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN -net_movement
            ELSE 0
        END as credit_balance
    FROM account_movements
    WHERE ABS(net_movement) > 0.01  -- Filter out zero balances
)
SELECT 
    'Calculated Trial Balance' as metric,
    account_type,
    COUNT(*) as account_count,
    SUM(total_debits) as type_total_debits,
    SUM(total_credits) as type_total_credits,
    SUM(debit_balance) as type_debit_balance,
    SUM(credit_balance) as type_credit_balance,
    SUM(debit_balance) - SUM(credit_balance) as type_net_balance
FROM trial_balance
GROUP BY account_type
ORDER BY account_type;

-- Trial balance totals validation
SELECT 
    'Trial Balance Totals' as metric,
    SUM(debit_balance) as total_debit_balances,
    SUM(credit_balance) as total_credit_balances,
    SUM(debit_balance) - SUM(credit_balance) as difference,
    CASE 
        WHEN ABS(SUM(debit_balance) - SUM(credit_balance)) < 0.01 THEN 'BALANCED'
        ELSE 'NOT BALANCED'
    END as status
FROM (
    SELECT 
        CASE 
            WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                SUM(jl.debit_amount) - SUM(jl.credit_amount)
            ELSE 0
        END as debit_balance,
        CASE 
            WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 
                SUM(jl.credit_amount) - SUM(jl.debit_amount)
            ELSE 0
        END as credit_balance
    FROM journal_lines jl
    JOIN accounts a ON jl.account_id = a.id
    JOIN journal_entries je ON jl.journal_entry_id = je.id
    WHERE je.status = 'POSTED' 
    AND je.deleted_at IS NULL 
    AND a.deleted_at IS NULL
    GROUP BY jl.account_id, a.type
) tb;

-- =====================================================
-- 4. ACCOUNTING EQUATION VALIDATION
-- =====================================================

SELECT 
    '==== ACCOUNTING EQUATION VALIDATION ====' as analysis_section;

-- Basic accounting equation check using account balances
WITH equation_balances AS (
    SELECT 
        SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) as total_assets,
        SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) as total_liabilities,
        SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END) as total_equity,
        SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END) as total_revenue,
        SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END) as total_expenses
    FROM accounts 
    WHERE is_active = true AND deleted_at IS NULL
)
SELECT 
    'Accounting Equation Check' as metric,
    total_assets,
    total_liabilities,
    total_equity,
    total_revenue,
    total_expenses,
    total_revenue - total_expenses as retained_earnings,
    total_equity + (total_revenue - total_expenses) as adjusted_equity,
    total_liabilities + total_equity + (total_revenue - total_expenses) as liabilities_plus_equity,
    total_assets - (total_liabilities + total_equity + (total_revenue - total_expenses)) as difference,
    CASE 
        WHEN ABS(total_assets - (total_liabilities + total_equity + (total_revenue - total_expenses))) < 0.01 
        THEN 'BALANCED' 
        ELSE 'NOT BALANCED' 
    END as equation_status
FROM equation_balances;

-- =====================================================
-- 5. PROFIT & LOSS ANALYSIS
-- =====================================================

SELECT 
    '==== PROFIT & LOSS ANALYSIS ====' as analysis_section;

-- P&L calculation from account balances
WITH pl_data AS (
    SELECT 
        SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END) as total_revenue,
        SUM(CASE 
            WHEN type = 'EXPENSE' AND category IN ('COST_OF_GOODS_SOLD', 'DIRECT_MATERIAL', 'DIRECT_LABOR') 
            THEN balance ELSE 0 
        END) as cost_of_goods_sold,
        SUM(CASE 
            WHEN type = 'EXPENSE' AND category IN ('OPERATING_EXPENSE', 'ADMINISTRATIVE_EXPENSE', 'SELLING_EXPENSE')
            THEN balance ELSE 0 
        END) as operating_expenses,
        SUM(CASE 
            WHEN type = 'EXPENSE' AND category NOT IN ('COST_OF_GOODS_SOLD', 'DIRECT_MATERIAL', 'DIRECT_LABOR', 'OPERATING_EXPENSE', 'ADMINISTRATIVE_EXPENSE', 'SELLING_EXPENSE')
            THEN balance ELSE 0 
        END) as other_expenses
    FROM accounts 
    WHERE is_active = true AND deleted_at IS NULL
)
SELECT 
    'Profit & Loss Summary' as metric,
    total_revenue,
    cost_of_goods_sold,
    total_revenue - cost_of_goods_sold as gross_profit,
    CASE WHEN total_revenue > 0 
         THEN ROUND(((total_revenue - cost_of_goods_sold) / total_revenue) * 100, 2) 
         ELSE 0 
    END as gross_profit_margin,
    operating_expenses,
    total_revenue - cost_of_goods_sold - operating_expenses as operating_income,
    other_expenses,
    total_revenue - cost_of_goods_sold - operating_expenses - other_expenses as net_income,
    CASE WHEN total_revenue > 0 
         THEN ROUND(((total_revenue - cost_of_goods_sold - operating_expenses - other_expenses) / total_revenue) * 100, 2) 
         ELSE 0 
    END as net_profit_margin
FROM pl_data;

-- =====================================================
-- 6. CASH ACCOUNTS ANALYSIS
-- =====================================================

SELECT 
    '==== CASH ACCOUNTS ANALYSIS ====' as analysis_section;

-- Cash and cash equivalent accounts
SELECT 
    'Cash Accounts Summary' as metric,
    code as account_code,
    name as account_name,
    balance as current_balance,
    CASE 
        WHEN category IN ('CURRENT_ASSET') AND (name ILIKE '%cash%' OR name ILIKE '%kas%') THEN 'CASH'
        WHEN category IN ('CURRENT_ASSET') AND (name ILIKE '%bank%') THEN 'BANK'
        ELSE 'OTHER'
    END as cash_type
FROM accounts 
WHERE is_active = true 
AND deleted_at IS NULL
AND type = 'ASSET'
AND (category = 'CURRENT_ASSET' OR name ILIKE '%cash%' OR name ILIKE '%kas%' OR name ILIKE '%bank%')
ORDER BY balance DESC;

-- =====================================================
-- 7. DATA QUALITY ISSUES DETECTION
-- =====================================================

SELECT 
    '==== DATA QUALITY ISSUES ====' as analysis_section;

-- Unbalanced journal entries
SELECT 
    'Unbalanced Journal Entries' as issue_type,
    COUNT(*) as issue_count,
    string_agg(code, ', ') as sample_codes
FROM journal_entries 
WHERE is_balanced = false 
AND deleted_at IS NULL;

-- Duplicate journal entry codes
SELECT 
    'Duplicate Journal Codes' as issue_type,
    COUNT(*) as duplicate_code_groups,
    string_agg(code, ', ') as duplicate_codes
FROM (
    SELECT code
    FROM journal_entries 
    WHERE deleted_at IS NULL
    GROUP BY code 
    HAVING COUNT(*) > 1
) duplicates;

-- Journal entries without references
SELECT 
    'Entries Without References' as issue_type,
    COUNT(*) as issue_count
FROM journal_entries 
WHERE (reference IS NULL OR reference = '')
AND deleted_at IS NULL;

-- Future dated entries
SELECT 
    'Future Dated Entries' as issue_type,
    COUNT(*) as issue_count,
    MIN(entry_date) as earliest_future_date,
    MAX(entry_date) as latest_future_date
FROM journal_entries 
WHERE entry_date > CURRENT_DATE
AND deleted_at IS NULL;

-- Accounts with invalid codes
SELECT 
    'Invalid Account Codes' as issue_type,
    COUNT(*) as issue_count
FROM accounts 
WHERE (code IS NULL OR code = '' OR LENGTH(code) < 3)
AND deleted_at IS NULL;

-- =====================================================
-- 8. SALES AND PURCHASE DATA VALIDATION
-- =====================================================

SELECT 
    '==== SALES & PURCHASE VALIDATION ====' as analysis_section;

-- Sales data overview
SELECT 
    'Sales Data Overview' as metric,
    COUNT(*) as total_sales,
    SUM(total_amount) as total_sales_amount,
    SUM(paid_amount) as total_paid_amount,
    SUM(outstanding_amount) as total_outstanding,
    COUNT(CASE WHEN status = 'PAID' THEN 1 END) as paid_sales,
    COUNT(CASE WHEN status IN ('PENDING', 'CONFIRMED', 'INVOICED') THEN 1 END) as pending_sales
FROM sales 
WHERE deleted_at IS NULL;

-- Purchase data overview
SELECT 
    'Purchase Data Overview' as metric,
    COUNT(*) as total_purchases,
    SUM(total_amount) as total_purchase_amount,
    SUM(paid_amount) as total_paid_amount,
    SUM(outstanding_amount) as total_outstanding,
    COUNT(CASE WHEN status = 'PAID' THEN 1 END) as paid_purchases,
    COUNT(CASE WHEN status IN ('PENDING', 'CONFIRMED') THEN 1 END) as pending_purchases
FROM purchases 
WHERE deleted_at IS NULL;

-- =====================================================
-- 9. JOURNAL LINES DETAILED ANALYSIS
-- =====================================================

SELECT 
    '==== JOURNAL LINES ANALYSIS ====' as analysis_section;

-- Journal lines summary by account type
SELECT 
    'Journal Lines by Account Type' as metric,
    a.type as account_type,
    COUNT(jl.*) as line_count,
    SUM(jl.debit_amount) as total_debits,
    SUM(jl.credit_amount) as total_credits,
    SUM(jl.debit_amount) - SUM(jl.credit_amount) as net_movement
FROM journal_lines jl
JOIN accounts a ON jl.account_id = a.id
JOIN journal_entries je ON jl.journal_entry_id = je.id
WHERE je.deleted_at IS NULL 
AND a.deleted_at IS NULL
GROUP BY a.type
ORDER BY a.type;

-- Journal lines without proper account linkage
SELECT 
    'Journal Lines with Missing Accounts' as issue_type,
    COUNT(*) as issue_count
FROM journal_lines jl
LEFT JOIN accounts a ON jl.account_id = a.id
WHERE a.id IS NULL;

-- Journal lines with zero amounts
SELECT 
    'Journal Lines with Zero Amounts' as issue_type,
    COUNT(*) as issue_count
FROM journal_lines 
WHERE debit_amount = 0 AND credit_amount = 0;

-- =====================================================
-- 10. MONTHLY FINANCIAL SUMMARY
-- =====================================================

SELECT 
    '==== MONTHLY FINANCIAL SUMMARY ====' as analysis_section;

-- Monthly revenue and expense trends
WITH monthly_financials AS (
    SELECT 
        DATE_TRUNC('month', je.entry_date) as month,
        a.type,
        SUM(jl.credit_amount - jl.debit_amount) as net_amount
    FROM journal_lines jl
    JOIN accounts a ON jl.account_id = a.id
    JOIN journal_entries je ON jl.journal_entry_id = je.id
    WHERE je.status = 'POSTED'
    AND je.deleted_at IS NULL
    AND a.deleted_at IS NULL
    AND a.type IN ('REVENUE', 'EXPENSE')
    GROUP BY DATE_TRUNC('month', je.entry_date), a.type
)
SELECT 
    'Monthly P&L Trends' as metric,
    month,
    SUM(CASE WHEN type = 'REVENUE' THEN net_amount ELSE 0 END) as monthly_revenue,
    SUM(CASE WHEN type = 'EXPENSE' THEN -net_amount ELSE 0 END) as monthly_expenses,
    SUM(CASE WHEN type = 'REVENUE' THEN net_amount ELSE 0 END) - 
    SUM(CASE WHEN type = 'EXPENSE' THEN -net_amount ELSE 0 END) as monthly_net_income
FROM monthly_financials
GROUP BY month
ORDER BY month DESC
LIMIT 12;

-- =====================================================
-- SUMMARY AND RECOMMENDATIONS
-- =====================================================

SELECT 
    '==== VALIDATION SUMMARY ====' as analysis_section;

-- Overall system health check
WITH health_metrics AS (
    SELECT 
        (SELECT COUNT(*) FROM journal_entries WHERE is_balanced = true AND deleted_at IS NULL) as balanced_entries,
        (SELECT COUNT(*) FROM journal_entries WHERE deleted_at IS NULL) as total_entries,
        (SELECT COUNT(*) FROM accounts WHERE is_active = true AND deleted_at IS NULL) as active_accounts,
        (SELECT COUNT(*) FROM accounts WHERE type = 'ASSET' AND is_active = true AND deleted_at IS NULL) as asset_accounts,
        (SELECT ABS(SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) - 
                    (SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) + 
                     SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END) +
                     SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END) -
                     SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END))) 
         FROM accounts WHERE is_active = true AND deleted_at IS NULL) as equation_difference
)
SELECT 
    'System Health Score' as metric,
    CASE WHEN total_entries > 0 
         THEN ROUND((balanced_entries::decimal / total_entries) * 100, 2) 
         ELSE 100 
    END as balance_accuracy_pct,
    active_accounts as total_active_accounts,
    asset_accounts as asset_account_count,
    equation_difference,
    CASE 
        WHEN equation_difference < 0.01 THEN 'BALANCED'
        ELSE 'NEEDS ATTENTION'
    END as accounting_equation_status,
    CASE 
        WHEN balanced_entries::decimal / NULLIF(total_entries, 0) >= 0.95 
        AND active_accounts >= 10 
        AND asset_accounts >= 1
        AND equation_difference < 0.01 
        THEN 'EXCELLENT'
        WHEN balanced_entries::decimal / NULLIF(total_entries, 0) >= 0.85 
        AND active_accounts >= 5
        AND asset_accounts >= 1
        THEN 'GOOD'
        WHEN balanced_entries::decimal / NULLIF(total_entries, 0) >= 0.70
        THEN 'NEEDS ATTENTION'
        ELSE 'CRITICAL'
    END as overall_health_status
FROM health_metrics;

SELECT 
    '=== ANALYSIS COMPLETE ===' as final_message,
    'Check results above for detailed financial data validation' as note,
    now() as completed_at;