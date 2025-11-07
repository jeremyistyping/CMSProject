package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=sistem_akuntans_test sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("🔍 JOURNAL TABLE STRUCTURE ANALYSIS")
	fmt.Println("=" + string(make([]byte, 50)) + "=")

	// Check unified_journal_ledger structure
	fmt.Println("\n📊 UNIFIED_JOURNAL_LEDGER TABLE STRUCTURE:")
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'unified_journal_ledger' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Printf("Error querying unified_journal_ledger structure: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var columnName, dataType, isNullable string
			rows.Scan(&columnName, &dataType, &isNullable)
			fmt.Printf("   %s: %s (%s)\n", columnName, dataType, isNullable)
		}
	}

	// Check unified_journal_lines structure
	fmt.Println("\n📊 UNIFIED_JOURNAL_LINES TABLE STRUCTURE:")
	rows, err = db.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'unified_journal_lines' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Printf("Error querying unified_journal_lines structure: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var columnName, dataType, isNullable string
			rows.Scan(&columnName, &dataType, &isNullable)
			fmt.Printf("   %s: %s (%s)\n", columnName, dataType, isNullable)
		}
	}

	// Check existing journal entries
	fmt.Println("\n📊 EXISTING JOURNAL ENTRIES:")
	var ledgerCount, linesCount int
	
	db.QueryRow("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&ledgerCount)
	db.QueryRow("SELECT COUNT(*) FROM unified_journal_lines").Scan(&linesCount)
	
	fmt.Printf("   Ledger entries: %d\n", ledgerCount)
	fmt.Printf("   Journal lines: %d\n", linesCount)

	// Sample data
	fmt.Println("\n📊 SAMPLE LEDGER DATA:")
	rows, err = db.Query("SELECT id, code, description FROM unified_journal_ledger LIMIT 3")
	if err != nil {
		log.Printf("Error querying sample data: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id int
			var code, description string
			rows.Scan(&id, &code, &description)
			fmt.Printf("   %d: %s - %s\n", id, code, description)
		}
	}
}