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

	fmt.Printf("🔄 Running manual balance synchronization...\n\n")

	// Run the manual sync
	rows, err := sqlDB.Query("SELECT * FROM sync_account_balances()")
	if err != nil {
		log.Fatalf("Failed to run balance sync: %v", err)
	}
	defer rows.Close()

	fmt.Printf("=== BALANCE CORRECTIONS ===\n")
	syncCount := 0
	totalAdjustment := 0.0

	for rows.Next() {
		var accountID int
		var oldBalance, newBalance, difference float64
		
		err := rows.Scan(&accountID, &oldBalance, &newBalance, &difference)
		if err != nil {
			log.Printf("Error scanning sync result: %v", err)
			continue
		}

		fmt.Printf("Account ID %d:\n", accountID)
		fmt.Printf("  Old balance: Rp %.0f\n", oldBalance)
		fmt.Printf("  New balance: Rp %.0f\n", newBalance)
		fmt.Printf("  Adjustment:  Rp %.0f\n", difference)
		fmt.Printf("  ---\n")
		
		syncCount++
		totalAdjustment += abs(difference)
	}

	fmt.Printf("\n=== SYNC SUMMARY ===\n")
	fmt.Printf("Accounts corrected: %d\n", syncCount)
	fmt.Printf("Total adjustments: Rp %.0f\n", totalAdjustment)

	if syncCount == 0 {
		fmt.Printf("✅ No corrections needed - all balances are accurate\n")
	} else {
		fmt.Printf("✅ Balance corrections applied successfully\n")
	}

	// Verify the corrections
	fmt.Printf("\n=== VERIFICATION ===\n")
	verifyQuery := `
		SELECT COUNT(*) FROM account_balance_monitoring WHERE status='MISMATCH'
	`
	
	var mismatchCount int
	err = sqlDB.QueryRow(verifyQuery).Scan(&mismatchCount)
	if err != nil {
		log.Printf("Warning: Could not verify corrections: %v", err)
	} else {
		if mismatchCount == 0 {
			fmt.Printf("✅ All account balances are now synchronized\n")
			fmt.Printf("🛡️  Balance protection system is working correctly\n")
		} else {
			fmt.Printf("⚠️  Still %d accounts with mismatches - may need review\n", mismatchCount)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}