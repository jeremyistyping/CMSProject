package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	// Load config and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("🔍 Checking Database Journal Tables")
	fmt.Println("====================================")

	// Check which journal tables actually exist
	journalTables := []string{
		"unified_journal_ledger",
		"unified_journal_lines",
		"ssot_journal_entries", 
		"ssot_journal_lines",
		"journal_entries",
		"journal_lines",
	}

	fmt.Println("\n1. Table Existence Check:")
	for _, table := range journalTables {
		var exists bool
		query := fmt.Sprintf(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = '%s'
			)
		`, table)
		
		db.Raw(query).Scan(&exists)
		if exists {
			fmt.Printf("   ✅ %s EXISTS\n", table)
		} else {
			fmt.Printf("   ❌ %s DOES NOT EXIST\n", table)
		}
	}

	// Check data in existing tables
	fmt.Println("\n2. Data Count Check:")
	
	// Check SSOT tables if they exist
	var ssotEntriesCount, ssotLinesCount int64
	db.Raw("SELECT COUNT(*) FROM ssot_journal_entries").Scan(&ssotEntriesCount)
	db.Raw("SELECT COUNT(*) FROM ssot_journal_lines").Scan(&ssotLinesCount)
	fmt.Printf("   📊 ssot_journal_entries: %d records\n", ssotEntriesCount)
	fmt.Printf("   📊 ssot_journal_lines: %d records\n", ssotLinesCount)

	// Check unified tables if they exist
	var unifiedLedgerCount, unifiedLinesCount int64
	err1 := db.Raw("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&unifiedLedgerCount).Error
	err2 := db.Raw("SELECT COUNT(*) FROM unified_journal_lines").Scan(&unifiedLinesCount).Error
	
	if err1 == nil && err2 == nil {
		fmt.Printf("   📊 unified_journal_ledger: %d records\n", unifiedLedgerCount)
		fmt.Printf("   📊 unified_journal_lines: %d records\n", unifiedLinesCount)
	} else {
		fmt.Println("   ❌ unified_journal tables don't exist or error accessing them")
	}

	// Check legacy journal tables
	var journalEntriesCount, journalLinesCount int64
	err3 := db.Raw("SELECT COUNT(*) FROM journal_entries").Scan(&journalEntriesCount).Error
	err4 := db.Raw("SELECT COUNT(*) FROM journal_lines").Scan(&journalLinesCount).Error
	
	if err3 == nil && err4 == nil {
		fmt.Printf("   📊 journal_entries: %d records\n", journalEntriesCount)
		fmt.Printf("   📊 journal_lines: %d records\n", journalLinesCount)
	} else {
		fmt.Println("   ❌ journal_entries/journal_lines tables don't exist or error accessing them")
	}

	// Show sample SSOT entries if they exist
	if ssotEntriesCount > 0 {
		fmt.Println("\n3. Recent SSOT Journal Entries:")
		var entries []struct {
			ID          uint   `json:"id"`
			EntryNumber string `json:"entry_number"`
			Description string `json:"description"`
			TotalDebit  string `json:"total_debit"`
			TotalCredit string `json:"total_credit"`
			Status      string `json:"status"`
		}

		db.Raw("SELECT id, entry_number, description, total_debit, total_credit, status FROM ssot_journal_entries ORDER BY id DESC LIMIT 5").Scan(&entries)
		
		for _, entry := range entries {
			fmt.Printf("   Entry %d: %s - %s (Dr: %s, Cr: %s) [%s]\n",
				entry.ID, entry.EntryNumber, entry.Description, entry.TotalDebit, entry.TotalCredit, entry.Status)
		}
	}

	// Show table structures
	fmt.Println("\n4. SSOT Journal Entries Structure:")
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}
	
	db.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'ssot_journal_entries' 
		ORDER BY ordinal_position
	`).Scan(&columns)
	
	for _, col := range columns {
		fmt.Printf("   %s: %s\n", col.ColumnName, col.DataType)
	}

	fmt.Println("\n✅ Database table analysis complete!")
}