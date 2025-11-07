package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
)

func main() {
	db := database.ConnectDB()
	
	// Check cash_banks table structure and data
	fmt.Println("=== CASH BANK ACCOUNTS ===")
	rows, _ := db.Raw("SELECT id, name, type, balance, account_id, is_active FROM cash_banks WHERE is_active = true ORDER BY id").Rows()
	defer rows.Close()
	
	for rows.Next() {
		var id uint
		var name, cbType string
		var balance float64
		var accountID *uint
		var isActive bool
		rows.Scan(&id, &name, &cbType, &balance, &accountID, &isActive)
		fmt.Printf("ID: %d, Name: %s, Type: %s, Balance: %.2f, AccountID: %v, Active: %t\n", 
			id, name, cbType, balance, accountID, isActive)
	}
	
	// Check accounts table for related records
	fmt.Println("\n=== ACCOUNTS TABLE (for cash bank accounts) ===")
	rows2, _ := db.Raw(`SELECT a.id, a.name, a.balance, a.is_active 
						FROM accounts a 
						INNER JOIN cash_banks cb ON a.id = cb.account_id 
						WHERE cb.is_active = true AND a.is_active = true`).Rows()
	defer rows2.Close()
	
	for rows2.Next() {
		var id uint
		var name string
		var balance float64
		var isActive bool
		rows2.Scan(&id, &name, &balance, &isActive)
		fmt.Printf("Account ID: %d, Name: %s, Balance: %.2f, Active: %t\n", 
			id, name, balance, isActive)
	}
	
	// Test the actual query used in GetBalanceSummary
	fmt.Println("\n=== TESTING SUMMARY QUERIES ===")
	
	var cashTotal float64
	db.Table("accounts a").
		Joins("INNER JOIN cash_banks cb ON a.id = cb.account_id").
		Where("cb.type = ? AND cb.is_active = ? AND a.is_active = ?", "CASH", true, true).
		Select("COALESCE(SUM(a.balance), 0)").
		Scan(&cashTotal)
	fmt.Printf("Cash Total from COA: %.2f\n", cashTotal)
	
	var bankTotal float64
	db.Table("accounts a").
		Joins("INNER JOIN cash_banks cb ON a.id = cb.account_id").
		Where("cb.type = ? AND cb.is_active = ? AND a.is_active = ?", "BANK", true, true).
		Select("COALESCE(SUM(a.balance), 0)").
		Scan(&bankTotal)
	fmt.Printf("Bank Total from COA: %.2f\n", bankTotal)
	
	// Alternative query using cash_banks balance
	var cashTotalDirect float64
	db.Table("cash_banks").
		Where("type = ? AND is_active = ?", "CASH", true).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&cashTotalDirect)
	fmt.Printf("Cash Total from cash_banks: %.2f\n", cashTotalDirect)
	
	var bankTotalDirect float64
	db.Table("cash_banks").
		Where("type = ? AND is_active = ?", "BANK", true).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&bankTotalDirect)
	fmt.Printf("Bank Total from cash_banks: %.2f\n", bankTotalDirect)
	
	// Check if account_id is NULL
	var nullAccountIDs int64
	db.Table("cash_banks").
		Where("account_id IS NULL AND is_active = ?", true).
		Count(&nullAccountIDs)
	fmt.Printf("\nCash Bank accounts with NULL account_id: %d\n", nullAccountIDs)
}