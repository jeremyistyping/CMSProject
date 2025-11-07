package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Printf("🧪 TESTING BALANCE SYNC MIGRATIONS...\n\n")
	
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

	// Test each migration file by running it
	migrations := []string{
		"balance_sync_system.sql",
		"balance_sync_system_fixed.sql", 
		"032_balance_sync_system_v2_fixed.sql",
	}

	for _, migration := range migrations {
		fmt.Printf("\n🧪 Testing migration: %s\n", migration)
		
		// Read the migration file
		content, err := os.ReadFile("migrations/" + migration)
		if err != nil {
			fmt.Printf("❌ Failed to read %s: %v\n", migration, err)
			continue
		}

		// Execute the migration
		_, err = sqlDB.Exec(string(content))
		if err != nil {
			fmt.Printf("❌ Migration %s failed: %v\n", migration, err)
		} else {
			fmt.Printf("✅ Migration %s executed successfully\n", migration)
		}
	}

	// Check if balance sync system is working
	fmt.Printf("\n=== VERIFYING BALANCE SYNC SYSTEM ===\n")
	
	var triggerExists bool
	err = sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.triggers 
			WHERE trigger_name = 'balance_sync_trigger'
		)
	`).Scan(&triggerExists)
	
	if err != nil {
		fmt.Printf("❌ Error checking trigger: %v\n", err)
	} else if triggerExists {
		fmt.Printf("✅ Balance sync trigger is active\n")
	} else {
		fmt.Printf("⚠️  Balance sync trigger not found\n")
	}

	var functionExists bool
	err = sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.routines 
			WHERE routine_name = 'sync_account_balances'
		)
	`).Scan(&functionExists)
	
	if err != nil {
		fmt.Printf("❌ Error checking function: %v\n", err)
	} else if functionExists {
		fmt.Printf("✅ Manual sync function is available\n")
	} else {
		fmt.Printf("⚠️  Manual sync function not found\n")
	}

	// Check migration logs
	fmt.Printf("\n=== CHECKING MIGRATION LOGS ===\n")
	
	rows, err := sqlDB.Query(`
		SELECT migration_name, status, executed_at 
		FROM migration_logs 
		WHERE migration_name LIKE '%balance_sync%' 
		ORDER BY executed_at DESC
	`)
	if err != nil {
		fmt.Printf("❌ Error checking migration logs: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Printf("📋 Balance sync migration logs:\n")
		for rows.Next() {
			var name, status, executed string
			rows.Scan(&name, &status, &executed)
			fmt.Printf("   %s | %s | %s\n", name, status, executed[:19])
		}
	}

	fmt.Printf("\n✅ MIGRATION TEST COMPLETED\n")
}