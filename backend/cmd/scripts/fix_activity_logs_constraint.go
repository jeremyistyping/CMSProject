package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔧 Fixing activity_logs user_id constraint...")

	// Initialize database connection
	db := database.ConnectDB()

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get SQL database: %v", err)
	}
	defer sqlDB.Close()

	// Check current constraint
	fmt.Println("\n📋 Checking current constraint...")
	var isNullable string
	err = db.Raw(`
		SELECT is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'activity_logs' 
		  AND column_name = 'user_id'
	`).Scan(&isNullable).Error
	
	if err != nil {
		log.Fatalf("❌ Failed to check constraint: %v", err)
	}
	
	fmt.Printf("Current user_id is_nullable: %s\n", isNullable)

	if isNullable == "NO" {
		// Make user_id nullable
		fmt.Println("\n🔄 Making user_id nullable...")
		err = db.Exec("ALTER TABLE activity_logs ALTER COLUMN user_id DROP NOT NULL").Error
		if err != nil {
			log.Fatalf("❌ Failed to alter table: %v", err)
		}
		fmt.Println("✅ Successfully made user_id nullable")
	} else {
		fmt.Println("✅ user_id is already nullable")
	}

	// Verify the change
	fmt.Println("\n🔍 Verifying changes...")
	type ColumnInfo struct {
		ColumnName  string
		IsNullable  string
		DataType    string
	}
	
	var columnInfo ColumnInfo
	err = db.Raw(`
		SELECT column_name, is_nullable, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'activity_logs' 
		  AND column_name = 'user_id'
	`).Scan(&columnInfo).Error
	
	if err != nil {
		log.Fatalf("❌ Failed to verify: %v", err)
	}
	
	fmt.Printf("Column: %s, Nullable: %s, Type: %s\n", 
		columnInfo.ColumnName, columnInfo.IsNullable, columnInfo.DataType)

	// Test anonymous user insert
	fmt.Println("\n🧪 Testing anonymous user log insert...")
	err = db.Exec(`
		INSERT INTO activity_logs 
		(user_id, username, role, method, path, action, resource, status_code, ip_address, duration, created_at)
		VALUES 
		(NULL, 'anonymous', 'guest', 'GET', '/test', 'test', 'test', 200, '127.0.0.1', 0, NOW())
	`).Error
	
	if err != nil {
		log.Printf("⚠️  Test insert failed: %v", err)
	} else {
		fmt.Println("✅ Test insert successful")
		
		// Clean up test data
		db.Exec("DELETE FROM activity_logs WHERE username = 'anonymous' AND path = '/test'")
	}

	fmt.Println("\n✅ Migration completed successfully!")
}
