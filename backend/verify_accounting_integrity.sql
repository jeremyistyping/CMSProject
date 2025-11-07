-- ========================================
-- ACCOUNTING INTEGRITY VERIFICATION QUERIES
-- ========================================
-- Use these queries to manually verify your accounting system
-- Run in your PostgreSQL/MySQL client

-- Query 1: Check Journal Balance (Debit = Credit)
-- All journal entries should have equal debit and credit amounts
SELECT 
    id,
    entry_number,
    entry_date,
    source_type,
    total_debit,
    total_credit,
    ABS(total_debit - total_credit) as difference,
    CASE 
        WHEN ABS(total_debit - total_credit) < 0.01 THEN '✅ Balanced'
        ELSE '❌ Unbalanced'
    END as status
FROM unified_journal_ledger
WHERE deleted_at IS NULL
ORDER BY ABS(total_debit - total_credit) DESC
LIMIT 50;

-- Query 2: Find Sales Without COGS Entries
-- All INVOICED/PAID sales should have corresponding COGS entries
SELECT 
    s.id as sale_id,
    s.invoice_number,
    s.date,
    s.status,
    s.total_amount,
    s.subtotal,
    CASE 
        WHEN EXISTS (
            SELECT 1 
            FROM unified_journal_ledger uje
            JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
            JOIN accounts a ON a.id = ujl.account_id
            WHERE uje.source_type = 'SALE' 
              AND uje.source_id = s.id
              AND a.code = '5101'  -- COGS account
              AND uje.deleted_at IS NULL
        ) THEN '✅ Has COGS'
        ELSE '❌ Missing COGS'
    END as cogs_status
FROM sales s
WHERE s.status IN ('INVOICED', 'PAID')
  AND s.deleted_at IS NULL
ORDER BY s.date DESC
LIMIT 100;

-- Query 3: Calculate Total COGS vs Revenue
-- Compare revenue and COGS to verify gross profit margin
SELECT 
    'Revenue (4xxx)' as account_type,
    SUM(ujl.credit_amount - ujl.debit_amount) as amount
FROM unified_journal_lines ujl
JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
JOIN accounts a ON a.id = ujl.account_id
WHERE a.code LIKE '4%'

UNION ALL

SELECT 
    'COGS (5101)' as account_type,
    SUM(ujl.debit_amount - ujl.credit_amount) as amount
FROM unified_journal_lines ujl
JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
JOIN accounts a ON a.id = ujl.account_id
WHERE a.code = '5101'

UNION ALL

SELECT 
    'Gross Profit' as account_type,
    SUM(CASE WHEN a.code LIKE '4%' THEN ujl.credit_amount - ujl.debit_amount ELSE 0 END) -
    SUM(CASE WHEN a.code = '5101' THEN ujl.debit_amount - ujl.credit_amount ELSE 0 END) as amount
FROM unified_journal_lines ujl
JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
JOIN accounts a ON a.id = ujl.account_id;

-- Query 4: Products with Zero Cost Price
-- These products will cause zero COGS when sold
SELECT 
    p.id,
    p.name,
    p.sku,
    p.price as selling_price,
    p.cost_price,
    p.stock,
    CASE 
        WHEN p.cost_price > 0 THEN ROUND((p.price - p.cost_price) / p.price * 100, 2)
        ELSE NULL
    END as margin_percentage,
    CASE 
        WHEN p.cost_price = 0 OR p.cost_price IS NULL THEN '❌ Missing'
        ELSE '✅ OK'
    END as cost_price_status
FROM products p
WHERE p.deleted_at IS NULL
  AND p.stock > 0
ORDER BY p.stock DESC, p.id
LIMIT 50;

-- Query 5: Account Balance vs Journal Lines Reconciliation
-- Verify that account balances match the sum of journal entries
SELECT 
    a.code,
    a.name,
    a.type,
    a.balance as stored_balance,
    COALESCE(
        CASE 
            WHEN UPPER(a.type) IN ('ASSET', 'EXPENSE') THEN 
                SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
            ELSE 
                SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
        END,
    0) as calculated_balance,
    ABS(a.balance - COALESCE(
        CASE 
            WHEN UPPER(a.type) IN ('ASSET', 'EXPENSE') THEN 
                SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
            ELSE 
                SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
        END,
    0)) as difference,
    CASE 
        WHEN ABS(a.balance - COALESCE(
            CASE 
                WHEN UPPER(a.type) IN ('ASSET', 'EXPENSE') THEN 
                    SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
                ELSE 
                    SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
            END,
        0)) < 0.01 THEN '✅ Match'
        ELSE '❌ Mismatch'
    END as status
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
WHERE a.is_header = false AND a.deleted_at IS NULL
GROUP BY a.id, a.code, a.name, a.type, a.balance
HAVING ABS(a.balance - COALESCE(
    CASE 
        WHEN UPPER(a.type) IN ('ASSET', 'EXPENSE') THEN 
            SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
        ELSE 
            SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
    END,
0)) > 0.01
ORDER BY difference DESC;

