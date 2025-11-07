-- Migration: Fix VAT/PPN Accounts Display Issue
-- Description: Ensure tax_account_settings has proper foreign key references to accounts 1240 and 2103
-- Version: 041
-- Created: 2025-11-02
-- Purpose: Fix the issue where PPN Masukan (1240) and PPN Keluaran (2103) don't appear in frontend

-- ============================================================================
-- STEP 1: Verify accounts 1240 and 2103 exist
-- ============================================================================

DO $$
DECLARE
    account_1240_id INT;
    account_2103_id INT;
    settings_count INT;
BEGIN
    -- Check if account 1240 exists
    SELECT id INTO account_1240_id
    FROM accounts
    WHERE code = '1240' AND deleted_at IS NULL
    LIMIT 1;
    
    -- Check if account 2103 exists
    SELECT id INTO account_2103_id
    FROM accounts
    WHERE code = '2103' AND deleted_at IS NULL
    LIMIT 1;
    
    RAISE NOTICE '';
    RAISE NOTICE '====================================';
    RAISE NOTICE '  VAT/PPN ACCOUNTS DIAGNOSTIC';
    RAISE NOTICE '====================================';
    
    IF account_1240_id IS NULL THEN
        RAISE EXCEPTION '❌ CRITICAL: Account 1240 (PPN Masukan) NOT FOUND in database!';
    ELSE
        RAISE NOTICE '✅ Account 1240 (PPN Masukan) found with ID: %', account_1240_id;
    END IF;
    
    IF account_2103_id IS NULL THEN
        RAISE EXCEPTION '❌ CRITICAL: Account 2103 (PPN Keluaran) NOT FOUND in database!';
    ELSE
        RAISE NOTICE '✅ Account 2103 (PPN Keluaran) found with ID: %', account_2103_id;
    END IF;
    
    -- Check if tax_account_settings exists
    SELECT COUNT(*) INTO settings_count
    FROM tax_account_settings
    WHERE deleted_at IS NULL;
    
    RAISE NOTICE 'ℹ️  Found % tax_account_settings record(s)', settings_count;
    RAISE NOTICE '====================================';
END $$;

-- ============================================================================
-- STEP 2: Update existing tax_account_settings to link with VAT accounts
-- ============================================================================

DO $$
DECLARE
    account_1240_id INT;
    account_2103_id INT;
    settings_record RECORD;
    updated_count INT := 0;
BEGIN
    -- Get account IDs
    SELECT id INTO account_1240_id FROM accounts WHERE code = '1240' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO account_2103_id FROM accounts WHERE code = '2103' AND deleted_at IS NULL LIMIT 1;
    
    -- Update ALL tax_account_settings records to ensure proper linkage
    FOR settings_record IN 
        SELECT id, purchase_input_vat_account_id, sales_output_vat_account_id
        FROM tax_account_settings
        WHERE deleted_at IS NULL
    LOOP
        -- Update the settings record with correct account IDs
        UPDATE tax_account_settings
        SET 
            purchase_input_vat_account_id = account_1240_id,
            sales_output_vat_account_id = account_2103_id,
            updated_at = NOW()
        WHERE id = settings_record.id;
        
        updated_count := updated_count + 1;
        
        RAISE NOTICE 'Updated tax_account_settings ID % with PPN accounts (1240: ID %, 2103: ID %)', 
            settings_record.id, account_1240_id, account_2103_id;
    END LOOP;
    
    IF updated_count = 0 THEN
        RAISE NOTICE '⚠️  No tax_account_settings records found to update';
        RAISE NOTICE 'ℹ️  Creating default tax_account_settings record...';
        
        -- Create a default settings record if none exists
        INSERT INTO tax_account_settings (
            sales_receivable_account_id,
            sales_cash_account_id,
            sales_bank_account_id,
            sales_revenue_account_id,
            sales_output_vat_account_id,
            purchase_payable_account_id,
            purchase_cash_account_id,
            purchase_bank_account_id,
            purchase_input_vat_account_id,
            purchase_expense_account_id,
            is_active,
            apply_to_all_companies,
            updated_by,
            created_at,
            updated_at
        )
        SELECT 
            COALESCE((SELECT id FROM accounts WHERE code = '1201' AND deleted_at IS NULL LIMIT 1), 1),  -- Sales Receivable
            COALESCE((SELECT id FROM accounts WHERE code = '1101' AND deleted_at IS NULL LIMIT 1), 1),  -- Cas h
            COALESCE((SELECT id FROM accounts WHERE code = '1102' AND deleted_at IS NULL LIMIT 1), 1),  -- Bank
            COALESCE((SELECT id FROM accounts WHERE code = '4101' AND deleted_at IS NULL LIMIT 1), 1),  -- Revenue
            account_2103_id,  -- PPN Keluaran
            COALESCE((SELECT id FROM accounts WHERE code = '2001' AND deleted_at IS NULL LIMIT 1), 1),  -- Purchase Payable
            COALESCE((SELECT id FROM accounts WHERE code = '1101' AND deleted_at IS NULL LIMIT 1), 1),  -- Cash
            COALESCE((SELECT id FROM accounts WHERE code = '1102' AND deleted_at IS NULL LIMIT 1), 1),  -- Bank
            account_1240_id,  -- PPN Masukan
            COALESCE((SELECT id FROM accounts WHERE code = '6001' AND deleted_at IS NULL LIMIT 1), 1),  -- Expense
            true,
            true,
            1,  -- System user
            NOW(),
            NOW()
        WHERE NOT EXISTS (
            SELECT 1 FROM tax_account_settings WHERE deleted_at IS NULL
        );
        
        IF FOUND THEN
            RAISE NOTICE '✅ Created default tax_account_settings with VAT accounts';
            updated_count := 1;
        END IF;
    ELSE
        RAISE NOTICE '✅ Updated % tax_account_settings record(s)', updated_count;
    END IF;
