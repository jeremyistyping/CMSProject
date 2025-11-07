package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Println("=== ANALYZING PENDING PAYMENTS ===")
	
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Get all pending payments
	var pendingPayments []models.Payment
	if err := db.Where("status = ?", "PENDING").
		Preload("Contact").
		Find(&pendingPayments).Error; err != nil {
		log.Fatal("Failed to fetch pending payments:", err)
	}

	// Get payment allocations separately
	var paymentAllocations []models.PaymentAllocation
	if len(pendingPayments) > 0 {
		paymentIDs := make([]uint, len(pendingPayments))
		for i, p := range pendingPayments {
			paymentIDs[i] = p.ID
		}
		db.Where("payment_id IN ?", paymentIDs).Find(&paymentAllocations)
	}

	fmt.Printf("\n📋 Found %d PENDING payments:\n", len(pendingPayments))
	fmt.Println("========================================")

	for i, payment := range pendingPayments {
		fmt.Printf("\n%d. Payment ID: %d\n", i+1, payment.ID)
		fmt.Printf("   Code: %s\n", payment.Code)
		fmt.Printf("   Date: %s\n", payment.Date.Format("2006-01-02"))
		fmt.Printf("   Amount: %.2f\n", payment.Amount)
		fmt.Printf("   Method: %s\n", payment.Method)
		fmt.Printf("   Status: %s\n", payment.Status)
		
		if payment.Contact.ID != 0 {
			fmt.Printf("   Contact: %s (ID: %d, Type: %s)\n", 
				payment.Contact.Name, payment.Contact.ID, payment.Contact.Type)
		}

		fmt.Printf("   Created: %s\n", payment.CreatedAt.Format("2006-01-02 15:04:05"))
		if payment.Notes != "" {
			fmt.Printf("   Notes: %s\n", payment.Notes)
		}

		// Check allocations for this payment
		paymentAllocForThis := []models.PaymentAllocation{}
		for _, alloc := range paymentAllocations {
			if alloc.PaymentID == payment.ID {
				paymentAllocForThis = append(paymentAllocForThis, alloc)
			}
		}
		fmt.Printf("   Allocations: %d\n", len(paymentAllocForThis))
		for j, alloc := range paymentAllocForThis {
			fmt.Printf("     %d. Amount: %.2f", j+1, alloc.AllocatedAmount)
			if alloc.InvoiceID != nil {
				fmt.Printf(", Invoice ID: %d", *alloc.InvoiceID)
			}
			if alloc.BillID != nil {
				fmt.Printf(", Bill ID: %d", *alloc.BillID)
			}
			fmt.Printf("\n")
		}
	}

	// Get related data for analysis
	fmt.Println("\n\n=== ANALYSIS ===")
	
	// Check journal entries for these payments
	var journalEntries []models.JournalEntry
	if len(pendingPayments) > 0 {
		paymentIDs := make([]uint, len(pendingPayments))
		for i, p := range pendingPayments {
			paymentIDs[i] = p.ID
		}
		
		db.Where("reference_type = ? AND reference_id IN ?", "PAYMENT", paymentIDs).
			Find(&journalEntries)
		
		fmt.Printf("📋 Journal Entries found: %d (should be %d if all completed)\n", 
			len(journalEntries), len(pendingPayments))
	}

	// Check cash bank transactions
	var cashBankTransactions []models.CashBankTransaction
	if len(pendingPayments) > 0 {
		paymentIDs := make([]uint, len(pendingPayments))
		for i, p := range pendingPayments {
			paymentIDs[i] = p.ID
		}
		
		db.Where("reference_type = ? AND reference_id IN ?", "PAYMENT", paymentIDs).
			Find(&cashBankTransactions)
		
		fmt.Printf("💰 Cash/Bank Transactions found: %d (should be %d if cash/bank updated)\n", 
			len(cashBankTransactions), len(pendingPayments))
	}

	// Check for any recent completed payments to compare
	var recentCompleted []models.Payment
	db.Where("status = ?", "COMPLETED").
		Order("created_at DESC").
		Limit(3).
		Preload("Contact").
		Find(&recentCompleted)

	fmt.Printf("\n📋 Recent COMPLETED payments for comparison:\n")
	for i, payment := range recentCompleted {
		fmt.Printf("%d. %s - %.2f - %s (%s)\n", 
			i+1, payment.Code, payment.Amount, payment.Contact.Name, payment.CreatedAt.Format("2006-01-02 15:04"))
	}

	// Summary statistics
	fmt.Println("\n=== SUMMARY ===")
	var totalPayments int64
	var completedPayments int64
	var pendingTotal float64

	db.Model(&models.Payment{}).Count(&totalPayments)
	db.Model(&models.Payment{}).Where("status = ?", "COMPLETED").Count(&completedPayments)
	
	for _, p := range pendingPayments {
		pendingTotal += p.Amount
	}

	fmt.Printf("Total Payments: %d\n", totalPayments)
	fmt.Printf("Completed: %d\n", completedPayments)
	fmt.Printf("Pending: %d (%.1f%%)\n", len(pendingPayments), 
		float64(len(pendingPayments))/float64(totalPayments)*100)
	fmt.Printf("Pending Amount: %.2f\n", pendingTotal)

	// Check for processing errors
	fmt.Println("\n=== POTENTIAL ISSUES ===")
	
	// Check for missing cash bank associations
	missingCashBank := 0
	for _, p := range pendingPayments {
		hasTransaction := false
		for _, t := range cashBankTransactions {
			if t.ReferenceID == p.ID {
				hasTransaction = true
				break
			}
		}
		if !hasTransaction {
			missingCashBank++
		}
	}
	if missingCashBank > 0 {
		fmt.Printf("❌ %d payments missing cash/bank transactions\n", missingCashBank)
	}

	// Check for missing journal entries
	missingJournal := 0
	for _, p := range pendingPayments {
		hasJournal := false
		for _, j := range journalEntries {
			if j.ReferenceID != nil && *j.ReferenceID == p.ID {
				hasJournal = true
				break
			}
		}
		if !hasJournal {
			missingJournal++
		}
	}
	if missingJournal > 0 {
		fmt.Printf("❌ %d payments missing journal entries\n", missingJournal)
	}

	fmt.Println("\n=== ANALYSIS COMPLETE ===")
}