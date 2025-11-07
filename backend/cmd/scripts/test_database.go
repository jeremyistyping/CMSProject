package main

import (
	"log"
	
	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Testing Database Connection...")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	log.Println("✅ Database connected successfully")
	
	// Test query
	var result struct {
		Count int64
	}
	
	db.Raw("SELECT COUNT(*) as count FROM accounts").Scan(&result)
	log.Printf("📊 Found %d accounts in database", result.Count)
	
	// Check if security_incidents table exists
	var tableExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'security_incidents'
		)
	`).Scan(&tableExists)
	
	if tableExists {
		log.Println("✅ security_incidents table already exists")
	} else {
		log.Println("❌ security_incidents table does not exist")
	}
	
	// Check for account 1200
	var accountExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM accounts 
			WHERE code = '1200' AND deleted_at IS NULL
		)
	`).Scan(&accountExists)
	
	if accountExists {
		log.Println("✅ Account 1200 (ACCOUNTS RECEIVABLE) exists")
	} else {
		log.Println("❌ Account 1200 (ACCOUNTS RECEIVABLE) does not exist")
	}
	
	// Check database indexes
	var indexCount int64
	db.Raw(`
		SELECT COUNT(*) as count
		FROM pg_indexes 
		WHERE indexname LIKE 'idx_blacklisted_%' 
		   OR indexname LIKE 'idx_notifications_%'
	`).Scan(&indexCount)
	
	log.Printf("📈 Found %d performance indexes", indexCount)
	
	log.Println("✅ Database test completed!")
}
