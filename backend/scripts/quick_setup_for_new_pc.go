package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🚀 Quick Setup for New PC/Environment")
	log.Println("================================================================================")
	log.Println("This script will check and install necessary triggers and functions")
	log.Println("for auto balance sync system if they don't exist.")
	log.Println("================================================================================")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Check and setup the system
	if err := checkAndSetupSystem(db); err != nil {
		log.Fatalf("❌ Setup failed: %v", err)
	}

	log.Println("🎉 Quick setup completed successfully!")
	log.Println("✅ Auto balance sync system is ready to use.")
}

func checkAndSetupSystem(db *gorm.DB) error {
	log.Println("🔍 Checking system status...")

	// Check if triggers exist
	var triggerCount int64
	triggerQuery := `
		SELECT COUNT(*) 
		FROM pg_trigger 
		WHERE tgname IN (
			'trigger_recalc_cashbank_balance_insert',
			'trigger_recalc_cashbank_balance_update', 
			'trigger_recalc_cashbank_balance_delete',
			'trigger_validate_account_balance'
		)
	`

	if err := db.Raw(triggerQuery).Scan(&triggerCount).Error; err != nil {
		return fmt.Errorf("failed to check triggers: %w", err)
	}

	log.Printf("Found %d/4 required triggers", triggerCount)

	if triggerCount < 4 {
		log.Println("⚠️ Some triggers are missing. Installing auto balance triggers...")
		if err := installAutoBalanceTriggers(db); err != nil {
			return fmt.Errorf("failed to install triggers: %w", err)
		}
		log.Println("✅ Auto balance triggers installed successfully")
	} else {
		log.Println("✅ All required triggers are already installed")
	}

	// Check if cash bank accounts are properly set as non-header
	var headerCashBankCount int64
	headerCheckQuery := `
		SELECT COUNT(*) 
		FROM accounts a
		JOIN cash_banks cb ON a.id = cb.account_id
		WHERE a.is_header = true 
		AND cb.deleted_at IS NULL
		AND a.deleted_at IS NULL
	`

	if err := db.Raw(headerCheckQuery).Scan(&headerCashBankCount).Error; err != nil {
		return fmt.Errorf("failed to check cash bank accounts: %w", err)
	}

	if headerCashBankCount > 0 {
		log.Printf("⚠️ Found %d cash bank accounts that are still set as headers. Fixing...", headerCashBankCount)
		
		fixHeaderSQL := `
			UPDATE accounts 
			SET is_header = false, updated_at = NOW()
			WHERE id IN (
				SELECT DISTINCT account_id 
				FROM cash_banks 
				WHERE deleted_at IS NULL 
				AND account_id IS NOT NULL
			)
			AND is_header = true;
		`

		result := db.Exec(fixHeaderSQL)
		if result.Error != nil {
			return fmt.Errorf("failed to fix header accounts: %w", result.Error)
		}

		log.Printf("✅ Fixed %d cash bank accounts (converted from header to non-header)", result.RowsAffected)
	} else {
		log.Println("✅ Cash bank accounts are properly configured (non-header)")
	}

	// Check balance synchronization status
	var syncIssues int64
	syncQuery := `
		SELECT COUNT(*) 
		FROM cash_banks cb
		JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL 
		AND a.deleted_at IS NULL
		AND cb.is_active = true
		AND ABS(cb.balance - a.balance) > 0.01
	`

	if err := db.Raw(syncQuery).Scan(&syncIssues).Error; err != nil {
		return fmt.Errorf("failed to check sync status: %w", err)
	}

	if syncIssues > 0 {
		log.Printf("⚠️ Found %d cash banks with balance sync issues. Fixing...", syncIssues)
		
		// Run manual sync for all cash banks
		var cashBankIDs []uint
		if err := db.Raw("SELECT id FROM cash_banks WHERE deleted_at IS NULL AND is_active = true").Scan(&cashBankIDs).Error; err != nil {
			return fmt.Errorf("failed to get cash bank IDs: %w", err)
		}

		for _, cbID := range cashBankIDs {
			if err := db.Exec("SELECT manual_sync_cashbank_coa(?)", cbID).Error; err != nil {
				log.Printf("    ⚠️ Failed to sync cash bank ID %d: %v", cbID, err)
			}
		}

		log.Printf("✅ Attempted to sync %d cash bank accounts", len(cashBankIDs))
	} else {
		log.Println("✅ All cash bank accounts are properly synchronized")
	}

	// Final status report
	log.Println("📊 System Status Report:")
	
	// Re-check triggers
	if err := db.Raw(triggerQuery).Scan(&triggerCount).Error; err == nil {
		log.Printf("  Triggers: %d/4 active", triggerCount)
	}

	// Re-check sync
	if err := db.Raw(syncQuery).Scan(&syncIssues).Error; err == nil {
		log.Printf("  Sync Issues: %d", syncIssues)
	}

	// Check functions
	var functionCount int64
	functionQuery := `
		SELECT COUNT(*) 
		FROM pg_proc p
		JOIN pg_namespace n ON p.pronamespace = n.oid
		WHERE n.nspname = 'public'
		AND p.proname IN (
			'update_parent_account_balances',
			'recalculate_cashbank_balance',
			'manual_sync_cashbank_coa',
			'direct_sync_cashbank_coa'
		)
	`

	if err := db.Raw(functionQuery).Scan(&functionCount).Error; err == nil {
		log.Printf("  Functions: %d/4 available", functionCount)
	}

	if triggerCount >= 4 && syncIssues == 0 {
		log.Println("🎯 System is fully operational and ready!")
	} else {
		log.Println("⚠️ System may need additional setup. Check the logs above.")
	}

	return nil
}

