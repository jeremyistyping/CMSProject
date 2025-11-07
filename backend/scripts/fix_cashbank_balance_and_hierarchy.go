package main

import (
	"fmt"
	"log"
	"math"
	"strings"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Starting Comprehensive Cash Bank & Account Balance Fix...")
	log.Println(strings.Repeat("=", 80))

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully")

	// Execute fixes in sequence
	if err := fixAccountHierarchy(db); err != nil {
		log.Printf("❌ Account hierarchy fix failed: %v", err)
		return
	}

	if err := syncCashBankWithCOA(db); err != nil {
		log.Printf("❌ Cash bank COA sync failed: %v", err)
		return
	}

	if err := updateParentAccountBalances(db); err != nil {
		log.Printf("❌ Parent balance update failed: %v", err)
		return
	}

	if err := validateFinalResults(db); err != nil {
		log.Printf("❌ Final validation failed: %v", err)
		return
	}

	log.Println("🎉 All fixes completed successfully!")
}

// fixAccountHierarchy fixes account header status and parent relationships
func fixAccountHierarchy(db *gorm.DB) error {
	log.Println("\n📋 STEP 1: Fixing Account Hierarchy...")

	// 1.1 Fix header status for accounts with children
	log.Println("  1.1 Updating header status for parent accounts...")
	
	updateHeadersSQL := `
		UPDATE accounts 
		SET is_header = true, updated_at = NOW()
		WHERE id IN (
			SELECT DISTINCT parent_id 
			FROM accounts 
			WHERE parent_id IS NOT NULL 
			AND deleted_at IS NULL
		)
		AND is_header = false
		AND deleted_at IS NULL;
	`
	
	result := db.Exec(updateHeadersSQL)
	if result.Error != nil {
		return fmt.Errorf("failed to update header status: %w", result.Error)
	}
	log.Printf("    ✅ Updated %d accounts to header status", result.RowsAffected)

	// 1.2 Fix non-header status for accounts without children
	log.Println("  1.2 Fixing non-header status for leaf accounts...")
	
	updateNonHeadersSQL := `
		UPDATE accounts 
		SET is_header = false, updated_at = NOW()
		WHERE is_header = true
		AND id NOT IN (
			SELECT DISTINCT parent_id 
			FROM accounts 
			WHERE parent_id IS NOT NULL 
			AND deleted_at IS NULL
		)
		AND deleted_at IS NULL;
	`
	
	result = db.Exec(updateNonHeadersSQL)
	if result.Error != nil {
		return fmt.Errorf("failed to update non-header status: %w", result.Error)
	}
	log.Printf("    ✅ Updated %d accounts to non-header status", result.RowsAffected)

	// 1.3 Ensure specific accounts are properly set as headers
	log.Println("  1.3 Ensuring key parent accounts are headers...")
	
	keyAccounts := []string{"1101", "1102"} // KAS and BANK
	for _, code := range keyAccounts {
		updateKeySQL := `
			UPDATE accounts 
			SET is_header = true, updated_at = NOW()
			WHERE code = ? AND deleted_at IS NULL;
		`
		result := db.Exec(updateKeySQL, code)
		if result.Error != nil {
			log.Printf("    ⚠️ Warning: Failed to update account %s: %v", code, result.Error)
		} else if result.RowsAffected > 0 {
			log.Printf("    ✅ Set account %s as header", code)
		}
	}

	log.Println("  ✅ Account hierarchy fixed successfully")
	return nil
}

