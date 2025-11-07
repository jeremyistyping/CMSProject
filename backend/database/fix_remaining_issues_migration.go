package database

import (
	"app-sistem-akuntansi/models"
	"log"
	"time"

	"gorm.io/gorm"
)

// FixRemainingIssuesMigration fixes remaining issues after the main fixes
func FixRemainingIssuesMigration(db *gorm.DB) {
	log.Println("🔧 Starting remaining issues fix migration...")

	migrationID := "fix_remaining_issues_v2024.1"

	// Check if this migration has already been run
	var existingMigration models.MigrationRecord
	if err := db.Where("migration_id = ?", migrationID).First(&existingMigration).Error; err == nil {
		log.Printf("Remaining issues fix migration '%s' already applied at %v, skipping...", migrationID, existingMigration.AppliedAt)
		return
	}

	// Fix 1: Drop and recreate views with proper column selection
	fixAuditViews(db)

	// Fix 2: Update migration versions to latest
	updateMigrationVersions(db)

	// Record this migration as completed
	migrationRecord := models.MigrationRecord{
		MigrationID: migrationID,
		Description: "Fix remaining issues - audit views column duplication, migration version updates",
		Version:     "2024.1",
		AppliedAt:   time.Now(),
	}

	if err := db.Create(&migrationRecord).Error; err != nil {
		log.Printf("Warning: Failed to record remaining issues migration completion: %v", err)
	} else {
		log.Println("✅ Remaining issues fix migration recorded successfully")
	}

	log.Println("✅ Remaining issues fix migration completed successfully")
}

// fixAuditViews fixes the audit log views to prevent column duplication
func fixAuditViews(db *gorm.DB) {
	log.Println("Fixing audit log views to prevent column duplication...")

	// Drop existing views first
	db.Exec(`DROP VIEW IF EXISTS critical_audit_changes`)
	db.Exec(`DROP VIEW IF EXISTS audit_trail_summary`)

	// Recreate audit trail summary
	db.Exec(`
		CREATE OR REPLACE VIEW audit_trail_summary AS
		SELECT 
			DATE(created_at) as audit_date,
			table_name,
			action,
			user_id,
			COUNT(*) as action_count,
			MIN(created_at) as first_action,
			MAX(created_at) as last_action
		FROM audit_logs
		WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY DATE(created_at), table_name, action, user_id
		ORDER BY audit_date DESC, action_count DESC
	`)

	// Recreate critical audit changes with explicit column selection
	db.Exec(`
		CREATE OR REPLACE VIEW critical_audit_changes AS
		SELECT 
			al.id,
			al.table_name,
			al.record_id,
			al.action,
			al.old_values,
			al.new_values,
			al.user_id,
			al.created_at,
			al.updated_at,
			u.username as user_username,
			COALESCE(CONCAT(u.first_name, ' ', u.last_name), u.username, 'Unknown User') as full_name
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.action IN ('DELETE', 'UPDATE')
			AND al.table_name IN ('journal_entries', 'accounts', 'transactions', 'sales', 'purchases')
			AND al.created_at >= CURRENT_DATE - INTERVAL '7 days'
		ORDER BY al.created_at DESC
	`)

	log.Println("✅ Audit log views fixed successfully")
}

// updateMigrationVersions updates migration record versions to latest
func updateMigrationVersions(db *gorm.DB) {
	log.Println("Updating migration versions to latest...")

	// Update any migration records that might need version updates
	updates := []struct {
		ID      string
		Version string
	}{
		{"fix_missing_columns_v2024.1", "2024.1"},
		{"database_enhancements_v2024.1", "2024.1"},
		{"product_image_fix_v1.0", "1.0"},
	}

	for _, update := range updates {
		result := db.Model(&models.MigrationRecord{}).
			Where("migration_id = ?", update.ID).
			Update("version", update.Version)
		
		if result.Error != nil {
			log.Printf("Warning: Could not update version for migration %s: %v", update.ID, result.Error)
		} else if result.RowsAffected > 0 {
			log.Printf("✅ Updated version for migration %s to %s", update.ID, update.Version)
		}
	}

	log.Println("✅ Migration versions updated")
}