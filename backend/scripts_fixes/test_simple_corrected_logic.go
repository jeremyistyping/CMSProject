package main

import (
	"log"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"gorm.io/gorm"
)

func main() {
	log.Printf("🧪 Testing corrected SSOT sales logic...")

	// Initialize database connection
	db := database.ConnectDB()

	// Initialize service
	ssotService := services.NewCorrectedSSOTSalesJournalService(db)

	log.Printf("\n📊 Initial account balances:")
	checkAccountBalances(db)

	// Create a test sale directly in database
	testSale := createTestSale(db)

	log.Printf("\n🧪 Test 1: Try creating journal for DRAFT sale (should FAIL)")
	_, err := ssotService.CreateSaleJournalEntry(testSale, 1)
	if err != nil {
		log.Printf("✅ CORRECT: DRAFT sale rejected: %v", err)
	} else {
		log.Printf("❌ ERROR: DRAFT sale should be rejected!")
	}

	log.Printf("\n🧪 Test 2: Update sale to INVOICED and create journal (should SUCCESS)")
	// Update sale to INVOICED
	testSale.Status = models.SaleStatusInvoiced
	testSale.InvoiceNumber = "INV-TEST-001"
	db.Save(testSale)

	entry, err := ssotService.CreateSaleJournalEntry(testSale, 1)
	if err != nil {
		log.Printf("❌ ERROR creating journal for INVOICED sale: %v", err)
	} else {
		log.Printf("✅ SUCCESS: INVOICED sale journal created: Entry %d", entry.ID)
		log.Printf("📊 Balances after INVOICED journal entry:")
		checkAccountBalances(db)
		checkJournalLines(db, uint(entry.ID))
	}

	log.Printf("\n✅ Testing completed!")
}

func createTestSale(db *gorm.DB) *models.Sale {
	sale := &models.Sale{
		Code:              "TST-001",
		CustomerID:        1, // Assume exists
		UserID:           1,
		Type:             "INVOICE", 
		Status:           models.SaleStatusDraft, // Start as DRAFT
		Date:             time.Now(),
		DueDate:          time.Now().AddDate(0, 0, 30),
		PaymentMethodType: "CASH",
		CashBankID:       uintPtr(1), // Assume cash account exists
		TotalAmount:      1110000,
		PPN:              110000,
		Currency:         "IDR",
		ExchangeRate:     1,
	}

	db.Create(sale)
	log.Printf("📄 Created test sale: ID %d, Status: %s", sale.ID, sale.Status)
	return sale
}

func checkAccountBalances(db *gorm.DB) {
	targetCodes := []string{"4101", "2103", "1101", "1201"}
	
	var accounts []models.Account
	db.Where("code IN ?", targetCodes).Find(&accounts)
	
	for _, acc := range accounts {
		log.Printf("   Account %s (%s): Balance %.2f", acc.Code, acc.Name, acc.Balance)
	}
}

func checkJournalLines(db *gorm.DB, entryID uint) {
	var lines []models.SSOTJournalLine
	db.Preload("Account").Where("journal_entry_id = ?", entryID).Find(&lines)
	
	log.Printf("   📋 Journal lines for entry %d:", entryID)
	for _, line := range lines {
		log.Printf("      - %s: Debit %.2f, Credit %.2f", 
			line.Account.Name, 
			line.DebitAmount.InexactFloat64(), 
			line.CreditAmount.InexactFloat64())
	}
}

func uintPtr(val uint) *uint {
	return &val
}