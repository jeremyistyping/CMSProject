package main

import (
	"encoding/json"
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
	log.Printf("🔬 Starting Payment Bug Test")
	
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// Initialize repositories and services
	accountRepo := repositories.NewAccountRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	
	paymentService := services.NewPaymentService(db, paymentRepo, accountRepo)
	salesPaymentService := services.NewSalesPaymentService(db, salesRepo, paymentService)
	
	// Test scenario: 1 million sale with 50% payment
	err = testPaymentBug(db, salesRepo, salesPaymentService, paymentService)
	if err != nil {
		log.Fatalf("❌ Test failed: %v", err)
	}
	
	log.Printf("✅ Test completed")
}

func testPaymentBug(db *gorm.DB, salesRepo *repositories.SalesRepository, salesPaymentService *services.SalesPaymentService, paymentService *services.PaymentService) error {
	userID := uint(1)
	
	// Step 1: Create a test sale of 1 million
	log.Printf("📊 Step 1: Creating test sale of 1,000,000")
	
	// Find or create test customer
	var customer models.Customer
	if err := db.Where("email = ?", "test@example.com").First(&customer).Error; err != nil {
		customer = models.Customer{
			Code:  "TEST001",
			Name:  "Test Customer",
			Email: "test@example.com",
			Type:  "INDIVIDUAL",
		}
		if err := db.Create(&customer).Error; err != nil {
			return fmt.Errorf("failed to create test customer: %v", err)
		}
	}
	
	// Create test sale
	sale := &models.Sale{
		Code:         fmt.Sprintf("SALE-TEST-%d", time.Now().Unix()),
		Date:         time.Now(),
		CustomerID:   customer.ID,
		TotalAmount:  1000000.0, // 1 million
		PaidAmount:   0,
		OutstandingAmount: 1000000.0,
		Status:       models.SaleStatusInvoiced,
		InvoiceNumber: fmt.Sprintf("INV-TEST-%d", time.Now().Unix()),
		DueDate:      time.Now().AddDate(0, 0, 30),
		UserID:       userID,
	}
	
	if err := db.Create(sale).Error; err != nil {
		return fmt.Errorf("failed to create test sale: %v", err)
	}
	
	log.Printf("✅ Test sale created: ID=%d, Amount=%.2f", sale.ID, sale.TotalAmount)
	
	// Step 2: Get bank balances before payment
	log.Printf("📊 Step 2: Recording balances before payment")
	
	bankAccounts := []string{"1104", "1102", "1101"} // Mandiri, BCA, Cash
	beforeBalances := make(map[string]float64)
	
	for _, code := range bankAccounts {
		var account models.Account
		if err := db.Where("code = ?", code).First(&account).Error; err != nil {
			log.Printf("⚠️ Account %s not found", code)
			continue
		}
		
		// Get balance from cash_banks table if exists
		var cashBank models.CashBank
		if err := db.Where("account_id = ?", account.ID).First(&cashBank).Error; err == nil {
			beforeBalances[code] = cashBank.Balance
			log.Printf("📈 %s (%s) balance before: %.2f", account.Name, code, cashBank.Balance)
		}
	}
	
	// Get AR balance before
	var arAccount models.Account
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err == nil {
		beforeBalances["1201"] = arAccount.Balance
		log.Printf("📈 AR (1201) balance before: %.2f", arAccount.Balance)
	}
	
	// Step 3: Make 50% payment (500,000)
	paymentAmount := 500000.0
	log.Printf("💰 Step 3: Making payment of %.2f (50%% of %.2f)", paymentAmount, sale.TotalAmount)
	
	// Use Bank Mandiri (account code 1104) for payment
	var bankMandiri models.CashBank
	if err := db.Joins("JOIN accounts ON cash_banks.account_id = accounts.id").
		Where("accounts.code = ?", "1104").First(&bankMandiri).Error; err != nil {
		return fmt.Errorf("Bank Mandiri not found: %v", err)
	}
	
	paymentRequest := models.SalePaymentRequest{
		Amount:        paymentAmount,
		PaymentDate:   time.Now(),
		PaymentMethod: "BANK_TRANSFER",
		Reference:     "TEST-PAY-50PCT",
		Notes:         "Test payment 50%",
		CashBankID:    &bankMandiri.ID,
	}
	
	// Create payment using SalesPaymentService (the one suspected of having bug)
	payment, err := salesPaymentService.CreateSalePaymentWithLock(sale.ID, paymentRequest, userID)
	if err != nil {
		return fmt.Errorf("failed to create payment: %v", err)
	}
	
	log.Printf("✅ Payment created: ID=%d, Amount=%.2f", payment.ID, payment.Amount)
	
	// Step 4: Check balances after payment
	log.Printf("📊 Step 4: Recording balances after payment")
	
	afterBalances := make(map[string]float64)
	
	for _, code := range bankAccounts {
		var account models.Account
		if err := db.Where("code = ?", code).First(&account).Error; err != nil {
			continue
		}
		
		// Get balance from cash_banks table if exists
		var cashBank models.CashBank
		if err := db.Where("account_id = ?", account.ID).First(&cashBank).Error; err == nil {
			afterBalances[code] = cashBank.Balance
			log.Printf("📈 %s (%s) balance after: %.2f", account.Name, code, cashBank.Balance)
		}
	}
	
	// Get AR balance after
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err == nil {
		afterBalances["1201"] = arAccount.Balance
		log.Printf("📈 AR (1201) balance after: %.2f", arAccount.Balance)
	}
	
	// Step 5: Analyze the results
	log.Printf("🔍 Step 5: Analyzing balance changes")
	
	for code, beforeBalance := range beforeBalances {
		afterBalance := afterBalances[code]
		change := afterBalance - beforeBalance
		
		log.Printf("📊 %s: %.2f -> %.2f (change: %.2f)", code, beforeBalance, afterBalance, change)
		
		if code == "1104" { // Bank Mandiri - should increase by payment amount
			if change != paymentAmount {
				log.Printf("❌ BUG DETECTED! Bank Mandiri should increase by %.2f but increased by %.2f", paymentAmount, change)
				if change == sale.TotalAmount {
					log.Printf("🚨 CONFIRMED BUG: Bank increased by total sale amount instead of payment amount!")
				}
				return fmt.Errorf("payment amount recording bug detected")
			} else {
				log.Printf("✅ Bank Mandiri balance change is correct")
			}
		} else if code == "1201" { // AR - should decrease by payment amount
			expectedChange := -paymentAmount
			if change != expectedChange {
				log.Printf("❌ BUG DETECTED! AR should decrease by %.2f but changed by %.2f", paymentAmount, change)
				return fmt.Errorf("AR balance recording bug detected")
			} else {
				log.Printf("✅ AR balance change is correct")
			}
		}
	}
	
	// Step 6: Check journal entries
	log.Printf("📖 Step 6: Checking journal entries")
	
	var journalEntries []models.JournalEntry
	if err := db.Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).
		Preload("JournalLines").Find(&journalEntries).Error; err != nil {
		log.Printf("⚠️ Could not fetch journal entries: %v", err)
	} else {
		log.Printf("📖 Found %d journal entries for payment", len(journalEntries))
		
		for _, entry := range journalEntries {
			log.Printf("📖 Entry %d: %s (Total Debit: %.2f, Total Credit: %.2f)", 
				entry.ID, entry.Description, entry.TotalDebit, entry.TotalCredit)
			
			for _, line := range entry.JournalLines {
				log.Printf("  📝 Line: Account %d, Debit: %.2f, Credit: %.2f - %s", 
					line.AccountID, line.DebitAmount, line.CreditAmount, line.Description)
			}
		}
	}
	
	// Cleanup - remove test data
	log.Printf("🧹 Cleanup: Removing test data")
	db.Where("id = ?", payment.ID).Delete(&models.SalePayment{})
	db.Where("id = ?", sale.ID).Delete(&models.Sale{})
	
	log.Printf("✅ Test completed successfully - no bug detected in this test")
	return nil
}