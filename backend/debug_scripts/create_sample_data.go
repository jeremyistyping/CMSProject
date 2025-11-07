package main

import (
	"fmt"
	"log"
	"time"
	"math/rand"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("🚀 Creating Sample Financial Data")
	fmt.Println("==================================")

	// Create sample sales data
	createSampleSales(db)
	
	// Create sample purchases data
	createSamplePurchases(db)
	
	// Create sample journal entries
	createSampleJournalEntries(db)

	fmt.Println("\n✅ Sample data created successfully!")
	fmt.Println("\n📊 Now you can test the reports with actual data:")
	fmt.Println("   - Balance Sheet")
	fmt.Println("   - Profit & Loss Statement")
	fmt.Println("   - Sales Summary")
	fmt.Println("   - Purchase Analysis")
}

func createSampleSales(db *gorm.DB) {
	fmt.Println("\n💼 Creating Sample Sales...")

	// Get customer contact
	var customer models.Contact
	db.Where("type = ? AND name ILIKE ?", "CUSTOMER", "%Customer%").First(&customer)
	if customer.ID == 0 {
		// Try to get existing customer with different name
		db.Where("type = ?", "CUSTOMER").First(&customer)
	}
	if customer.ID == 0 {
		// Create sample customer with unique code
		customer = models.Contact{
			Code:    fmt.Sprintf("CUST%d", time.Now().Unix()%10000),
			Name:    "PT. Sample Customer",
			Type:    "CUSTOMER",
			Phone:   "021-12345678",
			Email:   "customer@example.com",
			IsActive: true,
		}
		result := db.Create(&customer)
		if result.Error != nil {
			fmt.Printf("   Error creating customer: %v\n", result.Error)
			return
		}
	}

	// Get product
	var product models.Product
	db.Where("is_active = ?", true).First(&product)
	if product.ID == 0 {
		// Create sample product
		category := models.ProductCategory{
			Code: "CAT001",
			Name: "Sample Category",
		}
		db.Create(&category)
		
		product = models.Product{
			Code:        "PROD001",
			Name:        "Sample Product",
			SalePrice:   1000000,
			CostPrice:   750000,
			Stock:       100,
		CategoryID:  &category.ID,
			IsActive:    true,
		}
		db.Create(&product)
	}

	// Create sales
	sales := []models.Sale{
		{
			Code:            generateCode("SL"),
			Date:            time.Now().AddDate(0, 0, -30), // 30 days ago
			DueDate:         time.Now().AddDate(0, 0, -15),
			CustomerID:      customer.ID,
			UserID:          1, // Default user
			Type:            models.SaleTypeInvoice,
			Status:          models.SaleStatusCompleted,
			Subtotal:        10000000,
			DiscountAmount:  500000,
			TotalTax:        950000,
			TotalAmount:     10450000,
			PaidAmount:      10450000,
			OutstandingAmount: 0,
			Notes:           "Sample sale transaction",
		},
		{
			Code:            generateCode("SL"),
			Date:            time.Now().AddDate(0, 0, -15), // 15 days ago
			DueDate:         time.Now().AddDate(0, 0, -1),
			CustomerID:      customer.ID,
			UserID:          1, // Default user
			Type:            models.SaleTypeInvoice,
			Status:          models.SaleStatusCompleted,
			Subtotal:        5000000,
			DiscountAmount:  0,
			TotalTax:        500000,
			TotalAmount:     5500000,
			PaidAmount:      5500000,
			OutstandingAmount: 0,
			Notes:           "Sample sale transaction 2",
		},
		{
			Code:            generateCode("SL"),
			Date:            time.Now().AddDate(0, 0, -5), // 5 days ago
			DueDate:         time.Now().AddDate(0, 1, 0),
			CustomerID:      customer.ID,
			UserID:          1, // Default user
			Type:            models.SaleTypeInvoice,
			Status:          models.SaleStatusCompleted,  
			Subtotal:        3000000,
			DiscountAmount:  150000,
			TotalTax:        285000,
			TotalAmount:     3135000,
			PaidAmount:      3135000,
			OutstandingAmount: 0,
			Notes:           "Sample sale transaction 3",
		},
	}

	for i, sale := range sales {
		result := db.Create(&sale)
		if result.Error != nil {
			fmt.Printf("   Error creating sale %d: %v\n", i+1, result.Error)
		} else {
				// Create sale item
				saleItem := models.SaleItem{
					SaleID:     sale.ID,
					ProductID:  product.ID,
					Quantity:   i + 5, // 5, 6, 7
					UnitPrice:  product.SalePrice,
					LineTotal:  product.SalePrice * float64(i + 5),
					DiscountAmount: 0,
					TotalTax:   product.SalePrice * float64(i + 5) * 0.1,
					RevenueAccountID: 4101, // Sales revenue account
				}
			db.Create(&saleItem)
			fmt.Printf("   ✓ Created sale: %s (Amount: %.0f)\n", sale.Code, sale.TotalAmount)
		}
	}
}