END $$;

-- ============================================================================
-- STEP 3: Verification
-- ============================================================================

DO $$
DECLARE
    settings_record RECORD;
    account_1240_name TEXT;
    account_2103_name TEXT;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '====================================';
    RAISE NOTICE '  VERIFICATION RESULTS';
    RAISE NOTICE '====================================';
    
    -- Get account names
    SELECT name INTO account_1240_name FROM accounts WHERE code = '1240' AND deleted_at IS NULL LIMIT 1;
    SELECT name INTO account_2103_name FROM accounts WHERE code = '2103' AND deleted_at IS NULL LIMIT 1;
    
    -- Check all tax_account_settings
    FOR settings_record IN
        SELECT 
            ts.id,
            ts.is_active,
            ts.purchase_input_vat_account_id,
            ts.sales_output_vat_account_id,
            a1.code as ppn_masukan_code,
            a1.name as ppn_masukan_name,
            a2.code as ppn_keluaran_code,
            a2.name as ppn_keluaran_name
        FROM tax_account_settings ts
        LEFT JOIN accounts a1 ON ts.purchase_input_vat_account_id = a1.id
        LEFT JOIN accounts a2 ON ts.sales_output_vat_account_id = a2.id
        WHERE ts.deleted_at IS NULL
        ORDER BY ts.is_active DESC, ts.updated_at DESC
    LOOP
        RAISE NOTICE '';
        RAISE NOTICE 'Settings ID: % (Active: %)', settings_record.id, settings_record.is_active;
        RAISE NOTICE '  PPN Masukan (Input VAT):';
        IF settings_record.ppn_masukan_code IS NOT NULL THEN
            RAISE NOTICE '    ✅ %  - % (ID: %)', 
                settings_record.ppn_masukan_code, 
                settings_record.ppn_masukan_name,
                settings_record.purchase_input_vat_account_id;
        ELSE
            RAISE NOTICE '    ❌ NOT CONFIGURED';
        END IF;
        
        RAISE NOTICE '  PPN Keluaran (Output VAT):';
        IF settings_record.ppn_keluaran_code IS NOT NULL THEN
            RAISE NOTICE '    ✅ % - % (ID: %)', 
                settings_record.ppn_keluaran_code, 
                settings_record.ppn_keluaran_name,
                settings_record.sales_output_vat_account_id;
        ELSE
            RAISE NOTICE '    ❌ NOT CONFIGURED';
        END IF;
    END LOOP;
    
    RAISE NOTICE '';
    RAISE NOTICE '====================================';
    RAISE NOTICE '✅ Migration 041 completed successfully!';
    RAISE NOTICE '';
    RAISE NOTICE 'ℹ️  What this migration did:';
    RAISE NOTICE '   1. Verified accounts 1240 and 2103 exist';
    RAISE NOTICE '   2. Updated tax_account_settings with proper account links';
    RAISE NOTICE '   3. Created default settings if none existed';
    RAISE NOTICE '';
    RAISE NOTICE '📝 NEXT STEPS:';
    RAISE NOTICE '   1. Restart your backend server';
    RAISE NOTICE '   2. Clear browser cache or do hard refresh (Ctrl+Shift+R)';
    RAISE NOTICE '   3. Go to Settings > Tax Accounts page';
    RAISE NOTICE '   4. VAT/PPN accounts should now appear!';
    RAISE NOTICE '====================================';
END $$;
