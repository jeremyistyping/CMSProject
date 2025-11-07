package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
)

type JournalLineDetail struct {
	JournalID     uint
	JournalCode   string
	EntryDate     string
	Status        string
	AccountCode   string
	AccountName   string
	DebitAmount   float64
	CreditAmount  float64
	BalanceChange float64
}

func main() {
	log.Println("Analyzing Retained Earnings transactions...")
	
	db := database.ConnectDB()
	
	// Get all journal lines affecting Retained Earnings (3201)
	query := `
		SELECT 
			je.id as journal_id,
			je.code as journal_code,
			je.entry_date,
			je.status,
			a.code as account_code,
			a.name as account_name,
			jl.debit_amount,
			jl.credit_amount,
			CASE 
				WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') 
					THEN jl.credit_amount - jl.debit_amount
				ELSE jl.debit_amount - jl.credit_amount
			END as balance_change
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		JOIN accounts a ON a.id = jl.account_id
		WHERE a.code = '3201'
		AND je.deleted_at IS NULL
		ORDER BY je.entry_date, je.id, jl.id
	`
	
	var lines []JournalLineDetail
	if err := db.Raw(query).Scan(&lines).Error; err != nil {
		log.Fatalf("Failed to query journal lines: %v", err)
	}
	
	fmt.Printf("\n=== ALL TRANSACTIONS AFFECTING RETAINED EARNINGS (3201) ===\n\n")
	
	if len(lines) == 0 {
		fmt.Println("❌ No transactions found affecting Retained Earnings!")
		fmt.Println("\nThis explains why balance is stuck at 11.500.000")
		fmt.Println("The closing journal was created but balances were NOT updated!")
	} else {
		runningBalance := 0.0
		for i, line := range lines {
			fmt.Printf("%d. Journal #%d (%s) - %s [%s]\n", 
				i+1, line.JournalID, line.JournalCode, line.EntryDate[:10], line.Status)
			fmt.Printf("   Debit: %.2f, Credit: %.2f\n", line.DebitAmount, line.CreditAmount)
			fmt.Printf("   Balance Change: %.2f\n", line.BalanceChange)
			runningBalance += line.BalanceChange
			fmt.Printf("   Running Balance: %.2f\n\n", runningBalance)
		}
		
		fmt.Printf("=== SUMMARY ===\n")
		fmt.Printf("Total Transactions: %d\n", len(lines))
		fmt.Printf("Calculated Balance: %.2f\n", runningBalance)
	}
	
	// Check current balance in accounts table
	var currentBalance float64
	db.Raw("SELECT balance FROM accounts WHERE code = '3201'").Scan(&currentBalance)
	fmt.Printf("Actual Balance in accounts table: %.2f\n", currentBalance)
	
	// Check if there's a mismatch
	if len(lines) > 0 {
		calculatedBalance := 0.0
		for _, line := range lines {
			calculatedBalance += line.BalanceChange
		}
		
		diff := currentBalance - calculatedBalance
		if diff != 0 {
			fmt.Printf("\n❌ MISMATCH DETECTED!\n")
			fmt.Printf("   Calculated from journals: %.2f\n", calculatedBalance)
			fmt.Printf("   Actual in accounts table: %.2f\n", currentBalance)
			fmt.Printf("   Difference: %.2f\n", diff)
			fmt.Println("\n💡 This suggests the closing journal was created but account balances were NOT updated!")
		} else {
			fmt.Printf("\n✅ Balance matches journal transactions\n")
		}
	}
}
