package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Applying Auto Balance Sync Database Triggers...")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Apply the triggers
	if err := applyAutoBalanceTriggers(db); err != nil {
		log.Fatalf("❌ Failed to apply triggers: %v", err)
	}

	log.Println("🎉 Auto balance sync triggers applied successfully!")
}

func applyAutoBalanceTriggers(db *gorm.DB) error {
	log.Println("📝 Applying database functions and triggers...")

	// SQL commands to create the triggers
	sqlCommands := []string{
		// 1. Function to sync cash bank balance with linked COA account
		`CREATE OR REPLACE FUNCTION sync_cashbank_coa_balance()
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
		$$ LANGUAGE plpgsql;`,

		// 2. Function to recalculate cash bank balance from transactions
		`CREATE OR REPLACE FUNCTION recalculate_cashbank_balance()
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
		    
		    IF TG_OP = 'DELETE' THEN
		        RETURN OLD;
		    ELSE
		        RETURN NEW;
		    END IF;
		END;
		$$ LANGUAGE plpgsql;`,

		// 3. Function to update parent account balances recursively
		`CREATE OR REPLACE FUNCTION update_parent_account_balances(child_account_id INTEGER)
		RETURNS VOID AS $$
		DECLARE
		    parent_account_id INTEGER;
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
		$$ LANGUAGE plpgsql;`,

		// 4. Function to validate account balance consistency
		`CREATE OR REPLACE FUNCTION validate_account_balance_consistency()
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
		        PERFORM update_parent_account_balances(NEW.id);
		    END IF;
		    
		    RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		// Drop existing triggers first
		`DROP TRIGGER IF EXISTS trigger_sync_cashbank_coa ON cash_banks;`,
		`DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_insert ON cash_bank_transactions;`,
		`DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_update ON cash_bank_transactions;`,
		`DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_delete ON cash_bank_transactions;`,
		`DROP TRIGGER IF EXISTS trigger_validate_account_balance ON accounts;`,

		// 5. Create triggers
		`CREATE TRIGGER trigger_sync_cashbank_coa
		    AFTER UPDATE OF balance ON cash_banks
		    FOR EACH ROW
		    WHEN (OLD.balance IS DISTINCT FROM NEW.balance)
		    EXECUTE FUNCTION sync_cashbank_coa_balance();`,

		`CREATE TRIGGER trigger_recalc_cashbank_balance_insert
		    AFTER INSERT ON cash_bank_transactions
		    FOR EACH ROW
		    EXECUTE FUNCTION recalculate_cashbank_balance();`,

		`CREATE TRIGGER trigger_recalc_cashbank_balance_update
		    AFTER UPDATE OF amount, cash_bank_id ON cash_bank_transactions
		    FOR EACH ROW
		    WHEN (OLD.amount IS DISTINCT FROM NEW.amount OR OLD.cash_bank_id IS DISTINCT FROM NEW.cash_bank_id)
		    EXECUTE FUNCTION recalculate_cashbank_balance();`,

		`CREATE TRIGGER trigger_recalc_cashbank_balance_delete
		    AFTER DELETE ON cash_bank_transactions
		    FOR EACH ROW
		    EXECUTE FUNCTION recalculate_cashbank_balance();`,

		`CREATE TRIGGER trigger_validate_account_balance
		    BEFORE UPDATE OF balance ON accounts
		    FOR EACH ROW
		    EXECUTE FUNCTION validate_account_balance_consistency();`,

		// 6. Create indexes for better performance
		`CREATE INDEX IF NOT EXISTS idx_accounts_parent_id_balance 
		ON accounts (parent_id, balance) 
		WHERE deleted_at IS NULL;`,

		`CREATE INDEX IF NOT EXISTS idx_cash_bank_transactions_cash_bank_id_amount 
		ON cash_bank_transactions (cash_bank_id, amount) 
		WHERE deleted_at IS NULL;`,

		`CREATE INDEX IF NOT EXISTS idx_cash_banks_account_id 
		ON cash_banks (account_id) 
		WHERE deleted_at IS NULL;`,

		// Add comments
		`COMMENT ON FUNCTION sync_cashbank_coa_balance() IS 'Automatically syncs cash bank balance changes with linked COA account';`,
		`COMMENT ON FUNCTION recalculate_cashbank_balance() IS 'Recalculates cash bank balance from transactions whenever transactions change';`,
		`COMMENT ON FUNCTION update_parent_account_balances(INTEGER) IS 'Recursively updates parent account balances when child balances change';`,
		`COMMENT ON FUNCTION validate_account_balance_consistency() IS 'Validates and auto-corrects account balance consistency before updates';`,
	}

	// Execute each SQL command
	for i, sqlCmd := range sqlCommands {
		log.Printf("  Executing SQL command %d/%d...", i+1, len(sqlCommands))
		
		if err := db.Exec(sqlCmd).Error; err != nil {
			return fmt.Errorf("failed to execute SQL command %d: %w\nSQL: %s", i+1, err, sqlCmd)
		}
	}

	log.Println("✅ All database functions and triggers created successfully")
	return nil
}

// Test the triggers by performing some operations
func testTriggers(db *gorm.DB) error {
	log.Println("🧪 Testing auto balance sync triggers...")

	// Test will be performed by the actual application usage
	// For now, just validate that triggers exist
	
	var triggerCount int64
	if err := db.Raw(`
		SELECT COUNT(*) 
		FROM pg_trigger 
		WHERE tgname IN (
			'trigger_sync_cashbank_coa',
			'trigger_recalc_cashbank_balance_insert',
			'trigger_recalc_cashbank_balance_update', 
			'trigger_recalc_cashbank_balance_delete',
			'trigger_validate_account_balance'
		)
	`).Scan(&triggerCount).Error; err != nil {
		return fmt.Errorf("failed to validate triggers: %w", err)
	}

	log.Printf("✅ Found %d auto balance triggers in database", triggerCount)
	
	if triggerCount != 5 {
		return fmt.Errorf("expected 5 triggers, found %d", triggerCount)
	}

	return nil
}