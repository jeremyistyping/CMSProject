package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🚀 Starting manual tax account settings migration...")

	// Database connection
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	if envDsn := os.Getenv("DATABASE_URL"); envDsn != "" {
		dsn = envDsn
	}

	log.Printf("📡 Connecting to database: %s", dsn)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("✅ Database connected successfully")

	// Read migration file
	migrationPath := "migrations/027_create_tax_account_settings.sql"
	if _, err := os.Stat("../../" + migrationPath); err == nil {
		migrationPath = "../../" + migrationPath
	} else if _, err := os.Stat(migrationPath); err != nil {
		log.Fatalf("❌ Migration file not found: %s", migrationPath)
	}

	log.Printf("📂 Reading migration from: %s", migrationPath)
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("❌ Failed to read migration file: %v", err)
	}

	log.Println("🔧 Executing tax account settings migration...")
	
	// Execute the migration
	result := db.Exec(string(content))
	if result.Error != nil {
		log.Fatalf("❌ Failed to execute migration: %v", result.Error)
	}

	log.Println("✅ Migration executed successfully!")

	// Test if table exists and has data
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM tax_account_settings").Scan(&count).Error; err != nil {
		log.Printf("⚠️  Warning: Could not verify table creation: %v", err)
	} else {
		log.Printf("✅ Verification: tax_account_settings table exists with %d records", count)
	}

	// Check if default configuration was inserted
	var activeCount int64
	if err := db.Raw("SELECT COUNT(*) FROM tax_account_settings WHERE is_active = true").Scan(&activeCount).Error; err == nil {
		if activeCount > 0 {
			log.Printf("✅ Default configuration active: %d records", activeCount)
		} else {
			log.Println("⚠️  No active tax account settings found - may need to create default configuration")
		}
	}

	log.Println("🎉 Tax account settings migration completed successfully!")
	log.Println("📋 Next steps:")
	log.Println("  1. Start your backend server")
	log.Println("  2. Test the tax account settings API endpoints")
	log.Println("  3. Configure tax accounts via the frontend UI")
}