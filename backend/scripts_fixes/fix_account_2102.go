package main

import (
	"fmt"
	"log"
	
	"app-sistem-akuntansi/database"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize database
	db := database.ConnectDB()

	fmt.Println("=== FIXING ACCOUNT 2102 CONFIGURATION ===\n")

	// 1. Check current state of account 2102
	type Account struct {
		ID      uint
		Code    string
		Name    string
		Type    string
		Balance float64
	}
	
	var account Account
	err := db.Raw(`SELECT id, code, name, type, balance FROM accounts WHERE code = '2102'`).Scan(&account).Error
	
	if err != nil {
		fmt.Printf("❌ Error querying account 2102: %v\n", err)
		return
	}
	
	if account.ID == 0 {
		fmt.Println("❌ Account 2102 NOT FOUND in database!")
		return
	}
	
	fmt.Printf("Current state of Account 2102:\n")
	fmt.Printf("   ID: %d\n", account.ID)
	fmt.Printf("   Code: %s\n", account.Code)
	fmt.Printf("   Name: %s\n", account.Name)
	fmt.Printf("   Type: %s\n", account.Type)
	fmt.Printf("   Balance: Rp %.2f\n\n", account.Balance)

	// 2. Determine the issue
	fmt.Println("=== ISSUE ANALYSIS ===")
	if account.Code[0] == '2' && account.Type != "LIABILITY" {
		fmt.Printf("❌ ISSUE: Account code %s starts with '2' (LIABILITY range) but type is %s\n", account.Code, account.Type)
	}
	if account.Type == "ASSET" && (account.Name == "PPN Masukan" || account.Name == "PPN MASUKAN") {
		fmt.Println("❌ ISSUE: Account is marked as ASSET but has code in LIABILITY range (2xxx)")
	}
	
	fmt.Println("\n=== PROPOSED FIX ===")
	fmt.Println("Account 2102 with balance Rp 165.000 needs to be recategorized.")
	fmt.Println("\nOption 1: Change to LIABILITY (if it's actually a liability)")
	fmt.Println("   - Update type to 'LIABILITY'")
	fmt.Println("   - Rename to appropriate liability name (e.g., 'UTANG PPN' or 'PPN YANG HARUS DIBAYAR')")
	fmt.Println("\nOption 2: Change code to Asset range (if it's actually PPN Masukan)")
	fmt.Println("   - Update code to something like '1106' (PPN Masukan asset account)")
	fmt.Println("   - Keep type as 'ASSET' and name as 'PPN MASUKAN'")
	
	fmt.Println("\n=== RECOMMENDED ACTION ===")
	fmt.Println("Based on the code '2102' (LIABILITY range), the most likely fix is:")
	fmt.Println("   ✅ UPDATE account SET type = 'LIABILITY' WHERE code = '2102'")
	fmt.Println("   ✅ UPDATE account SET name = 'UTANG PPN LAIN-LAIN' WHERE code = '2102'")
	
	fmt.Println("\nOR, if this is truly a PPN Masukan (Input VAT) asset:")
	fmt.Println("   ✅ Move the balance to proper PPN Masukan account (1240)")
	fmt.Println("   ✅ Delete or deactivate account 2102")
	
	fmt.Println("\n⚠️  IMPORTANT: This script does NOT automatically fix the issue.")
	fmt.Println("   Please review the analysis and run SQL commands manually:")
	fmt.Println("\n   -- Option 1: Fix as LIABILITY")
	fmt.Println("   UPDATE accounts SET type = 'LIABILITY', name = 'UTANG PPN LAIN-LAIN' WHERE code = '2102';")
	fmt.Println("\n   -- Option 2: Merge into proper PPN Masukan account (1240)")
	fmt.Println("   -- First, transfer balance via journal entry")
	fmt.Println("   -- Then deactivate account 2102")
	fmt.Println("   UPDATE accounts SET is_active = false WHERE code = '2102';")
}