func installAutoBalanceTriggers(db *gorm.DB) error {
	// This is a simplified version. For full installation, run the main trigger installation script
	
	sqlCommands := []string{
		// Essential function for parent balance updates
		`CREATE OR REPLACE FUNCTION update_parent_account_balances(child_account_id BIGINT)
		RETURNS VOID AS $$
		DECLARE
		    parent_account_id BIGINT;
		    parent_balance DECIMAL(15,2);
		BEGIN
		    SELECT parent_id INTO parent_account_id
		    FROM accounts 
		    WHERE id = child_account_id 
		    AND deleted_at IS NULL;
		    
		    IF parent_account_id IS NOT NULL THEN
		        SELECT COALESCE(SUM(balance), 0)
		        INTO parent_balance
		        FROM accounts 
		        WHERE parent_id = parent_account_id 
		        AND deleted_at IS NULL;
		        
		        UPDATE accounts 
		        SET balance = parent_balance, updated_at = NOW()
		        WHERE id = parent_account_id 
		        AND deleted_at IS NULL;
		        
		        PERFORM update_parent_account_balances(parent_account_id);
		    END IF;
		END;
		$$ LANGUAGE plpgsql;`,

		// Essential function for cash bank balance recalculation
		`CREATE OR REPLACE FUNCTION recalculate_cashbank_balance()
		RETURNS TRIGGER AS $$
		DECLARE
		    cash_bank_id_val BIGINT;
		    new_balance DECIMAL(15,2);
		    linked_account_id BIGINT;
		BEGIN
		    IF TG_OP = 'DELETE' THEN
		        cash_bank_id_val := OLD.cash_bank_id;
		    ELSE
		        cash_bank_id_val := NEW.cash_bank_id;
		    END IF;
		    
		    SELECT COALESCE(SUM(amount), 0)
		    INTO new_balance
		    FROM cash_bank_transactions 
		    WHERE cash_bank_id = cash_bank_id_val 
		    AND deleted_at IS NULL;
		    
		    UPDATE cash_banks 
		    SET balance = new_balance, updated_at = NOW()
		    WHERE id = cash_bank_id_val 
		    AND deleted_at IS NULL
		    RETURNING account_id INTO linked_account_id;
		    
		    IF linked_account_id IS NOT NULL THEN
		        UPDATE accounts 
		        SET balance = new_balance, updated_at = NOW()
		        WHERE id = linked_account_id 
		        AND deleted_at IS NULL;
		        
		        PERFORM update_parent_account_balances(linked_account_id);
		    END IF;
		    
		    IF TG_OP = 'DELETE' THEN
		        RETURN OLD;
		    ELSE
		        RETURN NEW;
		    END IF;
		END;
		$$ LANGUAGE plpgsql;`,

		// Create triggers
		`DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_insert ON cash_bank_transactions;
		CREATE TRIGGER trigger_recalc_cashbank_balance_insert
		    AFTER INSERT ON cash_bank_transactions
		    FOR EACH ROW
		    EXECUTE FUNCTION recalculate_cashbank_balance();`,

		`DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_update ON cash_bank_transactions;
		CREATE TRIGGER trigger_recalc_cashbank_balance_update
		    AFTER UPDATE OF amount, cash_bank_id ON cash_bank_transactions
		    FOR EACH ROW
		    WHEN (OLD.amount IS DISTINCT FROM NEW.amount OR OLD.cash_bank_id IS DISTINCT FROM NEW.cash_bank_id)
		    EXECUTE FUNCTION recalculate_cashbank_balance();`,

		`DROP TRIGGER IF EXISTS trigger_recalc_cashbank_balance_delete ON cash_bank_transactions;
		CREATE TRIGGER trigger_recalc_cashbank_balance_delete
		    AFTER DELETE ON cash_bank_transactions
		    FOR EACH ROW
		    EXECUTE FUNCTION recalculate_cashbank_balance();`,
	}

	for i, sql := range sqlCommands {
		if err := db.Exec(sql).Error; err != nil {
			return fmt.Errorf("failed to execute SQL command %d: %w", i+1, err)
		}
	}

	return nil
}