package main

import (
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Println("🔍 Checking for duplicate journals...")
	
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}
	
	// Check for duplicate SALES journals
	type DuplicateInfo struct {
		TransactionType string
		TransactionID   uint
		JournalCount    int64
		TotalAmount     float64
	}
	
	var duplicates []DuplicateInfo
	err := db.Raw(`
		SELECT 
			transaction_type,
			transaction_id,
			COUNT(*) as journal_count,
			SUM(total_amount) as total_amount
		FROM simple_ssot_journals
		WHERE transaction_type IN ('SALES', 'SALES_PAYMENT')
		AND deleted_at IS NULL
		GROUP BY transaction_type, transaction_id
		ORDER BY journal_count DESC, transaction_id
	`).Scan(&duplicates).Error
	
	if err != nil {
		log.Fatalf("Error checking duplicates: %v", err)
	}
	
	log.Printf("\n📊 Found %d transactions with journals:\n", len(duplicates))
	log.Println("┌─────────────────┬────────────┬───────┬──────────────┐")
	log.Println("│ Type            │ Trans ID   │ Count │ Total Amount │")
	log.Println("├─────────────────┼────────────┼───────┼──────────────┤")
	
	hasDuplicates := false
	for _, dup := range duplicates {
		status := "✅"
		if dup.JournalCount > 1 {
			status = "❌ DUPLICATE!"
			hasDuplicates = true
		}
		log.Printf("│ %-15s │ %-10d │ %-5d │ %12.2f │ %s", 
			dup.TransactionType, dup.TransactionID, dup.JournalCount, dup.TotalAmount, status)
	}
	log.Println("└─────────────────┴────────────┴───────┴──────────────┘")
	
	// Show detailed journal entries
	log.Println("\n📋 Detailed Journal Entries:")
	
	type JournalDetail struct {
		JournalID       uint
		TransactionType string
		TransactionID   uint
		EntryNumber     string
		TotalAmount     float64
		AccountCode     string
		AccountName     string
		Debit           float64
		Credit          float64
	}
	
	var details []JournalDetail
	err = db.Raw(`
		SELECT 
			j.id as journal_id,
			j.transaction_type,
			j.transaction_id,
			j.entry_number,
			j.total_amount,
			ji.account_code,
			ji.account_name,
			ji.debit,
			ji.credit
		FROM simple_ssot_journals j
		LEFT JOIN simple_ssot_journal_items ji ON ji.journal_id = j.id
		WHERE j.transaction_type IN ('SALES', 'SALES_PAYMENT')
		AND j.deleted_at IS NULL
		ORDER BY j.created_at DESC, j.id, ji.account_code
	`).Scan(&details).Error
	
	if err != nil {
		log.Printf("Error loading details: %v", err)
	} else {
		currentJournalID := uint(0)
		for _, d := range details {
			if d.JournalID != currentJournalID {
				log.Printf("\n🔖 Journal #%d - %s (Trans #%d) - Total: %.2f", 
					d.JournalID, d.TransactionType, d.TransactionID, d.TotalAmount)
				currentJournalID = d.JournalID
			}
			log.Printf("   %s (%s): Debit %.2f, Credit %.2f", 
				d.AccountCode, d.AccountName, d.Debit, d.Credit)
		}
	}
	
	// Check account balances
	log.Println("\n💰 Current Account Balances:")
	
	var accounts []models.Account
	db.Where("code IN ?", []string{"1102", "1201", "4101", "2103"}).
		Order("code").
		Find(&accounts)
	
	log.Println("┌──────┬──────────────────────┬──────────────┐")
	log.Println("│ Code │ Name                 │ Balance      │")
	log.Println("├──────┼──────────────────────┼──────────────┤")
	for _, acc := range accounts {
		log.Printf("│ %-4s │ %-20s │ %12.2f │", acc.Code, acc.Name, acc.Balance)
	}
	log.Println("└──────┴──────────────────────┴──────────────┘")
	
	if hasDuplicates {
		log.Println("\n❌ DUPLICATE JOURNALS DETECTED!")
		log.Println("   Run cleanup script to fix: go run scripts/cleanup_duplicate_sales_journals.go")
	} else {
		log.Println("\n✅ No duplicate journals found")
	}
}

