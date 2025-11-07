package main

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	
	// Initialize database connection
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Printf("=== Fixing Payment Inconsistencies ===\n")
	
	// Start transaction for fixing data
	tx := db.Begin()
	
	// 1. Fix Purchase PO/2025/09/0011
	var purchase models.Purchase
	if err := tx.Where("code = ?", "PO/2025/09/0011").First(&purchase).Error; err != nil {
		log.Fatalf("Purchase not found: %v", err)
	}
	
	fmt.Printf("Before Fix - Purchase %s:\n", purchase.Code)
	fmt.Printf("  Total: %.2f | Paid: %.2f | Outstanding: %.2f\n", 
		purchase.TotalAmount, purchase.PaidAmount, purchase.OutstandingAmount)
	
	// 2. Get the recent payments that should be linked to this purchase
	var paymentsToLink []models.Payment
	if err := tx.Where("contact_id = ? AND status = 'COMPLETED' AND notes LIKE ?", 
		purchase.VendorID, "%PO/2025/09/0011%").
		Where("code IN ('PAY/2025/09/0007', 'PAY/2025/09/0008', 'PAY/2025/09/0009')").
		Find(&paymentsToLink).Error; err != nil {
		log.Fatalf("Failed to get payments: %v", err)
	}
	
	fmt.Printf("\\nPayments to Link:\n")
	totalPaymentAmount := 0.0
	for _, payment := range paymentsToLink {
		fmt.Printf("  Payment %s: %.2f\n", payment.Code, payment.Amount)
		totalPaymentAmount += payment.Amount
	}
	
	// 3. Create missing payment allocations
	fmt.Printf("\\nCreating missing payment allocations:\n")
	for _, payment := range paymentsToLink {
		// Check if allocation already exists
		var existingAllocation models.PaymentAllocation
		if err := tx.Where("payment_id = ? AND bill_id = ?", payment.ID, purchase.ID).
			First(&existingAllocation).Error; err == gorm.ErrRecordNotFound {
			
			// Create payment allocation
			allocation := models.PaymentAllocation{
				PaymentID:       payment.ID,
				BillID:          &purchase.ID,
				AllocatedAmount: payment.Amount,
			}
			
			if err := tx.Create(&allocation).Error; err != nil {
				tx.Rollback()
				log.Fatalf("Failed to create payment allocation: %v", err)
			}
			
			fmt.Printf("  ✅ Created allocation for payment %s\n", payment.Code)
		} else {
			fmt.Printf("  ⚠️  Allocation already exists for payment %s\n", payment.Code)
		}
	}
	
	// 4. Create missing purchase_payments records
	fmt.Printf("\\nCreating missing purchase_payments records:\n")
	for _, payment := range paymentsToLink {
		// Check if purchase_payment already exists
		var existingPP models.PurchasePayment
		if err := tx.Where("payment_id = ? AND purchase_id = ?", payment.ID, purchase.ID).
			First(&existingPP).Error; err == gorm.ErrRecordNotFound {
			
			// Create purchase payment record
			purchasePayment := models.PurchasePayment{
				PurchaseID:    purchase.ID,
				PaymentNumber: payment.Code,
				Date:          payment.Date,
				Amount:        payment.Amount,
				Method:        payment.Method,
				Reference:     payment.Reference,
				Notes:         fmt.Sprintf("Fixed payment allocation for %s", purchase.Code),
				UserID:        payment.UserID,
				PaymentID:     &payment.ID,
			}
			
			if err := tx.Create(&purchasePayment).Error; err != nil {
				tx.Rollback()
				log.Fatalf("Failed to create purchase payment: %v", err)
			}
			
			fmt.Printf("  ✅ Created purchase_payment for %s\n", payment.Code)
		} else {
			fmt.Printf("  ⚠️  Purchase payment already exists for %s\n", payment.Code)
		}
	}
	
	// 5. Fix purchase paid amount and outstanding amount
	fmt.Printf("\\nFixing purchase amounts:\n")
	
	// Calculate correct amounts
	var totalAllocatedAmount float64
	if err := tx.Model(&models.PaymentAllocation{}).
		Where("bill_id = ?", purchase.ID).
		Select("COALESCE(SUM(allocated_amount), 0)").
		Scan(&totalAllocatedAmount).Error; err != nil {
		tx.Rollback()
		log.Fatalf("Failed to calculate allocated amount: %v", err)
	}
	
	// Update purchase amounts
	oldPaidAmount := purchase.PaidAmount
	oldOutstandingAmount := purchase.OutstandingAmount
	
	purchase.PaidAmount = totalAllocatedAmount
	purchase.OutstandingAmount = purchase.TotalAmount - purchase.PaidAmount
	
	// Update status if fully paid
	if purchase.OutstandingAmount <= 0.01 {
		purchase.Status = models.PurchaseStatusPaid
		purchase.OutstandingAmount = 0
	}
	
	if err := tx.Save(&purchase).Error; err != nil {
		tx.Rollback()
		log.Fatalf("Failed to update purchase: %v", err)
	}
	
	fmt.Printf("Purchase Amount Updates:\n")
	fmt.Printf("  Paid Amount: %.2f -> %.2f\n", oldPaidAmount, purchase.PaidAmount)
	fmt.Printf("  Outstanding: %.2f -> %.2f\n", oldOutstandingAmount, purchase.OutstandingAmount)
	fmt.Printf("  Status: %s\n", purchase.Status)
	
	// 6. Optional: Fix other purchases with similar issues
	fmt.Printf("\\nChecking for other purchases with outstanding > total:\n")
	var problematicPurchases []models.Purchase
	if err := tx.Where("outstanding_amount > total_amount").Find(&problematicPurchases).Error; err != nil {
		log.Printf("Failed to find problematic purchases: %v", err)
	} else {
		for _, p := range problematicPurchases {
			fmt.Printf("  Purchase %s: Total=%.2f, Outstanding=%.2f\n", 
				p.Code, p.TotalAmount, p.OutstandingAmount)
			
			// Recalculate based on payment allocations
			var allocatedAmount float64
			if err := tx.Model(&models.PaymentAllocation{}).
				Where("bill_id = ?", p.ID).
				Select("COALESCE(SUM(allocated_amount), 0)").
				Scan(&allocatedAmount).Error; err != nil {
				log.Printf("    Failed to calculate allocated for %s: %v", p.Code, err)
				continue
			}
			
			// Fix the amounts
			p.PaidAmount = allocatedAmount
			p.OutstandingAmount = p.TotalAmount - p.PaidAmount
			
			if p.OutstandingAmount <= 0.01 {
				p.Status = models.PurchaseStatusPaid
				p.OutstandingAmount = 0
			}
			
			if err := tx.Save(&p).Error; err != nil {
				log.Printf("    Failed to fix purchase %s: %v", p.Code, err)
				continue
			}
			
			fmt.Printf("    ✅ Fixed: Paid=%.2f, Outstanding=%.2f, Status=%s\n", 
				p.PaidAmount, p.OutstandingAmount, p.Status)
		}
	}
	
	// Commit all changes
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}
	
	fmt.Printf("\\n🎉 All payment inconsistencies have been fixed!\n")
	
	// 7. Verify the fix
	var fixedPurchase models.Purchase
	if err := db.Where("code = ?", "PO/2025/09/0011").First(&fixedPurchase).Error; err != nil {
		log.Printf("Failed to verify fix: %v", err)
	} else {
		fmt.Printf("\\n=== Verification ===\n")
		fmt.Printf("Purchase %s after fix:\n", fixedPurchase.Code)
		fmt.Printf("  Total: %.2f | Paid: %.2f | Outstanding: %.2f\n", 
			fixedPurchase.TotalAmount, fixedPurchase.PaidAmount, fixedPurchase.OutstandingAmount)
		fmt.Printf("  Status: %s\n", fixedPurchase.Status)
		
		// Check allocations
		var allocationCount int64
		db.Model(&models.PaymentAllocation{}).Where("bill_id = ?", fixedPurchase.ID).Count(&allocationCount)
		fmt.Printf("  Payment Allocations: %d\n", allocationCount)
	}
}