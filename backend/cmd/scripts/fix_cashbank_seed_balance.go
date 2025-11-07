package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	fmt.Println("=== FIXING CASHBANK SEED BALANCE ISSUE ===")
	
	// 1. Show current state
	var cashBanks []models.CashBank
	db.Where("deleted_at IS NULL").Find(&cashBanks)

	fmt.Printf("\n1. CURRENT CASHBANK BALANCES:\n")
	fmt.Println("ID\tCode\t\tName\t\t\tBalance\t\tAccount ID")
	fmt.Println("--\t----\t\t----\t\t\t-------\t\t----------")
	
	var totalCashBankBalance float64
	for _, cb := range cashBanks {
		fmt.Printf("%d\t%s\t\t%-15s\t%.2f\t\t%v\n", 
			cb.ID, cb.Code, cb.Name, cb.Balance, cb.AccountID)
		totalCashBankBalance += cb.Balance
	}
	fmt.Printf("\nTotal CashBank Balance: %.2f\n", totalCashBankBalance)

	// 2. Show related COA balances
	fmt.Printf("\n2. RELATED COA BALANCES:\n")
	fmt.Println("CashBank\t\tCOA Code\tCOA Balance\tDiscrepancy")
	fmt.Println("--------\t\t--------\t-----------\t-----------")
	
	var totalCOABalance float64
	for _, cb := range cashBanks {
		if cb.AccountID != 0 {
			var account models.Account
			if err := db.Where("id = ? AND deleted_at IS NULL", cb.AccountID).First(&account).Error; err == nil {
				discrepancy := cb.Balance - account.Balance
				fmt.Printf("%-15s\t%s\t\t%.2f\t\t%.2f\n", 
					cb.Name, account.Code, account.Balance, discrepancy)
				totalCOABalance += account.Balance
			}
		}
	}
	fmt.Printf("\nTotal COA Balance: %.2f\n", totalCOABalance)
	fmt.Printf("Total Discrepancy: %.2f\n", totalCashBankBalance - totalCOABalance)

	// 3. Check if balances came from seed (no transactions)
	fmt.Printf("\n3. CHECKING FOR TRANSACTION HISTORY:\n")
	
	for _, cb := range cashBanks {
		if cb.Balance > 0 {
			var txCount int64
			db.Table("cash_bank_transactions").
				Where("cash_bank_id = ? AND deleted_at IS NULL", cb.ID).
				Count(&txCount)
			
			var journalCount int64 = 0
			if cb.AccountID != 0 {
				db.Table("journal_entries").
					Joins("JOIN journals ON journal_entries.journal_id = journals.id").
					Where("journal_entries.account_id = ? AND journals.deleted_at IS NULL", cb.AccountID).
					Count(&journalCount)
			}

			status := "✅ VALID (has transactions)"
			if txCount == 0 && journalCount == 0 {
				status = "🚨 FROM SEED (no transactions)"
			}

			fmt.Printf("%s (%.2f): %d cash_bank_transactions, %d journal_entries - %s\n", 
				cb.Name, cb.Balance, txCount, journalCount, status)
		}
	}

	// 4. Fix the issue by resetting seed balances
	fmt.Printf("\n4. FIXING SEED BALANCES:\n")
	
	fixCount := 0
	for _, cb := range cashBanks {
		if cb.Balance > 0 {
			// Check if this balance came from seed (no transactions)
			var txCount int64
			db.Table("cash_bank_transactions").
				Where("cash_bank_id = ? AND deleted_at IS NULL", cb.ID).
				Count(&txCount)
			
			var journalCount int64 = 0
			if cb.AccountID != 0 {
				db.Table("journal_entries").
					Joins("JOIN journals ON journal_entries.journal_id = journals.id").
					Where("journal_entries.account_id = ? AND journals.deleted_at IS NULL", cb.AccountID).
					Count(&journalCount)
			}

			// If balance exists but no transactions, it came from seed - reset it
			if txCount == 0 && journalCount == 0 && cb.Balance > 0 {
				fmt.Printf("Resetting %s balance from %.2f to 0.00 (seed balance)\n", 
					cb.Name, cb.Balance)
				
				if err := db.Model(&cb).Update("balance", 0).Error; err != nil {
					log.Printf("Error resetting balance for %s: %v", cb.Name, err)
				} else {
					fixCount++
					fmt.Printf("✅ Reset %s balance to 0.00\n", cb.Name)
				}
			}
		}
	}

	if fixCount == 0 {
		fmt.Println("No seed balances found to fix.")
	} else {
		fmt.Printf("\n✅ Successfully reset %d CashBank seed balances\n", fixCount)
	}

	// 5. Verify after fix
	fmt.Printf("\n5. VERIFICATION AFTER FIX:\n")
	
	db.Where("deleted_at IS NULL").Find(&cashBanks)
	
	totalAfterFix := 0.0
	hasBalance := false
	for _, cb := range cashBanks {
		if cb.Balance != 0 {
			fmt.Printf("%s: %.2f\n", cb.Name, cb.Balance)
			totalAfterFix += cb.Balance
			hasBalance = true
		}
	}
	
	if !hasBalance {
		fmt.Println("✅ All CashBank balances are now 0.00 (no seed balances)")
	} else {
		fmt.Printf("Remaining total balance: %.2f (from legitimate transactions)\n", totalAfterFix)
	}

	fmt.Println("\n=== CASHBANK SEED BALANCE FIX COMPLETED ===")
}
