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

	fmt.Printf("=== Debugging Payment Process ===\n")
	
	// 1. Check cash/bank accounts and balances
	var cashBanks []models.CashBank
	if err := db.Preload("Account").Find(&cashBanks).Error; err != nil {
		log.Printf("Failed to get cash/bank accounts: %v", err)
	} else {
		fmt.Printf("\n=== Cash/Bank Accounts ===\n")
		for _, cb := range cashBanks {
			fmt.Printf("ID: %d | Name: %s | Balance: %.2f | Account: %s\n", 
				cb.ID, cb.Name, cb.Balance, cb.Account.Name)
		}
	}
	
	// 2. Check purchase PO/2025/09/0011 details
	var purchase models.Purchase
	if err := db.Where("code = ?", "PO/2025/09/0011").Preload("Vendor").First(&purchase).Error; err != nil {
		log.Printf("Failed to get purchase: %v", err)
		return
	}
	
	fmt.Printf("\n=== Purchase PO/2025/09/0011 Details ===\n")
	fmt.Printf("ID: %d | Total: %.2f | Paid: %.2f | Outstanding: %.2f\n", 
		purchase.ID, purchase.TotalAmount, purchase.PaidAmount, purchase.OutstandingAmount)
	fmt.Printf("Status: %s | Payment Method: %s\n", purchase.Status, purchase.PaymentMethod)
	
	// 3. Check recent payments with COMPLETED status
	var recentPayments []models.Payment
	if err := db.Where("status = 'COMPLETED' AND contact_id = ?", purchase.VendorID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentPayments).Error; err != nil {
		log.Printf("Failed to get recent payments: %v", err)
	} else {
		fmt.Printf("\n=== Recent COMPLETED Payments to Vendor %d ===\n", purchase.VendorID)
		for _, p := range recentPayments {
			fmt.Printf("ID: %d | Code: %s | Amount: %.2f | Date: %s | Status: %s\n", 
				p.ID, p.Code, p.Amount, p.Date.Format("2006-01-02"), p.Status)
			fmt.Printf("  Notes: %s\n", p.Notes)
		}
	}
	
	// 4. Check payment allocations for the recent payments
	if len(recentPayments) > 0 {
		fmt.Printf("\n=== Payment Allocations for Recent Payments ===\n")
		for _, payment := range recentPayments {
			var allocations []models.PaymentAllocation
			if err := db.Where("payment_id = ?", payment.ID).Find(&allocations).Error; err != nil {
				log.Printf("Failed to get allocations for payment %d: %v", payment.ID, err)
				continue
			}
			
			if len(allocations) == 0 {
				fmt.Printf("Payment %s (ID: %d): NO ALLOCATIONS FOUND!\n", payment.Code, payment.ID)
			} else {
				for _, alloc := range allocations {
					billID := "NULL"
					if alloc.BillID != nil {
						billID = fmt.Sprintf("%d", *alloc.BillID)
					}
					fmt.Printf("Payment %s: Allocated %.2f to Bill ID %s\n", 
						payment.Code, alloc.AllocatedAmount, billID)
				}
			}
		}
	}
	
	// 5. Check journal entries for account payable and cash accounts
	fmt.Printf("\n=== Journal Entries Analysis ===\n")
	
	// Find AP account
	var apAccount models.Account
	if err := db.Where("code = ?", "2101").First(&apAccount).Error; err != nil {
		if err := db.Where("LOWER(name) LIKE ?", "%utang%usaha%").First(&apAccount).Error; err != nil {
			log.Printf("AP account not found: %v", err)
		}
	}
	
	if apAccount.ID != 0 {
		fmt.Printf("AP Account: %s (ID: %d) | Balance: %.2f\n", 
			apAccount.Name, apAccount.ID, apAccount.Balance)
		
		// Check recent journal lines for AP account
		var apJournalLines []models.JournalLine
		if err := db.Where("account_id = ?", apAccount.ID).
			Preload("JournalEntry").
			Order("created_at DESC").
			Limit(10).
			Find(&apJournalLines).Error; err != nil {
			log.Printf("Failed to get AP journal lines: %v", err)
		} else {
			fmt.Printf("Recent AP Journal Lines:\n")
			for _, jl := range apJournalLines {
				fmt.Printf("  Entry ID: %d | Date: %s | Debit: %.2f | Credit: %.2f | Description: %s\n",
					jl.JournalEntry.ID, jl.JournalEntry.EntryDate.Format("2006-01-02"),
					jl.DebitAmount, jl.CreditAmount, jl.Description)
			}
		}
	}
	
	// 6. Raw database check for inconsistencies
	fmt.Printf("\n=== Raw Database Consistency Check ===\n")
	
	// Check if outstanding amount calculation is consistent
	rows, err := db.Raw(`
		SELECT p.id, p.code, p.total_amount, p.paid_amount, p.outstanding_amount,
		       COALESCE(SUM(pp.amount), 0) as total_purchase_payments,
		       COALESCE(SUM(pa.allocated_amount), 0) as total_allocations
		FROM purchases p
		LEFT JOIN purchase_payments pp ON p.id = pp.purchase_id
		LEFT JOIN payment_allocations pa ON p.id = pa.bill_id
		WHERE p.code = 'PO/2025/09/0011'
		GROUP BY p.id, p.code, p.total_amount, p.paid_amount, p.outstanding_amount
	`).Rows()
	
	if err != nil {
		log.Printf("Failed to run consistency check: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id uint
			var code string
			var totalAmount, paidAmount, outstandingAmount float64
			var totalPurchasePayments, totalAllocations float64
			
			rows.Scan(&id, &code, &totalAmount, &paidAmount, &outstandingAmount, &totalPurchasePayments, &totalAllocations)
			
			fmt.Printf("Purchase %s (ID: %d):\n", code, id)
			fmt.Printf("  Total: %.2f | Paid: %.2f | Outstanding: %.2f\n", totalAmount, paidAmount, outstandingAmount)
			fmt.Printf("  Purchase Payments Sum: %.2f | Payment Allocations Sum: %.2f\n", totalPurchasePayments, totalAllocations)
			
			// Detect inconsistencies
			expectedOutstanding := totalAmount - paidAmount
			if outstandingAmount != expectedOutstanding {
				fmt.Printf("  ⚠️  INCONSISTENCY: Outstanding (%.2f) != Total-Paid (%.2f)\n", 
					outstandingAmount, expectedOutstanding)
			}
			
			if outstandingAmount > totalAmount {
				fmt.Printf("  ❌ ERROR: Outstanding (%.2f) > Total (%.2f)!\n", 
					outstandingAmount, totalAmount)
			}
		}
	}
}