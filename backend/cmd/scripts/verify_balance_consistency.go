package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔍 Final Balance Consistency Verification")

	db := database.ConnectDB()

	// Check CashBank ID 2 (BCA) specifically
	log.Printf("\n💰 CashBank ID 2 (Bank BCA - Operasional1) Verification:")
	
	var cashBank models.CashBank
	if err := db.First(&cashBank, 2).Error; err != nil {
		log.Printf("❌ CashBank not found: %v", err)
		return
	}
	
	log.Printf("   CashBank Balance: %.2f", cashBank.Balance)
	
	// Get related GL Account
	var glAccount models.Account
	if err := db.First(&glAccount, cashBank.AccountID).Error; err != nil {
		log.Printf("❌ GL Account not found: %v", err)
		return
	}
	
	log.Printf("   GL Account (%s) Balance: %.2f", glAccount.Code, glAccount.Balance)
	
	// Get latest CashBank transaction
	var latestTxn models.CashBankTransaction
	if err := db.Where("cash_bank_id = ?", cashBank.ID).Order("id DESC").First(&latestTxn).Error; err == nil {
		log.Printf("   Latest Transaction Balance: %.2f", latestTxn.BalanceAfter)
	}
	
	// Calculate GL account balance from SSOT journal entries
	var result struct {
		TotalDebits  float64
		TotalCredits float64
	}
	
	err := db.Raw(`
		SELECT 
			COALESCE(SUM(ujl.debit_amount), 0) as total_debits,
			COALESCE(SUM(ujl.credit_amount), 0) as total_credits
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
		WHERE ujl.account_id = ? AND uj.status = 'POSTED'
	`, glAccount.ID).Scan(&result).Error
	
	if err == nil {
		// For Asset account: Debit increases, Credit decreases
		calculatedBalance := result.TotalDebits - result.TotalCredits
		log.Printf("   Calculated from SSOT: %.2f (Debits: %.2f - Credits: %.2f)", 
			calculatedBalance, result.TotalDebits, result.TotalCredits)
	}

	// Check AR Account (1201)
	log.Printf("\n📋 AR Account (1201) Verification:")
	
	var arAccount models.Account
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		log.Printf("❌ AR Account not found: %v", err)
		return
	}
	
	log.Printf("   AR Account Balance: %.2f", arAccount.Balance)
	
	// Calculate AR balance from SSOT
	err = db.Raw(`
		SELECT 
			COALESCE(SUM(ujl.debit_amount), 0) as total_debits,
			COALESCE(SUM(ujl.credit_amount), 0) as total_credits
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
		WHERE ujl.account_id = ? AND uj.status = 'POSTED'
	`, arAccount.ID).Scan(&result).Error
	
	if err == nil {
		// For Asset account: Debit increases, Credit decreases
		calculatedBalance := result.TotalDebits - result.TotalCredits
		log.Printf("   Calculated from SSOT: %.2f (Debits: %.2f - Credits: %.2f)", 
			calculatedBalance, result.TotalDebits, result.TotalCredits)
	}

	// Check consistency
	log.Printf("\n✅ Consistency Check:")
	if cashBank.Balance == glAccount.Balance {
		log.Printf("   ✅ CashBank and GL Account balances are synchronized!")
	} else {
		log.Printf("   ❌ CashBank and GL Account balances are NOT synchronized!")
	}

	// Check recent SSOT journal entries
	log.Printf("\n📝 Recent SSOT Journal Entries (Last 3):")
	
	var recentJournals []models.SSOTJournalEntry
	if err := db.Order("id DESC").Limit(3).Find(&recentJournals).Error; err == nil {
		for i, journal := range recentJournals {
			log.Printf("   %d. Entry ID: %d, Number: %s, Status: %s", 
				i+1, journal.ID, journal.EntryNumber, journal.Status)
			
			// Get journal lines
			var lines []models.SSOTJournalLine
			if err := db.Where("journal_id = ?", journal.ID).Find(&lines).Error; err == nil {
				for _, line := range lines {
					var account models.Account
					if err := db.First(&account, line.AccountID).Error; err == nil {
						log.Printf("      - Account: %s (%s), Debit: %.2f, Credit: %.2f", 
							account.Code, account.Name, line.DebitAmount, line.CreditAmount)
					}
				}
			}
		}
	}

	log.Printf("\n🎉 Balance Verification Completed!")
}