package main

import (
	"fmt"
	"strings"
	"app-sistem-akuntansi/database"
)

type JournalDetail struct {
	JournalID    uint
	Code         string
	EntryDate    string
	Description  string
	ReferenceType string
	Status       string
	TotalDebit   float64
	TotalCredit  float64
	IsBalanced   bool
}

type AccountTransaction struct {
	AccountCode  string
	AccountName  string
	JournalCode  string
	EntryDate    string
	Description  string
	DebitAmount  float64
	CreditAmount float64
	Status       string
}

type ClosingEntry struct {
	ID           uint
	Code         string
	EntryDate    string
	Description  string
	TotalDebit   float64
	TotalCredit  float64
	Status       string
}

func main() {
	db := database.ConnectDB()
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("DIAGNOSTIC: ROOT CAUSE ANALYSIS - BALANCE SHEET TIDAK BALANCE")
	fmt.Println(strings.Repeat("=", 60) + "\n")
	
	// 1. Check current balance status
	fmt.Println("📊 [1] CURRENT BALANCE STATUS")
	fmt.Println(strings.Repeat("-", 60))
	
	type BalanceSummary struct {
		TotalAssets      float64
		TotalLiabilities float64
		TotalEquity      float64
		TotalRevenue     float64
		TotalExpense     float64
	}
	
	var summary BalanceSummary
	db.Raw(`
		SELECT 
			SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) AS total_assets,
			SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) AS total_liabilities,
			SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END) AS total_equity,
			SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END) AS total_revenue,
			SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END) AS total_expense
		FROM accounts
		WHERE is_active = true AND COALESCE(is_header, false) = false
	`).Scan(&summary)
	
	fmt.Printf("Assets:              Rp %15.2f\n", summary.TotalAssets)
	fmt.Printf("Liabilities:         Rp %15.2f\n", summary.TotalLiabilities)
	fmt.Printf("Equity:              Rp %15.2f\n", summary.TotalEquity)
	fmt.Printf("Total Liab + Equity: Rp %15.2f\n", summary.TotalLiabilities + summary.TotalEquity)
	
	diff := summary.TotalAssets - (summary.TotalLiabilities + summary.TotalEquity)
	fmt.Printf("\n⚠️  DIFFERENCE:        Rp %15.2f", diff)
	
	if diff == 0 {
		fmt.Println(" ✅ BALANCED")
	} else {
		fmt.Println(" ❌ NOT BALANCED")
	}
	
	fmt.Printf("\nRevenue (should be 0): Rp %15.2f\n", summary.TotalRevenue)
	fmt.Printf("Expense (should be 0): Rp %15.2f\n", summary.TotalExpense)
	
	// 2. Check for unbalanced journal entries
	fmt.Println("\n📋 [2] UNBALANCED JOURNAL ENTRIES")
	fmt.Println(strings.Repeat("-", 60))
	
	var unbalancedJournals []JournalDetail
	db.Raw(`
		SELECT 
			id as journal_id,
			code,
			entry_date::text,
			description,
			reference_type,
			status,
			total_debit,
			total_credit,
			is_balanced
		FROM journal_entries
		WHERE status = 'POSTED' 
			AND (is_balanced = false OR ABS(total_debit - total_credit) > 0.01)
		ORDER BY entry_date, id
	`).Scan(&unbalancedJournals)
	
	if len(unbalancedJournals) > 0 {
		fmt.Printf("❌ Found %d unbalanced posted journal entries:\n\n", len(unbalancedJournals))
		for _, j := range unbalancedJournals {
			fmt.Printf("  ID: %d | Code: %s | Date: %s\n", j.JournalID, j.Code, j.EntryDate)
			fmt.Printf("  Description: %s\n", j.Description)
			fmt.Printf("  Debit: %.2f | Credit: %.2f | Diff: %.2f\n", 
				j.TotalDebit, j.TotalCredit, j.TotalDebit - j.TotalCredit)
			fmt.Println()
		}
	} else {
		fmt.Println("✅ No unbalanced journal entries found")
	}
	
	// 3. Check Laba Ditahan (3201) transactions
	fmt.Println("\n💰 [3] LABA DITAHAN (3201) TRANSACTION HISTORY")
	fmt.Println(strings.Repeat("-", 60))
	
	var labaTransactions []AccountTransaction
	db.Raw(`
		SELECT 
			a.code as account_code,
			a.name as account_name,
			je.code as journal_code,
			je.entry_date::text,
			jl.description,
			jl.debit_amount,
			jl.credit_amount,
			je.status
		FROM journal_lines jl
		JOIN journal_entries je ON jl.journal_entry_id = je.id
		JOIN accounts a ON jl.account_id = a.id
		WHERE a.code = '3201'
			AND je.status = 'POSTED'
		ORDER BY je.entry_date, je.id
	`).Scan(&labaTransactions)
	
	fmt.Printf("Found %d transactions to Laba Ditahan:\n\n", len(labaTransactions))
	
	totalDebit := 0.0
	totalCredit := 0.0
	
	for i, t := range labaTransactions {
		netEffect := t.CreditAmount - t.DebitAmount
		totalDebit += t.DebitAmount
		totalCredit += t.CreditAmount
		
		fmt.Printf("%d. [%s] %s\n", i+1, t.EntryDate, t.JournalCode)
		fmt.Printf("   %s\n", t.Description)
		fmt.Printf("   DR: %15.2f | CR: %15.2f | Net: %15.2f\n", 
			t.DebitAmount, t.CreditAmount, netEffect)
		fmt.Println()
	}
	
	fmt.Printf("TOTAL to Laba Ditahan:\n")
	fmt.Printf("  Total Debit:  Rp %15.2f\n", totalDebit)
	fmt.Printf("  Total Credit: Rp %15.2f\n", totalCredit)
	fmt.Printf("  Net Balance:  Rp %15.2f\n", totalCredit - totalDebit)
	
	// Get actual balance from accounts table
	var actualBalance float64
	db.Raw(`SELECT balance FROM accounts WHERE code = '3201' AND is_active = true LIMIT 1`).Scan(&actualBalance)
	fmt.Printf("  Actual Balance in DB: Rp %15.2f\n", actualBalance)
	
	balanceDiff := actualBalance - (totalCredit - totalDebit)
	if balanceDiff != 0 {
		fmt.Printf("  ⚠️  MISMATCH: Rp %15.2f\n", balanceDiff)
	}
	
	// 4. Check all closing entries
	fmt.Println("\n🔒 [4] PERIOD CLOSING ENTRIES")
	fmt.Println(strings.Repeat("-", 60))
	
	var closingEntries []ClosingEntry
	db.Raw(`
		SELECT 
			id,
			code,
			entry_date::text,
			description,
			total_debit,
			total_credit,
			status
		FROM journal_entries
		WHERE reference_type = 'CLOSING'
		ORDER BY entry_date, id
	`).Scan(&closingEntries)
	
	if len(closingEntries) > 0 {
		fmt.Printf("Found %d closing entries:\n\n", len(closingEntries))
		for i, ce := range closingEntries {
			fmt.Printf("%d. ID: %d | Code: %s | Date: %s | Status: %s\n", 
				i+1, ce.ID, ce.Code, ce.EntryDate, ce.Status)
			fmt.Printf("   Description: %s\n", ce.Description)
			fmt.Printf("   Total Debit: %.2f | Total Credit: %.2f\n", ce.TotalDebit, ce.TotalCredit)
			
			// Get details of this closing entry
			var closingDetails []AccountTransaction
			db.Raw(`
				SELECT 
					a.code as account_code,
					a.name as account_name,
					jl.description,
					jl.debit_amount,
					jl.credit_amount
				FROM journal_lines jl
				JOIN accounts a ON jl.account_id = a.id
				WHERE jl.journal_entry_id = ?
				ORDER BY jl.line_number
			`, ce.ID).Scan(&closingDetails)
			
			for _, d := range closingDetails {
				fmt.Printf("     %s (%s): DR %.2f | CR %.2f\n", 
					d.AccountCode, d.AccountName, d.DebitAmount, d.CreditAmount)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("No closing entries found")
	}
	
	// 5. Check for negative balances (should not exist in assets)
	fmt.Println("\n⚠️  [5] ACCOUNTS WITH NEGATIVE BALANCES")
	fmt.Println(strings.Repeat("-", 60))
	
	var negativeAccounts []struct {
		Code    string
		Name    string
		Type    string
		Balance float64
	}
	
	db.Raw(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE is_active = true 
			AND COALESCE(is_header, false) = false
			AND (
				(type = 'ASSET' AND balance < 0) OR
				(type = 'LIABILITY' AND balance < 0) OR
				(type = 'EQUITY' AND balance < 0)
			)
		ORDER BY type, code
	`).Scan(&negativeAccounts)
	
	if len(negativeAccounts) > 0 {
		fmt.Printf("❌ Found %d accounts with abnormal negative balances:\n\n", len(negativeAccounts))
		for _, acc := range negativeAccounts {
			fmt.Printf("  %s - %s (%s): Rp %.2f\n", acc.Code, acc.Name, acc.Type, acc.Balance)
		}
	} else {
		fmt.Println("✅ No abnormal negative balances found")
	}
	
	// 6. Check journal entry totals vs posted amounts
	fmt.Println("\n🔍 [6] VERIFY JOURNAL ENTRY INTEGRITY")
	fmt.Println(strings.Repeat("-", 60))
	
	var integrityIssues []struct {
		JournalID       uint
		Code            string
		EntryDate       string
		StoredDebit     float64
		StoredCredit    float64
		CalculatedDebit float64
		CalculatedCredit float64
	}
	
	db.Raw(`
		SELECT 
			je.id as journal_id,
			je.code,
			je.entry_date::text,
			je.total_debit as stored_debit,
			je.total_credit as stored_credit,
			COALESCE(SUM(jl.debit_amount), 0) as calculated_debit,
			COALESCE(SUM(jl.credit_amount), 0) as calculated_credit
		FROM journal_entries je
		LEFT JOIN journal_lines jl ON je.id = jl.journal_entry_id
		WHERE je.status = 'POSTED'
		GROUP BY je.id, je.code, je.entry_date, je.total_debit, je.total_credit
		HAVING 
			ABS(je.total_debit - COALESCE(SUM(jl.debit_amount), 0)) > 0.01 OR
			ABS(je.total_credit - COALESCE(SUM(jl.credit_amount), 0)) > 0.01
		ORDER BY je.entry_date
	`).Scan(&integrityIssues)
	
	if len(integrityIssues) > 0 {
		fmt.Printf("❌ Found %d journal entries with integrity issues:\n\n", len(integrityIssues))
		for _, issue := range integrityIssues {
			fmt.Printf("  ID: %d | Code: %s | Date: %s\n", issue.JournalID, issue.Code, issue.EntryDate)
			fmt.Printf("  Stored:     DR %.2f | CR %.2f\n", issue.StoredDebit, issue.StoredCredit)
			fmt.Printf("  Calculated: DR %.2f | CR %.2f\n", issue.CalculatedDebit, issue.CalculatedCredit)
			fmt.Printf("  Difference: DR %.2f | CR %.2f\n", 
				issue.StoredDebit - issue.CalculatedDebit, 
				issue.StoredCredit - issue.CalculatedCredit)
			fmt.Println()
		}
	} else {
		fmt.Println("✅ All journal entries have matching totals")
	}
	
	// 7. Summary and Recommendations
	fmt.Println("\n📝 [7] SUMMARY & RECOMMENDATIONS")
	fmt.Println(strings.Repeat("=", 60))
	
	fmt.Println("\nISSUES FOUND:")
	issueCount := 0
	
	if diff != 0 {
		issueCount++
		fmt.Printf("  %d. Balance Sheet NOT balanced (diff: Rp %.2f)\n", issueCount, diff)
	}
	
	if len(unbalancedJournals) > 0 {
		issueCount++
		fmt.Printf("  %d. %d unbalanced journal entries\n", issueCount, len(unbalancedJournals))
	}
	
	if balanceDiff != 0 {
		issueCount++
		fmt.Printf("  %d. Laba Ditahan balance mismatch (diff: Rp %.2f)\n", issueCount, balanceDiff)
	}
	
	if len(negativeAccounts) > 0 {
		issueCount++
		fmt.Printf("  %d. %d accounts with negative balances\n", issueCount, len(negativeAccounts))
	}
	
	if len(integrityIssues) > 0 {
		issueCount++
		fmt.Printf("  %d. %d journal entries with integrity issues\n", issueCount, len(integrityIssues))
	}
	
	if len(closingEntries) > 1 {
		issueCount++
		fmt.Printf("  %d. Multiple closing entries detected (%d entries)\n", issueCount, len(closingEntries))
	}
	
	if issueCount == 0 {
		fmt.Println("  ✅ No major issues detected")
	}
	
	fmt.Println("\nRECOMMENDED ACTIONS:")
	fmt.Println("  1. Review all closing entries - may have duplicate closings")
	fmt.Println("  2. Check Laba Ditahan transactions for double posting")
	fmt.Println("  3. Investigate negative account balances")
	fmt.Println("  4. Run fix_double_posting.ps1 -Action analyze for detailed SQL analysis")
	fmt.Println("  5. Consider voiding duplicate closing entries")
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("END OF DIAGNOSTIC REPORT")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
