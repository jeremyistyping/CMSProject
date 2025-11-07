package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔍 Debugging Balance Update Issue")

	db := database.ConnectDB()

	// Get the accounts that should be updated
	var cashBankGLAccount models.Account
	if err := db.Where("code = ?", "1102").First(&cashBankGLAccount).Error; err != nil {
		log.Fatalf("❌ Cash account not found: %v", err)
	}

	var arAccount models.Account
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		log.Fatalf("❌ AR account not found: %v", err)
	}

	var cashBank models.CashBank
	if err := db.First(&cashBank, 2).Error; err != nil {
		log.Fatalf("❌ CashBank not found: %v", err)
	}

	log.Printf("🏦 Current Balances:")
	log.Printf("   Cash GL Account (%s): %.2f", cashBankGLAccount.Code, cashBankGLAccount.Balance)
	log.Printf("   AR Account (%s): %.2f", arAccount.Code, arAccount.Balance)
	log.Printf("   CashBank Balance: %.2f", cashBank.Balance)

	// Check the latest SSOT journal entry
	var latestJournal models.SSOTJournalEntry
	if err := db.Order("id DESC").First(&latestJournal).Error; err != nil {
		log.Fatalf("❌ No SSOT journal found: %v", err)
	}

	log.Printf("📝 Latest SSOT Journal: ID=%d, Status=%s, Total Debit=%.2f", 
		latestJournal.ID, latestJournal.Status, latestJournal.TotalDebit.InexactFloat64())

	// Check the journal lines
	var journalLines []models.SSOTJournalLine
	if err := db.Where("journal_id = ?", latestJournal.ID).Find(&journalLines).Error; err != nil {
		log.Fatalf("❌ No journal lines found: %v", err)
	}

	log.Printf("📝 Journal Lines:")
	for i, line := range journalLines {
		var account models.Account
		db.First(&account, line.AccountID)
		log.Printf("   Line %d: Account %s (%s), Debit: %.2f, Credit: %.2f", 
			i+1, account.Name, account.Code, 
			line.DebitAmount.InexactFloat64(), line.CreditAmount.InexactFloat64())
	}

	// Check recent cash bank transactions
	var recentTxns []models.CashBankTransaction
	if err := db.Where("cash_bank_id = ?", 2).Order("id DESC").Limit(3).Find(&recentTxns).Error; err != nil {
		log.Printf("⚠️ No recent cash bank transactions found: %v", err)
	} else {
		log.Printf("💰 Recent CashBank Transactions:")
		for i, txn := range recentTxns {
			log.Printf("   %d. ID=%d, Amount=%.2f, Balance After=%.2f, Date=%s, Notes=%s", 
				i+1, txn.ID, txn.Amount, txn.BalanceAfter, 
				txn.TransactionDate.Format("15:04:05"), txn.Notes)
		}
	}

	// Force refresh accounts by re-reading from DB
	log.Printf("\n🔄 Force refresh from DB...")
	
	if err := db.Table("accounts").Where("id = ?", cashBankGLAccount.ID).First(&cashBankGLAccount).Error; err == nil {
		log.Printf("   Refreshed Cash GL Account Balance: %.2f", cashBankGLAccount.Balance)
	}
	
	if err := db.Table("cash_banks").Where("id = ?", 2).First(&cashBank).Error; err == nil {
		log.Printf("   Refreshed CashBank Balance: %.2f", cashBank.Balance)
	}

	// Check if there are any pending balance updates
	log.Printf("\n🕒 Check account update timestamps:")
	log.Printf("   Cash GL Account updated: %s", cashBankGLAccount.UpdatedAt.Format("15:04:05"))
	log.Printf("   AR Account updated: %s", arAccount.UpdatedAt.Format("15:04:05"))
	log.Printf("   CashBank updated: %s", cashBank.UpdatedAt.Format("15:04:05"))

	log.Printf("✅ Debug completed")
}