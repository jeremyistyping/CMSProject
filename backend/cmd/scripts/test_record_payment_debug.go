package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"gorm.io/gorm"
)

func main() {
	log.Println("Testing Record Payment Debug...")

	// Initialize database
	db := database.ConnectDB()

	testRecordPayment(db)
}

func testRecordPayment(db *gorm.DB) {
	log.Println("\n=== Testing Record Payment ===")

	// Initialize repositories and services
	paymentRepo := repositories.NewPaymentRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	accountRepo := repositories.NewAccountRepository(db)

	// Create payment service
	paymentService := services.NewPaymentService(db, paymentRepo, cashBankRepo, contactRepo, accountRepo)

	// Test 1: Check required accounts
	checkPaymentAccounts(db)

	// Test 2: Check cash/bank accounts
	checkCashBankAccounts(db)

	// Test 3: Find a test sale to create payment for
	saleID := findTestSale(db)
	if saleID == 0 {
		log.Println("❌ No suitable sale found for testing")
		return
	}

	// Test 4: Try to create a payment
	testCreatePayment(paymentService, saleID)
}

func checkPaymentAccounts(db *gorm.DB) {
	log.Println("\n--- Checking Payment-Related Accounts ---")

	accountRepo := repositories.NewAccountRepository(db)
	requiredAccounts := map[string]string{
		"1101": "Kas (Cash)",
		"1201": "Piutang Usaha (Accounts Receivable)",
		"1200": "Accounts Receivable Alternative",
	}

	missingAccounts := []string{}
	for code, description := range requiredAccounts {
		account, err := accountRepo.GetAccountByCode(code)
		if err != nil {
			fmt.Printf("❌ MISSING: Account %s - %s: %v\n", code, description, err)
			missingAccounts = append(missingAccounts, code)
		} else {
			fmt.Printf("✅ FOUND: Account %s - %s (ID: %d, Name: %s)\n", code, description, account.ID, account.Name)
		}
	}

	if len(missingAccounts) > 0 {
		fmt.Printf("\n❌ Missing %d payment accounts: %v\n", len(missingAccounts), missingAccounts)
	} else {
		fmt.Println("✅ All payment accounts are present")
	}
}

func checkCashBankAccounts(db *gorm.DB) {
	log.Println("\n--- Checking Cash/Bank Accounts ---")

	var cashBanks []models.CashBank
	if err := db.Find(&cashBanks).Error; err != nil {
		log.Printf("❌ Error finding cash/bank accounts: %v", err)
		return
	}

	if len(cashBanks) == 0 {
		log.Println("❌ No cash/bank accounts found - this will cause payment failures!")
		return
	}

	log.Printf("Found %d cash/bank accounts:", len(cashBanks))
	for _, cb := range cashBanks {
		fmt.Printf("  - ID: %d, Name: %s, Type: %s, Balance: %.2f, Account ID: %d\n", 
			cb.ID, cb.Name, cb.Type, cb.Balance, cb.AccountID)
	}
}

func findTestSale(db *gorm.DB) uint {
	log.Println("\n--- Finding Test Sale ---")

	var sale models.Sale
	// Find a sale that is INVOICED and has outstanding amount
	if err := db.Where("status = ? AND outstanding_amount > 0", "INVOICED").First(&sale).Error; err != nil {
		log.Printf("❌ No suitable INVOICED sale found: %v", err)
		return 0
	}

	log.Printf("✅ Found test sale: ID=%d, Code=%s, Status=%s, Outstanding=%.2f", 
		sale.ID, sale.Code, sale.Status, sale.OutstandingAmount)
	return sale.ID
}

func testCreatePayment(paymentService *services.PaymentService, saleID uint) {
	log.Printf("\n--- Testing Payment Creation for Sale %d ---", saleID)

	// Get sale details
	var sale models.Sale
	if err := paymentService.DB().Preload("Customer").First(&sale, saleID).Error; err != nil {
		log.Printf("❌ Error getting sale: %v", err)
		return
	}

	// Get a cash bank account
	var cashBank models.CashBank
	if err := paymentService.DB().First(&cashBank).Error; err != nil {
		log.Printf("❌ No cash bank account available: %v", err)
		return
	}

	// Create payment request
	paymentRequest := services.PaymentCreateRequest{
		ContactID:  sale.CustomerID,
		CashBankID: cashBank.ID,
		Date:       time.Now(),
		Amount:     1000.0, // Test with small amount
		Method:     "CASH",
		Reference:  "TEST001",
		Notes:      "Test payment for debugging",
		Allocations: []services.InvoiceAllocation{
			{
				InvoiceID: saleID,
				Amount:    1000.0,
			},
		},
	}

	log.Printf("Creating test payment: Amount=%.2f, Customer=%d, CashBank=%d", 
		paymentRequest.Amount, paymentRequest.ContactID, paymentRequest.CashBankID)

	// Try to create payment
	userID := uint(1) // Admin user
	payment, err := paymentService.CreateReceivablePayment(paymentRequest, userID)
	if err != nil {
		log.Printf("❌ Payment creation failed: %v", err)
		
		// Analyze error
		errorStr := err.Error()
		if contains(errorStr, "account not found") {
			log.Println("💡 Missing accounting account issue")
		} else if contains(errorStr, "cash/bank account not found") {
			log.Println("💡 Cash/bank account issue")
		} else if contains(errorStr, "insufficient balance") {
			log.Println("💡 Insufficient cash/bank balance")
		} else if contains(errorStr, "not found") {
			log.Println("💡 Record not found issue")
		} else {
			log.Printf("💡 Other error: %s", errorStr)
		}
	} else {
		log.Printf("✅ Payment created successfully: ID=%d, Code=%s, Amount=%.2f", 
			payment.ID, payment.Code, payment.Amount)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}