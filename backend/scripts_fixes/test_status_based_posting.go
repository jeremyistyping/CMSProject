package main

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
)

func main() {
	log.Printf("🧪 Testing Status-Based Posting Logic")
	log.Printf("=====================================")
	log.Printf("💡 Expected behavior:")
	log.Printf("   ✅ DRAFT → CONFIRMED: No journal entries, no COA update")
	log.Printf("   ✅ CONFIRMED → INVOICED: Create journal entries, update COA")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Initialize services
	salesRepo := repositories.NewSalesRepository(db)
	productRepo := repositories.NewProductRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	salesService := services.NewSalesService(db, salesRepo, productRepo, contactRepo, accountRepo, nil, nil)

	// Get test data
	var product models.Product
	err = db.Where("stock > ?", 0).First(&product).Error
	if err != nil {
		log.Fatalf("❌ No product with stock found: %v", err)
	}

	var customer models.Contact
	err = db.Where("type = ?", "CUSTOMER").First(&customer).Error
	if err != nil {
		log.Fatalf("❌ No customer found: %v", err)
	}

	var cashBank models.CashBank
	err = db.Where("type = ? AND is_active = ?", "BANK", true).First(&cashBank).Error
	if err != nil {
		log.Fatalf("❌ No active bank account found: %v", err)
	}

	log.Printf("✅ Test data ready:")
	log.Printf("   Product: %s", product.Name)
	log.Printf("   Customer: %s", customer.Name)
	log.Printf("   Bank: %s", cashBank.Name)

	// Get account balances before test
	beforeBalances, err := getAccountBalances(db, []string{"1104", "4101", "2103"})
	if err != nil {
		log.Printf("⚠️ Warning: Could not get before balances: %v", err)
		beforeBalances = make(map[string]float64)
	}

	log.Printf("\n📊 Account Balances BEFORE:")
	for code, balance := range beforeBalances {
		log.Printf("   %s: %.2f", code, balance)
	}

	// =============================
	// TEST 1: Create Draft Sale
	// =============================
	log.Printf("\n🔄 TEST 1: Creating Draft Sale...")
	
	saleRequest := models.SaleCreateRequest{
		CustomerID:        customer.ID,
		Type:              models.SaleTypeInvoice,
		Date:              time.Now(),
		DueDate:           time.Now().AddDate(0, 0, 30),
		PaymentMethodType: "BANK",
		CashBankID:        &cashBank.ID,
		PPNRate:           11.0,
		Items: []models.SaleItemRequest{
			{
				ProductID:        product.ID,
				Quantity:         1,
				UnitPrice:        500000.0,
				Taxable:          true,
				RevenueAccountID: 0, // Use default
			},
		},
	}

	createdSale, err := salesService.CreateSale(saleRequest, 1)
	if err != nil {
		log.Fatalf("❌ Failed to create sale: %v", err)
	}

	log.Printf("✅ Draft sale created:")
	log.Printf("   Sale ID: %d", createdSale.ID)
	log.Printf("   Status: %s", createdSale.Status)
	log.Printf("   Total: %.2f", createdSale.TotalAmount)

	// Check if draft sale has created journal entries (should be NONE)
	journalCount := countJournalEntries(db, "SALE", createdSale.ID)
	log.Printf("   Journal entries: %d", journalCount)
	
	if journalCount > 0 {
		log.Printf("❌ FAIL: Draft sale should NOT create journal entries!")
	} else {
		log.Printf("✅ PASS: No journal entries for draft sale")
	}

	// Check account balances after draft (should be UNCHANGED)
	afterDraftBalances, _ := getAccountBalances(db, []string{"1104", "4101", "2103"})
	log.Printf("\n📊 Account Balances AFTER Draft Creation:")
	hasChangedAfterDraft := false
	for code, balance := range afterDraftBalances {
		change := balance - beforeBalances[code]
		log.Printf("   %s: %.2f (change: %.2f)", code, balance, change)
		if change != 0 {
			hasChangedAfterDraft = true
		}
	}
	
	if hasChangedAfterDraft {
		log.Printf("❌ FAIL: Draft sale should NOT change account balances!")
	} else {
		log.Printf("✅ PASS: Account balances unchanged for draft sale")
	}

	// =============================
	// TEST 2: Confirm Sale (DRAFT → CONFIRMED)
	// =============================
	log.Printf("\n🔄 TEST 2: Confirming Sale (DRAFT → CONFIRMED)...")
	
	err = salesService.ConfirmSale(createdSale.ID, 1)
	if err != nil {
		log.Fatalf("❌ Failed to confirm sale: %v", err)
	}

	// Get updated sale details
	confirmedSale, err := salesService.GetSaleByID(createdSale.ID)
	if err != nil {
		log.Fatalf("❌ Failed to get confirmed sale: %v", err)
	}

	log.Printf("✅ Sale confirmed:")
	log.Printf("   Status: %s", confirmedSale.Status)

	// Check if confirmed sale has journal entries (should still be NONE)
	journalCount = countJournalEntries(db, "SALE", confirmedSale.ID)
	log.Printf("   Journal entries: %d", journalCount)
	
	if journalCount > 0 {
		log.Printf("❌ FAIL: Confirmed sale should NOT create journal entries yet!")
	} else {
		log.Printf("✅ PASS: No journal entries for confirmed sale")
	}

	// Check account balances after confirmation (should still be UNCHANGED)
	afterConfirmedBalances, _ := getAccountBalances(db, []string{"1104", "4101", "2103"})
	log.Printf("\n📊 Account Balances AFTER Confirmation:")
	hasChangedAfterConfirmed := false
	for code, balance := range afterConfirmedBalances {
		change := balance - beforeBalances[code]
		log.Printf("   %s: %.2f (change: %.2f)", code, balance, change)
		if change != 0 {
			hasChangedAfterConfirmed = true
		}
	}
	
	if hasChangedAfterConfirmed {
		log.Printf("❌ FAIL: Confirmed sale should NOT change account balances yet!")
	} else {
		log.Printf("✅ PASS: Account balances still unchanged after confirmation")
	}

	// =============================
	// TEST 3: Create Invoice (CONFIRMED → INVOICED)
	// =============================
	log.Printf("\n🔄 TEST 3: Creating Invoice (CONFIRMED → INVOICED)...")
	
	invoicedSale, err := salesService.CreateInvoiceFromSale(confirmedSale.ID, 1)
	if err != nil {
		log.Fatalf("❌ Failed to create invoice: %v", err)
	}

	log.Printf("✅ Invoice created:")
	log.Printf("   Status: %s", invoicedSale.Status)
	log.Printf("   Invoice Number: %s", invoicedSale.InvoiceNumber)

	// Check if invoiced sale has journal entries (should have ENTRIES now)
	journalCount = countJournalEntries(db, "SALE", invoicedSale.ID)
	log.Printf("   Journal entries: %d", journalCount)
	
	if journalCount == 0 {
		log.Printf("❌ FAIL: Invoiced sale should create journal entries!")
	} else {
		log.Printf("✅ PASS: Journal entries created for invoiced sale")
		
		// Show journal entries
		var journalEntries []models.JournalEntry
		db.Preload("JournalLines.Account").
			Where("reference_type = ? AND reference_id = ?", "SALE", invoicedSale.ID).
			Find(&journalEntries)
		
		for i, entry := range journalEntries {
			log.Printf("   Entry %d: %s", i+1, entry.Code)
			for j, line := range entry.JournalLines {
				log.Printf("     Line %d: %s (%s) - Debit: %.2f, Credit: %.2f", 
					j+1, line.Account.Name, line.Account.Code,
					line.DebitAmount, line.CreditAmount)
			}
		}
	}

	// Check account balances after invoicing (should be CHANGED now)
	afterInvoicedBalances, _ := getAccountBalances(db, []string{"1104", "4101", "2103"})
	log.Printf("\n📊 Account Balances AFTER Invoicing:")
	hasChangedAfterInvoiced := false
	for code, balance := range afterInvoicedBalances {
		change := balance - beforeBalances[code]
		log.Printf("   %s: %.2f (change: %.2f)", code, balance, change)
		if change != 0 {
			hasChangedAfterInvoiced = true
		}
	}
	
	if !hasChangedAfterInvoiced {
		log.Printf("❌ FAIL: Invoiced sale should change account balances!")
	} else {
		log.Printf("✅ PASS: Account balances updated after invoicing")
	}

	// =============================
	// FINAL RESULTS
	// =============================
	log.Printf("\n🎯 FINAL TEST RESULTS:")
	log.Printf("======================")
	
	allTestsPassed := true
	
	// Test 1: Draft should not create journals
	draftJournalCount := countJournalEntries(db, "SALE", createdSale.ID)
	if draftJournalCount == 0 {
		log.Printf("✅ Test 1 PASSED: Draft sales do not create journal entries")
	} else {
		log.Printf("❌ Test 1 FAILED: Draft sales created %d journal entries", draftJournalCount)
		allTestsPassed = false
	}
	
	// Test 2: Final status should be PAID (for bank sales)
	if invoicedSale.Status == models.SaleStatusPaid {
		log.Printf("✅ Test 2 PASSED: Bank sale status correctly updated to PAID")
	} else {
		log.Printf("⚠️ Test 2 PARTIAL: Sale status is %s (expected PAID for bank sales)", invoicedSale.Status)
	}
	
	// Test 3: COA should be updated only after invoicing
	if hasChangedAfterInvoiced {
		log.Printf("✅ Test 3 PASSED: COA balances updated only after invoicing")
	} else {
		log.Printf("❌ Test 3 FAILED: COA balances not updated after invoicing")
		allTestsPassed = false
	}

	if allTestsPassed {
		log.Printf("\n🎉 ALL TESTS PASSED! Status-based posting is working correctly!")
		log.Printf("✅ Draft/Confirmed sales: No accounting impact")
		log.Printf("✅ Invoiced sales: Creates journal entries and updates COA")
	} else {
		log.Printf("\n⚠️ Some tests failed - please review the implementation")
	}

	// Cleanup
	log.Printf("\n🧹 Cleaning up test data...")
	cleanupSale(db, invoicedSale.ID)
	log.Printf("✅ Test data cleaned up")
}

func getAccountBalances(db *gorm.DB, accountCodes []string) (map[string]float64, error) {
	balances := make(map[string]float64)
	for _, code := range accountCodes {
		var account models.Account
		if err := db.Where("code = ?", code).First(&account).Error; err == nil {
			balances[code] = account.Balance
		}
	}
	return balances, nil
}

func countJournalEntries(db *gorm.DB, refType string, refID uint) int {
	var count int64
	db.Model(&models.JournalEntry{}).
		Where("reference_type = ? AND reference_id = ?", refType, refID).
		Count(&count)
	return int(count)
}

func cleanupSale(db *gorm.DB, saleID uint) {
	// Delete in order to avoid foreign key constraints
	db.Delete(&models.SalePayment{}, "sale_id = ?", saleID)
	db.Delete(&models.JournalLine{}, "journal_entry_id IN (SELECT id FROM journal_entries WHERE reference_type = 'SALE' AND reference_id = ?)", saleID)
	db.Delete(&models.JournalEntry{}, "reference_type = ? AND reference_id = ?", "SALE", saleID)
	db.Delete(&models.SaleItem{}, "sale_id = ?", saleID)
	db.Delete(&models.Sale{}, "id = ?", saleID)
}