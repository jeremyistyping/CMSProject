package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type LedgerRow struct {
	JournalID    uint      `gorm:"column:journal_id"`
	EntryNumber  string    `gorm:"column:entry_number"`
	EntryDate    time.Time `gorm:"column:entry_date"`
	Description  string    `gorm:"column:description"`
	Reference    string    `gorm:"column:reference"`
	AccountID    uint      `gorm:"column:account_id"`
	AccountCode  string    `gorm:"column:account_code"`
	AccountName  string    `gorm:"column:account_name"`
	DebitAmount  float64   `gorm:"column:debit_amount"`
	CreditAmount float64   `gorm:"column:credit_amount"`
	Status       string    `gorm:"column:status"`
	SourceType   string    `gorm:"column:source_type"`
}

func main() {
	fmt.Println("🔍 Testing General Ledger Query")
	fmt.Println("================================")
	fmt.Println("")

	db, err := connectDB()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Exact same query as General Ledger uses
	query := `
		SELECT 
			sje.id as journal_id,
			sje.entry_number,
			sje.entry_date,
			sjl.description,
			sje.reference,
			sjl.account_id,
			a.code as account_code,
			a.name as account_name,
			sjl.debit_amount,
			sjl.credit_amount,
			sje.status,
			sje.source_type
		FROM unified_journal_lines sjl
		JOIN unified_journal_ledger sje ON sje.id = sjl.journal_id
		LEFT JOIN accounts a ON a.id = sjl.account_id
		WHERE sje.status = ? 
			AND sje.entry_date BETWEEN ? AND ?
		ORDER BY sje.entry_date, sje.id, sjl.line_number
	`

	startDate := "2025-01-01"
	endDate := "2025-12-31"
	status := "POSTED"

	fmt.Printf("Query parameters:\n")
	fmt.Printf("  Status: %s\n", status)
	fmt.Printf("  Start Date: %s\n", startDate)
	fmt.Printf("  End Date: %s\n", endDate)
	fmt.Println("")

	var rows []LedgerRow
	err = db.Raw(query, status, startDate, endDate).Scan(&rows).Error
	
	if err != nil {
		fmt.Printf("❌ ERROR: %v\n", err)
		return
	}

	fmt.Printf("✅ Query executed successfully!\n")
	fmt.Printf("📊 Rows returned: %d\n\n", len(rows))

	if len(rows) == 0 {
		fmt.Println("⚠️  No rows returned. Checking why...")
		fmt.Println("")

		// Check each table individually
		var journalCount, lineCount, accountCount int64
		
		db.Table("unified_journal_ledger").
			Where("status = ? AND entry_date BETWEEN ? AND ?", status, startDate, endDate).
			Count(&journalCount)
		
		db.Table("unified_journal_lines").Count(&lineCount)
		db.Table("accounts").Count(&accountCount)

		fmt.Printf("Table checks:\n")
		fmt.Printf("  unified_journal_ledger (matching criteria): %d\n", journalCount)
		fmt.Printf("  unified_journal_lines (total): %d\n", lineCount)
		fmt.Printf("  accounts (total): %d\n", accountCount)
		fmt.Println("")

		// Check journal_id in lines
		type JournalIDCheck struct {
			JournalID uint `gorm:"column:journal_id"`
		}
		var lineJournalIDs []JournalIDCheck
		db.Table("unified_journal_lines").
			Select("DISTINCT journal_id").
			Scan(&lineJournalIDs)

		fmt.Printf("  Journal IDs in unified_journal_lines: ")
		for _, id := range lineJournalIDs {
			fmt.Printf("%d ", id.JournalID)
		}
		fmt.Println("")

		// Check if those journal IDs exist in ledger
		for _, id := range lineJournalIDs {
			var exists int64
			db.Table("unified_journal_ledger").
				Where("id = ?", id.JournalID).
				Count(&exists)
			
			if exists == 0 {
				fmt.Printf("  ⚠️  Journal ID %d in lines but NOT in ledger!\n", id.JournalID)
			}
		}

	} else {
		fmt.Println("Transaction Details:")
		fmt.Println("--------------------")
		for _, row := range rows {
			fmt.Printf("\n[Entry %d] %s | %s\n", 
				row.JournalID, 
				row.EntryDate.Format("2006-01-02"),
				row.EntryNumber)
			fmt.Printf("  Account: %s - %s\n", row.AccountCode, row.AccountName)
			fmt.Printf("  Description: %s\n", row.Description)
			fmt.Printf("  Debit: %.2f | Credit: %.2f\n", row.DebitAmount, row.CreditAmount)
			fmt.Printf("  Source: %s | Status: %s\n", row.SourceType, row.Status)
		}
	}

	fmt.Println("\n================================")
	fmt.Println("✅ Test complete")
}

func connectDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

