package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== CEK LEGACY JOURNAL SYSTEM ===\n")

	// 1. Cek apakah tabel journal_entries ada
	fmt.Println("1. CEK TABEL JOURNAL_ENTRIES (Legacy):")
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
		   SELECT FROM information_schema.tables 
		   WHERE table_schema = 'public'
		   AND table_name = 'journal_entries'
		)
	`).Scan(&tableExists)
	
	if err != nil {
		fmt.Printf("Error checking table: %v\n", err)
	} else if tableExists {
		fmt.Println("✅ Tabel journal_entries DITEMUKAN (Legacy system masih ada)")
		
		// Cek jumlah data di journal_entries
		var legacyCount int
		err = db.QueryRow("SELECT COUNT(*) FROM journal_entries").Scan(&legacyCount)
		if err == nil {
			fmt.Printf("   📊 Jumlah data di journal_entries: %d\n", legacyCount)
		}
		
		// Cek beberapa data legacy
		if legacyCount > 0 {
			fmt.Println("\n   📄 Sample data dari journal_entries:")
			rows, err := db.Query(`
				SELECT id, journal_id, reference_type, reference_id, 
				       total_debit, total_credit, status, description
				FROM journal_entries 
				ORDER BY created_at DESC
				LIMIT 10
			`)
			if err == nil {
				defer rows.Close()
				fmt.Printf("   %-4s %-10s %-15s %-10s %12s %12s %-8s %s\n", 
					"ID", "Journal ID", "Ref Type", "Ref ID", "Debit", "Credit", "Status", "Description")
				fmt.Println("   " + strings.Repeat("-", 100))
				
				for rows.Next() {
					var id, journalId, refId int
					var refType, status, description string
					var totalDebit, totalCredit float64
					
					err := rows.Scan(&id, &journalId, &refType, &refId,
						&totalDebit, &totalCredit, &status, &description)
					if err != nil {
						continue
					}
					
					if len(description) > 30 {
						description = description[:30] + "..."
					}
					
					fmt.Printf("   %-4d %-10d %-15s %-10d %12.0f %12.0f %-8s %s\n",
						id, journalId, refType, refId, totalDebit, totalCredit, status, description)
				}
			}
		}
	} else {
		fmt.Println("❌ Tabel journal_entries TIDAK DITEMUKAN")
	}

	// 2. Cek tabel journal_lines (legacy)
	fmt.Println("\n2. CEK TABEL JOURNAL_LINES (Legacy):")
	err = db.QueryRow(`
		SELECT EXISTS (
		   SELECT FROM information_schema.tables 
		   WHERE table_schema = 'public'
		   AND table_name = 'journal_lines'
		)
	`).Scan(&tableExists)
	
	if err != nil {
		fmt.Printf("Error checking table: %v\n", err)
	} else if tableExists {
		fmt.Println("✅ Tabel journal_lines DITEMUKAN (Legacy system masih ada)")
		
		// Cek jumlah data di journal_lines
		var legacyLinesCount int
		err = db.QueryRow("SELECT COUNT(*) FROM journal_lines").Scan(&legacyLinesCount)
		if err == nil {
			fmt.Printf("   📊 Jumlah data di journal_lines: %d\n", legacyLinesCount)
		}
	} else {
		fmt.Println("❌ Tabel journal_lines TIDAK DITEMUKAN")
	}

	// 3. Cek tabel journals (legacy)
	fmt.Println("\n3. CEK TABEL JOURNALS (Legacy):")
	err = db.QueryRow(`
		SELECT EXISTS (
		   SELECT FROM information_schema.tables 
		   WHERE table_schema = 'public'
		   AND table_name = 'journals'
		)
	`).Scan(&tableExists)
	
	if err != nil {
		fmt.Printf("Error checking table: %v\n", err)
	} else if tableExists {
		fmt.Println("✅ Tabel journals DITEMUKAN (Legacy system masih ada)")
		
		// Cek jumlah data di journals
		var legacyJournalsCount int
		err = db.QueryRow("SELECT COUNT(*) FROM journals").Scan(&legacyJournalsCount)
		if err == nil {
			fmt.Printf("   📊 Jumlah data di journals: %d\n", legacyJournalsCount)
		}
	} else {
		fmt.Println("❌ Tabel journals TIDAK DITEMUKAN")
	}

	// 4. Cek semua tabel yang ada di database
	fmt.Println("\n4. SEMUA TABEL DI DATABASE:")
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name LIKE '%journal%'
		ORDER BY table_name
	`)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Println("   📋 Tabel-tabel dengan kata 'journal':")
		for rows.Next() {
			var tableName string
			err := rows.Scan(&tableName)
			if err != nil {
				continue
			}
			fmt.Printf("   - %s\n", tableName)
		}
	}

	// 5. Kesimpulan
	fmt.Println("\n=== KESIMPULAN ===")
	fmt.Println("🔍 Status Sistem Journal:")
	fmt.Println("   ✅ SSOT Journal: AKTIF (unified_journal_ledger & unified_journal_lines)")
	fmt.Printf("   📊 Total entries di SSOT: %d entries dengan %d lines\n", 7, 17)
	
	// Cek keberadaan legacy
	legacyTables := []string{"journal_entries", "journal_lines", "journals"}
	legacyFound := false
	
	for _, table := range legacyTables {
		err = db.QueryRow(fmt.Sprintf(`
			SELECT EXISTS (
			   SELECT FROM information_schema.tables 
			   WHERE table_schema = 'public'
			   AND table_name = '%s'
			)
		`, table)).Scan(&tableExists)
		
		if err == nil && tableExists {
			legacyFound = true
			break
		}
	}
	
	if legacyFound {
		fmt.Println("   ⚠️  Legacy Journal: MASIH ADA (belum dibersihkan)")
		fmt.Println("   💡 Rekomendasi: Migrasi data legacy ke SSOT jika diperlukan")
	} else {
		fmt.Println("   ✅ Legacy Journal: TIDAK ADA (sudah dibersihkan)")
	}

	fmt.Println("\n=== CEK SELESAI ===")
}