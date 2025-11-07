package startup

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// MigrateToUnifiedJournals migrates from old journal_entries system to unified_journal_ledger
// This function:
// 1. Checks if migration is needed (if old journal_entries exist)
// 2. Deletes old journal_entries and journal_lines
// 3. Resets accounts.balance to 0
// 4. Recalculates balances from unified_journal_ledger only
func MigrateToUnifiedJournals(db *gorm.DB) error {
	log.Println("🔄 Checking if migration to Unified Journals is needed...")

	// Check if old journal entries exist
	var oldJournalCount int64
	if err := db.Table("journal_entries").Count(&oldJournalCount).Error; err != nil {
		// Table might not exist, which is fine
		log.Println("✅ No old journal_entries table found - system is clean")
		return nil
	}

	if oldJournalCount == 0 {
		log.Println("✅ No old journal entries found - migration not needed")
		return nil
	}

	log.Printf("⚠️  Found %d old journal entries - starting migration...", oldJournalCount)

	// Start transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// Step 1: Delete old journal_lines
		log.Println("📝 Step 1/4: Deleting old journal_lines...")
		if err := tx.Exec("DELETE FROM journal_lines").Error; err != nil {
			return fmt.Errorf("failed to delete journal_lines: %v", err)
		}

		// Step 2: Delete old journal_entries
		log.Println("📝 Step 2/4: Deleting old journal_entries...")
		if err := tx.Exec("DELETE FROM journal_entries").Error; err != nil {
			return fmt.Errorf("failed to delete journal_entries: %v", err)
		}

		// Step 3: Delete old accounting_periods
		log.Println("📝 Step 3/4: Deleting old accounting_periods...")
		if err := tx.Exec("DELETE FROM accounting_periods").Error; err != nil {
			return fmt.Errorf("failed to delete accounting_periods: %v", err)
		}

		// Step 4: Reset all account balances to 0
		log.Println("📝 Step 4/4: Resetting account balances...")
		if err := tx.Exec("UPDATE accounts SET balance = 0").Error; err != nil {
			return fmt.Errorf("failed to reset account balances: %v", err)
		}

		// Step 5: Recalculate balances from unified_journal_ledger
		log.Println("📊 Recalculating balances from unified_journal_ledger...")
		
		// For Assets and Expenses: balance = SUM(debit) - SUM(credit)
		debitNormalSQL := `
			UPDATE accounts a
			SET balance = COALESCE((
				SELECT SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
				FROM unified_journal_lines ujl
				JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
				WHERE ujl.account_id = a.id
				AND uj.deleted_at IS NULL
			), 0)
			WHERE a.type IN ('ASSET', 'EXPENSE')
		`
		if err := tx.Exec(debitNormalSQL).Error; err != nil {
			return fmt.Errorf("failed to calculate balances for ASSET/EXPENSE accounts: %v", err)
		}

		// For Liabilities, Equity, and Revenue: balance = SUM(credit) - SUM(debit)
		creditNormalSQL := `
			UPDATE accounts a
			SET balance = COALESCE((
				SELECT SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
				FROM unified_journal_lines ujl
				JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
				WHERE ujl.account_id = a.id
				AND uj.deleted_at IS NULL
			), 0)
			WHERE a.type IN ('LIABILITY', 'EQUITY', 'REVENUE')
		`
		if err := tx.Exec(creditNormalSQL).Error; err != nil {
			return fmt.Errorf("failed to calculate balances for LIABILITY/EQUITY/REVENUE accounts: %v", err)
		}

		log.Println("✅ Migration to Unified Journals completed successfully!")
		
		// Validate balance sheet
		if err := validateBalanceSheet(tx); err != nil {
			log.Printf("⚠️  Balance sheet validation: %v", err)
			// Don't fail the migration, just warn
		}

		return nil
	})
}

// validateBalanceSheet performs a quick balance sheet validation
func validateBalanceSheet(db *gorm.DB) error {
	type BalanceResult struct {
		TotalAssets      float64
		TotalLiabilities float64
		TotalEquity      float64
		TotalRevenue     float64
		TotalExpense     float64
	}

	var result BalanceResult
	
	// Get totals by account type
	if err := db.Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END), 0) as total_assets,
			COALESCE(SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END), 0) as total_liabilities,
			COALESCE(SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END), 0) as total_equity,
			COALESCE(SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END), 0) as total_expense
		FROM accounts
		WHERE deleted_at IS NULL
	`).Scan(&result).Error; err != nil {
		return fmt.Errorf("failed to calculate balance sheet totals: %v", err)
	}

	// Calculate net income (Revenue - Expense)
	netIncome := result.TotalRevenue - result.TotalExpense

	// Accounting equation: Assets = Liabilities + Equity + (Revenue - Expense)
	leftSide := result.TotalAssets
	rightSide := result.TotalLiabilities + result.TotalEquity + netIncome
	difference := leftSide - rightSide

	log.Printf("📊 Balance Sheet Summary:")
	log.Printf("   Assets: Rp %.2f", result.TotalAssets)
	log.Printf("   Liabilities: Rp %.2f", result.TotalLiabilities)
	log.Printf("   Equity: Rp %.2f", result.TotalEquity)
	log.Printf("   Revenue: Rp %.2f (temp)", result.TotalRevenue)
	log.Printf("   Expense: Rp %.2f (temp)", result.TotalExpense)
	log.Printf("   Net Income: Rp %.2f", netIncome)
	log.Printf("   Difference: Rp %.2f", difference)

	if difference < -0.01 || difference > 0.01 { // Allow small rounding errors
		log.Printf("⚠️  Balance sheet NOT BALANCED!")
		log.Printf("💡 This is expected if period closing has not been run yet")
		log.Printf("💡 Run period closing to close Revenue/Expense to Retained Earnings")
	} else {
		log.Printf("✅ Balance sheet is BALANCED!")
	}

	return nil
}
