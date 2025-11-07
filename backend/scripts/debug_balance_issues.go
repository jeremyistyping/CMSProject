package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔍 Debugging Balance Issues...")

	// Initialize database connection
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Debug cash bank issues
	debugCashBankIssues(db)
	
	// Debug balance sheet issues
	debugBalanceSheetIssues(db)
}

func debugCashBankIssues(db *gorm.DB) {
	log.Println("\n🔍 DEBUGGING CASH BANK ISSUES:")

	// Check specific cash bank CSH-2025-0001
	type CashBankDetail struct {
		ID        uint    `json:"id"`
		Code      string  `json:"code"`
		Name      string  `json:"name"`
		Balance   float64 `json:"balance"`
		AccountID uint    `json:"account_id"`
		IsActive  bool    `json:"is_active"`
	}

	var cashBank CashBankDetail
	query := `
		SELECT id, code, name, balance, account_id, is_active
		FROM cash_banks 
		WHERE code = 'CSH-2025-0001' 
		AND deleted_at IS NULL
	`

	if err := db.Raw(query).Scan(&cashBank).Error; err != nil {
		log.Printf("❌ Failed to fetch cash bank details: %v", err)
		return
	}

	log.Printf("📊 Cash Bank CSH-2025-0001 Details:")
	log.Printf("  ID: %d", cashBank.ID)
	log.Printf("  Name: %s", cashBank.Name)
	log.Printf("  Balance: %.2f", cashBank.Balance)
	log.Printf("  Account ID: %d", cashBank.AccountID)
	log.Printf("  Is Active: %t", cashBank.IsActive)

	// Check linked COA account
	type COADetail struct {
		ID      uint    `json:"id"`
		Code    string  `json:"code"`
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
		Type    string  `json:"type"`
	}

	var coaAccount COADetail
	coaQuery := `
		SELECT id, code, name, balance, type
		FROM accounts 
		WHERE id = ? 
		AND deleted_at IS NULL
	`

	if err := db.Raw(coaQuery, cashBank.AccountID).Scan(&coaAccount).Error; err != nil {
		log.Printf("❌ Failed to fetch COA account details: %v", err)
		return
	}

	log.Printf("📋 Linked COA Account Details:")
	log.Printf("  ID: %d", coaAccount.ID)
	log.Printf("  Code: %s", coaAccount.Code)
	log.Printf("  Name: %s", coaAccount.Name)
	log.Printf("  Balance: %.2f", coaAccount.Balance)
	log.Printf("  Type: %s", coaAccount.Type)

	// Check transactions for this cash bank
	type TransactionSum struct {
		TotalAmount float64 `json:"total_amount"`
		TxCount     int     `json:"tx_count"`
	}

	var txSum TransactionSum
	txQuery := `
		SELECT 
			COALESCE(SUM(amount), 0) as total_amount,
			COUNT(*) as tx_count
		FROM cash_bank_transactions 
		WHERE cash_bank_id = ? 
		AND deleted_at IS NULL
	`

	if err := db.Raw(txQuery, cashBank.ID).Scan(&txSum).Error; err != nil {
		log.Printf("❌ Failed to fetch transaction sum: %v", err)
		return
	}

	log.Printf("💰 Transaction Summary:")
	log.Printf("  Total Amount: %.2f", txSum.TotalAmount)
	log.Printf("  Transaction Count: %d", txSum.TxCount)

	// Check the hierarchy of this account
	checkAccountHierarchy(db, coaAccount.ID)
}

