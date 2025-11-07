package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	fmt.Println("📋 Checking current PPN accounts:")
	var accounts []models.Account
	db.Where("code LIKE '210%' OR name ILIKE '%ppn%'").Find(&accounts)
	
	if len(accounts) == 0 {
		fmt.Println("❌ No PPN accounts found!")
		return
	}
	
	for _, acc := range accounts {
		fmt.Printf("- %s: %s (%s)\n", acc.Code, acc.Name, acc.Type)
	}
	
	fmt.Println("\n📋 Checking for account 2102 and 2103:")
	var account2102 models.Account
	if err := db.Where("code = ?", "2102").First(&account2102).Error; err == nil {
		fmt.Printf("✅ Account 2102: %s - %s (%s)\n", account2102.Code, account2102.Name, account2102.Type)
	} else {
		fmt.Println("❌ Account 2102 not found")
	}
	
	var account2103 models.Account
	if err := db.Where("code = ?", "2103").First(&account2103).Error; err == nil {
		fmt.Printf("✅ Account 2103: %s - %s (%s)\n", account2103.Code, account2103.Name, account2103.Type)
	} else {
		fmt.Println("❌ Account 2103 not found - need to create it!")
	}
}