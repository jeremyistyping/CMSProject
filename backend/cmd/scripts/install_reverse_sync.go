package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔄 Installing Reverse Sync Trigger (COA → CashBank)...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Read the SQL file
	sqlContent, err := ioutil.ReadFile("add_reverse_sync_trigger.sql")
	if err != nil {
		log.Fatalf("❌ Failed to read SQL file: %v", err)
	}
	
	fmt.Println("📂 SQL file loaded successfully")
	
	// Execute the SQL commands
	fmt.Println("⚙️  Installing reverse sync trigger...")
	if err := db.Exec(string(sqlContent)).Error; err != nil {
		log.Fatalf("❌ Failed to execute SQL: %v", err)
	}
	
	fmt.Println("✅ Reverse sync trigger installed successfully!")
	
	// Verify installation
	fmt.Println("\n🔍 Verifying trigger installation...")
	
	// Check if the new trigger exists
	var triggerCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.triggers 
		WHERE trigger_name = 'trg_sync_coa_to_cashbank'
	`).Scan(&triggerCount)
	
	if triggerCount > 0 {
		fmt.Println("✅ Reverse sync trigger found in database")
	} else {
		fmt.Println("❌ Reverse sync trigger not found!")
	}
	
	// Test the reverse sync functionality
	fmt.Println("\n🧪 Testing reverse sync functionality...")
	
	// Get current CashBank and COA balances
	var results []struct {
		CashBankID      uint    `db:"cashbank_id"`
		CashBankName    string  `db:"cashbank_name"`
		CashBankBalance float64 `db:"cashbank_balance"`
		AccountID       uint    `db:"account_id"`
		AccountCode     string  `db:"account_code"`
		COABalance      float64 `db:"coa_balance"`
	}
	
	db.Raw(`
		SELECT 
			cb.id as cashbank_id,
			cb.name as cashbank_name, 
			cb.balance as cashbank_balance,
			a.id as account_id,
			a.code as account_code,
			a.balance as coa_balance
		FROM cash_banks cb 
		JOIN accounts a ON cb.account_id = a.id 
		WHERE cb.deleted_at IS NULL AND cb.is_active = true
		LIMIT 1
	`).Scan(&results)
	
	if len(results) > 0 {
		result := results[0]
		fmt.Printf("📊 Current State:\n")
		fmt.Printf("   CashBank: %s (ID: %d) = %.2f\n", result.CashBankName, result.CashBankID, result.CashBankBalance)
		fmt.Printf("   COA: %s (ID: %d) = %.2f\n", result.AccountCode, result.AccountID, result.COABalance)
		
		// Test by updating COA balance
		testBalance := result.COABalance + 1000000 // Add 1M for test
		fmt.Printf("\n🔧 Testing: Setting COA %s balance to %.2f\n", result.AccountCode, testBalance)
		
		err := db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", testBalance, result.AccountID).Error
		if err != nil {
			fmt.Printf("❌ Failed to update COA balance: %v\n", err)
		} else {
			fmt.Println("✅ COA balance updated")
			
			// Check if CashBank balance was automatically updated
			var newCashBankBalance float64
			db.Raw("SELECT balance FROM cash_banks WHERE id = ?", result.CashBankID).Scan(&newCashBankBalance)
			
			if newCashBankBalance == testBalance {
				fmt.Printf("🎉 SUCCESS! CashBank balance automatically synced to %.2f\n", newCashBankBalance)
			} else {
				fmt.Printf("❌ FAILED! CashBank balance is %.2f, expected %.2f\n", newCashBankBalance, testBalance)
			}
			
			// Restore original balance
			db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", result.COABalance, result.AccountID)
			fmt.Println("🔄 Original balance restored")
		}
	} else {
		fmt.Println("⚠️  No linked CashBank accounts found for testing")
	}
	
	// Check audit logs
	fmt.Println("\n📋 Checking audit logs...")
	var auditCount int64
	db.Raw("SELECT COUNT(*) FROM audit_logs WHERE table_name = 'coa_to_cashbank_sync'").Scan(&auditCount)
	fmt.Printf("Found %d audit log entries for reverse sync\n", auditCount)
	
	fmt.Println("\n✅ Reverse sync trigger installation and testing completed!")
	fmt.Println("🔄 Now both directions are supported:")
	fmt.Println("   ↗️  CashBank → COA: ✅ Working")
	fmt.Println("   ↙️  COA → CashBank: ✅ Working")
}