func checkAccountHierarchy(db *gorm.DB, accountID uint) {
	log.Printf("\n🏗️ CHECKING ACCOUNT HIERARCHY FOR ID %d:", accountID)

	type HierarchyInfo struct {
		ID       uint    `json:"id"`
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Balance  float64 `json:"balance"`
		ParentID *uint   `json:"parent_id"`
		IsHeader bool    `json:"is_header"`
		Level    int     `json:"level"`
	}

	// Get full path to root
	hierarchyQuery := `
		WITH RECURSIVE account_path AS (
			-- Start from the specific account
			SELECT 
				id, code, name, balance, parent_id, is_header,
				0 as level
			FROM accounts 
			WHERE id = ? AND deleted_at IS NULL
			
			UNION ALL
			
			-- Get parents recursively
			SELECT 
				p.id, p.code, p.name, p.balance, p.parent_id, p.is_header,
				ap.level + 1 as level
			FROM accounts p
			JOIN account_path ap ON p.id = ap.parent_id
			WHERE p.deleted_at IS NULL
		)
		SELECT * FROM account_path ORDER BY level;
	`

	var hierarchy []HierarchyInfo
	if err := db.Raw(hierarchyQuery, accountID).Scan(&hierarchy).Error; err != nil {
		log.Printf("❌ Failed to get account hierarchy: %v", err)
		return
	}

	for _, acc := range hierarchy {
		indent := ""
		for i := 0; i < acc.Level; i++ {
			indent += "  "
		}
		
		headerStatus := ""
		if acc.IsHeader {
			headerStatus = " [HEADER]"
		}
		
		parentInfo := "ROOT"
		if acc.ParentID != nil {
			parentInfo = fmt.Sprintf("Parent: %d", *acc.ParentID)
		}
		
		log.Printf("  %s%s (%s) - Balance: %.2f%s - %s", 
			indent, acc.Code, acc.Name, acc.Balance, headerStatus, parentInfo)
	}
}

func debugBalanceSheetIssues(db *gorm.DB) {
	log.Println("\n📊 DEBUGGING BALANCE SHEET ISSUES:")

	// Get detailed breakdown by account type and check for issues
	type AccountTypeBreakdown struct {
		Type         string  `json:"type"`
		TotalBalance float64 `json:"total_balance"`
		AccountCount int     `json:"account_count"`
		HeaderCount  int     `json:"header_count"`
		LeafCount    int     `json:"leaf_count"`
	}

	var breakdown []AccountTypeBreakdown
	breakdownQuery := `
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

	if err := db.Raw(breakdownQuery).Scan(&breakdown).Error; err != nil {
		log.Printf("❌ Failed to get account breakdown: %v", err)
		return
	}

	log.Println("📈 Account Type Breakdown:")
	for _, acc := range breakdown {
		log.Printf("  %s:", acc.Type)
		log.Printf("    Total Balance: %.2f", acc.TotalBalance)
		log.Printf("    Total Accounts: %d (Headers: %d, Leafs: %d)", 
			acc.AccountCount, acc.HeaderCount, acc.LeafCount)
	}

	// Check for accounts with unusual balances
	log.Println("\n🔍 Checking for unusual balances:")

	type UnusualBalance struct {
		Code    string  `json:"code"`
		Name    string  `json:"name"`
		Type    string  `json:"type"`
		Balance float64 `json:"balance"`
	}

	var unusualBalances []UnusualBalance
	unusualQuery := `
		SELECT code, name, type, balance
		FROM accounts 
		WHERE deleted_at IS NULL 
		AND (
			(type IN ('ASSET', 'EXPENSE') AND balance < 0) OR
			(type IN ('LIABILITY', 'EQUITY', 'REVENUE') AND balance > 0) OR
			ABS(balance) > 100000000
		)
		ORDER BY ABS(balance) DESC;
	`

	if err := db.Raw(unusualQuery).Scan(&unusualBalances).Error; err != nil {
		log.Printf("❌ Failed to check unusual balances: %v", err)
		return
	}

	if len(unusualBalances) == 0 {
		log.Println("  ✅ No unusual balances found")
	} else {
		log.Printf("  ⚠️ Found %d accounts with unusual balances:", len(unusualBalances))
		for _, acc := range unusualBalances {
			sign := ""
			if (acc.Type == "ASSET" || acc.Type == "EXPENSE") && acc.Balance < 0 {
				sign = " [NEGATIVE - UNUSUAL]"
			} else if (acc.Type == "LIABILITY" || acc.Type == "EQUITY" || acc.Type == "REVENUE") && acc.Balance > 0 {
				sign = " [POSITIVE - UNUSUAL]"
			} else if acc.Balance > 100000000 || acc.Balance < -100000000 {
				sign = " [VERY LARGE]"
			}
			
			log.Printf("    %s (%s) %s: %.2f%s", 
				acc.Code, acc.Name, acc.Type, acc.Balance, sign)
		}
	}
}