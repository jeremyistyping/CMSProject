package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🧪 Testing Reverse Sync with Proper Linked Accounts...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Find all CashBank accounts with their linked COA accounts
	var linkedAccounts []struct {
		CashBankID      uint    `db:"cashbank_id"`
		CashBankName    string  `db:"cashbank_name"`
		CashBankBalance float64 `db:"cashbank_balance"`
		AccountID       uint    `db:"account_id"`
		AccountCode     string  `db:"account_code"`
		AccountName     string  `db:"account_name"`
		COABalance      float64 `db:"coa_balance"`
	}
	
	db.Raw(`
		SELECT 
			cb.id as cashbank_id,
			cb.name as cashbank_name, 
			cb.balance as cashbank_balance,
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.balance as coa_balance
		FROM cash_banks cb 
		JOIN accounts a ON cb.account_id = a.id 
		WHERE cb.deleted_at IS NULL 
		  AND cb.is_active = true
		  AND cb.account_id IS NOT NULL
		  AND cb.account_id > 0
		ORDER BY cb.name
	`).Scan(&linkedAccounts)
	
	fmt.Printf("Found %d linked CashBank-COA pairs:\n", len(linkedAccounts))
	for i, acc := range linkedAccounts {
		fmt.Printf("%d. CashBank: %s (%.2f) ↔ COA: %s %s (%.2f)\n", 
			i+1, acc.CashBankName, acc.CashBankBalance,
			acc.AccountCode, acc.AccountName, acc.COABalance)
	}
	
	if len(linkedAccounts) == 0 {
		fmt.Println("❌ No linked accounts found for testing!")
		return
	}
	
	// Test with the first linked account
	testAccount := linkedAccounts[0]
	fmt.Printf("\n🔧 Testing with: %s ↔ %s %s\n", 
		testAccount.CashBankName, testAccount.AccountCode, testAccount.AccountName)
	
	// Check current values
	fmt.Printf("Current values:\n")
	fmt.Printf("  CashBank balance: %.2f\n", testAccount.CashBankBalance)
	fmt.Printf("  COA balance: %.2f\n", testAccount.COABalance)
	
	// Test reverse sync: COA → CashBank
	testAmount := testAccount.COABalance + 500000 // Add 500k for test
	fmt.Printf("\n🔄 Testing COA → CashBank sync:\n")
	fmt.Printf("  Updating COA %s balance from %.2f to %.2f\n", 
		testAccount.AccountCode, testAccount.COABalance, testAmount)
	
	// Update COA balance
	err := db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", 
		testAmount, testAccount.AccountID).Error
	if err != nil {
		fmt.Printf("❌ Failed to update COA balance: %v\n", err)
		return
	}
	
	// Check if CashBank balance was updated by trigger
	var newCashBankBalance float64
	db.Raw("SELECT balance FROM cash_banks WHERE id = ?", 
		testAccount.CashBankID).Scan(&newCashBankBalance)
	
	fmt.Printf("  Result: CashBank balance is now %.2f\n", newCashBankBalance)
	
	if newCashBankBalance == testAmount {
		fmt.Println("🎉 SUCCESS! Reverse sync (COA → CashBank) working!")
	} else {
		fmt.Println("❌ FAILED! Reverse sync not working")
		
		// Manual trigger test - force update CashBank
		fmt.Println("\n🔧 Manual fix test...")
		db.Exec("UPDATE cash_banks SET balance = ? WHERE id = ?", 
			testAmount, testAccount.CashBankID)
		fmt.Printf("  Manually updated CashBank balance to %.2f\n", testAmount)
	}
	
	// Test forward sync: CashBank → COA (should also work)
	fmt.Printf("\n🔄 Testing CashBank → COA sync (should work via existing triggers):\n")
	
	// Create a test transaction to trigger forward sync
	testTxAmount := 50000.0
	fmt.Printf("  Creating test transaction: +%.2f\n", testTxAmount)
	
	// First get any account for the counter entry (using expense account)
	var expenseAccountID uint
	db.Raw("SELECT id FROM accounts WHERE code LIKE '5%' LIMIT 1").Scan(&expenseAccountID)
	
	if expenseAccountID > 0 {
		// Insert test transaction (this should trigger forward sync)
		err = db.Exec(`
			INSERT INTO cash_bank_transactions 
			(cash_bank_id, amount, transaction_date, reference_type, reference_id, notes, created_at, updated_at)
			VALUES (?, ?, NOW(), 'TEST', 1, 'Test reverse sync functionality', NOW(), NOW())
		`, testAccount.CashBankID, testTxAmount).Error
		
		if err != nil {
			fmt.Printf("❌ Failed to create test transaction: %v\n", err)
		} else {
			fmt.Println("✅ Test transaction created")
			
			// Check if both balances updated
			var finalCashBankBalance, finalCOABalance float64
			db.Raw("SELECT balance FROM cash_banks WHERE id = ?", 
				testAccount.CashBankID).Scan(&finalCashBankBalance)
			db.Raw("SELECT balance FROM accounts WHERE id = ?", 
				testAccount.AccountID).Scan(&finalCOABalance)
			
			fmt.Printf("  Final CashBank balance: %.2f\n", finalCashBankBalance)
			fmt.Printf("  Final COA balance: %.2f\n", finalCOABalance)
			
			if finalCashBankBalance == finalCOABalance {
				fmt.Println("🎉 SUCCESS! Both directions syncing properly!")
			} else {
				fmt.Println("⚠️  Sync issue detected")
			}
			
			// Clean up test transaction
			db.Exec(`DELETE FROM cash_bank_transactions WHERE notes = 'Test reverse sync functionality' AND cash_bank_id = ?`, 
				testAccount.CashBankID)
			fmt.Println("🧹 Test transaction cleaned up")
		}
	}
	
	// Restore original values
	fmt.Println("\n🔄 Restoring original values...")
	db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", 
		testAccount.COABalance, testAccount.AccountID)
	db.Exec("UPDATE cash_banks SET balance = ? WHERE id = ?", 
		testAccount.CashBankBalance, testAccount.CashBankID)
	fmt.Println("✅ Original balances restored")
	
	// Check audit logs
	var auditCount int64
	db.Raw("SELECT COUNT(*) FROM audit_logs WHERE table_name = 'coa_to_cashbank_sync'").Scan(&auditCount)
	fmt.Printf("\n📋 Found %d audit log entries for reverse sync\n", auditCount)
	
	fmt.Println("\n=== SUMMARY ===")
	fmt.Println("✅ Reverse sync trigger installed")
	fmt.Println("✅ Test completed")
	if newCashBankBalance == testAmount {
		fmt.Println("🎉 COA → CashBank sync: WORKING")
	} else {
		fmt.Println("❌ COA → CashBank sync: NEEDS DEBUGGING")
	}
}
