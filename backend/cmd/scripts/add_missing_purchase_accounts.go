package main

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("🏗️ Add Missing Purchase Accounts")
	fmt.Println("=================================")

	// Database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Printf("❌ Database connection failed: %v", err)
		return
	}

	fmt.Println("✅ Database connected successfully\n")

	// Define missing accounts needed for purchase SSOT integration
	missingAccounts := []models.Account{
		{
			Code:         "1240",
			Name:         "PPN Masukan",
			Type:         models.AccountTypeAsset,
			Category:     "CURRENT_ASSETS",
			ParentID:     getParentAccountID(db, "1200"), // Try to find parent "ACCOUNTS RECEIVABLE"
			IsActive:     true,
			Balance:      0,
			Description:  "Account for input VAT (Value Added Tax) paid on purchases",
		},
		{
			Code:         "2111",
			Name:         "Utang PPh 21",
			Type:         models.AccountTypeLiability,
			Category:     "CURRENT_LIABILITIES",
			ParentID:     getParentAccountID(db, "2100"), // Try to find parent "CURRENT LIABILITIES"
			IsActive:     true,
			Balance:      0,
			Description:  "Account for PPh 21 (income tax on salary) payable",
		},
		{
			Code:         "2112",
			Name:         "Utang PPh 23",
			Type:         models.AccountTypeLiability,
			Category:     "CURRENT_LIABILITIES",
			ParentID:     getParentAccountID(db, "2100"), // Try to find parent "CURRENT LIABILITIES"
			IsActive:     true,
			Balance:      0,
			Description:  "Account for PPh 23 (withholding tax on services) payable",
		},
	}

	fmt.Println("📋 Accounts to be created:")
	for _, acc := range missingAccounts {
		fmt.Printf("  - %s: %s (%s)\n", acc.Code, acc.Name, acc.Type)
	}

	fmt.Println("\n🔄 Creating missing accounts...")
	
	var createdCount, skippedCount int

	for _, account := range missingAccounts {
		// Check if account already exists
		var existingAccount models.Account
		err := db.Where("code = ?", account.Code).First(&existingAccount).Error
		
		if err == nil {
			fmt.Printf("  📋 %s (%s): Already exists, skipping\n", account.Code, account.Name)
			skippedCount++
			continue
		}

		// Create the account
		err = db.Create(&account).Error
		if err != nil {
			fmt.Printf("  ❌ %s (%s): Failed to create - %v\n", account.Code, account.Name, err)
			continue
		}

		fmt.Printf("  ✅ %s (%s): Created successfully\n", account.Code, account.Name)
		createdCount++
	}

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  ✅ Created: %d accounts\n", createdCount)
	fmt.Printf("  📋 Skipped: %d accounts (already exist)\n", skippedCount)
	
	if createdCount > 0 {
		fmt.Printf("\n🎉 Missing accounts created! Now you can run the purchase SSOT reprocessing script.\n")
		fmt.Println("Next step: go run cmd/scripts/reprocess_existing_purchases_to_ssot.go")
	} else if skippedCount > 0 {
		fmt.Printf("\n✅ All required accounts already exist. You can proceed with purchase SSOT integration.\n")
	}

	// Verify all required accounts now exist
	fmt.Println("\n🔍 Verifying required accounts:")
	requiredAccounts := []string{"1301", "1240", "2101", "2111", "2112"}
	allAccountsExist := true

	for _, code := range requiredAccounts {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			fmt.Printf("  ❌ %s: Still missing\n", code)
			allAccountsExist = false
		} else {
			fmt.Printf("  ✅ %s: %s (Balance: Rp %.2f)\n", code, account.Name, account.Balance)
		}
	}

	if allAccountsExist {
		fmt.Printf("\n🎯 All required accounts for purchase SSOT integration are now available!\n")
	} else {
		fmt.Printf("\n⚠️ Some required accounts are still missing. Please check your account setup.\n")
	}
}

// Helper function to get parent account ID by code
func getParentAccountID(db *gorm.DB, parentCode string) *uint {
	if parentCode == "" {
		return nil
	}

	var parentAccount models.Account
	err := db.Where("code = ?", parentCode).First(&parentAccount).Error
	if err != nil {
		// Parent not found, return nil
		return nil
	}

	return &parentAccount.ID
}