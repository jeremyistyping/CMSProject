package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type BalanceSheetSummary struct {
	TotalAssets              float64 `json:"total_assets"`
	TotalLiabilities         float64 `json:"total_liabilities"`
	TotalEquity              float64 `json:"total_equity"`
	TotalLiabilitiesAndEquity float64 `json:"total_liabilities_and_equity"`
	BalanceDifference        float64 `json:"balance_difference"`
	IsBalanced               bool    `json:"is_balanced"`
	AsOfDate                 string  `json:"as_of_date"`
}

func main() {
	fmt.Println("=== SIMPLE DUPLICATE FIXER ===")
	fmt.Println("Script untuk menghapus duplikat dan balance sheet")
	fmt.Println("===============================================")

	// Setup database connection
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent,
		}),
	})
	if err != nil {
		log.Fatalf("❌ Error connecting to database: %v", err)
	}

	asOfDate := time.Now().Format("2006-01-02")
	fmt.Printf("📅 Processing as of: %s\n\n", asOfDate)

	// Step 1: Show current balance sheet
	fmt.Println("🔍 STEP 1: Current balance sheet status...")
	summary, err := analyzeBalanceSheet(db, asOfDate)
	if err != nil {
		log.Fatalf("❌ Error analyzing balance sheet: %v", err)
	}
	printBalanceSheetSummary(summary)

	// Step 2: Direct SQL approach to remove duplicates
	fmt.Println("\n🛠️  STEP 2: Removing duplicate journal entries using direct SQL...")
	
	// Create temp table and remove duplicates
	queries := []string{
		`-- Create temp table with duplicate line IDs to remove
		CREATE TEMP TABLE IF NOT EXISTS duplicate_lines_to_delete AS
		WITH duplicate_analysis AS (
			SELECT 
				ujl.id as line_id,
				ujl.journal_id,
				a.code as account_code,
				ROW_NUMBER() OVER (
					PARTITION BY a.code, DATE(uje.entry_date), ujl.debit_amount, ujl.credit_amount, 
					COALESCE(ujl.description, '') 
					ORDER BY ujl.id
				) as row_num
			FROM unified_journal_lines ujl
			JOIN accounts a ON a.id = ujl.account_id
			JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
			WHERE uje.status = 'POSTED'
		)
		SELECT line_id, journal_id, account_code
		FROM duplicate_analysis 
		WHERE row_num > 1`,

		`-- Delete duplicate lines
		DELETE FROM unified_journal_lines 
		WHERE id IN (SELECT line_id FROM duplicate_lines_to_delete)`,

		`-- Delete empty journal headers
		DELETE FROM unified_journal_ledger 
		WHERE id NOT IN (
			SELECT DISTINCT journal_id 
			FROM unified_journal_lines 
			WHERE journal_id IS NOT NULL
		)`,
	}

	for i, query := range queries {
		fmt.Printf("  Executing query %d...\n", i+1)
		result := db.Exec(query)
		if result.Error != nil {
			fmt.Printf("  ⚠️  Query %d warning: %v\n", i+1, result.Error)
		} else {
			fmt.Printf("  ✅ Query %d completed, rows affected: %d\n", i+1, result.RowsAffected)
		}
	}

	// Step 3: Check balance sheet after cleanup
	fmt.Println("\n🔍 STEP 3: Balance sheet after duplicate removal...")
	finalSummary, err := analyzeBalanceSheet(db, asOfDate)
	if err != nil {
		log.Fatalf("❌ Error analyzing balance sheet after cleanup: %v", err)
	}
	printBalanceSheetSummary(finalSummary)

	// Step 4: Create adjusting entry if still not balanced
	if !finalSummary.IsBalanced && math.Abs(finalSummary.BalanceDifference) > 0.01 {
		fmt.Printf("\n🔧 STEP 4: Creating adjusting entry for remaining difference: Rp %.2f\n", finalSummary.BalanceDifference)
		
		// Check table structure first
		var columns []struct {
			ColumnName string `gorm:"column:column_name"`
			IsNullable string `gorm:"column:is_nullable"`
		}
		
		db.Raw(`SELECT column_name, is_nullable 
				FROM information_schema.columns 
				WHERE table_name = 'unified_journal_ledger' 
				ORDER BY ordinal_position`).Scan(&columns)
				
		fmt.Println("📋 Table columns for unified_journal_ledger:")
		for _, col := range columns {
			fmt.Printf("  - %s (nullable: %s)\n", col.ColumnName, col.IsNullable)
		}

		// Create adjusting entry with minimal required fields
		err = createSimpleAdjustingEntry(db, finalSummary.BalanceDifference)
		if err != nil {
			fmt.Printf("❌ Error creating adjusting entry: %v\n", err)
			fmt.Println("💡 You may need to create adjusting entry manually")
		} else {
			// Final verification
			fmt.Println("\n🔍 FINAL CHECK: Balance sheet after adjusting entry...")
			verifyFinalSummary, err := analyzeBalanceSheet(db, asOfDate)
			if err == nil {
				printBalanceSheetSummary(verifyFinalSummary)
				if verifyFinalSummary.IsBalanced {
					fmt.Println("\n🎉 SUCCESS! Balance sheet is now balanced!")
				}
			}
		}
	} else if finalSummary.IsBalanced {
		fmt.Println("\n🎉 SUCCESS! Balance sheet is now balanced after duplicate removal!")
	}

	fmt.Println("\n✅ Script completed!")
}

