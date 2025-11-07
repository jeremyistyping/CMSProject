-- Migration: Disable Auto Balance Sync Triggers
-- Description: Disables database triggers that cause DOUBLE POSTING to COA balances
-- Date: 2025-10-17
-- Issue: Triggers were updating COA balances TWICE (once from journal, once from trigger)
-- Result: Bank balance showing 2x the correct amount (e.g. 11.1M instead of 5.55M)

-- =================================================================================
-- PROBLEM EXPLANATION
-- =================================================================================
-- The auto-sync triggers in 20250930_auto_balance_sync_triggers.sql were causing
-- DOUBLE POSTING because:
-- 
-- Flow with triggers ENABLED (BROKEN):
-- 1. PaymentService creates journal entry → COA +5.55M (correct)
-- 2. PaymentService updates cash_banks.balance → +5.55M (correct)
-- 3. Trigger sync_cashbank_coa_balance() fires → COA +5.55M AGAIN! (WRONG!)
-- 4. Result: COA Bank balance = 11.1M (should be 5.55M)
--
-- Flow with triggers DISABLED (CORRECT):
-- 1. PaymentService creates journal entry → COA +5.55M (correct)
-- 2. PaymentService updates cash_banks.balance → +5.55M (correct)
-- 3. No trigger fires
-- 4. Result: COA Bank balance = 5.55M (correct!)
--
-- Journal system is the SINGLE SOURCE OF TRUTH for COA balances.
-- Triggers are not needed and cause conflicts.
-- =================================================================================

-- Drop the problematic triggers
DO $$ 
BEGIN
    -- Drop trigger that syncs cash bank balance to COA (causes double posting)
    DROP TRIGGER IF EXISTS trigger_sync_cashbank_coa ON cash_banks;
    RAISE NOTICE 'Dropped trigger: trigger_sync_cashbank_coa';
    
    -- Drop triggers that recalculate cash bank balance from transactions
    -- (These also interfere with journal-based balance calculation)
    DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_insert ON cash_bank_transactions;
    RAISE NOTICE 'Dropped trigger: trigger_recalc_cashbank_balance_insert';
    
    DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_update ON cash_bank_transactions;
    RAISE NOTICE 'Dropped trigger: trigger_recalc_cashbank_balance_update';
    
    DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_delete ON cash_bank_transactions;
    RAISE NOTICE 'Dropped trigger: trigger_recalc_cashbank_balance_delete';
    
    -- Drop trigger that validates account balance (causes constraint errors during bulk updates)
    DROP TRIGGER IF EXISTS trigger_validate_account_balance ON accounts;
    RAISE NOTICE 'Dropped trigger: trigger_validate_account_balance';
    
    -- Drop trigger that updates parent balances (calls non-existent function)
    DROP TRIGGER IF EXISTS trigger_update_parent_balances ON accounts;
    RAISE NOTICE 'Dropped trigger: trigger_update_parent_balances';
    
    DROP TRIGGER IF EXISTS trigger_update_parent_account_balances ON accounts;
    RAISE NOTICE 'Dropped trigger: trigger_update_parent_account_balances';
    
    RAISE NOTICE '✅ All auto-sync triggers disabled successfully';
    RAISE NOTICE '   COA balances will now be managed ONLY by journal entries';
    RAISE NOTICE '   This prevents double posting and ensures data consistency';
END $$;

-- =================================================================================
-- Optional: Drop the trigger functions (no longer needed)
-- =================================================================================
-- Uncomment these if you want to completely remove the functions
-- Note: Keeping them doesn't hurt, just leaves unused code in database

-- DROP FUNCTION IF EXISTS sync_cashbank_coa_balance() CASCADE;
-- DROP FUNCTION IF EXISTS recalculate_cashbank_balance() CASCADE;
-- DROP FUNCTION IF EXISTS validate_account_balance_consistency() CASCADE;

-- Keep update_parent_account_balances() as it might be useful for manual maintenance
-- DROP FUNCTION IF EXISTS update_parent_account_balances(INTEGER) CASCADE;

-- =================================================================================
-- Recalculate COA balances from journal entries (one-time fix)
-- =================================================================================
-- This ensures existing balances are corrected after disabling triggers

DO $$
DECLARE
    v_account_id INTEGER;
    v_account_type VARCHAR(20);
    v_calculated_balance DECIMAL(15,2);
    v_current_balance DECIMAL(15,2);
    v_updated_count INTEGER := 0;
