package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("=== Account Mapping and SSOT Investigation ===")
	fmt.Println()

	// Check all accounts to see the mapping
	fmt.Println("1. All Account Details:")
	var accounts []models.Account
	err := db.Order("code").Find(&accounts).Error
	if err != nil {
		log.Printf("Error fetching accounts: %v", err)
	} else {
		for _, account := range accounts {
			fmt.Printf("  Account ID: %d, Code: %s, Name: %s, Type: %s, Balance: %.2f\n", 
				account.ID, account.Code, account.Name, account.Type, account.Balance)
		}
	}
	fmt.Println()

	// Check the SSOT journal with correct case
	fmt.Println("2. SSOT Journal Entries with Different Source Types:")
	type SSOTJournal struct {
		ID          uint      `json:"id"`
		EntryNumber string    `json:"entry_number"`
		SourceType  string    `json:"source_type"`
		SourceID    *uint     `json:"source_id"`
		TotalDebit  float64   `json:"total_debit"`
		TotalCredit float64   `json:"total_credit"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
	}

	var ssotJournals []SSOTJournal
	err = db.Raw(`
		SELECT id, entry_number, source_type, source_id, total_debit, total_credit, status, created_at
		FROM unified_journal_ledger 
		WHERE source_type IN ('SALE', 'sale', 'Sale')
		ORDER BY created_at DESC
	`).Scan(&ssotJournals).Error
	if err != nil {
		log.Printf("Error fetching SSOT journals: %v", err)
	} else {
		for _, journal := range ssotJournals {
			sourceID := "N/A"
			if journal.SourceID != nil {
				sourceID = fmt.Sprintf("%d", *journal.SourceID)
			}
			fmt.Printf("  SSOT Journal ID: %d, Entry: %s, Source: %s(%s), Debit: %.2f, Credit: %.2f, Status: %s\n", 
				journal.ID, journal.EntryNumber, journal.SourceType, sourceID, journal.TotalDebit, journal.TotalCredit, journal.Status)
		}
	}
	fmt.Println()

	// Check SSOT journal lines with account names
	fmt.Println("3. SSOT Journal Lines with Account Names:")
	type SSOTJournalLineWithAccount struct {
		ID           uint    `json:"id"`
		JournalID    uint    `json:"journal_id"`
		AccountID    uint    `json:"account_id"`
		AccountCode  string  `json:"account_code"`
		AccountName  string  `json:"account_name"`
		Description  string  `json:"description"`
		DebitAmount  float64 `json:"debit_amount"`
		CreditAmount float64 `json:"credit_amount"`
	}

	var ssotLinesWithAccounts []SSOTJournalLineWithAccount
	err = db.Raw(`
		SELECT ujl.id, ujl.journal_id, ujl.account_id, a.code as account_code, a.name as account_name, 
		       ujl.description, ujl.debit_amount, ujl.credit_amount
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
		LEFT JOIN accounts a ON ujl.account_id = a.id
		WHERE uj.source_type IN ('SALE', 'sale', 'Sale')
		ORDER BY ujl.id
	`).Scan(&ssotLinesWithAccounts).Error
	if err != nil {
		log.Printf("Error fetching SSOT journal lines with accounts: %v", err)
	} else {
		for _, line := range ssotLinesWithAccounts {
			fmt.Printf("  Line ID: %d, Journal: %d, Account: %d (%s - %s), Debit: %.2f, Credit: %.2f, Desc: %s\n", 
				line.ID, line.JournalID, line.AccountID, line.AccountCode, line.AccountName, line.DebitAmount, line.CreditAmount, line.Description)
		}
	}
	fmt.Println()

	// Check if there's balance update mechanism
	fmt.Println("4. Checking Account Balance Updates from SSOT:")
	var balanceCheckResults []struct {
		AccountID        uint    `json:"account_id"`
		AccountCode      string  `json:"account_code"`
		AccountName      string  `json:"account_name"`
		CurrentBalance   float64 `json:"current_balance"`
		CalculatedBalance float64 `json:"calculated_balance"`
	}

	err = db.Raw(`
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.balance as current_balance,
			COALESCE(
				(SELECT SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
				 FROM unified_journal_lines ujl 
				 WHERE ujl.account_id = a.id),
				0
			) as calculated_balance
		FROM accounts a
		WHERE a.id IN (
			SELECT DISTINCT account_id 
			FROM unified_journal_lines
		)
		ORDER BY a.code
	`).Scan(&balanceCheckResults).Error

	if err != nil {
		log.Printf("Error calculating balances: %v", err)
	} else {
		fmt.Println("  Account Balance Comparison (Current vs SSOT-Calculated):")
		for _, result := range balanceCheckResults {
			difference := result.CurrentBalance - result.CalculatedBalance
			status := "✅ MATCH"
			if difference != 0 {
				status = "❌ MISMATCH"
			}
			fmt.Printf("    %s: %s (%s) - Current: %.2f, SSOT: %.2f, Diff: %.2f %s\n", 
				result.AccountCode, result.AccountName, status, result.CurrentBalance, result.CalculatedBalance, difference, status)
		}
	}

	fmt.Println("\n=== Investigation Complete ===")
}