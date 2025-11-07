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
	
	fmt.Printf("=== CHECKING UNIFIED_JOURNAL_LINES FOR ACCOUNT ID %d ===\n", accountID)
	
	// Query unified_journal_lines table
	query := `
		SELECT 
			ujl.debit_amount,
			ujl.credit_amount,
			ujl.description,
			ujl.line_number,
			uil.entry_number,
			uil.created_at
		FROM unified_journal_lines ujl
		INNER JOIN unified_journal_ledger uil ON ujl.journal_id = uil.id
		WHERE ujl.account_id = $1
		ORDER BY uil.created_at;
	`

	rows, err := sqlDB.Query(query, accountID)
	if err != nil {
		log.Fatalf("Failed to query unified journal lines: %v", err)
	}
	defer rows.Close()

	var totalDebit, totalCredit float64
	
	for rows.Next() {
		var debitAmountStr, creditAmountStr, description, entryNumber, createdAt string
		var lineNumber int
		
		err := rows.Scan(&debitAmountStr, &creditAmountStr, &description, &lineNumber, &entryNumber, &createdAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		debitAmount, _ := strconv.ParseFloat(debitAmountStr, 64)
		creditAmount, _ := strconv.ParseFloat(creditAmountStr, 64)
		
		totalDebit += debitAmount
		totalCredit += creditAmount
		
		fmt.Printf("Entry: %s | Line %d | %s\n", entryNumber, lineNumber, createdAt[:19])
		fmt.Printf("  %s\n", description)
		fmt.Printf("  Debit: Rp %.0f | Credit: Rp %.0f\n", debitAmount, creditAmount)
		fmt.Println("  ---")
	}

	balance := totalDebit - totalCredit
	
	fmt.Println("\n=== UNIFIED JOURNAL SUMMARY ===")
	fmt.Printf("Total Debit:  Rp %.0f\n", totalDebit)
	fmt.Printf("Total Credit: Rp %.0f\n", totalCredit)
	fmt.Printf("Balance (Debit - Credit): Rp %.0f\n", balance)

	// Check what's in the regular journal_lines table too
	fmt.Println("\n=== CHECKING REGULAR JOURNAL_LINES ===")
	regularQuery := `
		SELECT COUNT(*) as count
		FROM journal_lines jl
		WHERE jl.account_id = $1;
	`
	
	var count int
	err = sqlDB.QueryRow(regularQuery, accountID).Scan(&count)
	if err != nil {
		log.Printf("Error getting count: %v", err)
	} else {
		fmt.Printf("Regular journal_lines count for account %d: %d\n", accountID, count)
	}
	
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