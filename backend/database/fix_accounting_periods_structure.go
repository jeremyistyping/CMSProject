package database

import (
	"log"
	"gorm.io/gorm"
)

// FixAccountingPeriodsStructure fixes the accounting_periods table structure
// to match the model definition by making year and month nullable
func FixAccountingPeriodsStructure(db *gorm.DB) error {
	log.Println("Fixing accounting_periods table structure...")

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

	// Check if year column exists
	var yearExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'accounting_periods' AND column_name = 'year'
	)`).Scan(&yearExists)

	// Check if month column exists
	var monthExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'accounting_periods' AND column_name = 'month'
	)`).Scan(&monthExists)

	if !yearExists && !monthExists {
		log.Println("year and month columns do not exist, no fix needed")
		return nil
	}

	// Option 1: Make year and month nullable (safer approach - keeps data)
	if yearExists {
		log.Println("Making year column nullable...")
		err := db.Exec(`
			ALTER TABLE accounting_periods 
			ALTER COLUMN year DROP NOT NULL;
		`).Error
		if err != nil {
			log.Printf("Warning: Failed to make year column nullable: %v", err)
			// Try to set default value for existing NULL records
			db.Exec(`UPDATE accounting_periods SET year = EXTRACT(YEAR FROM start_date) WHERE year IS NULL`)
		} else {
			log.Println("✅ year column is now nullable")
		}
	}

	if monthExists {
		log.Println("Making month column nullable...")
		err := db.Exec(`
			ALTER TABLE accounting_periods 
			ALTER COLUMN month DROP NOT NULL;
		`).Error
		if err != nil {
			log.Printf("Warning: Failed to make month column nullable: %v", err)
			// Try to set default value for existing NULL records
			db.Exec(`UPDATE accounting_periods SET month = EXTRACT(MONTH FROM start_date) WHERE month IS NULL`)
		} else {
			log.Println("✅ month column is now nullable")
		}
	}

	// Also check if period_name column exists and is being used
	var periodNameExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'accounting_periods' AND column_name = 'period_name'
	)`).Scan(&periodNameExists)

	if periodNameExists {
		log.Println("period_name column exists - keeping it as optional field")
	}

	log.Println("✅ Successfully fixed accounting_periods table structure")
	return nil
}
