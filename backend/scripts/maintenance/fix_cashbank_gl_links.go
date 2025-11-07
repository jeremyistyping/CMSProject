package main

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=your_password dbname=app_sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	
	// Update this connection string to match your actual database configuration
	// You might need to check your .env file or database configuration
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Fix cash bank GL account links
	if err := fixCashBankGLLinks(db); err != nil {
		log.Fatal("Failed to fix cash bank GL links:", err)
	}

	fmt.Println("Cash bank GL account links have been fixed successfully!")
}

func fixCashBankGLLinks(db *gorm.DB) error {
	// Get all cash banks that don't have GL account links
	var cashBanks []models.CashBank
	if err := db.Preload("Account").Find(&cashBanks).Error; err != nil {
		return fmt.Errorf("failed to fetch cash banks: %w", err)
	}

	fmt.Printf("Found %d cash/bank accounts\n", len(cashBanks))

	// Find accounts that need fixing
	var accountsToFix []models.CashBank
	for _, cb := range cashBanks {
		// Check if account_id is missing or points to deleted/non-existent account
		if cb.AccountID == 0 || cb.Account.ID == 0 {
			accountsToFix = append(accountsToFix, cb)
		}
	}

	fmt.Printf("Found %d accounts that need GL account linking\n", len(accountsToFix))

	if len(accountsToFix) == 0 {
		fmt.Println("All cash/bank accounts already have proper GL account links!")
		return nil
	}

	// Start transaction
	tx := db.Begin()

	for _, cb := range accountsToFix {
		fmt.Printf("Fixing account: %s (%s)\n", cb.Name, cb.Type)

		// Generate GL account code
		var code string
		if cb.Type == "CASH" {
			code = fmt.Sprintf("1100-%03d", time.Now().Unix()%1000)
		} else {
			code = fmt.Sprintf("1110-%03d", time.Now().Unix()%1000)
		}

		// Check if GL account with this name already exists
		var existingAccount models.Account
		err := tx.Where("name = ? AND type = 'ASSET' AND category = 'CURRENT_ASSET'", cb.Name).First(&existingAccount).Error
		
		var glAccountID uint
		if err == gorm.ErrRecordNotFound {
			// Create new GL account
			glAccount := models.Account{
				Code:        code,
				Name:        cb.Name,
				Type:        "ASSET",
				Category:    "CURRENT_ASSET",
				Level:       3,
				IsHeader:    false,
				IsActive:    true,
				Description: fmt.Sprintf("Auto-created GL account for %s account: %s", cb.Type, cb.Name),
			}

			// Find parent account (1100 - Current Assets)
			var parentAccount models.Account
			if err := tx.Where("code = ?", "1100").First(&parentAccount).Error; err == nil {
				glAccount.ParentID = &parentAccount.ID
			}

			if err := tx.Create(&glAccount).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create GL account for %s: %w", cb.Name, err)
			}
			glAccountID = glAccount.ID
			fmt.Printf("  Created new GL account: %s - %s\n", code, cb.Name)
		} else if err != nil {
			tx.Rollback()
			return fmt.Errorf("error checking for existing GL account: %w", err)
		} else {
			glAccountID = existingAccount.ID
			fmt.Printf("  Using existing GL account: %s - %s\n", existingAccount.Code, existingAccount.Name)
		}

		// Update cash bank account to link to GL account
		if err := tx.Model(&cb).Update("account_id", glAccountID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update cash bank account %s: %w", cb.Name, err)
		}

		fmt.Printf("  Linked cash/bank account '%s' to GL account ID %d\n", cb.Name, glAccountID)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Verify results
	fmt.Println("\nVerification:")
	var verifyResults []struct {
		CashBankID   uint   `gorm:"column:id"`
		CashBankCode string `gorm:"column:code"`
		CashBankName string `gorm:"column:name"`
		Type         string `gorm:"column:type"`
		AccountID    uint   `gorm:"column:account_id"`
		AccountCode  string `gorm:"column:account_code"`
		AccountName  string `gorm:"column:account_name"`
	}

	query := `
		SELECT 
			cb.id,
			cb.code,
			cb.name,
			cb.type,
			cb.account_id,
			a.code as account_code,
			a.name as account_name
		FROM cash_banks cb 
		LEFT JOIN accounts a ON cb.account_id = a.id 
		ORDER BY cb.id
	`

	if err := db.Raw(query).Scan(&verifyResults).Error; err != nil {
		return fmt.Errorf("failed to verify results: %w", err)
	}

	fmt.Printf("%-5s %-15s %-30s %-6s %-12s %-15s %-30s\n", 
		"ID", "Code", "Name", "Type", "GL-ID", "GL-Code", "GL-Name")
	fmt.Println(strings.Repeat("-", 120))

	for _, result := range verifyResults {
		glStatus := "✅ LINKED"
		if result.AccountID == 0 {
			glStatus = "❌ MISSING"
		}

		fmt.Printf("%-5d %-15s %-30s %-6s %-12d %-15s %-30s %s\n", 
			result.CashBankID, result.CashBankCode, result.CashBankName, 
			result.Type, result.AccountID, result.AccountCode, result.AccountName,
			glStatus)
	}

	return nil
}