func createSamplePurchases(db *gorm.DB) {
	fmt.Println("\n🛒 Creating Sample Purchases...")

	// Get vendor contact
	var vendor models.Contact
	db.Where("type = ? AND name ILIKE ?", "VENDOR", "%Vendor%").First(&vendor)
	if vendor.ID == 0 {
		// Try to get existing vendor with different name
		db.Where("type = ?", "VENDOR").First(&vendor)
	}
	if vendor.ID == 0 {
		// Create sample vendor with unique code
		vendor = models.Contact{
			Code:     fmt.Sprintf("VEND%d", time.Now().Unix()%10000),
			Name:     "PT. Sample Vendor",
			Type:     "VENDOR",
			Phone:    "021-87654321",
			Email:    "vendor@example.com",
			IsActive: true,
		}
		result := db.Create(&vendor)
		if result.Error != nil {
			fmt.Printf("   Error creating vendor: %v\n", result.Error)
			return
		}
	}

	// Get product
	var product models.Product
	db.Where("is_active = ?", true).First(&product)

	// Create purchases
	purchases := []models.Purchase{
		{
			Code:                   generateCode("PO"),
			Date:                   time.Now().AddDate(0, 0, -25), // 25 days ago
			DueDate:                time.Now().AddDate(0, 0, -10),
			VendorID:               vendor.ID,
			UserID:                 1, // Default user
			Status:                 models.PurchaseStatusCompleted,
			SubtotalBeforeDiscount: 7500000,
			OrderDiscountAmount:    250000,
			NetBeforeTax:           7250000,
			TaxAmount:              725000,
			TotalAmount:            7975000,
			PaidAmount:             7975000,
			OutstandingAmount:      0,
			MatchingStatus:         models.PurchaseMatchingMatched,
			PaymentMethod:          models.PurchasePaymentCredit,
			Notes:                  "Sample purchase transaction",
		},
		{
			Code:                   generateCode("PO"),
			Date:                   time.Now().AddDate(0, 0, -12), // 12 days ago
			DueDate:                time.Now().AddDate(0, 1, 0),
			VendorID:               vendor.ID,
			UserID:                 1, // Default user
			Status:                 models.PurchaseStatusCompleted,
			SubtotalBeforeDiscount: 4000000,
			OrderDiscountAmount:    0,
			NetBeforeTax:           4000000,
			TaxAmount:              400000,
			TotalAmount:            4400000,
			PaidAmount:             4400000,
			OutstandingAmount:      0,
			MatchingStatus:         models.PurchaseMatchingMatched,
			PaymentMethod:          models.PurchasePaymentCredit,
			Notes:                  "Sample purchase transaction 2",
		},
	}

	for i, purchase := range purchases {
		result := db.Create(&purchase)
		if result.Error != nil {
			fmt.Printf("   Error creating purchase %d: %v\n", i+1, result.Error)
		} else {
				// Create purchase item
				purchaseItem := models.PurchaseItem{
					PurchaseID:       purchase.ID,
					ProductID:        product.ID,
					Quantity:         i + 10, // 10, 11
					UnitPrice:        product.CostPrice,
					TotalPrice:       product.CostPrice * float64(i + 10),
					ExpenseAccountID: 5101, // COGS account
				}
			db.Create(&purchaseItem)
			fmt.Printf("   ✓ Created purchase: %s (Amount: %.0f)\n", purchase.Code, purchase.TotalAmount)
		}
	}
}

