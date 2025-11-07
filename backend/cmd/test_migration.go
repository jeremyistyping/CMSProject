package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Connect to database
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntans_test port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test if account_balances materialized view exists
	var exists bool
	result := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'account_balances' AND table_type = 'VIEW')").Scan(&exists)
	if result.Error != nil {
		log.Fatal("Failed to check account_balances view:", result.Error)
	}

	fmt.Printf("✅ account_balances materialized view exists: %v\n", exists)

	// Test if purchase_payments table exists
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'purchase_payments')").Scan(&exists)
	if result.Error != nil {
		log.Fatal("Failed to check purchase_payments table:", result.Error)
	}

	fmt.Printf("✅ purchase_payments table exists: %v\n", exists)

	// Test unified journal system
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_ledger')").Scan(&exists)
	if result.Error != nil {
		log.Fatal("Failed to check unified_journal_ledger table:", result.Error)
	}

	fmt.Printf("✅ unified_journal_ledger table exists: %v\n", exists)

	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_lines')").Scan(&exists)
	if result.Error != nil {
		log.Fatal("Failed to check unified_journal_lines table:", result.Error)
	}

	fmt.Printf("✅ unified_journal_lines table exists: %v\n", exists)

	// Check SSOT functions
	var funcExists bool
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'sync_account_balance_from_ssot' AND specific_name LIKE '%bigint%')").Scan(&funcExists)
	if result.Error != nil {
		fmt.Printf("⚠️  Could not check SSOT functions: %v\n", result.Error)
	} else {
		fmt.Printf("✅ SSOT sync function (bigint) exists: %v\n", funcExists)
	}

	fmt.Println("🎯 Migration status check completed!")
}