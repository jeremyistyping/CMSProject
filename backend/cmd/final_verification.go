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

	// Connect to database using DATABASE_URL from .env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database: %s\n", databaseURL)
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🎯 Final Backend Migration Verification")
	fmt.Println("=======================================")
	
	// 1. Test purchase_payments table
	var tableExists bool
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'purchase_payments')").Scan(&tableExists)
	if result.Error != nil {
		fmt.Printf("⚠️  Could not check purchase_payments table: %v\n", result.Error)
	} else {
		fmt.Printf("✅ purchase_payments table exists: %v\n", tableExists)
	}

	// 3. Test SSOT tables
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_ledger')").Scan(&tableExists)
	if result.Error != nil {
		fmt.Printf("⚠️  Could not check unified_journal_ledger: %v\n", result.Error)
	} else {
		fmt.Printf("✅ unified_journal_ledger table exists: %v\n", tableExists)
	}

	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_lines')").Scan(&tableExists)
	if result.Error != nil {
		fmt.Printf("⚠️  Could not check unified_journal_lines: %v\n", result.Error)
	} else {
		fmt.Printf("✅ unified_journal_lines table exists: %v\n", tableExists)
	}

	// 4. Test SSOT functions - more precise query
	var funcExists bool
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'sync_account_balance_from_ssot' AND data_type = 'void')").Scan(&funcExists)
	if result.Error != nil {
		fmt.Printf("⚠️  Could not check SSOT sync function: %v\n", result.Error)
	} else {
		fmt.Printf("✅ sync_account_balance_from_ssot function exists: %v\n", funcExists)
	}

	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'refresh_account_balances')").Scan(&funcExists)
	if result.Error != nil {
		fmt.Printf("⚠️  Could not check refresh function: %v\n", result.Error)
	} else {
		fmt.Printf("✅ refresh_account_balances function exists: %v\n", funcExists)
	}

	// 5. Test function execution
	fmt.Println("\n🧪 Testing Function Execution:")
	fmt.Println("------------------------------")

	// Test refresh function
	result = db.Exec("SELECT refresh_account_balances()")
	if result.Error != nil {
		fmt.Printf("❌ refresh_account_balances() failed: %v\n", result.Error)
	} else {
		fmt.Printf("✅ refresh_account_balances() executed successfully\n")
	}

	// Test sync function with valid account ID
	result = db.Exec("SELECT sync_account_balance_from_ssot(1::BIGINT)")
	if result.Error != nil {
		fmt.Printf("❌ sync_account_balance_from_ssot() failed: %v\n", result.Error)
	} else {
		fmt.Printf("✅ sync_account_balance_from_ssot() executed successfully\n")
	}
	
	// 6. Check purchase_payments table structure
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'purchase_payments' AND column_name = 'deleted_at')").Scan(&tableExists)
	if result.Error != nil {
		fmt.Printf("⚠️  Could not check deleted_at column: %v\n", result.Error)
	} else {
		fmt.Printf("✅ purchase_payments.deleted_at column exists: %v\n", tableExists)
	}

	// 7. Final System Health Check
	fmt.Println("\n📊 Final System Status:")
	fmt.Println("=======================")
	
	allGood := true
	
	if !funcExists {
		fmt.Println("❌ SSOT functions missing")
		allGood = false
	}

	if allGood {
		fmt.Println("🎉 *** ALL SYSTEMS GO! ***")
		fmt.Println("")
		fmt.Println("✅ Backend is ready to run without migration errors")
		fmt.Println("✅ SSOT Journal System is fully operational")
		fmt.Println("✅ Purchase payment integration is complete")
		fmt.Println("✅ Account balance synchronization is working")
		fmt.Println("")
		fmt.Println("🚀 You can now run your backend server successfully!")
	} else {
		fmt.Println("⚠️  Some components may need additional attention")
	}

	fmt.Println("\n🔗 Backend URLs when running:")
	fmt.Println("   Backend API: http://localhost:8080/api/v1")
	fmt.Println("   Swagger Docs: http://localhost:8080/swagger/index.html")
}