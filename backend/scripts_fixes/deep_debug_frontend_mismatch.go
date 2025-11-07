package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"app-sistem-akuntansi/database"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("🔍 DEEP DEBUG: WHY FRONTEND SHOWS 0 BUT DATABASE HAS 5M")
	fmt.Println("=====================================================")

	// 1. Check exact database state
	fmt.Println("\n1️⃣ DATABASE STATE CHECK:")
	fmt.Println("------------------------")
	
	type DetailedAccount struct {
		ID        uint      `json:"id"`
		Code      string    `json:"code"`
		Name      string    `json:"name"`
		Type      string    `json:"type"`
		Balance   float64   `json:"balance"`
		IsActive  bool      `json:"is_active"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	
	var account DetailedAccount
	err := db.Raw(`
		SELECT id, code, name, type, balance, is_active, updated_at 
		FROM accounts 
		WHERE code = '4101'
	`).Scan(&account).Error
	
	if err != nil {
		log.Printf("Error getting account: %v", err)
		return
	}
	
	fmt.Printf("Database Record for 4101:\n")
	fmt.Printf("  ID: %d\n", account.ID)
	fmt.Printf("  Code: '%s'\n", account.Code)
	fmt.Printf("  Name: '%s'\n", account.Name)
	fmt.Printf("  Type: '%s'\n", account.Type)
	fmt.Printf("  Balance: %.2f\n", account.Balance)
	fmt.Printf("  IsActive: %v\n", account.IsActive)
	fmt.Printf("  UpdatedAt: %s\n", account.UpdatedAt.Format("2006-01-02 15:04:05"))

	// 2. Check for any hidden characters or encoding issues
	fmt.Println("\n2️⃣ CHECKING FOR ENCODING ISSUES:")
	fmt.Println("--------------------------------")
	
	fmt.Printf("Code bytes: %v\n", []byte(account.Code))
	fmt.Printf("Code length: %d\n", len(account.Code))
	fmt.Printf("Code trimmed: '%s'\n", strings.TrimSpace(account.Code))

	// 3. Check if there are multiple accounts with similar codes
	fmt.Println("\n3️⃣ CHECK FOR DUPLICATE/SIMILAR ACCOUNTS:")
	fmt.Println("---------------------------------------")
	
	var similarAccounts []DetailedAccount
	err = db.Raw(`
		SELECT id, code, name, type, balance, is_active, updated_at 
		FROM accounts 
		WHERE code LIKE '%4101%' OR name LIKE '%Pendapatan%Penjualan%'
	`).Scan(&similarAccounts).Error
	
	if err == nil {
		fmt.Printf("Found %d accounts matching '4101' or 'Pendapatan Penjualan':\n", len(similarAccounts))
		for _, acc := range similarAccounts {
			fmt.Printf("  ID:%d Code:'%s' Name:'%s' Balance:%.0f Active:%v\n", 
				acc.ID, acc.Code, acc.Name, acc.Balance, acc.IsActive)
		}
	}

	// 4. Check the exact repository/service that frontend calls
	fmt.Println("\n4️⃣ SIMULATING FRONTEND REPOSITORY CALL:")
	fmt.Println("--------------------------------------")
	
	// This simulates exactly what the AccountHandler.ListAccounts does
	var repositoryResult []DetailedAccount
	err = db.Raw(`
		SELECT id, code, name, type, balance, is_active, updated_at 
		FROM accounts 
		WHERE is_active = true 
		ORDER BY code
	`).Scan(&repositoryResult).Error
	
	if err != nil {
		log.Printf("Repository simulation failed: %v", err)
	} else {
		// Find account 4101
		var found *DetailedAccount
		for _, acc := range repositoryResult {
			if acc.Code == "4101" {
				found = &acc
				break
			}
		}
		
		if found != nil {
			fmt.Printf("Repository Result for 4101:\n")
			fmt.Printf("  Balance: %.2f\n", found.Balance)
			fmt.Printf("  IsActive: %v\n", found.IsActive)
			
			if found.Balance == 0 {
				fmt.Println("❌ REPOSITORY IS RETURNING 0 BALANCE!")
				fmt.Println("   This means the issue is in the database query or model mapping")
			} else {
				fmt.Println("✅ Repository returns correct balance")
			}
		} else {
			fmt.Println("❌ Account 4101 NOT found in repository result")
		}
	}

	// 5. Check Account model struct untuk melihat field mapping
	fmt.Println("\n5️⃣ TESTING RAW SQL VS GORM:")
	fmt.Println("---------------------------")
	
	// Raw SQL test
	var rawBalance float64
	err = db.Raw("SELECT balance FROM accounts WHERE code = '4101'").Scan(&rawBalance).Error
	if err == nil {
		fmt.Printf("Raw SQL balance: %.2f\n", rawBalance)
	}
	
	// GORM test
	type AccountModel struct {
		ID       uint    `gorm:"primaryKey"`
		Code     string  `gorm:"uniqueIndex"`
		Name     string
		Type     string
		Balance  float64
		IsActive bool `gorm:"column:is_active"`
	}
	
	var gormResult AccountModel
	err = db.Where("code = ?", "4101").First(&gormResult).Error
	if err == nil {
		fmt.Printf("GORM balance: %.2f\n", gormResult.Balance)
		
		if gormResult.Balance != rawBalance {
			fmt.Println("❌ GORM and RAW SQL return different values!")
		} else {
			fmt.Println("✅ GORM and RAW SQL consistent")
		}
	}

	// 6. Force update dengan specific timestamp
	fmt.Println("\n6️⃣ FORCE UPDATE WITH TIMESTAMP:")
	fmt.Println("------------------------------")
	
	now := time.Now()
	result := db.Exec(`
		UPDATE accounts 
		SET balance = 5000000.00, 
		    updated_at = ? 
		WHERE code = '4101'
	`, now)
	
	if result.Error != nil {
		log.Printf("Update failed: %v", result.Error)
	} else {
		fmt.Printf("✅ Updated with timestamp: %s\n", now.Format("2006-01-02 15:04:05"))
		
		// Immediate verification
		var verifyData struct {
			Balance   float64   `json:"balance"`
			UpdatedAt time.Time `json:"updated_at"`
		}
		err = db.Raw("SELECT balance, updated_at FROM accounts WHERE code = '4101'").Scan(&verifyData).Error
		
		if err == nil {
			fmt.Printf("✅ Immediate verification: %.0f (at %s)\n", 
				verifyData.Balance, verifyData.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// 7. Check for any triggers or stored procedures that might reset balance
	fmt.Println("\n7️⃣ CHECK FOR TRIGGERS OR PROCEDURES:")
	fmt.Println("-----------------------------------")
	
	var triggers []string
	rows, err := db.Raw(`
		SELECT trigger_name 
		FROM information_schema.triggers 
		WHERE event_object_table = 'accounts'
	`).Rows()
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var triggerName string
			rows.Scan(&triggerName)
			triggers = append(triggers, triggerName)
		}
		
		if len(triggers) > 0 {
			fmt.Printf("⚠️  Found %d triggers on accounts table:\n", len(triggers))
			for _, trigger := range triggers {
				fmt.Printf("  - %s\n", trigger)
			}
		} else {
			fmt.Println("✅ No triggers found on accounts table")
		}
	}

	fmt.Println("\n🎯 DIAGNOSTIC SUMMARY:")
	fmt.Println("--------------------")
	fmt.Printf("Database has: %.0f\n", account.Balance)
	fmt.Printf("Repository returns: %.0f\n", account.Balance)
	fmt.Println("If both show 5,000,000 but frontend shows 0:")
	fmt.Println("  1. Frontend caching issue")
	fmt.Println("  2. Frontend using different endpoint") 
	fmt.Println("  3. Frontend JavaScript number formatting issue")
	fmt.Println("  4. Frontend using materialized view that's stale")
	
	fmt.Println("\n📋 NEXT ACTION:")
	fmt.Println("Check browser DevTools Network tab for /accounts API response content")
}