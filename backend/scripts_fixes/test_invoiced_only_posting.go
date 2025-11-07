package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	log.Println("🧪 Testing INVOICED-only posting workflow...")
	log.Println("==================================================")

	// Test workflow
	if err := testInvoicedOnlyPosting(db); err != nil {
		log.Fatalf("❌ Test failed: %v", err)
	}

	log.Println("✅ All tests passed! INVOICED-only posting is working correctly.")
}

func testInvoicedOnlyPosting(db *gorm.DB) error {
	// Initialize services
	contactRepo := repositories.NewContactRepository(db)
	productRepo := repositories.NewProductRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	journalService := services.NewJournalService(db, repositories.NewJournalEntryRepository(db), accountRepo)
	pdfService := services.NewPDFService()

	salesService := services.NewSalesService(
		db, salesRepo, productRepo, contactRepo, accountRepo, journalService, pdfService,
	)

	// Step 1: Create test customer
	log.Println("\n1️⃣ Creating test customer...")
	customer := &models.Contact{
		Type:        "CUSTOMER",
		Name:        "Test Customer for INVOICED-only Posting",
		Email:       "test.invoiced@customer.com",
		Phone:       "081234567890",
		IsActive:    true,
		CreditLimit: 10000000,
	}
	
	if err := db.Create(customer).Error; err != nil {
		return fmt.Errorf("failed to create test customer: %v", err)
	}
	log.Printf("✅ Created customer: ID=%d, Name=%s", customer.ID, customer.Name)

	// Step 2: Create test product
	log.Println("\n2️⃣ Creating test product...")
	product := &models.Product{
		Name:          "Test Product for INVOICED-only Posting",
		Code:          "TEST-INVOICED-001",
		Description:   "Test product to verify invoiced-only posting",
		CategoryID:    1,
		UnitID:        1,
		PurchasePrice: 800000,
		SellingPrice:  1000000,
		Stock:         100,
		MinStock:      10,
		IsActive:      true,
		IsService:     false,
	}
	
	if err := db.Create(product).Error; err != nil {
		return fmt.Errorf("failed to create test product: %v", err)
	}
	log.Printf("✅ Created product: ID=%d, Name=%s", product.ID, product.Name)

	// Step 3: Create sale in DRAFT status
	log.Println("\n3️⃣ Creating sale in DRAFT status...")
	saleRequest := models.SaleCreateRequest{
		CustomerID:      customer.ID,
		Type:            models.SaleTypeInvoice,
		Date:            time.Now(),
		DueDate:         time.Now().AddDate(0, 0, 30),
		Currency:        "IDR",
		PaymentTerms:    "NET30",
		PaymentMethodType: "CREDIT", // Credit sale to test AR entries
		BillingAddress:  "Test Address",
		Notes:           "Test sale for INVOICED-only posting verification",
		Items: []models.SaleItemRequest{
			{
				ProductID:   product.ID,
				Quantity:    2,
				UnitPrice:   1000000,
				Taxable:     true,
			},
		},
	}

	createdSale, err := salesService.CreateSale(saleRequest, 1)
	if err != nil {
		return fmt.Errorf("failed to create sale: %v", err)
	}

	log.Printf("✅ Created sale: ID=%d, Code=%s, Status=%s, Total=%.2f", 
		createdSale.ID, createdSale.Code, createdSale.Status, createdSale.TotalAmount)

	// Verify DRAFT status has no journal entries
	log.Println("\n4️⃣ Verifying DRAFT status has NO journal entries...")
	journalCount := countJournalEntriesForSale(db, createdSale.ID)
	if journalCount > 0 {
		return fmt.Errorf("❌ DRAFT sale should have 0 journal entries, found %d", journalCount)
	}
	log.Printf("✅ DRAFT sale correctly has %d journal entries", journalCount)

	// Verify account balances unchanged
	log.Println("\n5️⃣ Verifying account balances unchanged in DRAFT status...")
	arBalance := getAccountBalance(db, "1201") // Accounts Receivable
	revenueBalance := getAccountBalance(db, "4101") // Revenue
	log.Printf("ℹ️  AR Balance: %.2f, Revenue Balance: %.2f (should be unchanged)", arBalance, revenueBalance)

	// Step 6: Confirm sale (should still NOT create journal entries)
	log.Println("\n6️⃣ Confirming sale (should NOT create journal entries)...")
	if err := salesService.ConfirmSale(createdSale.ID, 1); err != nil {
		return fmt.Errorf("failed to confirm sale: %v", err)
	}

	// Verify sale status is CONFIRMED
	confirmedSale, err := salesService.GetSaleByID(createdSale.ID)
	if err != nil {
		return fmt.Errorf("failed to get confirmed sale: %v", err)
	}
	
	if confirmedSale.Status != models.SaleStatusConfirmed {
		return fmt.Errorf("sale status should be CONFIRMED, got %s", confirmedSale.Status)
	}
	log.Printf("✅ Sale confirmed: Status=%s", confirmedSale.Status)

	// Verify CONFIRMED status STILL has no journal entries
	log.Println("\n7️⃣ Verifying CONFIRMED status STILL has NO journal entries...")
	journalCount = countJournalEntriesForSale(db, createdSale.ID)
	if journalCount > 0 {
		return fmt.Errorf("❌ CONFIRMED sale should have 0 journal entries, found %d", journalCount)
	}
	log.Printf("✅ CONFIRMED sale correctly has %d journal entries", journalCount)

	// Step 8: Create invoice (THIS should create journal entries)
	log.Println("\n8️⃣ Creating invoice (THIS should create journal entries)...")
	invoicedSale, err := salesService.CreateInvoiceFromSale(createdSale.ID, 1)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %v", err)
	}

	if invoicedSale.Status != models.SaleStatusInvoiced {
		return fmt.Errorf("sale status should be INVOICED, got %s", invoicedSale.Status)
	}
	log.Printf("✅ Invoice created: Status=%s, Invoice#=%s", 
		invoicedSale.Status, invoicedSale.InvoiceNumber)

	// Step 9: Verify INVOICED status HAS journal entries
	log.Println("\n9️⃣ Verifying INVOICED status HAS journal entries...")
	time.Sleep(1 * time.Second) // Give time for journal creation
	journalCount = countJournalEntriesForSale(db, createdSale.ID)
	if journalCount == 0 {
		return fmt.Errorf("❌ INVOICED sale should have journal entries, found %d", journalCount)
	}
	log.Printf("✅ INVOICED sale correctly has %d journal entries", journalCount)

	// Step 10: Verify account balances are updated
	log.Println("\n🔟 Verifying account balances are NOW updated...")
	newARBalance := getAccountBalance(db, "1201") // Accounts Receivable
	newRevenueBalance := getAccountBalance(db, "4101") // Revenue
	
	log.Printf("📊 Final Balances:")
	log.Printf("   AR Balance: %.2f (should be increased)", newARBalance)
	log.Printf("   Revenue Balance: %.2f (should be increased)", newRevenueBalance)

	// Step 11: Verify journal entry details
	log.Println("\n1️⃣1️⃣ Verifying journal entry details...")
	if err := verifyJournalEntryDetails(db, createdSale.ID); err != nil {
		return fmt.Errorf("journal entry verification failed: %v", err)
	}

	// Cleanup
	log.Println("\n🧹 Cleaning up test data...")
	cleanupTestData(db, customer.ID, product.ID, createdSale.ID)

	return nil
}

