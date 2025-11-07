package database

import (
	"log"
	"gorm.io/gorm"
)

// AddAccountingPeriodDescription adds description column to accounting_periods table
func AddAccountingPeriodDescription(db *gorm.DB) error {
	log.Println("Checking accounting_periods table for description column...")

	// Check if accounting_periods table exists
	var tableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'accounting_periods'
	)`).Scan(&tableExists)

	if !tableExists {
		log.Println("accounting_periods table does not exist, skipping migration")
		return nil
	}

	// Check if description column already exists
	var columnExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'accounting_periods' AND column_name = 'description'
	)`).Scan(&columnExists)

	if columnExists {
		log.Println("description column already exists in accounting_periods table")
		return nil
	}

	// Add description column
	log.Println("Adding description column to accounting_periods table...")
	err := db.Exec(`
		ALTER TABLE accounting_periods 
		ADD COLUMN description TEXT;
	`).Error

	if err != nil {
		log.Printf("Error adding description column to accounting_periods table: %v", err)
		return err
	}

	log.Println("Successfully added description column to accounting_periods table")
	return nil
}
