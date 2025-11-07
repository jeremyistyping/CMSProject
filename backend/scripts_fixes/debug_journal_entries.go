package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("=== Journal Entries Debug Test ===")

	// Initialize database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	fmt.Println("✅ Connected to database successfully")

	// Check what tables exist
	fmt.Println("\n🔍 Checking available tables...")
	var tables []string
	db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE '%journal%' ORDER BY table_name").Scan(&tables)

	fmt.Printf("📋 Found %d journal-related tables:\n", len(tables))
	for _, table := range tables {
		fmt.Printf("  - %s\n", table)
	}

	// Check unified_journal_ledger table
	fmt.Println("\n📊 Checking unified_journal_ledger...")
	var ledgerCount int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&ledgerCount)
	fmt.Printf("Total entries in unified_journal_ledger: %d\n", ledgerCount)

	if ledgerCount > 0 {
		fmt.Println("Sample entries from unified_journal_ledger:")
		var sampleEntries []struct {
			ID          uint   `gorm:"column:id"`
			Status      string `gorm:"column:status"`
			EntryDate   string `gorm:"column:entry_date"`
			Description string `gorm:"column:description"`
		}
		db.Raw("SELECT id, status, entry_date, description FROM unified_journal_ledger ORDER BY entry_date DESC LIMIT 5").Scan(&sampleEntries)
		for _, entry := range sampleEntries {
			fmt.Printf("  - ID: %d, Status: %s, Date: %s, Desc: %s\n", entry.ID, entry.Status, entry.EntryDate, entry.Description)
		}
	}

	// Check unified_journal_lines table
	fmt.Println("\n📊 Checking unified_journal_lines...")
	var linesCount int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_lines").Scan(&linesCount)
	fmt.Printf("Total entries in unified_journal_lines: %d\n", linesCount)

	if linesCount > 0 {
		fmt.Println("Sample entries from unified_journal_lines:")
		var sampleLines []struct {
			ID           uint    `gorm:"column:id"`
			JournalID    uint    `gorm:"column:journal_id"`
			AccountID    uint    `gorm:"column:account_id"`
			DebitAmount  float64 `gorm:"column:debit_amount"`
			CreditAmount float64 `gorm:"column:credit_amount"`
		}
		db.Raw("SELECT id, journal_id, account_id, debit_amount, credit_amount FROM unified_journal_lines LIMIT 5").Scan(&sampleLines)
		for _, line := range sampleLines {
			fmt.Printf("  - ID: %d, JournalID: %d, AccountID: %d, Debit: %.2f, Credit: %.2f\n", 
				line.ID, line.JournalID, line.AccountID, line.DebitAmount, line.CreditAmount)
		}
	}

	// Check legacy journal_entries table
	fmt.Println("\n📊 Checking legacy journal_entries...")
	var legacyCount int64
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE status = 'POSTED'").Scan(&legacyCount)
	fmt.Printf("Total POSTED entries in journal_entries: %d\n", legacyCount)

	if legacyCount > 0 {
		fmt.Println("Sample POSTED entries from journal_entries:")
		var legacyEntries []struct {
			ID           uint    `gorm:"column:id"`
			AccountID    uint    `gorm:"column:account_id"`
			DebitAmount  float64 `gorm:"column:debit_amount"`
			CreditAmount float64 `gorm:"column:credit_amount"`
			Date         string  `gorm:"column:date"`
			Status       string  `gorm:"column:status"`
		}
		db.Raw("SELECT id, account_id, debit_amount, credit_amount, date, status FROM journal_entries WHERE status = 'POSTED' ORDER BY date DESC LIMIT 5").Scan(&legacyEntries)
		for _, entry := range legacyEntries {
			fmt.Printf("  - ID: %d, AccountID: %d, Debit: %.2f, Credit: %.2f, Date: %s, Status: %s\n", 
				entry.ID, entry.AccountID, entry.DebitAmount, entry.CreditAmount, entry.Date, entry.Status)
		}
	}

	// Check account balances table
	fmt.Println("\n📊 Checking account balances...")
	var accountsWithBalance []struct {
		ID      uint    `gorm:"column:id"`
		Code    string  `gorm:"column:code"`
		Name    string  `gorm:"column:name"`
		Balance float64 `gorm:"column:balance"`
	}
	db.Raw("SELECT id, code, name, balance FROM accounts WHERE balance != 0 ORDER BY ABS(balance) DESC LIMIT 10").Scan(&accountsWithBalance)
	
	fmt.Printf("Accounts with non-zero balances: %d\n", len(accountsWithBalance))
	for _, account := range accountsWithBalance {
		fmt.Printf("  - %s (%s): %.2f\n", account.Name, account.Code, account.Balance)
	}

	// Test the exact query used in Balance Sheet service
	fmt.Println("\n🔍 Testing Balance Sheet query...")
	testDate := "2025-01-15"
	
	var queryResults []struct {
		AccountID    uint    `gorm:"column:account_id"`
		AccountCode  string  `gorm:"column:account_code"`
		AccountName  string  `gorm:"column:account_name"`
		AccountType  string  `gorm:"column:account_type"`
		DebitTotal   float64 `gorm:"column:debit_total"`
		CreditTotal  float64 `gorm:"column:credit_total"`
		NetBalance   float64 `gorm:"column:net_balance"`
	}
	
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(ujl.debit_amount), 0) as debit_total,
			COALESCE(SUM(ujl.credit_amount), 0) as credit_total,
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE 
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as net_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE (uje.status = 'POSTED' AND uje.entry_date <= ?) OR uje.status IS NULL
		GROUP BY a.id, a.code, a.name, a.type
		HAVING a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
		ORDER BY a.code
		LIMIT 10
	`
	
	db.Raw(query, testDate).Scan(&queryResults)
	
	fmt.Printf("Balance Sheet query results: %d accounts found\n", len(queryResults))
	for _, result := range queryResults {
		if result.DebitTotal != 0 || result.CreditTotal != 0 {
			fmt.Printf("  - %s (%s) [%s]: Debit=%.2f, Credit=%.2f, Net=%.2f\n",
				result.AccountName, result.AccountCode, result.AccountType, result.DebitTotal, result.CreditTotal, result.NetBalance)
		}
	}
	
	// Alternative query using account balances directly
	fmt.Println("\n🔍 Testing alternative query with account.balance...")
	var altResults []struct {
		AccountCode  string  `gorm:"column:account_code"`
		AccountName  string  `gorm:"column:account_name"`
		AccountType  string  `gorm:"column:account_type"`
		Balance      float64 `gorm:"column:balance"`
	}
	
	altQuery := `
		SELECT 
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			a.balance
		FROM accounts a
		WHERE a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
		AND a.balance != 0
		ORDER BY ABS(a.balance) DESC
		LIMIT 10
	`
	
	db.Raw(altQuery).Scan(&altResults)
	
	fmt.Printf("Alternative query (account.balance): %d accounts with non-zero balance\n", len(altResults))
	for _, result := range altResults {
		fmt.Printf("  - %s (%s) [%s]: Balance=%.2f\n",
			result.AccountName, result.AccountCode, result.AccountType, result.Balance)
	}

	fmt.Println("\n✅ Journal entries debug test completed")
}