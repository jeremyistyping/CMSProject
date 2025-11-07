package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔍 Debugging Missing SSOT Journal Lines")

	db := database.ConnectDB()

	// Get the latest SSOT journal entry
	var latestEntry models.SSOTJournalEntry
	if err := db.Order("id DESC").First(&latestEntry).Error; err != nil {
		log.Fatalf("❌ No SSOT journal entries found: %v", err)
	}

	log.Printf("📝 Latest SSOT Journal Entry:")
	log.Printf("   ID: %d, Entry Number: %s", latestEntry.ID, latestEntry.EntryNumber)
	log.Printf("   Source: %s, Source ID: %v", latestEntry.SourceType, latestEntry.SourceID)
	log.Printf("   Status: %s, Total Debit: %.2f", latestEntry.Status, latestEntry.TotalDebit.InexactFloat64())

	// Check all journal lines for this entry
	var journalLines []models.SSOTJournalLine
	if err := db.Where("journal_id = ?", latestEntry.ID).Find(&journalLines).Error; err != nil {
		log.Printf("❌ No journal lines found: %v", err)
		return
	}

	log.Printf("📝 Journal Lines for Entry %d:", latestEntry.ID)
	for i, line := range journalLines {
		var account models.Account
		if err := db.First(&account, line.AccountID).Error; err == nil {
			log.Printf("   %d. Account ID: %d, Code: %s, Name: %s", 
				i+1, line.AccountID, account.Code, account.Name)
			log.Printf("      Debit: %.2f, Credit: %.2f", 
				line.DebitAmount.InexactFloat64(), line.CreditAmount.InexactFloat64())
		} else {
			log.Printf("   %d. Account ID: %d (ACCOUNT NOT FOUND)", i+1, line.AccountID)
			log.Printf("      Debit: %.2f, Credit: %.2f", 
				line.DebitAmount.InexactFloat64(), line.CreditAmount.InexactFloat64())
		}
	}

	// Check if GL Account ID 4 (Bank BCA) is in the journal lines
	targetAccountID := uint64(4)
	found := false
	for _, line := range journalLines {
		if line.AccountID == targetAccountID {
			found = true
			log.Printf("✅ Found line for GL Account ID %d: Debit=%.2f, Credit=%.2f", 
				targetAccountID, line.DebitAmount.InexactFloat64(), line.CreditAmount.InexactFloat64())
			break
		}
	}

	if !found {
		log.Printf("❌ GL Account ID %d NOT FOUND in journal lines", targetAccountID)
	}

	// Check if there are any journal lines affecting account ID 4
	var allLinesForAccount []models.SSOTJournalLine
	if err := db.Where("account_id = ?", targetAccountID).Find(&allLinesForAccount).Error; err == nil {
		log.Printf("📊 All SSOT Journal Lines for Account ID %d:", targetAccountID)
		if len(allLinesForAccount) == 0 {
			log.Printf("   ❌ NO JOURNAL LINES FOUND for this account")
		} else {
			for i, line := range allLinesForAccount {
				log.Printf("   %d. Journal ID: %d, Debit: %.2f, Credit: %.2f", 
					i+1, line.JournalID, line.DebitAmount.InexactFloat64(), line.CreditAmount.InexactFloat64())
			}
		}
	}

	// Check cash bank mapping
	var cashBank models.CashBank
	if err := db.Where("id = ?", 2).First(&cashBank).Error; err == nil {
		log.Printf("🏦 CashBank ID 2 mapping:")
		log.Printf("   Account ID: %d", cashBank.AccountID)
		
		var account models.Account
		if err := db.First(&account, cashBank.AccountID).Error; err == nil {
			log.Printf("   Account Code: %s, Name: %s", account.Code, account.Name)
		}
	}

	// Check if there's a mismatch in account ID usage
	log.Printf("\n🔍 Checking Account ID Usage in Recent Payment:")
	
	var latestPayment models.Payment
	if err := db.Order("id DESC").First(&latestPayment).Error; err == nil {
		if latestPayment.JournalEntryID != nil {
			var entry models.SSOTJournalEntry
			if err := db.First(&entry, *latestPayment.JournalEntryID).Error; err == nil {
				var lines []models.SSOTJournalLine
				if err := db.Where("journal_id = ?", entry.ID).Find(&lines).Error; err == nil {
					log.Printf("💳 Payment %d (Journal %d) used accounts:", latestPayment.ID, entry.ID)
					for _, line := range lines {
						var acc models.Account
						if err := db.First(&acc, line.AccountID).Error; err == nil {
							log.Printf("   Account ID %d: %s (%s)", line.AccountID, acc.Name, acc.Code)
						}
					}
				}
			}
		}
	}

	log.Printf("\n✅ Missing journal lines debugging completed")
}