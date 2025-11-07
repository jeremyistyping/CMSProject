package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=sistem_akuntans_test sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("🔍 SALES TABLE STRUCTURE ANALYSIS")
	fmt.Println("=" + string(make([]byte, 50)) + "=")

	// 1. Check sales table structure
	fmt.Println("\n📊 SALES TABLE COLUMNS:")
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'sales' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Printf("Error querying sales table structure: %v", err)
		return
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var columnName, dataType, isNullable string
		err := rows.Scan(&columnName, &dataType, &isNullable)
		if err != nil {
			log.Printf("Error scanning column info: %v", err)
			continue
		}
		columns = append(columns, columnName)
		fmt.Printf("   %s: %s (%s)\n", columnName, dataType, isNullable)
	}

	// 2. Check current sales data with correct columns
	fmt.Println("\n📊 CURRENT SALES DATA:")
	
	// Build query with available columns
	baseQuery := "SELECT id, code"
	for _, col := range columns {
		switch col {
		case "invoice_number", "invoice_code":
			baseQuery += ", " + col
		case "contact_name", "customer_name", "customer":
			baseQuery += ", " + col
		case "total_amount", "grand_total", "total":
			baseQuery += ", " + col
		case "status":
			baseQuery += ", " + col
		case "created_at":
			baseQuery += ", " + col
		}
	}
	baseQuery += " FROM sales WHERE status IN ('INVOICED', 'PAID') ORDER BY created_at DESC LIMIT 10"

	fmt.Printf("Query: %s\n", baseQuery)
	
	rows, err = db.Query(baseQuery)
	if err != nil {
		log.Printf("Error querying sales: %v", err)
		
		// Try simpler query
		rows, err = db.Query("SELECT id, code, status, created_at FROM sales WHERE status IN ('INVOICED', 'PAID') LIMIT 5")
		if err != nil {
			log.Printf("Error with simple query: %v", err)
			return
		}
	}
	defer rows.Close()

	salesCount := 0
	for rows.Next() {
		salesCount++
		// Just count for now since column structure is unknown
		fmt.Printf("   Sales record %d found\n", salesCount)
		
		// Skip scanning to avoid errors
		var dummy interface{}
		for i := 0; i < len(columns) && i < 6; i++ {
			rows.Scan(&dummy)
		}
	}

	fmt.Printf("\n   Total INVOICED/PAID Sales: %d\n", salesCount)

	// 3. Check revenue account balance
	fmt.Println("\n💰 REVENUE ACCOUNT (4101) BALANCE:")
	var revenueBalance float64
	err = db.QueryRow("SELECT balance FROM accounts WHERE code = '4101'").Scan(&revenueBalance)
	if err != nil {
		log.Printf("Error getting revenue balance: %v", err)
	} else {
		fmt.Printf("   Account 4101 Balance: Rp %.0f\n", revenueBalance)
	}

	// 4. Check available journal tables
	fmt.Println("\n📚 AVAILABLE JOURNAL TABLES:")
	journalTables := []string{
		"journal_entries", 
		"unified_journal_ledger", 
		"ssot_journal_entries",
		"sales_journal_entries",
	}
	
	for _, table := range journalTables {
		var exists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = $1
			)
		`, table).Scan(&exists)
		
		if err != nil {
			fmt.Printf("   ❓ %s: Error checking\n", table)
		} else if exists {
			var count int
			db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			fmt.Printf("   ✅ %s: EXISTS (%d records)\n", table, count)
		} else {
			fmt.Printf("   ❌ %s: NOT FOUND\n", table)
		}
	}

	// 5. Diagnosis
	fmt.Println("\n📋 ISSUE DIAGNOSIS:")
	if salesCount > 0 && revenueBalance == 0 {
		fmt.Printf("   ❌ CONFIRMED ISSUE: %d INVOICED sales but revenue balance is 0\n", salesCount)
		fmt.Println("   🔧 Sales journal integration is NOT working")
	} else if revenueBalance > 0 {
		fmt.Printf("   ✅ Revenue balance exists: Rp %.0f\n", revenueBalance) 
	}
	
	fmt.Println("\n🎯 RECOMMENDED ACTION:")
	fmt.Println("   1. Fix sales service to auto-create journal entries")
	fmt.Println("   2. Implement sales accounting integration") 
	fmt.Println("   3. Process existing INVOICED sales retroactively")
}