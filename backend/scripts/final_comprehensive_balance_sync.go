package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Final Comprehensive Balance Sync...")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Perform comprehensive balance sync
	if err := performComprehensiveSync(db); err != nil {
		log.Fatalf("❌ Comprehensive sync failed: %v", err)
	}

	log.Println("🎉 Comprehensive balance sync completed successfully!")
}

func performComprehensiveSync(db *gorm.DB) error {
	log.Println("📝 Performing comprehensive balance sync...")

	// Step 1: Sync all cash banks with their COA accounts
	log.Println("💰 Step 1: Syncing all cash banks...")
	
	var cashBanks []struct {
		ID        uint `json:"id"`
		Code      string `json:"code"`
		Name      string `json:"name"`
		AccountID uint `json:"account_id"`
	}

	if err := db.Raw(`
		SELECT id, code, name, account_id
		FROM cash_banks 
		WHERE deleted_at IS NULL AND is_active = true
		ORDER BY code
	`).Scan(&cashBanks).Error; err != nil {
		return fmt.Errorf("failed to get cash banks: %w", err)
	}

	log.Printf("Found %d active cash banks to sync", len(cashBanks))

	for _, cb := range cashBanks {
		log.Printf("  Syncing %s (%s)...", cb.Code, cb.Name)
		
		if err := db.Exec("SELECT manual_sync_cashbank_coa(?)", cb.ID).Error; err != nil {
			log.Printf("    ❌ Failed to sync %s: %v", cb.Code, err)
		} else {
			log.Printf("    ✅ Synced %s successfully", cb.Code)
		}
	}

	// Step 2: Update all parent account balances
	log.Println("🏗️ Step 2: Updating parent account balances...")

	// Get all accounts that could be parents (header accounts)
	var parentAccounts []struct {
		ID       uint   `json:"id"`
		Code     string `json:"code"`
		Name     string `json:"name"`
		IsHeader bool   `json:"is_header"`
	}

	if err := db.Raw(`
		SELECT DISTINCT p.id, p.code, p.name, p.is_header
		FROM accounts p
		WHERE p.deleted_at IS NULL 
		AND (p.is_header = true OR p.id IN (
			SELECT DISTINCT parent_id 
			FROM accounts 
			WHERE parent_id IS NOT NULL AND deleted_at IS NULL
		))
		ORDER BY p.code
	`).Scan(&parentAccounts).Error; err != nil {
		return fmt.Errorf("failed to get parent accounts: %w", err)
	}

	log.Printf("Found %d potential parent accounts", len(parentAccounts))

	for _, pa := range parentAccounts {
		log.Printf("  Updating parent %s (%s)...", pa.Code, pa.Name)
		
		if err := db.Exec("SELECT update_parent_account_balances(?)", pa.ID).Error; err != nil {
			log.Printf("    ⚠️ Failed to update parent %s: %v", pa.Code, err)
		} else {
			log.Printf("    ✅ Updated parent %s", pa.Code)
		}
	}

	// Step 3: Final validation
	log.Println("🔍 Step 3: Final validation...")

	// Check cash bank sync status
	var syncIssues []struct {
		Code            string  `json:"code"`
		Name            string  `json:"name"`
		CashBankBalance float64 `json:"cash_bank_balance"`
		COABalance      float64 `json:"coa_balance"`
		Difference      float64 `json:"difference"`
	}

	syncQuery := `
		SELECT 
			cb.code,
			cb.name,
			cb.balance as cash_bank_balance,
			a.balance as coa_balance,
			(cb.balance - a.balance) as difference
		FROM cash_banks cb
		JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL 
		AND a.deleted_at IS NULL
		AND cb.is_active = true
		AND ABS(cb.balance - a.balance) > 0.01
		ORDER BY ABS(cb.balance - a.balance) DESC
	`

	if err := db.Raw(syncQuery).Scan(&syncIssues).Error; err != nil {
		return fmt.Errorf("failed to check sync issues: %w", err)
	}

	if len(syncIssues) == 0 {
		log.Println("  ✅ All cash banks are in sync with COA accounts")
	} else {
		log.Printf("  ⚠️ Found %d cash banks still out of sync:", len(syncIssues))
		for _, issue := range syncIssues {
			log.Printf("    %s: Cash=%.2f, COA=%.2f, Diff=%.2f", 
				issue.Code, issue.CashBankBalance, issue.COABalance, issue.Difference)
		}
	}

	// Check parent-child consistency
	var parentChildIssues []struct {
		Code          string  `json:"code"`
		Name          string  `json:"name"`
		ParentBalance float64 `json:"parent_balance"`
		ChildrenSum   float64 `json:"children_sum"`
		Difference    float64 `json:"difference"`
	}

	parentQuery := `
		SELECT 
			p.code,
			p.name,
			p.balance as parent_balance,
			COALESCE(children_sum.total, 0) as children_sum,
			(p.balance - COALESCE(children_sum.total, 0)) as difference
		FROM accounts p
		LEFT JOIN (
			SELECT 
				parent_id,
				SUM(balance) as total
			FROM accounts 
			WHERE deleted_at IS NULL 
			GROUP BY parent_id
		) children_sum ON p.id = children_sum.parent_id
		WHERE p.is_header = true 
		AND p.deleted_at IS NULL
		AND ABS(p.balance - COALESCE(children_sum.total, 0)) > 0.01
		ORDER BY ABS(p.balance - COALESCE(children_sum.total, 0)) DESC
	`

	if err := db.Raw(parentQuery).Scan(&parentChildIssues).Error; err != nil {
		return fmt.Errorf("failed to check parent-child issues: %w", err)
	}

	if len(parentChildIssues) == 0 {
		log.Println("  ✅ All parent accounts have correct balances")
	} else {
		log.Printf("  ⚠️ Found %d parent accounts with balance issues:", len(parentChildIssues))
		for _, issue := range parentChildIssues {
			log.Printf("    %s: Parent=%.2f, Children=%.2f, Diff=%.2f", 
				issue.Code, issue.ParentBalance, issue.ChildrenSum, issue.Difference)
		}
	}

	// Show balance sheet summary
	var balanceTotals struct {
		Assets      float64 `json:"assets"`
		Liabilities float64 `json:"liabilities"`
		Equity      float64 `json:"equity"`
		Revenue     float64 `json:"revenue"`
		Expenses    float64 `json:"expenses"`
	}

	balanceQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END), 0) as assets,
			COALESCE(SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END), 0) as liabilities,
			COALESCE(SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END), 0) as equity,
			COALESCE(SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END), 0) as revenue,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END), 0) as expenses
		FROM accounts 
		WHERE deleted_at IS NULL
	`

	if err := db.Raw(balanceQuery).Scan(&balanceTotals).Error; err != nil {
		return fmt.Errorf("failed to calculate balance totals: %w", err)
	}

	log.Println("📊 Final Balance Sheet Summary:")
	log.Printf("  Assets (DR): %.2f", balanceTotals.Assets)
	log.Printf("  Liabilities (CR): %.2f", balanceTotals.Liabilities)
	log.Printf("  Equity (CR): %.2f", balanceTotals.Equity)
	log.Printf("  Revenue (CR): %.2f", balanceTotals.Revenue)
	log.Printf("  Expenses (DR): %.2f", balanceTotals.Expenses)

	// Check balance equation
	balanceEquation := balanceTotals.Assets + balanceTotals.Expenses + 
					  balanceTotals.Liabilities + balanceTotals.Equity + balanceTotals.Revenue

	log.Printf("  Balance Equation Check: %.2f", balanceEquation)

	if balanceEquation >= -0.01 && balanceEquation <= 0.01 {
		log.Println("  ✅ Balance sheet equation is satisfied!")
	} else {
		log.Printf("  ⚠️ Balance sheet imbalance: %.2f", balanceEquation)
	}

	return nil
}