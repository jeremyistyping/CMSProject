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
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	fmt.Println("🔍 SYSTEM STATUS VERIFICATION")
	fmt.Println("=============================")

	// Check trigger
	var triggerExists bool
	err = sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.triggers 
			WHERE trigger_name = 'balance_sync_trigger'
		)
	`).Scan(&triggerExists)
	
	if triggerExists {
		fmt.Println("✅ Balance sync trigger: ACTIVE")
	} else {
		fmt.Println("❌ Balance sync trigger: MISSING")
	}

	// Check stored procedure
	var procedureExists bool
	err = sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.routines 
			WHERE routine_name = 'sync_account_balances'
		)
	`).Scan(&procedureExists)
	
	if procedureExists {
		fmt.Println("✅ Manual sync function: AVAILABLE")
	} else {
		fmt.Println("❌ Manual sync function: MISSING")
	}

	// Check monitoring view
	var viewExists bool
	err = sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.views 
			WHERE table_name = 'account_balance_monitoring'
		)
	`).Scan(&viewExists)
	
	if viewExists {
		fmt.Println("✅ Monitoring view: CREATED")
	} else {
		fmt.Println("❌ Monitoring view: MISSING")
	}

	// Check balance health
	if viewExists {
		var mismatchCount int
		err = sqlDB.QueryRow("SELECT COUNT(*) FROM account_balance_monitoring WHERE status='MISMATCH'").Scan(&mismatchCount)
		if err == nil {
			if mismatchCount == 0 {
				fmt.Println("✅ Balance health: ALL SYNCHRONIZED")
			} else {
				fmt.Printf("⚠️  Balance health: %d MISMATCHES FOUND\n", mismatchCount)
			}
		}
	}

	// Overall status
	fmt.Println("\n📊 OVERALL STATUS:")
	if triggerExists && procedureExists && viewExists {
		fmt.Println("🛡️  BALANCE PROTECTION: FULLY ACTIVE")
		fmt.Println("✅ System is protected against balance mismatches")
	} else {
		fmt.Println("⚠️  BALANCE PROTECTION: INCOMPLETE")
		fmt.Println("❌ System needs setup to prevent balance issues")
	}
}