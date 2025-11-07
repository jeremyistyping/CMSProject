package database

import (
	"gorm.io/gorm"
	"log"
)

// AddSharesFieldsToCompanyProfile adds SharesOutstanding and ParValuePerShare fields to company_profiles table
func AddSharesFieldsToCompanyProfile(db *gorm.DB) error {
	log.Println("Adding SharesOutstanding and ParValuePerShare fields to company_profiles table...")
	
	// Add SharesOutstanding column
	if !db.Migrator().HasColumn("company_profiles", "shares_outstanding") {
		err := db.Exec(`
			ALTER TABLE company_profiles 
			ADD COLUMN shares_outstanding DECIMAL(20,4) DEFAULT 0;
		`).Error
		if err != nil {
			log.Printf("Error adding shares_outstanding column: %v", err)
			return err
		}
		log.Println("Added shares_outstanding column")
	}
	
	// Add ParValuePerShare column
	if !db.Migrator().HasColumn("company_profiles", "par_value_per_share") {
		err := db.Exec(`
			ALTER TABLE company_profiles 
			ADD COLUMN par_value_per_share DECIMAL(15,4) DEFAULT 1000;
		`).Error
		if err != nil {
			log.Printf("Error adding par_value_per_share column: %v", err)
			return err
		}
		log.Println("Added par_value_per_share column")
	}
	
	log.Println("Successfully added shares fields to company_profiles table")
	return nil
}

// RunSharesFieldsMigration runs the migration to add shares fields
func RunSharesFieldsMigration(db *gorm.DB) error {
	return AddSharesFieldsToCompanyProfile(db)
}
