-- Migration: Fix Account 2102 Issue and Ensure Balance Sheet Accuracy
-- Issue: Account 2102 was created with code in LIABILITY range (2xxx) but type ASSET and name "PPN Masukan"
-- This causes balance sheet to be off by Rp 165.000

-- Step 1: Check current state of account 2102
DO $$
DECLARE
    v_account_id INTEGER;
    v_account_code VARCHAR;
    v_account_name VARCHAR;
    v_account_type VARCHAR;
    v_account_balance DECIMAL;
BEGIN
    -- Get account 2102 details
    SELECT id, code, name, type, balance 
    INTO v_account_id, v_account_code, v_account_name, v_account_type, v_account_balance
    FROM accounts 
    WHERE code = '2102' AND deleted_at IS NULL;
    
    IF v_account_id IS NOT NULL THEN
        RAISE NOTICE 'Found account 2102: ID=%, Code=%, Name=%, Type=%, Balance=%', 
            v_account_id, v_account_code, v_account_name, v_account_type, v_account_balance;
            
        -- If account 2102 has balance, we need to transfer it
        IF v_account_balance <> 0 THEN
            RAISE NOTICE 'Account 2102 has non-zero balance: %. Creating adjustment journal entry.', v_account_balance;
            
            -- Create a journal entry to transfer balance from 2102 to proper account
            -- Since 2102 is named "PPN Masukan" but in liability range, we should transfer to 1240 (proper PPN MASUKAN account)
            INSERT INTO unified_journal_ledger (
                entry_number, 
                entry_date, 
                description, 
                status, 
                created_by,
                created_at,
                updated_at
            ) VALUES (
                'ADJ-2102-FIX-' || TO_CHAR(NOW(), 'YYYYMMDD'),
                CURRENT_DATE,
                'Adjustment: Transfer balance from incorrect account 2102 to proper PPN MASUKAN account 1240',
                'POSTED',
                1, -- Admin user
                NOW(),
                NOW()
            );
            
            -- Get the journal ID
            DECLARE
                v_journal_id BIGINT;
                v_ppn_masukan_id INTEGER;
            BEGIN
                SELECT id INTO v_journal_id FROM unified_journal_ledger 
                WHERE entry_number = 'ADJ-2102-FIX-' || TO_CHAR(NOW(), 'YYYYMMDD');
                
                -- Get proper PPN MASUKAN account (1240)
                SELECT id INTO v_ppn_masukan_id FROM accounts WHERE code = '1240' AND deleted_at IS NULL;
                
                IF v_ppn_masukan_id IS NOT NULL THEN
                    -- Credit account 2102 (to zero it out)
                    INSERT INTO unified_journal_lines (journal_id, account_id, description, debit_amount, credit_amount, line_number)
                    VALUES (v_journal_id, v_account_id, 'Transfer from incorrect account 2102', 0, ABS(v_account_balance), 1);
                    
                    -- Debit account 1240 (proper PPN MASUKAN)
                    INSERT INTO unified_journal_lines (journal_id, account_id, description, debit_amount, credit_amount, line_number)
                    VALUES (v_journal_id, v_ppn_masukan_id, 'Transfer to proper PPN MASUKAN account', ABS(v_account_balance), 0, 2);
                    
                    RAISE NOTICE 'Created adjustment journal entry to transfer balance from 2102 to 1240';
                ELSE
                    RAISE WARNING 'Account 1240 (PPN MASUKAN) not found. Manual adjustment required.';
                END IF;
            END;
        END IF;
        
        -- Soft delete account 2102
        UPDATE accounts 
        SET deleted_at = NOW(), 
            is_active = false,
            updated_at = NOW()
        WHERE code = '2102' AND deleted_at IS NULL;
        
        RAISE NOTICE 'Successfully soft-deleted account 2102';
    ELSE
        RAISE NOTICE 'Account 2102 not found. No action needed.';
    END IF;
END $$;

-- Step 2: Ensure correct PPN accounts exist with proper configuration
-- Account 1240: PPN MASUKAN (Asset - Input VAT)
INSERT INTO accounts (code, name, type, category, level, is_header, is_active, balance, created_at, updated_at)
VALUES ('1240', 'PPN MASUKAN', 'ASSET', 'CURRENT_ASSET', 3, false, true, 0, NOW(), NOW())
ON CONFLICT (code) WHERE deleted_at IS NULL
DO UPDATE SET 
    name = 'PPN MASUKAN',
    type = 'ASSET',
    is_active = true,
    updated_at = NOW();

-- Account 2103: PPN KELUARAN (Liability - Output VAT)  
INSERT INTO accounts (code, name, type, category, level, is_header, is_active, balance, created_at, updated_at)
VALUES ('2103', 'PPN KELUARAN', 'LIABILITY', 'CURRENT_LIABILITY', 3, false, true, 0, NOW(), NOW())
ON CONFLICT (code) WHERE deleted_at IS NULL
DO UPDATE SET 
    name = 'PPN KELUARAN',
    type = 'LIABILITY',
    is_active = true,
    updated_at = NOW();

-- Step 3: Set correct parent relationships for PPN accounts
UPDATE accounts 
SET parent_id = (SELECT id FROM accounts WHERE code = '1100' AND deleted_at IS NULL LIMIT 1)
WHERE code = '1240' AND deleted_at IS NULL;

UPDATE accounts 
SET parent_id = (SELECT id FROM accounts WHERE code = '2100' AND deleted_at IS NULL LIMIT 1)
WHERE code = '2103' AND deleted_at IS NULL;

-- Step 4: Verify no other misplaced PPN accounts
DO $$
DECLARE
    v_wrong_ppn RECORD;
BEGIN
    FOR v_wrong_ppn IN 
        SELECT code, name, type 
        FROM accounts 
        WHERE deleted_at IS NULL
          AND (
              (code LIKE '2%' AND UPPER(name) LIKE '%PPN MASUKAN%') OR
              (code LIKE '1%' AND UPPER(name) LIKE '%PPN KELUARAN%')
          )
    LOOP
        RAISE WARNING 'Found misplaced PPN account: Code=%, Name=%, Type=%. Please review manually.', 
            v_wrong_ppn.code, v_wrong_ppn.name, v_wrong_ppn.type;
    END LOOP;
END $$;

-- Step 5: Add comment for audit trail
COMMENT ON TABLE accounts IS 'Chart of Accounts - Fixed account 2102 issue on 2025-11-07';
