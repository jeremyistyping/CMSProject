package main

import (
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Println("🗑️  DELETING OLD COGS JOURNAL ENTRIES")
	log.Println("========================================")

	_ = config.LoadConfig()
	db := database.ConnectDB()
	log.Println("✅ Database connected\n")

	// Find COGS journal entries
	var cogsJournals []models.SSOTJournalEntry
	if err := db.Joins("JOIN unified_journal_lines ujl ON ujl.journal_id = unified_journal_ledger.id").
		Joins("JOIN accounts a ON a.id = ujl.account_id").
		Where("unified_journal_ledger.source_type = ?", "SALE").
		Where("a.code = ?", "5101").
		Group("unified_journal_ledger.id").
		Find(&cogsJournals).Error; err != nil {
		log.Fatalf("❌ Failed to find COGS journals: %v", err)
	}

	log.Printf("📊 Found %d COGS journal entries to delete\n", len(cogsJournals))

	if len(cogsJournals) == 0 {
		log.Println("ℹ️  No COGS entries to delete")
		return
	}

	// Show what will be deleted
	log.Println("\n📋 Journals to be deleted:")
	log.Println("   Journal ID | Source      | Entry Number          | Total Debit")
	log.Println("   -----------|-------------|-----------------------|-------------")
	for _, j := range cogsJournals {
		log.Printf("   %-10d | SALE #%-5d | %-21s | Rp %9.2f",
			j.ID, *j.SourceID, j.EntryNumber, j.TotalDebit.InexactFloat64())
	}

	// Auto-proceeding since these are test entries with wrong values
	log.Println("\n⚠️  Deleting old COGS entries...")
	log.Println("   (Auto-proceeding to fix test data with wrong cost prices)")

	// Delete journal lines first, then entries
	successCount := 0
	for _, journal := range cogsJournals {
		// Delete lines
		if err := db.Where("journal_id = ?", journal.ID).Delete(&models.SSOTJournalLine{}).Error; err != nil {
			log.Printf("   ❌ Failed to delete lines for journal #%d: %v", journal.ID, err)
			continue
		}

		// Delete entry
		if err := db.Delete(&journal).Error; err != nil {
			log.Printf("   ❌ Failed to delete journal #%d: %v", journal.ID, err)
			continue
		}

		log.Printf("   ✅ Deleted journal #%d and its lines", journal.ID)
		successCount++
	}

	// Update account balances - reverse the COGS and Inventory
	log.Println("\n💰 Reversing account balances...")
	
	// Get total COGS amount that was deleted
	var deletedCOGS float64
	for _, j := range cogsJournals {
		deletedCOGS += j.TotalDebit.InexactFloat64()
	}

	// Update COGS account (5101) - reduce balance
	var cogsAccount models.Account
	if err := db.Where("code = ?", "5101").First(&cogsAccount).Error; err == nil {
		cogsAccount.Balance -= deletedCOGS
		if err := db.Save(&cogsAccount).Error; err != nil {
			log.Printf("   ⚠️  Failed to update COGS account: %v", err)
		} else {
			log.Printf("   ✅ COGS account (5101) balance: Rp %.2f (reduced by Rp %.2f)", 
				cogsAccount.Balance, deletedCOGS)
		}
	}

	// Update Inventory account (1301) - increase balance
	var inventoryAccount models.Account
	if err := db.Where("code = ?", "1301").First(&inventoryAccount).Error; err == nil {
		inventoryAccount.Balance += deletedCOGS
		if err := db.Save(&inventoryAccount).Error; err != nil {
			log.Printf("   ⚠️  Failed to update Inventory account: %v", err)
		} else {
			log.Printf("   ✅ Inventory account (1301) balance: Rp %.2f (increased by Rp %.2f)", 
				inventoryAccount.Balance, deletedCOGS)
		}
	}

	log.Println("\n========================================")
	log.Printf("✅ Successfully deleted %d/%d COGS entries\n", successCount, len(cogsJournals))
	log.Println("\n💡 Next step:")
	log.Println("   Run: go run cmd/scripts/backfill_missing_cogs.go")
	log.Println("   To recreate COGS with correct cost prices")
	log.Println("========================================")
}
