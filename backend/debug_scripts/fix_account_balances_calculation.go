package main

import (
	"fmt"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔧 Fix Account Balances Calculation")
	fmt.Println("==================================")
	fmt.Println()

	db := database.ConnectDB()

	// First check what's in journal_lines
	fmt.Println("🔍 Step 1: Checking journal_lines data...")
	var journalLines []struct {
		ID             uint    `gorm:"column:id"`
		JournalEntryID uint    `gorm:"column:journal_entry_id"`
		AccountID      uint    `gorm:"column:account_id"`
		DebitAmount    float64 `gorm:"column:debit_amount"`
		CreditAmount   float64 `gorm:"column:credit_amount"`
		Description    string  `gorm:"column:description"`
	}

	db.Table("journal_lines").
		Select("id, journal_entry_id, account_id, debit_amount, credit_amount, description").
		Find(&journalLines)

	fmt.Printf("📋 Found %d journal lines:\n", len(journalLines))
	totalDebits := 0.0
	totalCredits := 0.0
	for _, line := range journalLines {
		fmt.Printf("   Line %d: Account %d, Debit: %.2f, Credit: %.2f, Desc: %s\n", 
			line.ID, line.AccountID, line.DebitAmount, line.CreditAmount, line.Description)
		totalDebits += line.DebitAmount
		totalCredits += line.CreditAmount
	}
	fmt.Printf("📊 Total Debits: %.2f, Total Credits: %.2f\n", totalDebits, totalCredits)
	fmt.Println()

	// Check which accounts are being used
	fmt.Println("🔍 Step 2: Checking which accounts have transactions...")
	var accountUsage []struct {
		AccountID    uint    `gorm:"column:account_id"`
		AccountCode  string  `gorm:"column:account_code"`
		AccountName  string  `gorm:"column:account_name"`
		AccountType  string  `gorm:"column:account_type"`
		TotalDebit   float64 `gorm:"column:total_debit"`
		TotalCredit  float64 `gorm:"column:total_credit"`
		TransactionCount int `gorm:"column:transaction_count"`
	}

	db.Raw(`
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(jl.debit_amount), 0) as total_debit,
			COALESCE(SUM(jl.credit_amount), 0) as total_credit,
			COUNT(jl.id) as transaction_count
		FROM accounts a
		LEFT JOIN journal_lines jl ON a.id = jl.account_id
		WHERE a.is_active = true 
		GROUP BY a.id, a.code, a.name, a.type
		HAVING COUNT(jl.id) > 0
		ORDER BY total_debit + total_credit DESC
	`).Find(&accountUsage)

	fmt.Printf("📋 Accounts with transactions (%d):\n", len(accountUsage))
	for _, acc := range accountUsage {
		balance := 0.0
		if acc.AccountType == "ASSET" || acc.AccountType == "EXPENSE" {
			balance = acc.TotalDebit - acc.TotalCredit
		} else {
			balance = acc.TotalCredit - acc.TotalDebit
		}
		fmt.Printf("   %s - %s (%s): Debit=%.2f, Credit=%.2f, Balance=%.2f, Count=%d\n", 
			acc.AccountCode, acc.AccountName, acc.AccountType, 
			acc.TotalDebit, acc.TotalCredit, balance, acc.TransactionCount)
	}
	fmt.Println()

	// Now recreate the account_balances view with correct calculation
	fmt.Println("🔧 Step 3: Recreating account_balances view...")
	
	// Drop existing view
	db.Exec("DROP VIEW IF EXISTS account_balances CASCADE")

	// Create new view with corrected logic
	createViewQuery := `
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
	WHERE a.is_active = true
	GROUP BY a.id, a.code, a.name, a.type
	`

	err := db.Exec(createViewQuery)
	if err.Error != nil {
		fmt.Printf("❌ Error creating view: %v\n", err.Error)
		return
	}

	fmt.Println("✅ Account balances view recreated!")

	// Test the new view
	fmt.Println("\n🧪 Step 4: Testing new account balances view...")
	var newBalances []struct {
		AccountCode  string  `gorm:"column:account_code"`
		AccountName  string  `gorm:"column:account_name"`
		AccountType  string  `gorm:"column:account_type"`
		TotalDebit   float64 `gorm:"column:total_debit"`
		TotalCredit  float64 `gorm:"column:total_credit"`
		Balance      float64 `gorm:"column:balance"`
	}

	db.Table("account_balances").
		Where("total_debit > 0 OR total_credit > 0").
		Order("ABS(balance) DESC").
		Find(&newBalances)

	fmt.Printf("📋 Active account balances (%d):\n", len(newBalances))
	for _, bal := range newBalances {
		fmt.Printf("   %s - %s (%s): Debit=%.2f, Credit=%.2f, Balance=%.2f\n", 
			bal.AccountCode, bal.AccountName, bal.AccountType, 
			bal.TotalDebit, bal.TotalCredit, bal.Balance)
	}

	// Check if we now have non-zero balances for revenue/expense accounts
	fmt.Println("\n🎯 Step 5: Checking Revenue & Expense Accounts...")
	
	var revenueAccounts []struct {
		AccountCode string  `gorm:"column:account_code"`
		AccountName string  `gorm:"column:account_name"`
		Balance     float64 `gorm:"column:balance"`
	}
	
	db.Table("account_balances").
		Where("account_type = 'REVENUE' AND balance != 0").
		Select("account_code, account_name, balance").
		Find(&revenueAccounts)

	var expenseAccounts []struct {
		AccountCode string  `gorm:"column:account_code"`
		AccountName string  `gorm:"column:account_name"`
		Balance     float64 `gorm:"column:balance"`
	}
	
	db.Table("account_balances").
		Where("account_type = 'EXPENSE' AND balance != 0").
		Select("account_code, account_name, balance").
		Find(&expenseAccounts)

	fmt.Printf("💰 Revenue Accounts with balance: %d\n", len(revenueAccounts))
	for _, acc := range revenueAccounts {
		fmt.Printf("   %s - %s: %.2f\n", acc.AccountCode, acc.AccountName, acc.Balance)
	}

	fmt.Printf("💸 Expense Accounts with balance: %d\n", len(expenseAccounts))
	for _, acc := range expenseAccounts {
		fmt.Printf("   %s - %s: %.2f\n", acc.AccountCode, acc.AccountName, acc.Balance)
	}

	fmt.Println()
	if len(revenueAccounts) > 0 || len(expenseAccounts) > 0 {
		fmt.Println("🎉 SUCCESS! Account balances now have non-zero values!")
		fmt.Println("✅ P&L and Balance Sheet reports should now show data")
	} else {
		fmt.Println("⚠️  Still no revenue/expense balances - need to check account mapping")
		fmt.Println("💡 The journal entries may not be properly mapped to revenue/expense accounts")
	}
}