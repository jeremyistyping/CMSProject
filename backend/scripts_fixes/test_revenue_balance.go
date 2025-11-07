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
	log.Printf("🧪 Testing Revenue Account Balance Increase")
	log.Printf("==========================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Get a revenue account to test with
	var revenueAccount models.Account
	err = db.Where("type = ?", "REVENUE").First(&revenueAccount).Error
	if err != nil {
		log.Fatalf("❌ No revenue account found: %v", err)
	}

	initialBalance := revenueAccount.Balance
	log.Printf("📊 Testing with Revenue Account: ID=%d, Name=%s", revenueAccount.ID, revenueAccount.Name)
	log.Printf("💰 Initial Balance: %.2f", initialBalance)

	// Initialize services
	salesRepo := repositories.NewSalesRepository(db)
	productRepo := repositories.NewProductRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	salesService := services.NewSalesService(db, salesRepo, productRepo, contactRepo, accountRepo, nil, nil)

	// Create manual test sale with specific revenue account
	testAmount := 150000.0
	sale := &models.Sale{
		Code:              "REV-TEST-001",
		CustomerID:        1,
		UserID:            1,
		Type:              models.SaleTypeInvoice,
		Status:            models.SaleStatusDraft,
		Date:              time.Now(),
		DueDate:           time.Now().AddDate(0, 0, 30),
		Currency:          "IDR",
		ExchangeRate:      1,
		PaymentMethodType: "CASH",
		CashBankID:        &[]uint{6}[0], // Petty cash
		Subtotal:          testAmount,
		TotalAmount:       testAmount,
		OutstandingAmount: testAmount,
		PaidAmount:        0,
	}

	// Create the sale
	err = db.Create(sale).Error
	if err != nil {
		log.Fatalf("❌ Failed to create test sale: %v", err)
	}

	// Create sale item with specific revenue account
	saleItem := &models.SaleItem{
		SaleID:           sale.ID,
		ProductID:        1, // Assuming product ID 1 exists
		Description:      "Test product for revenue account",
		Quantity:         1,
		UnitPrice:        testAmount,
		LineTotal:        testAmount,
		FinalAmount:      testAmount,
		TotalPrice:       testAmount,
		RevenueAccountID: revenueAccount.ID, // ← This is the key field!
	}

	err = db.Create(saleItem).Error
	if err != nil {
		log.Fatalf("❌ Failed to create sale item: %v", err)
	}

	log.Printf("✅ Created test sale: ID=%d with revenue account ID=%d", sale.ID, revenueAccount.ID)
	log.Printf("💰 Sale Amount: %.2f", testAmount)

	// Confirm the sale
	log.Printf("\n🔄 Confirming sale to test revenue balance increase...")
	err = salesService.ConfirmSale(sale.ID, 1)
	if err != nil {
		log.Fatalf("❌ Failed to confirm sale: %v", err)
	}

	// Check updated account balance
	err = db.First(&revenueAccount, revenueAccount.ID).Error
	if err != nil {
		log.Fatalf("❌ Failed to get updated account: %v", err)
	}

	finalBalance := revenueAccount.Balance
	balanceIncrease := finalBalance - initialBalance

	log.Printf("\n📊 RESULTS:")
	log.Printf("💰 Initial Balance: %.2f", initialBalance)
	log.Printf("💰 Final Balance: %.2f", finalBalance)
	log.Printf("📈 Balance Increase: %.2f", balanceIncrease)
	log.Printf("🎯 Expected Increase: %.2f", testAmount)

	// Verify results
	if balanceIncrease == testAmount {
		log.Printf("\n✅ ✅ SUCCESS! Revenue account balance increased correctly by %.2f", testAmount)
		log.Printf("🎉 The revenue account selected in the form is working properly!")
	} else {
		log.Printf("\n❌ ❌ MISMATCH: Expected increase %.2f, got %.2f", testAmount, balanceIncrease)
	}

	// Cleanup
	log.Printf("\n🧹 Cleaning up test data...")
	db.Delete(&models.SalePayment{}, "sale_id = ?", sale.ID)
	db.Delete(saleItem)
	db.Delete(sale)
	log.Printf("✅ Cleanup completed")
}