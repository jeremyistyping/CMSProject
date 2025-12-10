-- Migration: Update COA Budget Categories
-- Version: 079
-- Description: Add budget_category and work_package to existing COA accounts

-- Update Material accounts to OPERASIONAL_BUDGET
UPDATE coa_accounts 
SET budget_category = 'OPERASIONAL_BUDGET',
    work_package = 'Material'
WHERE code LIKE '51%' 
  AND budget_category IS NULL
  AND deleted_at IS NULL;

-- Update Labor/Upah accounts to LABOUR_BUDGET  
UPDATE coa_accounts 
SET budget_category = 'LABOUR_BUDGET',
    work_package = 'Tenaga Kerja'
WHERE code LIKE '52%' 
  AND budget_category IS NULL
  AND deleted_at IS NULL;

-- Update Equipment accounts to OPERASIONAL_BUDGET
UPDATE coa_accounts 
SET budget_category = 'OPERASIONAL_BUDGET',
    work_package = 'Peralatan'
WHERE code LIKE '53%' 
  AND budget_category IS NULL
  AND deleted_at IS NULL;

-- Update Subcontractor accounts to OPERASIONAL_BUDGET
UPDATE coa_accounts 
SET budget_category = 'OPERASIONAL_BUDGET',
    work_package = 'Subkontraktor'
WHERE code LIKE '54%' 
  AND budget_category IS NULL
  AND deleted_at IS NULL;

-- Update Overhead accounts to OTHER
UPDATE coa_accounts 
SET budget_category = 'OTHER',
    work_package = 'Overhead'
WHERE code LIKE '55%' 
  AND budget_category IS NULL
  AND deleted_at IS NULL;

-- Update Other Expense accounts to OTHER
UPDATE coa_accounts 
SET budget_category = 'OTHER',
    work_package = 'Lain-lain'
WHERE code LIKE '59%' 
  AND budget_category IS NULL
  AND deleted_at IS NULL;

-- Show updated records
SELECT code, name, budget_category, work_package 
FROM coa_accounts 
WHERE deleted_at IS NULL 
ORDER BY code;
