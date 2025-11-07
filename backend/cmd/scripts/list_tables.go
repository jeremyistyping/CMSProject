package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database using DATABASE_URL from .env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get underlying sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Query all tables
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name;
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	fmt.Println("=== ALL TABLES IN DATABASE ===")
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		fmt.Printf("- %s\n", tableName)
	}

	// Now look specifically for journal-related tables
	fmt.Println("\n=== JOURNAL-RELATED TABLES ===")
	journalQuery := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name LIKE '%journal%'
		ORDER BY table_name;
	`

	journalRows, err := sqlDB.Query(journalQuery)
	if err != nil {
		log.Fatalf("Failed to query journal tables: %v", err)
	}
	defer journalRows.Close()

	for journalRows.Next() {
		var tableName string
		err := journalRows.Scan(&tableName)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		fmt.Printf("- %s\n", tableName)
		
		// Get column info for journal tables
		columnQuery := `
			SELECT column_name, data_type 
			FROM information_schema.columns 
			WHERE table_name = $1 
			ORDER BY ordinal_position;
		`
		
		columnRows, err := sqlDB.Query(columnQuery, tableName)
		if err != nil {
			log.Printf("Failed to get columns for %s: %v", tableName, err)
			continue
		}
		
		fmt.Printf("  Columns:\n")
		for columnRows.Next() {
			var columnName, dataType string
			err := columnRows.Scan(&columnName, &dataType)
			if err != nil {
				log.Printf("Error scanning column: %v", err)
				continue
			}
			fmt.Printf("    - %s (%s)\n", columnName, dataType)
		}
		columnRows.Close()
	}
}