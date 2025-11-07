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
	log.Println("🧪 Testing Payment Journal Entry Fix")
	log.Println("=====================================")

	// Initialize database connection
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Test scenario: Record payment sebesar Rp 1.887.000
	testPaymentJournalFix(db)
}

func testPaymentJournalFix(db *gorm.DB) {
	log.Println("🔍 Starting Payment Journal Entry Fix Test")

	// Initialize repositories and services
	salesRepo := repositories.NewSalesRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	contactRepo := repositories.NewContactRepository(db)

	paymentService := services.NewPaymentService(
		db, paymentRepo, salesRepo, nil, cashBankRepo, accountRepo, contactRepo,
	)

	// Get test data - sesuai dengan screenshot Anda
	testAmount := float64(1887000) // Rp 1.887.000
	testCustomerID := uint(1)      // PT Global Tech
	testCashBankID := uint(2)      // Bank BCA (code 1102)
	testSaleID := uint(1)          // Invoice yang akan dibayar

	log.Printf("📋 Test Parameters:")
	log.Printf("   Amount: Rp %.2f", testAmount)
	log.Printf("   Customer ID: %d", testCustomerID)
	log.Printf("   Cash Bank ID: %d", testCashBankID)
	log.Printf("   Sale ID: %d", testSaleID)

	// 1. Check initial Bank BCA balance
	initialBankBCA, err := checkBankBCABalance(db)
	if err != nil {
		log.Printf("❌ Error checking initial Bank BCA balance: %v", err)
		return
	}
	log.Printf("💰 Initial Bank BCA Balance: Rp %.2f", initialBankBCA)

	// 2. Check initial Piutang Usaha balance
	initialAR, err := checkAccountsReceivableBalance(db)
	if err != nil {
		log.Printf("❌ Error checking initial AR balance: %v", err)
		return
	}
	log.Printf("💰 Initial Piutang Usaha Balance: Rp %.2f", initialAR)

	// 3. Create payment request
	paymentRequest := services.PaymentCreateRequest{
		ContactID:  testCustomerID,
		CashBankID: testCashBankID,
		Date:       time.Now(),
		Amount:     testAmount,
		Method:     "BANK_TRANSFER",
		Reference:  "TEST-PAYMENT-FIX",
		Notes:      "Testing payment journal entry fix",
		Allocations: []services.InvoiceAllocation{
			{
				InvoiceID: testSaleID,
				Amount:    testAmount,
			},
		},
	}

	log.Println("🚀 Creating payment via PaymentService.CreateReceivablePayment...")
	
	// 4. Create payment (this should only create ONE journal entry)
	payment, err := paymentService.CreateReceivablePayment(paymentRequest, 1)
	if err != nil {
		log.Printf("❌ Error creating payment: %v", err)
		return
	}

	log.Printf("✅ Payment created successfully:")
	log.Printf("   Payment ID: %d", payment.ID)
	log.Printf("   Payment Code: %s", payment.Code)
	log.Printf("   Amount: Rp %.2f", payment.Amount)

	// 5. Wait a moment for all transactions to complete
	time.Sleep(2 * time.Second)

	// 6. Check final Bank BCA balance
	finalBankBCA, err := checkBankBCABalance(db)
	if err != nil {
		log.Printf("❌ Error checking final Bank BCA balance: %v", err)
		return
	}
	log.Printf("💰 Final Bank BCA Balance: Rp %.2f", finalBankBCA)

	// 7. Check final Piutang Usaha balance
	finalAR, err := checkAccountsReceivableBalance(db)
	if err != nil {
		log.Printf("❌ Error checking final AR balance: %v", err)
		return
	}
	log.Printf("💰 Final Piutang Usaha Balance: Rp %.2f", finalAR)

	// 8. Calculate changes
	bankBCAChange := finalBankBCA - initialBankBCA
	arChange := initialAR - finalAR // AR should decrease (credit)

	log.Println("📊 RESULTS:")
	log.Printf("   Bank BCA Change: Rp %.2f", bankBCAChange)
	log.Printf("   Piutang Usaha Change: Rp %.2f", arChange)

	// 9. Validate results
	expectedChange := testAmount
	tolerance := 0.01 // Allow small floating point differences

	log.Println("🔍 VALIDATION:")

	// Check Bank BCA increase
	if abs(bankBCAChange-expectedChange) < tolerance {
		log.Printf("✅ Bank BCA: CORRECT! Expected +%.2f, Got +%.2f", expectedChange, bankBCAChange)
	} else {
		log.Printf("❌ Bank BCA: ERROR! Expected +%.2f, Got +%.2f", expectedChange, bankBCAChange)
		if bankBCAChange > expectedChange*1.5 {
			log.Printf("   This indicates DOUBLE JOURNAL ENTRY bug still exists!")
		}
	}

	// Check Piutang Usaha decrease
	if abs(arChange-expectedChange) < tolerance {
		log.Printf("✅ Piutang Usaha: CORRECT! Expected -%.2f, Got -%.2f", expectedChange, arChange)
	} else {
		log.Printf("❌ Piutang Usaha: ERROR! Expected -%.2f, Got -%.2f", expectedChange, arChange)
	}

	// 10. Check journal entries count
	checkJournalEntries(db, payment.ID)

	log.Println("🏁 Test completed!")
}

func checkBankBCABalance(db *gorm.DB) (float64, error) {
	var account models.Account
	err := db.Where("code = ? OR name LIKE ?", "1102", "%BCA%").First(&account).Error
	if err != nil {
		return 0, fmt.Errorf("Bank BCA account not found: %v", err)
	}
	return account.Balance, nil
}

func checkAccountsReceivableBalance(db *gorm.DB) (float64, error) {
	var account models.Account
	err := db.Where("code = ? OR name LIKE ?", "1201", "%Piutang%Usaha%").First(&account).Error
	if err != nil {
		return 0, fmt.Errorf("Accounts Receivable account not found: %v", err)
	}
	return account.Balance, nil
}

func checkJournalEntries(db *gorm.DB, paymentID uint) {
	log.Println("📋 Checking Journal Entries...")

	var journalEntries []models.JournalEntry
	err := db.Where("reference_type = ? AND reference_id = ?", "PAYMENT", paymentID).
		Preload("JournalLines").
		Find(&journalEntries).Error

	if err != nil {
		log.Printf("❌ Error fetching journal entries: %v", err)
		return
	}

	log.Printf("📄 Found %d journal entries for payment %d", len(journalEntries), paymentID)

	if len(journalEntries) == 1 {
		log.Printf("✅ CORRECT: Only 1 journal entry created (fix working!)")
	} else if len(journalEntries) > 1 {
		log.Printf("❌ ERROR: Multiple journal entries found - double entry bug still exists!")
	} else {
		log.Printf("❌ ERROR: No journal entries found - payment not properly recorded!")
	}

	// Show journal entry details
	for i, entry := range journalEntries {
		log.Printf("   Entry %d: ID=%d, Description='%s', Debit=%.2f, Credit=%.2f, Lines=%d",
			i+1, entry.ID, entry.Description, entry.TotalDebit, entry.TotalCredit, len(entry.JournalLines))

		// Show individual journal lines
		for j, line := range entry.JournalLines {
			var account models.Account
			db.First(&account, line.AccountID)
			log.Printf("      Line %d: Account=%s (%s), Debit=%.2f, Credit=%.2f",
				j+1, account.Code, account.Name, line.DebitAmount, line.CreditAmount)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Helper function to pretty print JSON
func prettyPrintJSON(data interface{}) {
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonBytes))
}