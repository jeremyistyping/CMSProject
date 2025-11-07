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

	fmt.Println("=== SSOT JOURNAL SYSTEM CHECK ===\n")

	// 1. Cek struktur unified_journal_ledger
	fmt.Println("1. UNIFIED_JOURNAL_LEDGER STRUCTURE:")
	rows, err := db.Query(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'unified_journal_ledger' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var columnName, dataType string
			err := rows.Scan(&columnName, &dataType)
			if err != nil {
				continue
			}
			fmt.Printf("- %s (%s)\n", columnName, dataType)
		}
	}

	// 2. Cek data di unified_journal_ledger
	fmt.Println("\n2. UNIFIED_JOURNAL_LEDGER DATA:")
	rows, err = db.Query(`
		SELECT id, source_type, source_code, entry_date, 
			   total_debit, total_credit, status, description
		FROM unified_journal_ledger 
		WHERE source_type = 'purchase' OR source_code LIKE 'PO%'
		ORDER BY entry_date DESC
		LIMIT 10
	`)
	if err != nil {
		fmt.Printf("Error reading unified_journal_ledger: %v\n", err)
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
			
			fmt.Printf("ID:%d - %s %s (%s): Dr=%.0f, Cr=%.0f [%s] - %s\n",
				id, sourceType, sourceCode, entryDate, totalDebit, totalCredit, status, description)
		}
	}

	// 3. Cek unified_journal_lines structure
	fmt.Println("\n3. UNIFIED_JOURNAL_LINES STRUCTURE:")
	rows, err = db.Query(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'unified_journal_lines' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var columnName, dataType string
			err := rows.Scan(&columnName, &dataType)
			if err != nil {
				continue
			}
			fmt.Printf("- %s (%s)\n", columnName, dataType)
		}
	}

	// 4. Cek journal lines untuk purchase
	fmt.Println("\n4. JOURNAL LINES FOR PURCHASES:")
	rows, err = db.Query(`
		SELECT ujl.journal_id, a.code as account_code, a.name as account_name,
			   ujl.debit_amount, ujl.credit_amount, ujl.description,
			   l.source_code
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger l ON ujl.journal_id = l.id
		JOIN accounts a ON ujl.account_id = a.id
		WHERE l.source_type = 'purchase' OR l.source_code LIKE 'PO%'
		ORDER BY l.source_code, a.code
	`)
	if err != nil {
		fmt.Printf("Error reading journal lines: %v\n", err)
	} else {
		defer rows.Close()
		currentRef := ""
		for rows.Next() {
			var ledgerId int
			var accountCode, accountName, description, refId string
			var debitAmount, creditAmount float64
			
			err := rows.Scan(&ledgerId, &accountCode, &accountName, 
				&debitAmount, &creditAmount, &description, &refId)
			if err != nil {
				fmt.Printf("Error scanning line: %v\n", err)
				continue
			}
			
			if currentRef != refId {
				fmt.Printf("\nJournal for %s (Ledger ID: %d):\n", refId, ledgerId)
				currentRef = refId
			}
			
			if debitAmount > 0 {
				fmt.Printf("  Dr. %s (%s): %.0f - %s\n", 
					accountCode, accountName, debitAmount, description)
			}
			if creditAmount > 0 {
				fmt.Printf("  Cr. %s (%s): %.0f - %s\n", 
					accountCode, accountName, creditAmount, description)
			}
		}
	}

	// 5. Count total journals
	fmt.Println("\n5. JOURNAL SUMMARY:")
	var totalLedger, totalLines int
	
	err = db.QueryRow("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&totalLedger)
	if err == nil {
		fmt.Printf("Total ledger entries: %d\n", totalLedger)
	}
	
	err = db.QueryRow("SELECT COUNT(*) FROM unified_journal_lines").Scan(&totalLines)
	if err == nil {
		fmt.Printf("Total journal lines: %d\n", totalLines)
	}

	// 6. Check balance per journal
	fmt.Println("\n6. BALANCE VERIFICATION:")
	rows, err = db.Query(`
		SELECT l.source_code, l.total_debit, l.total_credit,
			   l.total_debit - l.total_credit as difference
		FROM unified_journal_ledger l
		WHERE l.source_type = 'purchase' OR l.source_code LIKE 'PO%'
		ORDER BY l.source_code
	`)
	if err != nil {
		fmt.Printf("Error checking balance: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var refId string
			var totalDebit, totalCredit, difference float64
			
			err := rows.Scan(&refId, &totalDebit, &totalCredit, &difference)
			if err != nil {
				continue
			}
			
			fmt.Printf("%s: Dr=%.0f, Cr=%.0f", refId, totalDebit, totalCredit)
			if difference == 0 {
				fmt.Printf(" ✅ BALANCED\n")
			} else {
				fmt.Printf(" ❌ NOT BALANCED (Diff: %.0f)\n", difference)
			}
		}
	}

	fmt.Println("\n=== SSOT JOURNAL CHECK SELESAI ===")
}