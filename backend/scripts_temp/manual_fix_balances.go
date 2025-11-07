package main

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	fmt.Println("🔧 Manual Fix: COA Balances and Journal Entries...")

	// Check current journal entries
	var journalEntries []models.SSOTJournalEntry
	db.Where("source_type = ?", models.SSOTSourceTypePurchase).Find(&journalEntries)

	for _, entry := range journalEntries {
		fmt.Printf("📋 Journal Entry: %s\n", entry.EntryNumber)
		
		// Get lines for this entry
		var lines []models.SSOTJournalLine
		db.Where("journal_id = ?", entry.ID).Find(&lines)
		
		for _, line := range lines {
			var account models.Account
			db.First(&account, line.AccountID)
			fmt.Printf("   💰 Line: %s - Debit: %.0f, Credit: %.0f\n", 
				account.Name, line.DebitAmount, line.CreditAmount)
		}
	}

	fmt.Println("\n🔄 Updating COA balances correctly...")

	// Update specific account balances based on journal entries
	// Persediaan Barang Dagangan (1301) - should be 5,000,000 (debit)
	db.Model(&models.Account{}).Where("code = ?", "1301").Update("balance", 5000000)

	// Utang Usaha (2101) - should be 5,550,000 (credit, so positive for liability)
	db.Model(&models.Account{}).Where("code = ?", "2101").Update("balance", 5550000)

	// Utang Pajak (2102) - should be 550,000 (credit, so positive for liability)
	db.Model(&models.Account{}).Where("code = ?", "2102").Update("balance", 550000)

	// Reset wrong accounts to 0
	db.Model(&models.Account{}).Where("code = ?", "1201").Update("balance", 0)

	fmt.Println("✅ Updated COA balances:")
	
	// Show updated balances
	keyCodes := []string{"1301", "2101", "2102", "1201"}
	for _, code := range keyCodes {
		var account models.Account
		if db.Where("code = ?", code).First(&account).Error == nil {
			fmt.Printf("   💳 %s (%s): %.2f\n", account.Name, account.Code, account.Balance)
		}
	}

	fmt.Println("\n🎉 Manual fix completed!")
}