package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
)

type AccountInfo struct {
	ID      uint
	Code    string
	Name    string
	Balance float64
}

func main() {
	log.Println("Finding and fixing duplicate accounts...")
	
	db := database.ConnectDB()
	
	// Find all duplicate account codes
	var duplicateCodes []string
	db.Raw(`
		SELECT code
		FROM accounts
		WHERE deleted_at IS NULL
		GROUP BY code
		HAVING COUNT(*) > 1
		ORDER BY code
	`).Scan(&duplicateCodes)
	
	if len(duplicateCodes) == 0 {
		log.Println("✅ No duplicate accounts found")
		return
	}
	
	fmt.Printf("\n=== FOUND %d DUPLICATE ACCOUNT CODES ===\n\n", len(duplicateCodes))
	
	for _, code := range duplicateCodes {
		var accounts []AccountInfo
		db.Raw(`
			SELECT id, code, name, balance
			FROM accounts
			WHERE code = ? AND deleted_at IS NULL
			ORDER BY id
		`, code).Scan(&accounts)
		
		fmt.Printf("Code: %s (%d duplicates)\n", code, len(accounts))
		
		var keepAccount *AccountInfo
		var deleteAccounts []AccountInfo
		totalBalance := 0.0
		
		for i, acc := range accounts {
			fmt.Printf("  %d. ID=%d, Name=%s, Balance=%.2f\n", i+1, acc.ID, acc.Name, acc.Balance)
			totalBalance += acc.Balance
			
			// Keep the account with non-zero balance, or the first one if all are zero
			if acc.Balance != 0 || keepAccount == nil {
				if keepAccount != nil && keepAccount.Balance == 0 {
					// Previous keep account had 0 balance, switch to this one
					deleteAccounts = append(deleteAccounts, *keepAccount)
				}
				keepAccount = &acc
			} else {
				deleteAccounts = append(deleteAccounts, acc)
			}
		}
		
		fmt.Printf("  Total Balance: %.2f\n", totalBalance)
		fmt.Printf("  Will KEEP: ID=%d (%s) with balance %.2f\n", keepAccount.ID, keepAccount.Name, keepAccount.Balance)
		
		// Update the keep account with total balance if needed
		if keepAccount.Balance != totalBalance && totalBalance != 0 {
			fmt.Printf("  Updating balance from %.2f to %.2f\n", keepAccount.Balance, totalBalance)
			db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", totalBalance, keepAccount.ID)
		}
		
		// Soft delete the duplicate accounts
		for _, acc := range deleteAccounts {
			fmt.Printf("  Soft deleting: ID=%d (%s)\n", acc.ID, acc.Name)
			db.Exec("UPDATE accounts SET deleted_at = NOW() WHERE id = ?", acc.ID)
		}
		
		fmt.Println()
	}
	
	fmt.Println("✅ Duplicate accounts fixed!")
	fmt.Println("\nPlease verify account balances in Balance Sheet")
}
