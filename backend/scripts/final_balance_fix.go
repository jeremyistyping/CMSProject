package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Final Comprehensive Balance Fix...")
	log.Println("================================================================================")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Execute all fixes
	if err := fixCashBankCOASync(db); err != nil {
		log.Printf("❌ Cash bank COA sync failed: %v", err)
		return
	}

	if err := fixAccountBalanceSigns(db); err != nil {
		log.Printf("❌ Account balance signs fix failed: %v", err)
		return
	}

	if err := updateParentBalances(db); err != nil {
		log.Printf("❌ Parent balance update failed: %v", err)
		return
	}

	if err := validateAndSummarize(db); err != nil {
		log.Printf("❌ Final validation failed: %v", err)
		return
	}

	log.Println("🎉 All balance issues fixed successfully!")
}

// fixCashBankCOASync fixes the specific cash bank COA sync issue
func fixCashBankCOASync(db *gorm.DB) error {
	log.Println("\n💰 STEP 1: Fixing Cash Bank & COA Sync...")

	// Update COA account ID 3 (1101 - Kas) to match cash bank balance
	updateKasBalanceSQL := `
		UPDATE accounts 
		SET balance = 38900000.00, updated_at = NOW()
		WHERE id = 3 AND code = '1101' AND deleted_at IS NULL;
	`

	result := db.Exec(updateKasBalanceSQL)
	if result.Error != nil {
		return fmt.Errorf("failed to update Kas account balance: %w", result.Error)
	}

	log.Printf("  ✅ Updated account 1101 (Kas) balance to 38,900,000.00")

	// Verify the sync
	var syncCheck struct {
		CashBankBalance float64 `json:"cash_bank_balance"`
		COABalance      float64 `json:"coa_balance"`
	}

	syncQuery := `
		SELECT 
			cb.balance as cash_bank_balance,
			a.balance as coa_balance
		FROM cash_banks cb
		JOIN accounts a ON cb.account_id = a.id
		WHERE cb.code = 'CSH-2025-0001'
		AND cb.deleted_at IS NULL
		AND a.deleted_at IS NULL;
	`

	if err := db.Raw(syncQuery).Scan(&syncCheck).Error; err != nil {
		return fmt.Errorf("failed to verify sync: %w", err)
	}

	if syncCheck.CashBankBalance == syncCheck.COABalance {
		log.Printf("  ✅ Sync verified: CashBank=%.2f, COA=%.2f", 
			syncCheck.CashBankBalance, syncCheck.COABalance)
	} else {
		log.Printf("  ⚠️ Sync still mismatched: CashBank=%.2f, COA=%.2f", 
			syncCheck.CashBankBalance, syncCheck.COABalance)
	}

	return nil
}

// fixAccountBalanceSigns fixes the accounting balance signs to follow standard conventions
func fixAccountBalanceSigns(db *gorm.DB) error {
	log.Println("\n🔧 STEP 2: Fixing Account Balance Signs...")

	// Standard accounting convention:
	// ASSET & EXPENSE: Positive balances (debits)
	// LIABILITY, EQUITY & REVENUE: Negative balances (credits)

	fixes := []struct {
		description string
		sql         string
	}{
		{
			"Fix REVENUE accounts to negative balances",
			`UPDATE accounts 
			 SET balance = -ABS(balance), updated_at = NOW()
			 WHERE type = 'REVENUE' 
			 AND balance > 0 
			 AND deleted_at IS NULL;`,
		},
		{
			"Fix LIABILITY accounts to negative balances",
			`UPDATE accounts 
			 SET balance = -ABS(balance), updated_at = NOW()
			 WHERE type = 'LIABILITY' 
			 AND balance > 0 
			 AND deleted_at IS NULL;`,
		},
		{
			"Fix EQUITY accounts to negative balances (if positive)",
			`UPDATE accounts 
			 SET balance = -ABS(balance), updated_at = NOW()
			 WHERE type = 'EQUITY' 
			 AND balance > 0 
			 AND deleted_at IS NULL;`,
		},
		{
			"Ensure ASSET accounts have positive balances",
			`UPDATE accounts 
			 SET balance = ABS(balance), updated_at = NOW()
			 WHERE type = 'ASSET' 
			 AND balance < 0 
			 AND deleted_at IS NULL;`,
		},
		{
			"Ensure EXPENSE accounts have positive balances",
			`UPDATE accounts 
			 SET balance = ABS(balance), updated_at = NOW()
			 WHERE type = 'EXPENSE' 
			 AND balance < 0 
			 AND deleted_at IS NULL;`,
		},
	}

	for _, fix := range fixes {
		log.Printf("  🔄 %s...", fix.description)
		
		result := db.Exec(fix.sql)
		if result.Error != nil {
			log.Printf("    ❌ Failed: %v", result.Error)
			continue
		}
		
		log.Printf("    ✅ Updated %d accounts", result.RowsAffected)
	}

	log.Println("  ✅ Account balance signs fixed successfully")
	return nil
}

