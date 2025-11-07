package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CashBank struct {
	ID             uint    `gorm:"primaryKey"`
	AccountCode    string  `gorm:"column:account_code"`
	AccountName    string  `gorm:"column:account_name"`
	CurrentBalance float64 `gorm:"column:current_balance"`
	CreatedAt      string  `gorm:"column:created_at"`
	UpdatedAt      string  `gorm:"column:updated_at"`
}

type JournalEntry struct {
	ID           uint    `gorm:"primaryKey"`
	AccountCode  string  `gorm:"column:account_code"`
	DebitAmount  float64 `gorm:"column:debit_amount"`
	CreditAmount float64 `gorm:"column:credit_amount"`
	Description  string  `gorm:"column:description"`
	CreatedAt    string  `gorm:"column:created_at"`
	TransactionID string `gorm:"column:transaction_id"`
}

type PaymentTransaction struct {
	ID            uint    `gorm:"primaryKey"`
	TransactionID string  `gorm:"column:transaction_id"`
	AccountCode   string  `gorm:"column:account_code"`
	Amount        float64 `gorm:"column:amount"`
	Description   string  `gorm:"column:description"`
	CreatedAt     string  `gorm:"column:created_at"`
}

func main() {
	// Connect to database
	dsn := "root:password@tcp(localhost:3306)/accounting_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("=== DEBUGGING DOUBLE POSTING FOR BNK-2025-0004 ===")
	
	// 1. Check current balance in cash_bank table
	var cashBank CashBank
	result := db.Table("cash_bank").Where("account_code = ?", "BNK-2025-0004").First(&cashBank)
	if result.Error != nil {
		fmt.Printf("Error fetching cash_bank: %v\n", result.Error)
	} else {
		fmt.Printf("\n1. Current Balance in cash_bank table:\n")
		fmt.Printf("   Account Code: %s\n", cashBank.AccountCode)
		fmt.Printf("   Account Name: %s\n", cashBank.AccountName)
		fmt.Printf("   Current Balance: %.2f\n", cashBank.CurrentBalance)
		fmt.Printf("   Last Updated: %s\n", cashBank.UpdatedAt)
	}

	// 2. Check all journal entries for this account
	var journalEntries []JournalEntry
	result = db.Table("journal_entries").Where("account_code = ?", "BNK-2025-0004").Order("created_at DESC").Find(&journalEntries)
	if result.Error != nil {
		fmt.Printf("Error fetching journal entries: %v\n", result.Error)
	} else {
		fmt.Printf("\n2. All Journal Entries for BNK-2025-0004:\n")
		totalDebit := 0.0
		totalCredit := 0.0
		
		for i, entry := range journalEntries {
			fmt.Printf("   [%d] ID: %d, Debit: %.2f, Credit: %.2f, Description: %s, Transaction: %s, Created: %s\n", 
				i+1, entry.ID, entry.DebitAmount, entry.CreditAmount, entry.Description, entry.TransactionID, entry.CreatedAt)
			totalDebit += entry.DebitAmount
			totalCredit += entry.CreditAmount
		}
		
		fmt.Printf("\n   Total Debits: %.2f\n", totalDebit)
		fmt.Printf("   Total Credits: %.2f\n", totalCredit)
		fmt.Printf("   Net Balance (Debit-Credit): %.2f\n", totalDebit-totalCredit)
	}

	// 3. Check payment transactions
	var payments []PaymentTransaction
	result = db.Table("payment_transactions").Where("account_code = ?", "BNK-2025-0004").Order("created_at DESC").Find(&payments)
	if result.Error != nil {
		fmt.Printf("Error fetching payment transactions: %v\n", result.Error)
	} else {
		fmt.Printf("\n3. Payment Transactions for BNK-2025-0004:\n")
		for i, payment := range payments {
			fmt.Printf("   [%d] Transaction ID: %s, Amount: %.2f, Description: %s, Created: %s\n", 
				i+1, payment.TransactionID, payment.Amount, payment.Description, payment.CreatedAt)
		}
	}

	// 4. Check for duplicate journal entries
	fmt.Printf("\n4. Checking for Duplicate Journal Entries:\n")
	for _, payment := range payments {
		var duplicates []JournalEntry
		result = db.Table("journal_entries").Where("transaction_id = ? AND account_code = ?", payment.TransactionID, "BNK-2025-0004").Find(&duplicates)
		if len(duplicates) > 1 {
			fmt.Printf("   DUPLICATE FOUND for Transaction %s: %d journal entries\n", payment.TransactionID, len(duplicates))
			for _, dup := range duplicates {
				fmt.Printf("      - ID: %d, Debit: %.2f, Credit: %.2f, Description: %s\n", 
					dup.ID, dup.DebitAmount, dup.CreditAmount, dup.Description)
			}
		}
	}

	fmt.Printf("\n=== ANALYSIS COMPLETE ===\n")
}