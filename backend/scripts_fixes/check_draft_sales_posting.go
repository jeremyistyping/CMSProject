package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔍 Checking DRAFT sales posting to COA...")

	// Initialize database
	db := database.ConnectDB()

	checkDraftSalesJournalEntries(db)
}

func checkDraftSalesJournalEntries(db *gorm.DB) {
	log.Println("\n📊 === CHECKING DRAFT SALES POSTING ===")

	// 1. Count total sales by status
	log.Println("\n1️⃣ Sales Count by Status:")
	var statusCounts []struct {
		Status string
		Count  int64
	}
	
	db.Model(&models.Sale{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts)
	
	for _, sc := range statusCounts {
		log.Printf("   📋 Status: %-12s Count: %d", sc.Status, sc.Count)
	}

	// 2. Check DRAFT sales
	var draftSales []models.Sale
	err := db.Where("status = ?", "DRAFT").Find(&draftSales).Error
	if err != nil {
		log.Printf("❌ Error fetching DRAFT sales: %v", err)
		return
	}

	log.Printf("\n2️⃣ Found %d DRAFT Sales", len(draftSales))

	if len(draftSales) == 0 {
		log.Println("✅ No DRAFT sales found")
		return
	}

	// 3. Check if any DRAFT sales have journal entries
	log.Println("\n3️⃣ Checking Journal Entries for DRAFT Sales:")
	
	journalCount := 0
	for i, sale := range draftSales {
		// Check SSOT Journal
		var ssotJournal models.SimpleSSOTJournal
		err := db.Where("transaction_type = ? AND transaction_id = ?", "SALES", sale.ID).
			First(&ssotJournal).Error
		
		hasJournal := err == nil
		if hasJournal {
			journalCount++
			
			// Get journal items
			var journalItems []models.SimpleSSOTJournalItem
			db.Where("journal_id = ?", ssotJournal.ID).Find(&journalItems)
			
			log.Printf("   🚨 DRAFT Sale #%d (Code: %s) HAS JOURNAL ENTRIES:", 
				sale.ID, sale.Code)
			log.Printf("      📝 Journal ID: %d, Entry: %s", 
				ssotJournal.ID, ssotJournal.EntryNumber)
			log.Printf("      💰 Total Amount: %.2f", sale.TotalAmount)
			log.Printf("      📅 Date: %s", sale.Date.Format("2006-01-02"))
			log.Printf("      💳 Payment Method: %s", sale.PaymentMethodType)
			
			// Show journal items
			for _, item := range journalItems {
				log.Printf("         - Account %d (%s): Debit %.2f, Credit %.2f", 
					item.AccountID, item.AccountName, item.Debit, item.Credit)
			}
			
			// Check COA balance impact
			checkCOAImpact(db, journalItems)
		} else {
			if i < 5 { // Show first 5 only
				log.Printf("   ✅ DRAFT Sale #%d (Code: %s) - No journal entries", 
					sale.ID, sale.Code)
			}
		}
	}
	
	// 4. Summary
	log.Printf("\n4️⃣ SUMMARY:")
	log.Printf("   📊 Total DRAFT Sales: %d", len(draftSales))
	log.Printf("   🚨 DRAFT Sales with Journal Entries: %d", journalCount)
	
	if journalCount > 0 {
		log.Printf("   ❌ PROBLEM DETECTED: %d DRAFT sales have journal entries!", journalCount)
		log.Printf("   🔧 These should be cleaned up or their status should be corrected")
		
		// Suggest solution
		log.Printf("\n💡 SUGGESTED ACTIONS:")
		log.Printf("   1. Clean up journal entries for DRAFT sales")
		log.Printf("   2. Or update status to INVOICED if they should be posted")
		log.Printf("   3. Verify CreateSale logic is not creating journals for DRAFT")
	} else {
		log.Printf("   ✅ GOOD: No DRAFT sales have journal entries")
	}

	// 5. Check COA accounts that might be affected
	log.Println("\n5️⃣ Checking Key COA Account Balances:")
	checkKeyAccountBalances(db)
}

func checkCOAImpact(db *gorm.DB, journalItems []models.SimpleSSOTJournalItem) {
	for _, item := range journalItems {
		var coa models.COA
		if err := db.First(&coa, item.AccountID).Error; err == nil {
			log.Printf("         💳 Account %d (%s) Balance: %.2f (Type: %s)", 
				coa.ID, coa.Name, coa.Balance, coa.Type)
		}
	}
}

func checkKeyAccountBalances(db *gorm.DB) {
	keyAccounts := []uint{1101, 1102, 1201, 4101, 2103} // Kas, Bank, Piutang, Revenue, PPN
	
	for _, accountID := range keyAccounts {
		var coa models.COA
		if err := db.First(&coa, accountID).Error; err == nil {
			// Display balance correctly based on account type
			displayBalance := coa.Balance
			if coa.Type == "REVENUE" || coa.Type == "LIABILITY" {
				displayBalance = -coa.Balance // Convert for display
			}
			
			log.Printf("   💳 %d - %-25s: %15.2f (Type: %s)", 
				coa.ID, coa.Name, displayBalance, coa.Type)
		} else {
			log.Printf("   ❌ Account %d not found", accountID)
		}
	}
}