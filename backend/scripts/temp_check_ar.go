package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Account struct {
	ID       uint   `gorm:"primaryKey"`
	Code     string
	Name     string  
	Type     string
	Balance  float64
}

func main() {
	// Connect to database
	dsn := "accounting_user:Bismillah2024!@tcp(localhost:3306)/accounting_system?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("Database connection failed: %v\n", err)
		return
	}

	fmt.Println("🔍 CHECKING ACCOUNTS RECEIVABLE ACCOUNT:")
	fmt.Println("===========================================")
	
	// Check for AR account by code 1201
	var arAccount Account
	if err := db.Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		fmt.Printf("❌ Account 1201 (AR) not found: %v\n", err)
		
		// Try to find any 120x account
		var accounts []Account
		if err := db.Where("code LIKE ?", "120%").Find(&accounts).Error; err == nil {
			fmt.Printf("\n🔍 Found 120x accounts:\n")
			for _, acc := range accounts {
				fmt.Printf("   %s - %s (Type: %s, Balance: %.2f)\n", acc.Code, acc.Name, acc.Type, acc.Balance)
			}
		}
	} else {
		fmt.Printf("✅ Found AR Account: %s - %s (Type: %s, Balance: %.2f)\n", arAccount.Code, arAccount.Name, arAccount.Type, arAccount.Balance)
	}
	
	fmt.Println("\n🔍 CHECKING ALL ASSET ACCOUNTS:")
	fmt.Println("=====================================")
	
	var assetAccounts []Account
	if err := db.Where("type = ?", "ASSET").Or("code LIKE ?", "1%").Find(&assetAccounts).Error; err == nil {
		for _, acc := range assetAccounts {
			if acc.Balance != 0 {
				fmt.Printf("   %s - %s (Balance: %.2f)\n", acc.Code, acc.Name, acc.Balance)
			}
		}
	}
}