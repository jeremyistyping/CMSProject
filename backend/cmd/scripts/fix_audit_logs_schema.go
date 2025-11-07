package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Database connection
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	log.Println("🔧 Starting audit_logs schema fix...")

	// Read migration file
	sqlFile := "database/migrations/fix_audit_logs_schema.sql"
	sqlContent, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration
	if err := db.Exec(string(sqlContent)).Error; err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	log.Println("✅ Migration executed successfully!")

	// Verify changes
	type ColumnInfo struct {
		ColumnName            string
		DataType              string
		CharacterMaximumLength *int
		IsNullable            string
	}

	var columns []ColumnInfo
	if err := db.Raw(`
		SELECT 
			column_name, 
			data_type, 
			character_maximum_length,
			is_nullable
		FROM information_schema.columns 
		WHERE table_name = 'audit_logs' 
		AND column_name IN ('action', 'notes')
		ORDER BY ordinal_position
	`).Scan(&columns).Error; err != nil {
		log.Printf("⚠️ Warning: Failed to verify changes: %v", err)
		return
	}

	fmt.Println("\n📋 Verification Results:")
	fmt.Println("========================")
	for _, col := range columns {
		maxLen := "N/A"
		if col.CharacterMaximumLength != nil {
			maxLen = fmt.Sprintf("%d", *col.CharacterMaximumLength)
		}
		fmt.Printf("Column: %s | Type: %s | Max Length: %s | Nullable: %s\n", 
			col.ColumnName, col.DataType, maxLen, col.IsNullable)
	}

	fmt.Println("\n✅ Schema fix completed successfully!")
	fmt.Println("You can now restart your backend server.")
}
