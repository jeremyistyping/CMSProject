package main

import (
	"fmt"
	"log"
	"os"

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

	fmt.Printf("🔄 TRANSFERRING REVENUE BALANCE...\n\n")

	// Step 1: Check current balances
	var revenueBalance, pendapatanPenjualanBalance float64
	
	err = sqlDB.QueryRow("SELECT balance FROM accounts WHERE id = 23").Scan(&revenueBalance)
	if err != nil {
		log.Printf("Error getting REVENUE balance: %v", err)
		return
	}
	
	err = sqlDB.QueryRow("SELECT balance FROM accounts WHERE id = 24").Scan(&pendapatanPenjualanBalance)
	if err != nil {
		log.Printf("Error getting Pendapatan Penjualan balance: %v", err)
		return
	}

	fmt.Printf("Current Balances:\n")
	fmt.Printf("  Account 23 (REVENUE): Rp %.0f\n", revenueBalance)
	fmt.Printf("  Account 24 (Pendapatan Penjualan): Rp %.0f\n", pendapatanPenjualanBalance)

	if revenueBalance == 0 {
		fmt.Printf("❌ No balance to transfer from REVENUE account\n")
		return
	}

	// Step 2: Create journal entry to move the balance
	fmt.Printf("\n🔧 Creating journal entry to transfer Rp %.0f...\n", -revenueBalance)

	// Get next entry number
	var maxEntryNum int
	err = sqlDB.QueryRow(`
		SELECT COALESCE(MAX(CAST(SUBSTRING(entry_number FROM '[0-9]+') AS INTEGER)), 0)
		FROM unified_journal_ledger 
		WHERE entry_number LIKE 'JE-%'
	`).Scan(&maxEntryNum)
	
	if err != nil {
		log.Printf("Error getting max entry number: %v", err)
		return
	}
	
	nextEntryNum := fmt.Sprintf("JE-%05d", maxEntryNum+1)
	transferAmount := -revenueBalance // Convert to positive for transfer

	// Create journal ledger entry
	var journalID int
	err = sqlDB.QueryRow(`
		INSERT INTO unified_journal_ledger (
			entry_number, source_type, source_code, entry_date, 
			description, total_debit, total_credit, status, 
			is_balanced, is_auto_generated, created_by, created_at, updated_at
		) VALUES (
			$1, 'MANUAL', 'REVENUE_TRANSFER', CURRENT_DATE,
			'Transfer revenue from generic REVENUE to Pendapatan Penjualan', 
			$2, $2, 'POSTED', 
			true, true, 1, NOW(), NOW()
		) RETURNING id
	`, nextEntryNum, transferAmount).Scan(&journalID)

	if err != nil {
		log.Printf("Error creating journal ledger: %v", err)
		return
	}

	// Create debit entry for REVENUE (to reduce its credit balance)
	_, err = sqlDB.Exec(`
		INSERT INTO unified_journal_lines (
			journal_id, account_id, line_number, description,
			debit_amount, credit_amount, created_at, updated_at
		) VALUES ($1, 23, 1, 'Transfer to specific revenue account', $2, 0, NOW(), NOW())
	`, journalID, transferAmount)

	if err != nil {
		log.Printf("Error creating debit journal line: %v", err)
		return
	}

	// Create credit entry for Pendapatan Penjualan
	_, err = sqlDB.Exec(`
		INSERT INTO unified_journal_lines (
			journal_id, account_id, line_number, description,
			debit_amount, credit_amount, created_at, updated_at
		) VALUES ($1, 24, 2, 'Revenue from sales - Mouse Wireless Logitech', 0, $2, NOW(), NOW())
	`, journalID, transferAmount)

	if err != nil {
		log.Printf("Error creating credit journal line: %v", err)
		return
	}

	fmt.Printf("✅ Created balanced journal entry %s\n", nextEntryNum)

	// Step 3: Run balance sync to update account balances
	fmt.Printf("\n🔄 Running balance sync...\n")
	rows, err := sqlDB.Query("SELECT * FROM sync_account_balances()")
	if err != nil {
		log.Printf("Warning: Could not run balance sync: %v", err)
	} else {
		defer rows.Close()
		
		fmt.Printf("Balance updates:\n")
		for rows.Next() {
			var accountID int
			var oldBalance, newBalance, difference float64
			
			err := rows.Scan(&accountID, &oldBalance, &newBalance, &difference)
			if err != nil {
				continue
			}
			
			if accountID == 23 || accountID == 24 {
				var accountName string
				if accountID == 23 {
					accountName = "REVENUE"
				} else {
					accountName = "Pendapatan Penjualan"
				}
				fmt.Printf("  Account %d (%s): Rp %.0f → Rp %.0f\n", 
					accountID, accountName, oldBalance, newBalance)
			}
		}
	}

	// Step 4: Verify final balances
	fmt.Printf("\n✅ VERIFICATION:\n")
	err = sqlDB.QueryRow("SELECT balance FROM accounts WHERE id = 23").Scan(&revenueBalance)
	if err == nil {
		fmt.Printf("  Account 23 (REVENUE): Rp %.0f\n", revenueBalance)
	}
	
	err = sqlDB.QueryRow("SELECT balance FROM accounts WHERE id = 24").Scan(&pendapatanPenjualanBalance)
	if err == nil {
		fmt.Printf("  Account 24 (Pendapatan Penjualan): Rp %.0f\n", pendapatanPenjualanBalance)
	}

	if pendapatanPenjualanBalance < 0 {
		fmt.Printf("✅ Success! Revenue now properly recorded in Pendapatan Penjualan\n")
	} else {
		fmt.Printf("⚠️  Please check - Pendapatan Penjualan should show negative balance (credit normal)\n")
	}

	fmt.Printf("\n🎉 REVENUE TRANSFER COMPLETED!\n")
	fmt.Printf("Now 'Pendapatan Penjualan' should show the revenue from sales.\n")
}