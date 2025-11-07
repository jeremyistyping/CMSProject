package main

import (
	"fmt"
	"strings"
	"app-sistem-akuntansi/database"
)

func main() {
	db := database.ConnectDB()
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("CHECKING UNIFIED JOURNAL TABLES (SSOT SYSTEM)")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	
	// 1. Check unified_journal_ledger
	var unifiedCount, unifiedPosted, unifiedDraft int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&unifiedCount)
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger WHERE status = 'POSTED'").Scan(&unifiedPosted)
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger WHERE status = 'DRAFT'").Scan(&unifiedDraft)
	
	fmt.Println("📊 UNIFIED_JOURNAL_LEDGER:")
	fmt.Printf("  Total records:  %d\n", unifiedCount)
	fmt.Printf("  Posted:         %d\n", unifiedPosted)
	fmt.Printf("  Draft:          %d\n", unifiedDraft)
	fmt.Println()
	
	// 2. Check unified_journal_lines
	var linesCount int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_lines").Scan(&linesCount)
	fmt.Printf("📋 UNIFIED_JOURNAL_LINES: %d records\n\n", linesCount)
	
	// 3. Show sample of unified journal entries
	type UnifiedJournal struct {
		ID           uint
		Code         string
		EntryDate    string
		Description  string
		SourceType   string
		SourceID     *uint
		Status       string
		TotalDebit   float64
		TotalCredit  float64
	}
	
	var unifiedJournals []UnifiedJournal
	db.Raw(`
		SELECT 
			id,
			code,
			entry_date::text,
			SUBSTRING(description, 1, 60) as description,
			source_type,
			source_id,
			status,
			total_debit,
			total_credit
		FROM unified_journal_ledger
		ORDER BY entry_date DESC, id DESC
		LIMIT 20
	`).Scan(&unifiedJournals)
	
	fmt.Println("📖 UNIFIED JOURNAL LEDGER (Recent 20):")
	fmt.Println("ID   | Code         | Date       | Source Type | Status | Debit      | Credit")
	fmt.Println("-----+--------------+------------+-------------+--------+------------+-----------")
	
	for _, j := range unifiedJournals {
		fmt.Printf("%-4d | %-12s | %-10s | %-11s | %-6s | %10.0f | %10.0f\n",
			j.ID, j.Code, j.EntryDate[:10], j.SourceType, j.Status, j.TotalDebit, j.TotalCredit)
	}
	
	// 4. Check account balances calculated from unified journals
	type AccountBalance struct {
		AccountCode  string
		AccountName  string
		AccountType  string
		TotalDebit   float64
		TotalCredit  float64
		NetBalance   float64
	}
	
	var accountBalances []AccountBalance
	db.Raw(`
		SELECT 
			a.code as account_code,
			a.name as account_name,
			CASE 
				WHEN a.code LIKE '1%' THEN 'ASSET'
				WHEN a.code LIKE '2%' THEN 'LIABILITY'
				WHEN a.code LIKE '3%' THEN 'EQUITY'
				WHEN a.code LIKE '4%' THEN 'REVENUE'
				WHEN a.code LIKE '5%' THEN 'EXPENSE'
				ELSE 'OTHER'
			END as account_type,
			COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
			COALESCE(SUM(ujl.credit_amount), 0) as total_credit,
			CASE 
				WHEN a.code LIKE '1%' OR a.code LIKE '5%' 
				THEN COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as net_balance
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uj ON uj.id = ujl.journal_id
		LEFT JOIN accounts a ON a.id = ujl.account_id
		WHERE uj.status = 'POSTED'
		GROUP BY a.id, a.code, a.name
		HAVING ABS(COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)) > 0.01
		ORDER BY a.code
		LIMIT 30
	`).Scan(&accountBalances)
	
	fmt.Println("\n💰 ACCOUNT BALANCES (from Unified Journals):")
	fmt.Println("Code | Account Name              | Type      | Debit      | Credit     | Net Balance")
	fmt.Println("-----+---------------------------+-----------+------------+------------+------------")
	
	var totalAssets, totalLiabilities, totalEquity float64
	
	for _, ab := range accountBalances {
		fmt.Printf("%-4s | %-25s | %-9s | %10.0f | %10.0f | %11.0f\n",
			ab.AccountCode, ab.AccountName, ab.AccountType, 
			ab.TotalDebit, ab.TotalCredit, ab.NetBalance)
		
		switch ab.AccountType {
		case "ASSET":
			totalAssets += ab.NetBalance
		case "LIABILITY":
			totalLiabilities += ab.NetBalance
		case "EQUITY":
			totalEquity += ab.NetBalance
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("BALANCE SHEET FROM UNIFIED JOURNALS:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Total Assets:      Rp %15.2f\n", totalAssets)
	fmt.Printf("Total Liabilities: Rp %15.2f\n", totalLiabilities)
	fmt.Printf("Total Equity:      Rp %15.2f\n", totalEquity)
	fmt.Printf("Liab + Equity:     Rp %15.2f\n", totalLiabilities + totalEquity)
	fmt.Printf("Difference:        Rp %15.2f", totalAssets - (totalLiabilities + totalEquity))
	
	if totalAssets - (totalLiabilities + totalEquity) == 0 {
		fmt.Println(" ✅ BALANCED!")
	} else {
		fmt.Println(" ❌ NOT BALANCED")
	}
	
	// 5. Compare with accounts table
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("COMPARISON: Unified Journals vs Accounts Table")
	fmt.Println(strings.Repeat("=", 80) + "\n")
	
	type BalanceComparison struct {
		AccountCode      string
		AccountName      string
		UnifiedBalance   float64
		AccountsBalance  float64
		Difference       float64
	}
	
	var comparisons []BalanceComparison
	db.Raw(`
		WITH unified_balances AS (
			SELECT 
				a.code,
				a.name,
				CASE 
					WHEN a.code LIKE '1%' OR a.code LIKE '5%' 
					THEN COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
					ELSE COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
				END as unified_balance
			FROM unified_journal_lines ujl
			JOIN unified_journal_ledger uj ON uj.id = ujl.journal_id
			LEFT JOIN accounts a ON a.id = ujl.account_id
			WHERE uj.status = 'POSTED'
			GROUP BY a.id, a.code, a.name
		)
		SELECT 
			COALESCE(a.code, ub.code) as account_code,
			COALESCE(a.name, ub.name) as account_name,
			COALESCE(ub.unified_balance, 0) as unified_balance,
			COALESCE(a.balance, 0) as accounts_balance,
			COALESCE(a.balance, 0) - COALESCE(ub.unified_balance, 0) as difference
		FROM accounts a
		FULL OUTER JOIN unified_balances ub ON ub.code = a.code
		WHERE a.is_active = true AND COALESCE(a.is_header, false) = false
			AND (ABS(COALESCE(a.balance, 0)) > 0.01 OR ABS(COALESCE(ub.unified_balance, 0)) > 0.01)
		ORDER BY account_code
	`).Scan(&comparisons)
	
	fmt.Println("Code | Account Name              | Unified    | Accounts   | Difference")
	fmt.Println("-----+---------------------------+------------+------------+-----------")
	
	for _, cmp := range comparisons {
		status := ""
		if cmp.Difference != 0 {
			status = " ⚠️"
		}
		fmt.Printf("%-4s | %-25s | %10.0f | %10.0f | %10.0f%s\n",
			cmp.AccountCode, cmp.AccountName, cmp.UnifiedBalance, 
			cmp.AccountsBalance, cmp.Difference, status)
	}
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("END OF UNIFIED JOURNAL CHECK")
	fmt.Println(strings.Repeat("=", 80) + "\n")
}
