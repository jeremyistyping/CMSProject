package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type JournalEntry struct {
	ID          uint   `gorm:"primaryKey"`
	EntryDate   string `gorm:"column:entry_date"`
	Status      string `gorm:"column:status"`
	Description string `gorm:"column:description"`
	SourceType  string `gorm:"column:source_type"`
}

type JournalLine struct {
	ID           uint    `gorm:"primaryKey"`
	JournalID    uint    `gorm:"column:journal_id"`
	AccountID    uint    `gorm:"column:account_id"`
	DebitAmount  float64 `gorm:"column:debit_amount"`
	CreditAmount float64 `gorm:"column:credit_amount"`
	Description  string  `gorm:"column:description"`
}

type Account struct {
	ID      uint   `gorm:"primaryKey"`
	Code    string `gorm:"column:code"`
	Name    string `gorm:"column:name"`
	Type    string `gorm:"column:type"`
	Balance float64 `gorm:"column:balance"`
}

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=password dbname=accounting_system port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("=== DEBUG JOURNAL DATA ===")

	// Check unified_journal_ledger table
	var journalCount int64
	db.Table("unified_journal_ledger").Count(&journalCount)
	fmt.Printf("Total journal entries: %d\n", journalCount)

	if journalCount > 0 {
		var journals []JournalEntry
		db.Table("unified_journal_ledger").Limit(5).Find(&journals)
		fmt.Println("\nSample journal entries:")
		for _, j := range journals {
			fmt.Printf("ID: %d, Date: %s, Status: %s, Source: %s\n", j.ID, j.EntryDate, j.Status, j.SourceType)
		}
	}

	// Check unified_journal_lines table
	var lineCount int64
	db.Table("unified_journal_lines").Count(&lineCount)
	fmt.Printf("\nTotal journal lines: %d\n", lineCount)

	if lineCount > 0 {
		var lines []JournalLine
		db.Table("unified_journal_lines").Limit(5).Find(&lines)
		fmt.Println("\nSample journal lines:")
		for _, l := range lines {
			fmt.Printf("ID: %d, JournalID: %d, AccountID: %d, Debit: %.2f, Credit: %.2f\n", 
				l.ID, l.JournalID, l.AccountID, l.DebitAmount, l.CreditAmount)
		}
	}

	// Check accounts table
	var accountCount int64
	db.Table("accounts").Count(&accountCount)
	fmt.Printf("\nTotal accounts: %d\n", accountCount)

	if accountCount > 0 {
		var accounts []Account
		db.Table("accounts").Where("type IN (?)", []string{"ASSET", "LIABILITY", "EQUITY"}).Limit(10).Find(&accounts)
		fmt.Println("\nSample balance sheet accounts:")
		for _, a := range accounts {
			fmt.Printf("Code: %s, Name: %s, Type: %s, Balance: %.2f\n", a.Code, a.Name, a.Type, a.Balance)
		}
	}

	// Run the actual balance sheet query to see what it returns
	fmt.Println("\n=== BALANCE SHEET QUERY TEST ===")
	
	type AccountBalance struct {
		AccountID    uint    `gorm:"column:account_id"`
		AccountCode  string  `gorm:"column:account_code"`
		AccountName  string  `gorm:"column:account_name"`
		AccountType  string  `gorm:"column:account_type"`
		DebitTotal   float64 `gorm:"column:debit_total"`
		CreditTotal  float64 `gorm:"column:credit_total"`
		NetBalance   float64 `gorm:"column:net_balance"`
	}

	var balances []AccountBalance
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
		WHERE (uje.status = 'POSTED' AND uje.entry_date <= '2025-09-25') OR uje.status IS NULL
		GROUP BY a.id, a.code, a.name, a.type
		HAVING a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
		ORDER BY a.code
	`

	if err := db.Raw(query).Scan(&balances).Error; err != nil {
		fmt.Printf("Error running query: %v\n", err)
	} else {
		fmt.Printf("Query returned %d balance sheet accounts:\n", len(balances))
		for _, b := range balances {
			if b.DebitTotal != 0 || b.CreditTotal != 0 || b.NetBalance != 0 {
				fmt.Printf("  %s - %s (%s): Debit=%.2f, Credit=%.2f, Net=%.2f\n", 
					b.AccountCode, b.AccountName, b.AccountType, b.DebitTotal, b.CreditTotal, b.NetBalance)
			}
		}
	}

	// Check for POSTED journal entries specifically
	var postedCount int64
	db.Table("unified_journal_ledger").Where("status = 'POSTED'").Count(&postedCount)
	fmt.Printf("\nPOSTED journal entries: %d\n", postedCount)

	// Check journal entries with lines
	var entriesWithLines int64
	db.Table("unified_journal_ledger uje").
		Joins("INNER JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id").
		Where("uje.status = 'POSTED'").
		Count(&entriesWithLines)
	fmt.Printf("POSTED entries with journal lines: %d\n", entriesWithLines)
}