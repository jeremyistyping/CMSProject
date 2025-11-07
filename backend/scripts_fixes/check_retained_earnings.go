package main

import (
	"fmt"
	"log"
	
	"app-sistem-akuntansi/database"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize database
	db := database.ConnectDB()

	fmt.Println("=== CHECKING RETAINED EARNINGS ACCOUNT (3201) ===\n")

	// 1. Check if account 3201 exists
	type Account struct {
		ID      uint
		Code    string
		Name    string
		Type    string
		Balance float64
		IsActive bool
		IsHeader bool
	}
	
	var account Account
	err := db.Raw(`SELECT id, code, name, type, balance, is_active, COALESCE(is_header, false) as is_header 
		FROM accounts WHERE code = '3201'`).Scan(&account).Error
	
	if err != nil {
		fmt.Printf("❌ Error querying account 3201: %v\n", err)
		return
	}
	
	if account.ID == 0 {
		fmt.Println("❌ Account 3201 (Retained Earnings) NOT FOUND in database!")
		fmt.Println("\nSuggestion: Create account 3201 with:")
		fmt.Println("  Code: 3201")
		fmt.Println("  Name: LABA DITAHAN")
		fmt.Println("  Type: EQUITY")
		return
	}
	
	fmt.Printf("✅ Account 3201 EXISTS:\n")
	fmt.Printf("   ID: %d\n", account.ID)
	fmt.Printf("   Code: %s\n", account.Code)
	fmt.Printf("   Name: %s\n", account.Name)
	fmt.Printf("   Type: %s\n", account.Type)
	fmt.Printf("   Balance: Rp %.2f\n", account.Balance)
	fmt.Printf("   Is Active: %t\n", account.IsActive)
	fmt.Printf("   Is Header: %t\n\n", account.IsHeader)

	// 2. Check SSOT journal entries for account 3201
	type JournalEntry struct {
		ID          uint
		EntryNumber string
		EntryDate   string
		Description string
		DebitAmount float64
		CreditAmount float64
		Status      string
	}
	
	var ssotEntries []JournalEntry
	db.Raw(`
		SELECT 
			uje.id,
			uje.entry_number,
			uje.entry_date,
			uje.description,
			ujl.debit_amount,
			ujl.credit_amount,
			uje.status
		FROM unified_journal_ledger uje
		JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
		WHERE ujl.account_id = ?
		  AND uje.deleted_at IS NULL
		ORDER BY uje.entry_date DESC
	`, account.ID).Scan(&ssotEntries)
	
	fmt.Printf("=== SSOT JOURNAL ENTRIES FOR ACCOUNT 3201 ===\n")
	if len(ssotEntries) == 0 {
		fmt.Println("❌ NO SSOT journal entries found for account 3201!")
	} else {
		fmt.Printf("Found %d SSOT journal entries:\n\n", len(ssotEntries))
		totalDebit := 0.0
		totalCredit := 0.0
		for i, entry := range ssotEntries {
			fmt.Printf("%d. Entry #%s (Date: %s, Status: %s)\n", i+1, entry.EntryNumber, entry.EntryDate, entry.Status)
			fmt.Printf("   Description: %s\n", entry.Description)
			fmt.Printf("   Debit: Rp %.2f, Credit: Rp %.2f\n", entry.DebitAmount, entry.CreditAmount)
			if entry.Status == "POSTED" {
				totalDebit += entry.DebitAmount
				totalCredit += entry.CreditAmount
			}
		}
		fmt.Printf("\nTotal POSTED entries: Debit: Rp %.2f, Credit: Rp %.2f\n", totalDebit, totalCredit)
		fmt.Printf("Net Balance (Credit - Debit): Rp %.2f\n\n", totalCredit - totalDebit)
	}

	// 3. Check Legacy journal entries for account 3201
	var legacyEntries []JournalEntry
	db.Raw(`
		SELECT 
			je.id,
			je.entry_number,
			je.entry_date,
			je.description,
			jl.debit_amount,
			jl.credit_amount,
			je.status
		FROM journal_entries je
		JOIN journal_lines jl ON jl.journal_entry_id = je.id
		WHERE jl.account_id = ?
		  AND je.deleted_at IS NULL
		ORDER BY je.entry_date DESC
	`, account.ID).Scan(&legacyEntries)
	
	fmt.Printf("=== LEGACY JOURNAL ENTRIES FOR ACCOUNT 3201 ===\n")
	if len(legacyEntries) == 0 {
		fmt.Println("❌ NO legacy journal entries found for account 3201!")
	} else {
		fmt.Printf("Found %d legacy journal entries:\n\n", len(legacyEntries))
		totalDebit := 0.0
		totalCredit := 0.0
		for i, entry := range legacyEntries {
			fmt.Printf("%d. Entry #%s (Date: %s, Status: %s)\n", i+1, entry.EntryNumber, entry.EntryDate, entry.Status)
			fmt.Printf("   Description: %s\n", entry.Description)
			fmt.Printf("   Debit: Rp %.2f, Credit: Rp %.2f\n", entry.DebitAmount, entry.CreditAmount)
			if entry.Status == "POSTED" {
				totalDebit += entry.DebitAmount
				totalCredit += entry.CreditAmount
			}
		}
		fmt.Printf("\nTotal POSTED entries: Debit: Rp %.2f, Credit: Rp %.2f\n", totalDebit, totalCredit)
		fmt.Printf("Net Balance (Credit - Debit): Rp %.2f\n\n", totalCredit - totalDebit)
	}

	// 4. Check if account would appear in balance sheet query
	fmt.Println("=== CHECKING BALANCE SHEET QUERY ===")
	type BSAccount struct {
		AccountID   uint
		AccountCode string
		AccountName string
		AccountType string
		DebitTotal  float64
		CreditTotal float64
		NetBalance  float64
	}
	
	var bsAccount BSAccount
	query := `
		SELECT 
			MIN(a.id) as account_id,
			a.code as account_code,
			MAX(a.name) as account_name,
			UPPER(a.type) as account_type,
			COALESCE(SUM(ujl.debit_amount), 0) as debit_total,
			COALESCE(SUM(ujl.credit_amount), 0) as credit_total,
			CASE 
				WHEN UPPER(a.type) IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE 
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as net_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id 
			AND uje.status = 'POSTED' 
			AND uje.deleted_at IS NULL 
			AND uje.entry_date <= '2025-12-31'
		WHERE a.code = '3201'
		  AND COALESCE(a.is_header, false) = false
		  AND a.is_active = true
		GROUP BY a.code, UPPER(a.type)
	`
	
	err = db.Raw(query).Scan(&bsAccount).Error
	if err != nil {
		fmt.Printf("❌ Error in balance sheet query: %v\n", err)
	} else if bsAccount.AccountID == 0 {
		fmt.Println("❌ Account 3201 NOT included in balance sheet query results!")
		fmt.Println("\nPossible reasons:")
		fmt.Println("  1. No journal entries exist (HAVING clause filters out zero balances)")
		fmt.Println("  2. Account is_header = true")
		fmt.Println("  3. Account is_active = false")
	} else {
		fmt.Printf("✅ Account 3201 WOULD appear in balance sheet:\n")
		fmt.Printf("   Account ID: %d\n", bsAccount.AccountID)
		fmt.Printf("   Account Code: %s\n", bsAccount.AccountCode)
		fmt.Printf("   Account Name: %s\n", bsAccount.AccountName)
		fmt.Printf("   Account Type: %s\n", bsAccount.AccountType)
		fmt.Printf("   Debit Total: Rp %.2f\n", bsAccount.DebitTotal)
		fmt.Printf("   Credit Total: Rp %.2f\n", bsAccount.CreditTotal)
		fmt.Printf("   Net Balance: Rp %.2f\n", bsAccount.NetBalance)
	}

	fmt.Println("\n=== SUMMARY ===")
	if account.Balance == 0 && len(ssotEntries) == 0 && len(legacyEntries) == 0 {
		fmt.Println("⚠️  ISSUE CONFIRMED: Account 3201 exists but has NO journal entries!")
		fmt.Println("    This means Period Closing did NOT create journal entry for Retained Earnings.")
		fmt.Println("\n🔧 SOLUTION: Period Closing Journal needs to be fixed to include Retained Earnings.")
	} else if account.Balance != 0 && len(ssotEntries) == 0 {
		fmt.Printf("⚠️  Account 3201 has balance (%.2f) but NO SSOT journal entries!\n", account.Balance)
		fmt.Println("    Balance might be from legacy system or manual adjustment.")
	} else {
		fmt.Println("✅ Account 3201 has journal entries. Check if they sum up correctly.")
	}
}
