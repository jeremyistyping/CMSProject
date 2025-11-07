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

	// Get all accounts
	var accounts []models.Account
	if err := db.Order("code").Find(&accounts).Error; err != nil {
		log.Fatalf("Failed to get accounts: %v", err)
	}

	fmt.Printf("Total accounts found: %d\n\n", len(accounts))

	// Print all accounts with details
	for _, account := range accounts {
		parentInfo := "ROOT"
		if account.ParentID != nil {
			var parent models.Account
			if err := db.First(&parent, *account.ParentID).Error; err == nil {
				parentInfo = fmt.Sprintf("%s (%d)", parent.Code, *account.ParentID)
			} else {
				parentInfo = fmt.Sprintf("INVALID_PARENT (%d)", *account.ParentID)
			}
		}

		fmt.Printf("ID: %d | Code: %s | Name: %s | Type: %s | Level: %d | IsHeader: %t | Balance: %.2f | Parent: %s\n",
			account.ID, account.Code, account.Name, account.Type, account.Level, account.IsHeader, account.Balance, parentInfo)
	}

	// Check for duplicates
	fmt.Println("\nChecking for duplicate codes:")
	codeCount := make(map[string]int)
	for _, account := range accounts {
		codeCount[account.Code]++
	}

	duplicates := false
	for code, count := range codeCount {
		if count > 1 {
			fmt.Printf("DUPLICATE: Code '%s' appears %d times\n", code, count)
			duplicates = true
		}
	}

	if !duplicates {
		fmt.Println("No duplicate codes found.")
	}
}
