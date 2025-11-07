package main

import (
	"fmt"
	"strings"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println("DIAGNOSE UNIFIED JOURNAL & ACCOUNT BALANCES")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println()

	db := database.ConnectDB()

	// 1. Check unified journal entries
	fmt.Println("📊 [STEP 1] UNIFIED JOURNAL ENTRIES")
	fmt.Println(strings.Repeat("-", 100))
	
	type JournalEntry struct {
		ID          uint64
		EntryNumber string
		SourceType  string
		EntryDate   string
		Description string
		TotalDebit  float64
		TotalCredit float64
		Status      string
	}

	var journals []JournalEntry
	db.Raw(`
		SELECT 
			id, entry_number, source_type, entry_date, description,
			total_debit, total_credit, status
		FROM unified_journal_ledger
		WHERE deleted_at IS NULL
		ORDER BY entry_date, id
	`).Scan(&journals)

	fmt.Printf("Total Journal Entries: %d\n\n", len(journals))
	
	for _, j := range journals {
		fmt.Printf("ID: %d | %s | %s | %s\n", j.ID, j.EntryNumber, j.SourceType, j.EntryDate)
		fmt.Printf("   Desc: %s\n", j.Description)
		fmt.Printf("   Debit: Rp %15.2f | Credit: Rp %15.2f | Status: %s\n", 
			j.TotalDebit, j.TotalCredit, j.Status)
		
		// Show lines for this journal
		type JournalLine struct {
			AccountID   uint64
			AccountCode string
			AccountName string
			LineNumber  int
			Description string
			Debit       float64
			Credit      float64
		}
		
		var lines []JournalLine
		db.Raw(`
			SELECT 
				ujl.account_id,
				a.code as account_code,
				a.name as account_name,
				ujl.line_number,
				ujl.description,
				ujl.debit_amount as debit,
				ujl.credit_amount as credit
			FROM unified_journal_lines ujl
			JOIN accounts a ON a.id = ujl.account_id
			WHERE ujl.journal_id = ?
			ORDER BY ujl.line_number
		`, j.ID).Scan(&lines)
		
		for _, line := range lines {
			fmt.Printf("      L%d: %s (%s) | Debit: %12.2f | Credit: %12.2f | %s\n",
				line.LineNumber, line.AccountCode, line.AccountName,
				line.Debit, line.Credit, line.Description)
		}
		fmt.Println()
	}

	// 2. Check account balances calculated from journals
	fmt.Println("\n📊 [STEP 2] ACCOUNT BALANCES FROM JOURNALS")
	fmt.Println(strings.Repeat("-", 100))

	type AccountBalance struct {
		Code        string
		Name        string
		Type        string
		DBBalance   float64
		TotalDebit  float64
		TotalCredit float64
		CalcBalance float64
		Difference  float64
	}

	var balances []AccountBalance
	db.Raw(`
		SELECT 
			a.code,
			a.name,
			a.type,
			a.balance as db_balance,
			COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
			COALESCE(SUM(ujl.credit_amount), 0) as total_credit,
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as calc_balance,
			a.balance - CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as difference
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uj ON uj.id = ujl.journal_id 
			AND uj.status = 'POSTED' 
			AND uj.deleted_at IS NULL
		WHERE a.deleted_at IS NULL
			AND a.is_header = false
		GROUP BY a.id, a.code, a.name, a.type, a.balance
		HAVING ABS(a.balance) > 0.01 
			OR ABS(COALESCE(SUM(ujl.debit_amount), 0)) > 0.01
			OR ABS(COALESCE(SUM(ujl.credit_amount), 0)) > 0.01
		ORDER BY a.code
	`).Scan(&balances)

	fmt.Printf("%-6s | %-35s | %-10s | %15s | %15s | %15s | %15s | %15s\n",
		"Code", "Name", "Type", "DB Balance", "Total Debit", "Total Credit", "Calc Balance", "Difference")
	fmt.Println(strings.Repeat("-", 140))

	var totalDiff float64
	for _, b := range balances {
		marker := ""
		if b.Difference < -0.01 || b.Difference > 0.01 {
			marker = " ⚠️ MISMATCH"
			totalDiff += b.Difference
		}
		fmt.Printf("%-6s | %-35s | %-10s | %15.2f | %15.2f | %15.2f | %15.2f | %15.2f%s\n",
			b.Code, truncate(b.Name, 35), b.Type, b.DBBalance, b.TotalDebit, b.TotalCredit, 
			b.CalcBalance, b.Difference, marker)
	}
	
	fmt.Println(strings.Repeat("-", 140))
	fmt.Printf("Total Difference: Rp %15.2f\n\n", totalDiff)

	// 3. Check Balance Sheet Equation
	fmt.Println("\n📊 [STEP 3] BALANCE SHEET EQUATION CHECK")
	fmt.Println(strings.Repeat("-", 100))

	type BSTotal struct {
		TotalAssets      float64
		TotalLiabilities float64
		TotalEquity      float64
		TotalRevenue     float64
		TotalExpense     float64
	}

	var bsTotal BSTotal
	db.Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END), 0) as total_assets,
			COALESCE(SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END), 0) as total_liabilities,
			COALESCE(SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END), 0) as total_equity,
			COALESCE(SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END), 0) as total_expense
		FROM accounts
		WHERE deleted_at IS NULL
	`).Scan(&bsTotal)

	netIncome := bsTotal.TotalRevenue - bsTotal.TotalExpense
	leftSide := bsTotal.TotalAssets
	rightSide := bsTotal.TotalLiabilities + bsTotal.TotalEquity + netIncome
	diff := leftSide - rightSide

	fmt.Printf("Assets:       Rp %18.2f\n", bsTotal.TotalAssets)
	fmt.Printf("Liabilities:  Rp %18.2f\n", bsTotal.TotalLiabilities)
	fmt.Printf("Equity:       Rp %18.2f\n", bsTotal.TotalEquity)
	fmt.Printf("Revenue:      Rp %18.2f (temp)\n", bsTotal.TotalRevenue)
	fmt.Printf("Expense:      Rp %18.2f (temp)\n", bsTotal.TotalExpense)
	fmt.Printf("Net Income:   Rp %18.2f\n", netIncome)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Left Side (A):        Rp %18.2f\n", leftSide)
	fmt.Printf("Right Side (L+E+NI):  Rp %18.2f\n", rightSide)
	fmt.Printf("Difference:           Rp %18.2f ", diff)
	
	if diff < -0.01 || diff > 0.01 {
		fmt.Printf("❌ NOT BALANCED\n")
	} else {
		fmt.Printf("✅ BALANCED\n")
	}
	fmt.Println()

	// 4. Check for closing journal entries
	fmt.Println("\n📊 [STEP 4] CLOSING JOURNAL ENTRIES")
	fmt.Println(strings.Repeat("-", 100))

	var closingJournals []JournalEntry
	db.Raw(`
		SELECT 
			id, entry_number, source_type, entry_date, description,
			total_debit, total_credit, status
		FROM unified_journal_ledger
		WHERE source_type = 'CLOSING'
			AND deleted_at IS NULL
		ORDER BY entry_date DESC
	`).Scan(&closingJournals)

	if len(closingJournals) == 0 {
		fmt.Println("⚠️  No closing entries found")
	} else {
		fmt.Printf("Found %d closing entries:\n\n", len(closingJournals))
		for _, j := range closingJournals {
			fmt.Printf("ID: %d | %s | %s\n", j.ID, j.EntryNumber, j.EntryDate)
			fmt.Printf("   Desc: %s\n", j.Description)
			fmt.Printf("   Debit: Rp %15.2f | Credit: Rp %15.2f\n", j.TotalDebit, j.TotalCredit)
			
			// Show lines
			type ClosingLine struct {
				AccountCode string
				AccountName string
				Debit       float64
				Credit      float64
			}
			
			var lines []ClosingLine
			db.Raw(`
				SELECT 
					a.code as account_code,
					a.name as account_name,
					ujl.debit_amount as debit,
					ujl.credit_amount as credit
				FROM unified_journal_lines ujl
				JOIN accounts a ON a.id = ujl.account_id
				WHERE ujl.journal_id = ?
				ORDER BY ujl.line_number
			`, j.ID).Scan(&lines)
			
			for _, line := range lines {
				fmt.Printf("      %s (%s) | Dr: %12.2f | Cr: %12.2f\n",
					line.AccountCode, line.AccountName, line.Debit, line.Credit)
			}
			fmt.Println()
		}
	}

	// 5. Summary
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println("DIAGNOSIS SUMMARY")
	fmt.Println(strings.Repeat("=", 100))
	
	if totalDiff < -0.01 || totalDiff > 0.01 {
		fmt.Printf("❌ FOUND DISCREPANCY: Total difference of Rp %.2f between DB balance and calculated balance\n", totalDiff)
	} else {
		fmt.Println("✅ All account balances match journal entries")
	}
	
	if diff < -0.01 || diff > 0.01 {
		fmt.Printf("❌ BALANCE SHEET NOT BALANCED: Difference of Rp %.2f\n", diff)
	} else {
		fmt.Println("✅ Balance sheet is balanced")
	}
	
	fmt.Printf("\nTotal Journal Entries: %d\n", len(journals))
	fmt.Printf("Closing Entries: %d\n", len(closingJournals))
	fmt.Printf("Accounts with balances: %d\n", len(balances))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
