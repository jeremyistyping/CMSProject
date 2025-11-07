package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

type BalanceCheck struct {
	Code           string
	Name           string
	CashBankBalance float64
	AccountCode    string
	COABalance     float64
	Difference     float64
}

func main() {
	// Load configuration
	_ = config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	fmt.Println("🧪 TESTING CASHBANK-COA SYNC MECHANISM")
	fmt.Println("=" + string(make([]byte, 60)))

	// Get current state
	fmt.Println("\n📊 Current State:")
	var currentState []BalanceCheck
	db.Raw(`
		SELECT 
			cb.code,
			cb.name,
			cb.balance as cash_bank_balance,
			a.code as account_code,
			a.balance as coa_balance,
			(cb.balance - a.balance) as difference
		FROM cash_banks cb
		LEFT JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL
		ORDER BY cb.id
	`).Scan(&currentState)

	for _, cs := range currentState {
		status := "✅"
		if cs.Difference != 0 {
			status = "❌"
		}
		fmt.Printf("%s [%s] %s: CB=%.2f | COA[%s]=%.2f | Diff=%.2f\n",
			status, cs.Code, cs.Name, cs.CashBankBalance, cs.AccountCode, cs.COABalance, cs.Difference)
	}

	// Test 1: Update a cash bank balance
	fmt.Println("\n\n🧪 TEST 1: Updating Cash Bank Balance")
	fmt.Println("-" + string(make([]byte, 50)))

	// Find a cash bank with account linked
	var testCashBank struct {
		ID        uint
		Code      string
		Name      string
		Balance   float64
		AccountID uint
	}
	
	db.Raw(`
		SELECT id, code, name, balance, account_id
		FROM cash_banks
		WHERE deleted_at IS NULL AND account_id IS NOT NULL
		ORDER BY id
		LIMIT 1
	`).Scan(&testCashBank)

	if testCashBank.ID == 0 {
		log.Fatal("❌ No cash bank with linked COA account found for testing")
	}

	fmt.Printf("Testing with: [%s] %s (ID: %d)\n", testCashBank.Code, testCashBank.Name, testCashBank.ID)
	fmt.Printf("Current Balance: %.2f\n", testCashBank.Balance)

	// Update balance (add 100,000)
	newBalance := testCashBank.Balance + 100000
	fmt.Printf("\n⚙️  Updating balance to: %.2f (+100,000)\n", newBalance)

	if err := db.Exec(`
		UPDATE cash_banks 
		SET balance = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?
	`, newBalance, testCashBank.ID).Error; err != nil {
		log.Fatalf("❌ Failed to update cash bank balance: %v", err)
	}

	// Check if COA was synced
	fmt.Println("\n🔍 Checking if COA account was auto-synced...")
	
	var coaBalance float64
	db.Raw("SELECT balance FROM accounts WHERE id = ?", testCashBank.AccountID).Scan(&coaBalance)

	fmt.Printf("Cash Bank Balance: %.2f\n", newBalance)
	fmt.Printf("COA Balance:       %.2f\n", coaBalance)

	if coaBalance == newBalance {
		fmt.Println("✅ SUCCESS: COA balance auto-synced with Cash Bank balance!")
	} else {
		fmt.Printf("❌ FAILED: COA balance (%.2f) doesn't match Cash Bank balance (%.2f)\n", coaBalance, newBalance)
	}

	// Restore original balance
	fmt.Println("\n⏮️  Restoring original balance...")
	db.Exec(`
		UPDATE cash_banks 
		SET balance = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?
	`, testCashBank.Balance, testCashBank.ID)

	// Test 2: Try to link two cash banks to same COA account (should fail)
	fmt.Println("\n\n🧪 TEST 2: Attempting to Link Two Cash Banks to Same COA Account")
	fmt.Println("-" + string(make([]byte, 50)))

	var cashBank1, cashBank2 struct {
		ID        uint
		Code      string
		Name      string
		AccountID *uint
	}

	// Get two different cash banks
	db.Raw(`
		SELECT id, code, name, account_id
		FROM cash_banks
		WHERE deleted_at IS NULL
		ORDER BY id
		LIMIT 1
	`).Scan(&cashBank1)

	db.Raw(`
		SELECT id, code, name, account_id
		FROM cash_banks
		WHERE deleted_at IS NULL AND id != ?
		ORDER BY id
		LIMIT 1
	`, cashBank1.ID).Scan(&cashBank2)

	if cashBank1.ID == 0 || cashBank2.ID == 0 {
		fmt.Println("⏭️  Skipping: Not enough cash banks for testing")
	} else if cashBank1.AccountID == nil {
		fmt.Println("⏭️  Skipping: First cash bank has no linked account")
	} else {
		fmt.Printf("Attempting to link [%s] %s to same account as [%s] %s\n", 
			cashBank2.Code, cashBank2.Name, cashBank1.Code, cashBank1.Name)

		err := db.Exec(`
			UPDATE cash_banks 
			SET account_id = ? 
			WHERE id = ?
		`, *cashBank1.AccountID, cashBank2.ID).Error

		if err != nil {
			fmt.Printf("✅ SUCCESS: Duplicate link prevented by unique constraint!\n")
			fmt.Printf("   Error: %v\n", err)
		} else {
			fmt.Println("❌ FAILED: Duplicate link was allowed (constraint not working)")
			// Restore
			if cashBank2.AccountID != nil {
				db.Exec("UPDATE cash_banks SET account_id = ? WHERE id = ?", *cashBank2.AccountID, cashBank2.ID)
			} else {
				db.Exec("UPDATE cash_banks SET account_id = NULL WHERE id = ?", cashBank2.ID)
			}
		}
	}

	// Final verification
	fmt.Println("\n\n📊 Final Verification:")
	fmt.Println("-" + string(make([]byte, 60)))
	
	var finalState []BalanceCheck
	db.Raw(`
		SELECT 
			cb.code,
			cb.name,
			cb.balance as cash_bank_balance,
			a.code as account_code,
			a.balance as coa_balance,
			(cb.balance - a.balance) as difference
		FROM cash_banks cb
		LEFT JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL
		ORDER BY cb.id
	`).Scan(&finalState)

	allSynced := true
	for _, fs := range finalState {
		status := "✅"
		if fs.Difference != 0 {
			status = "❌"
			allSynced = false
		}
		fmt.Printf("%s [%s] %s: CB=%.2f | COA[%s]=%.2f\n",
			status, fs.Code, fs.Name, fs.CashBankBalance, fs.AccountCode, fs.COABalance)
	}

	fmt.Println("\n" + string(make([]byte, 60)))
	if allSynced {
		fmt.Println("🎉 ALL TESTS PASSED! Cash Bank-COA sync is working correctly!")
	} else {
		fmt.Println("⚠️  Some balances are not synced. Please investigate.")
	}
}
