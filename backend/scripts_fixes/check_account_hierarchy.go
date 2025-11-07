package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("=== Account Hierarchy Check ===")
	fmt.Println()

	// Check all accounts and their parent-child relationships
	fmt.Println("1. Account Hierarchy Analysis:")
	type AccountDetail struct {
		ID       uint    `json:"id"`
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		ParentID *uint   `json:"parent_id"`
		IsHeader bool    `json:"is_header"`
		Balance  float64 `json:"balance"`
		Type     string  `json:"type"`
	}

	var accounts []AccountDetail
	err := db.Raw(`
		SELECT id, code, name, parent_id, is_header, balance, type
		FROM accounts 
		WHERE deleted_at IS NULL
		ORDER BY code
	`).Scan(&accounts).Error

	if err != nil {
		log.Printf("Error fetching accounts: %v", err)
		return
	}

	// Group accounts by parent
	parentMap := make(map[uint][]AccountDetail)
	rootAccounts := []AccountDetail{}

	for _, acc := range accounts {
		if acc.ParentID == nil {
			rootAccounts = append(rootAccounts, acc)
		} else {
			parentMap[*acc.ParentID] = append(parentMap[*acc.ParentID], acc)
		}
	}

	// Print hierarchy recursively
	var printHierarchy func(account AccountDetail, level int)
	printHierarchy = func(account AccountDetail, level int) {
		indent := ""
		for i := 0; i < level; i++ {
			indent += "  "
		}
		
		headerStatus := "Detail"
		if account.IsHeader {
			headerStatus = "Header"
		}

		fmt.Printf("%s%s: %s [%s] - Balance: %.2f (Type: %s)\n", 
			indent, account.Code, account.Name, headerStatus, account.Balance, account.Type)

		// Print children
		if children, exists := parentMap[account.ID]; exists {
			for _, child := range children {
				printHierarchy(child, level+1)
			}
		}
	}

	for _, root := range rootAccounts {
		printHierarchy(root, 0)
	}

	// Check for any misplaced balances
	fmt.Println("\n2. Accounts with Non-Zero Balances:")
	for _, acc := range accounts {
		if acc.Balance != 0 {
			fmt.Printf("  %s: %s - Balance: %.2f (Header: %t, Type: %s)\n", 
				acc.Code, acc.Name, acc.Balance, acc.IsHeader, acc.Type)
		}
	}

	// Check for potential parent-child sum issues
	fmt.Println("\n3. Parent-Child Balance Verification:")
	for _, acc := range accounts {
		if acc.IsHeader {
			if children, exists := parentMap[acc.ID]; exists {
				childSum := 0.0
				for _, child := range children {
					childSum += child.Balance
				}
				
				diff := acc.Balance - childSum
				status := "✅ CORRECT"
				if diff != 0 {
					status = "❌ MISMATCH"
				}
				
				fmt.Printf("  %s: %s - Parent: %.2f, Children Sum: %.2f, Diff: %.2f %s\n",
					acc.Code, acc.Name, acc.Balance, childSum, diff, status)
			}
		}
	}

	fmt.Println("\n=== Account Hierarchy Check Complete ===")
}