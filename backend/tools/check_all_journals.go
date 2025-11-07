package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== ALL SSOT JOURNAL ENTRIES ===\n")

	// 1. Cek semua entries di unified_journal_ledger
	fmt.Println("1. ALL UNIFIED_JOURNAL_LEDGER ENTRIES:")
	rows, err := db.Query(`
		SELECT id, source_type, source_code, entry_date, 
			   total_debit, total_credit, status, description
		FROM unified_journal_ledger 
		ORDER BY id DESC
		LIMIT 20
	`)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id int
			var sourceType, sourceCode, status, description string
			var entryDate string
			var totalDebit, totalCredit float64
			
			err := rows.Scan(&id, &sourceType, &sourceCode, &entryDate, 
				&totalDebit, &totalCredit, &status, &description)
			if err != nil {
				fmt.Printf("Error scanning: %v\n", err)
				continue
			}
			
			fmt.Printf("ID:%d - %s %s (%s): Dr=%.0f, Cr=%.0f [%s]\n",
				id, sourceType, sourceCode, entryDate, totalDebit, totalCredit, status)
			fmt.Printf("  Description: %s\n\n", description)
		}
	}

	// 2. Cek journal lines untuk setiap ledger
	fmt.Println("2. DETAILED JOURNAL LINES:")
	rows, err = db.Query(`
		SELECT l.id, l.source_type, l.source_code,
			   ujl.line_number, a.code as account_code, a.name as account_name,
			   ujl.debit_amount, ujl.credit_amount, ujl.description
		FROM unified_journal_ledger l
		JOIN unified_journal_lines ujl ON ujl.journal_id = l.id
		JOIN accounts a ON ujl.account_id = a.id
		ORDER BY l.id, ujl.line_number
	`)
	if err != nil {
		fmt.Printf("Error reading journal lines: %v\n", err)
	} else {
		defer rows.Close()
		currentLedgerId := -1
		for rows.Next() {
			var ledgerId, lineNumber int
			var sourceType, sourceCode, accountCode, accountName, description string
			var debitAmount, creditAmount float64
			
			err := rows.Scan(&ledgerId, &sourceType, &sourceCode,
				&lineNumber, &accountCode, &accountName, 
				&debitAmount, &creditAmount, &description)
			if err != nil {
				fmt.Printf("Error scanning line: %v\n", err)
				continue
			}
			
			if currentLedgerId != ledgerId {
				fmt.Printf("\n--- Journal Entry ID:%d (%s %s) ---\n", 
					ledgerId, sourceType, sourceCode)
				currentLedgerId = ledgerId
			}
			
			if debitAmount > 0 {
				fmt.Printf("  %d. Dr. %s (%s): %.0f - %s\n", 
					lineNumber, accountCode, accountName, debitAmount, description)
			}
			if creditAmount > 0 {
				fmt.Printf("  %d. Cr. %s (%s): %.0f - %s\n", 
					lineNumber, accountCode, accountName, creditAmount, description)
			}
		}
	}

	// 3. Cek apakah purchase menggunakan source_type yang berbeda
	fmt.Println("\n\n3. CHECK PURCHASE-RELATED ENTRIES:")
	rows, err = db.Query(`
		SELECT DISTINCT source_type, source_code 
		FROM unified_journal_ledger 
		WHERE source_code LIKE '%PO%' OR source_code LIKE '%purchase%' 
		   OR source_type LIKE '%purchase%'
		ORDER BY source_type, source_code
	`)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var sourceType, sourceCode string
			err := rows.Scan(&sourceType, &sourceCode)
			if err != nil {
				continue
			}
			fmt.Printf("- %s: %s\n", sourceType, sourceCode)
		}
	}

	fmt.Println("\n=== ALL JOURNALS CHECK SELESAI ===")
}