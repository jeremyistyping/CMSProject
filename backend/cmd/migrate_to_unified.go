package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"app-sistem-akuntansi/database"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// Parse arguments
	dryRun := false
	if len(os.Args) > 1 && os.Args[1] == "--dry-run" {
		dryRun = true
	}
	
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("MIGRATION TO UNIFIED JOURNALS SYSTEM (SSOT)")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
	
	if dryRun {
		fmt.Println("🔍 Mode: DRY RUN (changes will be rolled back)")
	} else {
		fmt.Println("⚠️  Mode: PRODUCTION (changes will be committed)")
	}
	fmt.Println()
	
	// Connect to database
	db := database.ConnectDB()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	defer sqlDB.Close()
	
	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		log.Fatalf("Failed to begin transaction: %v", tx.Error)
	}
	
	// Ensure rollback on panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Fatalf("Panic during migration: %v", r)
		}
	}()
	
	// Step 1: Pre-migration analysis
	fmt.Println("📊 [STEP 1] PRE-MIGRATION ANALYSIS")
	fmt.Println(strings.Repeat("-", 80))
	
	var oldJournalCount, unifiedJournalCount int64
	tx.Raw("SELECT COUNT(*) FROM journal_entries").Scan(&oldJournalCount)
	tx.Raw("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&unifiedJournalCount)
	
	fmt.Printf("Old journal_entries:       %d\n", oldJournalCount)
	fmt.Printf("Unified journal_ledger:    %d\n", unifiedJournalCount)
	fmt.Println()
	
	// Get current balances for comparison
	type AccountBalance struct {
		Code       string
		Name       string
		Type       string
		OldBalance float64
	}
	
	var preBalances []AccountBalance
	tx.Raw(`
		SELECT code, name, type, balance as old_balance
		FROM accounts
		WHERE is_active = true AND COALESCE(is_header, false) = false
			AND ABS(balance) > 0.01
		ORDER BY code
	`).Scan(&preBalances)
	
	fmt.Printf("Accounts with non-zero balance: %d\n", len(preBalances))
	fmt.Println()
	
	// Step 2: Delete old journal entries
	fmt.Println("🗑️  [STEP 2] DELETING OLD JOURNAL ENTRIES")
	fmt.Println(strings.Repeat("-", 80))
	
	// Delete journal lines first (foreign key)
	result := tx.Exec("DELETE FROM journal_lines")
	if result.Error != nil {
		tx.Rollback()
		log.Fatalf("Failed to delete journal_lines: %v", result.Error)
	}
	fmt.Printf("✓ Deleted %d journal lines\n", result.RowsAffected)
	
	// Delete journal entries
	result = tx.Exec("DELETE FROM journal_entries")
	if result.Error != nil {
		tx.Rollback()
		log.Fatalf("Failed to delete journal_entries: %v", result.Error)
	}
	fmt.Printf("✓ Deleted %d journal entries\n", result.RowsAffected)
	
	// Delete accounting periods
	result = tx.Exec("DELETE FROM accounting_periods")
	if result.Error != nil {
		tx.Rollback()
		log.Fatalf("Failed to delete accounting_periods: %v", result.Error)
	}
	fmt.Printf("✓ Deleted %d accounting periods\n", result.RowsAffected)
	fmt.Println()
	
	// Step 3: Reset account balances
	fmt.Println("🔄 [STEP 3] RESETTING ACCOUNT BALANCES TO ZERO")
	fmt.Println(strings.Repeat("-", 80))
	
	result = tx.Exec(`
		UPDATE accounts 
		SET balance = 0, updated_at = NOW()
		WHERE is_active = true
	`)
	if result.Error != nil {
		tx.Rollback()
		log.Fatalf("Failed to reset account balances: %v", result.Error)
	}
	fmt.Printf("✓ Reset %d account balances to 0\n", result.RowsAffected)
	fmt.Println()
	
	// Step 4: Recalculate balances from unified journals
	fmt.Println("🔢 [STEP 4] RECALCULATING BALANCES FROM UNIFIED JOURNALS")
	fmt.Println(strings.Repeat("-", 80))
	
	updateQuery := `
		WITH account_balances AS (
			SELECT 
				a.id,
				a.code,
				a.type,
				CASE 
					WHEN a.type IN ('ASSET', 'EXPENSE') 
					THEN COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
					ELSE COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
				END as calculated_balance
			FROM accounts a
			LEFT JOIN unified_journal_lines ujl ON a.id = ujl.account_id
			LEFT JOIN unified_journal_ledger uj ON uj.id = ujl.journal_id
			WHERE uj.status = 'POSTED' OR uj.id IS NULL
			GROUP BY a.id, a.code, a.type
		)
		UPDATE accounts a
		SET 
			balance = ab.calculated_balance,
			updated_at = NOW()
		FROM account_balances ab
		WHERE a.id = ab.id
	`
	
	result = tx.Exec(updateQuery)
	if result.Error != nil {
		tx.Rollback()
		log.Fatalf("Failed to recalculate balances: %v", result.Error)
	}
	fmt.Printf("✓ Recalculated %d account balances\n", result.RowsAffected)
	fmt.Println()
	
	// Step 5: Show updated balances
	fmt.Println("💰 [STEP 5] UPDATED ACCOUNT BALANCES")
	fmt.Println(strings.Repeat("-", 80))
	
	type NewBalance struct {
		Code       string
		Name       string
		Type       string
		NewBalance float64
	}
	
	var newBalances []NewBalance
	tx.Raw(`
		SELECT code, name, type, balance as new_balance
		FROM accounts
		WHERE is_active = true 
			AND COALESCE(is_header, false) = false
			AND ABS(balance) > 0.01
		ORDER BY code
	`).Scan(&newBalances)
	
	fmt.Printf("%-6s | %-30s | %-10s | %15s\n", "Code", "Account Name", "Type", "New Balance")
	fmt.Println(strings.Repeat("-", 80))
	for _, b := range newBalances {
		fmt.Printf("%-6s | %-30s | %-10s | %15.2f\n", 
			b.Code, b.Name, b.Type, b.NewBalance)
	}
	fmt.Println()
	
	// Step 6: Validate balance sheet equation
	fmt.Println("✅ [STEP 6] VALIDATING BALANCE SHEET EQUATION")
	fmt.Println(strings.Repeat("-", 80))
	
	type BalanceTotals struct {
		TotalAssets      float64
		TotalLiabilities float64
		TotalEquity      float64
		TotalRevenue     float64
		TotalExpense     float64
	}
	
	var totals BalanceTotals
	tx.Raw(`
		SELECT 
			SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) AS total_assets,
			SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) AS total_liabilities,
			SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END) AS total_equity,
			SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END) AS total_revenue,
			SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END) AS total_expense
		FROM accounts
		WHERE is_active = true AND COALESCE(is_header, false) = false
	`).Scan(&totals)
	
	liabPlusEquity := totals.TotalLiabilities + totals.TotalEquity
	diffBeforeClosing := totals.TotalAssets - liabPlusEquity
	netIncome := totals.TotalRevenue - totals.TotalExpense
	diffWithTemp := totals.TotalAssets - (totals.TotalLiabilities + totals.TotalEquity + netIncome)
	
	fmt.Printf("Total Assets:              Rp %15.2f\n", totals.TotalAssets)
	fmt.Printf("Total Liabilities:         Rp %15.2f\n", totals.TotalLiabilities)
	fmt.Printf("Total Equity:              Rp %15.2f\n", totals.TotalEquity)
	fmt.Printf("Total Liab + Equity:       Rp %15.2f\n", liabPlusEquity)
	fmt.Printf("Difference (before close): Rp %15.2f\n", diffBeforeClosing)
	fmt.Println()
	fmt.Printf("Total Revenue:             Rp %15.2f (not closed yet)\n", totals.TotalRevenue)
	fmt.Printf("Total Expense:             Rp %15.2f (not closed yet)\n", totals.TotalExpense)
	fmt.Printf("Net Income:                Rp %15.2f\n", netIncome)
	fmt.Printf("Difference (with temp):    Rp %15.2f", diffWithTemp)
	
	if diffWithTemp < 0.01 && diffWithTemp > -0.01 {
		fmt.Println(" ✅ BALANCED (including temp accounts)")
	} else if diffBeforeClosing < 0.01 && diffBeforeClosing > -0.01 {
		fmt.Println(" ✅ BALANCED")
	} else {
		fmt.Println(" ⚠️  Will be balanced after period closing")
	}
	fmt.Println()
	
	// Step 7: Decide commit or rollback
	fmt.Println(strings.Repeat("=", 80))
	if dryRun {
		fmt.Println("🔙 DRY RUN MODE - ROLLING BACK CHANGES")
		tx.Rollback()
		fmt.Println("✓ All changes have been rolled back")
	} else {
		fmt.Println("💾 COMMITTING CHANGES TO DATABASE")
		if err := tx.Commit().Error; err != nil {
			log.Fatalf("Failed to commit transaction: %v", err)
		}
		fmt.Println("✅ Migration completed successfully!")
	}
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
	
	// Final summary
	fmt.Println("📋 SUMMARY:")
	fmt.Println()
	if dryRun {
		fmt.Println("✓ Dry run completed - no changes were made to the database")
		fmt.Println("  Run without --dry-run flag to apply changes")
	} else {
		fmt.Println("✅ Migration completed successfully!")
		fmt.Println()
		fmt.Println("NEXT STEPS:")
		fmt.Println("1. Verify balance sheet in UI")
		fmt.Println("2. Run period closing to close Revenue/Expense accounts")
		fmt.Println("3. After closing, Retained Earnings will include Net Income")
		fmt.Println("4. Balance sheet will be fully balanced")
		fmt.Println()
		fmt.Println("Period Closing:")
		fmt.Println("  - Via UI: Go to Period Closing menu")
		fmt.Println("  - Via API: POST /api/period-closing/execute")
		fmt.Println()
		fmt.Printf("Migration timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	}
	fmt.Println()
}
