package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
)

func main() {
	// Load config and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("🔄 Force Refreshing Account Balances from SSOT")
	fmt.Println("==============================================")

	// Key accounts to check and update
	accountCodes := []string{"1201", "4101", "4000", "2103", "1104"}
	
	for _, code := range accountCodes {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			fmt.Printf("   Account %s: Not found\n", code)
			continue
		}

		// Calculate SSOT balance manually
		var result struct {
			TotalDebit  decimal.Decimal
			TotalCredit decimal.Decimal
		}
		
		err = db.Raw(`
			SELECT 
				COALESCE(SUM(debit_amount), 0) as total_debit,
				COALESCE(SUM(credit_amount), 0) as total_credit
			FROM unified_journal_lines ujl
			JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
			WHERE ujl.account_id = ? AND ujd.status = 'POSTED' AND ujd.deleted_at IS NULL
		`, account.ID).Scan(&result).Error

		if err != nil {
			fmt.Printf("   Account %s (%s): Error calculating SSOT balance: %v\n", code, account.Name, err)
			continue
		}

		// Calculate net balance based on normal balance
		totalDebit, _ := result.TotalDebit.Float64()
		totalCredit, _ := result.TotalCredit.Float64()
		
		var newBalance float64
		normalBalance := account.GetNormalBalance()
		
		if normalBalance == models.NormalBalanceDebit {
			// Debit accounts: Balance = Debit - Credit
			newBalance = totalDebit - totalCredit
		} else {
			// Credit accounts: Balance = Credit - Debit (stored as negative)
			newBalance = totalCredit - totalDebit
			newBalance = -newBalance // Store credit balances as negative
		}

		// Update account balance
		oldBalance := account.Balance
		account.Balance = newBalance
		
		err = db.Save(&account).Error
		if err != nil {
			fmt.Printf("   Account %s (%s): Error updating balance: %v\n", code, account.Name, err)
		} else {
			fmt.Printf("   ✅ Account %s (%s): %.2f → %.2f (SSOT: Dr %.2f, Cr %.2f)\n", 
				code, account.Name, oldBalance, newBalance, totalDebit, totalCredit)
		}
	}

	fmt.Println("\n✅ Balance refresh complete!")
	fmt.Println("Now refresh the frontend Chart of Accounts page to see updated balances.")
}