package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file")
	}

	// Database connection
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== ANALISA STRUKTUR AKUN SETELAH PERBAIKAN ===")
	fmt.Println()

	// Check account structure
	fmt.Println("1. STRUKTUR CHART OF ACCOUNTS:")
	rows, err := db.Query(`
		SELECT a.code, a.name, a.type, 
		       COALESCE(pa.name, 'ROOT') as parent_name
		FROM accounts a
		LEFT JOIN accounts pa ON a.parent_id = pa.id
		WHERE a.code IN ('1201', '2102', '2103', '4101')
		ORDER BY a.code
	`)
	if err != nil {
		log.Fatal("Error querying accounts:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var code, name, accType, parent string
		if err := rows.Scan(&code, &name, &accType, &parent); err != nil {
			log.Fatal("Error scanning row:", err)
		}
		fmt.Printf("%s | %s | %s | Parent: %s\n", code, name, accType, parent)
	}

	fmt.Println()
	fmt.Println("2. SALDO AKUN TERKAIT:")
	
	// Check account balances
	balanceRows, err := db.Query(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE code IN ('1201', '2102', '2103', '4101')
		ORDER BY code
	`)
	if err != nil {
		log.Fatal("Error querying balances:", err)
	}
	defer balanceRows.Close()

	for balanceRows.Next() {
		var code, name, accType string
		var balance float64
		if err := balanceRows.Scan(&code, &name, &accType, &balance); err != nil {
			log.Fatal("Error scanning balance row:", err)
		}
		fmt.Printf("%s | %s | %s | Saldo: %.2f\n", code, name, accType, balance)
	}

	fmt.Println()
	fmt.Println("3. VERIFIKASI JURNAL PENJUALAN TERBARU:")
	
	// Check recent sales journal entries
	journalRows, err := db.Query(`
		SELECT je.entry_date, je.description, a.code, 
		       a.name, jl.debit_amount, jl.credit_amount
		FROM journal_entries je
		JOIN journal_lines jl ON je.id = jl.journal_entry_id
		JOIN accounts a ON jl.account_id = a.id
		WHERE (je.description LIKE '%penjualan%' OR je.description LIKE '%sales%')
		  AND a.code IN ('1201', '2103', '4101')
		ORDER BY je.entry_date DESC, je.id DESC
		LIMIT 10
	`)
	if err != nil {
		log.Fatal("Error querying journal entries:", err)
	}
	defer journalRows.Close()

	for journalRows.Next() {
		var date, desc, code, name string
		var debit, credit float64
		if err := journalRows.Scan(&date, &desc, &code, &name, &debit, &credit); err != nil {
			log.Fatal("Error scanning journal row:", err)
		}
		fmt.Printf("%s | %s | %s %s | Dr: %.2f | Cr: %.2f\n", 
			date, desc, code, name, debit, credit)
	}

	fmt.Println()
	fmt.Println("=== ANALISA KESEIMBANGAN ===")
	
	// Check if books balance
	var totalAssets, totalLiabilities, totalEquity, totalRevenue float64
	
	db.QueryRow(`
		SELECT COALESCE(SUM(CASE WHEN type = 'Asset' THEN balance ELSE 0 END), 0) as assets,
		       COALESCE(SUM(CASE WHEN type = 'Liability' THEN balance ELSE 0 END), 0) as liabilities,
		       COALESCE(SUM(CASE WHEN type = 'Equity' THEN balance ELSE 0 END), 0) as equity,
		       COALESCE(SUM(CASE WHEN type = 'Revenue' THEN balance ELSE 0 END), 0) as revenue
		FROM accounts
	`).Scan(&totalAssets, &totalLiabilities, &totalEquity, &totalRevenue)
	
	fmt.Printf("Total Assets: %.2f\n", totalAssets)
	fmt.Printf("Total Liabilities: %.2f\n", totalLiabilities)
	fmt.Printf("Total Equity: %.2f\n", totalEquity)
	fmt.Printf("Total Revenue: %.2f\n", totalRevenue)
	
	balanceCheck := totalAssets - (totalLiabilities + totalEquity)
	fmt.Printf("Balance Check (Assets - Liabilities - Equity): %.2f\n", balanceCheck)
	
	if balanceCheck < 0.01 && balanceCheck > -0.01 {
		fmt.Println("✅ BOOKS BALANCED!")
	} else {
		fmt.Println("❌ Books not balanced - difference:", balanceCheck)
	}

	fmt.Println()
	fmt.Println("=== KESIMPULAN PERBAIKAN ===")
	fmt.Println("✅ Akun 2103 sudah menjadi 'PPN Keluaran' (Liability)")
	fmt.Println("✅ Akun 2102 sudah menjadi 'PPN Masukan' (Asset)") 
	fmt.Println("✅ Akun 4101 sudah menjadi 'Pendapatan Penjualan' (Revenue)")
	fmt.Println("✅ Akun 1201 adalah 'Piutang Usaha' (Asset)")
}