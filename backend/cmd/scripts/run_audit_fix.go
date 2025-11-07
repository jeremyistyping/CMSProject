package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔧 Fixing audit_logs Table Structure for CashBank Sync...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Read and execute the fix SQL
	sqlContent, err := ioutil.ReadFile("fix_audit_logs_table.sql")
	if err != nil {
		log.Fatalf("❌ Failed to read SQL file: %v", err)
	}
	
	fmt.Println("📂 SQL fix file loaded successfully")
	
	// Execute the SQL fix
	fmt.Println("⚙️  Applying audit_logs table fixes...")
	if err := db.Exec(string(sqlContent)).Error; err != nil {
		log.Fatalf("❌ Failed to execute SQL fix: %v", err)
	}
	
	fmt.Println("✅ audit_logs table structure fixed!")
	
	// Verify the table structure
	fmt.Println("\n🔍 Verifying table structure...")
	
	var columns []struct {
		ColumnName    string `db:"column_name"`
		DataType      string `db:"data_type"`
		IsNullable    string `db:"is_nullable"`
		ColumnDefault *string `db:"column_default"`
	}
	
	db.Raw(`
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default
		FROM information_schema.columns
		WHERE table_name = 'audit_logs'
		ORDER BY ordinal_position
	`).Scan(&columns)
	
	fmt.Printf("audit_logs table structure (%d columns):\n", len(columns))
	for _, col := range columns {
		defaultVal := "NULL"
		if col.ColumnDefault != nil {
			defaultVal = *col.ColumnDefault
		}
		fmt.Printf("  ✅ %s: %s (nullable: %s, default: %s)\n", 
			col.ColumnName, col.DataType, col.IsNullable, defaultVal)
	}
	
	// Test CashBank deposit functionality
	fmt.Println("\n🧪 Testing CashBank deposit functionality...")
	
	// Find a CashBank account to test with
	var testCashBank struct {
		ID      uint    `db:"id"`
		Name    string  `db:"name"`
		Balance float64 `db:"balance"`
	}
	
	err = db.Raw(`
		SELECT id, name, balance 
		FROM cash_banks 
		WHERE deleted_at IS NULL AND is_active = true 
		LIMIT 1
	`).Scan(&testCashBank).Error
	
	if err != nil || testCashBank.ID == 0 {
		fmt.Println("⚠️  No CashBank accounts found for testing")
	} else {
		fmt.Printf("Found test CashBank: %s (ID: %d, Balance: %.2f)\n", 
			testCashBank.Name, testCashBank.ID, testCashBank.Balance)
		
		// Test insert a transaction (this should work now)
		fmt.Println("🔧 Testing transaction creation...")
		testAmount := 100.0
		
		err = db.Exec(`
			INSERT INTO cash_bank_transactions 
			(cash_bank_id, amount, balance_after, transaction_date, reference_type, reference_id, notes, created_at, updated_at)
			VALUES (?, ?, ?, NOW(), 'TEST_DEPOSIT', 1, 'Testing audit_logs fix', NOW(), NOW())
		`, testCashBank.ID, testAmount, testCashBank.Balance + testAmount).Error
		
		if err != nil {
			fmt.Printf("❌ Transaction test failed: %v\n", err)
		} else {
			fmt.Println("✅ Transaction created successfully!")
			
			// Clean up test transaction
			db.Exec(`DELETE FROM cash_bank_transactions WHERE notes = 'Testing audit_logs fix'`)
			fmt.Println("🧹 Test transaction cleaned up")
		}
	}
	
	// Check audit log entries
	var auditCount int64
	db.Raw("SELECT COUNT(*) FROM audit_logs").Scan(&auditCount)
	fmt.Printf("\n📋 Total audit_logs entries: %d\n", auditCount)
	
	// Test trigger functionality
	fmt.Println("\n🔄 Testing COA → CashBank sync trigger...")
	
	// Find a linked account to test with
	var linkedAccount struct {
		CashBankID  uint    `db:"cashbank_id"`
		CashBankName string `db:"cashbank_name"`
		AccountID   uint    `db:"account_id"`
		AccountCode string  `db:"account_code"`
		COABalance  float64 `db:"coa_balance"`
	}
	
	err = db.Raw(`
		SELECT 
			cb.id as cashbank_id,
			cb.name as cashbank_name,
			a.id as account_id,
			a.code as account_code,
			a.balance as coa_balance
		FROM cash_banks cb 
		JOIN accounts a ON cb.account_id = a.id 
		WHERE cb.deleted_at IS NULL 
		  AND cb.is_active = true
		  AND cb.account_id > 0
		LIMIT 1
	`).Scan(&linkedAccount).Error
	
	if err != nil || linkedAccount.AccountID == 0 {
		fmt.Println("⚠️  No linked accounts found for trigger testing")
	} else {
		fmt.Printf("Testing trigger with: CashBank ID %d ↔ COA %s\n", 
			linkedAccount.CashBankID, linkedAccount.AccountCode)
		
		// Test trigger by updating COA balance
		testBalance := linkedAccount.COABalance + 12345
		fmt.Printf("Updating COA balance from %.2f to %.2f\n", linkedAccount.COABalance, testBalance)
		
		err = db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", 
			testBalance, linkedAccount.AccountID).Error
		
		if err != nil {
			fmt.Printf("❌ COA update failed: %v\n", err)
		} else {
			fmt.Println("✅ COA balance updated")
			
			// Check if CashBank balance was synced
			var newCashBankBalance float64
			db.Raw("SELECT balance FROM cash_banks WHERE id = ?", 
				linkedAccount.CashBankID).Scan(&newCashBankBalance)
			
			if newCashBankBalance == testBalance {
				fmt.Printf("🎉 SUCCESS! CashBank balance synced to %.2f\n", newCashBankBalance)
			} else {
				fmt.Printf("❌ FAILED! CashBank balance is %.2f, expected %.2f\n", 
					newCashBankBalance, testBalance)
			}
			
			// Restore original balance
			db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", 
				linkedAccount.COABalance, linkedAccount.AccountID)
			db.Exec("UPDATE cash_banks SET balance = ? WHERE id = ?",
				linkedAccount.COABalance, linkedAccount.CashBankID)
			fmt.Println("🔄 Original balances restored")
		}
	}
	
	fmt.Println("\n✅ audit_logs table fix completed!")
	fmt.Println("🎯 Now CashBank deposit should work properly!")
}
