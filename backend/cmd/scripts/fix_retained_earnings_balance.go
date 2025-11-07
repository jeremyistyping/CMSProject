package main

import (
	"fmt"
	"log"
	
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("FIX RETAINED EARNINGS BALANCE")
	fmt.Println("========================================")
	fmt.Println()

	// Initialize database connection
	db := database.DB
	if db == nil {
		log.Fatal("Failed to get database connection. Make sure DB is initialized.")
	}

	fmt.Println("✅ Connected to database")
	fmt.Println()

	// Step 1: Show current state
	fmt.Println("=== CURRENT STATE ===")
	var equityAccounts []models.Account
	if err := db.Where("code IN (?, ?)", "3101", "3201").Find(&equityAccounts).Error; err != nil {
		log.Fatalf("Failed to get equity accounts: %v", err)
	}

	for _, acc := range equityAccounts {
		fmt.Printf("%s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
	}
	fmt.Println()

	// Step 2: Check Revenue/Expense accounts
	fmt.Println("=== REVENUE/EXPENSE ACCOUNTS (should be 0 after closing) ===")
	var tempAccounts []models.Account
	if err := db.Where("type IN (?, ?) AND balance != 0", models.AccountTypeRevenue, models.AccountTypeExpense).
		Find(&tempAccounts).Error; err != nil {
		log.Printf("Warning: Failed to check temp accounts: %v", err)
	}

	if len(tempAccounts) > 0 {
		for _, acc := range tempAccounts {
			fmt.Printf("%s - %s: %.2f (Type: %s)\n", acc.Code, acc.Name, acc.Balance, acc.Type)
		}
	} else {
		fmt.Println("All Revenue/Expense accounts are zero ✅")
	}
	fmt.Println()

	// Step 3: Calculate correct Retained Earnings from closing journals
	fmt.Println("=== CALCULATING CORRECT RETAINED EARNINGS ===")
	
	type Result struct {
		NetMovement float64
	}
	
	var result Result
	query := `
		SELECT COALESCE(SUM(jl.credit_amount - jl.debit_amount), 0) as net_movement
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		JOIN accounts a ON a.id = jl.account_id
		WHERE je.reference_type = 'CLOSING'
			AND je.status = 'POSTED'
			AND je.deleted_at IS NULL
			AND a.code = '3201'
	`
	
	if err := db.Raw(query).Scan(&result).Error; err != nil {
		log.Fatalf("Failed to calculate correct balance: %v", err)
	}

	fmt.Printf("Correct Retained Earnings balance (from closing journals): %.2f\n", result.NetMovement)
	fmt.Println()

	// Get current balance
	var retainedEarnings models.Account
	if err := db.Where("code = ?", "3201").First(&retainedEarnings).Error; err != nil {
		log.Fatalf("Failed to get Retained Earnings account: %v", err)
	}

	difference := retainedEarnings.Balance - result.NetMovement
	fmt.Printf("Current balance: %.2f\n", retainedEarnings.Balance)
	fmt.Printf("Difference: %.2f\n", difference)
	fmt.Println()

	// Step 4: Confirm fix
	if difference == 0 {
		fmt.Println("✅ Retained Earnings balance is already correct!")
		return
	}

	fmt.Println("⚠️  WARNING: This will update the Retained Earnings balance in the database.")
	fmt.Printf("   Current balance: %.2f\n", retainedEarnings.Balance)
	fmt.Printf("   Correct balance: %.2f\n", result.NetMovement)
	fmt.Printf("   Difference: %.2f\n", difference)
	fmt.Println()
	fmt.Print("Do you want to continue? (yes/no): ")
	
	var confirm string
	fmt.Scanln(&confirm)
	
	if confirm != "yes" {
		fmt.Println("Operation cancelled.")
		return
	}

	// Step 5: Fix the balance
	fmt.Println()
	fmt.Println("🔧 Fixing Retained Earnings balance...")
	
	updateQuery := `
		UPDATE accounts
		SET balance = (
			SELECT COALESCE(SUM(jl.credit_amount - jl.debit_amount), 0)
			FROM journal_lines jl
			JOIN journal_entries je ON je.id = jl.journal_entry_id
			WHERE je.reference_type = 'CLOSING'
				AND je.status = 'POSTED'
				AND je.deleted_at IS NULL
				AND jl.account_id = accounts.id
		)
		WHERE code = '3201' AND type = 'EQUITY'
	`
	
	if err := db.Exec(updateQuery).Error; err != nil {
		log.Fatalf("Failed to fix balance: %v", err)
	}

	fmt.Println("✅ Balance updated successfully!")
	fmt.Println()

	// Step 6: Verify fix
	fmt.Println("=== AFTER FIX ===")
	if err := db.Where("code IN (?, ?)", "3101", "3201").Find(&equityAccounts).Error; err != nil {
		log.Fatalf("Failed to verify fix: %v", err)
	}

	for _, acc := range equityAccounts {
		fmt.Printf("%s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
	}
	fmt.Println()

	// Step 7: Check balance sheet totals
	fmt.Println("=== BALANCE SHEET CHECK ===")
	
	type BalanceCheck struct {
		TotalAssets     float64
		TotalLiabilities float64
		TotalEquity     float64
		TotalLiabEquity float64
		Difference      float64
	}
	
	var check BalanceCheck
	balanceCheckQuery := `
		WITH account_balances AS (
			SELECT 
				type,
				CASE WHEN type = 'ASSET' THEN SUM(balance) ELSE 0 END as asset_balance,
				CASE WHEN type = 'LIABILITY' THEN SUM(balance) ELSE 0 END as liability_balance,
				CASE WHEN type = 'EQUITY' THEN SUM(balance) ELSE 0 END as equity_balance
			FROM accounts
			WHERE deleted_at IS NULL
				AND is_active = true
				AND COALESCE(is_header, false) = false
			GROUP BY type
		)
		SELECT 
			SUM(asset_balance) as total_assets,
			SUM(liability_balance) as total_liabilities,
			SUM(equity_balance) as total_equity,
			SUM(liability_balance) + SUM(equity_balance) as total_liab_equity,
			SUM(asset_balance) - (SUM(liability_balance) + SUM(equity_balance)) as difference
		FROM account_balances
	`
	
	if err := db.Raw(balanceCheckQuery).Scan(&check).Error; err != nil {
		log.Printf("Warning: Failed to check balance sheet: %v", err)
	} else {
		fmt.Printf("Total Assets: %.2f\n", check.TotalAssets)
		fmt.Printf("Total Liabilities: %.2f\n", check.TotalLiabilities)
		fmt.Printf("Total Equity: %.2f\n", check.TotalEquity)
		fmt.Printf("Total Liabilities + Equity: %.2f\n", check.TotalLiabEquity)
		fmt.Printf("Difference: %.2f\n", check.Difference)
		
		if check.Difference >= -0.01 && check.Difference <= 0.01 {
			fmt.Println("\n✅ Balance Sheet is BALANCED!")
		} else {
			fmt.Println("\n⚠️  Balance Sheet still has difference. Please investigate.")
		}
	}
	
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("✅ FIX COMPLETED!")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Restart your backend server")
	fmt.Println("2. Refresh your Balance Sheet report")
	fmt.Println("3. Verify that the difference is now 0")
}
