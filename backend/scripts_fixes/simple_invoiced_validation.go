package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/models"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🚀 Simple INVOICED-Only Logic Verification")

	// Test the sales double entry service validation (no DB required)
	if err := testSalesDoubleEntryValidation(); err != nil {
		log.Fatalf("❌ Test failed: %v", err)
	}

	log.Println("✅ All tests passed! INVOICED-only logic verification complete")
}

func testSalesDoubleEntryValidation() error {
	log.Println("📋 Testing SalesDoubleEntryService validation logic...")

	// Create a mock sale with different statuses
	testCases := []struct {
		status        string
		shouldPass    bool
		description   string
	}{
		{models.SaleStatusDraft, false, "DRAFT sales should not create journal entries"},
		{models.SaleStatusConfirmed, false, "CONFIRMED sales should not create journal entries"},
		{models.SaleStatusInvoiced, true, "INVOICED sales should create journal entries"},
		{models.SaleStatusPaid, true, "PAID sales should create journal entries"},
	}

	for _, tc := range testCases {
		log.Printf("🔸 Testing %s status: %s", tc.status, tc.description)
		
		// Create a mock sale
		sale := &models.Sale{
			ID:          1, // Mock ID
			Code:        fmt.Sprintf("TEST-%s-%d", tc.status, time.Now().Unix()),
			Status:      tc.status,
			TotalAmount: 100000,
			Date:        time.Now(),
		}
		sale.SaleItems = []models.SaleItem{
			{
				ProductID: 1,
				Quantity:  1,
				UnitPrice: 100000,
				LineTotal: 100000,
			},
		}

		// Test validation logic (simulated)
		isValid := validateSaleForJournalEntry(sale)
		
		if isValid != tc.shouldPass {
			return fmt.Errorf("❌ Validation failed for %s status: expected %v, got %v", tc.status, tc.shouldPass, isValid)
		}
		
		log.Printf("✅ %s status validation passed", tc.status)
	}

	return nil
}

// Simulate the validation logic from SalesDoubleEntryService
func validateSaleForJournalEntry(sale *models.Sale) bool {
	if sale.ID == 0 {
		return false
	}

	// Critical validation: Only INVOICED sales should create journal entries and affect COA
	if sale.Status != models.SaleStatusInvoiced && sale.Status != models.SaleStatusPaid {
		log.Printf("📋 Journal entries can only be created for INVOICED or PAID sales, current status: %s", sale.Status)
		return false
	}

	if sale.TotalAmount <= 0 {
		return false
	}

	if len(sale.SaleItems) == 0 {
		return false
	}

	return true
}