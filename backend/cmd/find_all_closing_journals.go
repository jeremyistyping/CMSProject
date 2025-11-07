package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
)

type JournalEntry struct {
	ID            uint
	Code          string
	Description   string
	EntryDate     string
	ReferenceType string
	Status        string
	TotalDebit    float64
	TotalCredit   float64
	IsBalanced    bool
}

type JournalLine struct {
	AccountCode  string
	AccountName  string
	DebitAmount  float64
	CreditAmount float64
	Description  string
}

func main() {
	db := database.ConnectDB()
	
	fmt.Println("\n=== SEARCHING FOR ALL CLOSING-RELATED JOURNAL ENTRIES ===\n")
	
	// 1. Search by code pattern "CLO"
	var closingJournals []JournalEntry
	db.Raw(`
		SELECT 
			id,
			code,
			description,
			entry_date::text,
			reference_type,
			status,
			total_debit,
			total_credit,
			is_balanced
		FROM journal_entries
		WHERE code LIKE 'CLO%' OR description LIKE '%closing%' OR description LIKE '%Closing%'
		ORDER BY entry_date, id
	`).Scan(&closingJournals)
	
	fmt.Printf("Found %d journal entries with closing pattern:\n\n", len(closingJournals))
	
	for i, je := range closingJournals {
		fmt.Printf("%d. ID: %d | Code: %s | Date: %s\n", i+1, je.ID, je.Code, je.EntryDate)
		fmt.Printf("   Reference Type: %s | Status: %s\n", je.ReferenceType, je.Status)
		fmt.Printf("   Description: %s\n", je.Description)
		fmt.Printf("   Total Debit: %.2f | Total Credit: %.2f | Balanced: %v\n", 
			je.TotalDebit, je.TotalCredit, je.IsBalanced)
		
		// Get journal lines
		var lines []JournalLine
		db.Raw(`
			SELECT 
				a.code as account_code,
				a.name as account_name,
				jl.debit_amount,
				jl.credit_amount,
				jl.description
			FROM journal_lines jl
			JOIN accounts a ON jl.account_id = a.id
			WHERE jl.journal_entry_id = ?
			ORDER BY jl.line_number
		`, je.ID).Scan(&lines)
		
		fmt.Println("   Lines:")
		for _, line := range lines {
			if line.DebitAmount > 0 {
				fmt.Printf("     DR %s (%s): %.2f - %s\n", line.AccountCode, line.AccountName, line.DebitAmount, line.Description)
			}
			if line.CreditAmount > 0 {
				fmt.Printf("     CR %s (%s): %.2f - %s\n", line.AccountCode, line.AccountName, line.CreditAmount, line.Description)
			}
		}
		fmt.Println()
	}
	
	// 2. Check all revenue/expense accounts balance
	fmt.Println("\n=== REVENUE/EXPENSE ACCOUNTS CURRENT BALANCE ===\n")
	
	type AccountBalance struct {
		Code    string
		Name    string
		Type    string
		Balance float64
	}
	
	var tempAccounts []AccountBalance
	db.Raw(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE is_active = true 
			AND COALESCE(is_header, false) = false
			AND type IN ('REVENUE', 'EXPENSE')
		ORDER BY type, code
	`).Scan(&tempAccounts)
	
	for _, acc := range tempAccounts {
		fmt.Printf("%s - %s (%s): %.2f\n", acc.Code, acc.Name, acc.Type, acc.Balance)
	}
	
	// 3. Check actual balance calculation
	fmt.Println("\n=== RECALCULATE ALL ACCOUNT BALANCES FROM JOURNAL LINES ===\n")
	
	type RecalcBalance struct {
		AccountCode     string
		AccountName     string
		AccountType     string
		CurrentBalance  float64
		CalculatedDebit float64
		CalculatedCredit float64
		NetCalculated   float64
		Difference      float64
	}
	
	var recalc []RecalcBalance
	db.Raw(`
		WITH account_calcs AS (
			SELECT 
				a.code,
				a.name,
				a.type,
				a.balance as current_balance,
				COALESCE(SUM(jl.debit_amount), 0) as calculated_debit,
				COALESCE(SUM(jl.credit_amount), 0) as calculated_credit
			FROM accounts a
			LEFT JOIN journal_lines jl ON a.id = jl.account_id
			LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id AND je.status = 'POSTED'
			WHERE a.is_active = true 
				AND COALESCE(a.is_header, false) = false
			GROUP BY a.id, a.code, a.name, a.type, a.balance
		)
		SELECT 
			code as account_code,
			name as account_name,
			type as account_type,
			current_balance,
			calculated_debit,
			calculated_credit,
			CASE 
				WHEN type IN ('ASSET', 'EXPENSE') THEN calculated_debit - calculated_credit
				ELSE calculated_credit - calculated_debit
			END as net_calculated,
			current_balance - CASE 
				WHEN type IN ('ASSET', 'EXPENSE') THEN calculated_debit - calculated_credit
				ELSE calculated_credit - calculated_debit
			END as difference
		FROM account_calcs
		WHERE ABS(current_balance - CASE 
				WHEN type IN ('ASSET', 'EXPENSE') THEN calculated_debit - calculated_credit
				ELSE calculated_credit - calculated_debit
			END) > 0.01
		ORDER BY type, code
	`).Scan(&recalc)
	
	if len(recalc) > 0 {
		fmt.Printf("❌ Found %d accounts with balance mismatch:\n\n", len(recalc))
		for _, r := range recalc {
			fmt.Printf("%s - %s (%s)\n", r.AccountCode, r.AccountName, r.AccountType)
			fmt.Printf("  Current Balance: %.2f\n", r.CurrentBalance)
			fmt.Printf("  Calculated:      DR %.2f - CR %.2f = %.2f\n", 
				r.CalculatedDebit, r.CalculatedCredit, r.NetCalculated)
			fmt.Printf("  Difference:      %.2f\n\n", r.Difference)
		}
		
		totalDiff := 0.0
		for _, r := range recalc {
			totalDiff += r.Difference
		}
		fmt.Printf("TOTAL BALANCE DIFFERENCE: %.2f\n", totalDiff)
	} else {
		fmt.Println("✅ All accounts match calculated balances from journal lines")
	}
}
