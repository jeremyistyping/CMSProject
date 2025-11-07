-- Migration: Prevent Shared COA Accounts
-- Purpose: Ensure each COA account can only be linked to one cash bank
-- This prevents balance conflicts and ensures data integrity

-- Step 1: Add unique partial index on account_id in cash_banks table
-- This ensures one COA account can only be used by one active cash bank
DROP INDEX IF EXISTS cash_banks_account_id_unique_idx;

CREATE UNIQUE INDEX cash_banks_account_id_unique_idx 
ON cash_banks (account_id) 
WHERE (deleted_at IS NULL AND account_id IS NOT NULL);

-- Step 2: Create function to validate cash_bank balance matches COA balance
CREATE OR REPLACE FUNCTION validate_cashbank_coa_balance()
RETURNS TRIGGER AS $$
DECLARE
    coa_balance DECIMAL(15,2);
    cashbank_balance DECIMAL(15,2);
BEGIN
    -- Only validate if account_id is set
    IF NEW.account_id IS NOT NULL THEN
        -- Get COA balance
        SELECT balance INTO coa_balance 
        FROM accounts 
        WHERE id = NEW.account_id AND deleted_at IS NULL;
        
        -- Get cash bank balance
        cashbank_balance := NEW.balance;
        
        -- Log warning if balances don't match (but don't block the operation)
        IF coa_balance IS NOT NULL AND ABS(coa_balance - cashbank_balance) > 0.01 THEN
            RAISE NOTICE 'WARNING: Cash Bank % balance (%) differs from COA account % balance (%)', 
                NEW.code, cashbank_balance, NEW.account_id, coa_balance;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 3: Create trigger to validate balance sync on cash_banks update
DROP TRIGGER IF EXISTS trg_validate_cashbank_coa_balance ON cash_banks;

CREATE TRIGGER trg_validate_cashbank_coa_balance
BEFORE UPDATE OF balance ON cash_banks
FOR EACH ROW
WHEN (NEW.balance IS DISTINCT FROM OLD.balance)
EXECUTE FUNCTION validate_cashbank_coa_balance();

-- Step 4: Create function to auto-sync COA balance when cash_bank balance changes
CREATE OR REPLACE FUNCTION sync_coa_balance_from_cashbank()
RETURNS TRIGGER AS $$
BEGIN
    -- Only sync if account_id is set and balance changed
    IF NEW.account_id IS NOT NULL AND NEW.balance IS DISTINCT FROM OLD.balance THEN
        -- Update COA account balance to match cash bank balance
        UPDATE accounts 
        SET balance = NEW.balance,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.account_id 
          AND deleted_at IS NULL;
        
        RAISE NOTICE 'Synced COA account % balance to %.2f (from cash bank %)', 
            NEW.account_id, NEW.balance, NEW.code;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 5: Create trigger to auto-sync COA balance
DROP TRIGGER IF EXISTS trg_sync_coa_balance_from_cashbank ON cash_banks;

CREATE TRIGGER trg_sync_coa_balance_from_cashbank
AFTER UPDATE OF balance ON cash_banks
FOR EACH ROW
WHEN (NEW.balance IS DISTINCT FROM OLD.balance AND NEW.account_id IS NOT NULL)
EXECUTE FUNCTION sync_coa_balance_from_cashbank();

-- Step 6: Create index for better performance
CREATE INDEX IF NOT EXISTS idx_cash_banks_account_id 
ON cash_banks(account_id) 
WHERE deleted_at IS NULL AND account_id IS NOT NULL;

-- Step 7: Add comments for documentation
COMMENT ON INDEX cash_banks_account_id_unique_idx IS 
'Ensures each COA account can only be linked to one active cash bank to prevent balance conflicts';

COMMENT ON FUNCTION sync_coa_balance_from_cashbank() IS 
'Automatically syncs COA account balance when cash bank balance changes to maintain consistency';

COMMENT ON FUNCTION validate_cashbank_coa_balance() IS 
'Validates and logs warnings when cash bank balance differs from linked COA account balance';
