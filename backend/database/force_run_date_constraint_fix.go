package database

import (
	"log"
	"gorm.io/gorm"
)

// ForceRunDateConstraintFix forces the date constraint fix to run
// Call this function if you need to manually trigger the fix
func ForceRunDateConstraintFix(db *gorm.DB) {
	log.Println("🔧 [MANUAL] Force running date constraint fix...")
	
	// Delete existing migration record if exists
	db.Exec("DELETE FROM migration_records WHERE migration_id = 'fix_journal_entry_date_constraint_v1.0'")
	log.Println("✅ Cleared existing migration record")
	
	// Run the fix
	FixJournalEntryDateConstraintMigration(db)
	
	log.Println("✅ [MANUAL] Date constraint fix completed")
}
