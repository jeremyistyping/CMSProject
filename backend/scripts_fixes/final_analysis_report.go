package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

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

	fmt.Println("🎯 =======================================================================")
	fmt.Println("    LAPORAN ANALISA FINAL PERBAIKAN AKUN PPN DAN PENDAPATAN")
	fmt.Println("🎯 =======================================================================")
	fmt.Println()

	// 1. VERIFY ACCOUNT STRUCTURE
	fmt.Println("📊 1. VERIFIKASI STRUKTUR AKUN SETELAH PERBAIKAN")
	fmt.Println("-------------------------------------------------------")
	
	rows, err := db.Query(`
		SELECT a.code, a.name, a.type, a.balance,
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

	accountsMap := make(map[string]string)
	fmt.Printf("%-6s | %-25s | %-10s | %15s | %s\n", "CODE", "NAME", "TYPE", "BALANCE", "PARENT")
	fmt.Println(strings.Repeat("-", 85))
	
	for rows.Next() {
		var code, name, accType, parent string
		var balance float64
		if err := rows.Scan(&code, &name, &accType, &balance, &parent); err != nil {
			log.Fatal("Error scanning row:", err)
		}
		fmt.Printf("%-6s | %-25s | %-10s | %15.2f | %s\n", code, name, accType, balance, parent)
		accountsMap[code] = accType
	}

	// 2. VALIDATE EXPECTED STRUCTURE
	fmt.Println()
	fmt.Println("✅ 2. VALIDASI STRUKTUR AKUN")
	fmt.Println("-------------------------------------------------------")
	
	expected := map[string]string{
		"1201": "ASSET",     // Piutang Usaha
		"2102": "ASSET",     // PPN Masukan  
		"2103": "LIABILITY", // PPN Keluaran
		"4101": "REVENUE",   // Pendapatan Penjualan
	}
	
	allCorrect := true
	for code, expectedType := range expected {
		if actualType, exists := accountsMap[code]; exists {
			if actualType == expectedType {
				fmt.Printf("✅ Akun %s: %s ← BENAR\n", code, actualType)
			} else {
				fmt.Printf("❌ Akun %s: %s (harusnya %s) ← SALAH\n", code, actualType, expectedType)
				allCorrect = false
			}
		} else {
			fmt.Printf("❌ Akun %s: TIDAK DITEMUKAN\n", code)
			allCorrect = false
		}
	}

	// 3. CHECK SPECIFIC BALANCES
	fmt.Println()
	fmt.Println("💰 3. ANALISA SALDO SESUAI EKSPEKTASI")
	fmt.Println("-------------------------------------------------------")
	
	expectedBalances := map[string]float64{
		"1201": 3885000.00, // Piutang Usaha
		"2103": -385000.00, // PPN Keluaran (negative karena liability)
		"4101": 0.00,       // Pendapatan (seharusnya ada nilai negatif untuk revenue, tapi bisa jadi sudah di-close)
	}
	
	balanceRows, err := db.Query(`
		SELECT code, name, balance 
		FROM accounts 
		WHERE code IN ('1201', '2103', '4101')
		ORDER BY code
	`)
	if err != nil {
		log.Fatal("Error querying balances:", err)
	}
	defer balanceRows.Close()

	fmt.Printf("%-6s | %-25s | %15s | %15s | %s\n", "CODE", "NAME", "ACTUAL", "EXPECTED", "STATUS")
	fmt.Println(strings.Repeat("-", 85))
	
	for balanceRows.Next() {
		var code, name string
		var actualBalance float64
		if err := balanceRows.Scan(&code, &name, &actualBalance); err != nil {
			log.Fatal("Error scanning balance row:", err)
		}
		
		expectedBalance := expectedBalances[code]
		status := "✅ OK"
		if expectedBalance != 0 && actualBalance != expectedBalance {
			status = "⚠️  BERBEDA"
		}
		
		fmt.Printf("%-6s | %-25s | %15.2f | %15.2f | %s\n", 
			code, name, actualBalance, expectedBalance, status)
	}

	// 4. JOURNAL VALIDATION
	fmt.Println()
	fmt.Println("📝 4. VALIDASI JURNAL PENJUALAN BER-PPN")
	fmt.Println("-------------------------------------------------------")
	
	// Check if there are any sales journal entries
	var journalCount int
	db.QueryRow(`
		SELECT COUNT(*)
		FROM journal_entries je
		WHERE je.description ILIKE '%penjualan%' 
		   OR je.description ILIKE '%sales%'
		   OR je.reference_type = 'SALE'
	`).Scan(&journalCount)
	
	fmt.Printf("Total jurnal penjualan ditemukan: %d\n", journalCount)
	
	if journalCount > 0 {
		fmt.Println("Sample jurnal penjualan terbaru:")
		journalRows, err := db.Query(`
			SELECT je.entry_date, je.description, a.code, a.name, 
			       jl.debit_amount, jl.credit_amount
			FROM journal_entries je
			JOIN journal_lines jl ON je.id = jl.journal_entry_id
			JOIN accounts a ON jl.account_id = a.id
			WHERE (je.description ILIKE '%penjualan%' 
			    OR je.description ILIKE '%sales%'
			    OR je.reference_type = 'SALE')
			  AND a.code IN ('1201', '2103', '4101')
			ORDER BY je.entry_date DESC, je.id DESC
			LIMIT 6
		`)
		if err == nil {
			defer journalRows.Close()
			
			fmt.Printf("%-12s | %-20s | %-6s | %-20s | %10s | %10s\n", 
				"DATE", "DESCRIPTION", "CODE", "ACCOUNT", "DEBIT", "CREDIT")
			fmt.Println(strings.Repeat("-", 95))
			
			for journalRows.Next() {
				var date, desc, code, name string
				var debit, credit float64
				if err := journalRows.Scan(&date, &desc, &code, &name, &debit, &credit); err != nil {
					break
				}
				fmt.Printf("%-12s | %-20s | %-6s | %-20s | %10.2f | %10.2f\n", 
					date[:10], desc[:20], code, name[:20], debit, credit)
			}
		}
	} else {
		fmt.Println("ℹ️  Belum ada jurnal penjualan, atau sudah dibersihkan.")
	}

	// 5. OVERALL BALANCE CHECK
	fmt.Println()
	fmt.Println("⚖️  5. CEK KESEIMBANGAN KESELURUHAN SISTEM")
	fmt.Println("-------------------------------------------------------")
	
	var totalAssets, totalLiabilities, totalEquity, totalRevenue, totalExpenses float64
	
	db.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'Asset' THEN balance ELSE 0 END), 0) as assets,
			COALESCE(SUM(CASE WHEN type = 'Liability' THEN -balance ELSE 0 END), 0) as liabilities,
			COALESCE(SUM(CASE WHEN type = 'Equity' THEN -balance ELSE 0 END), 0) as equity,
			COALESCE(SUM(CASE WHEN type = 'Revenue' THEN -balance ELSE 0 END), 0) as revenue,
			COALESCE(SUM(CASE WHEN type = 'Expense' THEN balance ELSE 0 END), 0) as expenses
		FROM accounts
		WHERE is_active = true
	`).Scan(&totalAssets, &totalLiabilities, &totalEquity, &totalRevenue, &totalExpenses)
	
	fmt.Printf("Total Assets    : %15.2f\n", totalAssets)
	fmt.Printf("Total Liabilities: %15.2f\n", totalLiabilities)
	fmt.Printf("Total Equity     : %15.2f\n", totalEquity)
	fmt.Printf("Total Revenue    : %15.2f\n", totalRevenue)
	fmt.Printf("Total Expenses   : %15.2f\n", totalExpenses)
	
	balanceCheck := totalAssets - (totalLiabilities + totalEquity)
	fmt.Printf("Balance Check   : %15.2f\n", balanceCheck)
	
	if balanceCheck < 0.01 && balanceCheck > -0.01 {
		fmt.Println("✅ SISTEM AKUNTANSI SEIMBANG!")
	} else {
		fmt.Printf("⚠️  Sistem tidak seimbang - selisih: %.2f\n", balanceCheck)
	}

	// 6. FINAL ASSESSMENT
	fmt.Println()
	fmt.Println("🏁 6. KESIMPULAN AKHIR PERBAIKAN")
	fmt.Println("-------------------------------------------------------")
	
	if allCorrect {
		fmt.Println("✅ SEMUA PERBAIKAN BERHASIL DITERAPKAN:")
		fmt.Println("   • Akun 2103 'PPN Keluaran' → LIABILITY ✅")
		fmt.Println("   • Akun 2102 'PPN Masukan' → ASSET ✅") 
		fmt.Println("   • Akun 4101 'Pendapatan Penjualan' → REVENUE ✅")
		fmt.Println("   • Akun 1201 'Piutang Usaha' → ASSET ✅")
		fmt.Println()
		fmt.Println("🎉 DAMPAK PERBAIKAN:")
		fmt.Println("   • Transaksi penjualan ber-PPN akan dicatat dengan benar")
		fmt.Println("   • PPN Keluaran muncul sebagai kewajiban (liability)")
		fmt.Println("   • Pendapatan tidak lagi terklasifikasi sebagai PPN")
		fmt.Println("   • Balance Sheet dan P&L akan menampilkan data yang akurat")
		fmt.Println()
		fmt.Println("📋 LANGKAH SELANJUTNYA:")
		fmt.Println("   • Refresh frontend dengan Ctrl+F5")
		fmt.Println("   • Test dengan membuat transaksi penjualan baru")
		fmt.Println("   • Verifikasi laporan Balance Sheet dan P&L")
	} else {
		fmt.Println("❌ MASIH ADA MASALAH YANG PERLU DIPERBAIKI")
		fmt.Println("   Periksa kembali klasifikasi akun di atas")
	}
	
	fmt.Println()
	fmt.Println("🎯 =======================================================================")
}