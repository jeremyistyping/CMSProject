package main

import (
	"database/sql"
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("🔍 Debugging account_balances materialized view...")

	// Check if materialized view exists in pg_matviews
	var matViewExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_matviews 
			WHERE schemaname = 'public' AND matviewname = 'account_balances'
		)
	`).Scan(&matViewExists)
	if err != nil {
		log.Printf("Error checking pg_matviews: %v", err)
	}
	log.Printf("📊 Materialized view in pg_matviews: %v", matViewExists)

	// Check in information_schema.views (views include materialized views)
	var viewExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.views
			WHERE table_schema = 'public' AND table_name = 'account_balances'
		)
	`).Scan(&viewExists)
	if err != nil {
		log.Printf("Error checking information_schema.views: %v", err)
	}
	log.Printf("📊 View in information_schema.views: %v", viewExists)

	// Try to query the view directly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM account_balances").Scan(&count)
	if err != nil {
		log.Printf("❌ Error querying account_balances: %v", err)
		
		// Check if it's just a regular table
		err2 := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'account_balances'
			)
		`).Scan(&viewExists)
		if err2 == nil && viewExists {
			log.Printf("🔍 account_balances exists as regular table")
		}
	} else {
		log.Printf("✅ Successfully queried account_balances: %d rows", count)
	}

	// List all columns from the view/table
	rows, err := db.Query(`
		SELECT column_name, data_type
		FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = 'account_balances'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Printf("❌ Error getting columns: %v", err)
	} else {
		log.Println("📋 Columns in account_balances:")
		for rows.Next() {
			var colName, dataType string
			if err := rows.Scan(&colName, &dataType); err != nil {
				log.Printf("Error scanning column: %v", err)
				continue
			}
			fmt.Printf("  - %s (%s)\n", colName, dataType)
		}
		rows.Close()
	}

	// Show sample data
	log.Println("📊 Sample data:")
	sampleRows, err := db.Query("SELECT account_id, account_code, account_name FROM account_balances LIMIT 3")
	if err != nil {
		log.Printf("❌ Error getting sample data: %v", err)
	} else {
		for sampleRows.Next() {
			var id int64
			var code, name string
			if err := sampleRows.Scan(&id, &code, &name); err != nil {
				log.Printf("Error scanning sample: %v", err)
				continue
			}
			fmt.Printf("  - ID: %d, Code: %s, Name: %s\n", id, code, name)
		}
		sampleRows.Close()
	}
}