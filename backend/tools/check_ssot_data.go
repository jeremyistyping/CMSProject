package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== CEK DATA SSOT JOURNAL SYSTEM ===\n")

	// 1. Cek semua data di unified_journal_ledger
	fmt.Println("1. DATA DI UNIFIED_JOURNAL_LEDGER:")
	rows, err := db.Query(`
		SELECT id, entry_number, source_type, source_code, entry_date,
		       total_debit, total_credit, status, description
		FROM unified_journal_ledger 
		ORDER BY created_at DESC
		LIMIT 15
	`)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Printf("%-4s %-15s %-12s %-15s %-12s %12s %12s %-8s %s\n", 
			"ID", "Entry Number", "Source Type", "Source Code", "Date", "Debit", "Credit", "Status", "Description")
		fmt.Println(strings.Repeat("-", 120))
		
		for rows.Next() {
			var id int
			var entryNumber, sourceType, sourceCode, status, description string
			var entryDate time.Time
			var totalDebit, totalCredit float64
			
			err := rows.Scan(&id, &entryNumber, &sourceType, &sourceCode, &entryDate,
				&totalDebit, &totalCredit, &status, &description)
			if err != nil {
				fmt.Printf("Error scanning: %v\n", err)
				continue
			}
			
			// Batasi deskripsi agar tidak terlalu panjang
			if len(description) > 40 {
				description = description[:40] + "..."
			}
			
			fmt.Printf("%-4d %-15s %-12s %-15s %-12s %12.0f %12.0f %-8s %s\n",
				id, entryNumber, sourceType, sourceCode, entryDate.Format("2006-01-02"), 
				totalDebit, totalCredit, status, description)
		}
	}

	// 2. Cek data per source_type
	fmt.Println("\n2. RINGKASAN DATA PER SOURCE TYPE:")
	rows, err = db.Query(`
		SELECT source_type, COUNT(*) as count, 
		       SUM(total_debit) as total_debit,
		       SUM(total_credit) as total_credit
		FROM unified_journal_ledger 
		GROUP BY source_type
		ORDER BY source_type
	`)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Printf("%-15s %8s %15s %15s\n", "Source Type", "Count", "Total Debit", "Total Credit")
		fmt.Println(strings.Repeat("-", 60))
		
		for rows.Next() {
			var sourceType string
			var count int
			var totalDebit, totalCredit float64
			
			err := rows.Scan(&sourceType, &count, &totalDebit, &totalCredit)
			if err != nil {
				continue
			}
			
			fmt.Printf("%-15s %8d %15.0f %15.0f\n", sourceType, count, totalDebit, totalCredit)
		}
	}

	// 3. Cek journal lines untuk satu entry
	fmt.Println("\n3. DETAIL LINES UNTUK JOURNAL PERTAMA:")
	rows, err = db.Query(`
		SELECT ujl.line_number, a.code as account_code, a.name as account_name,
		       ujl.debit_amount, ujl.credit_amount, ujl.description
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE ujl.journal_id = (SELECT id FROM unified_journal_ledger ORDER BY id LIMIT 1)
		ORDER BY ujl.line_number
	`)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Printf("%-4s %-10s %-25s %12s %12s %s\n", 
			"Line", "Acc Code", "Account Name", "Debit", "Credit", "Description")
		fmt.Println(strings.Repeat("-", 80))
		
		for rows.Next() {
			var lineNumber int
			var accountCode, accountName, lineDesc string
			var debitAmount, creditAmount float64
			
			err := rows.Scan(&lineNumber, &accountCode, &accountName, 
				&debitAmount, &creditAmount, &lineDesc)
			if err != nil {
				continue
			}
			
			// Batasi nama akun agar tidak terlalu panjang
			if len(accountName) > 25 {
				accountName = accountName[:22] + "..."
			}
			
			fmt.Printf("%-4d %-10s %-25s %12.0f %12.0f %s\n",
				lineNumber, accountCode, accountName, debitAmount, creditAmount, lineDesc)
		}
	}

	// 4. Cek status balanced
	fmt.Println("\n4. STATUS BALANCE:")
	rows, err = db.Query(`
		SELECT 
		    COUNT(*) as total_entries,
		    COUNT(*) FILTER (WHERE is_balanced = true) as balanced_entries,
		    COUNT(*) FILTER (WHERE status = 'POSTED') as posted_entries,
		    COUNT(*) FILTER (WHERE status = 'DRAFT') as draft_entries
		FROM unified_journal_ledger
	`)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var total, balanced, posted, draft int
			err := rows.Scan(&total, &balanced, &posted, &draft)
			if err != nil {
				continue
			}
			
			fmt.Printf("Total Entries: %d\n", total)
			fmt.Printf("Balanced Entries: %d\n", balanced)
			fmt.Printf("Posted Entries: %d\n", posted)
			fmt.Printf("Draft Entries: %d\n", draft)
		}
	}

	fmt.Println("\n=== CEK SELESAI ===")
}