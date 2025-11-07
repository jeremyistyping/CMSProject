-- Migration: Comprehensive Auto Balance Sync System
-- Description: Complete balance synchronization system with all fixes and improvements
-- Version: 2025-09-30-v2
-- Features:
-- - Unified trigger approach to avoid execution order issues
-- - BIGINT parameter types for compatibility
-- - Automatic cash bank to COA account synchronization
-- - Parent account balance rollup
-- - Manual sync functions for maintenance
-- - Account hierarchy validation

-- =================================================================================
-- 1. Drop existing functions and triggers to ensure clean installation
-- =================================================================================

-- Drop existing triggers first
DROP TRIGGER IF EXISTS trigger_sync_cashbank_coa ON cash_banks;
DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_insert ON cash_bank_transactions;
DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_update ON cash_bank_transactions;
DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_delete ON cash_bank_transactions;
DROP TRIGGER IF EXISTS trigger_validate_account_balance ON accounts;

-- Drop existing functions
DROP FUNCTION IF EXISTS sync_cashbank_coa_balance();
DROP FUNCTION IF EXISTS recalculate_cashbank_balance();
DROP FUNCTION IF EXISTS update_parent_account_balances(INTEGER);
DROP FUNCTION IF EXISTS update_parent_account_balances(BIGINT);
DROP FUNCTION IF EXISTS validate_account_balance_consistency();
DROP FUNCTION IF EXISTS manual_sync_cashbank_coa(INTEGER);
DROP FUNCTION IF EXISTS manual_sync_cashbank_coa(BIGINT);
DROP FUNCTION IF EXISTS ensure_cashbank_not_header();

-- =================================================================================
-- 2. Function to update parent account balances recursively (BIGINT compatible)
-- =================================================================================

CREATE OR REPLACE FUNCTION update_parent_account_balances(child_account_id BIGINT)
RETURNS VOID AS $$
DECLARE
    parent_account_id BIGINT;
    parent_balance DECIMAL(15,2);
BEGIN
    -- Get the parent of the current account
    SELECT parent_id INTO parent_account_id
    FROM accounts 
    WHERE id = child_account_id 
    AND deleted_at IS NULL;
    
    -- If there's a parent, update its balance
    IF parent_account_id IS NOT NULL THEN
        -- Calculate sum of all children balances
        SELECT COALESCE(SUM(balance), 0)
        INTO parent_balance
        FROM accounts 
        WHERE parent_id = parent_account_id 
        AND deleted_at IS NULL;
        
        -- Update parent balance
        UPDATE accounts 
        SET 
            balance = parent_balance,
            updated_at = NOW()
        WHERE id = parent_account_id 
        AND deleted_at IS NULL;
        
        -- Recursively update parent's parent
        PERFORM update_parent_account_balances(parent_account_id);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- =================================================================================
-- 3. Unified function to handle cash bank balance recalculation and COA sync
-- =================================================================================

CREATE OR REPLACE FUNCTION recalculate_cashbank_balance()
RETURNS TRIGGER AS $$
DECLARE
    cash_bank_id_val BIGINT;
    new_balance DECIMAL(15,2);
    linked_account_id BIGINT;
BEGIN
    -- Get the cash bank ID from either NEW or OLD record
    IF TG_OP = 'DELETE' THEN
        cash_bank_id_val := OLD.cash_bank_id;
    ELSE
        cash_bank_id_val := NEW.cash_bank_id;
    END IF;
    
    -- Calculate new balance from all transactions for this cash bank
    SELECT COALESCE(SUM(amount), 0)
    INTO new_balance
    FROM cash_bank_transactions 
    WHERE cash_bank_id = cash_bank_id_val 
    AND deleted_at IS NULL;
    
    -- Update cash bank balance and get linked account ID in one query
    UPDATE cash_banks 
    SET 
        balance = new_balance,
        updated_at = NOW()
    WHERE id = cash_bank_id_val 
    AND deleted_at IS NULL
    RETURNING account_id INTO linked_account_id;
    
    -- Update linked COA account balance directly (unified approach)
    IF linked_account_id IS NOT NULL THEN
        UPDATE accounts 
        SET 
            balance = new_balance,
            updated_at = NOW()
        WHERE id = linked_account_id 
        AND deleted_at IS NULL;
        
        -- Update parent balances recursively
        PERFORM update_parent_account_balances(linked_account_id);
    END IF;
    
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- =================================================================================
-- 4. Account balance validation function with hierarchy validation
-- =================================================================================