-- Query 6: Detailed COGS Breakdown by Sale
-- See COGS calculation for each sale
SELECT 
    s.id as sale_id,
    s.invoice_number,
    s.date,
    s.total_amount as sale_amount,
    COALESCE(SUM(ujl.debit_amount), 0) as cogs_amount,
    s.total_amount - COALESCE(SUM(ujl.debit_amount), 0) as gross_profit,
    CASE 
        WHEN s.total_amount > 0 THEN 
            ROUND((s.total_amount - COALESCE(SUM(ujl.debit_amount), 0)) / s.total_amount * 100, 2)
        ELSE 0
    END as gross_margin_percentage
FROM sales s
LEFT JOIN unified_journal_ledger uje ON uje.source_type = 'SALE' AND uje.source_id = s.id AND uje.deleted_at IS NULL
LEFT JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
LEFT JOIN accounts a ON a.id = ujl.account_id AND a.code = '5101'
WHERE s.status IN ('INVOICED', 'PAID')
  AND s.deleted_at IS NULL
GROUP BY s.id, s.invoice_number, s.date, s.total_amount
ORDER BY s.date DESC
LIMIT 50;

-- Query 7: Profit & Loss Summary (Quick View)
-- Get a quick P&L summary
WITH pl_summary AS (
    SELECT 
        SUM(CASE WHEN a.code LIKE '4%' THEN ujl.credit_amount - ujl.debit_amount ELSE 0 END) as revenue,
        SUM(CASE WHEN a.code = '5101' THEN ujl.debit_amount - ujl.credit_amount ELSE 0 END) as cogs,
        SUM(CASE WHEN a.code LIKE '52%' OR a.code LIKE '53%' OR a.code LIKE '54%' THEN ujl.debit_amount - ujl.credit_amount ELSE 0 END) as operating_expenses,
        SUM(CASE WHEN a.code LIKE '5%' AND a.code NOT IN ('5101') AND a.code NOT LIKE '52%' AND a.code NOT LIKE '53%' AND a.code NOT LIKE '54%' THEN ujl.debit_amount - ujl.credit_amount ELSE 0 END) as other_expenses
    FROM unified_journal_lines ujl
    JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
    JOIN accounts a ON a.id = ujl.account_id
)
SELECT 
    'Revenue' as line_item,
    revenue as amount,
    CASE WHEN revenue > 0 THEN 100.0 ELSE 0 END as percentage
FROM pl_summary

UNION ALL

SELECT 
    'Cost of Goods Sold' as line_item,
    cogs as amount,
    CASE WHEN revenue > 0 THEN ROUND((cogs / revenue) * 100, 2) ELSE 0 END as percentage
FROM pl_summary

UNION ALL

SELECT 
    'Gross Profit' as line_item,
    revenue - cogs as amount,
    CASE WHEN revenue > 0 THEN ROUND(((revenue - cogs) / revenue) * 100, 2) ELSE 0 END as percentage
FROM pl_summary

UNION ALL

SELECT 
    'Operating Expenses' as line_item,
    operating_expenses as amount,
    CASE WHEN revenue > 0 THEN ROUND((operating_expenses / revenue) * 100, 2) ELSE 0 END as percentage
FROM pl_summary

UNION ALL

SELECT 
    'Net Income' as line_item,
    revenue - cogs - operating_expenses - other_expenses as amount,
    CASE WHEN revenue > 0 THEN ROUND(((revenue - cogs - operating_expenses - other_expenses) / revenue) * 100, 2) ELSE 0 END as percentage
FROM pl_summary;

-- Query 8: Check for Duplicate Journal Entries
-- Find potential duplicate postings
SELECT 
    uje.source_type,
    uje.source_id,
    uje.source_code,
    COUNT(*) as journal_count,
    STRING_AGG(uje.id::TEXT, ', ') as journal_ids,
    SUM(uje.total_debit) as total_debit_sum
FROM unified_journal_ledger uje
WHERE uje.deleted_at IS NULL
  AND uje.status = 'POSTED'
GROUP BY uje.source_type, uje.source_id, uje.source_code
HAVING COUNT(*) > 2  -- Allow 2: one for sale, one for COGS
ORDER BY journal_count DESC;
