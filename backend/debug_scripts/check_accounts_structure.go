package main

import (
	"fmt"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔍 Checking Accounts Table Structure")
	fmt.Println("===================================")
	fmt.Println()

	db := database.ConnectDB()

	// Get column information
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}

	db.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = 'accounts' 
		ORDER BY ordinal_position
	`).Find(&columns)

	fmt.Println("📋 Accounts table columns:")
	for _, col := range columns {
		fmt.Printf("   - %s: %s\n", col.ColumnName, col.DataType)
	}

	// Also show sample data
	var sampleAccount map[string]interface{}
	db.Raw("SELECT * FROM accounts LIMIT 1").Find(&sampleAccount)

	fmt.Println("\n📝 Sample account data:")
	for key, value := range sampleAccount {
		fmt.Printf("   %s: %v\n", key, value)
	}
}