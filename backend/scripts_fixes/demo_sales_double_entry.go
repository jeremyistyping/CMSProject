package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("🚀 Sales Double Entry Accounting Demo")
	fmt.Println("=====================================")

	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.Account{},
		&models.CashBank{},
		&models.Contact{},
		&models.Product{},
		&models.Sale{},
		&models.SaleItem{},
		&models.SalePayment{},
		&models.JournalEntry{},
		&models.JournalLine{},
		&models.User{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Create test data
	userID, customerID, productID, cashBankID, revenueAccountID := createTestData(db)

	// Setup services
	accountRepo := repositories.NewAccountRepository(db)
	journalRepo := repositories.NewJournalEntryRepository(db)
	doubleEntryService := services.NewSalesDoubleEntryService(db, journalRepo, accountRepo)

	// Demo 1: Cash Sale
	fmt.Println("\n💰 Demo 1: Cash Sale Double Entry")
	fmt.Println("----------------------------------")
	
	cashSale := &models.Sale{
		Code:              "CASH-DEMO-001",
		CustomerID:        customerID,
		UserID:            userID,
		Type:              "INVOICE",
		Status:            "INVOICED",
		Date:              time.Now(),
		Currency:          "IDR",
		ExchangeRate:      1,
		PaymentMethodType: "CASH",
		CashBankID:        &cashBankID,
		PaymentTerms:      "COD",
		TotalAmount:       111000,
		SaleItems: []models.SaleItem{
			{
				ProductID:        productID,
				Description:      "Test Product",
				Quantity:         1,
				UnitPrice:        100000,
				LineTotal:        100000,
				FinalAmount:      111000, // Including tax
				RevenueAccountID: revenueAccountID,
			},
		},
	}

	// Save and process sale
	db.Create(cashSale)
	db.Preload("Customer").Preload("SaleItems").First(cashSale, cashSale.ID)

	err = doubleEntryService.CreateSaleJournalEntries(cashSale, userID)
	if err != nil {
		fmt.Printf("❌ Error creating journal entries: %v\n", err)
		return
	}

	// Create immediate payment
	payment, err := doubleEntryService.CreateImmediatePaymentEntry(cashSale, userID)
	if err != nil {
		fmt.Printf("❌ Error creating payment: %v\n", err)
		return
	}

	fmt.Printf("✅ Created cash sale journal entries for sale %s\n", cashSale.Code)
	if payment != nil {
		fmt.Printf("✅ Created immediate payment: %.2f IDR\n", payment.Amount)
	}

	// Demo 2: Credit Sale
	fmt.Println("\n📋 Demo 2: Credit Sale Double Entry")
	fmt.Println("------------------------------------")

	creditSale := &models.Sale{
		Code:              "CREDIT-DEMO-001",
		CustomerID:        customerID,
		UserID:            userID,
		Type:              "INVOICE",
		Status:            "INVOICED",
		Date:              time.Now(),
		DueDate:           time.Now().AddDate(0, 0, 30),
		Currency:          "IDR",
		ExchangeRate:      1,
		PaymentMethodType: "CREDIT",
		CashBankID:        nil,
		PaymentTerms:      "NET_30",
		TotalAmount:       111000,
		SaleItems: []models.SaleItem{
			{
				ProductID:        productID,
				Description:      "Test Product",
				Quantity:         1,
				UnitPrice:        100000,
				LineTotal:        100000,
				FinalAmount:      111000,
				RevenueAccountID: revenueAccountID,
			},
		},
	}

	// Save and process credit sale
	db.Create(creditSale)
	db.Preload("Customer").Preload("SaleItems").First(creditSale, creditSale.ID)

	err = doubleEntryService.CreateSaleJournalEntries(creditSale, userID)
	if err != nil {
		fmt.Printf("❌ Error creating journal entries: %v\n", err)
		return
	}

	// Try immediate payment (should be nil for credit)
	creditPayment, err := doubleEntryService.CreateImmediatePaymentEntry(creditSale, userID)
	if err != nil {
		fmt.Printf("❌ Error testing credit payment: %v\n", err)
		return
	}

	fmt.Printf("✅ Created credit sale journal entries for sale %s\n", creditSale.Code)
	if creditPayment == nil {
		fmt.Println("✅ No immediate payment created for credit sale (as expected)")
	}

	// Show journal entries created
	fmt.Println("\n📊 Journal Entries Summary")
	fmt.Println("--------------------------")

	var journalEntries []models.JournalEntry
	db.Preload("JournalLines").Find(&journalEntries)

	for _, journal := range journalEntries {
		fmt.Printf("Journal: %s - %s (Total: %.2f)\n", journal.Reference, journal.Description, journal.TotalDebit)
		for _, line := range journal.JournalLines {
			if line.DebitAmount > 0 {
				fmt.Printf("  DEBIT:  Account %d - %.2f - %s\n", line.AccountID, line.DebitAmount, line.Description)
			} else {
				fmt.Printf("  CREDIT: Account %d - %.2f - %s\n", line.AccountID, line.CreditAmount, line.Description)
			}
		}
		fmt.Println()
	}

	fmt.Println("🎉 Double Entry Accounting Demo Completed Successfully!")
	fmt.Println("")
	fmt.Println("💡 Summary:")
	fmt.Println("• Cash sales: Debit Cash Account, Credit Revenue Account + Immediate Payment")
	fmt.Println("• Credit sales: Debit Accounts Receivable, Credit Revenue Account (No immediate payment)")
	fmt.Println("• All entries are balanced (Total Debit = Total Credit)")
	fmt.Println("• Account balances are updated automatically when journals are posted")
}

func createTestData(db *gorm.DB) (uint, uint, uint, uint, uint) {
	// Create test user
	user := models.User{
		Username: "demo_user",
		Email:    "demo@example.com",
		Role:     "admin",
		IsActive: true,
	}
	db.Create(&user)

	// Create test customer
	customer := models.Contact{
		Type:     "CUSTOMER",
		Code:     "DEMO001",
		Name:     "Demo Customer",
		IsActive: true,
	}
	db.Create(&customer)

	// Create test product
	product := models.Product{
		Code:      "DEMO-PROD",
		Name:      "Demo Product",
		SalePrice: 100000,
		IsService: false,
		Stock:     100,
		IsActive:  true,
		Unit:      "pcs",
	}
	db.Create(&product)

	// Create cash account
	cashAccount := models.Account{
		Code:        "1101",
		Name:        "Cash",
		Type:        "ASSET",
		Category:    "CURRENT_ASSET",
		IsActive:    true,
		Balance:     0,
		Description: "Cash account",
	}
	db.Create(&cashAccount)

	// Create revenue account
	revenueAccount := models.Account{
		Code:        "4101",
		Name:        "Sales Revenue",
		Type:        "REVENUE",
		Category:    "OPERATING_REVENUE",
		IsActive:    true,
		Balance:     0,
		Description: "Revenue account",
	}
	db.Create(&revenueAccount)

	// Create cash bank record
	cashBank := models.CashBank{
		Code:      "CASH001",
		Name:      "Petty Cash",
		Type:      "CASH",
		AccountID: cashAccount.ID,
		Currency:  "IDR",
		Balance:   0,
		IsActive:  true,
		UserID:    user.ID,
	}
	db.Create(&cashBank)

	return user.ID, customer.ID, product.ID, cashBank.ID, revenueAccount.ID
}