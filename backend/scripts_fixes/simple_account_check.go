package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("🔍 SIMPLE CHECK: ACCOUNT 4101 STATUS")
	fmt.Println("=====================================")

	// 1. Check table structure first
	fmt.Println("\n1️⃣ CHECKING TABLE COLUMNS:")
	fmt.Println("---------------------------")
	
	rows, err := db.Raw("SELECT column_name FROM information_schema.columns WHERE table_name = 'accounts' ORDER BY ordinal_position").Rows()
	if err != nil {
		log.Printf("Error getting columns: %v", err)
		return
	}
	defer rows.Close()
	
	fmt.Println("Available columns in accounts table:")
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			log.Printf("Error scanning column: %v", err)
			continue
		}
		fmt.Printf("  - %s\n", columnName)
	}

	// 2. Simple check without status column
	fmt.Println("\n2️⃣ DIRECT ACCOUNT 4101 CHECK:")
	fmt.Println("-----------------------------")
	
	type SimpleAccount struct {
		ID      uint    `json:"id"`
		Code    string  `json:"code"`
		Name    string  `json:"name"`
		Type    string  `json:"type"`
		Balance float64 `json:"balance"`
	}
	
	var account4101 SimpleAccount
	err = db.Raw("SELECT id, code, name, type, balance FROM accounts WHERE code = '4101'").Scan(&account4101).Error
	if err != nil {
		log.Printf("Error getting account 4101: %v", err)
		return
	}
	
	fmt.Printf("Account 4101 Details:\n")
	fmt.Printf("  ID: %d\n", account4101.ID)
	fmt.Printf("  Code: %s\n", account4101.Code)
	fmt.Printf("  Name: %s\n", account4101.Name)
	fmt.Printf("  Type: %s\n", account4101.Type)
	fmt.Printf("  Balance: %.2f\n", account4101.Balance)
	
	if account4101.Balance == 0 {
		fmt.Println("❌ BALANCE IS 0 - MASALAH DITEMUKAN!")
	} else {
		fmt.Println("✅ Balance sudah benar")
	}

	// 3. Update balance to 5M
	fmt.Println("\n3️⃣ UPDATING BALANCE TO 5,000,000:")
	fmt.Println("---------------------------------")
	
	result := db.Exec("UPDATE accounts SET balance = 5000000.00 WHERE code = '4101'")
	if result.Error != nil {
		log.Printf("❌ Error updating balance: %v", result.Error)
	} else {
		fmt.Printf("✅ Updated %d rows\n", result.RowsAffected)
		
		// Verify update
		var newBalance float64
		err = db.Raw("SELECT balance FROM accounts WHERE code = '4101'").Scan(&newBalance).Error
		if err != nil {
			log.Printf("Error verifying update: %v", err)
		} else {
			fmt.Printf("✅ Verified new balance: %.2f\n", newBalance)
		}
	}

	// 4. Check all revenue accounts
	fmt.Println("\n4️⃣ ALL REVENUE ACCOUNTS:")
	fmt.Println("-----------------------")
	
	var revenueAccounts []SimpleAccount
	err = db.Raw("SELECT id, code, name, type, balance FROM accounts WHERE type = 'REVENUE' ORDER BY code").Scan(&revenueAccounts).Error
	if err != nil {
		log.Printf("Error getting revenue accounts: %v", err)
	} else {
		fmt.Printf("Revenue accounts in database:\n")
		for _, acc := range revenueAccounts {
			fmt.Printf("  %s (%s): Balance = %.0f\n", acc.Code, acc.Name, acc.Balance)
		}
	}

	fmt.Println("\n🎯 NEXT STEPS:")
	fmt.Println("1. Hard refresh browser (Ctrl+F5)")
	fmt.Println("2. Check Network tab in browser DevTools")
	fmt.Println("3. Look for /accounts API call response")
	fmt.Println("4. If still 0, there might be frontend caching issue")
}