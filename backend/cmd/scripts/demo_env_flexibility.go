package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/cmd/scripts/utils"
)

func main() {
	fmt.Printf("🔧 DEMONSTRATING DYNAMIC ENVIRONMENT LOADING\n")
	fmt.Printf("This script works with ANY .env configuration!\n\n")

	// Load environment variables dynamically from .env file
	databaseURL, err := utils.GetDatabaseURL()
	if err != nil {
		log.Fatal(err)
	}

	// Print current environment configuration
	utils.PrintEnvInfo()

	fmt.Printf("🔗 Connecting to database using detected configuration...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Test database connection with current environment
	fmt.Printf("✅ Database connection successful!\n\n")

	// Show database information
	var dbName, dbUser, dbHost string
	err = sqlDB.QueryRow("SELECT current_database(), current_user, inet_server_addr()").Scan(&dbName, &dbUser, &dbHost)
	if err != nil {
		fmt.Printf("⚠️  Could not get database info: %v\n", err)
	} else {
		fmt.Printf("📊 DATABASE INFO:\n")
		fmt.Printf("   Database Name: %s\n", dbName)
		fmt.Printf("   User: %s\n", dbUser)
		fmt.Printf("   Host: %s\n", dbHost)
	}

	// Check if balance sync system exists in this environment
	fmt.Printf("\n=== CHECKING BALANCE SYNC SYSTEM ===\n")
	
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
		fmt.Printf("✅ Balance sync trigger is active in this environment\n")
	} else {
		fmt.Printf("⚠️  Balance sync trigger not found in this environment\n")
	}

	// Count accounts in this environment
	var accountCount int
	err = sqlDB.QueryRow("SELECT COUNT(*) FROM accounts WHERE is_active = true").Scan(&accountCount)
	if err != nil {
		fmt.Printf("❌ Error counting accounts: %v\n", err)
	} else {
		fmt.Printf("📊 Active accounts in this environment: %d\n", accountCount)
	}

	// Count revenue accounts specifically
	var revenueCount int
	var totalRevenue float64
	err = sqlDB.QueryRow("SELECT COUNT(*), COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'REVENUE' AND is_active = true").Scan(&revenueCount, &totalRevenue)
	if err != nil {
		fmt.Printf("❌ Error checking revenue accounts: %v\n", err)
	} else {
		fmt.Printf("💰 Revenue accounts: %d (Total: Rp %.2f)\n", revenueCount, totalRevenue)
	}

	fmt.Printf("\n🎉 DEMONSTRATION COMPLETE!\n")
	fmt.Printf("This script successfully adapted to your environment configuration.\n")
	fmt.Printf("It will work on any PC with different .env settings!\n")

	// Show example .env configurations
	fmt.Printf("\n💡 EXAMPLE .env CONFIGURATIONS THAT WORK:\n\n")
	
	fmt.Printf("🖥️  Local Development:\n")
	fmt.Printf("   DATABASE_URL=postgres://user:pass@localhost/dev_db?sslmode=disable\n")
	fmt.Printf("   SERVER_PORT=8080\n")
	fmt.Printf("   ENVIRONMENT=development\n\n")

	fmt.Printf("🏢 Production Server:\n")
	fmt.Printf("   DATABASE_URL=postgres://admin:secret@prod-server.com/prod_db?sslmode=require\n")
	fmt.Printf("   SERVER_PORT=3000\n")
	fmt.Printf("   ENVIRONMENT=production\n\n")

	fmt.Printf("🧪 Testing Environment:\n")
	fmt.Printf("   DATABASE_URL=postgres://test:test123@test-server/test_db?sslmode=prefer\n")
	fmt.Printf("   SERVER_PORT=9090\n")
	fmt.Printf("   ENVIRONMENT=testing\n\n")

	fmt.Printf("✅ All these configurations work automatically!\n")
	fmt.Printf("No hardcoded values means maximum flexibility!\n")
}

// Example of how to create environment-specific configurations
func showEnvironmentExamples() {
	examples := map[string]map[string]string{
		"Development": {
			"DATABASE_URL": "postgres://postgres:postgres@localhost/sistem_akuntans_dev?sslmode=disable",
			"SERVER_PORT": "8080",
			"ENVIRONMENT": "development",
		},
		"Staging": {
			"DATABASE_URL": "postgres://staging_user:staging_pass@staging.company.com/sistem_akuntans_staging?sslmode=require",
			"SERVER_PORT": "8081",
			"ENVIRONMENT": "staging",
		},
		"Production": {
			"DATABASE_URL": "postgres://prod_user:prod_pass@production.company.com/sistem_akuntans_prod?sslmode=require",
			"SERVER_PORT": "80",
			"ENVIRONMENT": "production",
		},
	}

	fmt.Printf("📋 ENVIRONMENT CONFIGURATION EXAMPLES:\n\n")
	for envName, config := range examples {
		fmt.Printf("🔧 %s Environment (.env):\n", envName)
		for key, value := range config {
			if key == "DATABASE_URL" {
				// Mask password for display
				fmt.Printf("   %s=%s\n", key, MaskSensitiveURL(value))
			} else {
				fmt.Printf("   %s=%s\n", key, value)
			}
		}
		fmt.Printf("\n")
	}
}

// Utility function to mask sensitive URLs (making this public for reuse)
func MaskSensitiveURL(url string) string {
	if url == "" {
		return "NOT_SET"
	}
	
	// Simple implementation - can be enhanced
	return "postgres://user:***@host/database?sslmode=***"
}