package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("DATABASE TABLES LIST")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	type TableInfo struct {
		TableName string
	}

	var tables []TableInfo
	db.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`).Scan(&tables)

	fmt.Printf("Found %d tables:\n\n", len(tables))

	for i, table := range tables {
		fmt.Printf("%3d. %s\n", i+1, table.TableName)
	}
	fmt.Println()

	// Search for journal-related tables
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("JOURNAL-RELATED TABLES:")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	var journalTables []TableInfo
	db.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		  AND (table_name LIKE '%journal%' OR table_name LIKE '%ssot%')
		ORDER BY table_name
	`).Scan(&journalTables)

	for _, table := range journalTables {
		fmt.Printf("  - %s\n", table.TableName)
	}
	fmt.Println()
}

