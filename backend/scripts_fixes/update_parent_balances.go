package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("=== Parent Account Balance Update ===")
	fmt.Println()

	// Update parent account balances recursively from bottom up
	fmt.Println("1. Updating parent account balances...")

	// Step 1: Update all header accounts to be the sum of their children
	// We need to do this recursively from deepest level up
	
	// First, identify all accounts and their hierarchy levels
	type AccountHierarchy struct {
		ID       uint   `json:"id"`
		Code     string `json:"code"`
		Name     string `json:"name"`
		ParentID *uint  `json:"parent_id"`
		IsHeader bool   `json:"is_header"`
		Level    int    `json:"level"`
	}

	// Calculate hierarchy levels
	var allAccounts []AccountHierarchy
	err := db.Raw(`
		WITH RECURSIVE account_hierarchy AS (
			-- Base case: accounts without parents (level 0)
			SELECT 
				id, code, name, parent_id, is_header,
				0 as level
			FROM accounts 
			WHERE parent_id IS NULL AND deleted_at IS NULL
			
			UNION ALL
			
			-- Recursive case: accounts with parents
			SELECT 
				a.id, a.code, a.name, a.parent_id, a.is_header,
				ah.level + 1 as level
			FROM accounts a
			JOIN account_hierarchy ah ON a.parent_id = ah.id
			WHERE a.deleted_at IS NULL
		)
		SELECT * FROM account_hierarchy ORDER BY level DESC, code
	`).Scan(&allAccounts).Error

	if err != nil {
		log.Printf("Error calculating account hierarchy: %v", err)
		return
	}

	fmt.Printf("Found %d accounts in hierarchy\n", len(allAccounts))

	// Now update balances from deepest level to shallowest
	maxLevel := 0
	for _, acc := range allAccounts {
		if acc.Level > maxLevel {
			maxLevel = acc.Level
		}
	}

	fmt.Printf("Maximum hierarchy level: %d\n", maxLevel)

	// Process from deepest level up to root
	for level := maxLevel; level >= 0; level-- {
		fmt.Printf("  Processing level %d...\n", level)
		
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
			err = db.Raw(`
				SELECT COALESCE(SUM(balance), 0) as child_sum
				FROM accounts 
				WHERE parent_id = ? AND deleted_at IS NULL
			`, header.ID).Scan(&childSum).Error

			if err != nil {
				log.Printf("    Error calculating sum for account %s: %v", header.Code, err)
				continue
			}

			// Update the header account balance
			err = db.Model(&models.Account{}).Where("id = ?", header.ID).Update("balance", childSum).Error
			if err != nil {
				log.Printf("    Error updating balance for account %s: %v", header.Code, err)
			} else {
				fmt.Printf("    Updated %s (%s) balance to %.2f\n", header.Code, header.Name, childSum)
			}
		}
	}

	// Show final results
	fmt.Println("\n2. Final Account Balances:")
	var finalBalances []struct {
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Balance  float64 `json:"balance"`
		IsHeader bool    `json:"is_header"`
	}

	err = db.Raw(`
		SELECT code, name, balance, is_header
		FROM accounts 
		WHERE deleted_at IS NULL
		ORDER BY code
	`).Scan(&finalBalances).Error

	if err != nil {
		log.Printf("Error fetching final balances: %v", err)
	} else {
		for _, balance := range finalBalances {
			accountType := "Detail"
			if balance.IsHeader {
				accountType = "Header"
			}
			fmt.Printf("  %s: %s [%s] - Balance: %.2f\n", balance.Code, balance.Name, accountType, balance.Balance)
		}
	}

	// Verify balance sheet equation
	fmt.Println("\n3. Balance Sheet Verification:")
	var assetTotal, liabilityTotal, equityTotal float64
	
	// Assets (normally debit balance, positive)
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'ASSET' AND deleted_at IS NULL`).Scan(&assetTotal)
	
	// Liabilities (normally credit balance, negative in our system)
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'LIABILITY' AND deleted_at IS NULL`).Scan(&liabilityTotal)
	
	// Equity (normally credit balance, negative in our system)
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EQUITY' AND deleted_at IS NULL`).Scan(&equityTotal)

	// Revenue (normally credit balance, negative in our system)
	var revenueTotal float64
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'REVENUE' AND deleted_at IS NULL`).Scan(&revenueTotal)

	// Expenses (normally debit balance, positive in our system)
	var expenseTotal float64
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EXPENSE' AND deleted_at IS NULL`).Scan(&expenseTotal)

	fmt.Printf("  Assets Total: %.2f\n", assetTotal)
	fmt.Printf("  Liabilities Total: %.2f\n", liabilityTotal)
	fmt.Printf("  Equity Total: %.2f\n", equityTotal)
	fmt.Printf("  Revenue Total: %.2f\n", revenueTotal)
	fmt.Printf("  Expense Total: %.2f\n", expenseTotal)

	// Net Income = Revenue (negative) + Expenses (positive) = Revenue - Expenses
	netIncome := revenueTotal + expenseTotal
	fmt.Printf("  Net Income: %.2f\n", netIncome)

	// Balance equation: Assets = Liabilities + Equity + Net Income
	balanceEquation := assetTotal - (liabilityTotal + equityTotal + netIncome)
	fmt.Printf("  Balance Sheet Check (should be 0): %.2f\n", balanceEquation)

	if balanceEquation == 0 {
		fmt.Println("  ✅ Balance sheet is balanced!")
	} else {
		fmt.Printf("  ⚠️  Balance sheet has a difference of %.2f\n", balanceEquation)
	}

	fmt.Println("\n=== Parent Balance Update Complete ===")
}