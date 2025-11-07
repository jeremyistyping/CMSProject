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
	log.Printf("🧪 Testing Payment 50%% Fix - Ensuring COA reflects actual allocated amount")

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

	// Find an invoiced sale with substantial outstanding amount
	var sale models.Sale
	if err := db.Preload("Customer").Where("status = ?", "INVOICED").Where("outstanding_amount > 1000000").First(&sale).Error; err != nil {
		log.Fatalf("❌ No suitable invoiced sale found: %v", err)
	}

	log.Printf("📊 Found sale: ID=%d, Code=%s, Customer=%s", sale.ID, sale.Code, sale.Customer.Name)
	log.Printf("💰 Sale amounts: Total=%.2f, Outstanding=%.2f", sale.TotalAmount, sale.OutstandingAmount)

	// Get a cash/bank account
	var cashBank models.CashBank
	if err := db.First(&cashBank, 2).Error; err != nil {
		log.Fatalf("❌ Cash/bank account not found: %v", err)
	}

	// Get the related GL account for the cash/bank
	var cashBankGLAccount models.Account
	if err := db.First(&cashBankGLAccount, cashBank.AccountID).Error; err != nil {
		log.Fatalf("❌ Cash/bank GL account not found: %v", err)
	}

	// Get AR account
	var arAccount models.Account
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		log.Fatalf("❌ AR account not found: %v", err)
	}

	log.Printf("🏦 Using Cash/Bank: %s (GL: %s), Balance=%.2f", 
		cashBank.Name, cashBankGLAccount.Name, cashBankGLAccount.Balance)
	log.Printf("🏦 AR Account: %s, Balance=%.2f", arAccount.Name, arAccount.Balance)

	// Record initial balances
	initialCashBalance := cashBankGLAccount.Balance
	initialARBalance := arAccount.Balance
	initialCashBankBalance := cashBank.Balance

	// Calculate 50% of outstanding amount (this should be the ACTUAL amount processed)
	paymentAmount50Percent := sale.OutstandingAmount * 0.5
	
	// But request full amount (this is what user inputs)
	requestAmount := sale.OutstandingAmount  // User wants to pay full amount
	
	log.Printf("💳 Test Scenario:")
	log.Printf("   - User requests to pay: %.2f (100%% of outstanding)", requestAmount)
	log.Printf("   - But only %.2f (50%%) should be allocated and recorded in COA", paymentAmount50Percent)

	startTime := time.Now()

	// Create payment allocation for only 50% (this simulates partial allocation)
	allocations := []services.InvoiceAllocation{
		{
			InvoiceID: sale.ID,
			Amount:    paymentAmount50Percent,  // Only allocate 50%
		},
	}

	// Create payment with full request amount but only 50% allocation
	payment, err := paymentService.CreateReceivablePayment(services.PaymentCreateRequest{
		ContactID:    sale.CustomerID,
		Amount:       requestAmount,  // Full amount requested
		Date:         time.Now(),
		Method:       "BANK_TRANSFER",
		Reference:    "TEST-50PCT-FIX",
		Notes:        "Testing 50% allocation fix",
		CashBankID:   cashBank.ID,
		Allocations:  allocations,  // Only 50% allocated
	}, 1)

	duration := time.Since(startTime)
	
	if err != nil {
		log.Printf("❌ Payment creation failed after %.2f seconds: %v", duration.Seconds(), err)
		return
	}

	log.Printf("✅ Payment created successfully in %.2f seconds", duration.Seconds())
	log.Printf("💳 Payment Details:")
	log.Printf("   - Payment ID: %d", payment.ID)
	log.Printf("   - Payment Code: %s", payment.Code)
	log.Printf("   - Payment Amount (Record): %.2f", payment.Amount)

	// Check updated balances
	if err := db.First(&cashBankGLAccount, cashBank.AccountID).Error; err == nil {
		log.Printf("🏦 After Payment - Cash GL Account Balance: %.2f -> %.2f (Change: +%.2f)", 
			initialCashBalance, cashBankGLAccount.Balance, cashBankGLAccount.Balance - initialCashBalance)
	}
	
	if err := db.First(&arAccount, arAccount.ID).Error; err == nil {
		log.Printf("🏦 After Payment - AR Account Balance: %.2f -> %.2f (Change: %.2f)", 
			initialARBalance, arAccount.Balance, arAccount.Balance - initialARBalance)
	}

	if err := db.First(&cashBank, cashBank.ID).Error; err == nil {
		log.Printf("🏦 After Payment - CashBank Balance: %.2f -> %.2f (Change: +%.2f)", 
			initialCashBankBalance, cashBank.Balance, cashBank.Balance - initialCashBankBalance)
	}

	// Check SSOT journal entries
	var ssotEntries []models.SSOTJournalEntry
	db.Where("source_type = ? AND source_id = ?", "PAYMENT", payment.ID).Find(&ssotEntries)
	log.Printf("📝 SSOT Journal Entries: %d", len(ssotEntries))

	for _, entry := range ssotEntries {
		log.Printf("   Entry ID: %d, Status: %s, Total Debit: %.2f, Total Credit: %.2f", 
			entry.ID, entry.Status, entry.TotalDebit.InexactFloat64(), entry.TotalCredit.InexactFloat64())
		
		// Check journal lines
		var lines []models.SSOTJournalLine
		db.Where("journal_id = ?", entry.ID).Find(&lines)
		for i, line := range lines {
			var account models.Account
			db.First(&account, line.AccountID)
			log.Printf("     Line %d: %s (%s), Debit: %.2f, Credit: %.2f", 
				i+1, account.Name, account.Code, 
				line.DebitAmount.InexactFloat64(), line.CreditAmount.InexactFloat64())
		}
	}

	// Validation checks
	log.Printf("\n🔍 Validation Results:")
	
	// Expected changes based on 50% allocation
	expectedCashIncrease := paymentAmount50Percent
	expectedARDecrease := paymentAmount50Percent
	
	actualCashIncrease := cashBankGLAccount.Balance - initialCashBalance
	actualARDecrease := initialARBalance - arAccount.Balance
	
	// Check 1: Cash account should increase by allocated amount (50%)
	if actualCashIncrease == expectedCashIncrease {
		log.Printf("✅ Cash account correctly increased by %.2f (allocated amount)", actualCashIncrease)
	} else {
		log.Printf("❌ Cash account increase incorrect: got %.2f, expected %.2f", actualCashIncrease, expectedCashIncrease)
	}

	// Check 2: AR account should decrease by allocated amount (50%)
	if actualARDecrease == expectedARDecrease {
		log.Printf("✅ AR account correctly decreased by %.2f (allocated amount)", actualARDecrease)
	} else {
		log.Printf("❌ AR account decrease incorrect: got %.2f, expected %.2f", actualARDecrease, expectedARDecrease)
	}

	// Check 3: Journal entry amounts should match allocated amount
	if len(ssotEntries) > 0 {
		journalAmount := ssotEntries[0].TotalDebit.InexactFloat64()
		if journalAmount == expectedCashIncrease {
			log.Printf("✅ Journal entry amount correct: %.2f (matches allocated amount)", journalAmount)
		} else {
			log.Printf("❌ Journal entry amount incorrect: got %.2f, expected %.2f", journalAmount, expectedCashIncrease)
		}
	}

	// Check 4: CashBank balance should match allocated amount
	actualCashBankIncrease := cashBank.Balance - initialCashBankBalance
	if actualCashBankIncrease == expectedCashIncrease {
		log.Printf("✅ CashBank balance correctly increased by %.2f (allocated amount)", actualCashBankIncrease)
	} else {
		log.Printf("❌ CashBank balance increase incorrect: got %.2f, expected %.2f", actualCashBankIncrease, expectedCashIncrease)
	}

	log.Printf("\n🎉 Payment 50%% Fix Test Completed")
	
	// Final assessment
	allPassed := (actualCashIncrease == expectedCashIncrease) && 
				 (actualARDecrease == expectedARDecrease) &&
				 (actualCashBankIncrease == expectedCashIncrease)
	
	if allPassed {
		log.Printf("📋 RESULT: ✅ SUCCESS - Payment 50%% allocation is correctly recorded in COA")
	} else {
		log.Printf("📋 RESULT: ❌ FAILED - Payment allocation amount mismatch in COA")
	}
}