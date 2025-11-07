-- Script untuk memastikan akun-akun yang dibutuhkan untuk Asset Management sudah ada
-- Jalankan script ini sebelum menggunakan fitur asset management yang baru

-- 1. Ensure Fixed Assets account exists (1500)
INSERT IGNORE INTO accounts (id, code, name, description, type, category, parent_id, level, is_header, is_active, balance, created_at, updated_at)
VALUES 
(1500, '1500', 'Fixed Assets', 'Property, Plant, and Equipment', 'ASSET', 'FIXED_ASSET', NULL, 1, true, true, 0.00, NOW(), NOW());

-- 2. Ensure Accumulated Depreciation account exists (1501)
INSERT IGNORE INTO accounts (id, code, name, description, type, category, parent_id, level, is_header, is_active, balance, created_at, updated_at)
VALUES 
(1501, '1501', 'Accumulated Depreciation', 'Contra-asset account for accumulated depreciation', 'ASSET', 'FIXED_ASSET', 1500, 2, false, true, 0.00, NOW(), NOW());

-- 3. Ensure Depreciation Expense account exists (6201)
INSERT IGNORE INTO accounts (id, code, name, description, type, category, parent_id, level, is_header, is_active, balance, created_at, updated_at)
VALUES 
(6201, '6201', 'Depreciation Expense', 'Monthly/Annual depreciation expense', 'EXPENSE', 'DEPRECIATION_EXPENSE', NULL, 1, false, true, 0.00, NOW(), NOW());

-- 4. Ensure Cash account exists (1101) - if not already present
INSERT IGNORE INTO accounts (id, code, name, description, type, category, parent_id, level, is_header, is_active, balance, created_at, updated_at)
VALUES 
(1101, '1101', 'Kas', 'Cash on Hand', 'ASSET', 'CURRENT_ASSET', NULL, 1, false, true, 0.00, NOW(), NOW());

-- 5. Ensure Bank account exists (1102) - if not already present
INSERT IGNORE INTO accounts (id, code, name, description, type, category, parent_id, level, is_header, is_active, balance, created_at, updated_at)
VALUES 
(1102, '1102', 'Bank', 'Bank Account', 'ASSET', 'CURRENT_ASSET', NULL, 1, false, true, 0.00, NOW(), NOW());

-- 6. Ensure Accounts Payable exists (2001) - if not already present
INSERT IGNORE INTO accounts (id, code, name, description, type, category, parent_id, level, is_header, is_active, balance, created_at, updated_at)
VALUES 
(2001, '2001', 'Hutang Usaha', 'Accounts Payable', 'LIABILITY', 'CURRENT_LIABILITY', NULL, 1, false, true, 0.00, NOW(), NOW());

-- Display the newly created/existing accounts
SELECT 
    id,
    code,
    name,
    description,
    type,
    category,
    balance,
    is_active
FROM accounts 
WHERE code IN ('1500', '1501', '6201', '1101', '1102', '2001')
ORDER BY code;

-- Show the structure
SELECT 
    a.code,
    a.name,
    a.type,
    a.category,
    CASE 
        WHEN a.type = 'ASSET' AND a.category = 'FIXED_ASSET' THEN '💼 Fixed Asset'
        WHEN a.type = 'ASSET' AND a.category = 'CURRENT_ASSET' THEN '💰 Current Asset'  
        WHEN a.type = 'LIABILITY' THEN '📋 Liability'
        WHEN a.type = 'EXPENSE' THEN '💸 Expense'
        ELSE a.type
    END as account_classification,
    FORMAT(a.balance, 2) as balance_formatted
FROM accounts a
WHERE a.code IN ('1500', '1501', '6201', '1101', '1102', '2001')
ORDER BY a.code;
