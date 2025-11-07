package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	fmt.Printf("🔍 ANALYZING PAYMENT ISSUE...\n\n")

	// 1. Check all journal entries
	fmt.Printf("=== ALL JOURNAL ENTRIES ===\n")
	query := `
		SELECT 
			ujl.account_id,
			a.name as account_name,
			a.type as account_type,
			ujl.debit_amount,
			ujl.credit_amount,
			ujl.description,
			uil.entry_number,
			uil.created_at
		FROM unified_journal_lines ujl
		INNER JOIN unified_journal_ledger uil ON ujl.journal_id = uil.id
		INNER JOIN accounts a ON ujl.account_id = a.id
		ORDER BY uil.created_at, ujl.line_number;
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		log.Fatalf("Failed to query journal entries: %v", err)
	}
	defer rows.Close()

	accountBalances := make(map[int]float64)
	
	for rows.Next() {
		var accountID int
		var accountName, accountType, debitAmountStr, creditAmountStr, description, entryNumber, createdAt string
		
		err := rows.Scan(&accountID, &accountName, &accountType, &debitAmountStr, &creditAmountStr, &description, &entryNumber, &createdAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		debitAmount, _ := strconv.ParseFloat(debitAmountStr, 64)
		creditAmount, _ := strconv.ParseFloat(creditAmountStr, 64)
		
		// Track running balance
		if _, exists := accountBalances[accountID]; !exists {
			accountBalances[accountID] = 0
		}
		accountBalances[accountID] += debitAmount - creditAmount
		
		fmt.Printf("%s | %s\n", entryNumber, createdAt[:19])
		fmt.Printf("  Account %d (%s - %s)\n", accountID, accountName, accountType)
		fmt.Printf("  %s\n", description)
		fmt.Printf("  Debit: Rp %.0f | Credit: Rp %.0f\n", debitAmount, creditAmount)
		fmt.Printf("  Running Balance: Rp %.0f\n", accountBalances[accountID])
		fmt.Println("  ---")
	}

	// 2. Compare with actual account balances
	fmt.Printf("\n=== BALANCE COMPARISON ===\n")
	balanceQuery := `
		SELECT id, name, type, balance 
		FROM accounts 
		WHERE id IN (6, 9, 23, 40) 
		ORDER BY id;
	`
	
	balanceRows, err := sqlDB.Query(balanceQuery)
	if err != nil {
		log.Printf("Error getting account balances: %v", err)
		return
	}
	defer balanceRows.Close()

	fmt.Printf("%-4s %-20s %-12s %12s %12s %12s %s\n", 
		"ID", "Name", "Type", "Calculated", "Stored", "Difference", "Status")
	fmt.Println(strings.Repeat("-", 80))

	for balanceRows.Next() {
		var id int
		var name, accountType string
		var storedBalance float64
		
		err := balanceRows.Scan(&id, &name, &accountType, &storedBalance)
		if err != nil {
			log.Printf("Error scanning balance: %v", err)
			continue
		}

		calculatedBalance := accountBalances[id]
		difference := calculatedBalance - storedBalance
		status := "✅ OK"
		if difference != 0 {
			status = "❌ MISMATCH"
		}

		fmt.Printf("%-4d %-20s %-12s %12.0f %12.0f %12.0f %s\n", 
			id, name, accountType, calculatedBalance, storedBalance, difference, status)
	}

	// 3. Analyze the problem
	fmt.Printf("\n=== PROBLEM ANALYSIS ===\n")
	
	piutangCalculated := accountBalances[9]
	bankCalculated := accountBalances[6]
	
	fmt.Printf("Piutang Usaha (ID 9):\n")
	fmt.Printf("  - Calculated from journals: Rp %.0f\n", piutangCalculated)
	fmt.Printf("  - Should be: Rp 0 (fully paid)\n")
	
	fmt.Printf("\nBANK UOB (ID 6):\n")
	fmt.Printf("  - Calculated from journals: Rp %.0f\n", bankCalculated)
	fmt.Printf("  - Should be: Rp 5,550,000 (full payment received)\n")

	if piutangCalculated == 0 && bankCalculated == 5550000 {
		fmt.Printf("\n✅ JOURNAL ENTRIES ARE CORRECT!\n")
		fmt.Printf("The negative balance in accounts table is the issue.\n")
		fmt.Printf("Need to run balance sync to fix stored balances.\n")
	} else {
		fmt.Printf("\n❌ JOURNAL ENTRIES NEED REVIEW\n")
		fmt.Printf("There might be duplicate or incorrect entries.\n")
	}

	// 4. Check payment records
	fmt.Printf("\n=== PAYMENT ANALYSIS ===\n")
	paymentQuery := `
		SELECT id, amount, payment_date, reference, notes
		FROM payments
		ORDER BY created_at;
	`
	
	paymentRows, err := sqlDB.Query(paymentQuery)
	if err != nil {
		log.Printf("Error getting payments: %v", err)
		return
	}
	defer paymentRows.Close()

	totalPayments := 0.0
	paymentCount := 0
	
	for paymentRows.Next() {
		var id int
		var amount float64
		var paymentDate, reference, notes string
		
		err := paymentRows.Scan(&id, &amount, &paymentDate, &reference, &notes)
		if err != nil {
			log.Printf("Error scanning payment: %v", err)
			continue
		}

		totalPayments += amount
		paymentCount++
		
		fmt.Printf("Payment %d: Rp %.0f on %s\n", id, amount, paymentDate[:10])
		fmt.Printf("  Reference: %s\n", reference)
		fmt.Printf("  Notes: %s\n", notes)
		fmt.Println("  ---")
	}

	fmt.Printf("\nPayment Summary:\n")
	fmt.Printf("  Total payments: %d\n", paymentCount)
	fmt.Printf("  Total amount: Rp %.0f\n", totalPayments)
	
	if totalPayments == 5550000 {
		fmt.Printf("✅ Total payments match sale amount\n")
	} else {
		fmt.Printf("❌ Payment total mismatch\n")
	}

	fmt.Printf("\n=== RECOMMENDATION ===\n")
	fmt.Printf("1. Run balance sync to fix stored balances\n")
	fmt.Printf("2. Verify that Piutang Usaha should be 0 (not negative)\n")
	fmt.Printf("3. Ensure bank balance reflects total payments received\n")
}