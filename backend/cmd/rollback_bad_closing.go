package main

import (
	"fmt"
	"log"
	"strings"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println("ROLLBACK BAD PERIOD CLOSING & RESET TO UNIFIED JOURNALS")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println()

	db := database.ConnectDB()

	fmt.Println("⚠️  This will:")
	fmt.Println("   1. Delete all accounting_periods records")
	fmt.Println("   2. Recalculate ALL account balances from unified_journal_ledger ONLY")
	fmt.Println("   3. Reset accounts to match their journal entries")
	fmt.Println()

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		log.Fatalf("Failed to begin transaction: %v", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Fatalf("Panic: %v", r)
		}
	}()

	// Step 1: Delete all period closing records
	fmt.Println("📝 [STEP 1] Deleting period closing records...")
	result := tx.Exec("DELETE FROM accounting_periods")
	if result.Error != nil {
		tx.Rollback()
		log.Fatalf("Failed to delete accounting_periods: %v", result.Error)
	}
	fmt.Printf("✓ Deleted %d accounting period records\n\n", result.RowsAffected)

	// Step 2: Reset all account balances to 0
	fmt.Println("📝 [STEP 2] Resetting all account balances to 0...")
	result = tx.Exec("UPDATE accounts SET balance = 0")
	if result.Error != nil {
		tx.Rollback()
		log.Fatalf("Failed to reset account balances: %v", result.Error)
	}
	fmt.Printf("✓ Reset %d account balances\n\n", result.RowsAffected)

	// Step 3: Recalculate balances from unified journals ONLY
	fmt.Println("📝 [STEP 3] Recalculating balances from unified_journal_ledger...")

	// For Assets and Expenses: balance = SUM(debit) - SUM(credit)
	debitNormalSQL := `
		UPDATE accounts a
		SET balance = COALESCE((
			SELECT SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
			FROM unified_journal_lines ujl
			JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
			WHERE ujl.account_id = a.id
			AND uj.status = 'POSTED'
			AND uj.deleted_at IS NULL
		), 0)
		WHERE a.type IN ('ASSET', 'EXPENSE')
	`
	result = tx.Exec(debitNormalSQL)
	if result.Error != nil {
		tx.Rollback()
		log.Fatalf("Failed to calculate ASSET/EXPENSE balances: %v", result.Error)
	}
	fmt.Printf("✓ Recalculated %d ASSET/EXPENSE accounts\n", result.RowsAffected)

	// For Liabilities, Equity, and Revenue: balance = SUM(credit) - SUM(debit)
	creditNormalSQL := `
		UPDATE accounts a
		SET balance = COALESCE((
			SELECT SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
			FROM unified_journal_lines ujl
			JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
			WHERE ujl.account_id = a.id
			AND uj.status = 'POSTED'
			AND uj.deleted_at IS NULL
		), 0)
		WHERE a.type IN ('LIABILITY', 'EQUITY', 'REVENUE')
	`
	result = tx.Exec(creditNormalSQL)
	if result.Error != nil {
		tx.Rollback()
		log.Fatalf("Failed to calculate LIABILITY/EQUITY/REVENUE balances: %v", result.Error)
	}
	fmt.Printf("✓ Recalculated %d LIABILITY/EQUITY/REVENUE accounts\n\n", result.RowsAffected)

	// Step 4: Verify results
	fmt.Println("📝 [STEP 4] Verifying results...")

	type AccountBalance struct {
		Code    string
		Name    string
		Type    string
		Balance float64
	}

	var accounts []AccountBalance
	tx.Raw(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE deleted_at IS NULL
			AND is_header = false
			AND ABS(balance) > 0.01
		ORDER BY code
	`).Scan(&accounts)

	fmt.Printf("%-6s | %-35s | %-10s | %15s\n", "Code", "Name", "Type", "Balance")
	fmt.Println(strings.Repeat("-", 80))
	for _, acc := range accounts {
		fmt.Printf("%-6s | %-35s | %-10s | %15.2f\n",
			acc.Code, truncate(acc.Name, 35), acc.Type, acc.Balance)
	}
	fmt.Println()

	// Step 5: Check balance sheet
	fmt.Println("📝 [STEP 5] Checking balance sheet equation...")

	type BSTotal struct {
		TotalAssets      float64
		TotalLiabilities float64
		TotalEquity      float64
		TotalRevenue     float64
		TotalExpense     float64
	}

	var bsTotal BSTotal
	tx.Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END), 0) as total_assets,
			COALESCE(SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END), 0) as total_liabilities,
			COALESCE(SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END), 0) as total_equity,
			COALESCE(SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END), 0) as total_expense
		FROM accounts
		WHERE deleted_at IS NULL
	`).Scan(&bsTotal)

	netIncome := bsTotal.TotalRevenue - bsTotal.TotalExpense
	leftSide := bsTotal.TotalAssets
	rightSide := bsTotal.TotalLiabilities + bsTotal.TotalEquity + netIncome
	diff := leftSide - rightSide

	fmt.Printf("Assets:       Rp %18.2f\n", bsTotal.TotalAssets)
	fmt.Printf("Liabilities:  Rp %18.2f\n", bsTotal.TotalLiabilities)
	fmt.Printf("Equity:       Rp %18.2f\n", bsTotal.TotalEquity)
	fmt.Printf("Revenue:      Rp %18.2f (temp)\n", bsTotal.TotalRevenue)
	fmt.Printf("Expense:      Rp %18.2f (temp)\n", bsTotal.TotalExpense)
	fmt.Printf("Net Income:   Rp %18.2f\n", netIncome)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Difference:   Rp %18.2f ", diff)

	if diff < -0.01 || diff > 0.01 {
		fmt.Printf("❌ NOT BALANCED\n")
		tx.Rollback()
		log.Fatal("Balance sheet is not balanced after rollback - rolling back transaction")
	} else {
		fmt.Printf("✅ BALANCED\n")
	}
	fmt.Println()

	// Commit transaction
	fmt.Println("💾 Committing changes...")
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println("✅ ROLLBACK COMPLETED SUCCESSFULLY!")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Run unified period closing service to close Revenue/Expense accounts")
	fmt.Println("2. Verify balance sheet is balanced after closing")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
