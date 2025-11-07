package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Account struct {
	ID       uint   `gorm:"primaryKey"`
	Code     string `gorm:"uniqueIndex;size:20"`
	Name     string `gorm:"size:255"`
	Type     string `gorm:"size:50"`
	Balance  float64
	IsActive bool `gorm:"default:true"`
	Level    int
	IsHeader bool `gorm:"default:false"`
	ParentID *uint
}

type Journal struct {
	ID          uint      `gorm:"primaryKey"`
	Code        string    `gorm:"uniqueIndex;size:50"`
	Date        time.Time `gorm:"index"`
	Description string
	Status      string `gorm:"size:20;default:DRAFT"`
}

type JournalEntry struct {
	ID           uint `gorm:"primaryKey"`
	JournalID    uint
	AccountID    uint
	Description  string
	DebitAmount  float64
	CreditAmount float64
}

func main() {
	// Database connection string - adjust as needed
	// Update this to match your database configuration
	dsn := "root:@tcp(localhost:3306)/app_sistem_akuntansi?charset=utf8mb4&parseTime=True&loc=Local"
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔍 Balance Sheet Diagnostic & Fix")
	fmt.Println("===================================")

	// Check current state
	checkCurrentState(db)
	
	// Create/Update sample data
	fixBalanceSheetData(db)
	
	// Verify fix
	fmt.Println("\n✅ Verification after fix:")
	checkCurrentState(db)
	
	fmt.Println("\n🎉 Balance sheet fix completed!")
	fmt.Println("Try refreshing the balance sheet report in the frontend.")
}

func checkCurrentState(db *gorm.DB) {
	var totalAccounts int64
	db.Model(&Account{}).Count(&totalAccounts)
	fmt.Printf("Total accounts: %d\n", totalAccounts)

	// Count balance sheet accounts with balances
	var accountsWithBalance int64
	db.Model(&Account{}).
		Where("type IN ? AND balance != 0", []string{"ASSET", "LIABILITY", "EQUITY"}).
		Count(&accountsWithBalance)
	fmt.Printf("Balance sheet accounts with balances: %d\n", accountsWithBalance)

	// Show current balances
	var accounts []Account
	db.Where("type IN ? AND balance != 0", []string{"ASSET", "LIABILITY", "EQUITY"}).
		Order("type, code").Find(&accounts)

	if len(accounts) > 0 {
		fmt.Println("\n📊 Current Balance Sheet Accounts:")
		for _, acc := range accounts {
			fmt.Printf("- %s (%s): %s [Balance: %s]\n", 
				acc.Code, acc.Type, acc.Name, formatCurrency(acc.Balance))
		}
	} else {
		fmt.Println("⚠️  No balance sheet accounts with balances found!")
	}
}

func fixBalanceSheetData(db *gorm.DB) {
	fmt.Println("\n🛠️  Creating/Updating Balance Sheet Sample Data...")

	// Sample balance sheet accounts with balances
	sampleAccounts := []Account{
		{
			Code:     "1-1-001",
			Name:     "Kas",
			Type:     "ASSET",
			Balance:  50000000, // 50 million IDR
			IsActive: true,
			Level:    3,
			IsHeader: false,
		},
		{
			Code:     "1-1-002",
			Name:     "Bank",
			Type:     "ASSET",
			Balance:  100000000, // 100 million IDR
			IsActive: true,
			Level:    3,
			IsHeader: false,
		},
		{
			Code:     "1-1-003",
			Name:     "Piutang Dagang",
			Type:     "ASSET",
			Balance:  75000000, // 75 million IDR
			IsActive: true,
			Level:    3,
			IsHeader: false,
		},
		{
			Code:     "1-2-001",
			Name:     "Peralatan Kantor",
			Type:     "ASSET",
			Balance:  25000000, // 25 million IDR
			IsActive: true,
			Level:    3,
			IsHeader: false,
		},
		{
			Code:     "1-2-002",
			Name:     "Kendaraan",
			Type:     "ASSET",
			Balance:  80000000, // 80 million IDR
			IsActive: true,
			Level:    3,
			IsHeader: false,
		},
		{
			Code:     "2-1-001",
			Name:     "Utang Dagang",
			Type:     "LIABILITY",
			Balance:  40000000, // 40 million IDR
			IsActive: true,
			Level:    3,
			IsHeader: false,
		},
		{
			Code:     "2-1-002",
			Name:     "Utang Pajak",
			Type:     "LIABILITY",
			Balance:  15000000, // 15 million IDR
			IsActive: true,
			Level:    3,
			IsHeader: false,
		},
		{
			Code:     "2-2-001",
			Name:     "Utang Bank",
			Type:     "LIABILITY",
			Balance:  100000000, // 100 million IDR
			IsActive: true,
			Level:    3,
			IsHeader: false,
		},
		{
			Code:     "3-1-001",
			Name:     "Modal Disetor",
			Type:     "EQUITY",
			Balance:  150000000, // 150 million IDR
			IsActive: true,
			Level:    3,
			IsHeader: false,
		},
		{
			Code:     "3-2-001",
			Name:     "Laba Ditahan",
			Type:     "EQUITY",
			Balance:  25000000, // 25 million IDR
			IsActive: true,
			Level:    3,
			IsHeader: false,
		},
	}

	createdCount := 0
	updatedCount := 0

	for _, sampleAccount := range sampleAccounts {
		var existingAccount Account
		result := db.Where("code = ?", sampleAccount.Code).First(&existingAccount)
		
		if result.Error != nil {
			// Account doesn't exist, create it
			if err := db.Create(&sampleAccount).Error; err != nil {
				fmt.Printf("❌ Failed to create account %s: %v\n", sampleAccount.Code, err)
			} else {
				fmt.Printf("✅ Created: %s - %s (%s)\n", 
					sampleAccount.Code, sampleAccount.Name, formatCurrency(sampleAccount.Balance))
				createdCount++
			}
		} else {
			// Account exists, update balance and other fields
			updateData := map[string]interface{}{
				"balance":   sampleAccount.Balance,
				"is_active": sampleAccount.IsActive,
				"level":     sampleAccount.Level,
				"is_header": sampleAccount.IsHeader,
			}
			
			if err := db.Model(&existingAccount).Updates(updateData).Error; err != nil {
				fmt.Printf("❌ Failed to update account %s: %v\n", sampleAccount.Code, err)
			} else {
				fmt.Printf("✅ Updated: %s - %s (%s)\n", 
					sampleAccount.Code, sampleAccount.Name, formatCurrency(sampleAccount.Balance))
				updatedCount++
			}
		}
	}

	fmt.Printf("\n📊 Summary: Created %d accounts, Updated %d accounts\n", createdCount, updatedCount)
}

func formatCurrency(amount float64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("Rp %.0f juta", amount/1000000)
	} else if amount >= 1000 {
		return fmt.Sprintf("Rp %.0f ribu", amount/1000)
	} else {
		return fmt.Sprintf("Rp %.0f", amount)
	}
}