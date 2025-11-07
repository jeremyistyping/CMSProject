package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"gorm.io/gorm"
)

func main() {
	// Load config and connect to database using existing system
	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("🔍 Chart of Accounts Balance Diagnosis")
	fmt.Println("=====================================")

	// Test AccountResolver
	fmt.Println("\n1. Testing AccountResolver:")
	testAccountResolver(db)

	// Check key account balances
	fmt.Println("\n2. Key Account Balances:")
	checkKeyAccountBalances(db)

	// Check recent sales
	fmt.Println("\n3. Recent Sales Analysis:")
	checkRecentSales(db)

	fmt.Println("\n✅ Diagnosis complete!")
}

func testAccountResolver(db *gorm.DB) {
	resolver := services.NewAccountResolver(db)
	
	testCases := []services.AccountType{
		services.AccountTypeAccountsReceivable,
		services.AccountTypeSalesRevenue,
		services.AccountTypeBank,
		services.AccountTypeCash,
	}

	for _, accountType := range testCases {
		account, err := resolver.GetAccount(accountType)
		if err != nil {
			fmt.Printf("❌ %s: ERROR - %v\n", accountType, err)
		} else {
			fmt.Printf("✅ %s: %s (%s) - Balance: %.2f\n", 
				accountType, account.Name, account.Code, account.Balance)
		}
	}
}

func checkKeyAccountBalances(db *gorm.DB) {
	accountCodes := []string{"1201", "4101", "1102", "1103", "1104"}
	
	for _, code := range accountCodes {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			fmt.Printf("❌ Account %s: Not found\n", code)
			continue
		}

		normalBalance := account.GetNormalBalance()
		status := "✅ CORRECT"
		
		if normalBalance == models.NormalBalanceDebit && account.Balance < 0 {
			status = "❌ INCORRECT (should be positive for debit balance)"
		} else if normalBalance == models.NormalBalanceCredit && account.Balance > 0 {
			status = "❌ INCORRECT (should be negative for credit balance)"
		}

		fmt.Printf("%s (%s): Balance %.2f, Type: %s, Normal: %s - %s\n",
			code, account.Name, account.Balance, account.Type, normalBalance, status)
	}
}

func checkRecentSales(db *gorm.DB) {
	var sales []models.Sale
	err := db.Where("created_at > NOW() - INTERVAL '7 days'").
		Order("created_at DESC").
		Limit(5).
		Find(&sales).Error

	if err != nil {
		fmt.Printf("❌ Error fetching recent sales: %v\n", err)
		return
	}

	fmt.Printf("Found %d recent sales:\n", len(sales))
	for _, sale := range sales {
		fmt.Printf("Sale ID %d - Status: %s, Total: %.2f, Outstanding: %.2f\n",
			sale.ID, sale.Status, sale.TotalAmount, sale.OutstandingAmount)
	}
}