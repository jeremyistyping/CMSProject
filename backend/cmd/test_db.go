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

	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Attempting to connect to database...\n")
	fmt.Printf("📍 Database URL: %s\n", databaseURL)

	// Test database connection
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Printf("❌ Failed to connect to database: %v", err)
		
		// Try to provide helpful suggestions
		fmt.Printf("\n💡 Troubleshooting suggestions:\n")
		fmt.Printf("1. Make sure PostgreSQL is running\n")
		fmt.Printf("2. Check if the database exists\n")
		fmt.Printf("3. Verify the username and password in .env\n")
		fmt.Printf("4. Try creating the postgres user if it doesn't exist\n")
		fmt.Printf("5. Check if the database name matches your .env file\n")
		
		return
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("❌ Failed to get database instance: %v", err)
		return
	}
	defer sqlDB.Close()

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		log.Printf("❌ Failed to ping database: %v", err)
		return
	}

	fmt.Printf("✅ Database connection successful!\n")

	// Test basic query
	var version string
	if err := db.Raw("SELECT version()").Scan(&version).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to get PostgreSQL version: %v", err)
	} else {
		fmt.Printf("🗄️  PostgreSQL Version: %s\n", version)
	}

	// Check if database exists and show some info
	var dbName string
	if err := db.Raw("SELECT current_database()").Scan(&dbName).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to get database name: %v", err)
	} else {
		fmt.Printf("📊 Connected to database: %s\n", dbName)
	}

	// Check current user
	var currentUser string
	if err := db.Raw("SELECT current_user").Scan(&currentUser).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to get current user: %v", err)
	} else {
		fmt.Printf("👤 Connected as user: %s\n", currentUser)
	}

	// List some tables to verify database content
	var tableCount int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount).Error; err != nil {
		log.Printf("⚠️  Warning: Failed to count tables: %v", err)
	} else {
		fmt.Printf("📋 Number of tables in public schema: %d\n", tableCount)
	}

	fmt.Printf("\n🎉 Database connection test completed successfully!\n")
}