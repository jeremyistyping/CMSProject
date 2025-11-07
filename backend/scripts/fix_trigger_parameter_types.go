package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Fixing Database Function Parameter Types...")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Fix the function parameter types
	if err := fixParameterTypes(db); err != nil {
		log.Fatalf("❌ Failed to fix parameter types: %v", err)
	}

	log.Println("🎉 Database function parameter types fixed successfully!")
}

func fixParameterTypes(db *gorm.DB) error {
	log.Println("📝 Updating database function parameter types...")

	// SQL commands to fix the parameter types
	sqlCommands := []string{
		// Drop existing function first
		`DROP FUNCTION IF EXISTS update_parent_account_balances(INTEGER);`,
		`DROP FUNCTION IF EXISTS update_parent_account_balances(BIGINT);`,
		
		// Recreate with correct BIGINT parameter type
		`CREATE OR REPLACE FUNCTION update_parent_account_balances(child_account_id BIGINT)
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
		$$ LANGUAGE plpgsql;`,

		// Also fix the sync function to handle BIGINT properly
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
		    
		    -- Update parent account balances recursively (cast to BIGINT)
		    PERFORM update_parent_account_balances(NEW.account_id::BIGINT);
		    
		    RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		// Fix the recalculate balance function
		`CREATE OR REPLACE FUNCTION recalculate_cashbank_balance()
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
		    
		    -- Update cash bank balance and get linked account ID
		    UPDATE cash_banks 
		    SET 
		        balance = new_balance,
		        updated_at = NOW()
		    WHERE id = cash_bank_id_val 
		    AND deleted_at IS NULL
		    RETURNING account_id INTO linked_account_id;
		    
		    -- Update linked COA account balance
		    IF linked_account_id IS NOT NULL THEN
		        UPDATE accounts 
		        SET 
		            balance = new_balance,
		            updated_at = NOW()
		        WHERE id = linked_account_id 
		        AND deleted_at IS NULL;
		        
		        -- Update parent balances
		        PERFORM update_parent_account_balances(linked_account_id);
		    END IF;
		    
		    IF TG_OP = 'DELETE' THEN
		        RETURN OLD;
		    ELSE
		        RETURN NEW;
		    END IF;
		END;
		$$ LANGUAGE plpgsql;`,

		// Also fix the account validation function to use BIGINT
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
		        PERFORM update_parent_account_balances(NEW.id::BIGINT);
		    END IF;
		    
		    RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		// Update the comment
		`COMMENT ON FUNCTION update_parent_account_balances(BIGINT) IS 'Recursively updates parent account balances when child balances change (fixed BIGINT parameter)';`,
	}

	// Execute each SQL command
	for i, sqlCmd := range sqlCommands {
		log.Printf("  Executing SQL command %d/%d...", i+1, len(sqlCommands))
		
		if err := db.Exec(sqlCmd).Error; err != nil {
			return fmt.Errorf("failed to execute SQL command %d: %w\nSQL: %s", i+1, err, sqlCmd)
		}
	}

	log.Println("✅ All database function parameter types fixed successfully")
	return nil
}