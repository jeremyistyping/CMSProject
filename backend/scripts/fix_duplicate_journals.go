package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DuplicateJournal represents duplicate journal entries
type DuplicateJournal struct {
	AccountCode   string    `json:"account_code"`
	AccountName   string    `json:"account_name"`
	EntryDate     time.Time `json:"entry_date"`
	DebitAmount   float64   `json:"debit_amount"`
	CreditAmount  float64   `json:"credit_amount"`
	Description   string    `json:"description"`
	Count         int       `json:"count"`
	JournalIDs    []uint    `json:"journal_ids"`
}

// BalanceSheetSummary represents balance sheet totals
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
	fmt.Println("=== DUPLICATE JOURNAL FIXER FOR SSOT SYSTEM ===")
	fmt.Println("Script untuk menghapus jurnal duplikat dan balance sheet")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Setup database connection
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("❌ Error connecting to database: %v", err)
	}

	// Get today's date
	asOfDate := time.Now().Format("2006-01-02")
	fmt.Printf("📅 Analyzing duplicates as of: %s\n\n", asOfDate)

	// Step 1: Analyze current balance sheet
	fmt.Println("🔍 STEP 1: Current balance sheet status...")
	summary, err := analyzeBalanceSheet(db, asOfDate)
	if err != nil {
		log.Fatalf("❌ Error analyzing balance sheet: %v", err)
	}
	printBalanceSheetSummary(summary)

	// Step 2: Find duplicate journal entries
	fmt.Println("\n🔍 STEP 2: Finding duplicate journal entries...")
	duplicates, err := findDuplicateJournalEntries(db)
	if err != nil {
		log.Fatalf("❌ Error finding duplicates: %v", err)
	}

	if len(duplicates) == 0 {
		fmt.Println("✅ No duplicate journal entries found")
		return
	}

	fmt.Printf("⚠️  Found %d types of duplicate entries:\n", len(duplicates))
	for i, dup := range duplicates {
		fmt.Printf("  %d. %s (%s) on %s\n", i+1, dup.AccountCode, dup.AccountName, dup.EntryDate.Format("2006-01-02"))
		fmt.Printf("     Amount: Debit=%.2f, Credit=%.2f, Count=%d\n", dup.DebitAmount, dup.CreditAmount, dup.Count)
		fmt.Printf("     Journal IDs: %v\n", dup.JournalIDs)
	}

	// Step 3: Remove duplicates
	fmt.Println("\n🛠️  STEP 3: Removing duplicate journal entries...")
	err = removeDuplicateEntries(db, duplicates)
	if err != nil {
		log.Fatalf("❌ Error removing duplicates: %v", err)
	}

	// Step 4: Verify balance sheet after fixes
	fmt.Println("\n🔍 STEP 4: Verifying balance sheet after fixes...")
	finalSummary, err := analyzeBalanceSheet(db, asOfDate)
	if err != nil {
		log.Fatalf("❌ Error verifying balance sheet: %v", err)
	}

	fmt.Println("\n📊 FINAL BALANCE SHEET STATUS:")
	printBalanceSheetSummary(finalSummary)

	// Step 5: Create adjusting entry if still not balanced
	if !finalSummary.IsBalanced && math.Abs(finalSummary.BalanceDifference) > 0.01 {
		fmt.Printf("\n🔧 STEP 5: Creating adjusting entry for remaining difference: Rp %.2f\n", finalSummary.BalanceDifference)
		err = createAdjustingEntry(db, finalSummary.BalanceDifference)
		if err != nil {
			log.Fatalf("❌ Error creating adjusting entry: %v", err)
		}

		// Final check
		fmt.Println("\n🔍 FINAL CHECK: Balance sheet after adjusting entry...")
		verifyFinalSummary, err := analyzeBalanceSheet(db, asOfDate)
		if err != nil {
			log.Fatalf("❌ Error in final verification: %v", err)
		}
		printBalanceSheetSummary(verifyFinalSummary)

		if verifyFinalSummary.IsBalanced {
			fmt.Println("\n🎉 SUCCESS! Balance sheet is now balanced!")
		} else {
			fmt.Printf("\n⚠️  Balance sheet still shows difference: Rp %.2f\n", verifyFinalSummary.BalanceDifference)
		}
	} else if finalSummary.IsBalanced {
		fmt.Println("\n🎉 SUCCESS! Balance sheet is now balanced!")
	}
}

func setupDatabase() (*gorm.DB, error) {
	// Database connection setup for PostgreSQL
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Silent,
			},
		),
	})

	return db, err
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

