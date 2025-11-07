package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	db := database.ConnectDB()
	var accounts []models.Account
	db.Where("type = ?", "REVENUE").Find(&accounts)
	
	fmt.Println("Existing revenue accounts in database:")
	fmt.Println("======================================")
	
	if len(accounts) == 0 {
		fmt.Println("No revenue accounts found in database!")
		return
	}
	
	for _, account := range accounts {
		fmt.Printf("ID: %d, Code: %s, Name: %s, Active: %t\n", 
			account.ID, account.Code, account.Name, account.IsActive)
	}
}