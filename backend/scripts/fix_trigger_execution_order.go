package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Fixing Trigger Execution Order...")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Fix the trigger execution order
	if err := fixTriggerExecutionOrder(db); err != nil {
		log.Fatalf("❌ Failed to fix trigger execution order: %v", err)
	}

	log.Println("🎉 Trigger execution order fixed successfully!")
}

func fixTriggerExecutionOrder(db *gorm.DB) error {
	log.Println("📝 Fixing trigger execution order...")

	// SQL commands to fix the trigger execution order
	sqlCommands := []string{
		// Drop the problematic cash bank balance sync trigger
		`DROP TRIGGER IF EXISTS trigger_sync_cashbank_coa ON cash_banks;`,

		// Modify the recalculate_cashbank_balance function to handle both cash bank and COA updates
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
		    
		    -- Update linked COA account balance directly (avoid cascading triggers)
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
		$$ LANGUAGE plpgsql;`,

		// Create a simplified manual sync function for cash banks
		`CREATE OR REPLACE FUNCTION manual_sync_cashbank_coa(target_cash_bank_id BIGINT)
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
		$$ LANGUAGE plpgsql;`,

		// Update comments
		`COMMENT ON FUNCTION recalculate_cashbank_balance() IS 'Recalculates cash bank balance from transactions and syncs with COA account (unified approach)';`,
		`COMMENT ON FUNCTION manual_sync_cashbank_coa(BIGINT) IS 'Manually syncs a specific cash bank balance with its COA account';`,
	}

	// Execute each SQL command
	for i, sqlCmd := range sqlCommands {
		log.Printf("  Executing SQL command %d/%d...", i+1, len(sqlCommands))
		
		if err := db.Exec(sqlCmd).Error; err != nil {
			return fmt.Errorf("failed to execute SQL command %d: %w\nSQL: %s", i+1, err, sqlCmd)
		}
	}

	log.Println("✅ Trigger execution order fixed successfully")

	// Test the unified trigger approach
	if err := testUnifiedTriggerApproach(db); err != nil {
		return fmt.Errorf("unified trigger test failed: %w", err)
	}

	return nil
}

func testUnifiedTriggerApproach(db *gorm.DB) error {
	log.Println("🧪 Testing unified trigger approach...")

	// Check if we have any active cash banks
	var cashBankCount int64
	if err := db.Raw("SELECT COUNT(*) FROM cash_banks WHERE deleted_at IS NULL AND is_active = true").Scan(&cashBankCount).Error; err != nil {
		return fmt.Errorf("failed to count cash banks: %w", err)
	}

	log.Printf("Found %d active cash banks", cashBankCount)

	if cashBankCount == 0 {
		log.Println("ℹ️ No active cash banks found - skipping trigger test")
		return nil
	}

	// Test manual sync function
	var testCashBankID uint
	if err := db.Raw("SELECT id FROM cash_banks WHERE deleted_at IS NULL AND is_active = true LIMIT 1").Scan(&testCashBankID).Error; err != nil {
		return fmt.Errorf("failed to get test cash bank ID: %w", err)
	}

	log.Printf("Testing manual sync for cash bank ID: %d", testCashBankID)

	if err := db.Exec("SELECT manual_sync_cashbank_coa(?)", testCashBankID).Error; err != nil {
		return fmt.Errorf("manual sync function failed: %w", err)
	}

	log.Println("✅ Manual sync function works correctly")

	return nil
}