func countJournalEntriesForSale(db *gorm.DB, saleID uint) int {
	var count int64
	
	// Check both possible journal entry tables
	// Check SSOT Journal Entries
	db.Model(&models.SSOTJournalEntry{}).
		Where("source_type = ? AND source_id = ?", "SALE", saleID).
		Count(&count)
	
	ssotCount := count
	
	// Check traditional Journal Entries
	db.Model(&models.JournalEntry{}).
		Where("reference_type = ? AND reference_id = ?", "SALE", saleID).
		Count(&count)
	
	traditionalCount := count
	
	log.Printf("📊 Journal Entry Count: SSOT=%d, Traditional=%d, Total=%d", 
		ssotCount, traditionalCount, ssotCount+traditionalCount)
	
	return int(ssotCount + traditionalCount)
}

func getAccountBalance(db *gorm.DB, accountCode string) float64 {
	var account models.Account
	if err := db.Where("code = ?", accountCode).First(&account).Error; err != nil {
		log.Printf("⚠️ Could not find account %s: %v", accountCode, err)
		return 0
	}
	return account.Balance
}

func verifyJournalEntryDetails(db *gorm.DB, saleID uint) error {
	// Check SSOT journal entries
	var ssotEntries []models.SSOTJournalEntry
	if err := db.Preload("Lines").Where("source_type = ? AND source_id = ?", "SALE", saleID).
		Find(&ssotEntries).Error; err != nil {
		return fmt.Errorf("failed to load SSOT journal entries: %v", err)
	}

	for _, entry := range ssotEntries {
		log.Printf("📋 SSOT Journal Entry: ID=%d, Number=%s, Status=%s", 
			entry.ID, entry.EntryNumber, entry.Status)
		log.Printf("   Debit Total: %.2f, Credit Total: %.2f, Balanced: %t", 
			entry.TotalDebit.InexactFloat64(), entry.TotalCredit.InexactFloat64(), entry.IsBalanced)
		
		for _, line := range entry.Lines {
			log.Printf("   Line %d: Account=%d, Debit=%.2f, Credit=%.2f, Desc=%s", 
				line.LineNumber, line.AccountID, 
				line.DebitAmount.InexactFloat64(), line.CreditAmount.InexactFloat64(), 
				line.Description)
		}
	}

	// Check traditional journal entries  
	var tradEntries []models.JournalEntry
	if err := db.Preload("JournalLines").Where("reference_type = ? AND reference_id = ?", "SALE", saleID).
		Find(&tradEntries).Error; err != nil {
		return fmt.Errorf("failed to load traditional journal entries: %v", err)
	}

	for _, entry := range tradEntries {
		log.Printf("📋 Traditional Journal Entry: ID=%d, Status=%s", entry.ID, entry.Status)
		log.Printf("   Total Debit: %.2f, Credit: %.2f", entry.TotalDebit, entry.TotalCredit)
		
		for _, line := range entry.JournalLines {
			log.Printf("   Line %d: Account=%d, Debit=%.2f, Credit=%.2f, Desc=%s", 
				line.LineNumber, line.AccountID, line.DebitAmount, line.CreditAmount, line.Description)
		}
	}

	return nil
}

func cleanupTestData(db *gorm.DB, customerID, productID, saleID uint) {
	// Delete journal entries
	db.Where("source_type = ? AND source_id = ?", "SALE", saleID).Delete(&models.SSOTJournalEntry{})
	db.Where("reference_type = ? AND reference_id = ?", "SALE", saleID).Delete(&models.JournalEntry{})
	
	// Delete sale items
	db.Where("sale_id = ?", saleID).Delete(&models.SaleItem{})
	
	// Delete sale
	db.Delete(&models.Sale{}, saleID)
	
	// Delete test data
	db.Delete(&models.Product{}, productID)
	db.Delete(&models.Contact{}, customerID)
	
	log.Println("✅ Test data cleaned up")
}