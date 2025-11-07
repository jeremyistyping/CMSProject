package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
)

func main() {
	db := database.ConnectDB()

	var tables []string
	db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name").Scan(&tables)
	
	fmt.Println("=== DATABASE TABLES ===")
	for _, table := range tables {
		fmt.Println(table)
	}
	
	// Check for specific journal tables
	fmt.Println("\n=== JOURNAL RELATED TABLES ===")
	journalTables := []string{"journal_entries", "journal_lines", "unified_journal_ledger", "unified_journal_lines"}
	for _, table := range journalTables {
		var count int64
		db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?", table).Scan(&count)
		if count > 0 {
			fmt.Printf("✅ %s - EXISTS\n", table)
		} else {
			fmt.Printf("❌ %s - MISSING\n", table)
		}
	}
}