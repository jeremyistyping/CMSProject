package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	db := database.ConnectDB()
	
	fmt.Println("=== DEBUGGING FindAll() QUERY ===")
	
	// Test the exact same query as FindAll()
	fmt.Println("\n1. Raw query with same conditions as FindAll():")
	rows, _ := db.Raw(`
		SELECT id, code, name, type, balance, is_active, deleted_at
		FROM cash_banks 
		WHERE is_active = true AND deleted_at IS NULL
		ORDER BY id
	`).Rows()
	defer rows.Close()
	
	allAccountsCount := 0
	for rows.Next() {
		var id uint
		var code, name, cbType string
		var balance float64
		var isActive bool
		var deletedAt *string
		
		rows.Scan(&id, &code, &name, &cbType, &balance, &isActive, &deletedAt)
		fmt.Printf("ID: %d | Code: %s | Name: %s | Type: %s | Balance: %.2f | Active: %t | Deleted: %v\n", 
			id, code, name, cbType, balance, isActive, deletedAt)
		allAccountsCount++
	}
	fmt.Printf("Total from raw query: %d\n", allAccountsCount)
	
	// Test with GORM
	fmt.Println("\n2. Using GORM Find with same conditions:")
	var accounts []models.CashBank
	err := db.Where("is_active = ?", true).Find(&accounts).Error
	if err != nil {
		fmt.Printf("GORM Error: %v\n", err)
		return
	}
	
	fmt.Printf("Total from GORM: %d\n", len(accounts))
	for _, account := range accounts {
		fmt.Printf("ID: %d | Code: %s | Name: %s | Type: %s | Balance: %.2f | Active: %t\n", 
			account.ID, account.Code, account.Name, account.Type, account.Balance, account.IsActive)
	}
	
	// Test with Preload like in repository
	fmt.Println("\n3. Using GORM Find with Preload (same as FindAll()):")
	var preloadAccounts []models.CashBank
	err = db.Preload("Account").Where("is_active = ?", true).Find(&preloadAccounts).Error
	if err != nil {
		fmt.Printf("GORM Preload Error: %v\n", err)
		return
	}
	
	fmt.Printf("Total from GORM with Preload: %d\n", len(preloadAccounts))
	for _, account := range preloadAccounts {
		accountName := "NULL"
		if account.Account.ID > 0 {
			accountName = account.Account.Name
		}
		fmt.Printf("ID: %d | Code: %s | Name: %s | Type: %s | Balance: %.2f | Active: %t | COA: %s\n", 
			account.ID, account.Code, account.Name, account.Type, account.Balance, account.IsActive, accountName)
	}
	
	// Check specific account ID 7
	fmt.Println("\n4. Checking specific account ID 7:")
	rows2, _ := db.Raw(`
		SELECT id, code, name, type, balance, is_active, account_id, deleted_at
		FROM cash_banks 
		WHERE id = 7
	`).Rows()
	defer rows2.Close()
	
	found := false
	for rows2.Next() {
		var id uint
		var code, name, cbType string
		var balance float64
		var isActive bool
		var accountID *uint
		var deletedAt *string
		
		rows2.Scan(&id, &code, &name, &cbType, &balance, &isActive, &accountID, &deletedAt)
		fmt.Printf("ID: %d | Code: %s | Name: %s | Type: %s | Balance: %.2f | Active: %t | AccountID: %v | Deleted: %v\n", 
			id, code, name, cbType, balance, isActive, accountID, deletedAt)
		found = true
	}
	
	if !found {
		fmt.Println("Account ID 7 not found in database!")
	}
	
	// Check if account_id = 5 exists in accounts table
	fmt.Println("\n5. Checking if linked COA account exists:")
	rows3, _ := db.Raw(`
		SELECT cb.id, cb.name as cb_name, cb.account_id, a.id as coa_id, a.name as coa_name, a.is_active as coa_active
		FROM cash_banks cb
		LEFT JOIN accounts a ON cb.account_id = a.id
		WHERE cb.id = 7
	`).Rows()
	defer rows3.Close()
	
	for rows3.Next() {
		var cbID uint
		var cbName string
		var accountID *uint
		var coaID *uint
		var coaName *string
		var coaActive *bool
		
		rows3.Scan(&cbID, &cbName, &accountID, &coaID, &coaName, &coaActive)
		fmt.Printf("CashBank ID: %d | CB Name: %s | AccountID: %v | COA ID: %v | COA Name: %v | COA Active: %v\n", 
			cbID, cbName, accountID, coaID, coaName, coaActive)
		
		if coaID == nil || coaActive == nil || !*coaActive {
			fmt.Println("❌ PROBLEM: COA account is missing or inactive - this might cause Preload to filter out the record!")
		}
	}
}