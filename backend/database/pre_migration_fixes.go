package database

import (
	"log"

	"gorm.io/gorm"
)

// PreMigrationFixes handles issues that need to be resolved before core model migration
func PreMigrationFixes(db *gorm.DB) {
	log.Println("🔧 Running pre-migration fixes...")

	// Fix 1: Drop audit-related views that might conflict with audit_logs table migration
	dropAuditRelatedViews(db)

	// Fix 2: Handle any other pre-migration conflicts
	handleColumnTypeConflicts(db)

	log.Println("✅ Pre-migration fixes completed")
}

// dropAuditRelatedViews drops views that depend on audit_logs table
func dropAuditRelatedViews(db *gorm.DB) {
	log.Println("Dropping audit-related views that might conflict with migration...")

	// List of views that might depend on audit_logs table
	viewsToDrop := []string{
		"critical_audit_changes",
		"audit_trail_summary",
		"audit_logs_view",
		"recent_audit_activities",
		"user_audit_summary",
	}

	for _, view := range viewsToDrop {
		err := db.Exec("DROP VIEW IF EXISTS " + view).Error
		if err != nil {
			log.Printf("Note: Could not drop view %s: %v", view, err)
		} else {
			log.Printf("✅ Dropped view: %s", view)
		}
	}

	log.Println("✅ Audit-related views cleanup completed")
}

// handleColumnTypeConflicts handles column type conflicts before migration
func handleColumnTypeConflicts(db *gorm.DB) {
	log.Println("Handling column type conflicts...")

	// Check if audit_logs table exists
	var tableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'audit_logs'
	)`).Scan(&tableExists)

	if !tableExists {
		log.Println("audit_logs table doesn't exist yet, skipping column type fixes")
		return
	}

	// Check current action column size
	var columnInfo struct {
		CharacterMaximumLength *int `json:"character_maximum_length"`
		DataType               string `json:"data_type"`
	}

	err := db.Raw(`
		SELECT character_maximum_length, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'audit_logs' AND column_name = 'action'
	`).Scan(&columnInfo).Error

	if err != nil {
		log.Printf("Could not check action column info: %v", err)
		return
	}

	log.Printf("Current audit_logs.action column: type=%s, max_length=%v", 
		columnInfo.DataType, columnInfo.CharacterMaximumLength)

	// If column is too small, try to alter it (after dropping views)
	if columnInfo.CharacterMaximumLength != nil && *columnInfo.CharacterMaximumLength < 50 {
		log.Println("Expanding audit_logs.action column to accommodate larger values...")

		err = db.Exec(`ALTER TABLE audit_logs ALTER COLUMN action TYPE varchar(50)`).Error
		if err != nil {
			log.Printf("Could not expand action column (non-critical): %v", err)
		} else {
			log.Println("✅ Expanded audit_logs.action column to varchar(50)")
		}
	}

	log.Println("✅ Column type conflicts handled")
}