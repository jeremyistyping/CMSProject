package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	fmt.Println("=== DEBUG CURRENT ASSETS ===")
	
	// 1. Get all current asset accounts (code starts with 11)
	var currentAssetAccounts []models.Account
	if err := db.Where("code LIKE '11%' AND is_active = ? AND deleted_at IS NULL", true).Find(&currentAssetAccounts).Error; err != nil {
		log.Fatal("Failed to get current asset accounts:", err)
	}

	fmt.Println("\n1. CURRENT ASSET ACCOUNTS:")
	fmt.Println("Code\t\tName\t\t\t\tBalance")
	fmt.Println("----\t\t----\t\t\t\t-------")
	
	totalCurrentAssets := 0.0
	for _, account := range currentAssetAccounts {
		fmt.Printf("%s\t\t%-20s\t\t%.2f\n", account.Code, account.Name, account.Balance)
		if !account.IsHeader {
			totalCurrentAssets += account.Balance
		}
	}
	
	fmt.Printf("\nTotal Current Assets (dari account.balance): %.2f\n", totalCurrentAssets)
	
	// 2. Calculate based on journal entries
	fmt.Println("\n2. CALCULATION BASED ON JOURNAL ENTRIES:")
	
	totalFromJournal := 0.0
	for _, account := range currentAssetAccounts {
		if account.IsHeader {
			continue
		}
		
		var totalDebits, totalCredits float64
		db.Table("journal_entries").
			Joins("JOIN journals ON journal_entries.journal_id = journals.id").
			Where("journal_entries.account_id = ? AND journals.status = 'POSTED' AND journals.deleted_at IS NULL", account.ID).
			Select("COALESCE(SUM(journal_entries.debit_amount), 0) as total_debits, COALESCE(SUM(journal_entries.credit_amount), 0) as total_credits").
			Row().Scan(&totalDebits, &totalCredits)
		
		// For assets: Debit increases, Credit decreases
		journalBalance := account.Balance + totalDebits - totalCredits
		totalFromJournal += journalBalance
		
		fmt.Printf("Account %s (%s):\n", account.Code, account.Name)
		fmt.Printf("  - Starting Balance: %.2f\n", account.Balance)
		fmt.Printf("  - Total Debits: %.2f\n", totalDebits)
		fmt.Printf("  - Total Credits: %.2f\n", totalCredits)
		fmt.Printf("  - Journal Balance: %.2f\n", journalBalance)
		fmt.Println()
	}
	
	fmt.Printf("Total Current Assets (dari journal calculation): %.2f\n", totalFromJournal)
	
	// 3. Check what's displayed in dashboard (from topAccounts query)
	fmt.Println("\n3. TOP ACCOUNTS QUERY (Dashboard):")
	
	type AccountData struct {
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
		Type    string  `json:"type"`
		Code    string  `json:"code"`
	}
	
	var topAccounts []AccountData
	db.Raw(`
		SELECT 
			name,
			code,
			ABS(balance) as balance,
			type
		FROM accounts 
		WHERE deleted_at IS NULL 
			AND is_active = true
			AND balance != 0
		ORDER BY ABS(balance) DESC
		LIMIT 10
	`).Scan(&topAccounts)
	
	fmt.Println("Top Accounts by Balance:")
	for _, account := range topAccounts {
		fmt.Printf("  %s (%s): %.2f - %s\n", account.Code, account.Name, account.Balance, account.Type)
	}
	
	// 4. Check if there's any account with balance 555000
	var accountWith555000 []models.Account
	db.Where("ABS(balance) = 555000 AND deleted_at IS NULL").Find(&accountWith555000)
	
	if len(accountWith555000) > 0 {
		fmt.Println("\n4. ACCOUNTS WITH BALANCE 555000:")
		for _, account := range accountWith555000 {
			fmt.Printf("  %s (%s): %.2f\n", account.Code, account.Name, account.Balance)
		}
	} else {
		fmt.Println("\n4. NO ACCOUNTS WITH BALANCE 555000 FOUND")
	}
	
	// 5. Check specific account 1100
	var account1100 models.Account
	if err := db.Where("code = '1100' AND deleted_at IS NULL").First(&account1100).Error; err == nil {
		fmt.Printf("\n5. ACCOUNT 1100 (CURRENT ASSETS header): %.2f\n", account1100.Balance)
	} else {
		fmt.Println("\n5. ACCOUNT 1100 NOT FOUND")
	}

	os.Exit(0)
}
