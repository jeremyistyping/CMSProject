package main

import (
	"fmt"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔍 Checking Old Journal Table Structure")
	fmt.Println("======================================")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully")

	// Check journal_entries table structure
	fmt.Println("\n📋 journal_entries table structure:")
	rows, err := db.Raw(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = 'journal_entries' 
		ORDER BY ordinal_position
	`).Rows()
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer rows.Close()
	
	fmt.Println("Column Name | Data Type | Nullable | Default")
	fmt.Println("------------|-----------|----------|--------")
	for rows.Next() {
		var colName, dataType, nullable, defaultVal string
		rows.Scan(&colName, &dataType, &nullable, &defaultVal)
		if defaultVal == "" {
			defaultVal = "NULL"
		}
		fmt.Printf("%-15s | %-15s | %-8s | %s\n", colName, dataType, nullable, defaultVal)
	}

	// Sample data from journal_entries
	fmt.Println("\n📊 Sample journal_entries data:")
	rows2, err := db.Raw(`
		SELECT id, description, created_at, total_debit, total_credit, status 
		FROM journal_entries 
		ORDER BY created_at DESC 
		LIMIT 3
	`).Rows()
	
	if err != nil {
		fmt.Printf("Error getting sample data: %v\n", err)
		return
	}
	defer rows2.Close()
	
	fmt.Println("ID | Description | Date | Debit | Credit | Status")
	fmt.Println("---|-------------|------|-------|--------|-------")
	for rows2.Next() {
		var id uint64
		var desc string
		var createdAt, totalDebit, totalCredit, status interface{}
		rows2.Scan(&id, &desc, &createdAt, &totalDebit, &totalCredit, &status)
		fmt.Printf("%d | %s | %v | %v | %v | %v\n", id, desc, createdAt, totalDebit, totalCredit, status)
	}

	fmt.Println("\n✅ Table structure check completed!")
}