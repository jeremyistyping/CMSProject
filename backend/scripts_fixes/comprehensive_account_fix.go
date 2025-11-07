package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("=== Comprehensive Account Structure Fix ===")
	fmt.Println()

	// 1. Fix account 2102 (PPN) hierarchy - should be under liabilities, not assets
	fmt.Println("1. Fixing Account 2102 (PPN) hierarchy...")
	
	// Find the current liabilities parent account
	var currentLiabilitiesID uint
	err := db.Raw(`SELECT id FROM accounts WHERE code = '2100' AND name = 'CURRENT LIABILITIES'`).Scan(&currentLiabilitiesID).Error
	if err != nil {
		log.Printf("Error finding Current Liabilities account: %v", err)
		return
	}

	// Move PPN account to current liabilities and fix its type
	err = db.Exec(`
		UPDATE accounts 
		SET parent_id = ?, type = 'LIABILITY'
		WHERE code = '2102' AND name = 'PPN Masukan'
	`, currentLiabilitiesID).Error
	
	if err != nil {
		log.Printf("Error fixing PPN account: %v", err)
	} else {
		fmt.Println("  ✅ Fixed PPN account hierarchy and type")
	}

	// 2. Move the revenue balance from header account to detail account
	fmt.Println("\n2. Moving revenue balance from header to detail account...")
	
	// Get the current balance of account 23 (4000 - REVENUE header)
	var revenueHeaderBalance float64
	err = db.Raw(`SELECT balance FROM accounts WHERE id = 23`).Scan(&revenueHeaderBalance).Error
	if err != nil {
		log.Printf("Error getting revenue header balance: %v", err)
	} else if revenueHeaderBalance != 0 {
		// Move the balance to account 4101 (Pendapatan Penjualan)
		err = db.Exec(`UPDATE accounts SET balance = balance + ? WHERE code = '4101'`, revenueHeaderBalance).Error
		if err != nil {
			log.Printf("Error moving balance to detail account: %v", err)
		} else {
			// Clear the header account balance
			err = db.Exec(`UPDATE accounts SET balance = 0 WHERE id = 23`).Error
			if err != nil {
				log.Printf("Error clearing header balance: %v", err)
			} else {
				fmt.Printf("  ✅ Moved %.2f from header account 4000 to detail account 4101\n", revenueHeaderBalance)
			}
		}
	}

	// 3. Update the SSOT journal line to use correct account
	fmt.Println("\n3. Updating SSOT journal line to use correct revenue account...")
	
	// Find the ID of the correct revenue account (4101)
	var revenueAccountID uint
	err = db.Raw(`SELECT id FROM accounts WHERE code = '4101'`).Scan(&revenueAccountID).Error
	if err != nil {
		log.Printf("Error finding revenue detail account: %v", err)
	} else {
		// Update the SSOT journal line to use the correct account
		err = db.Exec(`
			UPDATE unified_journal_lines 
			SET account_id = ?
			WHERE account_id = 23 AND description = 'Sales Revenue'
		`, revenueAccountID).Error
		
		if err != nil {
			log.Printf("Error updating SSOT journal line: %v", err)
		} else {
			fmt.Printf("  ✅ Updated SSOT journal line to use account ID %d (4101)\n", revenueAccountID)
		}
	}

	// 4. Recalculate all account balances from SSOT
	fmt.Println("\n4. Recalculating all account balances from SSOT...")
	
	// Reset all balances first
	err = db.Exec("UPDATE accounts SET balance = 0").Error
	if err != nil {
		log.Printf("Error resetting balances: %v", err)
		return
	}

	// Calculate and update balances from SSOT journal lines
	var accountUpdates []struct {
		AccountID       uint    `json:"account_id"`
		CalculatedBalance float64 `json:"calculated_balance"`
	}

	err = db.Raw(`
		SELECT 
			account_id,
			SUM(debit_amount) - SUM(credit_amount) as calculated_balance
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
		WHERE uj.status = 'POSTED'
		GROUP BY account_id
	`).Scan(&accountUpdates).Error

	if err != nil {
		log.Printf("Error calculating balances: %v", err)
		return
	}

	// Update each account balance
	for _, update := range accountUpdates {
		err = db.Model(&models.Account{}).Where("id = ?", update.AccountID).Update("balance", update.CalculatedBalance).Error
		if err != nil {
			log.Printf("  Error updating account %d: %v", update.AccountID, err)
		}
	}
	fmt.Printf("  ✅ Updated %d account balances from SSOT\n", len(accountUpdates))

	// 5. Update parent account balances recursively
	fmt.Println("\n5. Updating parent account balances...")
	
	// Update parent accounts in multiple passes to handle deep hierarchies
	for pass := 0; pass < 5; pass++ {
		err = db.Exec(`
			UPDATE accounts 
			SET balance = (
				SELECT COALESCE(SUM(child.balance), 0)
				FROM accounts child 
				WHERE child.parent_id = accounts.id AND child.deleted_at IS NULL
			)
			WHERE is_header = true AND deleted_at IS NULL
		`).Error

		if err != nil {
			log.Printf("Error updating parent balances (pass %d): %v", pass+1, err)
		}
	}
	fmt.Println("  ✅ Updated parent account balances")

	// 6. Show final results
	fmt.Println("\n6. Final Account Balance Summary:")
	
	var finalResults []struct {
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Balance  float64 `json:"balance"`
		Type     string  `json:"type"`
		IsHeader bool    `json:"is_header"`
	}

	err = db.Raw(`
		SELECT code, name, balance, type, is_header
		FROM accounts 
		WHERE deleted_at IS NULL AND balance != 0
		ORDER BY code
	`).Scan(&finalResults).Error

	if err != nil {
		log.Printf("Error fetching final results: %v", err)
	} else {
		for _, result := range finalResults {
			accountType := "Detail"
			if result.IsHeader {
				accountType = "Header"
			}
			fmt.Printf("  %s: %s [%s %s] - Balance: %.2f\n", 
				result.Code, result.Name, result.Type, accountType, result.Balance)
		}
	}

	// 7. Balance sheet verification
	fmt.Println("\n7. Balance Sheet Verification:")
	var assets, liabilities, equity, revenue, expenses float64

	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'ASSET' AND deleted_at IS NULL`).Scan(&assets)
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'LIABILITY' AND deleted_at IS NULL`).Scan(&liabilities)
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EQUITY' AND deleted_at IS NULL`).Scan(&equity)
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'REVENUE' AND deleted_at IS NULL`).Scan(&revenue)
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EXPENSE' AND deleted_at IS NULL`).Scan(&expenses)

	fmt.Printf("  Assets: %.2f\n", assets)
	fmt.Printf("  Liabilities: %.2f\n", liabilities)
	fmt.Printf("  Equity: %.2f\n", equity)
	fmt.Printf("  Revenue: %.2f\n", revenue)
	fmt.Printf("  Expenses: %.2f\n", expenses)

	// Net Income = Revenue + Expenses (since revenue is negative)
	netIncome := revenue + expenses
	fmt.Printf("  Net Income: %.2f\n", netIncome)

	// Balance equation: Assets = Liabilities + Equity + Net Income
	balanceCheck := assets - (liabilities + equity + netIncome)
	fmt.Printf("  Balance Check: %.2f\n", balanceCheck)

	if balanceCheck == 0 {
		fmt.Println("  ✅ Balance sheet is balanced!")
	} else {
		fmt.Printf("  ⚠️  Balance difference: %.2f\n", balanceCheck)
	}

	fmt.Println("\n=== Comprehensive Fix Complete ===")
	fmt.Println()
	fmt.Println("Summary of fixes applied:")
	fmt.Println("✅ 1. Fixed PPN account hierarchy (moved to liabilities)")
	fmt.Println("✅ 2. Moved revenue balance from header to detail account")
	fmt.Println("✅ 3. Updated SSOT journal to use correct revenue account")
	fmt.Println("✅ 4. Recalculated all balances from SSOT transactions")
	fmt.Println("✅ 5. Updated parent account hierarchical balances")
}