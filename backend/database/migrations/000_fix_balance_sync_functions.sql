-- Fix missing balance sync functions
-- Some triggers reference functions that don't exist

DO $$
BEGIN
    -- Check if recalculate_cashbank_balance function exists
    IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'recalculate_cashbank_balance') THEN
        RAISE NOTICE '🔧 Creating missing recalculate_cashbank_balance function';
        
        -- Create the function
        CREATE OR REPLACE FUNCTION recalculate_cashbank_balance()
        RETURNS TRIGGER AS $func$
        DECLARE
            new_balance DECIMAL(20,2);
        BEGIN
            -- Calculate balance from all transactions
            SELECT COALESCE(SUM(amount), 0) INTO new_balance
            FROM cash_bank_transactions
            WHERE cash_bank_id = COALESCE(NEW.cash_bank_id, OLD.cash_bank_id)
              AND deleted_at IS NULL;
            
            -- Update cash_banks balance
            UPDATE cash_banks
            SET balance = new_balance,
                updated_at = NOW()
            WHERE id = COALESCE(NEW.cash_bank_id, OLD.cash_bank_id);
            
            RETURN COALESCE(NEW, OLD);
        END;
        $func$ LANGUAGE plpgsql;
        
        RAISE NOTICE '✅ Created recalculate_cashbank_balance function';
    ELSE
        RAISE NOTICE '✅ Function recalculate_cashbank_balance already exists';
    END IF;

    -- Check if update_parent_account_balances function exists
    IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_parent_account_balances') THEN
        RAISE NOTICE '🔧 Creating missing update_parent_account_balances function';
        
        -- Create the function
        CREATE OR REPLACE FUNCTION update_parent_account_balances(account_id_param BIGINT)
        RETURNS VOID AS $func$
        DECLARE
            parent_id_val BIGINT;
            parent_balance DECIMAL(20,2);
        BEGIN
            -- Get parent account id
            SELECT parent_id INTO parent_id_val
            FROM accounts
            WHERE id = account_id_param;
            
            -- If no parent, exit
            IF parent_id_val IS NULL THEN
                RETURN;
            END IF;
            
            -- Calculate parent balance from all children
            SELECT COALESCE(SUM(balance), 0) INTO parent_balance
            FROM accounts
            WHERE parent_id = parent_id_val
              AND deleted_at IS NULL;
            
            -- Update parent balance
            UPDATE accounts
            SET balance = parent_balance,
                updated_at = NOW()
            WHERE id = parent_id_val;
            
            -- Recursively update grandparent
            PERFORM update_parent_account_balances(parent_id_val);
        END;
        $func$ LANGUAGE plpgsql;
        
        RAISE NOTICE '✅ Created update_parent_account_balances function';
    ELSE
        RAISE NOTICE '✅ Function update_parent_account_balances already exists';
    END IF;

    RAISE NOTICE '🎯 Balance sync functions verified/created';
END $$;
