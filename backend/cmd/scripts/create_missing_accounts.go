package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"

	"gorm.io/gorm"
)

func main() {
	log.Println("Creating Missing Accounts for Sales Confirmation...")

	// Initialize database
	db := database.ConnectDB()

	createMissingAccounts(db)
}

func createMissingAccounts(db *gorm.DB) {
	log.Println("\n=== Creating Missing Accounts ===")

	accountRepo := repositories.NewAccountRepository(db)

	// Define missing accounts to create
	missingAccounts := []struct {
		Code        string
		Name        string
		Type        models.AccountType
		Description string
	}{
		{
			Code:        "5100",
			Name:        "Cost of Goods Sold",
			Type:        models.AccountTypeExpense,
			Description: "Alternative COGS account for sales confirmation",
		},
		{
			Code:        "1300",
			Name:        "Inventory",
			Type:        models.AccountTypeAsset,
			Description: "Alternative inventory account for sales confirmation",
		},
		{
			Code:        "4102",
			Name:        "Shipping Revenue",
			Type:        models.AccountTypeRevenue,
			Description: "Revenue from shipping and delivery charges",
		},
		{
			Code:        "4100",
			Name:        "Sales Revenue",
			Type:        models.AccountTypeRevenue,
			Description: "Alternative sales revenue account",
		},
	}

	for _, acc := range missingAccounts {
		// Check if account already exists
		existingAccount, err := accountRepo.GetAccountByCode(acc.Code)
		if err == nil {
			fmt.Printf("✅ Account %s already exists: %s (ID: %d)\n", acc.Code, existingAccount.Name, existingAccount.ID)
			continue
		}

		// Create the account
		fmt.Printf("🔄 Creating account %s - %s...\n", acc.Code, acc.Name)
		
		// Create account using the repository
		accountRequest := &models.AccountCreateRequest{
			Code:           acc.Code,
			Name:           acc.Name,
			Type:           acc.Type,
			Description:    acc.Description,
			OpeningBalance: 0, // Default opening balance
		}

		createdAccount, err := accountRepo.Create(nil, accountRequest)
		if err != nil {
			fmt.Printf("❌ Failed to create account %s: %v\n", acc.Code, err)
		} else {
			fmt.Printf("✅ Successfully created account %s - %s (ID: %d)\n", 
				acc.Code, createdAccount.Name, createdAccount.ID)
		}
	}

	fmt.Println("\n=== Account Creation Complete ===")
	fmt.Println("✅ All required accounts should now be available for sales confirmation.")
	fmt.Println("💡 You can now try to confirm the sale again.")
}