func createSampleJournalEntries(db *gorm.DB) {
	fmt.Println("\n📝 Creating Sample Journal Entries...")

	// Get necessary accounts
	var cashAccount, revenueAccount, expenseAccount, cogsAccount models.Account
	
	db.Where("code = ? OR name ILIKE ?", "1101", "%kas%").First(&cashAccount)
	db.Where("code = ? OR name ILIKE ?", "4101", "%pendapatan%").First(&revenueAccount)  
	db.Where("code = ? OR name ILIKE ?", "5101", "%harga pokok%").First(&cogsAccount)
	db.Where("code = ? OR name ILIKE ?", "5202", "%listrik%").First(&expenseAccount)

	if cashAccount.ID == 0 || revenueAccount.ID == 0 {
		fmt.Println("   ⚠️  Warning: Required accounts not found, skipping journal entries")
		return
	}

	// Journal Entry 1: Sales Transaction
	journalEntry1 := models.JournalEntry{
		Code:        generateCode("JE"),
		EntryDate:   time.Now().AddDate(0, 0, -30),
		Description: "Sales revenue for the month",
		Status:      models.JournalStatusPosted,
		TotalDebit:  10000000,
		TotalCredit: 10000000,
		IsBalanced:  true,
		UserID:      1,
	}
	
	result := db.Create(&journalEntry1)
	if result.Error != nil {
		fmt.Printf("   Error creating journal entry 1: %v\n", result.Error)
	} else {
		// Journal lines for sales
		lines := []models.JournalLine{
			{
				JournalEntryID: journalEntry1.ID,
				AccountID:      cashAccount.ID,
				DebitAmount:    10000000,
				CreditAmount:   0,
				Description:    "Cash received from sales",
			},
			{
				JournalEntryID: journalEntry1.ID,
				AccountID:      revenueAccount.ID,
				DebitAmount:    0,
				CreditAmount:   10000000,
				Description:    "Sales revenue",
			},
		}
		
		for _, line := range lines {
			db.Create(&line)
		}
		
		fmt.Printf("   ✓ Created journal entry: %s (Sales)\n", journalEntry1.Code)
	}

	// Journal Entry 2: Purchase Transaction  
		if cogsAccount.ID > 0 {
			journalEntry2 := models.JournalEntry{
				Code:        generateCode("JE"),
				EntryDate:   time.Now().AddDate(0, 0, -25),
				Description: "Purchase of inventory",
				Status:      models.JournalStatusPosted,
				TotalDebit:  7500000,
				TotalCredit: 7500000,
				IsBalanced:  true,
				UserID:      1,
			}
		
		result := db.Create(&journalEntry2)
		if result.Error != nil {
			fmt.Printf("   Error creating journal entry 2: %v\n", result.Error)
		} else {
			// Journal lines for purchase
			lines := []models.JournalLine{
				{
					JournalEntryID: journalEntry2.ID,
					AccountID:      cogsAccount.ID,
					DebitAmount:    7500000,
					CreditAmount:   0,
					Description:    "Cost of goods purchased",
				},
				{
					JournalEntryID: journalEntry2.ID,
					AccountID:      cashAccount.ID,
					DebitAmount:    0,
					CreditAmount:   7500000,
					Description:    "Cash paid for purchase",
				},
			}
			
			for _, line := range lines {
				db.Create(&line)
			}
			
			fmt.Printf("   ✓ Created journal entry: %s (Purchase)\n", journalEntry2.Code)
		}
	}

	// Journal Entry 3: Operating Expense
		if expenseAccount.ID > 0 {
			journalEntry3 := models.JournalEntry{
				Code:        generateCode("JE"),
				EntryDate:   time.Now().AddDate(0, 0, -15),
				Description: "Monthly electricity expense",
				Status:      models.JournalStatusPosted,
				TotalDebit:  500000,
				TotalCredit: 500000,
				IsBalanced:  true,
				UserID:      1,
			}
		
		result := db.Create(&journalEntry3)
		if result.Error != nil {
			fmt.Printf("   Error creating journal entry 3: %v\n", result.Error)
		} else {
			// Journal lines for expense
			lines := []models.JournalLine{
				{
					JournalEntryID: journalEntry3.ID,
					AccountID:      expenseAccount.ID,
					DebitAmount:    500000,
					CreditAmount:   0,
					Description:    "Electricity expense",
				},
				{
					JournalEntryID: journalEntry3.ID,
					AccountID:      cashAccount.ID,
					DebitAmount:    0,
					CreditAmount:   500000,
					Description:    "Cash paid for electricity",
				},
			}
			
			for _, line := range lines {
				db.Create(&line)
			}
			
			fmt.Printf("   ✓ Created journal entry: %s (Expense)\n", journalEntry3.Code)
		}
	}

	fmt.Println("   📊 Journal entries created and posted successfully!")
}

// generateCode generates a simple code for testing purposes
func generateCode(prefix string) string {
	timestamp := time.Now().Unix()
	randomNum := rand.Intn(1000)
	return fmt.Sprintf("%s%d%03d", prefix, timestamp%100000, randomNum)
}
