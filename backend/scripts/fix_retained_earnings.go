package main

import (
	"log"
	
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Println("🔍 Checking Retained Earnings Account...")
	
	// Initialize database
	db := database.ConnectDB()
	
	// Get accounting config
	accountingConfig := config.GetAccountingConfig()
	retainedEarningsID := accountingConfig.DefaultAccounts.RetainedEarnings
	
	log.Printf("📋 Config says Retained Earnings ID should be: %d", retainedEarningsID)
	
	// Try to find account by ID
	var accountByID models.Account
	err := db.Unscoped().Where("id = ?", retainedEarningsID).First(&accountByID).Error
	if err == nil {
		log.Printf("✅ Found account by ID %d: Code=%s, Name=%s", 
			accountByID.ID, accountByID.Code, accountByID.Name)
		if accountByID.DeletedAt.Valid {
			log.Printf("⚠️  WARNING: Account is soft-deleted! DeletedAt=%v", accountByID.DeletedAt.Time)
			log.Println("🔧 Restoring account...")
			if err := db.Model(&accountByID).Update("deleted_at", nil).Error; err != nil {
				log.Fatalf("Failed to restore account: %v", err)
			}
			log.Println("✅ Account restored successfully!")
		} else {
			log.Println("✅ Account is active and available")
		}
		return
	}
	
	log.Printf("❌ Account ID %d not found: %v", retainedEarningsID, err)
	
	// Try to find by code
	var accountByCode models.Account
	err = db.Where("code = ? AND deleted_at IS NULL", "3201").First(&accountByCode).Error
	if err == nil {
		log.Printf("✅ Found account by code '3201': ID=%d, Name=%s", 
			accountByCode.ID, accountByCode.Name)
		log.Printf("⚠️  ID mismatch! Config expects %d but found %d", 
			retainedEarningsID, accountByCode.ID)
		log.Printf("💡 You need to update accounting_config.json to use ID %d", accountByCode.ID)
		return
	}
	
	log.Println("❌ Account with code '3201' not found either")
	log.Println("🔧 Creating Retained Earnings account...")
	
	// Find parent (Equity 3000)
	var parentAccount models.Account
	err = db.Where("code = ?", "3000").First(&parentAccount).Error
	if err != nil {
		log.Fatalf("Parent account '3000' (EQUITY) not found: %v", err)
	}
	
	// Create the account with the expected ID
	newAccount := models.Account{
		Code:     "3201",
		Name:     "LABA DITAHAN",
		Type:     models.AccountTypeEquity,
		Category: models.CategoryEquity,
		Level:    2,
		IsHeader: false,
		IsActive: true,
		Balance:  0,
		ParentID: &parentAccount.ID,
	}
	
	// Set the ID explicitly if possible (may not work on all DB engines)
	if retainedEarningsID > 0 {
		newAccount.ID = retainedEarningsID
	}
	
	if err := db.Create(&newAccount).Error; err != nil {
		log.Fatalf("Failed to create account: %v", err)
	}
	
	log.Printf("✅ Created Retained Earnings account: ID=%d, Code=%s, Name=%s", 
		newAccount.ID, newAccount.Code, newAccount.Name)
	
	if newAccount.ID != retainedEarningsID {
		log.Printf("⚠️  WARNING: Created with ID %d but config expects %d", 
			newAccount.ID, retainedEarningsID)
		log.Printf("💡 Update accounting_config.json 'retained_earnings' to %d", newAccount.ID)
	}
}
