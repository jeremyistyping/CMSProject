package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	config.LoadConfig()
	db := database.ConnectDB()
	
	fmt.Println("=== FINDING SOFT-DELETED ACCOUNTS WITH BALANCE ===\n")
	
	var accounts []struct {
		ID        uint
		Code      string
		Name      string
		Type      string
		Balance   float64
		DeletedAt string
	}
	
	err := db.Raw(`
		SELECT id, code, name, type, balance, deleted_at::text
		FROM accounts 
		WHERE deleted_at IS NOT NULL
		  AND balance != 0
		ORDER BY ABS(balance) DESC
	`).Scan(&accounts).Error
	
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	
	if len(accounts) == 0 {
		fmt.Println("✅ No soft-deleted accounts with non-zero balance found")
		fmt.Println("\nThis means the Rp 165.000 difference must be from another issue.")
		fmt.Println("\n📋 INVESTIGATION SUGGESTIONS:")
		fmt.Println("1. Check if there are duplicate accounts with same code")
		fmt.Println("2. Check if there are accounts with wrong type classification")
		fmt.Println("3. Check PPN netting calculation")
		fmt.Println("4. Verify Retained Earnings (3201) inclusion")
		return
	}
	
	fmt.Printf("Found %d soft-deleted accounts with balance:\n\n", len(accounts))
	
	var totalAssets, totalLiabilities, totalEquity float64
	
	for i, acc := range accounts {
		fmt.Printf("%d. Account %s - %s\n", i+1, acc.Code, acc.Name)
		fmt.Printf("   Type: %s\n", acc.Type)
		fmt.Printf("   Balance: Rp %.0f\n", acc.Balance)
		fmt.Printf("   Deleted At: %s\n\n", acc.DeletedAt)
		
		switch acc.Type {
		case "ASSET":
			totalAssets += acc.Balance
		case "LIABILITY":
			totalLiabilities += acc.Balance
		case "EQUITY":
			totalEquity += acc.Balance
		}
	}
	
	fmt.Println("=== SUMMARY ===")
	fmt.Printf("Total deleted ASSETS with balance: Rp %.0f\n", totalAssets)
	fmt.Printf("Total deleted LIABILITIES with balance: Rp %.0f\n", totalLiabilities)
	fmt.Printf("Total deleted EQUITY with balance: Rp %.0f\n", totalEquity)
	fmt.Printf("\nNet impact on balance sheet: Rp %.0f\n", totalAssets-(totalLiabilities+totalEquity))
	
	fmt.Println("\n📋 RECOMMENDATION:")
	fmt.Println("These deleted accounts are being excluded by the updated balance sheet query.")
	fmt.Println("If the balance sheet still shows a difference after restart, the issue is elsewhere.")
	
	// Also check for accounts with inconsistent data
	fmt.Println("\n=== CHECKING FOR OTHER ANOMALIES ===\n")
	
	var duplicates []struct {
		Code  string
		Count int64
		Type  string
	}
	
	db.Raw(`
		SELECT code, COUNT(*) as count, MAX(type) as type
		FROM accounts 
		WHERE deleted_at IS NULL
		  AND is_active = true
		GROUP BY code
		HAVING COUNT(*) > 1
		ORDER BY count DESC
	`).Scan(&duplicates)
	
	if len(duplicates) > 0 {
		fmt.Printf("⚠️  Found %d duplicate account codes:\n\n", len(duplicates))
		for _, dup := range duplicates {
			fmt.Printf("  Code %s (%s): %d instances\n", dup.Code, dup.Type, dup.Count)
		}
	} else {
		fmt.Println("✅ No duplicate account codes found")
	}
}
