package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

func main() {
	log.Println("🧪 Testing Sales Payment Fix...")
	
	// Initialize database connection
	db := database.ConnectDB()
	
	log.Println("✅ Database initialized successfully")
	
	// Initialize repositories and services
	paymentRepo := repositories.NewPaymentRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	purchaseRepo := repositories.NewPurchaseRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	
	paymentService := services.NewPaymentService(
		db, 
		paymentRepo, 
		salesRepo, 
		purchaseRepo, 
		cashBankRepo, 
		accountRepo, 
		contactRepo,
	)
	
	log.Println("✅ Services initialized successfully")
	
	// Test payment creation
	testSalesPayment(paymentService, db)
}

func testSalesPayment(paymentService *services.PaymentService, db *gorm.DB) {
	log.Println("\n📝 Testing Sales Payment Creation...")
	
	// Get the first available sale with outstanding amount
	var sale models.Sale
	if err := db.Where("outstanding_amount > 0").First(&sale).Error; err != nil {
		log.Printf("⚠️ No sales with outstanding amount found, creating test sale")
		if err := createTestSale(db, &sale); err != nil {
			log.Fatalf("❌ Failed to create test sale: %v", err)
		}
	}
	
	log.Printf("📋 Using Sale ID: %d, Outstanding: %.2f", sale.ID, sale.OutstandingAmount)
	
	// Get the first available customer
	var customer models.Contact
	if err := db.Where("type = ?", "CUSTOMER").First(&customer).Error; err != nil {
		log.Fatalf("❌ No customers found: %v", err)
	}
	
	// Get the first available cash/bank account
	var cashBank models.CashBank
	if err := db.First(&cashBank).Error; err != nil {
		log.Fatalf("❌ No cash/bank accounts found: %v", err)
	}
	
	log.Printf("💰 Using Cash Bank: %s (ID: %d), Current Balance: %.2f", 
		cashBank.Name, cashBank.ID, cashBank.Balance)
	
	// Record current balances for comparison
	originalCashBalance := cashBank.Balance
	originalSaleOutstanding := sale.OutstandingAmount
	
	// Create payment request
	paymentAmount := 1000.00 // Test with fixed amount
	if paymentAmount > sale.OutstandingAmount {
		paymentAmount = sale.OutstandingAmount // Don't exceed outstanding
	}
	
	paymentRequest := services.PaymentCreateRequest{
		ContactID:   customer.ID,
		CashBankID:  cashBank.ID,
		Date:        time.Now(),
		Amount:      paymentAmount,
		Method:      "Bank Transfer", // This was the issue - should be treated as receivable
		Reference:   "TEST-PAYMENT-FIX",
		Notes:       "Testing sales payment accounting fix",
		Allocations: []services.InvoiceAllocation{
			{
				InvoiceID: sale.ID,
				Amount:    paymentAmount,
			},
		},
	}
	
	log.Printf("💳 Creating payment: Amount=%.2f, Method=%s", paymentAmount, paymentRequest.Method)
	
	// Get valid user ID
	var user models.User
	if err := db.First(&user).Error; err != nil {
		log.Fatalf("❌ No users found: %v", err)
	}
	
	// Create receivable payment
	payment, err := paymentService.CreateReceivablePayment(paymentRequest, user.ID)
	if err != nil {
		log.Fatalf("❌ Failed to create payment: %v", err)
	}
	
	log.Printf("✅ Payment created: ID=%d, Code=%s", payment.ID, payment.Code)
	
	// Wait a moment for async journal creation
	time.Sleep(2 * time.Second)
	
	// Verify the results
	verifyPaymentResults(db, cashBank.ID, sale.ID, originalCashBalance, originalSaleOutstanding, paymentAmount)
}

func createTestSale(db *gorm.DB, sale *models.Sale) error {
	// Get first customer
	var customer models.Contact
	if err := db.Where("type = ?", "CUSTOMER").First(&customer).Error; err != nil {
		return fmt.Errorf("no customers found: %v", err)
	}
	
	// Create test sale
	*sale = models.Sale{
		Code:              "TEST-SALE-001",
		CustomerID:        customer.ID,
		Date:              time.Now(),
		DueDate:           time.Now().AddDate(0, 0, 30),
		TotalAmount:       5000.00,
		PaidAmount:        0.00,
		OutstandingAmount: 5000.00,
		Status:            models.SaleStatusInvoiced,
		Notes:             "Test sale for payment fix verification",
	}
	
	return db.Create(sale).Error
}

func verifyPaymentResults(db *gorm.DB, cashBankID uint, saleID uint, originalCashBalance, originalSaleOutstanding, paymentAmount float64) {
	log.Println("\n🔍 Verifying Payment Results...")
	
	// Check cash bank balance
	var cashBank models.CashBank
	if err := db.First(&cashBank, cashBankID).Error; err != nil {
		log.Printf("❌ Failed to get cash bank: %v", err)
		return
	}
	
	expectedCashBalance := originalCashBalance + paymentAmount
	if cashBank.Balance == expectedCashBalance {
		log.Printf("✅ Cash Bank Balance CORRECT: %.2f -> %.2f (+%.2f)", 
			originalCashBalance, cashBank.Balance, paymentAmount)
	} else {
		log.Printf("❌ Cash Bank Balance INCORRECT: Expected %.2f, Got %.2f", 
			expectedCashBalance, cashBank.Balance)
	}
	
	// Check sale outstanding amount
	var sale models.Sale
	if err := db.First(&sale, saleID).Error; err != nil {
		log.Printf("❌ Failed to get sale: %v", err)
		return
	}
	
	expectedOutstanding := originalSaleOutstanding - paymentAmount
	if sale.OutstandingAmount == expectedOutstanding {
		log.Printf("✅ Sale Outstanding CORRECT: %.2f -> %.2f (-%.2f)", 
			originalSaleOutstanding, sale.OutstandingAmount, paymentAmount)
	} else {
		log.Printf("❌ Sale Outstanding INCORRECT: Expected %.2f, Got %.2f", 
			expectedOutstanding, sale.OutstandingAmount)
	}
	
	// Check if there's a journal entry
	var journalCount int64
	db.Model(&models.JournalEntry{}).Where("reference_type = ? AND description LIKE ?", 
		models.JournalRefPayment, "%Payment%").Count(&journalCount)
	
	log.Printf("📝 Journal Entries Found: %d", journalCount)
	
	// Check specific journal lines for the accounts
	var journalLines []models.JournalLine
	db.Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_entries.reference_type = ? AND journal_entries.description LIKE ?", 
			models.JournalRefPayment, "%Payment%").
		Order("journal_lines.created_at DESC").
		Limit(10).
		Find(&journalLines)
	
	log.Printf("📋 Recent Payment Journal Lines: %d", len(journalLines))
	for _, line := range journalLines {
		if line.DebitAmount > 0 {
			log.Printf("  DR Account %d: %.2f (%s)", line.AccountID, line.DebitAmount, line.Description)
		} else {
			log.Printf("  CR Account %d: %.2f (%s)", line.AccountID, line.CreditAmount, line.Description)
		}
	}
	
	log.Println("\n🎉 Payment verification completed!")
}