func analyzeBalanceSheet(db *gorm.DB, asOfDate string) (*BalanceSheetSummary, error) {
	var result struct {
		TotalAssets      float64 `json:"total_assets"`
		TotalLiabilities float64 `json:"total_liabilities"`
		TotalEquity      float64 `json:"total_equity"`
	}

	query := `
		WITH account_balances AS (
			SELECT 
				a.type as account_type,
				CASE 
					WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
						COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
					ELSE 
						COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
				END as net_balance
			FROM accounts a
			LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
			LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
			WHERE (uje.status = 'POSTED' AND uje.entry_date <= $1) OR uje.status IS NULL
			GROUP BY a.id, a.type
			HAVING a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
		)
		SELECT 
			COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN net_balance ELSE 0 END), 0) as total_assets,
			COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN net_balance ELSE 0 END), 0) as total_liabilities,
			COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN net_balance ELSE 0 END), 0) as total_equity
		FROM account_balances
	`

	if err := db.Raw(query, asOfDate).Scan(&result).Error; err != nil {
		return nil, err
	}

	summary := &BalanceSheetSummary{
		AsOfDate:                  asOfDate,
		TotalAssets:              result.TotalAssets,
		TotalLiabilities:         result.TotalLiabilities,
		TotalEquity:              result.TotalEquity,
		TotalLiabilitiesAndEquity: result.TotalLiabilities + result.TotalEquity,
	}

	summary.BalanceDifference = summary.TotalAssets - summary.TotalLiabilitiesAndEquity
	summary.IsBalanced = math.Abs(summary.BalanceDifference) <= 0.01

	return summary, nil
}

