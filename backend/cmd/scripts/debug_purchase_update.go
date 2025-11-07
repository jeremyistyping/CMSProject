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

	// Get purchases with outstanding amounts first
	fmt.Printf("=== Finding Purchases with Outstanding Amounts ===\n")
	var purchasesWithOutstanding []models.Purchase
	if err := db.Where("outstanding_amount > 0 AND status = 'APPROVED'").Preload("Vendor").Find(&purchasesWithOutstanding).Error; err != nil {
		log.Printf("Failed to get purchases with outstanding: %v", err)
	} else {
		for _, p := range purchasesWithOutstanding {
			fmt.Printf("Purchase: %s | Vendor: %s | Outstanding: %.2f | Status: %s\n", 
				p.Code, p.Vendor.Name, p.OutstandingAmount, p.Status)
		}
	}
	
	// Find recent payments to PT Global Tech (vendor yang sering muncul di screenshots)
	fmt.Printf("\n=== Recent Payments to PT Global Tech ===\n")
	var globalTechPayments []models.Payment
	if err := db.Joins("JOIN contacts ON payments.contact_id = contacts.id").
		Where("contacts.name LIKE ?", "%Global Tech%").
		Order("payments.date DESC").
		Preload("Contact").
		Find(&globalTechPayments).Error; err != nil {
		log.Printf("Failed to get Global Tech payments: %v", err)
	} else {
		for _, p := range globalTechPayments {
			fmt.Printf("Payment: %s | Amount: %.2f | Date: %s | Status: %s\n", 
				p.Code, p.Amount, p.Date.Format("2006-01-02"), p.Status)
		}
	}
	
	// Focus on purchase PO/2025/09/0011 (the one that got 44.4M payments)
	var targetPurchase models.Purchase
	var purchaseID uint
	if err := db.Where("code = ?", "PO/2025/09/0011").Preload("Vendor").First(&targetPurchase).Error; err != nil {
		log.Printf("Target purchase not found, using first outstanding: %v", err)
		if len(purchasesWithOutstanding) > 0 {
			purchaseID = purchasesWithOutstanding[0].ID
		} else {
			purchaseID = uint(1)
		}
	} else {
		purchaseID = targetPurchase.ID
		fmt.Printf("\n=== Target Purchase Found ===\n")
		fmt.Printf("Purchase ID: %d | Code: %s | Outstanding: %.2f\n", 
			targetPurchase.ID, targetPurchase.Code, targetPurchase.OutstandingAmount)
	}
	
	fmt.Printf("=== Debugging Purchase ID %d ===\n", purchaseID)
	
	// 1. Get purchase details
	var purchase models.Purchase
	if err := db.Preload("Vendor").First(&purchase, purchaseID).Error; err != nil {
		log.Fatalf("Failed to get purchase: %v", err)
	}
	
	fmt.Printf("Purchase Code: %s\n", purchase.Code)
	fmt.Printf("Vendor: %s\n", purchase.Vendor.Name)
	fmt.Printf("Total Amount: %.2f\n", purchase.TotalAmount)
	fmt.Printf("Paid Amount: %.2f\n", purchase.PaidAmount)
	fmt.Printf("Outstanding Amount: %.2f\n", purchase.OutstandingAmount)
	fmt.Printf("Status: %s\n", purchase.Status)
	fmt.Printf("Payment Method: %s\n", purchase.PaymentMethod)
	
	// 2. Get payment records for this purchase
	var purchasePayments []models.PurchasePayment
	if err := db.Where("purchase_id = ?", purchaseID).Order("date DESC").Find(&purchasePayments).Error; err != nil {
		log.Printf("Failed to get purchase payments: %v", err)
	} else {
		fmt.Printf("\n=== Purchase Payments ===\n")
		totalPaid := 0.0
		for _, pp := range purchasePayments {
			fmt.Printf("Payment: %s | Date: %s | Amount: %.2f | Method: %s\n", 
				pp.PaymentNumber, pp.Date.Format("2006-01-02"), pp.Amount, pp.Method)
			totalPaid += pp.Amount
		}
		fmt.Printf("Total Payments from purchase_payments: %.2f\n", totalPaid)
	}
	
	// 3. Get payment allocations linked to this purchase
	var paymentAllocations []models.PaymentAllocation
	if err := db.Where("bill_id = ?", purchaseID).Preload("Payment").Find(&paymentAllocations).Error; err != nil {
		log.Printf("Failed to get payment allocations: %v", err)
	} else {
		fmt.Printf("\n=== Payment Allocations ===\n")
		totalAllocated := 0.0
		for _, pa := range paymentAllocations {
			fmt.Printf("Payment ID: %d | Allocated: %.2f | Payment Code: %s | Payment Status: %s\n", 
				pa.PaymentID, pa.AllocatedAmount, pa.Payment.Code, pa.Payment.Status)
			totalAllocated += pa.AllocatedAmount
		}
		fmt.Printf("Total Allocated from payment_allocations: %.2f\n", totalAllocated)
	}
	
	// 4. Get payments from main Payment table for this vendor
	var payments []models.Payment
	if err := db.Where("contact_id = ? AND status = 'COMPLETED'", purchase.VendorID).Order("date DESC").Find(&payments).Error; err != nil {
		log.Printf("Failed to get payments: %v", err)
	} else {
		fmt.Printf("\n=== Payments from Payment Table (for vendor %d) ===\n", purchase.VendorID)
		for _, p := range payments {
			fmt.Printf("Payment: %s | Date: %s | Amount: %.2f | Status: %s | Notes: %s\n", 
				p.Code, p.Date.Format("2006-01-02"), p.Amount, p.Status, p.Notes)
		}
	}
	
	// 5. Check journal entries related to payments
	var journalEntries []models.JournalEntry
	if err := db.Where("reference_type = 'PAYMENT' AND description LIKE ?", "%"+purchase.Code+"%").Preload("JournalLines").Find(&journalEntries).Error; err != nil {
		log.Printf("Failed to get journal entries: %v", err)
	} else {
		fmt.Printf("\n=== Journal Entries for Purchase %s ===\n", purchase.Code)
		for _, je := range journalEntries {
			fmt.Printf("Journal Entry ID: %d | Date: %s | Description: %s\n", 
				je.ID, je.EntryDate.Format("2006-01-02"), je.Description)
			fmt.Printf("  Total Debit: %.2f | Total Credit: %.2f | Status: %s\n", 
				je.TotalDebit, je.TotalCredit, je.Status)
			
			for _, jl := range je.JournalLines {
				fmt.Printf("  Line: Account ID %d | Debit: %.2f | Credit: %.2f | Description: %s\n", 
					jl.AccountID, jl.DebitAmount, jl.CreditAmount, jl.Description)
			}
		}
	}
	
	// 6. Raw SQL to check what's happening
	fmt.Printf("\n=== Raw Purchase Data Check ===\n")
	rows, err := db.Raw("SELECT id, code, paid_amount, outstanding_amount, status, total_amount FROM purchases WHERE id = ?", purchaseID).Rows()
	if err != nil {
		log.Printf("Failed to run raw query: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id uint
			var code string
			var paidAmount, outstandingAmount, totalAmount float64
			var status string
			
			rows.Scan(&id, &code, &paidAmount, &outstandingAmount, &status, &totalAmount)
			fmt.Printf("Raw Data - ID: %d | Code: %s | Total: %.2f | Paid: %.2f | Outstanding: %.2f | Status: %s\n", 
				id, code, totalAmount, paidAmount, outstandingAmount, status)
		}
	}
}