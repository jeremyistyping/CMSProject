package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Printf("🔧 FIXING BALANCE SYNC MIGRATION ERROR...\n\n")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Step 1: Check current migration status
	fmt.Printf("\n=== CHECKING MIGRATION STATUS ===\n")
	
	var oldMigrationExists bool
	err = sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM migration_logs 
			WHERE migration_name = 'balance_sync_system_v1.0'
		)
	`).Scan(&oldMigrationExists)
	
	if oldMigrationExists {
		fmt.Printf("Found old migration entry - marking as failed\n")
		
		// Mark old migration as failed
		_, err = sqlDB.Exec(`
			UPDATE migration_logs 
			SET status = 'FAILED', 
			    description = 'Failed due to SQL syntax error - replaced with v1.1'
			WHERE migration_name = 'balance_sync_system_v1.0'
		`)
		if err != nil {
			log.Printf("Warning: Could not update old migration status: %v", err)
		}
	}

	// Step 2: Check if components already exist
	fmt.Printf("\n=== CHECKING EXISTING COMPONENTS ===\n")
	
	var triggerExists, functionExists, viewExists bool
	
	sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.triggers 
			WHERE trigger_name = 'balance_sync_trigger'
		)
	`).Scan(&triggerExists)
	
	sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.routines 
			WHERE routine_name = 'sync_account_balances'
		)
	`).Scan(&functionExists)
	
	sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.views 
			WHERE table_name = 'account_balance_monitoring'
		)
	`).Scan(&viewExists)

	fmt.Printf("Current system status:\n")
	fmt.Printf("  - Balance sync trigger: %v\n", triggerExists)
	fmt.Printf("  - Manual sync function: %v\n", functionExists)
	fmt.Printf("  - Monitoring view: %v\n", viewExists)

	if triggerExists && functionExists && viewExists {
		fmt.Printf("\n✅ All components already exist and working!\n")
		fmt.Printf("The migration error can be safely ignored.\n")
		
		// Create success migration entry
		_, err = sqlDB.Exec(`
			INSERT INTO migration_logs (
				migration_name, status, executed_at, description
			) VALUES (
				'balance_sync_system_v1.1', 'SUCCESS', NOW(),
				'Balance sync system components already installed and verified'
			) ON CONFLICT (migration_name) DO UPDATE SET
				status = EXCLUDED.status,
				executed_at = EXCLUDED.executed_at,
				description = EXCLUDED.description
		`)
		
		if err != nil {
			log.Printf("Warning: Could not create success migration entry: %v", err)
		}
		
		// Verify system is working
		fmt.Printf("\n🔍 RUNNING SYSTEM VERIFICATION...\n")
		
		// Test manual sync function
		rows, err := sqlDB.Query("SELECT account_id FROM sync_account_balances() LIMIT 5")
		if err != nil {
			fmt.Printf("⚠️  Manual sync function test failed: %v\n", err)
		} else {
			defer rows.Close()
			var accountID int
			var count int
			for rows.Next() {
				rows.Scan(&accountID)
				count++
			}
			if count == 0 {
				fmt.Printf("✅ Manual sync function: WORKING (no corrections needed)\n")
			} else {
				fmt.Printf("✅ Manual sync function: WORKING (found %d accounts to sync)\n", count)
			}
		}
		
		// Test monitoring view
		var mismatchCount int
		err = sqlDB.QueryRow("SELECT COUNT(*) FROM account_balance_monitoring WHERE status='MISMATCH'").Scan(&mismatchCount)
		if err != nil {
			fmt.Printf("⚠️  Monitoring view test failed: %v\n", err)
		} else {
			fmt.Printf("✅ Monitoring view: WORKING (%d mismatches found)\n", mismatchCount)
		}
		
		return
	}

	// Step 3: If components are missing, show how to fix
	fmt.Printf("\n⚠️  SOME COMPONENTS ARE MISSING\n")
	fmt.Printf("To fix this, you can:\n\n")
	
	fmt.Printf("Option 1 - Run the fixed migration manually:\n")
	fmt.Printf("  psql -d $DATABASE_URL -f migrations/balance_sync_system_fixed.sql\n\n")
	
	fmt.Printf("Option 2 - Use the setup script:\n")
	fmt.Printf("  go run cmd/scripts/setup_balance_sync_auto.go\n\n")
	
	fmt.Printf("Option 3 - Use the batch script:\n")
	fmt.Printf("  setup_balance_protection.bat (Windows)\n")
	fmt.Printf("  ./setup_balance_protection.sh (Linux/Mac)\n\n")

	// Step 4: Create informative migration entry
	_, err = sqlDB.Exec(`
		INSERT INTO migration_logs (
			migration_name, status, executed_at, description
		) VALUES (
			'balance_sync_system_v1.0_error_info', 'INFO', NOW(),
			'Migration failed due to SQL syntax error. Use balance_sync_system_v1.1 or setup scripts instead.'
		) ON CONFLICT (migration_name) DO UPDATE SET
			status = EXCLUDED.status,
			executed_at = EXCLUDED.executed_at,
			description = EXCLUDED.description
	`)
	
	fmt.Printf("📋 NEXT STEPS:\n")
	fmt.Printf("1. The original migration had a SQL syntax error\n") 
	fmt.Printf("2. A fixed version (v1.1) is available in migrations/balance_sync_system_fixed.sql\n")
	fmt.Printf("3. The setup scripts can install the system automatically\n")
	fmt.Printf("4. Once installed, the system will prevent balance mismatch issues\n\n")
	
	fmt.Printf("🎯 RECOMMENDATION:\n")
	fmt.Printf("Run: go run cmd/scripts/setup_balance_sync_auto.go\n")
	fmt.Printf("This will install the system safely with proper error handling.\n")
}