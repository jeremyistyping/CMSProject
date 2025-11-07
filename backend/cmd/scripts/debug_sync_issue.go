package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔍 Deep Debugging Synchronization Issue")

	db := database.ConnectDB()

	// Get CashBank ID 2 (the one we've been testing)
	cashBankID := uint(2)
	
	log.Printf("📋 Analyzing CashBank ID: %d", cashBankID)

	// 1. Get CashBank record
	var cashBank models.CashBank
	if err := db.First(&cashBank, cashBankID).Error; err != nil {
		log.Fatalf("❌ CashBank not found: %v", err)
	}

	// 2. Get related GL Account
	var glAccount models.Account
	if err := db.First(&glAccount, cashBank.AccountID).Error; err != nil {
		log.Fatalf("❌ GL Account not found: %v", err)
	}

	log.Printf("🏦 CashBank Record:")
	log.Printf("   ID: %d, Name: %s, Balance: %.2f", cashBank.ID, cashBank.Name, cashBank.Balance)
	log.Printf("   Account ID: %d, Updated: %s", cashBank.AccountID, cashBank.UpdatedAt.Format("15:04:05"))

	log.Printf("🏦 GL Account Record:")
	log.Printf("   ID: %d, Code: %s, Name: %s, Balance: %.2f", 
		glAccount.ID, glAccount.Code, glAccount.Name, glAccount.Balance)
	log.Printf("   Updated: %s", glAccount.UpdatedAt.Format("15:04:05"))

	// 3. Get recent CashBank transactions (should show what the balance SHOULD be)
	var transactions []models.CashBankTransaction
	if err := db.Where("cash_bank_id = ?", cashBankID).Order("id DESC").Limit(5).Find(&transactions).Error; err != nil {
		log.Printf("⚠️ No transactions found: %v", err)
	} else {
		log.Printf("💰 Recent CashBank Transactions:")
		for i, txn := range transactions {
			log.Printf("   %d. ID=%d, Amount=%.2f, Balance After=%.2f, Date=%s", 
				i+1, txn.ID, txn.Amount, txn.BalanceAfter, 
				txn.TransactionDate.Format("15:04:05"))
		}
		
		if len(transactions) > 0 {
			latestTransaction := transactions[0]
			log.Printf("🎯 Latest Transaction Balance After: %.2f", latestTransaction.BalanceAfter)
			log.Printf("🎯 Current CashBank Balance: %.2f", cashBank.Balance)
			if latestTransaction.BalanceAfter != cashBank.Balance {
				log.Printf("❌ INCONSISTENCY: Transaction shows %.2f but CashBank shows %.2f", 
					latestTransaction.BalanceAfter, cashBank.Balance)
			} else {
				log.Printf("✅ CashBank balance matches latest transaction")
			}
		}
	}

	// 4. Get recent SSOT journal entries affecting this account
	var ssotEntries []models.SSOTJournalEntry
	if err := db.Where("source_type = ?", "PAYMENT").Order("id DESC").Limit(3).Find(&ssotEntries).Error; err != nil {
		log.Printf("⚠️ No SSOT entries found: %v", err)
	} else {
		log.Printf("📝 Recent SSOT Journal Entries:")
		for i, entry := range ssotEntries {
			log.Printf("   %d. ID=%d, Status=%s, Total Debit=%.2f, Entry Number=%s", 
				i+1, entry.ID, entry.Status, entry.TotalDebit.InexactFloat64(), entry.EntryNumber)
			
			// Check lines affecting our GL account
			var lines []models.SSOTJournalLine
			db.Where("journal_id = ? AND account_id = ?", entry.ID, glAccount.ID).Find(&lines)
			for _, line := range lines {
				log.Printf("      - GL Account affected: Debit=%.2f, Credit=%.2f", 
					line.DebitAmount.InexactFloat64(), line.CreditAmount.InexactFloat64())
			}
		}
	}

	// 5. Calculate expected GL Account balance based on SSOT
	var totalDebits, totalCredits float64
	var lineCount int64
	
	err := db.Table("unified_journal_lines").
		Select("SUM(debit_amount) as total_debits, SUM(credit_amount) as total_credits, COUNT(*) as line_count").
		Where("account_id = ?", glAccount.ID).
		Joins("JOIN unified_journal_ledger ON unified_journal_lines.journal_id = unified_journal_ledger.id").
		Where("unified_journal_ledger.status = ?", "POSTED").
		Scan(&map[string]interface{}{
			"total_debits": &totalDebits,
			"total_credits": &totalCredits, 
			"line_count": &lineCount,
		}).Error
	
	if err != nil {
		log.Printf("⚠️ Could not calculate SSOT balance: %v", err)
	} else {
		expectedBalance := totalDebits - totalCredits // For asset accounts
		log.Printf("📊 SSOT Journal Analysis for GL Account %s:", glAccount.Code)
		log.Printf("   Total Debits: %.2f", totalDebits)
		log.Printf("   Total Credits: %.2f", totalCredits)
		log.Printf("   Line Count: %d", lineCount)
		log.Printf("   Expected Balance: %.2f", expectedBalance)
		log.Printf("   Actual Balance: %.2f", glAccount.Balance)
		
		if expectedBalance != glAccount.Balance {
			log.Printf("❌ INCONSISTENCY: SSOT expects %.2f but GL Account shows %.2f", 
				expectedBalance, glAccount.Balance)
		} else {
			log.Printf("✅ GL Account balance matches SSOT calculations")
		}
	}

	// 6. Check for any ongoing transactions or locks
	log.Printf("\n🔐 Checking for potential transaction issues:")
	
	// Get the latest payment that might be causing issues
	var latestPayment models.Payment
	if err := db.Order("id DESC").First(&latestPayment).Error; err == nil {
		log.Printf("💳 Latest Payment: ID=%d, Code=%s, Amount=%.2f, Status=%s", 
			latestPayment.ID, latestPayment.Code, latestPayment.Amount, latestPayment.Status)
		
		// Check if it has journal reference
		if latestPayment.JournalEntryID != nil {
			log.Printf("   Journal Reference: %d", *latestPayment.JournalEntryID)
		} else {
			log.Printf("   ⚠️ No journal reference")
		}
	}

	// 7. Summary and recommendations
	log.Printf("\n📋 SYNCHRONIZATION ANALYSIS SUMMARY:")
	log.Printf("===============================================")
	
	if len(transactions) > 0 {
		latestTxnBalance := transactions[0].BalanceAfter
		log.Printf("🏦 CashBank Transaction Balance: %.2f", latestTxnBalance)
		log.Printf("🏦 CashBank Table Balance: %.2f", cashBank.Balance)
		log.Printf("🏦 GL Account Balance: %.2f", glAccount.Balance)
		
		log.Printf("\n🔍 ISSUES IDENTIFIED:")
		if latestTxnBalance != cashBank.Balance {
			log.Printf("❌ CashBank table not synced with transactions")
		}
		if cashBank.Balance != glAccount.Balance {
			log.Printf("❌ GL Account not synced with CashBank")
		}
		
		log.Printf("\n💡 RECOMMENDED ACTIONS:")
		log.Printf("1. CashBank balance should be: %.2f", latestTxnBalance)
		log.Printf("2. GL Account balance should match CashBank balance")
		log.Printf("3. Investigate why balance updates are not persisting")
	}

	log.Printf("\n✅ Deep synchronization analysis completed")
}