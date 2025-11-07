package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	// Load configuration
	_ = config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	fmt.Println("🔧 Running Migration: Prevent Shared COA Accounts")
	fmt.Println("=" + string(make([]byte, 60)))

	// Execute migration statements one by one
	fmt.Println("\n📋 Executing migration...")
	
	// Step 1: Create unique index
	fmt.Println("  Step 1: Creating unique index...")
	if err := db.Exec("DROP INDEX IF EXISTS cash_banks_account_id_unique_idx").Error; err != nil {
		log.Printf("⚠️ Warning dropping index: %v", err)
	}
	if err := db.Exec(`
		CREATE UNIQUE INDEX cash_banks_account_id_unique_idx 
		ON cash_banks (account_id) 
		WHERE (deleted_at IS NULL AND account_id IS NOT NULL)
	`).Error; err != nil {
		log.Fatalf("❌ Failed to create unique index: %v", err)
	}
	fmt.Println("    ✅ Unique index created")
	
	// Step 2: Create validation function
	fmt.Println("  Step 2: Creating validation function...")
	if err := db.Exec(`
		CREATE OR REPLACE FUNCTION validate_cashbank_coa_balance()
		RETURNS TRIGGER AS $$
		DECLARE
		    coa_balance DECIMAL(15,2);
		    cashbank_balance DECIMAL(15,2);
		BEGIN
		    IF NEW.account_id IS NOT NULL THEN
		        SELECT balance INTO coa_balance 
		        FROM accounts 
		        WHERE id = NEW.account_id AND deleted_at IS NULL;
		        
		        cashbank_balance := NEW.balance;
		        
		        IF coa_balance IS NOT NULL AND ABS(coa_balance - cashbank_balance) > 0.01 THEN
		            RAISE NOTICE 'WARNING: Cash Bank % balance (%) differs from COA account % balance (%)', 
		                NEW.code, cashbank_balance, NEW.account_id, coa_balance;
		        END IF;
		    END IF;
		    
		    RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`).Error; err != nil {
		log.Fatalf("❌ Failed to create validation function: %v", err)
	}
	fmt.Println("    ✅ Validation function created")
	
	// Step 3: Create validation trigger
	fmt.Println("  Step 3: Creating validation trigger...")
	db.Exec("DROP TRIGGER IF EXISTS trg_validate_cashbank_coa_balance ON cash_banks")
	if err := db.Exec(`
		CREATE TRIGGER trg_validate_cashbank_coa_balance
		BEFORE UPDATE OF balance ON cash_banks
		FOR EACH ROW
		WHEN (NEW.balance IS DISTINCT FROM OLD.balance)
		EXECUTE FUNCTION validate_cashbank_coa_balance()
	`).Error; err != nil {
		log.Fatalf("❌ Failed to create validation trigger: %v", err)
	}
	fmt.Println("    ✅ Validation trigger created")
	
	// Step 4: Create sync function
	fmt.Println("  Step 4: Creating auto-sync function...")
	if err := db.Exec(`
		CREATE OR REPLACE FUNCTION sync_coa_balance_from_cashbank()
		RETURNS TRIGGER AS $$
		BEGIN
		    IF NEW.account_id IS NOT NULL AND NEW.balance IS DISTINCT FROM OLD.balance THEN
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
		$$ LANGUAGE plpgsql
	`).Error; err != nil {
		log.Fatalf("❌ Failed to create sync function: %v", err)
	}
	fmt.Println("    ✅ Auto-sync function created")
	
	// Step 5: Create sync trigger
	fmt.Println("  Step 5: Creating auto-sync trigger...")
	db.Exec("DROP TRIGGER IF EXISTS trg_sync_coa_balance_from_cashbank ON cash_banks")
	if err := db.Exec(`
		CREATE TRIGGER trg_sync_coa_balance_from_cashbank
		AFTER UPDATE OF balance ON cash_banks
		FOR EACH ROW
		WHEN (NEW.balance IS DISTINCT FROM OLD.balance AND NEW.account_id IS NOT NULL)
		EXECUTE FUNCTION sync_coa_balance_from_cashbank()
	`).Error; err != nil {
		log.Fatalf("❌ Failed to create sync trigger: %v", err)
	}
	fmt.Println("    ✅ Auto-sync trigger created")
	
	// Step 6: Create performance index
	fmt.Println("  Step 6: Creating performance index...")
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_cash_banks_account_id 
		ON cash_banks(account_id) 
		WHERE deleted_at IS NULL AND account_id IS NOT NULL
	`).Error; err != nil {
		log.Printf("⚠️ Warning creating performance index: %v", err)
	}
	fmt.Println("    ✅ Performance index created")

	fmt.Println("✅ Migration executed successfully!")
	
	// Verify constraints and triggers
	fmt.Println("\n🔍 Verifying migration...")
	
	// Check unique index
	var indexExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE indexname = 'cash_banks_account_id_unique_idx'
		)
	`).Scan(&indexExists)
	
	if indexExists {
		fmt.Println("  ✅ Unique index 'cash_banks_account_id_unique_idx' created")
	} else {
		fmt.Println("  ❌ Unique index not found")
	}
	
	// Check triggers
	var syncTriggerExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_trigger 
			WHERE tgname = 'trg_sync_coa_balance_from_cashbank'
		)
	`).Scan(&syncTriggerExists)
	
	if syncTriggerExists {
		fmt.Println("  ✅ Auto-sync trigger 'trg_sync_coa_balance_from_cashbank' created")
	} else {
		fmt.Println("  ❌ Auto-sync trigger not found")
	}
	
	var validateTriggerExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_trigger 
			WHERE tgname = 'trg_validate_cashbank_coa_balance'
		)
	`).Scan(&validateTriggerExists)
	
	if validateTriggerExists {
		fmt.Println("  ✅ Validation trigger 'trg_validate_cashbank_coa_balance' created")
	} else {
		fmt.Println("  ❌ Validation trigger not found")
	}
	
	fmt.Println("\n🎉 Migration completed successfully!")
	fmt.Println("\nNow:")
	fmt.Println("  1. Each cash bank can only link to one unique COA account")
	fmt.Println("  2. Cash bank balance changes will auto-sync to COA")
	fmt.Println("  3. Warnings will be logged if balances drift out of sync")
}
