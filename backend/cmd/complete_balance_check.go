package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
)

type Account struct {
	Code    string
	Name    string
	Type    string
	Balance float64
}

func main() {
	db := database.ConnectDB()
	
	var accounts []Account
	db.Raw(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE deleted_at IS NULL 
		AND balance != 0
		AND COALESCE(is_header, false) = false
		ORDER BY code
	`).Scan(&accounts)
	
	fmt.Println("\n=== ALL NON-ZERO ACCOUNT BALANCES ===\n")
	
	totalAssets := 0.0
	totalLiabilities := 0.0
	totalEquity := 0.0
	totalRevenue := 0.0
	totalExpense := 0.0
	
	fmt.Println("ASSETS:")
	for _, acc := range accounts {
		if acc.Type == "ASSET" {
			fmt.Printf("  %s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalAssets += acc.Balance
		}
	}
	fmt.Printf("TOTAL ASSETS: %.2f\n\n", totalAssets)
	
	fmt.Println("LIABILITIES:")
	for _, acc := range accounts {
		if acc.Type == "LIABILITY" {
			fmt.Printf("  %s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalLiabilities += acc.Balance
		}
	}
	fmt.Printf("TOTAL LIABILITIES: %.2f\n\n", totalLiabilities)
	
	fmt.Println("EQUITY:")
	for _, acc := range accounts {
		if acc.Type == "EQUITY" {
			fmt.Printf("  %s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalEquity += acc.Balance
		}
	}
	fmt.Printf("TOTAL EQUITY: %.2f\n\n", totalEquity)
	
	fmt.Println("REVENUE:")
	for _, acc := range accounts {
		if acc.Type == "REVENUE" {
			fmt.Printf("  %s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalRevenue += acc.Balance
		}
	}
	fmt.Printf("TOTAL REVENUE: %.2f\n\n", totalRevenue)
	
	fmt.Println("EXPENSE:")
	for _, acc := range accounts {
		if acc.Type == "EXPENSE" {
			fmt.Printf("  %s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalExpense += acc.Balance
		}
	}
	fmt.Printf("TOTAL EXPENSE: %.2f\n\n", totalExpense)
	
	fmt.Println("=== BALANCE SHEET EQUATION ===")
	fmt.Printf("Assets:                     %.2f\n", totalAssets)
	fmt.Printf("Liabilities + Equity:       %.2f + %.2f = %.2f\n", 
		totalLiabilities, totalEquity, totalLiabilities+totalEquity)
	
	diff := totalAssets - (totalLiabilities + totalEquity)
	fmt.Printf("Difference:                 %.2f\n", diff)
	
	if diff == 0 {
		fmt.Println("✅ BALANCED!")
	} else {
		fmt.Println("❌ NOT BALANCED!")
		
		netIncome := totalRevenue - totalExpense
		fmt.Printf("\nNet Income (Revenue-Expense): %.2f\n", netIncome)
		
		if netIncome != 0 {
			fmt.Println("⚠️  Revenue/Expense accounts still have balances (period not closed)")
		}
	}
}
