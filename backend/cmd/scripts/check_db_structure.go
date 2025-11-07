package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Println("🔍 Checking database structure...")

	// Initialize database connection
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("❌ Failed to connect to database")
	}

	// Check if tables exist by trying to query them
	tables := map[string]interface{}{
		"accounts":    &models.Account{},
		"coa":         &models.COA{},
		"cash_banks":  &models.CashBank{},
		"purchases":   &models.Purchase{},
	}

	fmt.Println("\n📊 DATABASE TABLES CHECK")
	fmt.Println("========================")

	for tableName, model := range tables {
		var count int64
		err := db.Model(model).Count(&count).Error
		if err != nil {
			fmt.Printf("❌ Table '%s': ERROR - %v\n", tableName, err)
		} else {
			fmt.Printf("✅ Table '%s': EXISTS - %d records\n", tableName, count)
		}
	}

	// Check specific account structure
	fmt.Println("\n📋 ACCOUNTS TABLE STRUCTURE")
	fmt.Println("============================")
	
	var sampleAccount models.Account
	if err := db.First(&sampleAccount).Error; err == nil {
		fmt.Printf("Sample Account ID: %d\n", sampleAccount.ID)
		fmt.Printf("Sample Account Code: %s\n", sampleAccount.Code)
		fmt.Printf("Sample Account Name: %s\n", sampleAccount.Name)
		fmt.Printf("Sample Account Type: %s\n", sampleAccount.Type)
	} else {
		fmt.Printf("❌ Cannot get sample account: %v\n", err)
	}

	// Check COA structure if exists
	fmt.Println("\n📋 COA TABLE STRUCTURE")
	fmt.Println("=======================")
	
	var sampleCOA models.COA
	if err := db.First(&sampleCOA).Error; err == nil {
		fmt.Printf("Sample COA ID: %d\n", sampleCOA.ID)
		fmt.Printf("Sample COA Code: %s\n", sampleCOA.Code)
		fmt.Printf("Sample COA Name: %s\n", sampleCOA.Name)
		fmt.Printf("Sample COA Type: %s\n", sampleCOA.Type)
		fmt.Printf("Sample COA Balance: %.2f\n", sampleCOA.Balance)
	} else {
		fmt.Printf("❌ Cannot get sample COA: %v\n", err)
	}

	// Check Cash Banks structure
	fmt.Println("\n💰 CASH_BANKS TABLE STRUCTURE")
	fmt.Println("==============================")
	
	var sampleCashBank models.CashBank
	if err := db.First(&sampleCashBank).Error; err == nil {
		fmt.Printf("Sample Cash Bank ID: %d\n", sampleCashBank.ID)
		fmt.Printf("Sample Cash Bank Name: %s\n", sampleCashBank.Name)
		fmt.Printf("Sample Cash Bank AccountID: %d\n", sampleCashBank.AccountID)
		fmt.Printf("Sample Cash Bank Balance: %.2f\n", sampleCashBank.Balance)
		fmt.Printf("Sample Cash Bank Active: %t\n", sampleCashBank.IsActive)
	} else {
		fmt.Printf("❌ Cannot get sample cash bank: %v\n", err)
	}

	// Check relationship between cash_banks and accounts/coa
	fmt.Println("\n🔗 RELATIONSHIP CHECK")
	fmt.Println("======================")
	
	query := `
		SELECT 
			cb.id as cash_bank_id,
			cb.name as cash_bank_name,
			cb.account_id,
			cb.balance as cash_bank_balance,
			acc.code as account_code,
			acc.name as account_name,
			acc.type as account_type
		FROM cash_banks cb
		LEFT JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.is_active = true
		LIMIT 5
	`
	
	var relationships []map[string]interface{}
	if err := db.Raw(query).Scan(&relationships).Error; err == nil {
		fmt.Printf("Found %d cash bank -> accounts relationships:\n", len(relationships))
		for _, rel := range relationships {
			fmt.Printf("  - Cash Bank %v (%v) -> Account %v (%v) - Balance: %v\n", 
				rel["cash_bank_id"], rel["cash_bank_name"], 
				rel["account_code"], rel["account_name"], 
				rel["cash_bank_balance"])
		}
	} else {
		fmt.Printf("❌ Error checking relationships: %v\n", err)
	}

	log.Println("✅ Database structure check completed")
}