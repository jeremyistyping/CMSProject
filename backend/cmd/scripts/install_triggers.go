package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔧 Installing CashBank-COA Sync Database Triggers...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Read the SQL file
	sqlFile := filepath.Join("database", "migrations", "cashbank_coa_sync_trigger.sql")
	sqlContent, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("❌ Failed to read SQL file: %v", err)
	}
	
	fmt.Printf("📂 SQL file loaded: %s\n", sqlFile)
	
	// Execute the SQL commands
	fmt.Println("⚙️  Executing trigger installation...")
	if err := db.Exec(string(sqlContent)).Error; err != nil {
		log.Fatalf("❌ Failed to execute SQL: %v", err)
	}
	
	fmt.Println("✅ Database triggers installed successfully!")
	
	// Verify installation
	fmt.Println("\n🔍 Verifying trigger installation...")
	
	// Check triggers
	var triggerCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.triggers 
		WHERE trigger_name LIKE '%sync_cashbank%'
	`).Scan(&triggerCount)
	
	if triggerCount > 0 {
		fmt.Printf("✅ Found %d CashBank sync trigger(s)\n", triggerCount)
		
		// Get trigger details
		var triggers []struct {
			TriggerName string `db:"trigger_name"`
			TableName   string `db:"event_object_table"`
			Timing      string `db:"action_timing"`
			Event       string `db:"event_manipulation"`
		}
		
		db.Raw(`
			SELECT trigger_name, event_object_table, action_timing, event_manipulation 
			FROM information_schema.triggers 
			WHERE trigger_name LIKE '%sync_cashbank%'
		`).Scan(&triggers)
		
		for _, trigger := range triggers {
			fmt.Printf("   📌 %s on %s (%s %s)\n", 
				trigger.TriggerName, trigger.TableName, trigger.Timing, trigger.Event)
		}
	} else {
		fmt.Println("❌ No triggers found after installation!")
	}
	
	// Check functions
	var functionCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.routines 
		WHERE routine_name LIKE '%cashbank%' AND routine_type = 'FUNCTION'
	`).Scan(&functionCount)
	
	if functionCount > 0 {
		fmt.Printf("✅ Found %d CashBank sync function(s)\n", functionCount)
	}
	
	// Test integrity validation function
	fmt.Println("\n🔍 Testing integrity validation function...")
	var integrityCheck []struct {
		CashBankID      int     `db:"cash_bank_id"`
		CashBankName    string  `db:"cash_bank_name"`
		COAAccountID    int     `db:"coa_account_id"`
		COAAccountCode  string  `db:"coa_account_code"`
		CashBankBalance float64 `db:"cash_bank_balance"`
		COABalance      float64 `db:"coa_balance"`
		TransactionSum  float64 `db:"transaction_sum"`
		IsSynced        bool    `db:"is_synced"`
		Discrepancy     float64 `db:"discrepancy"`
	}
	
	db.Raw("SELECT * FROM validate_cashbank_coa_integrity()").Scan(&integrityCheck)
	
	syncedCount := 0
	for _, item := range integrityCheck {
		if item.IsSynced {
			syncedCount++
		}
		fmt.Printf("   💰 %s: CB=%.2f, COA=%.2f, TX=%.2f, Synced=%v\n", 
			item.CashBankName, item.CashBankBalance, item.COABalance, 
			item.TransactionSum, item.IsSynced)
	}
	
	fmt.Printf("\n📊 Summary: %d/%d accounts properly synced\n", syncedCount, len(integrityCheck))
	
	if syncedCount == len(integrityCheck) {
		fmt.Println("🎉 All CashBank-COA accounts are properly synchronized!")
	} else {
		fmt.Printf("⚠️  %d accounts need synchronization\n", len(integrityCheck)-syncedCount)
	}
	
	fmt.Println("\n✅ Trigger installation complete!")
}
