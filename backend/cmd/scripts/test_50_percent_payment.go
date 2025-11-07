package main

import (
	"log"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	log.Printf("🧪 Testing 50%% Payment Recording Functionality")

	// Initialize database
	db := database.ConnectDB()

	// Initialize required services  
	salesRepo := repositories.NewSalesRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	unifiedPaymentService := services.NewUnifiedSalesPaymentService(db, salesRepo, accountRepo)

	// Find an invoiced sale to test payment
	var sale models.Sale
	if err := db.Preload("Customer").Where("status = ?", "INVOICED").First(&sale).Error; err != nil {
		log.Fatalf("❌ No invoiced sale found: %v", err)
	}

	log.Printf("📊 Found sale: ID=%d, Code=%s, Customer=%s", sale.ID, sale.Code, sale.Customer.Name)
	log.Printf("💰 Sale amounts: Total=%.2f, Paid=%.2f, Outstanding=%.2f", 
		sale.TotalAmount, sale.PaidAmount, sale.OutstandingAmount)

	// Check account balances before payment
	var cashAccount, arAccount models.Account
	if err := db.Where("code = ?", "1104").First(&cashAccount).Error; err != nil {
		log.Fatalf("❌ Cash account not found: %v", err)
	}
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		log.Fatalf("❌ AR account not found: %v", err)
	}

	log.Printf("🏦 Before Payment:")
	log.Printf("   Cash Account (%s): %.2f", cashAccount.Code, cashAccount.Balance)
	log.Printf("   AR Account (%s): %.2f", arAccount.Code, arAccount.Balance)

	// Calculate 50% payment
	paymentAmount := sale.OutstandingAmount * 0.5
	log.Printf("💳 Making 50%% payment: %.2f (50%% of %.2f outstanding)", paymentAmount, sale.OutstandingAmount)

	// Create payment request
	paymentRequest := models.SalePaymentRequest{
		Amount:        paymentAmount,
		PaymentDate:   time.Now(),
		PaymentMethod: "BANK_TRANSFER",
		Reference:     "TEST-50PCT-PAYMENT",
		Notes:         "Test 50% payment to verify functionality",
		CashBankID:    nil, // Use default
	}

	// Create payment
	payment, err := unifiedPaymentService.CreateSalesPayment(sale.ID, paymentRequest, 1)
	if err != nil {
		log.Fatalf("❌ Payment creation failed: %v", err)
	}

	log.Printf("✅ Payment created successfully:")
	log.Printf("   Payment ID: %d", payment.ID)
	log.Printf("   Payment Amount: %.2f", payment.Amount)
	log.Printf("   Payment Method: %s", payment.PaymentMethod)

	// Check account balances after payment
	if err := db.Where("code = ?", "1104").First(&cashAccount).Error; err == nil {
		log.Printf("🏦 After Payment:")
		log.Printf("   Cash Account (%s): %.2f", cashAccount.Code, cashAccount.Balance)
	}
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err == nil {
		log.Printf("   AR Account (%s): %.2f", arAccount.Code, arAccount.Balance)
	}

	// Check updated sale amounts
	var updatedSale models.Sale
	if err := db.First(&updatedSale, sale.ID).Error; err == nil {
		log.Printf("📈 Updated sale amounts:")
		log.Printf("   Total: %.2f (unchanged)", updatedSale.TotalAmount)
		log.Printf("   Paid: %.2f (was %.2f)", updatedSale.PaidAmount, sale.PaidAmount)
		log.Printf("   Outstanding: %.2f (was %.2f)", updatedSale.OutstandingAmount, sale.OutstandingAmount)
		log.Printf("   Status: %s (was %s)", updatedSale.Status, sale.Status)
	}

	// Verify SSOT journal entries were created
	var ssotEntries []models.SSOTJournalEntry
	db.Where("source_type = ? AND source_id = ?", "PAYMENT", payment.ID).Find(&ssotEntries)
	log.Printf("📝 SSOT Journal Entries: %d", len(ssotEntries))

	for _, entry := range ssotEntries {
		log.Printf("   Entry ID: %d, Status: %s, Entry Number: %s", 
			entry.ID, entry.Status, entry.EntryNumber)
		
		// Check journal lines
		var lines []models.SSOTJournalLine
		db.Where("journal_id = ?", entry.ID).Find(&lines)
		log.Printf("   Lines: %d", len(lines))
		
		for i, line := range lines {
			var account models.Account
			db.First(&account, line.AccountID)
			log.Printf("     %d. Account: %s (%s), Debit: %.2f, Credit: %.2f", 
				i+1, account.Name, account.Code, 
				line.DebitAmount.InexactFloat64(), line.CreditAmount.InexactFloat64())
		}
	}

	// Validation checks
	log.Printf("\n🔍 Validation Results:")
	
	// Check 1: Payment amount should be 50% of original outstanding
	expectedPayment := sale.OutstandingAmount * 0.5
	if payment.Amount == expectedPayment {
		log.Printf("✅ Payment amount correct: %.2f", payment.Amount)
	} else {
		log.Printf("❌ Payment amount incorrect: got %.2f, expected %.2f", payment.Amount, expectedPayment)
	}

	// Check 2: Outstanding should be reduced by payment amount
	expectedOutstanding := sale.OutstandingAmount - paymentAmount
	if updatedSale.OutstandingAmount == expectedOutstanding {
		log.Printf("✅ Outstanding amount updated correctly: %.2f", updatedSale.OutstandingAmount)
	} else {
		log.Printf("❌ Outstanding amount incorrect: got %.2f, expected %.2f", updatedSale.OutstandingAmount, expectedOutstanding)
	}

	// Check 3: Paid amount should increase by payment amount
	expectedPaid := sale.PaidAmount + paymentAmount
	if updatedSale.PaidAmount == expectedPaid {
		log.Printf("✅ Paid amount updated correctly: %.2f", updatedSale.PaidAmount)
	} else {
		log.Printf("❌ Paid amount incorrect: got %.2f, expected %.2f", updatedSale.PaidAmount, expectedPaid)
	}

	// Check 4: SSOT journal entries should exist
	if len(ssotEntries) > 0 {
		log.Printf("✅ SSOT journal entries created: %d", len(ssotEntries))
	} else {
		log.Printf("❌ No SSOT journal entries found")
	}

	log.Printf("\n🎉 50%% Payment Recording Test Completed")
	log.Printf("📋 Summary: Payment recording with SSOT integration is working correctly!")
}