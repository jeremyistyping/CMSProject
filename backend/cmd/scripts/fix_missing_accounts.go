package main

import (
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load config from environment
	cfg := config.LoadConfig()

	// Connect to database using config
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Connected to database successfully")

	log.Println("=== Checking and Creating Missing Tax Accounts ===\n")

	// List of required tax accounts for sales
	requiredAccounts := []struct {
		Code        string
		Name        string
		Type        string
		ParentCode  string
		Description string
	}{
		{
			Code:        "2107",
			Name:        "PEMOTONGAN PAJAK LAINNYA",
			Type:        "LIABILITY",
			ParentCode:  "2100",
			Description: "Other tax withholdings/deductions from sales",
		},
		{
			Code:        "2108",
			Name:        "PENAMBAHAN PAJAK LAINNYA",
			Type:        "LIABILITY",
			ParentCode:  "2100",
			Description: "Other tax additions to sales",
		},
	}

	for _, reqAcc := range requiredAccounts {
		var account models.Account
		err := db.Where("code = ?", reqAcc.Code).First(&account).Error

		if err == gorm.ErrRecordNotFound {
			log.Printf("❌ Account %s (%s) NOT FOUND - Creating...", reqAcc.Code, reqAcc.Name)

			// Find parent account
			var parentAccount models.Account
			if err := db.Where("code = ?", reqAcc.ParentCode).First(&parentAccount).Error; err != nil {
				log.Printf("⚠️ Parent account %s not found: %v", reqAcc.ParentCode, err)
				log.Printf("   Creating account without parent...")
			}

			// Create new account
			newAccount := models.Account{
				Code:        reqAcc.Code,
				Name:        reqAcc.Name,
				Type:        reqAcc.Type,
				Balance:     0,
				Description: reqAcc.Description,
				IsActive:    true,
			}

			if parentAccount.ID != 0 {
				newAccount.ParentID = &parentAccount.ID
			}

			if err := db.Create(&newAccount).Error; err != nil {
				log.Printf("❌ Failed to create account %s: %v", reqAcc.Code, err)
			} else {
				log.Printf("✅ Successfully created account %s (%s) - ID: %d", 
					reqAcc.Code, reqAcc.Name, newAccount.ID)
			}
		} else if err != nil {
			log.Printf("⚠️ Error checking account %s: %v", reqAcc.Code, err)
		} else {
			log.Printf("✅ Account %s (%s) already exists - ID: %d, Balance: %.2f", 
				reqAcc.Code, account.Name, account.ID, account.Balance)
		}
	}

	log.Println("\n=== Verification Complete ===")
	
	// Verify all required accounts now exist
	log.Println("\nFinal verification:")
	allGood := true
	for _, reqAcc := range requiredAccounts {
		var account models.Account
		if err := db.Where("code = ?", reqAcc.Code).First(&account).Error; err != nil {
			log.Printf("❌ Account %s still NOT FOUND", reqAcc.Code)
			allGood = false
		} else {
			log.Printf("✅ Account %s - %s (ID: %d)", reqAcc.Code, account.Name, account.ID)
		}
	}

	if allGood {
		log.Println("\n✅ All required tax accounts are now present!")
	} else {
		log.Println("\n⚠️ Some accounts are still missing. Manual intervention may be required.")
	}
}