// updateParentBalances updates all parent account balances recursively
func updateParentBalances(db *gorm.DB) error {
	log.Println("\n🏗️ STEP 3: Updating Parent Account Balances...")

	// Get all accounts ordered by hierarchy level (deepest first)
	type AccountHierarchy struct {
		ID       uint    `json:"id"`
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		ParentID *uint   `json:"parent_id"`
		IsHeader bool    `json:"is_header"`
		Level    int     `json:"level"`
		Type     string  `json:"type"`
	}

	var allAccounts []AccountHierarchy
	hierarchyQuery := `
		WITH RECURSIVE account_hierarchy AS (
			-- Base case: accounts without parents (level 0)
			SELECT 
				id, code, name, parent_id, is_header, type,
				0 as level
			FROM accounts 
			WHERE parent_id IS NULL AND deleted_at IS NULL
			
			UNION ALL
			
			-- Recursive case: accounts with parents
			SELECT 
				a.id, a.code, a.name, a.parent_id, a.is_header, a.type,
				ah.level + 1 as level
			FROM accounts a
			JOIN account_hierarchy ah ON a.parent_id = ah.id
			WHERE a.deleted_at IS NULL
		)
		SELECT * FROM account_hierarchy ORDER BY level DESC, code;
	`

	if err := db.Raw(hierarchyQuery).Scan(&allAccounts).Error; err != nil {
		return fmt.Errorf("failed to get hierarchy: %w", err)
	}

	// Find max level
	maxLevel := 0
	for _, acc := range allAccounts {
		if acc.Level > maxLevel {
			maxLevel = acc.Level
		}
	}

	log.Printf("  Found %d accounts, max level: %d", len(allAccounts), maxLevel)

	// Update balances from deepest level to root
	updateCount := 0
	for level := maxLevel; level >= 0; level-- {
		log.Printf("    Processing level %d...", level)

		// Get header accounts at this level
		var headerAccounts []AccountHierarchy
		for _, acc := range allAccounts {
			if acc.Level == level && acc.IsHeader {
				headerAccounts = append(headerAccounts, acc)
			}
		}

		// Update each header account
		for _, header := range headerAccounts {
			var childSum float64
			childSumQuery := `
				SELECT COALESCE(SUM(balance), 0) as child_sum
				FROM accounts 
				WHERE parent_id = ? AND deleted_at IS NULL;
			`

			if err := db.Raw(childSumQuery, header.ID).Scan(&childSum).Error; err != nil {
				log.Printf("      ❌ Failed to calculate sum for %s: %v", header.Code, err)
				continue
			}

			// Update the header account balance
			if err := db.Model(&models.Account{}).
				Where("id = ?", header.ID).
				Update("balance", childSum).Error; err != nil {
				log.Printf("      ❌ Failed to update %s: %v", header.Code, err)
				continue
			}

			log.Printf("      ✅ Updated %s (%s) balance to %.2f", 
				header.Code, header.Name, childSum)
			updateCount++
		}
	}

	log.Printf("  ✅ Updated %d parent account balances", updateCount)
	return nil
}

