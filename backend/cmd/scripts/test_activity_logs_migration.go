package main

import (
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	log.Println("==============================================")
	log.Println("TESTING ACTIVITY LOGS USER_ID MIGRATION")
	log.Println("==============================================")

	// Initialize database connection
	log.Println("Connecting to database...")
	db := database.ConnectDB()

	// Test migration
	log.Println("\n--- Running FixActivityLogsUserIDMigration ---")
	if err := database.FixActivityLogsUserIDMigration(db); err != nil {
		log.Printf("❌ Migration failed: %v", err)
	} else {
		log.Println("✅ Migration completed successfully")
	}

	// Verify result
	log.Println("\n--- Verifying Result ---")
	var isNullable string
	err := db.Raw(`
		SELECT is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'activity_logs' 
		AND column_name = 'user_id'
	`).Scan(&isNullable).Error

	if err != nil {
		log.Printf("❌ Failed to verify: %v", err)
	} else {
		log.Printf("✅ user_id is_nullable: %s", isNullable)
		if isNullable == "YES" {
			log.Println("✅ SUCCESS: user_id is now nullable")
		} else {
			log.Println("❌ FAILED: user_id is still NOT NULL")
		}
	}

	// Check migration logs
	log.Println("\n--- Checking Migration Logs ---")
	var count int64
	err = db.Raw(`
		SELECT COUNT(*) FROM migration_logs 
		WHERE migration_name = 'fix_activity_logs_user_id_nullable' 
		AND status = 'SUCCESS'
	`).Scan(&count).Error

	if err != nil {
		log.Printf("❌ Failed to check migration logs: %v", err)
	} else {
		log.Printf("✅ Migration log entries: %d", count)
	}

	// Test with anonymous user activity log
	log.Println("\n--- Testing Anonymous User Log ---")
	testSQL := `
		INSERT INTO activity_logs (
			user_id, username, role, method, path, 
			action, resource, status_code, ip_address, 
			user_agent, duration, description, metadata, 
			is_error, created_at
		) VALUES (
			NULL, 'anonymous', 'guest', 'GET', '/api/v1/test',
			'test_action', 'test_resource', 200, '127.0.0.1',
			'Test Agent', 100, 'Test log', '{}',
			false, NOW()
		)
	`

	err = db.Exec(testSQL).Error
	if err != nil {
		log.Printf("❌ Failed to insert test log: %v", err)
	} else {
		log.Println("✅ Successfully inserted activity log with user_id = NULL")
		
		// Cleanup test data
		db.Exec("DELETE FROM activity_logs WHERE username = 'anonymous' AND user_agent = 'Test Agent'")
		log.Println("✅ Test data cleaned up")
	}

	log.Println("\n==============================================")
	log.Println("MIGRATION TEST COMPLETED")
	log.Println("==============================================")
}
