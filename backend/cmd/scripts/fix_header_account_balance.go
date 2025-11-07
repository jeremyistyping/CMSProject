package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	fmt.Println("=== FIXING HEADER ACCOUNT BALANCES ===")
	
	// 1. Find all header accounts with non-zero balance
	var headerAccounts []models.Account
	if err := db.Where("is_header = ? AND balance != 0 AND deleted_at IS NULL", true).Find(&headerAccounts).Error; err != nil {
		log.Fatal("Failed to get header accounts:", err)
	}

	if len(headerAccounts) == 0 {
		fmt.Println("No header accounts with non-zero balance found.")
		return
	}

	fmt.Printf("Found %d header accounts with non-zero balance:\n", len(headerAccounts))
	for _, account := range headerAccounts {
		fmt.Printf("  %s (%s): %.2f\n", account.Code, account.Name, account.Balance)
	}

	// 2. Check if there are any journal entries for these header accounts
	fmt.Println("\n=== CHECKING JOURNAL ENTRIES FOR HEADER ACCOUNTS ===")
	
	for _, account := range headerAccounts {
		var journalCount int64
		db.Table("journal_entries").
			Joins("JOIN journals ON journal_entries.journal_id = journals.id").
			Where("journal_entries.account_id = ? AND journals.deleted_at IS NULL", account.ID).
			Count(&journalCount)

		fmt.Printf("Account %s (%s) has %d journal entries\n", account.Code, account.Name, journalCount)
		
		if journalCount > 0 {
			fmt.Printf("WARNING: Header account %s has journal entries - this should not happen!\n", account.Code)
		}
	}

	// 3. Fix header account balances by setting them to zero
	fmt.Println("\n=== FIXING HEADER ACCOUNT BALANCES ===")
	
	for _, account := range headerAccounts {
		fmt.Printf("Setting balance of %s (%s) from %.2f to 0.00\n", account.Code, account.Name, account.Balance)
		
		if err := db.Model(&account).Update("balance", 0).Error; err != nil {
			log.Printf("Error updating account %s: %v", account.Code, err)
		} else {
			fmt.Printf("✅ Successfully updated account %s\n", account.Code)
		}
	}

	// 4. Verify the fix
	fmt.Println("\n=== VERIFICATION ===")
	
	// Check account 1100 specifically
	var account1100 models.Account
	if err := db.Where("code = '1100' AND deleted_at IS NULL").First(&account1100).Error; err == nil {
		fmt.Printf("Account 1100 balance after fix: %.2f\n", account1100.Balance)
	}

	// Check current assets calculation
	fmt.Println("\nCurrent assets calculation after fix:")
	var currentAssetAccounts []models.Account
	db.Where("code LIKE '11%' AND is_active = ? AND is_header = ? AND deleted_at IS NULL", true, false).Find(&currentAssetAccounts)
	
	totalCurrentAssets := 0.0
	for _, account := range currentAssetAccounts {
		totalCurrentAssets += account.Balance
	}
	fmt.Printf("Total Current Assets from detail accounts: %.2f\n", totalCurrentAssets)

	// Check what dashboard would show now
	type AccountData struct {
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
		Type    string  `json:"type"`
		Code    string  `json:"code"`
	}
	
	var topAccounts []AccountData
	db.Raw(`
		SELECT 
			code,
			name,
			ABS(balance) as balance,
			type
		FROM accounts 
		WHERE deleted_at IS NULL 
			AND is_active = true
			AND balance != 0
		ORDER BY ABS(balance) DESC
		LIMIT 5
	`).Scan(&topAccounts)
	
	fmt.Println("\nTop 5 accounts after fix:")
	if len(topAccounts) == 0 {
		fmt.Println("  No accounts with non-zero balance found")
	} else {
		for _, account := range topAccounts {
			fmt.Printf("  %s (%s): %.2f - %s\n", account.Code, account.Name, account.Balance, account.Type)
		}
	}

	fmt.Println("\n✅ Header account balance fix completed!")
}
