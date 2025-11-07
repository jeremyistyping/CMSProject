package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

	log.Println("🔍 Checking account_balances materialized view status...")

	// Check if materialized view exists
	var viewExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_matviews 
			WHERE schemaname = 'public' AND matviewname = 'account_balances'
		)
	`).Scan(&viewExists)
	if err != nil {
		log.Fatalf("Failed to check materialized view existence: %v", err)
	}

	log.Printf("📊 Materialized view exists: %v", viewExists)

	if !viewExists {
		log.Println("❌ Materialized view not found. Creating it now...")
		if err := createMaterializedView(db); err != nil {
			log.Fatalf("Failed to create materialized view: %v", err)
		}
		log.Println("✅ Materialized view created successfully")
	} else {
		log.Println("✅ Materialized view exists")
		
		// Check if view has data
		var rowCount int
		err = db.QueryRow("SELECT COUNT(*) FROM account_balances").Scan(&rowCount)
		if err != nil {
			log.Printf("❌ Error querying materialized view: %v", err)
			log.Println("🔄 Refreshing materialized view...")
			if err := refreshMaterializedView(db); err != nil {
				log.Fatalf("Failed to refresh materialized view: %v", err)
			}
		} else {
			log.Printf("📊 Materialized view has %d rows", rowCount)
		}
	}

	// Final verification
	log.Println("🧪 Running final verification...")
	if err := verifyMaterializedView(db); err != nil {
		log.Fatalf("Final verification failed: %v", err)
	}

	log.Println("🎉 Account balances materialized view is working properly!")
}

func createMaterializedView(db *sql.DB) error {
	// Read the migration file
	migrationPath := filepath.Join(".", "migrations", "030_create_account_balances_materialized_view.sql")
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %v", err)
	}

	// Execute the migration
	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %v", err)
	}

	return nil
}

func refreshMaterializedView(db *sql.DB) error {
	_, err := db.Exec("REFRESH MATERIALIZED VIEW account_balances")
	return err
}

func verifyMaterializedView(db *sql.DB) error {
	// Check if materialized view exists and is accessible
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM account_balances").Scan(&count)
	if err != nil {
		return fmt.Errorf("materialized view is not accessible: %v", err)
	}

	log.Printf("✅ Materialized view verified with %d records", count)

	// Check if required columns exist
	rows, err := db.Query(`
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = 'account_balances' AND table_schema = 'public'
		ORDER BY column_name
	`)
	if err != nil {
		return fmt.Errorf("failed to check columns: %v", err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return fmt.Errorf("failed to scan column: %v", err)
		}
		columns = append(columns, column)
	}

	log.Printf("📋 Available columns: %v", columns)

	// Check for required columns
	requiredColumns := []string{
		"account_id", "account_code", "account_name", "account_type", 
		"current_balance", "calculated_balance", "balance_difference",
	}

	for _, reqCol := range requiredColumns {
		found := false
		for _, col := range columns {
			if col == reqCol {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required column '%s' not found in materialized view", reqCol)
		}
	}

	log.Println("✅ All required columns are present")
	return nil
}