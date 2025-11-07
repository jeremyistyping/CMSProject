package main

import (
	"log"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

func main() {
	log.Printf("🧪 Testing corrected sales posting logic...")

	// Initialize database connection
	db := database.ConnectDB()

	// Initialize repositories
	salesRepo := repositories.NewSalesRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	journalRepo := repositories.NewJournalEntryRepository(db)

	// Initialize services
	salesDoubleEntryService := services.NewSalesDoubleEntryService(db, journalRepo, accountRepo)
	salesService := services.NewSalesService(db, salesRepo, accountRepo, salesDoubleEntryService)

	log.Printf("\n📊 Initial account balances:")
	checkAccountBalances(db)

	log.Printf("\n🧪 Test 1: Creating DRAFT sale (should NOT post journal entries)")
	draftSale := createTestDraftSale(salesService)
	if draftSale != nil {
		log.Printf("✅ DRAFT sale created: ID %d, Status: %s", draftSale.ID, draftSale.Status)
		log.Printf("📊 Balances after DRAFT sale creation:")
		checkAccountBalances(db)
		checkJournalEntries(db, draftSale.ID)
	}

	log.Printf("\n🧪 Test 2: Converting sale to INVOICED (should post journal entries)")
	if draftSale != nil {
		invoicedSale, err := salesService.CreateInvoiceFromSale(draftSale.ID, 1)
		if err != nil {
			log.Printf("❌ Error creating invoice: %v", err)
		} else {
			log.Printf("✅ Sale converted to INVOICED: ID %d, Status: %s", invoicedSale.ID, invoicedSale.Status)
			log.Printf("📊 Balances after INVOICING (should show changes):")
			checkAccountBalances(db)
			checkJournalEntries(db, invoicedSale.ID)
		}
	}

	log.Printf("\n✅ Testing completed!")
}

func createTestDraftSale(salesService *services.SalesService) *models.Sale {
	// Create test customer if not exists
	// For simplicity, assume customer ID 1 exists or use a mock

	saleRequest := &models.SaleCreateRequest{
		CustomerID:        1, // Assume exists
		Type:             "INVOICE",
		Date:             time.Now(),
		DueDate:          time.Now().AddDate(0, 0, 30),
		PaymentMethodType: "CASH",
		CashBankID:       uintPtr(1), // Assume cash account exists
		Items: []models.SaleItemRequest{
			{
				ProductID:   1, // Assume product exists
				Quantity:    2,
				UnitPrice:   500000,
				TotalPrice:  1000000,
				PPNPercent:  11,
				PPNAmount:   110000,
				FinalAmount: 1110000,
			},
		},
	}

	sale, err := salesService.CreateSale(*saleRequest, 1) // userID 1
	if err != nil {
		log.Printf("❌ Error creating DRAFT sale: %v", err)
		return nil
	}

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

func checkJournalEntries(db *gorm.DB, saleID uint) {
	// Check SSOT journal entries
	var ssotEntries []models.SSOTJournalEntry
	saleIDUint64 := uint64(saleID)
	db.Where("source_type = ? AND source_id = ?", "SALES", &saleIDUint64).Find(&ssotEntries)
	
	log.Printf("   📋 SSOT Journal entries for sale %d: %d entries", saleID, len(ssotEntries))
	
	for _, entry := range ssotEntries {
		log.Printf("      - Entry %d: %s, Status: %s, Amount: %.2f", 
			entry.ID, entry.Description, entry.Status, entry.TotalDebit.InexactFloat64())
	}
	
	// Check legacy journal entries too
	var legacyEntries []models.JournalEntry
	db.Where("reference_type = ? AND reference_id = ?", "SALE", saleID).Find(&legacyEntries)
	
	log.Printf("   📋 Legacy Journal entries for sale %d: %d entries", saleID, len(legacyEntries))
}

func uintPtr(val uint) *uint {
	return &val
}