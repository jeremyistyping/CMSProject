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

type JournalEntry struct {
	ID          uint      `gorm:"column:id"`
	EntryNumber string    `gorm:"column:entry_number"`
	SourceType  string    `gorm:"column:source_type"`
	EntryDate   time.Time `gorm:"column:entry_date"`
	Description string    `gorm:"column:description"`
	Status      string    `gorm:"column:status"`
	TotalDebit  float64   `gorm:"column:total_debit"`
	TotalCredit float64   `gorm:"column:total_credit"`
}

type JournalLine struct {
	ID           uint    `gorm:"column:id"`
	JournalID    uint    `gorm:"column:journal_id"`
	AccountID    uint    `gorm:"column:account_id"`
	AccountCode  string  `gorm:"column:account_code"`
	AccountName  string  `gorm:"column:account_name"`
	Description  string  `gorm:"column:description"`
	DebitAmount  float64 `gorm:"column:debit_amount"`
	CreditAmount float64 `gorm:"column:credit_amount"`
}

func main() {
	fmt.Println("🔍 Checking Journal Entries in Database")
	fmt.Println("=========================================")
	fmt.Println("")

	db, err := connectDB()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Check unified_journal_ledger
	fmt.Println("📊 Checking unified_journal_ledger table:")
	fmt.Println("-----------------------------------------")
	
	var entries []JournalEntry
	result := db.Table("unified_journal_ledger").
		Where("deleted_at IS NULL").
		Order("entry_date DESC").
		Limit(10).
		Find(&entries)

	if result.Error != nil {
		log.Printf("Error querying journal entries: %v", result.Error)
	} else {
		fmt.Printf("Total entries found: %d\n\n", len(entries))
		
		if len(entries) > 0 {
			for _, entry := range entries {
				fmt.Printf("ID: %d | Entry: %s | Type: %s | Date: %s | Status: %s\n",
					entry.ID,
					entry.EntryNumber,
					entry.SourceType,
					entry.EntryDate.Format("2006-01-02"),
					entry.Status)
				fmt.Printf("   Description: %s\n", entry.Description)
				fmt.Printf("   Debit: %.2f | Credit: %.2f\n\n", entry.TotalDebit, entry.TotalCredit)
			}
		} else {
			fmt.Println("⚠️  No journal entries found in unified_journal_ledger")
		}
	}

	// Check unified_journal_lines
	fmt.Println("\n📋 Checking unified_journal_lines table:")
	fmt.Println("-----------------------------------------")
	
	var lines []JournalLine
	query := `
		SELECT 
			ujl.id,
			ujl.journal_id,
			ujl.account_id,
			a.code as account_code,
			a.name as account_name,
			ujl.description,
			ujl.debit_amount,
			ujl.credit_amount
		FROM unified_journal_lines ujl
		LEFT JOIN accounts a ON a.id = ujl.account_id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE ujl.deleted_at IS NULL
		  AND uje.deleted_at IS NULL
		ORDER BY ujl.journal_id, ujl.line_number
		LIMIT 20
	`
	
	result2 := db.Raw(query).Scan(&lines)
	if result2.Error != nil {
		log.Printf("Error querying journal lines: %v", result2.Error)
	} else {
		fmt.Printf("Total lines found: %d\n\n", len(lines))
		
		if len(lines) > 0 {
			currentJournalID := uint(0)
			for _, line := range lines {
				if line.JournalID != currentJournalID {
					fmt.Printf("\n--- Journal ID: %d ---\n", line.JournalID)
					currentJournalID = line.JournalID
				}
				fmt.Printf("  Line %d: %s (%s) | Debit: %.2f | Credit: %.2f\n",
					line.ID,
					line.AccountCode,
					line.AccountName,
					line.DebitAmount,
					line.CreditAmount)
			}
		} else {
			fmt.Println("⚠️  No journal lines found")
		}
	}

	// Check for specific date range
	fmt.Println("\n\n📅 Checking entries for 2025:")
	fmt.Println("-----------------------------")
	
	var count2025 int64
	db.Table("unified_journal_ledger").
		Where("entry_date >= ? AND entry_date <= ? AND status = ? AND deleted_at IS NULL",
			"2025-01-01", "2025-12-31", "POSTED").
		Count(&count2025)
	
	fmt.Printf("Entries in 2025 (POSTED status): %d\n", count2025)

	// Check all statuses
	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	db.Table("unified_journal_ledger").
		Select("status, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("status").
		Scan(&statusCounts)
	
	fmt.Println("\nEntries by status:")
	for _, sc := range statusCounts {
		fmt.Printf("  %s: %d\n", sc.Status, sc.Count)
	}

	fmt.Println("\n=========================================")
	fmt.Println("✅ Database check complete")
}

func connectDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	return db, err
}

