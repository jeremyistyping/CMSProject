package database

import (
	"log"
	"time"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

// FixJournalEntryDateConstraintMigration fixes the overly restrictive date validation constraint
// that prevents period closing for future dates
func FixJournalEntryDateConstraintMigration(db *gorm.DB) {
	migrationID := "fix_journal_entry_date_constraint_v1.0"
	
	// Check if this migration has already been applied
	var existing models.MigrationRecord
	err := db.Where("migration_id = ?", migrationID).First(&existing).Error
	if err == nil {
		log.Printf("Migration '%s' already applied at %v, skipping...", migrationID, existing.AppliedAt)
		return
	}
	
	log.Println("Running journal entry date constraint fix migration...")
	log.Println("Issue: Constraint limits entry_date to CURRENT_DATE + 1 year")
	log.Println("Fix: Relax constraint to allow dates up to year 2099 for period closing")
	
	// Check if journal_entries table exists
	var tableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'journal_entries'
	)`).Scan(&tableExists)
	
	if !tableExists {
		log.Println("journal_entries table does not exist yet, skipping constraint fix")
		return
	}
	
	// Drop and recreate the constraint with relaxed validation
	err = db.Exec(`
		ALTER TABLE journal_entries 
		DROP CONSTRAINT IF EXISTS chk_journal_entries_date_valid,
		ADD CONSTRAINT chk_journal_entries_date_valid 
		CHECK (entry_date >= '2000-01-01' AND entry_date <= '2099-12-31')
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update date constraint: %v", err)
		return
	}
	
	log.Println("✅ Updated chk_journal_entries_date_valid constraint")
	
	// Verify the new constraint
	var constraintDef string
	err = db.Raw(`
		SELECT pg_get_constraintdef(oid) as constraint_definition
		FROM pg_constraint 
		WHERE conname = 'chk_journal_entries_date_valid'
	`).Scan(&constraintDef).Error
	
	if err == nil {
		log.Printf("✅ New constraint definition: %s", constraintDef)
	}
	
	// Record this migration as completed
	migrationRecord := models.MigrationRecord{
		MigrationID: migrationID,
		Description: "Fix journal entry date constraint to allow future period closing (up to year 2099)",
		Version:     "1.0",
		AppliedAt:   time.Now(),
	}
	
	if err := db.Create(&migrationRecord).Error; err != nil {
		log.Printf("Warning: Failed to record migration completion: %v", err)
	} else {
		log.Println("✅ Journal entry date constraint fix migration completed successfully")
	}
}
