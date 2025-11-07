package main

import (
	"fmt"
	"app-sistem-akuntansi/database"
)

func main() {
	db := database.ConnectDB()
	
	fmt.Println("=== DETAILED CASH ACCOUNTS ANALYSIS ===")
	
	// Get all cash accounts with detailed info
	fmt.Println("\n1. ALL CASH ACCOUNTS FROM cash_banks table:")
	rows, _ := db.Raw(`
		SELECT id, code, name, type, balance, is_active, account_id, currency, created_at
		FROM cash_banks 
		WHERE type = 'CASH'
		ORDER BY id
	`).Rows()
	defer rows.Close()
	
	totalCashFromDB := 0.0
	activeCashCount := 0
	
	for rows.Next() {
		var id uint
		var code, name, cbType, currency string
		var balance float64
		var isActive bool
		var accountID *uint
		var createdAt string
		
		rows.Scan(&id, &code, &name, &cbType, &balance, &isActive, &accountID, &currency, &createdAt)
		
		status := "INACTIVE"
		if isActive {
			status = "ACTIVE"
			totalCashFromDB += balance
			activeCashCount++
		}
		
		fmt.Printf("ID: %d | Code: %s | Name: %s | Balance: %.2f | Status: %s | AccountID: %v | Currency: %s\n", 
			id, code, name, balance, status, accountID, currency)
	}
	
	fmt.Printf("\nActive Cash Accounts: %d\n", activeCashCount)
	fmt.Printf("Total Cash Balance (Raw DB): %.2f\n", totalCashFromDB)
	
	// Get cash accounts that match summary calculation
	fmt.Println("\n2. CASH ACCOUNTS USED IN SUMMARY CALCULATION:")
	var summaryTotal float64
	db.Table("cash_banks").
		Where("type = ? AND is_active = ?", "CASH", true).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&summaryTotal)
	
	fmt.Printf("Summary Total (from repository logic): %.2f\n", summaryTotal)
	
	// Check specific accounts with non-zero balances
	fmt.Println("\n3. CASH ACCOUNTS WITH NON-ZERO BALANCES:")
	rows2, _ := db.Raw(`
		SELECT id, code, name, balance, is_active
		FROM cash_banks 
		WHERE type = 'CASH' AND balance != 0
		ORDER BY balance DESC
	`).Rows()
	defer rows2.Close()
	
	nonZeroCount := 0
	for rows2.Next() {
		var id uint
		var code, name string
		var balance float64
		var isActive bool
		
		rows2.Scan(&id, &code, &name, &balance, &isActive)
		
		status := "INACTIVE"
		if isActive {
			status = "ACTIVE"
		}
		
		fmt.Printf("ID: %d | Code: %s | Name: %s | Balance: %.2f | Status: %s\n", 
			id, code, name, balance, status)
		nonZeroCount++
	}
	
	if nonZeroCount == 0 {
		fmt.Println("No cash accounts with non-zero balances found!")
	}
	
	// Check the service layer GetCashBankAccounts result
	fmt.Println("\n4. TESTING SERVICE LAYER GetCashBankAccounts():")
	rows3, _ := db.Raw(`
		SELECT cb.id, cb.code, cb.name, cb.type, cb.balance, cb.is_active, 
		       a.id as account_id, a.name as account_name, a.balance as coa_balance
		FROM cash_banks cb
		LEFT JOIN accounts a ON cb.account_id = a.id  
		WHERE cb.is_active = true
		ORDER BY cb.type, cb.id
	`).Rows()
	defer rows3.Close()
	
	fmt.Println("ID | Code | Name | Type | CB_Balance | COA_Balance | AccountID | AccountName")
	fmt.Println("---|------|------|------|------------|-------------|-----------|-------------")
	
	for rows3.Next() {
		var id uint
		var code, name, cbType string
		var balance, coaBalance float64
		var isActive bool
		var accountID *uint
		var accountName *string
		
		rows3.Scan(&id, &code, &name, &cbType, &balance, &isActive, &accountID, &accountName, &coaBalance)
		
		accountIDStr := "NULL"
		accountNameStr := "NULL"
		if accountID != nil {
			accountIDStr = fmt.Sprintf("%d", *accountID)
		}
		if accountName != nil {
			accountNameStr = *accountName
		}
		
		fmt.Printf("%d | %s | %s | %s | %.2f | %.2f | %s | %s\n", 
			id, code, name, cbType, balance, coaBalance, accountIDStr, accountNameStr)
	}
	
	// Check if there are any cash transactions
	fmt.Println("\n5. RECENT CASH TRANSACTIONS:")
	var transactionCount int64
	db.Table("cash_bank_transactions cbt").
		Joins("INNER JOIN cash_banks cb ON cbt.cash_bank_id = cb.id").
		Where("cb.type = ?", "CASH").
		Count(&transactionCount)
	
	fmt.Printf("Total cash transactions: %d\n", transactionCount)
	
	if transactionCount > 0 {
		rows4, _ := db.Raw(`
			SELECT cbt.id, cb.name, cbt.amount, cbt.balance_after, cbt.transaction_date, cbt.notes
			FROM cash_bank_transactions cbt
			INNER JOIN cash_banks cb ON cbt.cash_bank_id = cb.id
			WHERE cb.type = 'CASH'
			ORDER BY cbt.transaction_date DESC
			LIMIT 10
		`).Rows()
		defer rows4.Close()
		
		fmt.Println("\nRecent cash transactions:")
		for rows4.Next() {
			var id uint
			var name, notes, transactionDate string
			var amount, balanceAfter float64
			
			rows4.Scan(&id, &name, &amount, &balanceAfter, &transactionDate, &notes)
			fmt.Printf("TX ID: %d | Account: %s | Amount: %.2f | Balance After: %.2f | Date: %s | Notes: %s\n", 
				id, name, amount, balanceAfter, transactionDate, notes)
		}
	}
}