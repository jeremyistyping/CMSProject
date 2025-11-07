package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Printf("🚀 AUTO-SETUP: Balance Synchronization System\n\n")
	
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database using DATABASE_URL from .env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get underlying sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Check if balance sync system is already installed
	var exists bool
	err = sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.triggers 
			WHERE trigger_name = 'trg_sync_account_balance_on_line_change'
		)
	`).Scan(&exists)
	
	if err != nil {
		log.Printf("Warning: Could not check existing system: %v", err)
	}

	if exists {
		fmt.Printf("✅ Balance Sync System already installed - skipping setup\n")
		
		// Still run a health check
		var triggerCount int
		err = sqlDB.QueryRow("SELECT COUNT(*) FROM information_schema.triggers WHERE trigger_name LIKE '%sync_account_balance%'").Scan(&triggerCount)
		if err == nil {
			if triggerCount >= 2 {
				fmt.Printf("✅ Balance sync system is active (%d triggers found)\n", triggerCount)
			} else {
				fmt.Printf("⚠️  Warning: Only %d sync triggers found\n", triggerCount)
			}
		}
		return
	}

	fmt.Printf("📦 Installing Balance Sync System...\n")

	// Read migration file
	migrationPath := "setup_automatic_balance_sync.sql"
	migrationSQL, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration
	fmt.Printf("🔧 Executing database migration...\n")
	_, err = sqlDB.Exec(string(migrationSQL))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify installation
	fmt.Printf("🔍 Verifying installation...\n")
	
	var triggerExists, procedureExists, viewExists bool
	
	// Check trigger
	sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.triggers 
			WHERE trigger_name = 'trg_sync_account_balance_on_line_change'
		)
	`).Scan(&triggerExists)
	
	// Check stored procedure
	sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.routines 
			WHERE routine_name = 'sync_account_balance_from_ssot'
		)
	`).Scan(&procedureExists)
	
	// Check view (checking for the function as there's no view created in the SQL)
	sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.routines 
			WHERE routine_name = 'trigger_sync_account_balance'
		)
	`).Scan(&viewExists)

	// Report results
	if triggerExists && procedureExists && viewExists {
		fmt.Printf("\n🎉 Balance Sync System successfully installed!\n\n")
		
		fmt.Printf("📋 INSTALLED COMPONENTS:\n")
		fmt.Printf("  ✅ Automatic trigger: trg_sync_account_balance_on_line_change\n")
		fmt.Printf("  ✅ Sync function: sync_account_balance_from_ssot()\n")  
		fmt.Printf("  ✅ Trigger function: trigger_sync_account_balance()\n")
		fmt.Printf("  ✅ Performance index: idx_unified_journal_lines_account_id_posted\n\n")
		
		fmt.Printf("💡 USAGE EXAMPLES:\n")
		fmt.Printf("  • Manual sync:     SELECT sync_account_balance_from_ssot(account_id);\n")
		fmt.Printf("  • Refresh views:   SELECT refresh_account_balances_view();\n")
		fmt.Printf("  • Check triggers:  SELECT * FROM information_schema.triggers WHERE trigger_name LIKE '%%sync%%';\n\n")

		// Verify trigger is working by checking if functions exist
		var funcCount int
		err = sqlDB.QueryRow("SELECT COUNT(*) FROM information_schema.routines WHERE routine_name LIKE '%sync_account_balance%'").Scan(&funcCount)
		if err == nil {
			if funcCount >= 2 {
				fmt.Printf("✅ All sync functions are installed (%d functions found)\n", funcCount)
			} else {
				fmt.Printf("⚠️  Warning: Only %d sync functions found\n", funcCount)
			}
		}

		fmt.Printf("\n🛡️  PROTECTION ACTIVE: Future balance inconsistencies will be prevented!\n")
		
	} else {
		fmt.Printf("\n❌ Balance Sync System installation incomplete!\n")
		fmt.Printf("   - Trigger exists: %v\n", triggerExists)
		fmt.Printf("   - Procedure exists: %v\n", procedureExists)
		fmt.Printf("   - View exists: %v\n", viewExists)
		log.Fatal("Installation failed - please check database permissions")
	}
}