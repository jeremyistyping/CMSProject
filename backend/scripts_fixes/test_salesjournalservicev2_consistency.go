package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Printf("🔍 Testing SalesJournalServiceV2 Consistency Fixes...")

	// Database connection
	dsn := "host=localhost user=postgres password=your_password dbname=accounting_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test 1: StatusValidationHelper
	log.Printf("\n📋 Test 1: StatusValidationHelper")
	testStatusValidation()

	// Test 2: Deprecated Method Guards
	log.Printf("\n🔒 Test 2: Deprecated Method Guards")
	testDeprecatedMethodGuards(db)

	// Test 3: Payment Status Validation
	log.Printf("\n💰 Test 3: Payment Status Validation")
	testPaymentStatusValidation(db)

	log.Printf("\n✅ All tests completed!")
}

func testStatusValidation() {
	validator := services.NewStatusValidationHelper()

	// Test allowed statuses
	allowedStatuses := []string{"INVOICED", "PAID", "OVERDUE"}
	for _, status := range allowedStatuses {
		if !validator.ShouldAllowPayment(status) {
			log.Printf("❌ FAIL: Status '%s' should be allowed", status)
		} else {
			log.Printf("✅ PASS: Status '%s' is correctly allowed", status)
		}
	}

	// Test blocked statuses
	blockedStatuses := []string{"DRAFT", "CONFIRMED", "CANCELLED"}
	for _, status := range blockedStatuses {
		if validator.ShouldAllowPayment(status) {
			log.Printf("❌ FAIL: Status '%s' should be blocked", status)
		} else {
			log.Printf("✅ PASS: Status '%s' is correctly blocked", status)
		}
	}

	// Test validation error messages
	err := validator.ValidatePaymentAllocation("DRAFT", 123)
	if err == nil {
		log.Printf("❌ FAIL: Should return error for DRAFT status")
	} else {
		log.Printf("✅ PASS: Correct error for DRAFT status: %v", err)
	}
}

func testDeprecatedMethodGuards(db *gorm.DB) {
	validator := services.NewStatusValidationHelper()

	// Test with deprecated methods disabled
	os.Setenv("DISABLE_DEPRECATED_PAYMENT_METHODS", "true")
	
	err := validator.ValidateDeprecatedMethodUsage("createReceivablePaymentJournalWithSSOT")
	if err == nil {
		log.Printf("❌ FAIL: Should return error when deprecated methods disabled")
	} else {
		log.Printf("✅ PASS: Deprecated method correctly blocked: %v", err)
	}

	// Test with deprecated methods enabled
	os.Setenv("DISABLE_DEPRECATED_PAYMENT_METHODS", "false")
	
	err = validator.ValidateDeprecatedMethodUsage("createReceivablePaymentJournalWithSSOT")
	if err != nil {
		log.Printf("❌ FAIL: Should allow deprecated method when not disabled")
	} else {
		log.Printf("✅ PASS: Deprecated method correctly allowed when not disabled")
	}

	// Reset environment
	os.Unsetenv("DISABLE_DEPRECATED_PAYMENT_METHODS")
}

func testPaymentStatusValidation(db *gorm.DB) {
	// Create test sales with different statuses
	testSales := []models.Sale{
		{
			ID: 1001,
			Status: "DRAFT",
			InvoiceNumber: "INV-DRAFT-001",
			CustomerID: 1,
			TotalAmount: 100.0,
			CreatedAt: time.Now(),
		},
		{
			ID: 1002,
			Status: "INVOICED", 
			InvoiceNumber: "INV-INVOICED-001",
			CustomerID: 1,
			TotalAmount: 200.0,
			CreatedAt: time.Now(),
		},
		{
			ID: 1003,
			Status: "OVERDUE",
			InvoiceNumber: "INV-OVERDUE-001", 
			CustomerID: 1,
			TotalAmount: 300.0,
			CreatedAt: time.Now(),
		},
	}

	validator := services.NewStatusValidationHelper()

	for _, sale := range testSales {
		err := validator.ValidatePaymentAllocation(sale.Status, sale.ID)
		
		if sale.Status == "DRAFT" && err == nil {
			log.Printf("❌ FAIL: Should block payment for DRAFT status")
		} else if sale.Status == "DRAFT" && err != nil {
			log.Printf("✅ PASS: Correctly blocked payment for DRAFT status")
		} else if (sale.Status == "INVOICED" || sale.Status == "OVERDUE") && err != nil {
			log.Printf("❌ FAIL: Should allow payment for %s status", sale.Status)
		} else if (sale.Status == "INVOICED" || sale.Status == "OVERDUE") && err == nil {
			log.Printf("✅ PASS: Correctly allowed payment for %s status", sale.Status)
		}
	}
}

// Test helper function
func simulatePaymentAttempt(status string, saleID uint) error {
	validator := services.NewStatusValidationHelper()
	return validator.ValidatePaymentAllocation(status, saleID)
}

// Integration test example
func testIntegrationScenario() {
	log.Printf("\n🧪 Integration Test: Real-world scenario")
	
	scenarios := []struct {
		name string
		status string
		expectedBlocked bool
	}{
		{"Draft Invoice Payment", "DRAFT", true},
		{"Confirmed Invoice Payment", "CONFIRMED", true},
		{"Invoiced Invoice Payment", "INVOICED", false},
		{"Overdue Invoice Payment", "OVERDUE", false},
		{"Paid Invoice Payment", "PAID", false},
	}

	for _, scenario := range scenarios {
		err := simulatePaymentAttempt(scenario.status, 999)
		blocked := (err != nil)
		
		if blocked == scenario.expectedBlocked {
			log.Printf("✅ PASS: %s - correctly %s", scenario.name, 
				map[bool]string{true: "blocked", false: "allowed"}[blocked])
		} else {
			log.Printf("❌ FAIL: %s - should be %s but was %s", scenario.name,
				map[bool]string{true: "blocked", false: "allowed"}[scenario.expectedBlocked],
				map[bool]string{true: "blocked", false: "allowed"}[blocked])
		}
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println(`
🔧 SalesJournalServiceV2 Consistency Test
==========================================
Testing all implemented fixes:
1. StatusValidationHelper consistency
2. Deprecated method guards  
3. Payment allocation validation
4. Environment variable controls
==========================================
`)
}