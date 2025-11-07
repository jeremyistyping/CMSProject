package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	log.Printf("🎯 SIMPLE DIRECT TEST: Verifying fixes...")

	// Initialize database connection
	db := database.ConnectDB()

	log.Printf("\n📊 BEFORE - Account balances:")
	printAccountBalances(db)

	log.Printf("\n=== Creating a DRAFT sale with CASH payment ===")
	
	// Create a DRAFT sale directly in database
	draftSale := &models.Sale{
		Code:              fmt.Sprintf("TEST-%d", time.Now().Unix()%100000),
		CustomerID:        1,
		UserID:           1,
		Type:             "INVOICE",
		Status:           models.SaleStatusDraft, // DRAFT status
		Date:             time.Now(),
		DueDate:          time.Now().AddDate(0, 0, 30),
		PaymentMethodType: "CASH", // CASH payment
		CashBankID:       uintPtr(1),
		TotalAmount:      1110000,
		PPN:              110000,
		Subtotal:         1000000,
		Currency:         "IDR",
		ExchangeRate:     1,
	}
	
	err := db.Create(draftSale).Error
	if err != nil {
		log.Printf("❌ Error creating sale: %v", err)
		return
	}
	
	log.Printf("📄 Created DRAFT sale: ID %d, Status: %s, Payment: %s", 
		draftSale.ID, draftSale.Status, draftSale.PaymentMethodType)

	log.Printf("\n📊 AFTER creating DRAFT sale - Account balances:")
	printAccountBalances(db)
	
	log.Printf("\n🔍 Checking for journal entries for DRAFT sale...")
	checkJournalEntries(db, draftSale.ID)
	
	log.Printf("\n=== RESULTS ===")
	log.Printf("1. Check if Piutang Usaha (1201) increased for a CASH sale")
	log.Printf("2. Check if Revenue (4101) and PPN (2103) changed for DRAFT sale")
	log.Printf("3. Check if any journal entries were created for DRAFT sale")
	
	log.Printf("\n📊 Expected behavior:")
	log.Printf("   ✅ NO changes to any account balances for DRAFT sale")
	log.Printf("   ✅ NO journal entries created for DRAFT sale")
	log.Printf("   ✅ Cash payment should NOT affect Piutang Usaha")
	log.Printf("   ✅ Only INVOICED sales should create journal entries")
}

func printAccountBalances(db *gorm.DB) {
	accounts := []string{"1101", "1201", "4101", "2103"}
	
	for _, code := range accounts {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			log.Printf("   %s: NOT FOUND", code)
			continue
		}
		
		var name string
		switch code {
		case "1101":
			name = "Kas"
		case "1201":
			name = "Piutang Usaha (AR)"
		case "4101":
			name = "Revenue"
		case "2103":
			name = "PPN Keluaran"
		}
		
		log.Printf("   %s (%s): %.2f", code, name, account.Balance)
	}
}

func checkJournalEntries(db *gorm.DB, saleID uint) {
	// Check legacy journal entries
	var legacyCount int64
	db.Model(&models.JournalEntry{}).
		Where("reference_type = ? AND reference_id = ?", "SALE", saleID).
		Count(&legacyCount)
	
	// Check SSOT journal entries
	var ssotCount int64
	saleIDUint64 := uint64(saleID)
	db.Table("ssot_journal_entries").
		Where("source_type = ? AND source_id = ?", "SALES", saleIDUint64).
		Count(&ssotCount)
	
	// Check unified journal entries
	var unifiedCount int64
	db.Table("unified_journal_entries").
		Where("source_type = ? AND source_id = ?", "SALE", saleID).
		Count(&unifiedCount)
	
	log.Printf("   Legacy journal entries: %d", legacyCount)
	log.Printf("   SSOT journal entries: %d", ssotCount)
	log.Printf("   Unified journal entries: %d", unifiedCount)
	
	totalEntries := legacyCount + ssotCount + unifiedCount
	
	if totalEntries == 0 {
		log.Printf("   ✅ CORRECT: No journal entries for DRAFT sale")
	} else {
		log.Printf("   ❌ ERROR: Found %d journal entries for DRAFT sale!", totalEntries)
	}
}

func uintPtr(val uint) *uint {
	return &val
}