package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database using DATABASE_URL from .env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get underlying sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Account ID 9 is Piutang Usaha
	accountID := 9
	
	// Query all journal lines for this account
	query := `
		SELECT 
			jl.debit_amount,
			jl.credit_amount,
			jl.description,
			je.code,
			je.created_at
		FROM journal_lines jl
		INNER JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE jl.account_id = $1
		ORDER BY je.created_at;
	`

	rows, err := sqlDB.Query(query, accountID)
	if err != nil {
		log.Fatalf("Failed to query journal entries: %v", err)
	}
	defer rows.Close()

	var totalDebit, totalCredit float64
	
	fmt.Printf("=== PIUTANG USAHA (Account ID %d) JOURNAL ENTRIES ===\n", accountID)
	
	for rows.Next() {
		var debitAmountStr, creditAmountStr, description, entryCode, createdAt string
		
		err := rows.Scan(&debitAmountStr, &creditAmountStr, &description, &entryCode, &createdAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		debitAmount, _ := strconv.ParseFloat(debitAmountStr, 64)
		creditAmount, _ := strconv.ParseFloat(creditAmountStr, 64)
		
		totalDebit += debitAmount
		totalCredit += creditAmount
		
		fmt.Printf("Entry: %s | %s\n", entryCode, createdAt[:19])
		fmt.Printf("  %s\n", description)
		fmt.Printf("  Debit: Rp %.0f | Credit: Rp %.0f\n", debitAmount, creditAmount)
		fmt.Println("  ---")
	}

	balance := totalDebit - totalCredit
	
	fmt.Println("\n=== SUMMARY ===")
	fmt.Printf("Total Debit:  Rp %.0f\n", totalDebit)
	fmt.Printf("Total Credit: Rp %.0f\n", totalCredit)
	fmt.Printf("Balance (Debit - Credit): Rp %.0f\n", balance)
	
	// Check current balance in accounts table
	var currentBalance float64
	err = sqlDB.QueryRow("SELECT balance FROM accounts WHERE id = $1", accountID).Scan(&currentBalance)
	if err != nil {
		log.Printf("Error getting current balance: %v", err)
	} else {
		fmt.Printf("Current balance in accounts table: Rp %.0f\n", currentBalance)
		
		if balance != currentBalance {
			fmt.Printf("❌ MISMATCH! Calculated: %.0f vs Stored: %.0f\n", balance, currentBalance)
		} else {
			fmt.Printf("✅ MATCH! Both show: %.0f\n", balance)
		}
	}
}