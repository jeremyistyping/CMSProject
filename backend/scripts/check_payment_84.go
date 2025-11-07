package main

import (
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Println("🔍 Checking Payment #84 details...")
	
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}
	
	// Check if payment #84 exists in sale_payments
	var salePayment models.SalePayment
	if err := db.First(&salePayment, 84).Error; err != nil {
		log.Printf("❌ SalePayment #84 not found: %v", err)
	} else {
		log.Printf("\n📝 SalePayment #84:")
		log.Printf("  Sale ID: %d", salePayment.SaleID)
		log.Printf("  Amount: %.2f", salePayment.Amount)
		log.Printf("  Payment Date: %s", salePayment.PaymentDate.Format("2006-01-02"))
		log.Printf("  Payment Method: %s", salePayment.PaymentMethod)
		log.Printf("  Cash Bank ID: %v", salePayment.CashBankID)
		log.Printf("  Created At: %s", salePayment.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	
	// Check all journals for SALES_PAYMENT type
	type JournalInfo struct {
		JournalID       uint
		TransactionID   uint
		EntryNumber     string
		TotalAmount     float64
		CreatedAt       string
		AccountCode     string
		Debit           float64
		Credit          float64
	}
	
	var journals []JournalInfo
	err := db.Raw(`
		SELECT 
			j.id as journal_id,
			j.transaction_id,
			j.entry_number,
			j.total_amount,
			j.created_at,
			ji.account_code,
			ji.debit,
			ji.credit
		FROM simple_ssot_journals j
		LEFT JOIN simple_ssot_journal_items ji ON ji.journal_id = j.id
		WHERE j.transaction_type = 'SALES_PAYMENT'
		AND j.deleted_at IS NULL
		ORDER BY j.created_at DESC, ji.account_code
	`).Scan(&journals).Error
	
	if err != nil {
		log.Fatalf("Error loading journals: %v", err)
	}
	
	log.Printf("\n📋 All SALES_PAYMENT Journals:")
	log.Println("┌────────────┬──────────────┬─────────────┬──────────────┬─────────────────────┐")
	log.Println("│ Journal ID │ Trans ID     │ Account     │ Debit        │ Credit              │")
	log.Println("├────────────┼──────────────┼─────────────┼──────────────┼─────────────────────┤")
	
	currentJournalID := uint(0)
	for _, j := range journals {
		if j.JournalID != currentJournalID {
			log.Printf("├────────────┴──────────────┴─────────────┴──────────────┴─────────────────────┤")
			log.Printf("│ Journal #%-4d | Trans #%-5d | Total: %-10.2f | Created: %s", 
				j.JournalID, j.TransactionID, j.TotalAmount, j.CreatedAt)
			currentJournalID = j.JournalID
		}
		log.Printf("│ %-10s │ %-12s │ %-11s │ %12.2f │ %19.2f │", 
			"", "", j.AccountCode, j.Debit, j.Credit)
	}
	log.Println("└────────────┴──────────────┴─────────────┴──────────────┴─────────────────────┘")
	
	// Check if there are duplicate payment journals
	type DupCheck struct {
		TransactionID uint
		Count         int64
	}
	
	var dups []DupCheck
	db.Raw(`
		SELECT transaction_id, COUNT(*) as count
		FROM simple_ssot_journals
		WHERE transaction_type = 'SALES_PAYMENT'
		AND deleted_at IS NULL
		GROUP BY transaction_id
		HAVING COUNT(*) > 1
	`).Scan(&dups)
	
	if len(dups) > 0 {
		log.Printf("\n❌ DUPLICATE PAYMENT JOURNALS FOUND:")
		for _, d := range dups {
			log.Printf("  Payment #%d has %d journal entries (should be 1)", d.TransactionID, d.Count)
		}
	} else {
		log.Printf("\n✅ No duplicate payment journals")
	}
	
	// Calculate expected Bank balance
	log.Printf("\n🧮 Balance Calculation:")
	log.Printf("  From Journal #24 (Payment #84): Debit Bank +5,550,000")
	log.Printf("  Expected Bank Balance: 5,550,000")
	log.Printf("  Actual Bank Balance:   11,100,000")
	log.Printf("  Difference:            5,550,000 (EXACTLY ONE PAYMENT AMOUNT!)")
	log.Printf("\n❓ This suggests Bank was credited/debited an extra 5.55M somewhere")
}