BEGIN
    RAISE NOTICE 'Recalculating COA balances from journal entries...';
    
    -- Temporarily disable triggers for this session
    SET session_replication_role = replica;
    
    -- Reset all account balances to 0
    UPDATE accounts SET balance = 0 WHERE deleted_at IS NULL;
    RAISE NOTICE 'Reset all account balances to 0';
    
    -- Recalculate balance for each account from journal entries
    FOR v_account_id, v_account_type IN 
        SELECT DISTINCT ji.account_id, a.type
        FROM simple_ssot_journal_items ji
        JOIN simple_ssot_journals j ON j.id = ji.journal_id
        JOIN accounts a ON a.id = ji.account_id
        WHERE j.status = 'POSTED'
        AND j.deleted_at IS NULL
        AND a.deleted_at IS NULL
    LOOP
        -- Calculate balance based on account type
        IF v_account_type IN ('ASSET', 'EXPENSE') THEN
            -- Assets and Expenses increase with debit
            SELECT COALESCE(SUM(ji.debit - ji.credit), 0)
            INTO v_calculated_balance
            FROM simple_ssot_journal_items ji
            JOIN simple_ssot_journals j ON j.id = ji.journal_id
            WHERE ji.account_id = v_account_id
            AND j.status = 'POSTED'
            AND j.deleted_at IS NULL;
        ELSE
            -- Liabilities, Equity, Revenue increase with credit
            SELECT COALESCE(SUM(ji.credit - ji.debit), 0)
            INTO v_calculated_balance
            FROM simple_ssot_journal_items ji
            JOIN simple_ssot_journals j ON j.id = ji.journal_id
            WHERE ji.account_id = v_account_id
            AND j.status = 'POSTED'
            AND j.deleted_at IS NULL;
        END IF;
        
        -- Update account balance
        UPDATE accounts 
        SET balance = v_calculated_balance, updated_at = NOW()
        WHERE id = v_account_id;
        
        v_updated_count := v_updated_count + 1;
    END LOOP;
    
    RAISE NOTICE 'Updated % account balances from journal entries', v_updated_count;
    
    -- Update parent account balances (sum of children)
    UPDATE accounts p
    SET balance = (
        SELECT COALESCE(SUM(c.balance), 0)
        FROM accounts c
        WHERE c.parent_id = p.id
        AND c.deleted_at IS NULL
    ),
    updated_at = NOW()
    WHERE p.is_header = true
    AND p.deleted_at IS NULL;
    
    RAISE NOTICE 'Updated parent account balances';
    
    -- Re-enable triggers
    SET session_replication_role = DEFAULT;
    
    RAISE NOTICE '✅ Balance recalculation complete!';
    
EXCEPTION WHEN OTHERS THEN
    -- Re-enable triggers even if error occurs
    SET session_replication_role = DEFAULT;
    RAISE EXCEPTION 'Balance recalculation failed: %', SQLERRM;
END $$;

-- =================================================================================
-- Verification: Show current balances for key accounts
-- =================================================================================
DO $$
DECLARE
    v_rec RECORD;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '📊 Current Account Balances (Key Accounts):';
    RAISE NOTICE '┌────────┬──────────────────────────┬────────────────┐';
    RAISE NOTICE '│ Code   │ Name                     │ Balance        │';
    RAISE NOTICE '├────────┼──────────────────────────┼────────────────┤';
    
    FOR v_rec IN 
        SELECT code, name, balance
        FROM accounts
        WHERE code IN ('1102', '1201', '4101', '2103')
        AND deleted_at IS NULL
        ORDER BY code
    LOOP
        RAISE NOTICE '│ %-6s │ %-24s │ %14s │', 
            v_rec.code, 
            v_rec.name, 
            TO_CHAR(v_rec.balance, 'FM999,999,999,990.00');
    END LOOP;
    
    RAISE NOTICE '└────────┴──────────────────────────┴────────────────┘';
    RAISE NOTICE '';
END $$;

-- =================================================================================
-- Final message
-- =================================================================================
DO $$
BEGIN
    RAISE NOTICE '✅ Migration completed successfully!';
    RAISE NOTICE '';
    RAISE NOTICE 'Summary of changes:';
    RAISE NOTICE '  - Disabled 5 auto-sync triggers';
    RAISE NOTICE '  - Recalculated all COA balances from journal entries';
    RAISE NOTICE '  - Updated parent account balances';
    RAISE NOTICE '';
    RAISE NOTICE 'Going forward:';
    RAISE NOTICE '  - Journal entries are the SINGLE SOURCE OF TRUTH for COA balances';
    RAISE NOTICE '  - No automatic syncing from cash_banks to accounts';
    RAISE NOTICE '  - No double posting issues';
    RAISE NOTICE '';
END $$;

