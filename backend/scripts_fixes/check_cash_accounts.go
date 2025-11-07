package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔍 Checking Cash Bank Accounts and their Account Types")
	log.Printf("===================================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Get all cash banks with their associated accounts
	var cashBanks []models.CashBank
	err = db.Preload("Account").Find(&cashBanks).Error
	if err != nil {
		log.Fatalf("❌ Failed to fetch cash banks: %v", err)
	}

	log.Printf("📋 Found %d cash bank entries:", len(cashBanks))
	
	for _, cashBank := range cashBanks {
		log.Printf("\n💳 CashBank ID: %d", cashBank.ID)
		log.Printf("   Name: %s", cashBank.Name)
		log.Printf("   Active: %v", cashBank.IsActive)
		log.Printf("   Account ID: %d", cashBank.AccountID)
		
		if cashBank.Account != nil {
			log.Printf("   Account Name: %s", cashBank.Account.Name)
			log.Printf("   Account Type: %s", cashBank.Account.Type)
			log.Printf("   Account Code: %s", cashBank.Account.Code)
			
			// Check if this is considered a cash account
			isCashAccount := cashBank.Account.Type == models.AccountTypeCash || cashBank.Account.Type == models.AccountTypeBank
			log.Printf("   Is Cash Account: %v (Type=%s)", isCashAccount, cashBank.Account.Type)
		} else {
			log.Printf("   ⚠️  No associated account found!")
		}
	}

	// Also check what account types exist in general
	log.Printf("\n🏦 All Account Types in system:")
	var accounts []models.Account
	err = db.Select("DISTINCT type").Find(&accounts).Error
	if err != nil {
		log.Printf("❌ Failed to get account types: %v", err)
		return
	}

	uniqueTypes := make(map[string]int)
	db.Model(&models.Account{}).Select("type, COUNT(*) as count").Group("type").Find(&accounts)

	for _, account := range accounts {
		log.Printf("   - %s", account.Type)
	}

	// Check account type constants
	log.Printf("\n📝 Expected Account Type Constants:")
	log.Printf("   CASH: %s", models.AccountTypeCash)
	log.Printf("   BANK: %s", models.AccountTypeBank)
	log.Printf("   ACCOUNTS_RECEIVABLE: %s", models.AccountTypeAccountsReceivable)
	log.Printf("   CURRENT_ASSET: %s", models.AccountTypeCurrentAsset)
}