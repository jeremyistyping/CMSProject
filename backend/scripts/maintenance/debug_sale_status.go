package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	// Check sale with ID 21
	var sale models.Sale
	result := db.Preload("Customer").First(&sale, 21)
	if result.Error != nil {
		log.Printf("Error finding sale ID 21: %v", result.Error)
		
		// Create a test sale if it doesn't exist
		log.Println("Creating test sale with ID 21...")
		testSale := createTestSale(db)
		if testSale != nil {
			log.Printf("Created test sale: ID=%d, Status=%s, TotalAmount=%.2f, OutstandingAmount=%.2f", 
				testSale.ID, testSale.Status, testSale.TotalAmount, testSale.OutstandingAmount)
		}
		return
	}

	log.Printf("Found sale ID 21:")
	log.Printf("  Code: %s", sale.Code)
	log.Printf("  Status: %s", sale.Status)
	log.Printf("  Type: %s", sale.Type)
	log.Printf("  Customer ID: %d", sale.CustomerID)
	if sale.Customer.ID > 0 {
		log.Printf("  Customer Name: %s", sale.Customer.Name)
	}
	log.Printf("  Total Amount: %.2f", sale.TotalAmount)
	log.Printf("  Paid Amount: %.2f", sale.PaidAmount)
	log.Printf("  Outstanding Amount: %.2f", sale.OutstandingAmount)
	log.Printf("  Invoice Number: %s", sale.InvoiceNumber)
	log.Printf("  Created At: %v", sale.CreatedAt)

	// Check if sale can accept payments
	canReceivePayment := sale.Status == models.SaleStatusInvoiced || sale.Status == models.SaleStatusOverdue
	log.Printf("  Can receive payment: %t", canReceivePayment)

	if !canReceivePayment {
		log.Printf("  --> Sale needs to be INVOICED or OVERDUE to receive payments")
		log.Printf("  --> Current status '%s' does not allow payments", sale.Status)
		
		// If sale is DRAFT, let's confirm it to INVOICED
		if sale.Status == models.SaleStatusDraft {
			log.Println("  --> Updating sale status to INVOICED...")
			updateSaleToInvoiced(db, &sale)
		}
	}

	// Check existing payments
	var payments []models.SalePayment
	db.Where("sale_id = ?", 21).Find(&payments)
	log.Printf("  Existing payments: %d", len(payments))
	for i, payment := range payments {
		log.Printf("    Payment %d: Amount=%.2f, Method=%s, Date=%v", 
			i+1, payment.Amount, payment.Method, payment.Date)
	}
}

func createTestSale(db *gorm.DB) *models.Sale {
	// First, ensure we have a test customer
	var customer models.Contact
	result := db.Where("type = ? AND name LIKE ?", "CUSTOMER", "%Test Customer%").First(&customer)
	if result.Error != nil {
		// Create test customer
		customer = models.Contact{
			Type:        "CUSTOMER",
			Name:        "Test Customer",
			Email:       "test@customer.com",
			Phone:       "081234567890",
			IsActive:    true,
			CreditLimit: 10000000,
		}
		if err := db.Create(&customer).Error; err != nil {
			log.Printf("Error creating test customer: %v", err)
			return nil
		}
		log.Printf("Created test customer: ID=%d, Name=%s", customer.ID, customer.Name)
	}

	// Create test sale
	testSale := models.Sale{
		Code:             "INV-2025-TEST21",
		CustomerID:       customer.ID,
		UserID:           1, // Assuming admin user exists
		Type:             models.SaleTypeInvoice,
		Status:           models.SaleStatusInvoiced,
		Date:             time.Now(),
		DueDate:          time.Now().AddDate(0, 0, 30),
		Currency:         "IDR",
		ExchangeRate:     1,
		TotalAmount:      1000000,
		PaidAmount:       0,
		OutstandingAmount: 1000000,
		InvoiceNumber:    "INV/2025/01/TEST21",
		PPNPercent:       11,
		PaymentTerms:     "NET30",
		BillingAddress:   "Test Address",
		Notes:           "Test sale for payment testing",
	}

	// Force ID to be 21
	if err := db.Exec("INSERT INTO sales (id, code, customer_id, user_id, type, status, date, due_date, currency, exchange_rate, total_amount, paid_amount, outstanding_amount, invoice_number, ppn_percent, payment_terms, billing_address, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())",
		21, testSale.Code, testSale.CustomerID, testSale.UserID, testSale.Type, testSale.Status, testSale.Date, testSale.DueDate, testSale.Currency, testSale.ExchangeRate, testSale.TotalAmount, testSale.PaidAmount, testSale.OutstandingAmount, testSale.InvoiceNumber, testSale.PPNPercent, testSale.PaymentTerms, testSale.BillingAddress, testSale.Notes).Error; err != nil {
		log.Printf("Error creating test sale with ID 21: %v", err)
		return nil
	}

	// Retrieve the created sale
	var createdSale models.Sale
	db.First(&createdSale, 21)
	return &createdSale
}

func updateSaleToInvoiced(db *gorm.DB, sale *models.Sale) {
	// Update sale status to INVOICED
	updates := map[string]interface{}{
		"status": models.SaleStatusInvoiced,
	}
	
	// Generate invoice number if not exists
	if sale.InvoiceNumber == "" {
		updates["invoice_number"] = generateInvoiceNumber()
	}
	
	// Set outstanding amount if not set
	if sale.OutstandingAmount == 0 && sale.TotalAmount > 0 {
		updates["outstanding_amount"] = sale.TotalAmount - sale.PaidAmount
	}

	if err := db.Model(sale).Updates(updates).Error; err != nil {
		log.Printf("Error updating sale to INVOICED: %v", err)
		return
	}

	log.Printf("Successfully updated sale ID %d to INVOICED status", sale.ID)
}

func generateInvoiceNumber() string {
	return fmt.Sprintf("INV/%04d/%02d/%04d", time.Now().Year(), time.Now().Month(), 21)
}
