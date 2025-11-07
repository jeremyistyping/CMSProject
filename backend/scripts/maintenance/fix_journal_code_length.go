package main

import (
	"log"
	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("=== Fix Journal Code Length Issue ===")
	
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
	
	// Check current journal table structure
	log.Println("Checking current journal table structure...")
	
	var columns []struct {
		ColumnName    string `gorm:"column:column_name"`
		DataType      string `gorm:"column:data_type"`
		CharMaxLength *int   `gorm:"column:character_maximum_length"`
	}
	
	err = db.Raw(`
		SELECT column_name, data_type, character_maximum_length 
		FROM information_schema.columns 
		WHERE table_name = 'journals' AND column_name = 'code'
	`).Scan(&columns).Error
	if err != nil {
		log.Printf("Warning: Failed to check journal table structure: %v", err)
	} else {
		for _, col := range columns {
			if col.CharMaxLength != nil {
				log.Printf("Current code column: %s (%s, max_length: %d)", col.ColumnName, col.DataType, *col.CharMaxLength)
			}
		}
	}
	
	// Fix journal code length constraint
	log.Println("Fixing journal code length constraint...")
	
	err = db.Exec("ALTER TABLE journals ALTER COLUMN code TYPE VARCHAR(30)").Error
	if err != nil {
		log.Printf("Failed to alter journal code column length: %v", err)
	} else {
		log.Println("✅ Successfully increased journal code column length to 30 characters")
	}
	
	// Check if revenue account exists
	log.Println("Checking if revenue account 4900 exists...")
	
	var count int64
	err = db.Raw("SELECT COUNT(*) FROM accounts WHERE code = '4900' AND type = 'REVENUE'").Scan(&count).Error
	if err != nil {
		log.Printf("Failed to check revenue account: %v", err)
	} else {
		if count == 0 {
			log.Println("Revenue account 4900 doesn't exist, creating it...")
			
			err = db.Exec(`
				INSERT INTO accounts (code, name, type, category, is_active, description, level, created_at, updated_at)
				VALUES ('4900', 'Other Income', 'REVENUE', 'OTHER_REVENUE', true, 'Auto-created account for cash/bank deposits', 2, NOW(), NOW())
				ON CONFLICT (code) DO NOTHING
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
	
	// Also check/create withdrawal expense account
	log.Println("Checking if expense account 5900 exists...")
	
	err = db.Raw("SELECT COUNT(*) FROM accounts WHERE code = '5900' AND type = 'EXPENSE'").Scan(&count).Error
	if err != nil {
		log.Printf("Failed to check expense account: %v", err)
	} else {
		if count == 0 {
			log.Println("Expense account 5900 doesn't exist, creating it...")
			
			err = db.Exec(`
				INSERT INTO accounts (code, name, type, category, is_active, description, level, created_at, updated_at)
				VALUES ('5900', 'General Expense', 'EXPENSE', 'OTHER_EXPENSE', true, 'Auto-created account for cash/bank withdrawals', 2, NOW(), NOW())
				ON CONFLICT (code) DO NOTHING
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
	
	log.Println("=== Fix completed successfully! ===")
	log.Println("The journal code length has been increased and required accounts have been created.")
	log.Println("You can now process deposits and withdrawals without the character limit error.")
}
