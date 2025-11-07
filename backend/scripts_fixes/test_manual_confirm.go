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
	log.Printf("🧪 Testing Sale Confirmation Fix - Manual Test")
	log.Printf("=============================================")

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

	// Step 1: Create a manual DRAFT cash sale
	log.Printf("\n📝 Step 1: Creating manual DRAFT cash sale...")
	
	sale := &models.Sale{
		Code:              "TEST-CASH-001",
		CustomerID:        1, // Assuming customer ID 1 exists
		UserID:            1,
		Type:              models.SaleTypeInvoice,
		Status:            models.SaleStatusDraft,
		Date:              time.Now(),
		DueDate:           time.Now().AddDate(0, 0, 30),
		Currency:          "IDR",
		ExchangeRate:      1,
		PaymentMethodType: "CASH",
		CashBankID:        &[]uint{6}[0], // Use Petty cash (ID 6)
		Subtotal:          100000,
		TotalAmount:       100000,
		OutstandingAmount: 100000,
		PaidAmount:        0,
	}

	// Create the sale directly
	err = db.Create(sale).Error
	if err != nil {
		log.Fatalf("❌ Failed to create manual sale: %v", err)
	}
	
	log.Printf("✅ Created manual DRAFT sale: ID=%d, Code=%s", sale.ID, sale.Code)
	log.Printf("   Status: %s, Total: %.2f, PaymentMethod: %s", 
		sale.Status, sale.TotalAmount, sale.PaymentMethodType)

	// Step 2: Confirm the sale
	log.Printf("\n🔄 Step 2: Confirming sale...")
	err = salesService.ConfirmSale(sale.ID, 1) // Using userID 1
	if err != nil {
		log.Fatalf("❌ Failed to confirm sale: %v", err)
	}
	log.Printf("✅ Sale confirmed successfully")

	// Step 3: Get updated sale details
	log.Printf("\n📋 Step 3: Getting updated sale details...")
	updatedSale, err := salesService.GetSaleByID(sale.ID)
	if err != nil {
		log.Fatalf("❌ Failed to get updated sale: %v", err)
	}

	log.Printf("   Status: %s", updatedSale.Status)
	log.Printf("   Total: %.2f", updatedSale.TotalAmount)
	log.Printf("   Paid: %.2f", updatedSale.PaidAmount)
	log.Printf("   Outstanding: %.2f", updatedSale.OutstandingAmount)

	// Step 4: Verify results
	log.Printf("\n🔍 Step 4: Verifying results...")
	
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

	// Step 5: Check payments created
	log.Printf("\n💳 Step 5: Checking payments created...")
	var payments []models.SalePayment
	err = db.Where("sale_id = ?", sale.ID).Find(&payments).Error
	if err != nil {
		log.Printf("⚠️  Warning: Could not get payments: %v", err)
	} else {
		log.Printf("💰 Found %d payment(s) for sale %d", len(payments), sale.ID)
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

	// Final result
	log.Printf("\n🎯 SUMMARY")
	log.Printf("=========")
	log.Printf("Sale ID: %d", sale.ID)
	log.Printf("Sale Code: %s", updatedSale.Code)
	log.Printf("Initial Status: DRAFT")
	log.Printf("Final Status: %s", updatedSale.Status)
	log.Printf("Payment Method: %s", updatedSale.PaymentMethodType)
	log.Printf("Total Amount: %.2f", updatedSale.TotalAmount)
	log.Printf("Paid Amount: %.2f", updatedSale.PaidAmount)
	log.Printf("Outstanding: %.2f", updatedSale.OutstandingAmount)
	log.Printf("Payments Created: %d", len(payments))

	if testPassed {
		log.Printf("\n🎉 🎉 ALL TESTS PASSED! Sale confirmation fix is working correctly.")
		log.Printf("✅ Sales with CASH payment method are now correctly marked as PAID after confirmation.")
	} else {
		log.Printf("\n❌ ❌ SOME TESTS FAILED. The sale confirmation fix needs further investigation.")
	}

	// Cleanup - delete the test sale
	log.Printf("\n🧹 Cleaning up test data...")
	db.Delete(&models.SalePayment{}, "sale_id = ?", sale.ID)
	db.Delete(sale)
	log.Printf("✅ Test data cleaned up")
}