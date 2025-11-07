package main

import (
	"database/sql"
	"fmt"
	"log"
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

	asOf := "2025-09-25" // adjust if needed
	asOfDate, _ := time.Parse("2006-01-02", asOf)
	fmt.Println("=== DEBUG TRIAL BALANCE QUERY ===")
	fmt.Println("As Of:", asOfDate.Format("2006-01-02"))

	query := `
		SELECT 
			ujl.account_id,
			a.code AS account_code,
			a.name AS account_name,
			CASE 
				WHEN a.code LIKE '1%%' THEN 'Asset'
				WHEN a.code LIKE '2%%' THEN 'Liability'
				WHEN a.code LIKE '3%%' THEN 'Equity'
				WHEN a.code LIKE '4%%' THEN 'Revenue'
				WHEN a.code LIKE '5%%' THEN 'Expense'
				ELSE 'Other'
			END AS account_type,
			COALESCE(SUM(ujl.debit_amount), 0) AS total_debit,
			COALESCE(SUM(ujl.credit_amount), 0) AS total_credit
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		LEFT JOIN accounts a ON a.id = ujl.account_id
		WHERE uje.entry_date <= $1
			AND uje.status = 'POSTED'
		GROUP BY ujl.account_id, a.code, a.name
		ORDER BY a.code`

	rows, err := db.Query(query, asOfDate)
	if err != nil {
		log.Fatalf("Query error: %v", err)
	}
	defer rows.Close()

	fmt.Printf("%-8s %-8s %-30s %-10s %12s %12s\n", "AccID", "Code", "Name", "Type", "Debit", "Credit")
	fmt.Println("------------------------------------------------------------------------------------------")
	count := 0
	var totalDebit, totalCredit float64
	for rows.Next() {
		var accID int64
		var code, name, typ string
		var debit, credit float64
		if err := rows.Scan(&accID, &code, &name, &typ, &debit, &credit); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%-8d %-8s %-30s %-10s %12.0f %12.0f\n", accID, code, name, typ, debit, credit)
		count++
		totalDebit += debit
		totalCredit += credit
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("------------------------------------------------------------------------------------------")
	fmt.Printf("Rows: %d  TotalDebit: %.0f  TotalCredit: %.0f  Balanced: %v\n", count, totalDebit, totalCredit, totalDebit == totalCredit)
}
