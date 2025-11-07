-- Migration: Fix Account Mapping for Tax and Revenue
-- Date: 2025-09-26
-- Purpose: Ensure proper account types for PPN and Revenue accounts

-- Fix PPN Keluaran (output tax) - should be LIABILITY
UPDATE accounts 
SET 
    name = 'PPN Keluaran',
    type = 'LIABILITY',
    is_active = true,
    updated_at = NOW()
WHERE code = '2103';

-- Ensure PPN Masukan (input tax) - should be ASSET  
UPDATE accounts
SET 
    name = 'PPN Masukan', 
    type = 'ASSET',
    is_active = true,
    updated_at = NOW()
WHERE code = '2102';

-- Fix Sales Revenue account - should be REVENUE
UPDATE accounts
SET 
    name = 'Pendapatan Penjualan',
    type = 'REVENUE', 
    is_active = true,
    updated_at = NOW()
WHERE code = '4101';

-- Create missing accounts if they don't exist
INSERT INTO accounts (code, name, type, is_active, is_header, created_at, updated_at)
SELECT '2103', 'PPN Keluaran', 'LIABILITY', true, false, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM accounts WHERE code = '2103');

INSERT INTO accounts (code, name, type, is_active, is_header, created_at, updated_at)
SELECT '4101', 'Pendapatan Penjualan', 'REVENUE', true, false, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM accounts WHERE code = '4101');

-- Log the migration
INSERT INTO migration_logs (name, executed_at, description) 
VALUES (
    '031_fix_account_mapping_for_tax_and_revenue.sql',
    NOW(),
    'Fixed account types and names for PPN and Revenue accounts to ensure proper accounting classification'
);