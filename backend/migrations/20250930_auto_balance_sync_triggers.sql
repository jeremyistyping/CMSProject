-- Migration: Auto Balance Sync Triggers
-- Description: Creates database triggers to automatically sync cash bank balances with COA accounts
-- Date: 2025-09-30

-- =================================================================================
-- 1. Function to sync cash bank balance with linked COA account
-- =================================================================================

CREATE OR REPLACE FUNCTION sync_cashbank_coa_balance()
RETURNS TRIGGER AS $$
BEGIN
    -- Update the linked COA account balance to match cash bank balance
    UPDATE accounts 
    SET 
        balance = NEW.balance,
        updated_at = NOW()
    WHERE id = NEW.account_id 
    AND deleted_at IS NULL;
    
    -- Update parent account balances recursively
    PERFORM update_parent_account_balances(NEW.account_id);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =================================================================================
-- 2. Function to recalculate cash bank balance from transactions
-- =================================================================================

CREATE OR REPLACE FUNCTION recalculate_cashbank_balance()
RETURNS TRIGGER AS $$
DECLARE
    cash_bank_id_val INTEGER;
    new_balance DECIMAL(15,2);
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
    
    -- Update cash bank balance
    UPDATE cash_banks 
    SET 
        balance = new_balance,
        updated_at = NOW()
    WHERE id = cash_bank_id_val 
    AND deleted_at IS NULL;
    
    -- Sync with COA account (will trigger the sync function above)
    UPDATE cash_banks 
    SET updated_at = NOW()
    WHERE id = cash_bank_id_val;
    
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- =================================================================================
-- 3. Function to update parent account balances recursively
-- =================================================================================

CREATE OR REPLACE FUNCTION update_parent_account_balances(child_account_id INTEGER)
RETURNS VOID AS $$
DECLARE
    parent_id INTEGER;
    parent_balance DECIMAL(15,2);
    current_parent_balance DECIMAL(15,2);
BEGIN
    -- Get the parent of the current account
    SELECT parent_id INTO parent_id
    FROM accounts 
    WHERE id = child_account_id 
    AND deleted_at IS NULL;
    
    -- If there's a parent, update its balance
    IF parent_id IS NOT NULL THEN
        -- Get current parent balance for comparison
        SELECT balance INTO current_parent_balance
        FROM accounts 
        WHERE id = parent_id 
        AND deleted_at IS NULL;
        
        -- Calculate sum of all children balances
        SELECT COALESCE(SUM(balance), 0)
        INTO parent_balance
        FROM accounts 
        WHERE parent_id = parent_id 
        AND deleted_at IS NULL;
        
        -- Update parent balance only if it's different
        IF ABS(current_parent_balance - parent_balance) > 0.01 THEN
            UPDATE accounts 
            SET 
                balance = parent_balance,
                updated_at = NOW()
            WHERE id = parent_id 
            AND deleted_at IS NULL;
            
            -- Log the update
            RAISE NOTICE 'Updated parent account % balance: % → %', parent_id, current_parent_balance, parent_balance;
        END IF;
        
        -- Recursively update parent's parent
        PERFORM update_parent_account_balances(parent_id);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- =================================================================================
-- 4. Function to validate account balance consistency
-- =================================================================================

CREATE OR REPLACE FUNCTION validate_account_balance_consistency()
RETURNS TRIGGER AS $$
DECLARE
    calculated_balance DECIMAL(15,2);
    parent_id_val INTEGER;
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
            RAISE NOTICE 'Auto-corrected header account % balance: % → %', NEW.id, OLD.balance, calculated_balance;
        END IF;
    END IF;
    
    -- Update parent balances when a child balance changes
    IF OLD.balance IS DISTINCT FROM NEW.balance THEN
        PERFORM update_parent_account_balances(NEW.id);
        RAISE NOTICE 'Triggered parent balance update for account % (balance changed from % to %)', NEW.id, OLD.balance, NEW.balance;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =================================================================================
-- 5. Create triggers
-- =================================================================================

-- Trigger for cash bank balance updates
DROP TRIGGER IF EXISTS trigger_sync_cashbank_coa ON cash_banks;
CREATE TRIGGER trigger_sync_cashbank_coa
    AFTER UPDATE OF balance ON cash_banks
    FOR EACH ROW
    WHEN (OLD.balance IS DISTINCT FROM NEW.balance)
    EXECUTE FUNCTION sync_cashbank_coa_balance();

-- Trigger for cash bank transaction changes
DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_insert ON cash_bank_transactions;
CREATE TRIGGER trigger_recalc_cashbank_balance_insert
    AFTER INSERT ON cash_bank_transactions
    FOR EACH ROW
    EXECUTE FUNCTION recalculate_cashbank_balance();

DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_update ON cash_bank_transactions;
CREATE TRIGGER trigger_recalc_cashbank_balance_update
    AFTER UPDATE OF amount, cash_bank_id ON cash_bank_transactions
    FOR EACH ROW
    WHEN (OLD.amount IS DISTINCT FROM NEW.amount OR OLD.cash_bank_id IS DISTINCT FROM NEW.cash_bank_id)
    EXECUTE FUNCTION recalculate_cashbank_balance();

DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_delete ON cash_bank_transactions;
CREATE TRIGGER trigger_recalc_cashbank_balance_delete
    AFTER DELETE ON cash_bank_transactions
    FOR EACH ROW
    EXECUTE FUNCTION recalculate_cashbank_balance();

-- Trigger for account balance validation and parent rollup
DROP TRIGGER IF EXISTS trigger_validate_account_balance ON accounts;
CREATE TRIGGER trigger_validate_account_balance
    BEFORE UPDATE OF balance ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION validate_account_balance_consistency();

-- =================================================================================
-- 6. Create indexes for better performance
-- =================================================================================

CREATE INDEX IF NOT EXISTS idx_accounts_parent_id_balance 
ON accounts (parent_id, balance) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_cash_bank_transactions_cash_bank_id_amount 
ON cash_bank_transactions (cash_bank_id, amount) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_cash_banks_account_id 
ON cash_banks (account_id) 
WHERE deleted_at IS NULL;

-- =================================================================================
-- Comments and documentation
-- =================================================================================

COMMENT ON FUNCTION sync_cashbank_coa_balance() IS 'Automatically syncs cash bank balance changes with linked COA account';
COMMENT ON FUNCTION recalculate_cashbank_balance() IS 'Recalculates cash bank balance from transactions whenever transactions change';
COMMENT ON FUNCTION update_parent_account_balances(INTEGER) IS 'Recursively updates parent account balances when child balances change';
COMMENT ON FUNCTION validate_account_balance_consistency() IS 'Validates and auto-corrects account balance consistency before updates';