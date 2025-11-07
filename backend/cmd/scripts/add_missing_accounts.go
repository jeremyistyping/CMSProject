package main

import (
	"log"
	
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔧 Adding Missing Accounts...")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	log.Println("✅ Database connected successfully")
	
	// Check if account 1200 exists
	var account models.Account
	if err := db.Where("code = ?", "1200").First(&account).Error; err != nil {
		log.Println("🔧 Account 1200 (ACCOUNTS RECEIVABLE) not found, seeding...")
		if err := database.SeedAccounts(db); err != nil {
			log.Printf("❌ Failed to seed accounts: %v", err)
		} else {
			log.Println("✅ Accounts seeded successfully")
		}
	} else {
		log.Println("✅ Account 1200 (ACCOUNTS RECEIVABLE) already exists")
	}
	
	// Verify account 1200 was created
	if err := db.Where("code = ?", "1200").First(&account).Error; err != nil {
		log.Println("❌ Account 1200 still not found after seeding")
	} else {
		log.Printf("✅ Verified Account 1200: %s", account.Name)
	}
	
	// Check if account 1104 exists (BANK UOB)
	if err := db.Where("code = ?", "1104").First(&account).Error; err != nil {
		log.Println("✅ Account 1104 (BANK UOB) already exists from seed")
	} else {
		log.Printf("✅ Account 1104: %s (Balance: %.2f)", account.Name, account.Balance)
	}
	
	// Show account hierarchy
	log.Println("📋 Account Hierarchy:")
	
	var accounts []models.Account
	db.Where("code IN (?)", []string{"1100", "1200", "1201", "1104"}).Order("code").Find(&accounts)
	
	for _, acc := range accounts {
		log.Printf("  %s - %s (Level: %d, Header: %v)", acc.Code, acc.Name, acc.Level, acc.IsHeader)
	}
	
	log.Println("✅ Account check completed!")
}
