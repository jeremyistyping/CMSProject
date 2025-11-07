package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	config.LoadConfig()
	db := database.ConnectDB()
	
	fmt.Println("=== CLEARING ACCOUNT 2102 BALANCE ===\n")
	
	// 1. Check current status
	var account struct {
		ID        uint
		Code      string
		Name      string
		Type      string
		Balance   float64
		DeletedAt *string
	}
	
	err := db.Raw(`
		SELECT id, code, name, type, balance, deleted_at
		FROM accounts 
		WHERE code = '2102'
	`).Scan(&account).Error
	
	if err != nil || account.ID == 0 {
		fmt.Println("❌ Account 2102 not found")
		return
	}
	
	fmt.Printf("Current status:\n")
	fmt.Printf("  ID: %d\n", account.ID)
	fmt.Printf("  Code: %s\n", account.Code)
	fmt.Printf("  Name: %s\n", account.Name)
	fmt.Printf("  Type: %s\n", account.Type)
	fmt.Printf("  Balance: Rp %.0f\n", account.Balance)
	if account.DeletedAt != nil {
		fmt.Printf("  Status: SOFT DELETED (%s)\n\n", *account.DeletedAt)
	} else {
		fmt.Printf("  Status: ACTIVE\n\n")
	}
	
	if account.Balance == 0 {
		fmt.Println("✅ Balance already 0, no action needed")
		return
	}
	
	// 2. Clear the balance
	fmt.Printf("Clearing balance of Rp %.0f...\n", account.Balance)
	
	result := db.Exec(`
		UPDATE accounts 
		SET balance = 0,
		    updated_at = NOW()
		WHERE code = '2102'
	`)
	
	if result.Error != nil {
		fmt.Printf("❌ Failed to clear balance: %v\n", result.Error)
		return
	}
	
	fmt.Printf("✅ Successfully cleared balance of account 2102\n")
	fmt.Printf("   Rows affected: %d\n\n", result.RowsAffected)
	
	// 3. Verify
	var newBalance float64
	db.Raw(`SELECT balance FROM accounts WHERE code = '2102'`).Scan(&newBalance)
	
	fmt.Printf("Verification:\n")
	fmt.Printf("  New balance: Rp %.0f\n", newBalance)
	
	if newBalance == 0 {
		fmt.Println("\n✅ SUCCESS! Account 2102 balance is now 0")
		fmt.Println("   Balance sheet should now be balanced!")
		fmt.Println("\n📋 NEXT STEPS:")
		fmt.Println("   1. Restart backend to load updated balance sheet service")
		fmt.Println("   2. Generate balance sheet again (date: 31/12/2025)")
		fmt.Println("   3. Verify that diff is now 0")
	} else {
		fmt.Printf("\n⚠️  WARNING: Balance is still %.0f\n", newBalance)
	}
}