// syncCashBankWithCOA synchronizes cash bank balances with linked COA accounts
func syncCashBankWithCOA(db *gorm.DB) error {
	log.Println("\n💰 STEP 2: Synchronizing Cash Bank with COA Balances...")

	// 2.1 Find all cash bank accounts with their linked COA accounts
	type CashBankCOALink struct {
		CashBankID      uint    `json:"cash_bank_id"`
		CashBankCode    string  `json:"cash_bank_code"`
		CashBankName    string  `json:"cash_bank_name"`
		CashBankBalance float64 `json:"cash_bank_balance"`
		COAAccountID    uint    `json:"coa_account_id"`
		COACode         string  `json:"coa_code"`
		COAName         string  `json:"coa_name"`
		COABalance      float64 `json:"coa_balance"`
		TransactionSum  float64 `json:"transaction_sum"`
		Discrepancy     float64 `json:"discrepancy"`
	}

	var links []CashBankCOALink
	
	query := `
		SELECT 
			cb.id as cash_bank_id,
			cb.code as cash_bank_code,
			cb.name as cash_bank_name,
			cb.balance as cash_bank_balance,
			a.id as coa_account_id,
			a.code as coa_code,
			a.name as coa_name,
			a.balance as coa_balance,
			COALESCE(tx_sum.transaction_sum, 0) as transaction_sum,
			(cb.balance - a.balance) as discrepancy
		FROM cash_banks cb
		JOIN accounts a ON cb.account_id = a.id
		LEFT JOIN (
			SELECT 
				cash_bank_id, 
				SUM(amount) as transaction_sum
			FROM cash_bank_transactions 
			WHERE deleted_at IS NULL 
			GROUP BY cash_bank_id
		) tx_sum ON cb.id = tx_sum.cash_bank_id
		WHERE cb.deleted_at IS NULL 
		AND a.deleted_at IS NULL
		AND cb.is_active = true
		ORDER BY cb.code;
	`

	if err := db.Raw(query).Scan(&links).Error; err != nil {
		return fmt.Errorf("failed to fetch cash bank COA links: %w", err)
	}

	log.Printf("  Found %d cash bank accounts with COA links", len(links))

	// 2.2 Sync balances
	syncCount := 0
	for _, link := range links {
		log.Printf("  📊 %s (%s): CashBank=%.2f, COA=%.2f, TxSum=%.2f", 
			link.CashBankCode, link.CashBankName, 
			link.CashBankBalance, link.COABalance, link.TransactionSum)

		// Determine the correct balance (prefer transaction sum if available)
		correctBalance := link.TransactionSum
		if link.TransactionSum == 0 && link.CashBankBalance != 0 {
			// If no transactions but cash bank has balance, use cash bank balance
			correctBalance = link.CashBankBalance
		}

		// Update both cash bank and COA to the correct balance
		tx := db.Begin()
		
		// Update cash bank balance
		if err := tx.Model(&models.CashBank{}).
			Where("id = ?", link.CashBankID).
			Update("balance", correctBalance).Error; err != nil {
			tx.Rollback()
			log.Printf("    ❌ Failed to update cash bank balance: %v", err)
			continue
		}

		// Update COA balance
		if err := tx.Model(&models.Account{}).
			Where("id = ?", link.COAAccountID).
			Update("balance", correctBalance).Error; err != nil {
			tx.Rollback()
			log.Printf("    ❌ Failed to update COA balance: %v", err)
			continue
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("    ❌ Failed to commit sync: %v", err)
			continue
		}

		log.Printf("    ✅ Synced to balance: %.2f", correctBalance)
		syncCount++
	}

	log.Printf("  ✅ Successfully synchronized %d cash bank accounts", syncCount)
	return nil
}

