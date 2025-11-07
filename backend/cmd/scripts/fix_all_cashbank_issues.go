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

	fmt.Println("=== FIXING ALL CASHBANK BALANCE ISSUES ===")
	
	// 1. Get all cash banks and their transactions
	var cashBanks []models.CashBank
	db.Where("deleted_at IS NULL").Find(&cashBanks)

	fmt.Printf("\n1. CURRENT STATE ANALYSIS:\n")
	fmt.Println("CashBank\t\tCurrent Balance\tTransaction Sum\tShould Be\tAction Needed")
	fmt.Println("--------\t\t---------------\t---------------\t---------\t-------------")
	
	type FixAction struct {
		CashBank         models.CashBank
		CurrentBalance   float64
		TransactionSum   float64
		CorrectBalance   float64
		NeedsFixing      bool
		HasSeedBalance   bool
	}
	
	var fixActions []FixAction

	for _, cb := range cashBanks {
		// Calculate sum of transactions
		var transactionSum float64
		db.Table("cash_bank_transactions").
			Where("cash_bank_id = ? AND deleted_at IS NULL", cb.ID).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&transactionSum)

		correctBalance := transactionSum
		needsFixing := cb.Balance != correctBalance
		hasSeedBalance := cb.Balance > transactionSum // More than transactions = seed balance

		action := "✅ OK"
		if needsFixing {
			if hasSeedBalance {
				action = "🚨 Remove seed balance"
			} else {
				action = "🔧 Fix balance"
			}
		}

		fmt.Printf("%-15s\t%.2f\t\t%.2f\t\t%.2f\t\t%s\n", 
			cb.Name, cb.Balance, transactionSum, correctBalance, action)

		fixActions = append(fixActions, FixAction{
			CashBank:       cb,
			CurrentBalance: cb.Balance,
			TransactionSum: transactionSum,
			CorrectBalance: correctBalance,
			NeedsFixing:    needsFixing,
			HasSeedBalance: hasSeedBalance,
		})
	}

	// 2. Apply fixes
	fmt.Printf("\n2. APPLYING FIXES:\n")
	
	fixedCount := 0
	for _, action := range fixActions {
		if action.NeedsFixing {
			fmt.Printf("Fixing %s: %.2f -> %.2f", 
				action.CashBank.Name, action.CurrentBalance, action.CorrectBalance)
			
			if action.HasSeedBalance {
				fmt.Printf(" (removing seed balance of %.2f)", 
					action.CurrentBalance - action.TransactionSum)
			}
			
			if err := db.Model(&action.CashBank).Update("balance", action.CorrectBalance).Error; err != nil {
				log.Printf(" ❌ ERROR: %v", err)
			} else {
				fmt.Printf(" ✅ FIXED\n")
				fixedCount++
			}
		}
	}

	if fixedCount == 0 {
		fmt.Println("No fixes needed.")
	} else {
		fmt.Printf("\n✅ Fixed %d CashBank balances\n", fixedCount)
	}

	// 3. Now check COA sync issues
	fmt.Printf("\n3. CHECKING COA SYNCHRONIZATION:\n")
	fmt.Println("CashBank\t\tCOA Code\tCashBank Balance\tCOA Balance\tJournal Entries\tSync Status")
	fmt.Println("--------\t\t--------\t----------------\t-----------\t---------------\t-----------")
	
	// Refresh cashbanks after fixes
	db.Where("deleted_at IS NULL").Find(&cashBanks)
	
	coaSyncIssues := 0
	for _, cb := range cashBanks {
		if cb.AccountID != 0 {
			var account models.Account
			if err := db.Where("id = ? AND deleted_at IS NULL", cb.AccountID).First(&account).Error; err == nil {
				
				var journalCount int64
				db.Table("journal_entries").
					Joins("JOIN journals ON journal_entries.journal_id = journals.id").
					Where("journal_entries.account_id = ? AND journals.deleted_at IS NULL", cb.AccountID).
					Count(&journalCount)

				syncStatus := "✅ SYNCED"
				if cb.Balance != account.Balance {
					syncStatus = "❌ OUT OF SYNC"
					coaSyncIssues++
				}

				fmt.Printf("%-15s\t%s\t\t%.2f\t\t\t%.2f\t\t%d\t\t%s\n", 
					cb.Name, account.Code, cb.Balance, account.Balance, journalCount, syncStatus)
			}
		}
	}

	if coaSyncIssues > 0 {
		fmt.Printf("\n⚠️  Found %d COA sync issues. This indicates that:\n", coaSyncIssues)
		fmt.Println("   - CashBank transactions are not creating proper journal entries")
		fmt.Println("   - OR journal entries are not updating COA account balances")
		fmt.Println("   - This needs to be fixed in the payment/transaction creation logic")
	}

	// 4. Final verification
	fmt.Printf("\n4. FINAL VERIFICATION:\n")
	
	db.Where("deleted_at IS NULL").Find(&cashBanks)
	
	allGood := true
	totalCashBankBalance := 0.0
	totalTransactionSum := 0.0
	
	for _, cb := range cashBanks {
		var transactionSum float64
		db.Table("cash_bank_transactions").
			Where("cash_bank_id = ? AND deleted_at IS NULL", cb.ID).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&transactionSum)

		if cb.Balance != transactionSum {
			fmt.Printf("❌ %s: Balance %.2f != Transaction Sum %.2f\n", 
				cb.Name, cb.Balance, transactionSum)
			allGood = false
		}

		totalCashBankBalance += cb.Balance
		totalTransactionSum += transactionSum
	}

	if allGood && totalCashBankBalance == totalTransactionSum {
		fmt.Printf("✅ All CashBank balances are now consistent with transactions\n")
		fmt.Printf("Total CashBank Balance: %.2f\n", totalCashBankBalance)
		fmt.Printf("Total Transaction Sum: %.2f\n", totalTransactionSum)
	} else {
		fmt.Printf("❌ Still have inconsistencies\n")
	}

	// 5. Clean up seed data from seed.go
	fmt.Printf("\n5. SEED DATA CLEANUP RECOMMENDATION:\n")
	fmt.Println("To prevent this issue in the future, update database/seed.go:")
	fmt.Println("1. Remove line 252: db.Model(&models.CashBank{}).Where(\"id = ?\", 6).Update(\"balance\", 10000000)")
	fmt.Println("2. Set Balance: 0 in lines 261, 268, 275 for initial CashBank records")
	fmt.Println("3. CashBank balances should only come from legitimate transactions")

	fmt.Println("\n=== CASHBANK FIX COMPLETED ===")
}
