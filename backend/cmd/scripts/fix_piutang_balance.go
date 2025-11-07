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
	
	fmt.Printf("🔍 Recalculating balance for Account ID %d (Piutang Usaha)...\n", accountID)
	
	// Get current balance
	var currentBalance float64
	err = sqlDB.QueryRow("SELECT balance FROM accounts WHERE id = $1", accountID).Scan(&currentBalance)
	if err != nil {
		log.Fatalf("Error getting current balance: %v", err)
	}
	fmt.Printf("Current stored balance: Rp %.0f\n", currentBalance)
	
	// Calculate correct balance from unified_journal_lines
	query := `
		SELECT 
			COALESCE(SUM(CAST(ujl.debit_amount AS DECIMAL)), 0) as total_debit,
			COALESCE(SUM(CAST(ujl.credit_amount AS DECIMAL)), 0) as total_credit
		FROM unified_journal_lines ujl
		WHERE ujl.account_id = $1;
	`

	var totalDebitStr, totalCreditStr string
	err = sqlDB.QueryRow(query, accountID).Scan(&totalDebitStr, &totalCreditStr)
	if err != nil {
		log.Fatalf("Failed to calculate balance: %v", err)
	}

	totalDebit, _ := strconv.ParseFloat(totalDebitStr, 64)
	totalCredit, _ := strconv.ParseFloat(totalCreditStr, 64)
	correctBalance := totalDebit - totalCredit
	
	fmt.Printf("Calculated from unified_journal_lines:\n")
	fmt.Printf("  Total Debit:  Rp %.0f\n", totalDebit)
	fmt.Printf("  Total Credit: Rp %.0f\n", totalCredit)
	fmt.Printf("  Net Balance:  Rp %.0f\n", correctBalance)
	
	if correctBalance == currentBalance {
		fmt.Printf("✅ Balance is already correct!\n")
		return
	}
	
	fmt.Printf("\n❌ Balance mismatch detected!\n")
	fmt.Printf("   Current: Rp %.0f\n", currentBalance)
	fmt.Printf("   Correct: Rp %.0f\n", correctBalance)
	fmt.Printf("   Difference: Rp %.0f\n", correctBalance - currentBalance)
	
	// Update the balance in accounts table
	fmt.Printf("\n🔧 Updating account balance...\n")
	
	updateQuery := `UPDATE accounts SET balance = $1, updated_at = NOW() WHERE id = $2`
	result, err := sqlDB.Exec(updateQuery, correctBalance, accountID)
	if err != nil {
		log.Fatalf("Failed to update balance: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Warning: Could not get rows affected: %v", err)
	} else if rowsAffected == 0 {
		log.Fatalf("No rows were updated - account ID %d not found", accountID)
	}
	
	// Verify the update
	var newBalance float64
	err = sqlDB.QueryRow("SELECT balance FROM accounts WHERE id = $1", accountID).Scan(&newBalance)
	if err != nil {
		log.Printf("Warning: Could not verify new balance: %v", err)
	} else {
		fmt.Printf("✅ Balance updated successfully!\n")
		fmt.Printf("   Old balance: Rp %.0f\n", currentBalance)
		fmt.Printf("   New balance: Rp %.0f\n", newBalance)
		
		if newBalance == correctBalance {
			fmt.Printf("✅ Verification passed - balance is now correct!\n")
		} else {
			fmt.Printf("❌ Verification failed - balance mismatch: %.0f vs %.0f\n", newBalance, correctBalance)
		}
	}
	
	fmt.Printf("\n🎉 Piutang Usaha balance fix completed!\n")
	fmt.Printf("   Account ID %d now shows: Rp %.0f\n", accountID, correctBalance)
	fmt.Printf("   This represents the remaining accounts receivable after payment.\n")
}