CREATE OR REPLACE FUNCTION validate_account_balance_consistency()
RETURNS TRIGGER AS $$
DECLARE
    calculated_balance DECIMAL(15,2);
BEGIN
    -- If this is a header account, ensure balance equals sum of children
    IF NEW.is_header = true THEN
        SELECT COALESCE(SUM(balance), 0)
        INTO calculated_balance
        FROM accounts 
        WHERE parent_id = NEW.id 
        AND deleted_at IS NULL;
        
        -- Auto-correct the balance if it doesn't match
        IF ABS(NEW.balance - calculated_balance) > 0.01 THEN
            NEW.balance := calculated_balance;
        END IF;
    END IF;
    
    -- Update parent balances when a child balance changes
    IF OLD.balance IS DISTINCT FROM NEW.balance THEN
        PERFORM update_parent_account_balances(NEW.id::BIGINT);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =================================================================================
-- 5. Manual sync functions for maintenance and debugging
-- =================================================================================

-- Manual sync function for a specific cash bank
CREATE OR REPLACE FUNCTION manual_sync_cashbank_coa(target_cash_bank_id BIGINT)
RETURNS VOID AS $$
DECLARE
    cash_bank_balance DECIMAL(15,2);
    linked_account_id BIGINT;
BEGIN
    -- Get cash bank balance and linked account
    SELECT balance, account_id INTO cash_bank_balance, linked_account_id
    FROM cash_banks 
    WHERE id = target_cash_bank_id 
    AND deleted_at IS NULL;
    
    -- Update linked COA account balance
    IF linked_account_id IS NOT NULL THEN
        UPDATE accounts 
        SET 
            balance = cash_bank_balance,
            updated_at = NOW()
        WHERE id = linked_account_id 
        AND deleted_at IS NULL;
        
        -- Update parent balances
        PERFORM update_parent_account_balances(linked_account_id);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Manual sync function for all cash banks
CREATE OR REPLACE FUNCTION manual_sync_all_cashbank_coa()
RETURNS TEXT AS $$
DECLARE
    cash_bank_record RECORD;
    sync_count INTEGER := 0;
BEGIN
    -- Loop through all active cash banks
    FOR cash_bank_record IN 
        SELECT id, account_id, balance, bank_name 
        FROM cash_banks 
        WHERE deleted_at IS NULL AND is_active = true
    LOOP
        -- Sync each cash bank with its COA account
        IF cash_bank_record.account_id IS NOT NULL THEN
            UPDATE accounts 
            SET 
                balance = cash_bank_record.balance,
                updated_at = NOW()
            WHERE id = cash_bank_record.account_id 
            AND deleted_at IS NULL;
            
            -- Update parent balances
            PERFORM update_parent_account_balances(cash_bank_record.account_id);
            sync_count := sync_count + 1;
        END IF;
    END LOOP;
    
    RETURN format('Synced %s cash bank accounts with their COA accounts', sync_count);
END;
$$ LANGUAGE plpgsql;

-- Function to ensure cash bank accounts are not header accounts
CREATE OR REPLACE FUNCTION ensure_cashbank_not_header()
RETURNS TEXT AS $$
DECLARE
    update_count INTEGER;
