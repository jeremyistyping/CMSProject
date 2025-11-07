package main

import (
	"fmt"
	"log"
	"strings"

	"app-sistem-akuntansi/models"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("LIST ALL EXPENSE & COGS ACCOUNTS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Get all accounts starting with 5 or 6
	var accounts []models.Account
	err = db.Where("code LIKE ? OR code LIKE ?", "5%", "6%").
		Where("is_active = ?", true).
		Order("code").
		Find(&accounts).Error
	
	if err != nil {
		log.Fatalf("Failed to fetch accounts: %v", err)
	}

	if len(accounts) == 0 {
		fmt.Println("❌ No expense/COGS accounts found!")
		fmt.Println()
		fmt.Println("You need to create COGS and Expense accounts:")
		fmt.Println("  5001 - Beban Pokok Penjualan (COGS)")
		fmt.Println("  6001 - Beban Operasional (Operating Expense)")
		return
	}

	fmt.Printf("Found %d expense/COGS accounts:\n\n", len(accounts))

	cogsAccounts := []models.Account{}
	expenseAccounts := []models.Account{}

	for _, acc := range accounts {
		fmt.Printf("%s - %s (Balance: Rp %.2f)\n", acc.Code, acc.Name, acc.Balance)
		
		if strings.HasPrefix(acc.Code, "5") {
			cogsAccounts = append(cogsAccounts, acc)
			fmt.Println("   📊 Category: COGS / Cost of Goods Sold")
		} else if strings.HasPrefix(acc.Code, "6") {
			expenseAccounts = append(expenseAccounts, acc)
			fmt.Println("   📊 Category: Operating/Other Expenses")
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  COGS Accounts (5xxx):     %d accounts\n", len(cogsAccounts))
	fmt.Printf("  Expense Accounts (6xxx):  %d accounts\n", len(expenseAccounts))
	fmt.Println()

	if len(cogsAccounts) > 0 {
		fmt.Println("✅ COGS accounts available:")
		for _, acc := range cogsAccounts {
			fmt.Printf("   - %s: %s\n", acc.Code, acc.Name)
		}
	} else {
		fmt.Println("❌ No COGS accounts found. Need to create account 5001.")
	}

	fmt.Println()

	if len(expenseAccounts) > 0 {
		fmt.Println("✅ Expense accounts available:")
		for _, acc := range expenseAccounts {
			fmt.Printf("   - %s: %s\n", acc.Code, acc.Name)
		}
	} else {
		fmt.Println("❌ No expense accounts found. Need to create account 6001.")
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
}

