package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔍 CHECKING: Database Journal Table Structure")
	fmt.Println("============================================")

	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	// Check what journal tables exist
	fmt.Println("\n1. CHECKING JOURNAL TABLES:")
	fmt.Println("---------------------------")

	var tableNames []string

	// Get all table names
	err := db.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name LIKE '%journal%' 
		ORDER BY table_name
	`).Pluck("table_name", &tableNames).Error

	if err != nil {
		log.Fatalf("Error querying tables: %v", err)
	}

	fmt.Printf("Found %d journal-related tables:\n", len(tableNames))
	for _, name := range tableNames {
		fmt.Printf("  - %s\n", name)
	}

	// Check unified journal tables specifically
	fmt.Println("\n2. CHECKING UNIFIED JOURNAL STRUCTURE:")
	fmt.Println("--------------------------------------")

	var entryCount, lineCount int64

	// Count entries and lines in unified journals
	db.Raw("SELECT COUNT(*) FROM unified_journal_entries").Scan(&entryCount)
	db.Raw("SELECT COUNT(*) FROM unified_journal_lines").Scan(&lineCount)

	fmt.Printf("Unified Journal Entries: %d\n", entryCount)
	fmt.Printf("Unified Journal Lines: %d\n", lineCount)

	// Get sample entry data
	fmt.Println("\n3. SAMPLE JOURNAL ENTRY DATA:")
	fmt.Println("-----------------------------")

	type JournalSample struct {
		ID          uint    `json:"id"`
		EntryNumber string  `json:"entry_number"`
		SourceType  string  `json:"source_type"`
		SourceID    *uint64 `json:"source_id"`
		Description string  `json:"description"`
		TotalAmount float64 `json:"total_amount"`
		Status      string  `json:"status"`
	}

	var sampleEntries []JournalSample
	db.Raw(`
		SELECT id, entry_number, source_type, source_id, description, total_amount, status
		FROM unified_journal_entries 
		WHERE source_type = 'SALES'
		ORDER BY id DESC
		LIMIT 5
	`).Scan(&sampleEntries)

	for _, entry := range sampleEntries {
		fmt.Printf("Entry ID %d: %s (Source: %s, Status: %s)\n", 
			entry.ID, entry.Description, entry.SourceType, entry.Status)
		fmt.Printf("  Entry Number: %s, Amount: %.2f\n", entry.EntryNumber, entry.TotalAmount)

		// Get lines for this entry
		type LineSample struct {
			ID           uint    `json:"id"`
			AccountID    uint64  `json:"account_id"`
			Description  string  `json:"description"`
			DebitAmount  float64 `json:"debit_amount"`
			CreditAmount float64 `json:"credit_amount"`
			AccountCode  string  `json:"account_code"`
			AccountName  string  `json:"account_name"`
		}

		var lines []LineSample
		db.Raw(`
			SELECT l.id, l.account_id, l.description, l.debit_amount, l.credit_amount,
			       a.code as account_code, a.name as account_name
			FROM unified_journal_lines l
			JOIN accounts a ON l.account_id = a.id
			WHERE l.journal_id = ?
			ORDER BY l.line_number
		`, entry.ID).Scan(&lines)

		for _, line := range lines {
			fmt.Printf("    Line: %s (%s) - Debit: %.2f, Credit: %.2f\n",
				line.AccountCode, line.AccountName, line.DebitAmount, line.CreditAmount)
		}
		fmt.Println()
	}

	// Check for any SSOT tables
	fmt.Println("4. CHECKING FOR SSOT TABLES:")
	fmt.Println("-----------------------------")

	var ssotTables []string
	err = db.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name LIKE '%ssot%' 
		ORDER BY table_name
	`).Pluck("table_name", &ssotTables).Error
	
	if err != nil {
		log.Printf("Warning: Error querying SSOT tables: %v", err)
	}

	fmt.Printf("Found %d SSOT tables:\n", len(ssotTables))
	for _, name := range ssotTables {
		fmt.Printf("  - %s\n", name)
	}

	fmt.Println("\n🔍 JOURNAL TABLE CHECK COMPLETE")
}