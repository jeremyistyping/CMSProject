package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("==========================================")
	fmt.Println("🔧 FIXING WRONG PAYMENT ALLOCATIONS")
	fmt.Println("==========================================")

	db := database.ConnectDB()

	// Find customer payments that are allocated to bills (wrong!)
	fmt.Println("\n🔍 Finding customer payments wrongly allocated to bills...")

	var wrongAllocations []struct {
		AllocationID    uint    `gorm:"column:allocation_id"`
		PaymentID       uint    `gorm:"column:payment_id"`
		PaymentCode     string  `gorm:"column:payment_code"`
		PaymentAmount   float64 `gorm:"column:payment_amount"`
		ContactName     string  `gorm:"column:contact_name"`
		ContactType     string  `gorm:"column:contact_type"`
		BillID          uint    `gorm:"column:bill_id"`
		AllocatedAmount float64 `gorm:"column:allocated_amount"`
	}

	db.Raw(`
		SELECT 
			pa.id as allocation_id,
			p.id as payment_id,
			p.code as payment_code,
			p.amount as payment_amount,
			c.name as contact_name,
			c.type as contact_type,
			pa.bill_id,
			pa.allocated_amount
		FROM payment_allocations pa
		JOIN payments p ON pa.payment_id = p.id
		JOIN contacts c ON p.contact_id = c.id
		WHERE pa.bill_id IS NOT NULL 
		  AND c.type = 'CUSTOMER'
		ORDER BY pa.id DESC
	`).Scan(&wrongAllocations)

	if len(wrongAllocations) == 0 {
		fmt.Println("✅ No wrong allocations found")
		return
	}

	fmt.Printf("❌ Found %d customer payments wrongly allocated to bills:\n\n", len(wrongAllocations))

	for i, alloc := range wrongAllocations {
		fmt.Printf("  %d. Payment: %s (Rp %.2f) - %s (%s)\n", 
			i+1, alloc.PaymentCode, alloc.PaymentAmount, alloc.ContactName, alloc.ContactType)
		fmt.Printf("     ❌ Currently allocated to Bill ID: %d (Amount: Rp %.2f)\n", 
			alloc.BillID, alloc.AllocatedAmount)
		fmt.Printf("     Should be allocated to an Invoice instead!\n\n")
	}

	fmt.Print("🔧 Do you want to fix these wrong allocations? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" && response != "yes" {
		fmt.Println("🛑 No changes made")
		return
	}

	// Fix each wrong allocation
	fmt.Println("\n🔧 FIXING WRONG ALLOCATIONS...")
	
	totalFixed := 0
	for _, alloc := range wrongAllocations {
		if err := fixWrongAllocation(db, alloc); err != nil {
			log.Printf("❌ Failed to fix allocation %d: %v", alloc.AllocationID, err)
		} else {
			log.Printf("✅ Fixed allocation for payment %s", alloc.PaymentCode)
			totalFixed++
		}
	}

	fmt.Printf("\n🎉 Successfully fixed %d out of %d allocations!\n", totalFixed, len(wrongAllocations))
}

func fixWrongAllocation(db *gorm.DB, wrongAlloc struct {
	AllocationID    uint    `gorm:"column:allocation_id"`
	PaymentID       uint    `gorm:"column:payment_id"`
	PaymentCode     string  `gorm:"column:payment_code"`
	PaymentAmount   float64 `gorm:"column:payment_amount"`
	ContactName     string  `gorm:"column:contact_name"`
	ContactType     string  `gorm:"column:contact_type"`
	BillID          uint    `gorm:"column:bill_id"`
	AllocatedAmount float64 `gorm:"column:allocated_amount"`
}) error {
	
	log.Printf("🔧 Fixing payment %s (%s)", wrongAlloc.PaymentCode, wrongAlloc.ContactName)
	
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Step 1: Remove the allocated amount from the bill
	log.Printf("  1. Removing allocation from Bill ID %d", wrongAlloc.BillID)
	
	err := tx.Exec(`
		UPDATE purchases 
		SET outstanding_amount = outstanding_amount + ?,
			status = CASE 
				WHEN outstanding_amount + ? > 0.01 THEN 'APPROVED'
				ELSE status
			END
		WHERE id = ?
	`, wrongAlloc.AllocatedAmount, wrongAlloc.AllocatedAmount, wrongAlloc.BillID).Error
	
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update bill outstanding: %v", err)
	}

	// Step 2: Find a suitable invoice for this customer
	var saleID uint
	var saleCode string
	var currentOutstanding float64
	
	// Get payment contact ID
	var contactID uint
	tx.Raw("SELECT contact_id FROM payments WHERE id = ?", wrongAlloc.PaymentID).Scan(&contactID)
	
	// Find unpaid invoice for this customer
	err = tx.Raw(`
		SELECT id, code, outstanding_amount 
		FROM sales 
		WHERE customer_id = ? 
		  AND outstanding_amount > 0 
		  AND status = 'INVOICED'
		ORDER BY date ASC 
		LIMIT 1
	`, contactID).Row().Scan(&saleID, &saleCode, &currentOutstanding)
	
	if err != nil {
		// No unpaid invoice found, we'll need to create a generic allocation
		log.Printf("  ⚠️ No unpaid invoice found for customer, keeping as generic allocation")
		
		// Update allocation to remove bill_id (make it generic)
		err = tx.Exec(`
			UPDATE payment_allocations 
			SET bill_id = NULL 
			WHERE id = ?
		`, wrongAlloc.AllocationID).Error
		
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update allocation: %v", err)
		}
		
		tx.Commit()
		log.Printf("  ✅ Converted to generic allocation")
		return nil
	}

	log.Printf("  2. Found suitable invoice: %s (Outstanding: Rp %.2f)", saleCode, currentOutstanding)

	// Step 3: Calculate allocation amount
	allocAmount := wrongAlloc.AllocatedAmount
	if allocAmount > currentOutstanding {
		allocAmount = currentOutstanding
		log.Printf("  ⚠️ Reducing allocation to match invoice outstanding: Rp %.2f", allocAmount)
	}

	// Step 4: Update allocation to point to invoice instead of bill
	err = tx.Exec(`
		UPDATE payment_allocations 
		SET bill_id = NULL,
			invoice_id = ?,
			allocated_amount = ?
		WHERE id = ?
	`, saleID, allocAmount, wrongAlloc.AllocationID).Error
	
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update allocation: %v", err)
	}

	// Step 5: Update invoice outstanding amount
	newOutstanding := currentOutstanding - allocAmount
	status := "INVOICED"
	if newOutstanding <= 0.01 {
		newOutstanding = 0
		status = "PAID"
	}

	err = tx.Exec(`
		UPDATE sales 
		SET outstanding_amount = ?,
			status = ?
		WHERE id = ?
	`, newOutstanding, status, saleID).Error
	
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update invoice outstanding: %v", err)
	}

	log.Printf("  3. Updated invoice %s: Outstanding %.2f → %.2f", 
		saleCode, currentOutstanding, newOutstanding)

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit: %v", err)
	}

	log.Printf("  ✅ Successfully reallocated payment to invoice %s", saleCode)
	return nil
}