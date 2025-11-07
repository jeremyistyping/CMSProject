package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔍 Testing Customer History Query...")

	// Initialize database connection
	db := database.ConnectDB()

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get SQL database: %v", err)
	}
	defer sqlDB.Close()

	// Test parameters
	customerID := uint(114)
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fmt.Printf("\nTesting with:\n")
	fmt.Printf("Customer ID: %d\n", customerID)
	fmt.Printf("Start Date: %s\n", startDate.Format("2006-01-02"))
	fmt.Printf("End Date: %s\n", endDate.Format("2006-01-02"))

	// Check if customer exists
	fmt.Println("\n📋 Checking customer...")
	var customer models.Contact
	if err := db.Where("id = ? AND type = ?", customerID, models.ContactTypeCustomer).First(&customer).Error; err != nil {
		log.Printf("❌ Customer not found: %v", err)
	} else {
		fmt.Printf("✅ Customer found: %s (Code: %s)\n", customer.Name, customer.Code)
	}

	// Test sales query
	fmt.Println("\n📋 Testing sales query...")
	var sales []models.Sale
	query := db.Where("customer_id = ? AND date BETWEEN ? AND ?", customerID, startDate, endDate).
		Order("date DESC")
	
	// Print SQL query
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&sales)
	})
	fmt.Printf("SQL Query: %s\n", sql)

	if err := query.Find(&sales).Error; err != nil {
		log.Printf("❌ Sales query error: %v", err)
	} else {
		fmt.Printf("✅ Found %d sales\n", len(sales))
		for i, sale := range sales {
			fmt.Printf("  %d. Code: %s, Date: %s, Amount: %.2f\n", 
				i+1, sale.Code, sale.Date.Format("2006-01-02"), sale.TotalAmount)
		}
	}

	// Test invoices query
	fmt.Println("\n📋 Testing invoices query...")
	var invoices []models.Invoice
	if err := db.Where("customer_id = ? AND date BETWEEN ? AND ?", customerID, startDate, endDate).
		Order("date DESC").Find(&invoices).Error; err != nil {
		log.Printf("❌ Invoices query error: %v", err)
	} else {
		fmt.Printf("✅ Found %d invoices\n", len(invoices))
		for i, inv := range invoices {
			fmt.Printf("  %d. Code: %s, Date: %s, Amount: %.2f\n", 
				i+1, inv.Code, inv.Date.Format("2006-01-02"), inv.TotalAmount)
		}
	}

	// Test payments query
	fmt.Println("\n📋 Testing payments query...")
	var payments []models.Payment
	if err := db.Where("contact_id = ? AND date BETWEEN ? AND ?", customerID, startDate, endDate).
		Order("date DESC").Find(&payments).Error; err != nil {
		log.Printf("❌ Payments query error: %v", err)
	} else {
		fmt.Printf("✅ Found %d payments\n", len(payments))
		for i, pmt := range payments {
			fmt.Printf("  %d. Code: %s, Date: %s, Amount: %.2f\n", 
				i+1, pmt.Code, pmt.Date.Format("2006-01-02"), pmt.Amount)
		}
	}

	// Check Settings table for company info
	fmt.Println("\n📋 Testing settings query...")
	var settings models.Settings
	if err := db.First(&settings).Error; err != nil {
		log.Printf("❌ Settings query error: %v", err)
	} else {
		fmt.Printf("✅ Company settings found: %s\n", settings.CompanyName)
	}

	fmt.Println("\n✅ Test completed!")
}