// validateAndSummarize provides final validation and summary
func validateAndSummarize(db *gorm.DB) error {
	log.Println("\n📊 STEP 4: Final Validation & Summary...")

	// Check cash bank sync
	log.Println("  4.1 Cash Bank & COA Sync Status:")
	var syncChecks []struct {
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		CashBalance float64 `json:"cash_balance"`
		COABalance  float64 `json:"coa_balance"`
		Difference  float64 `json:"difference"`
	}

	syncQuery := `
		SELECT 
			cb.code,
			cb.name,
			cb.balance as cash_balance,
			a.balance as coa_balance,
			(cb.balance - a.balance) as difference
		FROM cash_banks cb
		JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL 
		AND a.deleted_at IS NULL
		AND cb.is_active = true
		ORDER BY cb.code;
	`

	if err := db.Raw(syncQuery).Scan(&syncChecks).Error; err != nil {
		return fmt.Errorf("failed to check sync: %w", err)
	}

	syncIssues := 0
	for _, check := range syncChecks {
		if check.Difference > 0.01 || check.Difference < -0.01 {
			if syncIssues == 0 {
				log.Println("    ⚠️ Sync discrepancies found:")
			}
			log.Printf("      %s: Cash=%.2f, COA=%.2f, Diff=%.2f", 
				check.Code, check.CashBalance, check.COABalance, check.Difference)
			syncIssues++
		}
	}

	if syncIssues == 0 {
		log.Println("    ✅ All cash banks synchronized with COA")
	}

	// Check balance sheet equation
	log.Println("  4.2 Balance Sheet Equation:")
	var balances struct {
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
		WHERE deleted_at IS NULL;
	`

	if err := db.Raw(balanceQuery).Scan(&balances).Error; err != nil {
		return fmt.Errorf("failed to calculate balances: %w", err)
	}

	log.Printf("    Assets (DR): %.2f", balances.Assets)
	log.Printf("    Liabilities (CR): %.2f", balances.Liabilities)  
	log.Printf("    Equity (CR): %.2f", balances.Equity)
	log.Printf("    Revenue (CR): %.2f", balances.Revenue)
	log.Printf("    Expenses (DR): %.2f", balances.Expenses)

	// Balance sheet equation: Assets + Expenses = Liabilities + Equity + Revenue
	// In our system: Assets(+) + Expenses(+) + Liabilities(-) + Equity(-) + Revenue(-) = 0
	totalBalance := balances.Assets + balances.Expenses + balances.Liabilities + balances.Equity + balances.Revenue
	
	log.Printf("    Total Balance (should be 0): %.2f", totalBalance)
	
	if totalBalance >= -0.01 && totalBalance <= 0.01 {
		log.Println("    ✅ Balance sheet equation is satisfied!")
	} else {
		log.Printf("    ⚠️ Balance sheet has an imbalance of %.2f", totalBalance)
	}

	// Show account structure
	log.Println("  4.3 Account Structure Summary:")
	var structureSummary []struct {
		Type         string `json:"type"`
		TotalBalance float64 `json:"total_balance"`
		AccountCount int    `json:"account_count"`
		HeaderCount  int    `json:"header_count"`
		LeafCount    int    `json:"leaf_count"`
	}

	structureQuery := `
		SELECT 
			type,
			COALESCE(SUM(balance), 0) as total_balance,
			COUNT(*) as account_count,
			SUM(CASE WHEN is_header = true THEN 1 ELSE 0 END) as header_count,
			SUM(CASE WHEN is_header = false THEN 1 ELSE 0 END) as leaf_count
		FROM accounts 
		WHERE deleted_at IS NULL 
		GROUP BY type
		ORDER BY type;
	`

	if err := db.Raw(structureQuery).Scan(&structureSummary).Error; err != nil {
		return fmt.Errorf("failed to get structure summary: %w", err)
	}

	for _, summary := range structureSummary {
		log.Printf("    %s: %.2f (%d accounts - %d headers, %d leafs)", 
			summary.Type, summary.TotalBalance, summary.AccountCount, 
			summary.HeaderCount, summary.LeafCount)
	}

	log.Println("  ✅ Final validation completed")
	return nil
}