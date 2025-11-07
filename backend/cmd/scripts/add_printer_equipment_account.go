package main

import (
	"log"
	
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🖨️ Adding Printer & Computer Equipment Accounts...")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	log.Println("✅ Database connected successfully")
	
	// Get parent account ID (1500 - FIXED ASSETS)
	var parentAccount models.Account
	if err := db.Where("code = ?", "1500").First(&parentAccount).Error; err != nil {
		log.Fatal("❌ Parent account 1500 (FIXED ASSETS) not found")
	}
	
	// Define new equipment accounts
	newAccounts := []models.Account{
		{
			Code: "1504",
			Name: "Peralatan Komputer",
			Type: models.AccountTypeAsset,
			Category: models.CategoryFixedAsset,
			Level: 3,
			IsHeader: false,
			IsActive: true,
			Balance: 0,
			ParentID: &parentAccount.ID,
			Description: "Computer equipment including laptops, desktops, monitors",
		},
		{
			Code: "1505", 
			Name: "Printer dan Mesin Kantor",
			Type: models.AccountTypeAsset,
			Category: models.CategoryFixedAsset,
			Level: 3,
			IsHeader: false,
			IsActive: true,
			Balance: 0,
			ParentID: &parentAccount.ID,
			Description: "Printers, copiers, scanners and other office machines",
		},
	}
	
	// Add each account if it doesn't exist
	for _, account := range newAccounts {
		var existingAccount models.Account
		result := db.Where("code = ?", account.Code).First(&existingAccount)
		
		if result.Error != nil {
			// Account doesn't exist, create it
			if err := db.Create(&account).Error; err != nil {
				log.Printf("❌ Failed to create account %s: %v", account.Code, err)
			} else {
				log.Printf("✅ Created account %s - %s", account.Code, account.Name)
			}
		} else {
			// Account exists, update it but preserve balance
			log.Printf("🔒 Account %s already exists, updating details (preserving balance: %.2f)", 
				existingAccount.Code, existingAccount.Balance)
			existingAccount.Name = account.Name
			existingAccount.Type = account.Type
			existingAccount.Category = account.Category
			existingAccount.Level = account.Level
			existingAccount.IsHeader = account.IsHeader
			existingAccount.IsActive = account.IsActive
			existingAccount.Description = account.Description
			existingAccount.ParentID = account.ParentID
			
			if err := db.Save(&existingAccount).Error; err != nil {
				log.Printf("❌ Failed to update account %s: %v", account.Code, err)
			} else {
				log.Printf("✅ Updated account %s - %s", account.Code, account.Name)
			}
		}
	}
	
	// Verify the accounts were created/updated
	log.Println("📋 Fixed Asset Accounts:")
	var fixedAssets []models.Account
	db.Where("parent_id = ?", parentAccount.ID).Order("code").Find(&fixedAssets)
	
	for _, acc := range fixedAssets {
		log.Printf("  %s - %s (Balance: %.2f)", acc.Code, acc.Name, acc.Balance)
	}
	
	log.Println("🖨️ Printer & Computer Equipment accounts setup completed!")
}
