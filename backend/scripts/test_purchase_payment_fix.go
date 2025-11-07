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
	log.Println("🧪 Testing Purchase Payment Fix...")
	
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
	testPurchasePayment(paymentService, db)
}

func testPurchasePayment(paymentService *services.PaymentService, db *gorm.DB) {
	log.Println("\n📝 Testing Purchase Payment Creation...")
	
	// Get the first available purchase with outstanding amount
	var purchase models.Purchase
	if err := db.Where("status = ? AND payment_method = ?", "APPROVED", "CREDIT").First(&purchase).Error; err != nil {
		log.Printf("⚠️ No approved credit purchases found, creating test purchase")
		if err := createTestPurchase(db, &purchase); err != nil {
			log.Fatalf("❌ Failed to create test purchase: %v", err)
		}
	}
	
	log.Printf("📋 Using Purchase ID: %d, Total: %.2f", purchase.ID, purchase.TotalAmount)
	
	// Get the first available vendor
	var vendor models.Contact
	if err := db.Where("type = ?", "VENDOR").First(&vendor).Error; err != nil {
		log.Fatalf("❌ No vendors found: %v", err)
	}
	
	// Get the first available cash/bank account with sufficient balance
	var cashBank models.CashBank
	if err := db.Where("balance >= ?", 5000000).First(&cashBank).Error; err != nil {
		log.Printf("⚠️ No accounts with sufficient balance found, using any available account")
		if err := db.First(&cashBank).Error; err != nil {
			log.Fatalf("❌ No cash/bank accounts found: %v", err)
		}
	}
	
	log.Printf("💰 Using Cash Bank: %s (ID: %d), Current Balance: %.2f", 
		cashBank.Name, cashBank.ID, cashBank.Balance)
	
	// Record current balances for comparison
	originalCashBalance := cashBank.Balance
	
	// Create payment request
	paymentAmount := 1000000.00 // Test with 1M Rupiah
	if paymentAmount > cashBank.Balance {
		paymentAmount = cashBank.Balance * 0.1 // Use 10% of available balance
	}
	
	if paymentAmount > purchase.TotalAmount {
		paymentAmount = purchase.TotalAmount // Don't exceed total amount
	}
	
	paymentRequest := services.PaymentCreateRequest{
		ContactID:   vendor.ID,
		CashBankID:  cashBank.ID,
		Date:        time.Now(),
		Amount:      paymentAmount,
		Method:      "Bank Transfer", // This should be treated as payable payment
		Reference:   "TEST-PURCHASE-PAYMENT-FIX",
		Notes:       "Testing purchase payment accounting fix",
		BillAllocations: []services.BillAllocation{
			{
				BillID: purchase.ID,
				Amount: paymentAmount,
			},
		},
	}
	
	log.Printf("💳 Creating purchase payment: Amount=%.2f, Method=%s", paymentAmount, paymentRequest.Method)
	
	// Get valid user ID
	var user models.User
	if err := db.First(&user).Error; err != nil {
		log.Fatalf("❌ No users found: %v", err)
	}
	
	// Create payable payment
	payment, err := paymentService.CreatePayablePayment(paymentRequest, user.ID)
	if err != nil {
		log.Fatalf("❌ Failed to create payment: %v", err)
	}
	
	log.Printf("✅ Payment created: ID=%d, Code=%s", payment.ID, payment.Code)
	
	// Wait a moment for any async operations
	time.Sleep(2 * time.Second)
	
	// Verify the results
	verifyPurchasePaymentResults(db, cashBank.ID, purchase.ID, originalCashBalance, paymentAmount)
}

func createTestPurchase(db *gorm.DB, purchase *models.Purchase) error {
	// Get first vendor
	var vendor models.Contact
	if err := db.Where("type = ?", "VENDOR").First(&vendor).Error; err != nil {
		return fmt.Errorf("no vendors found: %v", err)
	}
	
	// Get first user
	var user models.User
	if err := db.First(&user).Error; err != nil {
		return fmt.Errorf("no users found: %v", err)
	}
	
	// Create test purchase
	*purchase = models.Purchase{
		Code:           "TEST-PURCHASE-001",
		VendorID:       vendor.ID,
		UserID:         user.ID,
		Date:           time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 30),
		TotalAmount:    5000000.00,
		PaidAmount:     0.00,
		OutstandingAmount: 5000000.00,
		Status:         models.PurchaseStatusApproved,
		PaymentMethod:  models.PurchasePaymentCredit,
		Notes:          "Test purchase for payment fix verification",
	}
	
	return db.Create(purchase).Error
}

func verifyPurchasePaymentResults(db *gorm.DB, cashBankID uint, purchaseID uint, originalCashBalance, paymentAmount float64) {
	log.Println("\n🔍 Verifying Purchase Payment Results...")
	
	// Check cash bank balance - should DECREASE for purchase payments
	var cashBank models.CashBank
	if err := db.First(&cashBank, cashBankID).Error; err != nil {
		log.Printf("❌ Failed to get cash bank: %v", err)
		return
	}
	
	expectedCashBalance := originalCashBalance - paymentAmount
	if cashBank.Balance == expectedCashBalance {
		log.Printf("✅ Cash Bank Balance CORRECT: %.2f -> %.2f (-%.2f)", 
			originalCashBalance, cashBank.Balance, paymentAmount)
	} else {
		log.Printf("❌ Cash Bank Balance INCORRECT: Expected %.2f, Got %.2f", 
			expectedCashBalance, cashBank.Balance)
	}
	
	// Check purchase paid amount
	var purchase models.Purchase
	if err := db.First(&purchase, purchaseID).Error; err != nil {
		log.Printf("❌ Failed to get purchase: %v", err)
		return
	}
	
	if purchase.PaidAmount == paymentAmount {
		log.Printf("✅ Purchase Paid Amount CORRECT: 0.00 -> %.2f (+%.2f)", 
			purchase.PaidAmount, paymentAmount)
	} else {
		log.Printf("❌ Purchase Paid Amount INCORRECT: Expected %.2f, Got %.2f", 
			paymentAmount, purchase.PaidAmount)
	}
	
	// Check if there's a journal entry
	var journalCount int64
	db.Model(&models.JournalEntry{}).Where("reference_type = ? AND description LIKE ?", 
		models.JournalRefPayment, "%Vendor Payment%").Count(&journalCount)
	
	log.Printf("📝 Vendor Payment Journal Entries Found: %d", journalCount)
	
	// Check specific journal lines for the accounts
	var journalLines []models.JournalLine
	db.Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id").
		Where("journal_entries.reference_type = ? AND journal_entries.description LIKE ?", 
			models.JournalRefPayment, "%Vendor Payment%").
		Order("journal_lines.created_at DESC").
		Limit(10).
		Find(&journalLines)
	
	log.Printf("📋 Recent Vendor Payment Journal Lines: %d", len(journalLines))
	for _, line := range journalLines {
		if line.DebitAmount > 0 {
			log.Printf("  DR Account %d: %.2f (%s)", line.AccountID, line.DebitAmount, line.Description)
		} else {
			log.Printf("  CR Account %d: %.2f (%s)", line.AccountID, line.CreditAmount, line.Description)
		}
	}
	
	log.Println("\n🎉 Purchase payment verification completed!")
	log.Println("Expected behavior: Cash/Bank balance DECREASES (we're paying money out)")
}