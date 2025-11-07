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
	log.Printf("🧪 Starting Payment SSOT Integration Test")

	// Initialize database
	db := database.ConnectDB()

	// Initialize required services
	salesRepo := repositories.NewSalesRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	
	// Test current UnifiedSalesPaymentService
	unifiedPaymentService := services.NewUnifiedSalesPaymentService(db, salesRepo, accountRepo)
	
	// Also test SSOT Sales Journal Service
	ssotSalesJournal := services.NewSSOTSalesJournalService(db)

	// Find an invoiced sale to test payment
	var sale models.Sale
	if err := db.Preload("Customer").Where("status = ?", "INVOICED").First(&sale).Error; err != nil {
		log.Fatalf("❌ No invoiced sale found: %v", err)
	}

	log.Printf("📊 Found sale for testing: ID=%d, Code=%s, Customer=%s, Total=%.2f, Outstanding=%.2f", 
		sale.ID, sale.Code, sale.Customer.Name, sale.TotalAmount, sale.OutstandingAmount)

	// Check current account balances before payment
	var cashAccount, arAccount models.Account
	if err := db.Where("code = ?", "1104").First(&cashAccount).Error; err != nil { // Bank Mandiri
		log.Fatalf("❌ Cash account not found: %v", err)
	}
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err != nil { // Accounts Receivable
		log.Fatalf("❌ AR account not found: %v", err)
	}

	log.Printf("💰 Before Payment - Cash Account (%s): %.2f", cashAccount.Code, cashAccount.Balance)
	log.Printf("💰 Before Payment - AR Account (%s): %.2f", arAccount.Code, arAccount.Balance)

	// Create a 50% payment request
	paymentAmount := sale.OutstandingAmount * 0.5
	paymentRequest := models.SalePaymentRequest{
		Amount:        paymentAmount,
		PaymentDate:   time.Now(),
		PaymentMethod: "BANK_TRANSFER",
		Reference:     "TEST-PAYMENT-50PCT",
		Notes:         "Test 50% payment for SSOT integration",
		CashBankID:    nil, // Use default
	}

	log.Printf("💳 Creating payment: Amount=%.2f (50%% of %.2f)", paymentAmount, sale.OutstandingAmount)

	// Test 1: Current UnifiedSalesPaymentService
	log.Printf("\n🔄 Test 1: Using UnifiedSalesPaymentService")
	payment, err := unifiedPaymentService.CreateSalesPayment(sale.ID, paymentRequest, 1)
	if err != nil {
		log.Printf("❌ UnifiedSalesPaymentService failed: %v", err)
	} else {
		log.Printf("✅ UnifiedSalesPaymentService success: Payment ID=%d, Amount=%.2f", payment.ID, payment.Amount)
		
		// Check if journal entries were created
		var journalEntries []models.JournalEntry
		db.Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).Find(&journalEntries)
		log.Printf("📝 Old Journal Entries created: %d", len(journalEntries))
		
		// Check SSOT journal entries
		var ssotEntries []models.SSOTJournalEntry
		db.Where("source_type = ? AND source_id = ?", "PAYMENT", payment.ID).Find(&ssotEntries)
		log.Printf("📝 SSOT Journal Entries created: %d", len(ssotEntries))
	}

	// Check account balances after payment
	if err := db.Where("code = ?", "1104").First(&cashAccount).Error; err == nil {
		log.Printf("💰 After Payment - Cash Account (%s): %.2f", cashAccount.Code, cashAccount.Balance)
	}
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err == nil {
		log.Printf("💰 After Payment - AR Account (%s): %.2f", arAccount.Code, arAccount.Balance)
	}

	// Test 2: Direct SSOT Journal Service (if payment was created)
	if payment != nil {
		log.Printf("\n🔄 Test 2: Testing SSOT Journal Service directly")
		ssotEntry, err := ssotSalesJournal.CreatePaymentJournalEntry(payment, 1)
		if err != nil {
			log.Printf("❌ SSOT Journal Service failed: %v", err)
		} else {
			log.Printf("✅ SSOT Journal Service success: Entry ID=%d", ssotEntry.ID)
			
			// Check SSOT journal lines
			var journalLines []models.SSOTJournalLine
			db.Where("journal_id = ?", ssotEntry.ID).Find(&journalLines)
			log.Printf("📚 SSOT Journal Lines created: %d", len(journalLines))
			
			for i, line := range journalLines {
				var account models.Account
				db.First(&account, line.AccountID)
				log.Printf("   %d. Account: %s (%s), Debit: %.2f, Credit: %.2f", 
					i+1, account.Name, account.Code, line.DebitAmount, line.CreditAmount)
			}
		}
	}

	// Check final account balances
	if err := db.Where("code = ?", "1104").First(&cashAccount).Error; err == nil {
		log.Printf("💰 Final - Cash Account (%s): %.2f", cashAccount.Code, cashAccount.Balance)
	}
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err == nil {
		log.Printf("💰 Final - AR Account (%s): %.2f", arAccount.Code, arAccount.Balance)
	}

	// Summary
	log.Printf("\n📋 Test Summary:")
	log.Printf("   - UnifiedSalesPaymentService uses old journal system")
	log.Printf("   - SSOT Sales Journal Service uses new unified journal system")
	log.Printf("   - Need to integrate SSOT into UnifiedSalesPaymentService for proper balance updates")
	
	log.Printf("✅ Payment SSOT Integration Test completed")
}