BEGIN
    -- Update cash bank linked accounts to ensure they are not header accounts
    UPDATE accounts 
    SET 
        is_header = false,
        updated_at = NOW()
    WHERE id IN (
        SELECT DISTINCT account_id 
        FROM cash_banks 
        WHERE account_id IS NOT NULL 
        AND deleted_at IS NULL
    )
    AND is_header = true
    AND deleted_at IS NULL;
    
    GET DIAGNOSTICS update_count = ROW_COUNT;
    
    RETURN format('Updated %s cash bank accounts to be non-header accounts', update_count);
END;
$$ LANGUAGE plpgsql;

-- =================================================================================
-- 6. Create optimized triggers with the unified approach
-- =================================================================================

-- Trigger for cash bank transaction changes (unified approach)
CREATE TRIGGER trigger_recalc_cashbank_balance_insert
    AFTER INSERT ON cash_bank_transactions
    FOR EACH ROW
    EXECUTE FUNCTION recalculate_cashbank_balance();

CREATE TRIGGER trigger_recalc_cashbank_balance_update
    AFTER UPDATE OF amount, cash_bank_id ON cash_bank_transactions
    FOR EACH ROW
    WHEN (OLD.amount IS DISTINCT FROM NEW.amount OR OLD.cash_bank_id IS DISTINCT FROM NEW.cash_bank_id)
    EXECUTE FUNCTION recalculate_cashbank_balance();

CREATE TRIGGER trigger_recalc_cashbank_balance_delete
    AFTER DELETE ON cash_bank_transactions
    FOR EACH ROW
    EXECUTE FUNCTION recalculate_cashbank_balance();

-- Trigger for account balance validation and parent rollup
CREATE TRIGGER trigger_validate_account_balance
    BEFORE UPDATE OF balance ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION validate_account_balance_consistency();

-- =================================================================================
-- 7. Create performance indexes
-- =================================================================================

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_accounts_parent_id_balance 
ON accounts (parent_id, balance) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_cash_bank_transactions_cash_bank_id_amount 
ON cash_bank_transactions (cash_bank_id, amount) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_cash_banks_account_id 
ON cash_banks (account_id) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_accounts_is_header 
ON accounts (is_header) 
WHERE deleted_at IS NULL;

-- =================================================================================
-- 8. Apply initial data fixes
-- =================================================================================

-- Ensure all cash bank linked accounts are not header accounts
SELECT ensure_cashbank_not_header();

-- Sync all existing cash bank balances with their COA accounts
SELECT manual_sync_all_cashbank_coa();

-- =================================================================================
-- 9. Add function documentation
-- =================================================================================

COMMENT ON FUNCTION update_parent_account_balances(BIGINT) IS 'Recursively updates parent account balances when child balances change (BIGINT compatible)';
COMMENT ON FUNCTION recalculate_cashbank_balance() IS 'Unified function to recalculate cash bank balance from transactions and sync with COA account';
COMMENT ON FUNCTION validate_account_balance_consistency() IS 'Validates and auto-corrects account balance consistency before updates';
COMMENT ON FUNCTION manual_sync_cashbank_coa(BIGINT) IS 'Manually syncs a specific cash bank balance with its COA account';
COMMENT ON FUNCTION manual_sync_all_cashbank_coa() IS 'Manually syncs all cash bank balances with their respective COA accounts';
COMMENT ON FUNCTION ensure_cashbank_not_header() IS 'Ensures all cash bank linked accounts are marked as non-header accounts';

-- =================================================================================
-- Migration completion log
-- =================================================================================

DO $$
BEGIN
    RAISE NOTICE 'Comprehensive Auto Balance Sync System installed successfully';
    RAISE NOTICE '✓ Unified trigger approach implemented';
    RAISE NOTICE '✓ BIGINT parameter compatibility ensured';
    RAISE NOTICE '✓ Cash bank to COA synchronization active';
    RAISE NOTICE '✓ Parent account balance rollup active';
    RAISE NOTICE '✓ Manual sync functions available';
    RAISE NOTICE '✓ Account hierarchy validation active';
    RAISE NOTICE '✓ Performance indexes created';
    RAISE NOTICE '✓ Initial data fixes applied';
END;
$$;