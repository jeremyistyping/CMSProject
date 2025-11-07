package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔧 Fixing Shared COA Account Issue...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Find CashBanks sharing the same COA account
	var sharedAccounts []struct {
		AccountID       uint    `db:"account_id"`
		CashBankCount   int     `db:"cash_bank_count"`
		TotalBalance    float64 `db:"total_balance"`
		TransactionSum  float64 `db:"transaction_sum"`
	}
	
	db.Raw(`
		SELECT 
			cb.account_id,
			COUNT(cb.id) as cash_bank_count,
			SUM(cb.balance) as total_balance,
			COALESCE(SUM(tx_sum.transaction_sum), 0) as transaction_sum
		FROM cash_banks cb
		LEFT JOIN (
			SELECT 
				cash_bank_id,
				SUM(amount) as transaction_sum
			FROM cash_bank_transactions 
			WHERE deleted_at IS NULL 
			GROUP BY cash_bank_id
		) tx_sum ON cb.id = tx_sum.cash_bank_id
		WHERE cb.deleted_at IS NULL 
		  AND cb.is_active = true 
		  AND cb.account_id IS NOT NULL 
		  AND cb.account_id > 0
		GROUP BY cb.account_id
		HAVING COUNT(cb.id) > 1
	`).Scan(&sharedAccounts)
	
	if len(sharedAccounts) == 0 {
		fmt.Println("✅ No shared COA accounts found")
		return
	}
	
	fmt.Printf("Found %d COA account(s) shared by multiple CashBanks:\n", len(sharedAccounts))
	
	for _, shared := range sharedAccounts {
		fmt.Printf("\n🔧 Fixing COA Account ID %d:\n", shared.AccountID)
		fmt.Printf("   - Used by %d CashBanks\n", shared.CashBankCount)
		fmt.Printf("   - Total CashBank Balance: %.2f\n", shared.TotalBalance)
		fmt.Printf("   - Total Transaction Sum: %.2f\n", shared.TransactionSum)
		
		// Use transaction sum as correct balance
		correctBalance := shared.TransactionSum
		fmt.Printf("   - Setting COA balance to: %.2f\n", correctBalance)
		
		// Update the shared COA account balance
		err := db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", correctBalance, shared.AccountID).Error
		if err != nil {
			fmt.Printf("   ❌ Failed to update COA account: %v\n", err)
		} else {
			fmt.Printf("   ✅ COA account balance updated\n")
		}
		
		// Update each individual CashBank balance to match their transaction sum
		var cashBanks []struct {
			ID             uint    `db:"id"`
			Name           string  `db:"name"`
			Balance        float64 `db:"balance"`
			TransactionSum float64 `db:"transaction_sum"`
		}
		
		db.Raw(`
			SELECT 
				cb.id,
				cb.name,
				cb.balance,
				COALESCE(tx_sum.transaction_sum, 0) as transaction_sum
			FROM cash_banks cb
			LEFT JOIN (
				SELECT 
					cash_bank_id,
					SUM(amount) as transaction_sum
				FROM cash_bank_transactions 
				WHERE deleted_at IS NULL 
				GROUP BY cash_bank_id
			) tx_sum ON cb.id = tx_sum.cash_bank_id
			WHERE cb.account_id = ? AND cb.deleted_at IS NULL AND cb.is_active = true
		`, shared.AccountID).Scan(&cashBanks)
		
		for _, cb := range cashBanks {
			if cb.Balance != cb.TransactionSum {
				fmt.Printf("   🔄 Updating %s: %.2f -> %.2f\n", cb.Name, cb.Balance, cb.TransactionSum)
				err := db.Exec("UPDATE cash_banks SET balance = ? WHERE id = ?", cb.TransactionSum, cb.ID).Error
				if err != nil {
					fmt.Printf("      ❌ Failed to update: %v\n", err)
				} else {
					fmt.Printf("      ✅ Updated successfully\n")
				}
			} else {
				fmt.Printf("   ✅ %s already synced\n", cb.Name)
			}
		}
	}
	
	fmt.Println("\n✅ Shared account fix completed!")
}
