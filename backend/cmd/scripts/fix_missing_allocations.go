package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// Script to fix missing payment allocations
func main() {
	fmt.Println("==========================================")
	fmt.Println("🔧 FIXING MISSING PAYMENT ALLOCATIONS")
	fmt.Println("==========================================")

	// Connect to database
	db := database.ConnectDB()

	// Find payments without allocations
	var paymentsWithoutAllocations []models.Payment
	db.Raw(`
		SELECT p.* 
		FROM payments p
		LEFT JOIN payment_allocations pa ON pa.payment_id = p.id
		WHERE pa.id IS NULL AND p.status = 'COMPLETED'
	`).Scan(&paymentsWithoutAllocations)

	if len(paymentsWithoutAllocations) == 0 {
		fmt.Println("✅ No payments found without allocations")
		return
	}

	fmt.Printf("Found %d payments without allocations:\n", len(paymentsWithoutAllocations))

	for i, payment := range paymentsWithoutAllocations {
		fmt.Printf("  %d. Payment ID: %d, Code: %s, Amount: %.2f\n", 
			i+1, payment.ID, payment.Code, payment.Amount)
	}

	fmt.Print("\n🤔 Do you want to create allocations for these payments? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" && response != "yes" {
		fmt.Println("🛑 Cancelled by user")
		return
	}

	// Process each payment
	totalFixed := 0
	for _, payment := range paymentsWithoutAllocations {
		if err := createMissingAllocation(db, payment); err != nil {
			log.Printf("❌ Failed to create allocation for payment %d: %v", payment.ID, err)
		} else {
			log.Printf("✅ Created allocation for payment %d", payment.ID)
			totalFixed++
		}
	}

	fmt.Printf("\n🎉 Successfully fixed %d out of %d payments!\n", totalFixed, len(paymentsWithoutAllocations))
	
	// Run diagnostic again to verify
	fmt.Println("\n🔍 Running diagnostic to verify fixes...")
	
	var remainingIssues int64
	db.Raw(`
		SELECT COUNT(*)
		FROM payments p
		LEFT JOIN payment_allocations pa ON pa.payment_id = p.id
		WHERE pa.id IS NULL AND p.status = 'COMPLETED'
	`).Scan(&remainingIssues)

	if remainingIssues == 0 {
		fmt.Println("🎉 SUCCESS: All payment allocation issues have been resolved!")
		fmt.Println("💯 Payment system integrity should now be at 100%")
	} else {
		fmt.Printf("⚠️ Warning: %d payments still need attention\n", remainingIssues)
	}
}

func createMissingAllocation(db *gorm.DB, payment models.Payment) error {
	log.Printf("🔧 Creating allocation for payment %d (%s)", payment.ID, payment.Code)

	// Load payment with contact info
	db.Preload("Contact").First(&payment, payment.ID)

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Strategy: Try to find related invoice/bill based on payment date and contact
	var allocation models.PaymentAllocation

	// First try to find invoice for this customer around payment date
	if payment.Contact.Type == "CUSTOMER" {
		var sale models.Sale
		err := tx.Where("customer_id = ? AND ABS(EXTRACT(DAY FROM AGE(?, date))) <= 30 AND outstanding_amount > 0", 
			payment.ContactID, payment.Date).
			Order("ABS(EXTRACT(DAY FROM AGE(?, date)))", payment.Date).
			First(&sale).Error

		if err == nil && sale.OutstandingAmount >= payment.Amount {
			// Found matching sale
			allocation = models.PaymentAllocation{
				PaymentID:       payment.ID,
				InvoiceID:       &sale.ID,
				AllocatedAmount: payment.Amount,
			}

			// Update sale outstanding amount
			sale.OutstandingAmount -= payment.Amount
			if sale.OutstandingAmount <= 0.01 {
				sale.OutstandingAmount = 0
				sale.Status = models.SaleStatusPaid
			}
			
			if err := tx.Save(&sale).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update sale: %v", err)
			}

			log.Printf("  📋 Allocated to Invoice %d", sale.ID)
		}
	}

	// If no invoice found, try purchase/bill
	if allocation.ID == 0 && payment.Contact.Type == "VENDOR" {
		var purchase models.Purchase
		err := tx.Where("vendor_id = ? AND ABS(EXTRACT(DAY FROM AGE(?, date))) <= 30 AND outstanding_amount > 0", 
			payment.ContactID, payment.Date).
			Order("ABS(EXTRACT(DAY FROM AGE(?, date)))", payment.Date).
			First(&purchase).Error

		if err == nil && purchase.OutstandingAmount >= payment.Amount {
			// Found matching purchase
			allocation = models.PaymentAllocation{
				PaymentID:       payment.ID,
				BillID:          &purchase.ID,
				AllocatedAmount: payment.Amount,
			}

			// Update purchase outstanding amount
			purchase.OutstandingAmount -= payment.Amount
			if purchase.OutstandingAmount <= 0.01 {
				purchase.OutstandingAmount = 0
				purchase.Status = models.PurchaseStatusPaid
			}
			
			if err := tx.Save(&purchase).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update purchase: %v", err)
			}

			log.Printf("  📋 Allocated to Bill %d", purchase.ID)
		}
	}

	// If still no allocation, create a generic one
	if allocation.ID == 0 {
		// For payments without specific invoice/bill, create allocation to the contact's default
		allocation = models.PaymentAllocation{
			PaymentID:       payment.ID,
			AllocatedAmount: payment.Amount,
		}
		
		// Determine if this should be invoice or bill based on contact type
		if payment.Contact.Type == "CUSTOMER" {
			// Look for any unpaid invoice from this customer
			var sale models.Sale
			if err := tx.Where("customer_id = ? AND outstanding_amount > 0", payment.ContactID).
				Order("date ASC").First(&sale).Error; err == nil {
				allocation.InvoiceID = &sale.ID
				
				// Update outstanding amount
				if sale.OutstandingAmount >= payment.Amount {
					sale.OutstandingAmount -= payment.Amount
					if sale.OutstandingAmount <= 0.01 {
						sale.OutstandingAmount = 0
						sale.Status = models.SaleStatusPaid
					}
					tx.Save(&sale)
				}
				log.Printf("  📋 Allocated to oldest unpaid invoice %d", sale.ID)
			}
		} else if payment.Contact.Type == "VENDOR" {
			// Look for any unpaid bill from this vendor
			var purchase models.Purchase
			if err := tx.Where("vendor_id = ? AND outstanding_amount > 0", payment.ContactID).
				Order("date ASC").First(&purchase).Error; err == nil {
				allocation.BillID = &purchase.ID
				
				// Update outstanding amount
				if purchase.OutstandingAmount >= payment.Amount {
					purchase.OutstandingAmount -= payment.Amount
					if purchase.OutstandingAmount <= 0.01 {
						purchase.OutstandingAmount = 0
						purchase.Status = models.PurchaseStatusPaid
					}
					tx.Save(&purchase)
				}
				log.Printf("  📋 Allocated to oldest unpaid bill %d", purchase.ID)
			}
		}
	}

	// Create the allocation
	if err := tx.Create(&allocation).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create allocation: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit: %v", err)
	}

	return nil
}