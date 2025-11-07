package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("=== Testing Sales Creation with sales_person_id ===")
	
	// 1. List available employees
	fmt.Println("\n1. Available employees:")
	var employees []models.Contact
	db.Where("type = ? AND is_active = ?", "EMPLOYEE", true).Find(&employees)
	
	for _, emp := range employees {
		fmt.Printf("  ID: %d, Code: %s, Name: %s\n", emp.ID, emp.Code, emp.Name)
	}
	
	if len(employees) == 0 {
		fmt.Println("No employees found!")
		return
	}
	
	// 2. Check customers
	fmt.Println("\n2. Checking for customers...")
	var customers []models.Contact
	db.Where("type = ? AND is_active = ?", "CUSTOMER", true).Find(&customers)
	
	fmt.Printf("Found %d customers\n", len(customers))
	
	var customerID uint
	if len(customers) == 0 {
		// Create a test customer
		fmt.Println("Creating test customer...")
		testCustomer := models.Contact{
			Code:     "CUST001",
			Name:     "Test Customer",
			Type:     "CUSTOMER",
			IsActive: true,
			Email:    "customer@test.com",
			Phone:    "0123456789",
		}
		
		if err := db.Create(&testCustomer).Error; err != nil {
			fmt.Printf("Failed to create test customer: %v\n", err)
			return
		}
		customerID = testCustomer.ID
		fmt.Printf("Created test customer with ID: %d\n", customerID)
	} else {
		customerID = customers[0].ID
		fmt.Printf("Using existing customer: %s (ID: %d)\n", customers[0].Name, customerID)
	}
	
	// 3. Check products
	fmt.Println("\n3. Checking for products...")
	var products []models.Product
	db.Where("is_active = ?", true).Find(&products)
	
	fmt.Printf("Found %d products\n", len(products))
	
	var productID uint
	if len(products) == 0 {
		fmt.Println("No products found. Please create a product first.")
		return
	} else {
		productID = products[0].ID
		fmt.Printf("Using product: %s (ID: %d, Sale Price: %.2f)\n", products[0].Name, productID, products[0].SalePrice)
	}
	
	// 3a. Get a revenue account
	fmt.Println("\n3a. Getting revenue account...")
	var revenueAccounts []models.Account
	db.Where("type = ? AND category = ?", "REVENUE", "SALES").Find(&revenueAccounts)
	
	var revenueAccountID uint
	if len(revenueAccounts) == 0 {
		// Try to find any revenue account
		db.Where("type = ?", "REVENUE").Find(&revenueAccounts)
		if len(revenueAccounts) == 0 {
			fmt.Println("No revenue accounts found. Please create a revenue account first.")
			return
		}
	}
	revenueAccountID = revenueAccounts[0].ID
	fmt.Printf("Using revenue account: %s (ID: %d)\n", revenueAccounts[0].Name, revenueAccountID)
	
	// 4. Create test sale using employee ID 8 (jerly)
	fmt.Println("\n4. Creating test sale with sales_person_id: 8...")
	
	var userID uint = 1 // Assuming user ID 1 exists
	salesPersonID := uint(8) // jerly
	
	testSale := models.Sale{
		Code:          fmt.Sprintf("TEST-SALE-%d", time.Now().Unix()),
		CustomerID:    customerID,
		UserID:        userID,
		SalesPersonID: &salesPersonID,
		Type:          models.SaleTypeInvoice,
		Date:          time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		Currency:      "IDR",
		ExchangeRate:  1.0,
		Status:        models.SaleStatusDraft,
	}
	
	// Create the sale
	err := db.Create(&testSale).Error
	if err != nil {
		fmt.Printf("❌ Failed to create sale: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Successfully created sale with ID: %d\n", testSale.ID)
	
	// 5. Add sale item
	fmt.Println("\n5. Adding sale item...")
	
	testSaleItem := models.SaleItem{
		SaleID:           testSale.ID,
		ProductID:        productID,
		Description:      "Test item",
		Quantity:         2,
		UnitPrice:        100000,
		LineTotal:        200000,
		FinalAmount:      200000,
		Taxable:          true,
		RevenueAccountID: revenueAccountID,
	}
	
	err = db.Create(&testSaleItem).Error
	if err != nil {
		fmt.Printf("❌ Failed to create sale item: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Successfully created sale item with ID: %d\n", testSaleItem.ID)
	
	// 6. Update sale totals
	fmt.Println("\n6. Updating sale totals...")
	
	testSale.Subtotal = 200000
	testSale.TaxableAmount = 200000
	testSale.PPN = 22000 // 11% PPN
	testSale.TotalTax = 22000
	testSale.TotalAmount = 222000
	testSale.OutstandingAmount = 222000
	
	err = db.Save(&testSale).Error
	if err != nil {
		fmt.Printf("❌ Failed to update sale totals: %v\n", err)
		return
	}
	
	fmt.Println("✅ Successfully updated sale totals")
	
	// 7. Verify the created sale with relations
	fmt.Println("\n7. Verifying created sale...")
	
	var createdSale models.Sale
	err = db.Preload("Customer").Preload("SalesPerson").Preload("SaleItems").First(&createdSale, testSale.ID).Error
	if err != nil {
		fmt.Printf("❌ Failed to retrieve created sale: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Sale Details:\n")
	fmt.Printf("   ID: %d\n", createdSale.ID)
	fmt.Printf("   Code: %s\n", createdSale.Code)
	fmt.Printf("   Customer: %s\n", createdSale.Customer.Name)
	if createdSale.SalesPerson != nil {
		fmt.Printf("   Sales Person: %s (ID: %d)\n", createdSale.SalesPerson.Name, *createdSale.SalesPersonID)
	} else {
		fmt.Printf("   Sales Person: NULL\n")
	}
	fmt.Printf("   Total Amount: %.2f\n", createdSale.TotalAmount)
	fmt.Printf("   Sale Items: %d\n", len(createdSale.SaleItems))
	
	fmt.Println("\n=== Test Complete ===")
	fmt.Println("✅ Sales creation with sales_person_id is now working!")
	fmt.Println("The foreign key constraint fix was successful.")
}
