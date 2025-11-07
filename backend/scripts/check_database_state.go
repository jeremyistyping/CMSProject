package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔍 Checking Database State")
	fmt.Println("==========================")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully")

	// Check views
	fmt.Println("\n1. Checking views...")
	rows, err := db.Raw(`
		SELECT table_name, table_type 
		FROM information_schema.tables 
		WHERE table_name = 'account_balances'
		OR table_schema = 'public' AND table_name LIKE '%account%balance%'
	`).Rows()
	
	if err != nil {
		log.Printf("Error checking views: %v", err)
	} else {
		defer rows.Close()
		fmt.Println("Views/Tables related to account_balances:")
		for rows.Next() {
			var tableName, tableType string
			rows.Scan(&tableName, &tableType)
			fmt.Printf("  - %s (%s)\n", tableName, tableType)
		}
	}

	// Check materialized views
	fmt.Println("\n2. Checking materialized views...")
	rows2, err := db.Raw(`SELECT matviewname FROM pg_matviews WHERE matviewname LIKE '%balance%'`).Rows()
	
	if err != nil {
		log.Printf("Error checking materialized views: %v", err)
	} else {
		defer rows2.Close()
		fmt.Println("Materialized views with 'balance':")
		for rows2.Next() {
			var matviewname string
			rows2.Scan(&matviewname)
			fmt.Printf("  - %s\n", matviewname)
		}
	}

	// Check SSOT tables
	fmt.Println("\n3. Checking SSOT tables...")
	ssotTables := []string{"unified_journal_ledger", "unified_journal_lines", "journal_event_log"}
	for _, table := range ssotTables {
		var exists bool
		err := db.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = ?
			)
		`, table).Scan(&exists).Error
		
		if err != nil {
			log.Printf("Error checking table %s: %v", table, err)
		} else {
			if exists {
				fmt.Printf("  ✅ %s exists\n", table)
			} else {
				fmt.Printf("  ❌ %s missing\n", table)
			}
		}
	}

	// Check migration logs
	fmt.Println("\n4. Checking migration status...")
	rows3, err := db.Raw(`
		SELECT migration_name, status, message 
		FROM migration_logs 
		WHERE migration_name LIKE '%ssot%' OR migration_name LIKE '%journal%'
		ORDER BY executed_at DESC
		LIMIT 5
	`).Rows()
	
	if err != nil {
		log.Printf("Error checking migration logs: %v", err)
	} else {
		defer rows3.Close()
		fmt.Println("Recent SSOT/Journal migrations:")
		for rows3.Next() {
			var migrationName, status, message string
			rows3.Scan(&migrationName, &status, &message)
			fmt.Printf("  - %s: %s (%s)\n", migrationName, status, message)
		}
	}

	fmt.Println("\n✅ Database state check completed!")
}