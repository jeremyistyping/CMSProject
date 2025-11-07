package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("=== CHECKING ACCOUNTS TABLE (ASSET Type) ===")
	var accounts []models.Account
	result := db.Where("type = ? AND deleted_at IS NULL", "ASSET").
		Order("code ASC").
		Find(&accounts)

	if result.Error != nil {
		log.Fatal("Failed to query accounts:", result.Error)
	}

	fmt.Printf("%-4s %-12s %-30s %-8s %-15s %-10s %s\n", 
		"ID", "Code", "Name", "Balance", "Category", "Active", "Header")
	fmt.Println("====================================================================================")
	
	for _, account := range accounts {
		fmt.Printf("%-4d %-12s %-30s %-8.0f %-15s %-10t %t\n",
			account.ID, account.Code, account.Name, account.Balance, 
			account.Category, account.IsActive, account.IsHeader)
	}

	fmt.Printf("\nTotal Asset Accounts: %d\n", len(accounts))
}
