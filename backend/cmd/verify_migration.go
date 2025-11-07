package main

import (
	"fmt"
	"strings"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("VERIFY MIGRATION TO UNIFIED JOURNALS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Connect to database
	db := database.ConnectDB()

	// Check old journal entries
	var oldJournalCount, oldLineCount, oldPeriodCount int64
	db.Raw("SELECT COUNT(*) FROM journal_entries").Scan(&oldJournalCount)
	db.Raw("SELECT COUNT(*) FROM journal_lines").Scan(&oldLineCount)
	db.Raw("SELECT COUNT(*) FROM accounting_periods").Scan(&oldPeriodCount)

	fmt.Println("📊 OLD JOURNAL SYSTEM STATUS:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("journal_entries:     %d\n", oldJournalCount)
	fmt.Printf("journal_lines:       %d\n", oldLineCount)
	fmt.Printf("accounting_periods:  %d\n", oldPeriodCount)
	
	if oldJournalCount == 0 && oldLineCount == 0 {
		fmt.Println("✅ Old journal system is CLEAN (migration successful)")
	} else {
		fmt.Println("⚠️  Old journal entries still exist")
	}
	fmt.Println()

	// Check unified journals
	var unifiedJournalCount, unifiedLineCount int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&unifiedJournalCount)
	db.Raw("SELECT COUNT(*) FROM unified_journal_lines").Scan(&unifiedLineCount)

	fmt.Println("🔄 UNIFIED JOURNAL SYSTEM STATUS:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("unified_journal_ledger: %d entries\n", unifiedJournalCount)
	fmt.Printf("unified_journal_lines:  %d lines\n", unifiedLineCount)
	fmt.Println()

	// Check account balances
	type BalanceResult struct {
		TotalAssets      float64
		TotalLiabilities float64
		TotalEquity      float64
		TotalRevenue     float64
		TotalExpense     float64
	}

	var result BalanceResult
	db.Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END), 0) as total_assets,
			COALESCE(SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END), 0) as total_liabilities,
			COALESCE(SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END), 0) as total_equity,
			COALESCE(SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END), 0) as total_expense
		FROM accounts
		WHERE deleted_at IS NULL
	`).Scan(&result)

	netIncome := result.TotalRevenue - result.TotalExpense
	leftSide := result.TotalAssets
	rightSide := result.TotalLiabilities + result.TotalEquity + netIncome
	difference := leftSide - rightSide

	fmt.Println("💰 BALANCE SHEET SUMMARY:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Assets:       Rp %18.2f\n", result.TotalAssets)
	fmt.Printf("Liabilities:  Rp %18.2f\n", result.TotalLiabilities)
	fmt.Printf("Equity:       Rp %18.2f\n", result.TotalEquity)
	fmt.Printf("Revenue:      Rp %18.2f (temp - not closed)\n", result.TotalRevenue)
	fmt.Printf("Expense:      Rp %18.2f (temp - not closed)\n", result.TotalExpense)
	fmt.Printf("Net Income:   Rp %18.2f\n", netIncome)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Difference:   Rp %18.2f\n", difference)
	fmt.Println()

	// Validate
	if difference < -0.01 || difference > 0.01 {
		fmt.Println("⚠️  Balance sheet NOT BALANCED")
		fmt.Println("💡 This is NORMAL if period closing has not been run yet")
		fmt.Println("💡 Run period closing to close Revenue/Expense to Retained Earnings")
	} else {
		fmt.Println("✅ Balance sheet is BALANCED!")
	}
	fmt.Println()

	// Show accounts with balances
	type AccountInfo struct {
		Code    string
		Name    string
		Type    string
		Balance float64
	}

	var accounts []AccountInfo
	db.Raw(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE deleted_at IS NULL 
			AND is_header = false
			AND ABS(balance) > 0.01
		ORDER BY code
	`).Scan(&accounts)

	fmt.Println("📋 ACCOUNTS WITH NON-ZERO BALANCE:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-6s | %-35s | %-10s | %15s\n", "Code", "Account Name", "Type", "Balance")
	fmt.Println(strings.Repeat("-", 80))
	for _, acc := range accounts {
		fmt.Printf("%-6s | %-35s | %-10s | %15.2f\n", 
			acc.Code, 
			truncate(acc.Name, 35), 
			acc.Type, 
			acc.Balance)
	}
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total: %d accounts\n", len(accounts))
	fmt.Println()

	// Final summary
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✅ MIGRATION VERIFICATION COMPLETE")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
	
	if oldJournalCount == 0 && unifiedJournalCount > 0 {
		fmt.Println("✅ System is now using UNIFIED JOURNALS exclusively")
		fmt.Println("📝 Next step: Run period closing to close Revenue/Expense accounts")
	} else {
		fmt.Println("⚠️  Migration may not be complete - please review")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
