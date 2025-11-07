package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("DATABASE SCHEMA CHECK")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	// Check accounts table columns
	fmt.Println("📊 Accounts Table Columns:")
	fmt.Println(string(make([]byte, 80)))
	
	type ColumnInfo struct {
		ColumnName string
		DataType   string
	}
	
	var accountColumns []ColumnInfo
	db.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'accounts'
		ORDER BY ordinal_position
	`).Scan(&accountColumns)

	for _, col := range accountColumns {
		fmt.Printf("  %-30s %s\n", col.ColumnName, col.DataType)
	}
	fmt.Println()

	// Check unified_journal_ledger table columns
	fmt.Println("📊 Unified Journal Ledger Table Columns:")
	fmt.Println(string(make([]byte, 80)))
	
	var journalColumns []ColumnInfo
	db.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'unified_journal_ledger'
		ORDER BY ordinal_position
	`).Scan(&journalColumns)

	for _, col := range journalColumns {
		fmt.Printf("  %-30s %s\n", col.ColumnName, col.DataType)
	}
	fmt.Println()

	// Sample data from accounts
	fmt.Println("📊 Sample Accounts Data:")
	fmt.Println(string(make([]byte, 80)))
	
	type Account struct {
		ID       uint
		Code     string
		Name     string
		Type     string
		Balance  float64
		IsHeader bool
	}
	
	var accounts []Account
	db.Raw(`SELECT id, code, name, type, balance, is_header FROM accounts LIMIT 5`).Scan(&accounts)
	
	for _, acc := range accounts {
		fmt.Printf("ID: %d | Code: %s | Name: %s | Type: %s | Balance: %.2f | Header: %v\n",
			acc.ID, acc.Code, acc.Name, acc.Type, acc.Balance, acc.IsHeader)
	}
	fmt.Println()

	// Sample data from unified_journal_ledger
	fmt.Println("📊 Sample Unified Journal Ledger Data:")
	fmt.Println(string(make([]byte, 80)))
	
	type JournalLine struct {
		ID           uint
		JournalID    uint
		AccountCode  string
		DebitAmount  float64
		CreditAmount float64
	}
	
	var lines []JournalLine
	db.Raw(`
		SELECT id, journal_id, account_code, debit_amount, credit_amount 
		FROM unified_journal_ledger 
		LIMIT 5
	`).Scan(&lines)
	
	for _, line := range lines {
		fmt.Printf("ID: %d | JournalID: %d | Account: %s | Dr: %.2f | Cr: %.2f\n",
			line.ID, line.JournalID, line.AccountCode, line.DebitAmount, line.CreditAmount)
	}
	fmt.Println()

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("✅ SCHEMA CHECK COMPLETE")
	fmt.Println("=" + string(make([]byte, 80)))
}

