package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// FixActivityLogsUserIDMigration ensures activity_logs.user_id is nullable for anonymous users
// This is idempotent and safe to run on every startup across environments
func FixActivityLogsUserIDMigration(db *gorm.DB) error {
	migrationName := "fix_activity_logs_user_id_nullable"
	
	// Check if already completed
	var count int64
	err := db.Raw(`
		SELECT COUNT(*) FROM migration_logs 
		WHERE migration_name = ? AND status = 'SUCCESS'
	`, migrationName).Scan(&count).Error
	
	if err == nil && count > 0 {
		log.Println("✅ Activity logs user_id fix already applied")
		return nil
	}
	
	log.Println("🔧 Fixing activity_logs.user_id constraint to allow NULL...")
	
	// Step 1: Check if user_id column is already nullable
	var isNullable string
	err = db.Raw(`
		SELECT is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'activity_logs' 
		AND column_name = 'user_id'
	`).Scan(&isNullable).Error
	
	if err != nil {
		return fmt.Errorf("failed to check user_id nullability: %w", err)
	}
	
	if isNullable == "YES" {
		log.Println("✅ user_id is already nullable - skipping")
		logMigrationResult(db, migrationName, "SUCCESS", "user_id already nullable", 0)
		return nil
	}
	
	log.Println("🔄 Making user_id nullable...")
	
	// Step 2: Drop existing foreign key constraint if exists
	err = db.Exec(`
		ALTER TABLE activity_logs 
		DROP CONSTRAINT IF EXISTS fk_activity_logs_user
	`).Error
	
	if err != nil {
		return fmt.Errorf("failed to drop foreign key constraint: %w", err)
	}
	
	// Step 3: Make user_id nullable
	err = db.Exec(`
		ALTER TABLE activity_logs 
		ALTER COLUMN user_id DROP NOT NULL
	`).Error
	
	if err != nil {
		return fmt.Errorf("failed to make user_id nullable: %w", err)
	}
	
	// Step 4: Re-add foreign key constraint that allows NULL values
	err = db.Exec(`
		ALTER TABLE activity_logs 
		ADD CONSTRAINT fk_activity_logs_user 
		FOREIGN KEY (user_id) 
		REFERENCES users(id) 
		ON DELETE CASCADE
	`).Error
	
	if err != nil {
		// Don't fail if FK already exists
		if isAlreadyExistsError(err) {
			log.Println("⚠️  Foreign key constraint already exists - continuing")
		} else {
			return fmt.Errorf("failed to add foreign key constraint: %w", err)
		}
	}
	
	// Step 5: Add comment
	err = db.Exec(`
		COMMENT ON COLUMN activity_logs.user_id IS 
		'ID of the user who performed the action (NULL for anonymous/unauthenticated users)'
	`).Error
	
	if err != nil {
		log.Printf("⚠️  Failed to add comment: %v", err)
		// Don't fail on comment errors
	}
	
	// Log success
	logMigrationResult(db, migrationName, "SUCCESS", "user_id made nullable for anonymous users", 0)
	
	log.Println("✅ Activity logs user_id constraint fixed successfully")
	log.Println("   ℹ️  Anonymous users can now be logged with user_id = NULL")
	
	return nil
}
