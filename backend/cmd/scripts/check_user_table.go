package main

import (
	"fmt"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("📊 Checking Users Table Structure")
	fmt.Println("=================================")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully\n")

	// Get users table structure
	var columns []struct {
		ColumnName string
		DataType   string
		IsNullable string
	}
	
	err := db.Raw(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'users' 
		ORDER BY ordinal_position
	`).Scan(&columns).Error
	
	if err != nil {
		fmt.Printf("❌ Error checking users table: %v\n", err)
		return
	}
	
	fmt.Println("📋 Users Table Structure:")
	for _, col := range columns {
		fmt.Printf("   %-20s %s %s\n", col.ColumnName, col.DataType, col.IsNullable)
	}
	
	// Sample some users
	var users []struct {
		ID    uint64
		Email string
		Role  string
	}
	
	db.Raw(`
		SELECT id, email, role
		FROM users 
		ORDER BY id 
		LIMIT 10
	`).Scan(&users)
	
	fmt.Println("\n📄 Sample Users:")
	if len(users) == 0 {
		fmt.Println("   No users found in database")
	}
	for _, user := range users {
		fmt.Printf("   ID: %d - %s (%s)\n", user.ID, user.Email, user.Role)
	}
}