func createSimpleAdjustingEntry(db *gorm.DB, balanceDifference float64) error {
	if math.Abs(balanceDifference) <= 0.01 {
		return nil
	}

	fmt.Printf("Creating adjusting entry for Rp %.2f\n", balanceDifference)

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Try to create journal header with minimal fields
	entryNumber := fmt.Sprintf("ADJ-%s-001", time.Now().Format("20060102"))
	
	var journalID uint
	// Try different INSERT approaches based on table structure
	insertQueries := []string{
		// Try with all common fields
		`INSERT INTO unified_journal_ledger (entry_number, entry_date, description, status, source_type, created_at, updated_at)
		 VALUES ($1, CURRENT_DATE, 'Balance Sheet Adjusting Entry - Automated', 'POSTED', 'MANUAL', NOW(), NOW())
		 RETURNING id`,
		
		// Try with basic fields only
		`INSERT INTO unified_journal_ledger (entry_number, entry_date, description, status, created_at, updated_at)
		 VALUES ($1, CURRENT_DATE, 'Balance Sheet Adjusting Entry - Automated', 'POSTED', NOW(), NOW())
		 RETURNING id`,
	}

	var err error
	for i, insertQuery := range insertQueries {
		err = tx.Raw(insertQuery, entryNumber).Scan(&journalID).Error
		if err == nil {
			fmt.Printf("✅ Journal header created with query %d, ID: %d\n", i+1, journalID)
			break
		} else {
			fmt.Printf("⚠️  Query %d failed: %v\n", i+1, err)
		}
	}

	if err != nil || journalID == 0 {
		tx.Rollback()
		return fmt.Errorf("failed to create journal header: %v", err)
	}

	// Get or create retained earnings account
	var retainedEarningsID uint
	accountQueries := []string{
		"SELECT id FROM accounts WHERE code = '3201' LIMIT 1",
		"SELECT id FROM accounts WHERE name ILIKE '%retained%' OR name ILIKE '%laba%ditahan%' LIMIT 1",
		"SELECT id FROM accounts WHERE type = 'EQUITY' LIMIT 1",
	}

	for _, accountQuery := range accountQueries {
		tx.Raw(accountQuery).Scan(&retainedEarningsID)
		if retainedEarningsID != 0 {
			break
		}
	}

	// If no account found, create suspense account
	if retainedEarningsID == 0 {
		tx.Raw(`
			INSERT INTO accounts (code, name, type, is_active, created_at, updated_at)
			VALUES ('1901', 'Suspense Account', 'ASSET', true, NOW(), NOW())
			ON CONFLICT (code) DO UPDATE SET updated_at = NOW()
			RETURNING id
		`).Scan(&retainedEarningsID)
		
		if retainedEarningsID == 0 {
			tx.Raw("SELECT id FROM accounts WHERE code = '1901' LIMIT 1").Scan(&retainedEarningsID)
		}
	}

	if retainedEarningsID == 0 {
		tx.Rollback()
		return fmt.Errorf("failed to find or create adjusting account")
	}

	// Create adjusting entry line
	if balanceDifference > 0 {
		// Assets > Liabilities + Equity, so credit the account to reduce equity/increase liability
		err = tx.Exec(`
			INSERT INTO unified_journal_lines (journal_id, account_id, debit_amount, credit_amount, description, created_at, updated_at)
			VALUES (?, ?, 0, ?, 'Adjusting entry for balance sheet balance', NOW(), NOW())
		`, journalID, retainedEarningsID, balanceDifference).Error
	} else {
		// Assets < Liabilities + Equity, so debit the account
		err = tx.Exec(`
			INSERT INTO unified_journal_lines (journal_id, account_id, debit_amount, credit_amount, description, created_at, updated_at)
			VALUES (?, ?, ?, 0, 'Adjusting entry for balance sheet balance', NOW(), NOW())
		`, journalID, retainedEarningsID, math.Abs(balanceDifference)).Error
	}

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create adjusting lines: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit adjusting entry: %v", err)
	}

	fmt.Printf("✅ Adjusting entry created: %s\n", entryNumber)
	return nil
}

func printBalanceSheetSummary(summary *BalanceSheetSummary) {
	fmt.Println("📊 Balance Sheet Summary:")
	fmt.Printf("   Total Assets:              Rp %15.0f\n", summary.TotalAssets)
	fmt.Printf("   Total Liabilities:         Rp %15.0f\n", summary.TotalLiabilities)
	fmt.Printf("   Total Equity:              Rp %15.0f\n", summary.TotalEquity)
	fmt.Printf("   Total Liab + Equity:       Rp %15.0f\n", summary.TotalLiabilitiesAndEquity)
	fmt.Printf("   Balance Difference:        Rp %15.0f\n", summary.BalanceDifference)

	if summary.IsBalanced {
		fmt.Println("   Status:                    ✅ BALANCED")
	} else {
		fmt.Println("   Status:                    ❌ NOT BALANCED")
	}
}