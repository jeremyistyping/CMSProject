package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("=== Date Filtering Debug Test ===")

	// Initialize database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	fmt.Println("✅ Connected to database successfully")

	// Check dates in unified_journal_ledger
	fmt.Println("\n🔍 Checking dates in unified_journal_ledger...")
	var journalDates []struct {
		ID        uint   `gorm:"column:id"`
		EntryDate string `gorm:"column:entry_date"`
		Status    string `gorm:"column:status"`
	}
	db.Raw("SELECT id, entry_date, status FROM unified_journal_ledger ORDER BY entry_date").Scan(&journalDates)

	fmt.Printf("All journal entries dates:\n")
	for _, entry := range journalDates {
		fmt.Printf("  - ID: %d, Date: %s, Status: %s\n", entry.ID, entry.EntryDate, entry.Status)
	}

	// Test different date filters
	testDates := []string{
		"2025-01-15",
		"2025-09-26",
		"2025-12-31",
		"2024-01-01",
	}

	fmt.Println("\n🔍 Testing different date filters...")
	for _, testDate := range testDates {
		fmt.Printf("\n--- Testing date filter: %s ---\n", testDate)

		// Count entries that match the date filter
		var count int64
		db.Raw(`
			SELECT COUNT(*)
			FROM unified_journal_ledger uje
			WHERE uje.status = 'POSTED' AND uje.entry_date <= ?
		`, testDate).Scan(&count)

		fmt.Printf("POSTED entries <= %s: %d\n", testDate, count)

		if count > 0 {
			// Get account balances with this date filter
			var results []struct {
				AccountCode  string  `gorm:"column:account_code"`
				AccountName  string  `gorm:"column:account_name"`
				AccountType  string  `gorm:"column:account_type"`
				DebitTotal   float64 `gorm:"column:debit_total"`
				CreditTotal  float64 `gorm:"column:credit_total"`
				NetBalance   float64 `gorm:"column:net_balance"`
			}

			query := `
				SELECT 
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
				AND (COALESCE(SUM(ujl.debit_amount), 0) != 0 OR COALESCE(SUM(ujl.credit_amount), 0) != 0)
				ORDER BY ABS(COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)) DESC
				LIMIT 5
			`

			db.Raw(query, testDate).Scan(&results)

			fmt.Printf("  Accounts with journal activity:\n")
			for _, result := range results {
				fmt.Printf("    - %s (%s) [%s]: Debit=%.2f, Credit=%.2f, Net=%.2f\n",
					result.AccountName, result.AccountCode, result.AccountType,
					result.DebitTotal, result.CreditTotal, result.NetBalance)
			}
		}
	}

	// Check the actual current date format expected
	fmt.Println("\n🔍 Testing with today's date and different formats...")
	now := time.Now()
	dateFormats := []string{
		now.Format("2006-01-02"),
		now.Format("2006-01-02T15:04:05Z07:00"),
		now.Format("2006-01-02 15:04:05"),
	}

	for _, dateFormat := range dateFormats {
		var count int64
		db.Raw(`
			SELECT COUNT(*)
			FROM unified_journal_ledger uje
			WHERE uje.status = 'POSTED' AND uje.entry_date <= ?
		`, dateFormat).Scan(&count)

		fmt.Printf("Format '%s': %d entries\n", dateFormat, count)
	}

	// Show the exact problematic query from Balance Sheet service
	fmt.Println("\n🔍 Running exact Balance Sheet query with debug...")
	testDate := "2025-01-15"
	
	var debugResults []struct {
		AccountID    uint    `gorm:"column:account_id"`
		AccountCode  string  `gorm:"column:account_code"`
		AccountName  string  `gorm:"column:account_name"`
		AccountType  string  `gorm:"column:account_type"`
		DebitTotal   float64 `gorm:"column:debit_total"`
		CreditTotal  float64 `gorm:"column:credit_total"`
		NetBalance   float64 `gorm:"column:net_balance"`
	}

	exactQuery := `
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

	db.Raw(exactQuery, testDate).Scan(&debugResults)

	fmt.Printf("Exact Balance Sheet query results for %s:\n", testDate)
	for _, result := range debugResults {
		fmt.Printf("  - %s (%s) [%s]: Debit=%.2f, Credit=%.2f, Net=%.2f\n",
			result.AccountName, result.AccountCode, result.AccountType,
			result.DebitTotal, result.CreditTotal, result.NetBalance)
	}

	fmt.Println("\n✅ Date filtering debug test completed")
}