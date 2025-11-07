package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("❌ DATABASE_URL environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer db.Close()

	// Check materialized view using pg_matviews
	var viewExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_matviews 
			WHERE matviewname = 'account_balances'
		);
	`).Scan(&viewExists)
	
	if err != nil {
		log.Printf("⚠️  Failed to check materialized view: %v", err)
	} else {
		fmt.Printf("🔍 Materialized view 'account_balances' exists: %v\n", viewExists)
	}

	// Also check if it exists as a regular table/view
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'account_balances'
		);
	`).Scan(&tableExists)
	
	if err != nil {
		log.Printf("⚠️  Failed to check table: %v", err)
	} else {
		fmt.Printf("🔍 Table 'account_balances' exists: %v\n", tableExists)
	}

	// If it exists, get some info about it
	if tableExists {
		var rowCount int
		err = db.QueryRow(`SELECT COUNT(*) FROM account_balances;`).Scan(&rowCount)
		if err != nil {
			log.Printf("⚠️  Failed to count rows: %v", err)
		} else {
			fmt.Printf("📊 Row count in account_balances: %d\n", rowCount)
		}
	}
}