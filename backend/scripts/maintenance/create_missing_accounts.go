package main

import (
	"log"
	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("=== Create Missing Accounts ===")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	log.Println("Connected to database successfully")
	
	// Check if revenue account 4900 exists
	log.Println("Checking if revenue account 4900 exists...")
	
	var count int64
	err = db.Raw("SELECT COUNT(*) FROM accounts WHERE code = '4900' AND type = 'REVENUE'").Scan(&count).Error
	if err != nil {
		log.Printf("Failed to check revenue account: %v", err)
	} else {
		if count == 0 {
			log.Println("Revenue account 4900 doesn't exist, creating it...")
			
			// Simple INSERT without ON CONFLICT
			err = db.Exec(`
				INSERT INTO accounts (code, name, type, category, is_active, description, level, created_at, updated_at)
				VALUES ('4900', 'Other Income', 'REVENUE', 'OTHER_REVENUE', true, 'Auto-created account for cash/bank deposits', 2, NOW(), NOW())
			`).Error
			if err != nil {
				log.Printf("Failed to create revenue account: %v", err)
			} else {
				log.Println("✅ Successfully created revenue account 4900")
			}
		} else {
			log.Println("✅ Revenue account 4900 already exists")
		}
	}
	
	// Check if expense account 5900 exists
	log.Println("Checking if expense account 5900 exists...")
	
	err = db.Raw("SELECT COUNT(*) FROM accounts WHERE code = '5900' AND type = 'EXPENSE'").Scan(&count).Error
	if err != nil {
		log.Printf("Failed to check expense account: %v", err)
	} else {
		if count == 0 {
			log.Println("Expense account 5900 doesn't exist, creating it...")
			
			// Simple INSERT without ON CONFLICT
			err = db.Exec(`
				INSERT INTO accounts (code, name, type, category, is_active, description, level, created_at, updated_at)
				VALUES ('5900', 'General Expense', 'EXPENSE', 'OTHER_EXPENSE', true, 'Auto-created account for cash/bank withdrawals', 2, NOW(), NOW())
			`).Error
			if err != nil {
				log.Printf("Failed to create expense account: %v", err)
			} else {
				log.Println("✅ Successfully created expense account 5900")
			}
		} else {
			log.Println("✅ Expense account 5900 already exists")
		}
	}
	
	log.Println("=== Account creation completed! ===")
}
