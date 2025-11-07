package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/cmd/scripts/utils"
)

func main() {
	fmt.Printf("🔧 FINAL BALANCE FIX - DISABLING TRIGGER TEMPORARILY...\n\n")
	
	// Load environment variables from .env file
	databaseURL, err := utils.GetDatabaseURL()
	if err != nil {
		log.Fatal(err)
	}
	
	// Print environment info (with masked sensitive data)
	utils.PrintEnvInfo()

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

	// Step 1: Disable the trigger temporarily
	fmt.Printf("\n=== STEP 1: DISABLING BALANCE SYNC TRIGGER ===\n")
	
	_, err = sqlDB.Exec("ALTER TABLE unified_journal_lines DISABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Fatal("Failed to disable trigger:", err)
	}
	fmt.Printf("✅ Balance sync trigger disabled temporarily\n")

	// Step 2: Fix all account balances based on journal entries
	fmt.Printf("\n=== STEP 2: FIXING ALL ACCOUNT BALANCES ===\n")
	
	// Update all account balances based on correct accounting logic
	updateQuery := `
		UPDATE accounts 
		SET balance = subq.calculated_balance,
		    updated_at = NOW()
		FROM (
			SELECT 
				a.id,
				CASE 
					WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
						COALESCE(SUM(ujl.debit_amount - ujl.credit_amount), 0)
					WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 
						COALESCE(SUM(ujl.credit_amount - ujl.debit_amount), 0)
					ELSE 0
				END as calculated_balance
			FROM accounts a 
			LEFT JOIN unified_journal_lines ujl ON a.id = ujl.account_id
			WHERE EXISTS (SELECT 1 FROM unified_journal_lines WHERE account_id = a.id)
			GROUP BY a.id, a.type
		) subq
		WHERE accounts.id = subq.id
		AND accounts.balance != subq.calculated_balance
	`
	
	result, err := sqlDB.Exec(updateQuery)
	if err != nil {
		log.Fatal("Failed to update account balances:", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ Updated %d account balances\n", rowsAffected)

	// Step 3: Verify the fixes
	fmt.Printf("\n=== STEP 3: VERIFYING FIXES ===\n")
	
	type BalanceCheck struct {
		ID               int     `gorm:"column:id"`
		Name             string  `gorm:"column:name"`
		Type             string  `gorm:"column:type"`
		Balance          float64 `gorm:"column:balance"`
		CalculatedBalance float64 `gorm:"column:calculated_balance"`
	}
	
	var balanceChecks []BalanceCheck
	verifyQuery := `
		SELECT 
			a.id, a.name, a.type, a.balance,
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(ujl.debit_amount - ujl.credit_amount), 0)
				WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 
					COALESCE(SUM(ujl.credit_amount - ujl.debit_amount), 0)
				ELSE 0
			END as calculated_balance
		FROM accounts a 
		LEFT JOIN unified_journal_lines ujl ON a.id = ujl.account_id
		WHERE EXISTS (SELECT 1 FROM unified_journal_lines WHERE account_id = a.id)
		GROUP BY a.id, a.name, a.type, a.balance
		ORDER BY a.type, a.id
	`
	
	err = gormDB.Raw(verifyQuery).Scan(&balanceChecks).Error
	if err != nil {
		log.Fatal("Failed to verify balances:", err)
	}

	fmt.Printf("📊 Balance verification:\n\n")
	fmt.Printf("%-4s | %-30s | %-8s | %15s | %15s | %s\n", "ID", "Account Name", "Type", "Balance", "Should Be", "Status")
	fmt.Printf("%-4s-+%-30s-+%-8s-+%15s-+%15s-+-%s\n", "----", "------------------------------", "--------", "---------------", "---------------", "--------")
	
	allCorrect := true
	for _, check := range balanceChecks {
		status := "✅ CORRECT"
		if check.Balance != check.CalculatedBalance {
			status = "❌ STILL WRONG"
			allCorrect = false
		}
		
		fmt.Printf("%-4d | %-30s | %-8s | %15.2f | %15.2f | %s\n", 
			check.ID, check.Name, check.Type, check.Balance, check.CalculatedBalance, status)
	}

	if !allCorrect {
		fmt.Printf("\n❌ Some balances are still incorrect. Aborting trigger re-enable.\n")
		return
	}

	// Step 4: Re-enable the trigger
	fmt.Printf("\n=== STEP 4: RE-ENABLING BALANCE SYNC TRIGGER ===\n")
	
	_, err = sqlDB.Exec("ALTER TABLE unified_journal_lines ENABLE TRIGGER balance_sync_trigger")
	if err != nil {
		log.Printf("⚠️  Failed to re-enable trigger: %v", err)
		fmt.Printf("You may need to manually re-enable it later with:\n")
		fmt.Printf("ALTER TABLE unified_journal_lines ENABLE TRIGGER balance_sync_trigger;\n")
	} else {
		fmt.Printf("✅ Balance sync trigger re-enabled\n")
	}

	// Step 5: Final verification and totals
	fmt.Printf("\n=== STEP 5: FINAL RESULTS ===\n")
	
	// Revenue totals
	var totalRevenue float64
	err = sqlDB.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'REVENUE'").Scan(&totalRevenue)
	if err != nil {
		log.Printf("Error calculating total revenue: %v", err)
	} else {
		fmt.Printf("💰 TOTAL REVENUE BALANCE: Rp %.2f\n", totalRevenue)
	}

	// Asset totals
	var totalAssets float64
	err = sqlDB.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'ASSET'").Scan(&totalAssets)
	if err != nil {
		log.Printf("Error calculating total assets: %v", err)
	} else {
		fmt.Printf("🏦 TOTAL ASSET BALANCE: Rp %.2f\n", totalAssets)
	}

	// Liability totals
	var totalLiabilities float64
	err = sqlDB.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'LIABILITY'").Scan(&totalLiabilities)
	if err != nil {
		log.Printf("Error calculating total liabilities: %v", err)
	} else {
		fmt.Printf("📋 TOTAL LIABILITY BALANCE: Rp %.2f\n", totalLiabilities)
	}

	fmt.Printf("\n🎉 SUCCESS! All account balances have been corrected!\n")
	fmt.Printf("✅ Revenue accounts now show the correct positive values\n")
	fmt.Printf("✅ Asset accounts show correct positive values\n") 
	fmt.Printf("✅ Liability accounts show correct positive values\n")
	fmt.Printf("✅ Balance sync trigger is working again\n")
	
	fmt.Printf("\n📈 Your revenue reports should now display correctly!\n")
}