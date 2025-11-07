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
	log.Printf("🧪 Testing Payment Timeout Fix")

	// Initialize database
	db := database.ConnectDB()

	// Initialize required services
	paymentRepo := repositories.NewPaymentRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	purchaseRepo := repositories.NewPurchaseRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	contactRepo := repositories.NewContactRepository(db)

	paymentService := services.NewPaymentService(db, paymentRepo, salesRepo, purchaseRepo, cashBankRepo, accountRepo, contactRepo)

	// Find an invoiced sale
	var sale models.Sale
	if err := db.Preload("Customer").Where("status = ?", "INVOICED").Where("outstanding_amount > 0").First(&sale).Error; err != nil {
		log.Fatalf("❌ No invoiced sale found: %v", err)
	}

	log.Printf("📊 Found sale: ID=%d, Code=%s, Outstanding=%.2f", sale.ID, sale.Code, sale.OutstandingAmount)

	// Get a cash/bank account
	var cashBank models.CashBank
	if err := db.First(&cashBank, 2).Error; err != nil { // Assuming ID 2 exists
		log.Fatalf("❌ Cash/bank account not found: %v", err)
	}

	log.Printf("🏦 Using Cash/Bank: ID=%d, Name=%s, Balance=%.2f", 
		cashBank.ID, cashBank.Name, cashBank.Balance)

	// Create payment request with timeout test
	paymentAmount := 1000000.00 // 1M IDR - smaller amount for quick test
	
	log.Printf("💳 Creating payment: Amount=%.2f", paymentAmount)
	log.Printf("⏰ Start time: %s", time.Now().Format("15:04:05.000"))

	startTime := time.Now()

	// Create payment allocation
	allocations := []services.InvoiceAllocation{
		{
			InvoiceID: sale.ID,
			Amount:    paymentAmount,
		},
	}

	// Create receivable payment
	payment, err := paymentService.CreateReceivablePayment(services.PaymentCreateRequest{
		ContactID:    sale.CustomerID,
		Amount:       paymentAmount,
		Date:         time.Now(),
		Method:       "BANK_TRANSFER",
		Reference:    "TIMEOUT-FIX-TEST",
		Notes:        "Testing timeout fix for SSOT journal creation",
		CashBankID:   cashBank.ID,
		Allocations:  allocations,
	}, 1)

	duration := time.Since(startTime)
	
	if err != nil {
		log.Printf("❌ Payment creation failed after %.2f seconds: %v", duration.Seconds(), err)
		return
	}

	log.Printf("✅ Payment created successfully!")
	log.Printf("⏰ End time: %s", time.Now().Format("15:04:05.000"))
	log.Printf("⌛ Duration: %.2f seconds", duration.Seconds())
	log.Printf("💳 Payment ID: %d", payment.ID)
	log.Printf("💰 Payment Code: %s", payment.Code)
	log.Printf("💴 Payment Amount: %.2f", payment.Amount)

	// Check if journal entries were created
	var ssotEntries []models.SSOTJournalEntry
	db.Where("source_type = ? AND source_id = ?", "PAYMENT", payment.ID).Find(&ssotEntries)
	log.Printf("📝 SSOT Journal Entries: %d", len(ssotEntries))

	for _, entry := range ssotEntries {
		log.Printf("   Entry ID: %d, Status: %s, Entry Number: %s", 
			entry.ID, entry.Status, entry.EntryNumber)
	}

	// Check updated sale
	var updatedSale models.Sale
	db.First(&updatedSale, sale.ID)
	log.Printf("📈 Sale Outstanding: %.2f -> %.2f", sale.OutstandingAmount, updatedSale.OutstandingAmount)

	// Performance assessment
	if duration.Seconds() < 10 {
		log.Printf("🎉 SUCCESS: Payment completed in %.2f seconds (< 10s threshold)", duration.Seconds())
	} else if duration.Seconds() < 30 {
		log.Printf("⚠️  SLOW: Payment completed in %.2f seconds (10-30s)", duration.Seconds()) 
	} else {
		log.Printf("❌ TIMEOUT RISK: Payment took %.2f seconds (> 30s)", duration.Seconds())
	}

	log.Printf("✅ Payment timeout fix test completed")
}