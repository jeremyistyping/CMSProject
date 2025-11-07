package main

import (
	"fmt"
	"log"
	"strings"
	"app-sistem-akuntansi/database"
)

type BalanceSync struct {
	CashBankID      uint    `json:"cash_bank_id"`
	CashBankCode    string  `json:"cash_bank_code"`
	CashBankName    string  `json:"cash_bank_name"`
	CashBankType    string  `json:"cash_bank_type"`
	CashBankBalance float64 `json:"cash_bank_balance"`
	GLAccountID     uint    `json:"gl_account_id"`
	GLCode          string  `json:"gl_code"`
	GLName          string  `json:"gl_name"`
	GLBalance       float64 `json:"gl_balance"`
	IsSync          bool    `json:"is_sync"`
	Difference      float64 `json:"difference"`
}

func main() {
	// Connect to database
	db := database.ConnectDB()
	
	fmt.Println("🔍 Checking balance synchronization between Cash/Bank accounts and GL accounts...")
	fmt.Println(strings.Repeat("=", 80))
	
	var results []BalanceSync
	
	// Query to join cash_bank_accounts with accounts (GL accounts)
	err := db.Raw(`
		SELECT 
			cb.id as cash_bank_id,
			cb.code as cash_bank_code,
			cb.name as cash_bank_name,
			cb.type as cash_bank_type,
			cb.balance as cash_bank_balance,
			cb.account_id as gl_account_id,
			COALESCE(acc.code, 'UNLINKED') as gl_code,
			COALESCE(acc.name, 'NO GL ACCOUNT') as gl_name,
			COALESCE(acc.balance, 0) as gl_balance,
			CASE WHEN cb.balance = COALESCE(acc.balance, 0) THEN 1 ELSE 0 END as is_sync,
			cb.balance - COALESCE(acc.balance, 0) as difference
		FROM cash_banks cb
		LEFT JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL
		ORDER BY cb.type, cb.code
	`).Scan(&results).Error
	
	if err != nil {
		log.Fatal("Failed to query balance sync:", err)
	}
	
	fmt.Printf("%-15s %-20s %-10s %-15s %-15s %-15s %-15s %-10s %-15s\n", 
		"CB_CODE", "CB_NAME", "TYPE", "CB_BALANCE", "GL_CODE", "GL_NAME", "GL_BALANCE", "SYNC", "DIFFERENCE")
	fmt.Println(strings.Repeat("=", 150))
	
	syncCount := 0
	unsyncCount := 0
	
	for _, result := range results {
		syncStatus := "❌ NO"
		if result.IsSync {
			syncStatus = "✅ YES"
			syncCount++
		} else {
			unsyncCount++
		}
		
		fmt.Printf("%-15s %-20s %-10s %-15.2f %-15s %-15s %-15.2f %-10s %-15.2f\n",
			result.CashBankCode,
			truncateString(result.CashBankName, 20),
			result.CashBankType,
			result.CashBankBalance,
			result.GLCode,
			truncateString(result.GLName, 15),
			result.GLBalance,
			syncStatus,
			result.Difference,
		)
	}
	
	fmt.Println(strings.Repeat("=", 150))
	fmt.Printf("Summary:\n")
	fmt.Printf("✅ Synchronized accounts: %d\n", syncCount)
	fmt.Printf("❌ Unsynchronized accounts: %d\n", unsyncCount)
	fmt.Printf("📊 Total accounts: %d\n", len(results))
	
	if unsyncCount > 0 {
		fmt.Println("\n🚨 Issues found:")
		fmt.Println("Some cash/bank accounts have different balances than their linked GL accounts.")
		fmt.Println("This means transactions are not being properly synchronized between the two systems.")
		
		fmt.Println("\n💡 Possible causes:")
		fmt.Println("1. Manual balance updates in cash_bank_accounts without updating GL accounts")
		fmt.Println("2. Missing journal entries when cash/bank transactions are processed")
		fmt.Println("3. GL account balance calculation issues")
		fmt.Println("4. Data inconsistency from previous operations")
		
		fmt.Println("\n🔧 Recommended actions:")
		fmt.Println("1. Implement balance synchronization in transaction processing")
		fmt.Println("2. Create journal entries for all cash/bank transactions")
		fmt.Println("3. Add validation to ensure balances stay in sync")
		fmt.Println("4. Run balance sync repair script")
	}
}

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
