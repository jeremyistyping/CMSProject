package main

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
)

func main() {
	log.Printf("🧪 Testing Sale Confirmation Fix - Direct Database Test")
	log.Printf("=====================================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Initialize repositories
	salesRepo := repositories.NewSalesRepository(db)
	productRepo := repositories.NewProductRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	accountRepo := repositories.NewAccountRepository(db)

	// Initialize sales service
	salesService := services.NewSalesService(db, salesRepo, productRepo, contactRepo, accountRepo, nil, nil)
	
	// Step 1: Find or create a draft cash sale for testing
	log.Printf("\n📋 Step 1: Finding or creating test sale...")
	
	// First try to find existing draft sale
	var existingSale models.Sale
	err = db.Where("status = ? AND payment_method_type = ?", models.SaleStatusDraft, "CASH").
		Preload("SaleItems").
		Preload("Customer").
		First(&existingSale).Error

	var testSale *models.Sale
	if err == nil {
		testSale = &existingSale
		log.Printf("✅ Found existing draft cash sale: ID=%d, Code=%s", testSale.ID, testSale.Code)
	} else {
		// Create a new draft cash sale for testing
		log.Printf("📝 Creating new test sale...")
		
		// Get first customer
		var customer models.Contact
		err = db.Where("type = ?", "CUSTOMER").First(&customer).Error
		if err != nil {
			log.Fatalf("❌ No customer found in database: %v", err)
		}

		// Get first product and ensure it has stock
		var product models.Product
		err = db.Where("stock > ?", 0).First(&product).Error
		if err != nil {
			// If no product with stock, create one or update existing
			err = db.First(&product).Error
			if err != nil {
				log.Fatalf("❌ No product found in database: %v", err)
			}
			// Update stock
			product.Stock = 100
			db.Save(&product)
			log.Printf("📎 Updated product %d stock to 100", product.ID)
		}

		// Get first cash bank account with linked account
		var cashBank models.CashBank
		err = db.Where("is_active = ? AND account_id IS NOT NULL", true).First(&cashBank).Error
		if err != nil {
			// Try without account_id filter
			err = db.Where("is_active = ?", true).First(&cashBank).Error
			if err != nil {
				log.Fatalf("❌ No active cash bank found: %v", err)
			}
			log.Printf("⚠️  Warning: Using cash bank without linked account: %s", cashBank.Name)
		}

		dueDate := time.Now().AddDate(0, 0, 30) // 30 days from now

		// Create test sale request
		saleRequest := models.SaleCreateRequest{
			CustomerID:        customer.ID,
			Type:              models.SaleTypeInvoice,
			Date:              time.Now(),
			DueDate:           dueDate,
			PaymentMethodType: "CASH",
			CashBankID:        &cashBank.ID,
			Items: []models.SaleItemRequest{
				{
					ProductID: product.ID,
					Quantity:  1,
					UnitPrice: 50000,
				},
			},
		}

		createdSale, err := salesService.CreateSale(saleRequest, 1) // Using userID 1
		if err != nil {
			log.Fatalf("❌ Failed to create test sale: %v", err)
		}
		testSale = createdSale
		log.Printf("✅ Created test sale: ID=%d, Code=%s", testSale.ID, testSale.Code)
	}

	// Step 2: Verify initial sale status
	log.Printf("\n🔍 Step 2: Verifying initial sale status...")
	log.Printf("   Status: %s", testSale.Status)
	log.Printf("   Total: %.2f", testSale.TotalAmount)
	log.Printf("   Paid: %.2f", testSale.PaidAmount)
	log.Printf("   Outstanding: %.2f", testSale.OutstandingAmount)
	log.Printf("   Payment Method: %s", testSale.PaymentMethodType)
	
	if testSale.Status != models.SaleStatusDraft {
		log.Fatalf("❌ Expected DRAFT status, got %s", testSale.Status)
	}

	// Step 3: Get initial cash bank balance
	log.Printf("\n📊 Step 3: Getting initial cash bank balance...")
	var initialBalance float64
	if testSale.CashBankID != nil {
		var cashBank models.CashBank
		err = db.First(&cashBank, *testSale.CashBankID).Error
		if err != nil {
			log.Printf("⚠️  Warning: Could not get cash bank: %v", err)
		} else {
			initialBalance = cashBank.Balance
			log.Printf("💰 Initial Cash Bank balance (ID %d): %.2f", cashBank.ID, initialBalance)
		}
	}

	// Step 4: Get initial account balances
	log.Printf("\n📊 Step 4: Getting initial account balances...")
	
	// Get Cash/Bank account balance
	var cashAccount models.Account
	cashAccountID := uint(1101) // Assuming Kas account
	err = db.First(&cashAccount, cashAccountID).Error
	if err == nil {
		log.Printf("💰 Initial Cash Account balance (ID %d): %.2f", cashAccount.ID, cashAccount.Balance)
	}
	
	// Get Sales Revenue account balance
	var revenueAccount models.Account
	revenueAccountID := uint(4101) // Assuming Pendapatan Penjualan
	err = db.First(&revenueAccount, revenueAccountID).Error
	if err == nil {
		log.Printf("💰 Initial Sales Revenue balance (ID %d): %.2f", revenueAccount.ID, revenueAccount.Balance)
	}

	// Step 5: Confirm the sale
	log.Printf("\n🔄 Step 5: Confirming sale...")
	err = salesService.ConfirmSale(testSale.ID, 1) // Using userID 1
	if err != nil {
		log.Fatalf("❌ Failed to confirm sale: %v", err)
	}
	log.Printf("✅ Sale confirmed successfully")

	// Step 6: Get updated sale details
	log.Printf("\n📋 Step 6: Getting updated sale details...")
	updatedSale, err := salesService.GetSaleByID(testSale.ID)
	if err != nil {
		log.Fatalf("❌ Failed to get updated sale: %v", err)
	}

	log.Printf("   Status: %s", updatedSale.Status)
	log.Printf("   Total: %.2f", updatedSale.TotalAmount)
	log.Printf("   Paid: %.2f", updatedSale.PaidAmount)
	log.Printf("   Outstanding: %.2f", updatedSale.OutstandingAmount)

	// Step 7: Verify sale status and amounts
	log.Printf("\n🔍 Step 7: Verifying sale status and amounts...")
	
	testPassed := true
	
	if updatedSale.Status == models.SaleStatusPaid {
		log.Printf("✅ ✅ Sale status is correctly PAID")
	} else {
		log.Printf("❌ ❌ Sale status should be PAID but is %s", updatedSale.Status)
		testPassed = false
	}

	if updatedSale.PaidAmount == updatedSale.TotalAmount {
		log.Printf("✅ ✅ PaidAmount matches TotalAmount (%.2f)", updatedSale.PaidAmount)
	} else {
		log.Printf("❌ ❌ PaidAmount (%.2f) should equal TotalAmount (%.2f)", 
			updatedSale.PaidAmount, updatedSale.TotalAmount)
		testPassed = false
	}

	if updatedSale.OutstandingAmount == 0 {
		log.Printf("✅ ✅ OutstandingAmount is correctly 0")
	} else {
		log.Printf("❌ ❌ OutstandingAmount should be 0 but is %.2f", updatedSale.OutstandingAmount)
		testPassed = false
	}

	// Step 8: Check payments created
	log.Printf("\n💳 Step 8: Checking payments created...")
	var payments []models.SalePayment
	err = db.Where("sale_id = ?", testSale.ID).Find(&payments).Error
	if err != nil {
		log.Printf("⚠️  Warning: Could not get payments: %v", err)
	} else {
		log.Printf("💰 Found %d payment(s) for sale %d", len(payments), testSale.ID)
		for _, payment := range payments {
			log.Printf("   - Payment ID=%d, Amount=%.2f, Method=%s, Status=%s", 
				payment.ID, payment.Amount, payment.PaymentMethod, payment.Status)
		}
		
		if len(payments) > 0 {
			log.Printf("✅ ✅ Immediate payment was created")
		} else {
			log.Printf("❌ ❌ No immediate payment found")
			testPassed = false
		}
	}

	// Step 9: Check journal entries created
	log.Printf("\n📋 Step 9: Checking journal entries created...")
	var journalEntries []models.SSOTJournalEntry
	err = db.Where("source_type = ? AND source_id = ?", "SALE", testSale.ID).Find(&journalEntries).Error
	if err != nil {
		log.Printf("⚠️  Warning: Could not get journal entries: %v", err)
	} else {
		log.Printf("📋 Found %d journal entry/entries for sale %d", len(journalEntries), testSale.ID)
		for _, entry := range journalEntries {
			log.Printf("   - Entry ID=%d, TotalDebit=%.2f, TotalCredit=%.2f, Source=%s", 
				entry.ID, entry.TotalDebit, entry.TotalCredit, entry.SourceType)
		}
		
		if len(journalEntries) > 0 {
			log.Printf("✅ ✅ Journal entries were created")
		} else {
			log.Printf("❌ ❌ No journal entries found")
			testPassed = false
		}
	}

	// Step 10: Check updated account balances
	log.Printf("\n📊 Step 10: Checking updated account balances...")
	
	// Check Cash Account balance
	var initialCashBalance float64
	if err == nil {
		initialCashBalance = cashAccount.Balance
	}
	err = db.First(&cashAccount, cashAccountID).Error
	if err == nil {
		balanceIncrease := cashAccount.Balance - initialCashBalance
		log.Printf("💰 Final Cash Account balance (ID %d): %.2f", cashAccount.ID, cashAccount.Balance)
		log.Printf("📈 Cash account balance change: %.2f", balanceIncrease)
	}
	
	// Check Sales Revenue balance
	err = db.First(&revenueAccount, revenueAccountID).Error
	if err == nil {
		log.Printf("💰 Final Sales Revenue balance (ID %d): %.2f", revenueAccount.ID, revenueAccount.Balance)
	}

	// Final result
	log.Printf("\n🎯 SUMMARY")
	log.Printf("=========")
	log.Printf("Sale ID: %d", testSale.ID)
	log.Printf("Sale Code: %s", updatedSale.Code)
	log.Printf("Initial Status: DRAFT")
	log.Printf("Final Status: %s", updatedSale.Status)
	log.Printf("Payment Method: %s", updatedSale.PaymentMethodType)
	log.Printf("Total Amount: %.2f", updatedSale.TotalAmount)
	log.Printf("Paid Amount: %.2f", updatedSale.PaidAmount)
	log.Printf("Outstanding: %.2f", updatedSale.OutstandingAmount)
	log.Printf("Payments Created: %d", len(payments))
	log.Printf("Journal Entries: %d", len(journalEntries))

	if testPassed {
		log.Printf("\n🎉 🎉 ALL TESTS PASSED! Sale confirmation fix is working correctly.")
		log.Printf("✅ Sales with CASH payment method are now correctly marked as PAID after confirmation.")
	} else {
		log.Printf("\n❌ ❌ SOME TESTS FAILED. The sale confirmation fix needs further investigation.")
	}
}