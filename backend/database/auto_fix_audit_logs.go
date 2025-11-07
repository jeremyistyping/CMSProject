package database

import (
	"log"

	"gorm.io/gorm"
)

// AutoFixAuditLogsSchema checks and fixes audit_logs schema issues
func AutoFixAuditLogsSchema(db *gorm.DB) error {
	log.Println("🔍 Checking audit_logs schema...")

	changesMade := false

	// Check if action column size is correct
	var actionSize int
	err := db.Raw(`
		SELECT character_maximum_length 
		FROM information_schema.columns 
		WHERE table_name = 'audit_logs' 
		AND column_name = 'action'
	`).Scan(&actionSize).Error

	if err != nil {
		log.Printf("⚠️  Failed to check action column size: %v", err)
		return err
	}

	// Fix action column size if needed
	if actionSize < 50 {
		log.Printf("🔧 Fixing action column size (current: %d, target: 50)...", actionSize)
		if err := db.Exec("ALTER TABLE audit_logs ALTER COLUMN action TYPE VARCHAR(50)").Error; err != nil {
			log.Printf("❌ Failed to fix action column: %v", err)
			return err
		}
		log.Println("✅ Action column size fixed to VARCHAR(50)")
		changesMade = true
	} else {
		log.Printf("✅ Action column size already correct (VARCHAR(%d))", actionSize)
	}

	// Check if notes column exists
	var notesExists bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = 'audit_logs' 
			AND column_name = 'notes'
		)
	`).Scan(&notesExists).Error

	if err != nil {
		log.Printf("⚠️  Failed to check notes column: %v", err)
		return err
	}

	// Add notes column if missing
	if !notesExists {
		log.Println("🔧 Adding notes column (required for SQL triggers)...")
		if err := db.Exec("ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS notes TEXT").Error; err != nil {
			log.Printf("❌ Failed to add notes column: %v", err)
			return err
		}
		log.Println("✅ Notes column added successfully")
		changesMade = true
	} else {
		log.Println("✅ Notes column already exists")
	}

	if changesMade {
		log.Println("")
		log.Println("🎉 audit_logs schema has been updated!")
		log.Println("   • This fix is automatically applied on every backend start")
		log.Println("   • Safe to run multiple times (idempotent)")
	} else {
		log.Println("")
		log.Println("✅ audit_logs schema is already up to date - no changes needed")
	}

	return nil
}
