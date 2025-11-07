package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
)

// Script to add missing accounts needed for sales journal entries
// - 2107: Other Tax Deductions (Liability)
// - 2108: Other Tax Additions (Liability) 
// - 4102: Shipping/Service Revenue (Revenue)
//
// Usage: go run cmd/scripts/add_missing_sales_accounts.go
// Note: Reads DATABASE_URL from .env file automatically

func main() {
	// Load config from .env file (same as main app)
	cfg := config.LoadConfig()
	
	log.Println("🔧 Adding missing sales journal accounts...")
	log.Printf("📍 Connecting to database: %s", maskDatabaseURL(cfg.DatabaseURL))
	
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}


	// Define missing accounts
	accountsToAdd := []models.Account{
		{
			Code:     "2107",
			Name:     "PEMOTONGAN PAJAK LAINNYA",
			Type:     models.AccountTypeLiability,
			Category: models.CategoryCurrentLiability,
			Level:    3,
			IsHeader: false,
			IsActive: true,
			Balance:  0,
		},
		{
			Code:     "2108", 
			Name:     "PENAMBAHAN PAJAK LAINNYA",
			Type:     models.AccountTypeLiability,
			Category: models.CategoryCurrentLiability,
			Level:    3,
			IsHeader: false,
			IsActive: true,
			Balance:  0,
		},
		{
			Code:     "4102",
			Name:     "PENDAPATAN JASA/ONGKIR",
			Type:     models.AccountTypeRevenue,
			Category: models.CategoryOperatingRevenue,
			Level:    2,
			IsHeader: false,
			IsActive: true,
			Balance:  0,
		},
	}

	// Add accounts with duplicate check
	for _, account := range accountsToAdd {
		if err := addAccountIfNotExists(db, account); err != nil {
			log.Printf("❌ Failed to add account %s: %v", account.Code, err)
			os.Exit(1)
		}
	}

	log.Println("✅ Successfully added all missing accounts")
	log.Println("💡 You can now restart the backend server")
}

// maskDatabaseURL masks sensitive info in database URL for logging
func maskDatabaseURL(url string) string {
	// Simple masking: show only host and database name
	// Example: postgres://user:***@localhost/dbname -> postgres://***@localhost/dbname
	if len(url) > 20 {
		return url[:10] + "***" + url[len(url)-30:]
	}
	return "***"
}

func addAccountIfNotExists(db *gorm.DB, account models.Account) error {
	// Check if account already exists
	var existing models.Account
	err := db.Where("code = ? AND deleted_at IS NULL", account.Code).First(&existing).Error
	
	if err == gorm.ErrRecordNotFound {
		// Account doesn't exist, create it
		if err := db.Create(&account).Error; err != nil {
			return fmt.Errorf("failed to create account: %v", err)
		}
		log.Printf("✅ Created account: %s - %s", account.Code, account.Name)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to check existing account: %v", err)
	}

	// Account already exists
	log.Printf("ℹ️  Account %s (%s) already exists, skipping", account.Code, account.Name)
	return nil
}
