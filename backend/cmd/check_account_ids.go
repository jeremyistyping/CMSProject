package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
)

type AccountDetail struct {
	ID        uint
	Code      string
	Name      string
	Balance   float64
	DeletedAt *string
}

func main() {
	log.Println("Checking account IDs used in closing journal...")
	
	db := database.ConnectDB()
	
	// Get all 3201 accounts
	var accounts []AccountDetail
	db.Raw(`
		SELECT id, code, name, balance, deleted_at::text as deleted_at
		FROM accounts
		WHERE code = '3201'
		ORDER BY id
	`).Scan(&accounts)
	
	fmt.Println("\n=== ALL ACCOUNTS WITH CODE 3201 ===")
	for _, acc := range accounts {
		deleted := ""
		if acc.DeletedAt != nil {
			deleted = " [DELETED]"
		}
		fmt.Printf("ID=%d: %s - %s, Balance=%.2f%s\n", 
			acc.ID, acc.Code, acc.Name, acc.Balance, deleted)
	}
	
	// Get account IDs used in closing journal
	type JournalLineInfo struct {
		AccountID uint
		AccountCode string
		AccountName string
		DebitAmount float64
		CreditAmount float64
	}
	
	var journalLines []JournalLineInfo
	db.Raw(`
		SELECT 
			jl.account_id,
			a.code as account_code,
			a.name as account_name,
			jl.debit_amount,
			jl.credit_amount
		FROM journal_lines jl
		JOIN journal_entries je ON je.id = jl.journal_entry_id
		JOIN accounts a ON a.id = jl.account_id
		WHERE je.code = 'CLO-2025-10-12-31'
		ORDER BY jl.line_number
	`).Scan(&journalLines)
	
	fmt.Println("\n=== CLOSING JOURNAL LINES (CLO-2025-10-12-31) ===")
	for i, line := range journalLines {
		fmt.Printf("%d. Account ID=%d (%s - %s)\n", 
			i+1, line.AccountID, line.AccountCode, line.AccountName)
		fmt.Printf("   Debit=%.2f, Credit=%.2f\n", 
			line.DebitAmount, line.CreditAmount)
	}
	
	// Check which account ID should have been updated
	fmt.Println("\n=== ANALYSIS ===")
	for _, line := range journalLines {
		if line.AccountCode == "3201" {
			fmt.Printf("Closing journal used Account ID=%d for Retained Earnings\n", line.AccountID)
			
			// Check if this ID has the balance
			for _, acc := range accounts {
				if acc.ID == line.AccountID {
					if acc.DeletedAt != nil {
						fmt.Printf("❌ WARNING: Account ID=%d is DELETED but was used in closing journal!\n", acc.ID)
					}
					fmt.Printf("Account ID=%d current balance: %.2f\n", acc.ID, acc.Balance)
				}
			}
		}
	}
}