// updateParentAccountBalances updates parent account balances to be sum of children
func updateParentAccountBalances(db *gorm.DB) error {
	log.Println("\n🏗️ STEP 3: Updating Parent Account Balances...")

	// 3.1 Get hierarchy levels with recursive CTE
	type AccountHierarchy struct {
		ID       uint    `json:"id"`
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		ParentID *uint   `json:"parent_id"`
		IsHeader bool    `json:"is_header"`
		Level    int     `json:"level"`
		Balance  float64 `json:"balance"`
	}

	var allAccounts []AccountHierarchy
	hierarchyQuery := `
		WITH RECURSIVE account_hierarchy AS (
			-- Base case: accounts without parents (level 0)
			SELECT 
				id, code, name, parent_id, is_header, balance,
				0 as level
			FROM accounts 
			WHERE parent_id IS NULL AND deleted_at IS NULL
			
			UNION ALL
			
			-- Recursive case: accounts with parents
			SELECT 
				a.id, a.code, a.name, a.parent_id, a.is_header, a.balance,
				ah.level + 1 as level
			FROM accounts a
			JOIN account_hierarchy ah ON a.parent_id = ah.id
			WHERE a.deleted_at IS NULL
		)
		SELECT * FROM account_hierarchy ORDER BY level DESC, code;
	`

	if err := db.Raw(hierarchyQuery).Scan(&allAccounts).Error; err != nil {
		return fmt.Errorf("failed to calculate account hierarchy: %w", err)
	}

	log.Printf("  Found %d accounts in hierarchy", len(allAccounts))

	// 3.2 Find maximum level
	maxLevel := 0
	for _, acc := range allAccounts {
		if acc.Level > maxLevel {
			maxLevel = acc.Level
		}
	}
	log.Printf("  Maximum hierarchy level: %d", maxLevel)

	// 3.3 Update balances from deepest level to root
	updateCount := 0
	for level := maxLevel; level >= 0; level-- {
		log.Printf("    Processing level %d...", level)

		// Get all header accounts at this level
		var headerAccounts []AccountHierarchy
		for _, acc := range allAccounts {
			if acc.Level == level && acc.IsHeader {
				headerAccounts = append(headerAccounts, acc)
			}
		}

		// Update each header account's balance
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
				log.Printf("      ❌ Failed to update balance for %s: %v", header.Code, err)
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

// validateFinalResults validates that all fixes were applied correctly
func validateFinalResults(db *gorm.DB) error {
	log.Println("\n🔍 STEP 4: Validating Final Results...")

	// 4.1 Check cash bank vs COA sync
	log.Println("  4.1 Checking Cash Bank vs COA synchronization...")
	
	type SyncCheck struct {
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		CashBalance float64 `json:"cash_balance"`
		COABalance  float64 `json:"coa_balance"`
		Difference  float64 `json:"difference"`
	}

	var syncChecks []SyncCheck
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
		AND ABS(cb.balance - a.balance) > 0.01
		ORDER BY ABS(cb.balance - a.balance) DESC;
	`

	if err := db.Raw(syncQuery).Scan(&syncChecks).Error; err != nil {
		return fmt.Errorf("failed to check sync status: %w", err)
	}

	if len(syncChecks) == 0 {
		log.Println("    ✅ All cash bank accounts are synchronized with COA")
	} else {
		log.Printf("    ⚠️ Found %d accounts with sync discrepancies:", len(syncChecks))
		for _, check := range syncChecks {
			log.Printf("      %s: CashBank=%.2f, COA=%.2f, Diff=%.2f", 
				check.Code, check.CashBalance, check.COABalance, check.Difference)
		}
	}

	// 4.2 Check parent-child balance consistency
	log.Println("  4.2 Checking parent-child balance consistency...")
	
	type ParentChildCheck struct {
		Code         string  `json:"code"`
		Name         string  `json:"name"`
		ParentBalance float64 `json:"parent_balance"`
		ChildrenSum   float64 `json:"children_sum"`
		Difference    float64 `json:"difference"`
	}

	var parentChecks []ParentChildCheck
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
		ORDER BY ABS(p.balance - COALESCE(children_sum.total, 0)) DESC;
	`

	if err := db.Raw(parentQuery).Scan(&parentChecks).Error; err != nil {
		return fmt.Errorf("failed to check parent consistency: %w", err)
	}

	if len(parentChecks) == 0 {
		log.Println("    ✅ All parent accounts have correct child sum balances")
	} else {
		log.Printf("    ⚠️ Found %d parent accounts with inconsistent balances:", len(parentChecks))
		for _, check := range parentChecks {
			log.Printf("      %s: Parent=%.2f, ChildrenSum=%.2f, Diff=%.2f", 
				check.Code, check.ParentBalance, check.ChildrenSum, check.Difference)
		}
	}

	// 4.3 Show final balance summary
	log.Println("  4.3 Final balance summary...")
	
	type BalanceSummary struct {
		AccountType  string  `json:"account_type"`
		TotalBalance float64 `json:"total_balance"`
		AccountCount int     `json:"account_count"`
	}

	var summaries []BalanceSummary
	summaryQuery := `
		SELECT 
			type as account_type,
			COALESCE(SUM(balance), 0) as total_balance,
			COUNT(*) as account_count
		FROM accounts 
		WHERE deleted_at IS NULL 
		AND type IN ('ASSET', 'LIABILITY', 'EQUITY')
		GROUP BY type
		ORDER BY type;
	`

	if err := db.Raw(summaryQuery).Scan(&summaries).Error; err != nil {
		return fmt.Errorf("failed to get balance summary: %w", err)
	}

	log.Println("    📊 Balance Summary by Type:")
	var assetsTotal, liabilitiesTotal, equityTotal float64
	for _, summary := range summaries {
		log.Printf("      %s: %.2f (%d accounts)", 
			summary.AccountType, summary.TotalBalance, summary.AccountCount)
		
		switch summary.AccountType {
		case "ASSET":
			assetsTotal = summary.TotalBalance
		case "LIABILITY":
			liabilitiesTotal = summary.TotalBalance
		case "EQUITY":
			equityTotal = summary.TotalBalance
		}
	}

	// Check balance equation: Assets = Liabilities + Equity
	balanceEquation := assetsTotal - (liabilitiesTotal + equityTotal)
	log.Printf("    📈 Balance Sheet Equation Check:")
	log.Printf("      Assets: %.2f", assetsTotal)
	log.Printf("      Liabilities: %.2f", liabilitiesTotal)
	log.Printf("      Equity: %.2f", equityTotal)
	log.Printf("      Difference (should be 0): %.2f", balanceEquation)
	
	if math.Abs(balanceEquation) <= 0.01 {
		log.Println("      ✅ Balance sheet is balanced!")
	} else {
		log.Printf("      ⚠️ Balance sheet has a difference of %.2f", balanceEquation)
	}

	log.Println("  ✅ Final validation completed")
	return nil
}