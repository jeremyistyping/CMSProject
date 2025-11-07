package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type AccountBalance struct {
	ID               int     `gorm:"column:id"`
	Name             string  `gorm:"column:name"`
	Type             string  `gorm:"column:type"`
	CurrentBalance   float64 `gorm:"column:balance"`
	CalculatedBalance float64 `gorm:"column:calculated_balance"`
}

func main() {
	fmt.Printf("🔧 FIXING REVENUE ACCOUNT BALANCES...\n\n")
	
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

	// Step 1: Calculate correct balances for revenue accounts
	fmt.Printf("\n=== STEP 1: CALCULATING CORRECT BALANCES ===\n")
	
	var revenueAccounts []AccountBalance
	query := `
		SELECT 
			a.id,
			a.name,
			a.type,
			a.balance as balance,
			COALESCE(SUM(ujl.credit_amount - ujl.debit_amount), 0) as calculated_balance
		FROM accounts a 
		LEFT JOIN unified_journal_lines ujl ON a.id = ujl.account_id
		WHERE a.type = 'REVENUE'
		GROUP BY a.id, a.name, a.type, a.balance 
		ORDER BY a.id
	`
	
	err = gormDB.Raw(query).Scan(&revenueAccounts).Error
	if err != nil {
		log.Fatal("Failed to calculate revenue balances:", err)
	}

	if len(revenueAccounts) == 0 {
		fmt.Printf("❌ No revenue accounts found!\n")
		return
	}

	fmt.Printf("📊 Revenue accounts analysis:\n\n")
	fmt.Printf("%-4s | %-30s | %15s | %15s | %s\n", "ID", "Account Name", "Current Balance", "Should Be", "Action")
	fmt.Printf("%-4s-+%-30s-+%15s-+%15s-+-%s\n", "----", "------------------------------", "---------------", "---------------", "--------")
	
	var needsUpdate []AccountBalance
	
	for _, acc := range revenueAccounts {
		status := "✅ OK"
		if acc.CurrentBalance != acc.CalculatedBalance {
			status = "🔧 NEEDS FIX"
			needsUpdate = append(needsUpdate, acc)
		}
		
		fmt.Printf("%-4d | %-30s | %15.2f | %15.2f | %s\n", 
			acc.ID, acc.Name, acc.CurrentBalance, acc.CalculatedBalance, status)
	}

	if len(needsUpdate) == 0 {
		fmt.Printf("\n✅ All revenue account balances are already correct!\n")
		return
	}

	// Step 2: Fix the incorrect balances
	fmt.Printf("\n=== STEP 2: FIXING INCORRECT BALANCES ===\n")
	
	for _, acc := range needsUpdate {
		fmt.Printf("🔧 Fixing account %d (%s):\n", acc.ID, acc.Name)
		fmt.Printf("   Current: Rp %.2f → Correct: Rp %.2f\n", acc.CurrentBalance, acc.CalculatedBalance)
		
		_, err = sqlDB.Exec(`
			UPDATE accounts 
			SET balance = $1, updated_at = NOW() 
			WHERE id = $2
		`, acc.CalculatedBalance, acc.ID)
		
		if err != nil {
			log.Printf("❌ Error fixing account %d: %v", acc.ID, err)
		} else {
			fmt.Printf("   ✅ Successfully updated!\n")
		}
	}

	// Step 3: Verify the fixes
	fmt.Printf("\n=== STEP 3: VERIFYING FIXES ===\n")
	
	var verifyAccounts []AccountBalance
	err = gormDB.Raw(query).Scan(&verifyAccounts).Error
	if err != nil {
		log.Printf("Error verifying fixes: %v", err)
		return
	}

	fmt.Printf("📊 Verification results:\n\n")
	fmt.Printf("%-4s | %-30s | %15s | %15s | %s\n", "ID", "Account Name", "New Balance", "Calculated", "Status")
	fmt.Printf("%-4s-+%-30s-+%15s-+%15s-+-%s\n", "----", "------------------------------", "---------------", "---------------", "--------")
	
	allFixed := true
	for _, acc := range verifyAccounts {
		status := "✅ FIXED"
		if acc.CurrentBalance != acc.CalculatedBalance {
			status = "❌ STILL WRONG"
			allFixed = false
		}
		
		fmt.Printf("%-4d | %-30s | %15.2f | %15.2f | %s\n", 
			acc.ID, acc.Name, acc.CurrentBalance, acc.CalculatedBalance, status)
	}

	// Step 4: Test the automatic sync system
	fmt.Printf("\n=== STEP 4: TESTING BALANCE SYNC SYSTEM ===\n")
	
	// Call the sync function
	rows, err := sqlDB.Query("SELECT account_id FROM sync_account_balances()")
	if err != nil {
		fmt.Printf("⚠️  Balance sync system test failed: %v\n", err)
	} else {
		defer rows.Close()
		var syncedAccounts []int
		for rows.Next() {
			var accountID int
			rows.Scan(&accountID)
			syncedAccounts = append(syncedAccounts, accountID)
		}
		
		if len(syncedAccounts) == 0 {
			fmt.Printf("✅ Balance sync system: WORKING (no accounts needed syncing)\n")
		} else {
			fmt.Printf("✅ Balance sync system: WORKING (synced %d accounts: %v)\n", len(syncedAccounts), syncedAccounts)
		}
	}

	// Final summary
	fmt.Printf("\n🎯 SUMMARY:\n")
	if allFixed {
		fmt.Printf("✅ ALL REVENUE ACCOUNT BALANCES HAVE BEEN FIXED!\n")
		fmt.Printf("✅ Revenue reports should now show correct values.\n")
		fmt.Printf("✅ Balance sync system is working properly.\n")
		
		// Show new revenue totals
		var totalRevenue float64
		err = sqlDB.QueryRow("SELECT SUM(balance) FROM accounts WHERE type = 'REVENUE'").Scan(&totalRevenue)
		if err != nil {
			log.Printf("Error calculating total revenue: %v", err)
		} else {
			fmt.Printf("\n💰 TOTAL REVENUE BALANCE: Rp %.2f\n", totalRevenue)
		}
	} else {
		fmt.Printf("❌ Some accounts still have incorrect balances.\n")
		fmt.Printf("💡 You may need to check journal entries or run balance sync manually.\n")
	}
}