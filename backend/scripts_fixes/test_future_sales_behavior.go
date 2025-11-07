package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"gorm.io/gorm"
)

func main() {
	log.Printf("🔮 Testing future sales creation behavior...")

	// Initialize database connection
	db := database.ConnectDB()

	// Record initial balances
	log.Printf("\n📊 BEFORE creating new sales - Account balances:")
	initialBalances := getAccountBalances(db)
	printBalances(initialBalances)

	log.Printf("\n🧪 Test 1: Create multiple DRAFT sales (should NOT affect balances)")
	for i := 1; i <= 3; i++ {
		testSale := createNewDraftSale(db, i)
		log.Printf("   ✅ Created DRAFT sale %d: ID %d, Status: %s", i, testSale.ID, testSale.Status)
	}

	log.Printf("\n📊 AFTER creating 3 DRAFT sales - Account balances:")
	afterDraftBalances := getAccountBalances(db)
	printBalances(afterDraftBalances)

	// Check if balances changed
	if balancesEqual(initialBalances, afterDraftBalances) {
		log.Printf("✅ PERFECT: DRAFT sales did NOT change account balances")
	} else {
		log.Printf("❌ ERROR: DRAFT sales incorrectly changed account balances!")
		printBalanceDifferences(initialBalances, afterDraftBalances)
	}

	// Check for unwanted journal entries
	checkUnwantedJournalEntries(db)

	log.Printf("\n🧪 Test 2: Verify sales service workflow is protected")
	testFullSalesWorkflow(db)

	log.Printf("\n🔮 Future protection verification completed!")
}

func createNewDraftSale(db *gorm.DB, index int) *models.Sale {
	sale := &models.Sale{
		Code:              fmt.Sprintf("FUTURE-%03d", index),
		CustomerID:        1, // Assume exists
		UserID:           1,
		Type:             "INVOICE", 
		Status:           models.SaleStatusDraft, // Always starts as DRAFT
		Date:             time.Now(),
		DueDate:          time.Now().AddDate(0, 0, 30),
		PaymentMethodType: "BANK", // Test with different payment methods
		CashBankID:       uintPtr(2), // Different account
		TotalAmount:      500000 * float64(index), // Different amounts
		PPN:              55000 * float64(index),
		Currency:         "IDR",
		ExchangeRate:     1,
	}

	result := db.Create(sale)
	if result.Error != nil {
		log.Printf("❌ Error creating sale: %v", result.Error)
		return nil
	}
	
	return sale
}

func testFullSalesWorkflow(db *gorm.DB) {
	log.Printf("\n   Testing protected sales service workflow...")
	
	// Initialize services (this might trigger problematic services)
	ssotService := services.NewCorrectedSSOTSalesJournalService(db)
	
	// Create a test sale and try various statuses
	testSale := &models.Sale{
		Code:              "WORKFLOW-TEST",
		CustomerID:        1,
		UserID:           1,
		Type:             "INVOICE", 
		Status:           models.SaleStatusDraft,
		Date:             time.Now(),
		TotalAmount:      1000000,
		PPN:              100000,
		Currency:         "IDR",
		ExchangeRate:     1,
		PaymentMethodType: "CREDIT",
	}
	
	db.Create(testSale)
	beforeWorkflowBalances := getAccountBalances(db)
	
	// Test 1: Try DRAFT (should fail)
	_, err := ssotService.CreateSaleJournalEntry(testSale, 1)
	if err != nil {
		log.Printf("   ✅ DRAFT correctly rejected: %v", err)
	} else {
		log.Printf("   ❌ DRAFT incorrectly accepted!")
	}
	
	// Test 2: Try CONFIRMED (should fail) 
	testSale.Status = models.SaleStatusConfirmed
	db.Save(testSale)
	_, err = ssotService.CreateSaleJournalEntry(testSale, 1)
	if err != nil {
		log.Printf("   ✅ CONFIRMED correctly rejected: %v", err)
	} else {
		log.Printf("   ❌ CONFIRMED incorrectly accepted!")
	}
	
	afterWorkflowBalances := getAccountBalances(db)
	if balancesEqual(beforeWorkflowBalances, afterWorkflowBalances) {
		log.Printf("   ✅ Workflow protection working - no unwanted balance changes")
	} else {
		log.Printf("   ❌ Workflow protection failed - balances changed!")
	}
}

func getAccountBalances(db *gorm.DB) map[string]float64 {
	targetCodes := []string{"4101", "2103", "1101", "1102", "1201"}
	balances := make(map[string]float64)
	
	var accounts []models.Account
	db.Where("code IN ?", targetCodes).Find(&accounts)
	
	for _, acc := range accounts {
		balances[acc.Code] = acc.Balance
	}
	
	return balances
}

func printBalances(balances map[string]float64) {
	for code, balance := range balances {
		log.Printf("   Account %s: Balance %.2f", code, balance)
	}
}

func balancesEqual(before, after map[string]float64) bool {
	for code, beforeBalance := range before {
		afterBalance, exists := after[code]
		if !exists || beforeBalance != afterBalance {
			return false
		}
	}
	return true
}

func printBalanceDifferences(before, after map[string]float64) {
	log.Printf("   💥 Balance differences detected:")
	for code, beforeBalance := range before {
		afterBalance := after[code]
		if beforeBalance != afterBalance {
			difference := afterBalance - beforeBalance
			log.Printf("     Account %s: %.2f → %.2f (change: %+.2f)", 
				code, beforeBalance, afterBalance, difference)
		}
	}
}

func checkUnwantedJournalEntries(db *gorm.DB) {
	// Check for recent SSOT entries that shouldn't exist
	var recentEntries []models.SSOTJournalEntry
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	
	db.Where("created_at > ? AND source_type = ?", fiveMinutesAgo, "SALES").Find(&recentEntries)
	
	if len(recentEntries) > 1 { // We created 1 legitimate entry in previous test
		log.Printf("⚠️  Found %d recent SALES journal entries (expected 1 or 0)", len(recentEntries))
		for _, entry := range recentEntries {
			log.Printf("     Entry %d: %s, Status: %s", entry.ID, entry.Description, entry.Status)
		}
	} else {
		log.Printf("✅ No unwanted recent journal entries found")
	}
}

func uintPtr(val uint) *uint {
	return &val
}