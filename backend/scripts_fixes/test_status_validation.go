package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
)

func main() {
	log.Printf("🧪 Testing Status-Based Validation in Double Entry Service")
	log.Printf("=========================================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Initialize double entry service
	journalRepo := repositories.NewJournalEntryRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	doubleEntryService := services.NewSalesDoubleEntryService(db, journalRepo, accountRepo)

	// Create a test sale with different statuses
	testSale := &models.Sale{
		ID:          999, // Fake ID for testing
		Code:        "TEST-VALIDATION-001",
		TotalAmount: 1000000.0,
		PPNAmount:   100000.0,
		SaleItems: []models.SaleItem{
			{
				ID:               1,
				RevenueAccountID: 1, // Exists
				LineTotal:        900000.0,
				PPNAmount:        100000.0,
			},
		},
	}

	// Test cases
	testCases := []struct {
		status      string
		shouldPass  bool
		description string
	}{
		{models.SaleStatusDraft, false, "DRAFT status should be REJECTED"},
		{models.SaleStatusConfirmed, false, "CONFIRMED status should be REJECTED"},
		{models.SaleStatusInvoiced, true, "INVOICED status should be ACCEPTED"},
		{models.SaleStatusPaid, true, "PAID status should be ACCEPTED"},
		{models.SaleStatusCancelled, false, "CANCELLED status should be REJECTED"},
	}

	log.Printf("\n🔄 Running validation tests...")
	
	allTestsPassed := true
	
	for i, tc := range testCases {
		testSale.Status = tc.status
		
		log.Printf("\nTest %d: %s", i+1, tc.description)
		log.Printf("  Status: %s", tc.status)
		
		err := doubleEntryService.CreateSaleJournalEntries(testSale, 1)
		
		if tc.shouldPass {
			if err == nil {
				log.Printf("  ✅ PASS: Status %s correctly allowed", tc.status)
			} else {
				log.Printf("  ❌ FAIL: Status %s should be allowed but got error: %v", tc.status, err)
				allTestsPassed = false
			}
		} else {
			if err != nil {
				log.Printf("  ✅ PASS: Status %s correctly rejected - %v", tc.status, err)
			} else {
				log.Printf("  ❌ FAIL: Status %s should be rejected but was allowed", tc.status)
				allTestsPassed = false
			}
		}
	}

	log.Printf("\n🎯 VALIDATION TEST RESULTS:")
	log.Printf("============================")
	
	if allTestsPassed {
		log.Printf("🎉 ALL TESTS PASSED!")
		log.Printf("✅ Only INVOICED and PAID sales can create journal entries")
		log.Printf("✅ DRAFT and CONFIRMED sales are correctly rejected")
		log.Printf("✅ Status-based posting validation is working correctly!")
	} else {
		log.Printf("⚠️ Some validation tests failed")
		log.Printf("Please review the validateSaleForJournalEntry function")
	}

	// Test the actual behavior with a real sale from database
	log.Printf("\n🔍 Testing with real sales from database...")
	
	var realSales []models.Sale
	db.Preload("SaleItems").Limit(3).Find(&realSales)
	
	for i, sale := range realSales {
		log.Printf("\nReal Sale %d:", i+1)
		log.Printf("  ID: %d, Code: %s, Status: %s", sale.ID, sale.Code, sale.Status)
		
		err := doubleEntryService.CreateSaleJournalEntries(&sale, 1)
		
		if sale.Status == models.SaleStatusInvoiced || sale.Status == models.SaleStatusPaid {
			if err == nil {
				log.Printf("  ✅ Status %s: Journal creation would succeed", sale.Status)
			} else {
				log.Printf("  ⚠️ Status %s: Got error (might be business rule): %v", sale.Status, err)
			}
		} else {
			if err != nil {
				log.Printf("  ✅ Status %s: Correctly rejected - %v", sale.Status, err)
			} else {
				log.Printf("  ❌ Status %s: Should be rejected but was allowed!", sale.Status)
			}
		}
	}
}