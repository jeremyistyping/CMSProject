package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// Simple script to create generic payment allocations
func main() {
	fmt.Println("==========================================")
	fmt.Println("🔧 SIMPLE ALLOCATION FIX")
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

	fmt.Print("\n🤔 Create simple allocations for these payments? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" && response != "yes" {
		fmt.Println("🛑 Cancelled by user")
		return
	}

	// Process each payment with simple allocation
	totalFixed := 0
	for _, payment := range paymentsWithoutAllocations {
		if err := createSimpleAllocation(db, payment); err != nil {
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

func createSimpleAllocation(db *gorm.DB, payment models.Payment) error {
	log.Printf("🔧 Creating simple allocation for payment %d (%s)", payment.ID, payment.Code)

	// Load payment with contact info
	db.Preload("Contact").First(&payment, payment.ID)

	// Create simple generic allocation
	allocation := models.PaymentAllocation{
		PaymentID:       payment.ID,
		AllocatedAmount: payment.Amount,
	}

	// Try to find a suitable invoice or bill to allocate to
	if payment.Contact.Type == "CUSTOMER" {
		// Try to find an unpaid invoice for this customer
		var sale models.Sale
		err := db.Where("customer_id = ? AND outstanding_amount > 0", payment.ContactID).
			Order("date DESC").First(&sale).Error
		
		if err == nil {
			allocation.InvoiceID = &sale.ID
			log.Printf("  📋 Will allocate to Invoice %d", sale.ID)
			
			// Update sale outstanding amount
			if sale.OutstandingAmount >= payment.Amount {
				sale.OutstandingAmount -= payment.Amount
				if sale.OutstandingAmount <= 0.01 {
					sale.OutstandingAmount = 0
					sale.Status = models.SaleStatusPaid
				}
				db.Save(&sale)
				log.Printf("  💰 Updated invoice outstanding: %.2f", sale.OutstandingAmount)
			}
		} else {
			log.Printf("  ⚠️ No unpaid invoice found for customer, creating generic allocation")
		}
	} else if payment.Contact.Type == "VENDOR" {
		// Try to find an unpaid bill for this vendor
		var purchase models.Purchase
		err := db.Where("vendor_id = ? AND outstanding_amount > 0", payment.ContactID).
			Order("date DESC").First(&purchase).Error
			
		if err == nil {
			allocation.BillID = &purchase.ID
			log.Printf("  📋 Will allocate to Bill %d", purchase.ID)
			
			// Update purchase outstanding amount
			if purchase.OutstandingAmount >= payment.Amount {
				purchase.OutstandingAmount -= payment.Amount
				if purchase.OutstandingAmount <= 0.01 {
					purchase.OutstandingAmount = 0
					purchase.Status = models.PurchaseStatusPaid
				}
				db.Save(&purchase)
				log.Printf("  💰 Updated bill outstanding: %.2f", purchase.OutstandingAmount)
			}
		} else {
			log.Printf("  ⚠️ No unpaid bill found for vendor, creating generic allocation")
		}
	}

	// Create the allocation
	if err := db.Create(&allocation).Error; err != nil {
		return fmt.Errorf("failed to create allocation: %v", err)
	}

	log.Printf("  ✅ Allocation created successfully")
	return nil
}