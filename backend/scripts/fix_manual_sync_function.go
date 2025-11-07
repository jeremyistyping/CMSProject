package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Fixing Manual Sync Function...")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Fix the manual sync function
	if err := fixManualSyncFunction(db); err != nil {
		log.Fatalf("❌ Failed to fix manual sync function: %v", err)
	}

	log.Println("🎉 Manual sync function fixed successfully!")
}

func fixManualSyncFunction(db *gorm.DB) error {
	log.Println("📝 Fixing manual sync function...")

	// Drop existing function and recreate with improved logic
	sqlCommands := []string{
		// Drop existing manual sync function
		`DROP FUNCTION IF EXISTS manual_sync_cashbank_coa(BIGINT);`,

		// Create improved manual sync function with debugging
		`CREATE OR REPLACE FUNCTION manual_sync_cashbank_coa(target_cash_bank_id BIGINT)
		RETURNS VOID AS $$
		DECLARE
		    cash_bank_balance DECIMAL(15,2);
		    linked_account_id BIGINT;
		    calculated_balance DECIMAL(15,2);
		    rows_affected INTEGER;
		BEGIN
		    -- Get cash bank balance and linked account
		    SELECT balance, account_id 
		    INTO cash_bank_balance, linked_account_id
		    FROM cash_banks 
		    WHERE id = target_cash_bank_id 
		    AND deleted_at IS NULL;
		    
		    -- Verify we found the cash bank
		    IF linked_account_id IS NULL THEN
		        RAISE NOTICE 'Cash bank ID % not found or has no linked account', target_cash_bank_id;
		        RETURN;
		    END IF;
		    
		    -- Calculate balance from transactions (SSOT)
		    SELECT COALESCE(SUM(amount), 0)
		    INTO calculated_balance
		    FROM cash_bank_transactions 
		    WHERE cash_bank_id = target_cash_bank_id 
		    AND deleted_at IS NULL;
		    
		    -- Update cash bank balance first
		    UPDATE cash_banks 
		    SET 
		        balance = calculated_balance,
		        updated_at = NOW()
		    WHERE id = target_cash_bank_id 
		    AND deleted_at IS NULL;
		    
		    GET DIAGNOSTICS rows_affected = ROW_COUNT;
		    RAISE NOTICE 'Updated cash bank ID % balance to % (rows affected: %)', 
		                 target_cash_bank_id, calculated_balance, rows_affected;
		    
		    -- Update linked COA account balance
		    UPDATE accounts 
		    SET 
		        balance = calculated_balance,
		        updated_at = NOW()
		    WHERE id = linked_account_id 
		    AND deleted_at IS NULL;
		    
		    GET DIAGNOSTICS rows_affected = ROW_COUNT;
		    RAISE NOTICE 'Updated account ID % balance to % (rows affected: %)', 
		                 linked_account_id, calculated_balance, rows_affected;
		    
		    -- Update parent balances recursively
		    PERFORM update_parent_account_balances(linked_account_id);
		    
		END;
		$$ LANGUAGE plpgsql;`,

		// Also create a direct sync function that uses the current cash bank balance
		`CREATE OR REPLACE FUNCTION direct_sync_cashbank_coa(target_cash_bank_id BIGINT)
		RETURNS VOID AS $$
		DECLARE
		    cash_bank_balance DECIMAL(15,2);
		    linked_account_id BIGINT;
		    rows_affected INTEGER;
		BEGIN
		    -- Get cash bank balance and linked account
		    SELECT balance, account_id 
		    INTO cash_bank_balance, linked_account_id
		    FROM cash_banks 
		    WHERE id = target_cash_bank_id 
		    AND deleted_at IS NULL;
		    
		    -- Verify we found the cash bank
		    IF linked_account_id IS NULL THEN
		        RAISE NOTICE 'Cash bank ID % not found or has no linked account', target_cash_bank_id;
		        RETURN;
		    END IF;
		    
		    -- Update linked COA account balance to match cash bank
		    UPDATE accounts 
		    SET 
		        balance = cash_bank_balance,
		        updated_at = NOW()
		    WHERE id = linked_account_id 
		    AND deleted_at IS NULL;
		    
		    GET DIAGNOSTICS rows_affected = ROW_COUNT;
		    RAISE NOTICE 'Direct sync: Updated account ID % balance to % (rows affected: %)', 
		                 linked_account_id, cash_bank_balance, rows_affected;
		    
		    -- Update parent balances recursively
		    PERFORM update_parent_account_balances(linked_account_id);
		    
		END;
		$$ LANGUAGE plpgsql;`,

		// Update comments
		`COMMENT ON FUNCTION manual_sync_cashbank_coa(BIGINT) IS 'Manually syncs cash bank balance with COA account using transaction-calculated balance (SSOT)';`,
		`COMMENT ON FUNCTION direct_sync_cashbank_coa(BIGINT) IS 'Directly syncs COA account balance to match current cash bank balance';`,
	}

	// Execute each SQL command
	for i, sqlCmd := range sqlCommands {
		log.Printf("  Executing SQL command %d/%d...", i+1, len(sqlCommands))
		
		if err := db.Exec(sqlCmd).Error; err != nil {
			return fmt.Errorf("failed to execute SQL command %d: %w\nSQL: %s", i+1, err, sqlCmd)
		}
	}

	log.Println("✅ Manual sync function fixed successfully")

	// Test the fixed function
	if err := testFixedSyncFunction(db); err != nil {
		return fmt.Errorf("sync function test failed: %w", err)
	}

	return nil
}

func testFixedSyncFunction(db *gorm.DB) error {
	log.Println("🧪 Testing fixed sync function...")

	// Get the problematic cash banks
	problematicCashBanks := []string{"CSH-2025-0001", "BNK-2025-0003"}

	for _, code := range problematicCashBanks {
		log.Printf("  Testing sync for %s...", code)

		// Get cash bank ID
		var cashBankID uint
		if err := db.Raw("SELECT id FROM cash_banks WHERE code = ? AND deleted_at IS NULL", code).Scan(&cashBankID).Error; err != nil {
			log.Printf("    ❌ Failed to find cash bank %s: %v", code, err)
			continue
		}

		// Test direct sync first
		if err := db.Exec("SELECT direct_sync_cashbank_coa(?)", cashBankID).Error; err != nil {
			log.Printf("    ❌ Direct sync failed for %s: %v", code, err)
		} else {
			log.Printf("    ✅ Direct sync completed for %s", code)
		}

		// Check the result
		balanceQuery := `
			SELECT 
				cb.balance as cash_balance,
				a.balance as coa_balance
			FROM cash_banks cb
			JOIN accounts a ON cb.account_id = a.id
			WHERE cb.code = ? AND cb.deleted_at IS NULL
		`

		var result struct {
			CashBalance float64 `json:"cash_balance"`
			COABalance  float64 `json:"coa_balance"`
		}

		if err := db.Raw(balanceQuery, code).Scan(&result).Error; err != nil {
			log.Printf("    ❌ Failed to check sync result for %s: %v", code, err)
			continue
		}

		log.Printf("    Result: Cash=%.2f, COA=%.2f", result.CashBalance, result.COABalance)

		if result.CashBalance == result.COABalance {
			log.Printf("    ✅ %s is now in sync", code)
		} else {
			log.Printf("    ⚠️ %s still not in sync (difference: %.2f)", code, result.CashBalance - result.COABalance)
		}
	}

	return nil
}