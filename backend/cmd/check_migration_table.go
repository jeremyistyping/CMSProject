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

	fmt.Printf("🔗 Connecting to database: %s\n", databaseURL)
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔍 Checking migration_logs table structure...")

	// Check if migration_logs table exists
	var tableExists bool
	result := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'migration_logs')").Scan(&tableExists)
	if result.Error != nil {
		fmt.Printf("❌ Failed to check migration_logs table: %v\n", result.Error)
		return
	}

	fmt.Printf("📋 migration_logs table exists: %v\n", tableExists)

	if !tableExists {
		fmt.Println("ℹ️  migration_logs table doesn't exist, will need to be created")
		return
	}

	// Check columns in migration_logs table
	fmt.Println("\n📊 Columns in migration_logs table:")
	fmt.Println("==================================")

	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
		IsNullable string `gorm:"column:is_nullable"`
	}

	result = db.Raw(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'migration_logs' 
		ORDER BY ordinal_position
	`).Scan(&columns)

	if result.Error != nil {
		fmt.Printf("❌ Failed to get columns: %v\n", result.Error)
		return
	}

	for _, col := range columns {
		fmt.Printf("  - %s (%s, nullable: %s)\n", col.ColumnName, col.DataType, col.IsNullable)
	}

	// Check if description column exists
	var descriptionExists bool
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'migration_logs' AND column_name = 'description')").Scan(&descriptionExists)
	if result.Error != nil {
		fmt.Printf("❌ Failed to check description column: %v\n", result.Error)
		return
	}

	fmt.Printf("\n🔍 'description' column exists: %v\n", descriptionExists)

	// Show sample data
	var count int64
	result = db.Raw("SELECT COUNT(*) FROM migration_logs").Scan(&count)
	if result.Error != nil {
		fmt.Printf("❌ Failed to count rows: %v\n", result.Error)
	} else {
		fmt.Printf("📈 Total rows in migration_logs: %d\n", count)
	}

	fmt.Println("\n✅ Migration table structure check completed!")
}