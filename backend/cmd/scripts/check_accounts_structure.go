package main

import (
	"fmt"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("📊 Checking Accounts Table Structure")
	fmt.Println("====================================")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully\n")

	// Get accounts table structure
	var columns []struct {
		ColumnName string
		DataType   string
		IsNullable string
	}
	
	err := db.Raw(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'accounts' 
		ORDER BY ordinal_position
	`).Scan(&columns).Error
	
	if err != nil {
		fmt.Printf("❌ Error checking accounts table: %v\n", err)
		return
	}
	
	fmt.Println("📋 Accounts Table Structure:")
	for _, col := range columns {
		fmt.Printf("   %-20s %s %s\n", col.ColumnName, col.DataType, col.IsNullable)
	}
	
	// Sample some accounts
	var accounts []struct {
		ID   uint64
		Name string
		Type string
	}
	
	db.Raw(`
		SELECT id, name, 
		       COALESCE(type, 'UNKNOWN') as type
		FROM accounts 
		ORDER BY id 
		LIMIT 10
	`).Scan(&accounts)
	
	fmt.Println("\n📄 Sample Accounts:")
	for _, acc := range accounts {
		fmt.Printf("   ID: %d - %s (%s)\n", acc.ID, acc.Name, acc.Type)
	}
}