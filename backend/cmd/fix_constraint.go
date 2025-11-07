package main

import (
	"log"
	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load config from app config
	cfg := config.LoadConfig()

	// Connect to database using same config as main app
	log.Printf("Connecting to database: %s", cfg.DatabaseURL)
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")

	// 1. Check current constraint
	log.Println("\n1️⃣  Checking current constraint...")
	var constraintDef string
	err = db.Raw(`
		SELECT pg_get_constraintdef(oid) as constraint_definition
		FROM pg_constraint 
		WHERE conname = 'chk_journal_entries_date_valid'
		  AND conrelid = 'journal_entries'::regclass
	`).Scan(&constraintDef).Error
	
	if err != nil {
		log.Printf("⚠️  Warning: Could not read current constraint: %v", err)
	} else {
		log.Printf("Current constraint: %s", constraintDef)
	}

	// 2. Drop existing constraint
	log.Println("\n2️⃣  Dropping existing constraint...")
	err = db.Exec(`
		ALTER TABLE journal_entries 
		DROP CONSTRAINT IF EXISTS chk_journal_entries_date_valid
	`).Error
	
	if err != nil {
		log.Fatalf("❌ Failed to drop constraint: %v", err)
	}
	log.Println("✅ Constraint dropped")

	// 3. Create new relaxed constraint
	log.Println("\n3️⃣  Creating new relaxed constraint...")
	err = db.Exec(`
		ALTER TABLE journal_entries 
		ADD CONSTRAINT chk_journal_entries_date_valid 
		CHECK (entry_date >= '2000-01-01'::date AND entry_date <= '2099-12-31'::date)
	`).Error
	
	if err != nil {
		log.Fatalf("❌ Failed to create new constraint: %v", err)
	}
	log.Println("✅ New constraint created")

	// 4. Verify new constraint
	log.Println("\n4️⃣  Verifying new constraint...")
	err = db.Raw(`
		SELECT pg_get_constraintdef(oid) as constraint_definition
		FROM pg_constraint 
		WHERE conname = 'chk_journal_entries_date_valid'
		  AND conrelid = 'journal_entries'::regclass
	`).Scan(&constraintDef).Error
	
	if err != nil {
		log.Fatalf("❌ Failed to verify new constraint: %v", err)
	}
	
	log.Printf("✅ New constraint verified: %s", constraintDef)
	log.Println("\n🎉 SUCCESS! Constraint has been updated.")
	log.Println("📅 Journal entries can now have entry_date from 2000-01-01 to 2099-12-31")
	log.Println("✅ Period closing for year 2026 should now work!")
	log.Println("")
	log.Println("⚠️  IMPORTANT: Restart your backend server to apply changes!")
}
