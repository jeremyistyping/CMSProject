package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	
	// Check if accounting_periods table exists
	var tableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'accounting_periods'
	)`).Scan(&tableExists)
	
	if !tableExists {
		log.Println("accounting_periods table does not exist")
		return
	}
	
	log.Println("✅ accounting_periods table exists")
	
	// Get all columns from accounting_periods table
	type Column struct {
		ColumnName string
		DataType   string
		IsNullable string
	}
	
	var columns []Column
	err := db.Raw(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'accounting_periods'
		ORDER BY ordinal_position
	`).Scan(&columns).Error
	
	if err != nil {
		log.Fatalf("Failed to get columns: %v", err)
	}
	
	fmt.Println("\n=== accounting_periods table columns ===")
	descriptionFound := false
	for _, col := range columns {
		fmt.Printf("- %s (%s, nullable: %s)\n", col.ColumnName, col.DataType, col.IsNullable)
		if col.ColumnName == "description" {
			descriptionFound = true
		}
	}
	
	if descriptionFound {
		fmt.Println("\n✅ SUCCESS: 'description' column exists in accounting_periods table")
	} else {
		fmt.Println("\n❌ FAILED: 'description' column NOT FOUND in accounting_periods table")
	}
}