func findDuplicateJournalEntries(db *gorm.DB) ([]DuplicateJournal, error) {
	var duplicates []DuplicateJournal

	query := `
		WITH duplicate_lines AS (
			SELECT 
				a.code as account_code,
				a.name as account_name,
				DATE(uje.entry_date) as entry_date,
				ujl.debit_amount,
				ujl.credit_amount,
				ujl.description,
				COUNT(*) as count,
				ARRAY_AGG(ujl.id) as line_ids,
				ARRAY_AGG(ujl.journal_id) as journal_ids
			FROM unified_journal_lines ujl
			JOIN accounts a ON a.id = ujl.account_id
			JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
			WHERE uje.status = 'POSTED'
			GROUP BY a.code, a.name, DATE(uje.entry_date), ujl.debit_amount, ujl.credit_amount, ujl.description
			HAVING COUNT(*) > 1
		)
		SELECT 
			account_code,
			account_name,
			entry_date,
			debit_amount,
			credit_amount,
			description,
			count,
			journal_ids
		FROM duplicate_lines
		ORDER BY count DESC, entry_date DESC
	`

	var results []struct {
		AccountCode   string    `json:"account_code"`
		AccountName   string    `json:"account_name"`
		EntryDate     time.Time `json:"entry_date"`
		DebitAmount   float64   `json:"debit_amount"`
		CreditAmount  float64   `json:"credit_amount"`
		Description   string    `json:"description"`
		Count         int       `json:"count"`
		JournalIDsStr string    `json:"journal_ids"`
	}

	if err := db.Raw(query).Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, result := range results {
		// Parse journal IDs from PostgreSQL array format
		journalIDsStr := strings.Trim(result.JournalIDsStr, "{}")
		journalIDsStr = strings.ReplaceAll(journalIDsStr, " ", "")
		journalIDs := []uint{}

		if journalIDsStr != "" {
			idStrs := strings.Split(journalIDsStr, ",")
			for _, idStr := range idStrs {
				var id uint
				fmt.Sscanf(idStr, "%d", &id)
				journalIDs = append(journalIDs, id)
			}
		}

		duplicates = append(duplicates, DuplicateJournal{
			AccountCode:   result.AccountCode,
			AccountName:   result.AccountName,
			EntryDate:     result.EntryDate,
			DebitAmount:   result.DebitAmount,
			CreditAmount:  result.CreditAmount,
			Description:   result.Description,
			Count:         result.Count,
			JournalIDs:    journalIDs,
		})
	}

	return duplicates, nil
}

func removeDuplicateEntries(db *gorm.DB, duplicates []DuplicateJournal) error {
	if len(duplicates) == 0 {
		return nil
	}

	fmt.Printf("Removing %d types of duplicate entries...\n", len(duplicates))

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	totalRemoved := 0

	for _, duplicate := range duplicates {
		// Keep the first entry, remove the rest
		if len(duplicate.JournalIDs) <= 1 {
			continue
		}

		// Remove duplicate journal lines (keep the first one)
		journalIDsToRemove := duplicate.JournalIDs[1:] // Skip first entry

		for _, journalID := range journalIDsToRemove {
			// Remove journal lines for this journal ID that match the duplicate criteria
			result := tx.Exec(`
				DELETE FROM unified_journal_lines 
				WHERE journal_id = ? 
				AND account_id = (SELECT id FROM accounts WHERE code = ?)
				AND debit_amount = ? 
				AND credit_amount = ?
				AND description = ?
			`, journalID, duplicate.AccountCode, duplicate.DebitAmount, duplicate.CreditAmount, duplicate.Description)

			if result.Error != nil {
				tx.Rollback()
				return fmt.Errorf("failed to remove duplicate lines for journal %d: %v", journalID, result.Error)
			}

			if result.RowsAffected > 0 {
				totalRemoved++
				fmt.Printf("  📝 Removed duplicate entry: %s (%s) from journal %d\n",
					duplicate.AccountCode, duplicate.AccountName, journalID)

				// Check if this journal has any remaining lines
				var remainingLines int64
				tx.Model(&struct {
					ID uint `gorm:"primaryKey"`
				}{}).Table("unified_journal_lines").Where("journal_id = ?", journalID).Count(&remainingLines)

				// If no lines left, remove the journal header too
				if remainingLines == 0 {
					tx.Exec("DELETE FROM unified_journal_ledger WHERE id = ?", journalID)
					fmt.Printf("  🗑️  Removed empty journal header: %d\n", journalID)
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit duplicate removal: %v", err)
	}

	fmt.Printf("✅ Successfully removed %d duplicate entries!\n", totalRemoved)
	return nil
}

func createAdjustingEntry(db *gorm.DB, balanceDifference float64) error {
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

	// Create journal header
	entryNumber := fmt.Sprintf("ADJ-%s-001", time.Now().Format("20060102"))
	
	var journalID uint
	err := tx.Raw(`
		INSERT INTO unified_journal_ledger (entry_number, entry_date, description, status, created_at, updated_at)
		VALUES (?, CURRENT_DATE, 'Balance Sheet Adjusting Entry - Automated', 'POSTED', NOW(), NOW())
		RETURNING id
	`, entryNumber).Scan(&journalID).Error

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create journal header: %v", err)
	}

	// Get Retained Earnings account (or create suspense account)
	var retainedEarningsID uint
	err = tx.Raw("SELECT id FROM accounts WHERE code = '3201' OR name LIKE '%Retained Earnings%' OR name LIKE '%Laba Ditahan%' LIMIT 1").Scan(&retainedEarningsID).Error
	
	if err != nil || retainedEarningsID == 0 {
		// Create suspense account if retained earnings doesn't exist
		err = tx.Raw(`
			INSERT INTO accounts (code, name, type, is_active, created_at, updated_at)
			VALUES ('1901', 'Suspense Account', 'ASSET', true, NOW(), NOW())
			ON CONFLICT (code) DO NOTHING
			RETURNING id
		`).Scan(&retainedEarningsID).Error

		if err != nil {
			// Try to get existing suspense account
			tx.Raw("SELECT id FROM accounts WHERE code = '1901' LIMIT 1").Scan(&retainedEarningsID)
		}
	}

	if retainedEarningsID == 0 {
		tx.Rollback()
		return fmt.Errorf("failed to find or create adjusting account")
	}

	// Create adjusting entry lines
	if balanceDifference > 0 {
		// Assets > Liabilities + Equity, so credit retained earnings
		err = tx.Exec(`
			INSERT INTO unified_journal_lines (journal_id, account_id, debit_amount, credit_amount, description, created_at, updated_at)
			VALUES (?, ?, 0, ?, 'Adjusting entry for balance sheet balance', NOW(), NOW())
		`, journalID, retainedEarningsID, balanceDifference).Error
	} else {
		// Assets < Liabilities + Equity, so debit retained earnings
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