package main

import (
	"fmt"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔧 Creating Account Balances View")
	fmt.Println("=================================")
	fmt.Println()

	db := database.ConnectDB()

	// Drop any existing view first
	fmt.Println("🗑️  Dropping existing account_balances if exists...")
	db.Exec("DROP MATERIALIZED VIEW IF EXISTS account_balances CASCADE")
	db.Exec("DROP VIEW IF EXISTS account_balances CASCADE")

	// Create the view with proper SQL
	fmt.Println("🏗️  Creating account_balances view...")
	createQuery := `
	CREATE VIEW account_balances AS
	SELECT 
		a.id as account_id,
		a.code as account_code,
		a.name as account_name,
		a.type as account_type,
		COALESCE(SUM(jl.debit_amount), 0) as total_debit,
		COALESCE(SUM(jl.credit_amount), 0) as total_credit,
		CASE 
			WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
				COALESCE(SUM(jl.debit_amount), 0) - COALESCE(SUM(jl.credit_amount), 0)
			ELSE 
				COALESCE(SUM(jl.credit_amount), 0) - COALESCE(SUM(jl.debit_amount), 0)
		END as balance,
		NOW() as last_updated
	FROM accounts a
	LEFT JOIN journal_lines jl ON a.id = jl.account_id
	LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id
	WHERE a.is_active = true AND (je.status = 'POSTED' OR je.status IS NULL)
	GROUP BY a.id, a.code, a.name, a.type
	`

	err := db.Exec(createQuery)
	if err.Error != nil {
		fmt.Printf("❌ Error creating view: %v\n", err.Error)
		return
	}

	fmt.Println("✅ Account balances view created successfully!")

	// Test the view
	var count int64
	err = db.Table("account_balances").Count(&count)
	if err != nil {
		fmt.Printf("❌ Error testing view: %v\n", err)
		return
	}

	fmt.Printf("✅ Account balances view contains %d accounts\n", count)

	// Show accounts with balances
	var results []struct {
		AccountCode string  `gorm:"column:account_code"`
		AccountName string  `gorm:"column:account_name"`
		AccountType string  `gorm:"column:account_type"`
		TotalDebit  float64 `gorm:"column:total_debit"`
		TotalCredit float64 `gorm:"column:total_credit"`
		Balance     float64 `gorm:"column:balance"`
	}

	db.Table("account_balances").
		Where("total_debit > 0 OR total_credit > 0").
		Order("ABS(balance) DESC").
		Limit(10).
		Find(&results)

	if len(results) > 0 {
		fmt.Println("\n💰 Accounts with Activity:")
		for _, r := range results {
			fmt.Printf("   %s - %s (%s)\n", r.AccountCode, r.AccountName, r.AccountType)
			fmt.Printf("      Debit: %.2f, Credit: %.2f, Balance: %.2f\n", 
				r.TotalDebit, r.TotalCredit, r.Balance)
		}
	} else {
		fmt.Println("\n⚠️  No accounts with activity found")
	}

	fmt.Println("\n🎉 Account balances view is ready!")
	fmt.Println("✅ Frontend financial reports should now work properly")
	fmt.Println("✅ Try refreshing your browser and generating reports")
}