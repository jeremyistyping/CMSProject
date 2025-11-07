package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔍 Debugging Triggers...")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Debug triggers
	if err := debugTriggers(db); err != nil {
		log.Fatalf("❌ Debug failed: %v", err)
	}

	log.Println("🎉 Debug completed!")
}

func debugTriggers(db *gorm.DB) error {
	log.Println("📋 Checking trigger status...")

	// Check all triggers
	var triggers []struct {
		TableName   string `json:"table_name"`
		TriggerName string `json:"trigger_name"`
		Event       string `json:"event"`
		Timing      string `json:"timing"`
		Enabled     string `json:"enabled"`
	}

	triggerQuery := `
		SELECT 
			t.table_name,
			tr.trigger_name,
			tr.event_manipulation as event,
			tr.action_timing as timing,
			CASE WHEN pt.tgenabled = 'O' THEN 'enabled' ELSE 'disabled' END as enabled
		FROM information_schema.triggers tr
		JOIN pg_trigger pt ON pt.tgname = tr.trigger_name
		JOIN information_schema.tables t ON t.table_name = tr.event_object_table
		WHERE tr.trigger_schema = 'public'
		AND tr.trigger_name LIKE '%balance%'
		ORDER BY t.table_name, tr.trigger_name;
	`

	if err := db.Raw(triggerQuery).Scan(&triggers).Error; err != nil {
		return fmt.Errorf("failed to get triggers: %w", err)
	}

	log.Printf("Found %d balance-related triggers:", len(triggers))
	for _, trigger := range triggers {
		log.Printf("  Table: %s, Trigger: %s, Event: %s, Timing: %s, Status: %s", 
			trigger.TableName, trigger.TriggerName, trigger.Event, trigger.Timing, trigger.Enabled)
	}

	// Test transaction trigger manually
	log.Println("🧪 Testing trigger functionality manually...")

	// Find a cash bank for testing
	var cashBankID uint
	var accountID uint
	testQuery := `
		SELECT cb.id, cb.account_id
		FROM cash_banks cb
		WHERE cb.deleted_at IS NULL AND cb.is_active = true
		LIMIT 1
	`
	
	type CashBankInfo struct {
		ID        uint `json:"id"`
		AccountID uint `json:"account_id"`
	}
	
	var cbInfo CashBankInfo
	if err := db.Raw(testQuery).Scan(&cbInfo).Error; err != nil {
		return fmt.Errorf("failed to get cash bank for testing: %w", err)
	}

	cashBankID = cbInfo.ID
	accountID = cbInfo.AccountID

	log.Printf("Testing with cash bank ID: %d, account ID: %d", cashBankID, accountID)

	// Get initial balances
	var initialCashBalance, initialCOABalance float64
	
	if err := db.Raw("SELECT balance FROM cash_banks WHERE id = ?", cashBankID).Scan(&initialCashBalance).Error; err != nil {
		return fmt.Errorf("failed to get initial cash balance: %w", err)
	}
	
	if err := db.Raw("SELECT balance FROM accounts WHERE id = ?", accountID).Scan(&initialCOABalance).Error; err != nil {
		return fmt.Errorf("failed to get initial COA balance: %w", err)
	}

	log.Printf("Initial balances - Cash: %.2f, COA: %.2f", initialCashBalance, initialCOABalance)

	// Manually call the trigger function to see if it works
	log.Println("🔧 Testing trigger function directly...")

	if err := db.Exec("SELECT manual_sync_cashbank_coa(?)", cashBankID).Error; err != nil {
		log.Printf("⚠️ Manual sync failed: %v", err)
	} else {
		log.Println("✅ Manual sync function executed successfully")

		// Check balances after manual sync
		var afterCOABalance float64
		if err := db.Raw("SELECT balance FROM accounts WHERE id = ?", accountID).Scan(&afterCOABalance).Error; err != nil {
			return fmt.Errorf("failed to get COA balance after sync: %w", err)
		}
		
		log.Printf("After manual sync - Cash: %.2f, COA: %.2f", initialCashBalance, afterCOABalance)
	}

	// Test if trigger is actually being called
	log.Println("📊 Testing if triggers are being executed...")
	
	// Enable trigger logging (if available)
	log.Println("💡 To debug further, we'll check trigger execution with a test transaction...")

	// Create a minimal test transaction to see what happens
	testTxQuery := `
		INSERT INTO cash_bank_transactions (cash_bank_id, amount, reference_type, transaction_date, notes)
		VALUES (?, 500.00, 'TEST', NOW(), 'Debug trigger test')
		RETURNING id
	`
	
	var testTxID uint
	if err := db.Raw(testTxQuery, cashBankID).Scan(&testTxID).Error; err != nil {
		log.Printf("⚠️ Failed to create test transaction: %v", err)
	} else {
		log.Printf("✅ Created test transaction ID: %d", testTxID)

		// Check if balances were updated by trigger
		var afterCashBalance, afterCOABalance float64
		
		if err := db.Raw("SELECT balance FROM cash_banks WHERE id = ?", cashBankID).Scan(&afterCashBalance).Error; err != nil {
			return fmt.Errorf("failed to get cash balance after test: %w", err)
		}
		
		if err := db.Raw("SELECT balance FROM accounts WHERE id = ?", accountID).Scan(&afterCOABalance).Error; err != nil {
			return fmt.Errorf("failed to get COA balance after test: %w", err)
		}

		log.Printf("After test transaction - Cash: %.2f, COA: %.2f", afterCashBalance, afterCOABalance)

		if afterCashBalance == initialCashBalance + 500.00 {
			log.Println("✅ Cash bank balance trigger worked correctly")
		} else {
			log.Printf("❌ Cash bank balance trigger failed: expected %.2f, got %.2f", initialCashBalance + 500.00, afterCashBalance)
		}

		if afterCOABalance == afterCashBalance {
			log.Println("✅ COA balance sync worked correctly")
		} else {
			log.Printf("❌ COA balance sync failed: cash %.2f != COA %.2f", afterCashBalance, afterCOABalance)
		}

		// Clean up test transaction
		if err := db.Exec("DELETE FROM cash_bank_transactions WHERE id = ?", testTxID).Error; err != nil {
			log.Printf("⚠️ Failed to clean up test transaction: %v", err)
		} else {
			log.Println("🧹 Test transaction cleaned up")
		}
	}

	return nil
}