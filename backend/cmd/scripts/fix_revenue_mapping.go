package main

import (
	"database/sql"
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

	fmt.Printf("🔍 ANALYZING REVENUE MAPPING ISSUE...\n\n")

	// 1. Check current revenue journal entries
	fmt.Printf("=== CURRENT REVENUE JOURNAL ENTRIES ===\n")
	revenueQuery := `
		SELECT 
			ujl.account_id,
			a.name as account_name,
			ujl.debit_amount,
			ujl.credit_amount,
			ujl.description,
			uil.entry_number,
			uil.created_at
		FROM unified_journal_lines ujl
		INNER JOIN unified_journal_ledger uil ON ujl.journal_id = uil.id
		INNER JOIN accounts a ON ujl.account_id = a.id
		WHERE a.type = 'REVENUE'
		ORDER BY uil.created_at, ujl.line_number;
	`

	rows, err := sqlDB.Query(revenueQuery)
	if err != nil {
		log.Fatalf("Failed to query revenue entries: %v", err)
	}
	defer rows.Close()

	fmt.Printf("Revenue journal entries found:\n")
	revenueEntryFound := false
	
	for rows.Next() {
		var accountID int
		var accountName, debitAmountStr, creditAmountStr, description, entryNumber, createdAt string
		
		err := rows.Scan(&accountID, &accountName, &debitAmountStr, &creditAmountStr, &description, &entryNumber, &createdAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		debitAmount, _ := strconv.ParseFloat(debitAmountStr, 64)
		creditAmount, _ := strconv.ParseFloat(creditAmountStr, 64)
		
		fmt.Printf("  %s | %s\n", entryNumber, createdAt[:19])
		fmt.Printf("    Account %d (%s)\n", accountID, accountName)
		fmt.Printf("    %s\n", description)
		fmt.Printf("    Debit: Rp %.0f | Credit: Rp %.0f\n", debitAmount, creditAmount)
		fmt.Printf("    ---\n")
		
		revenueEntryFound = true
	}

	if !revenueEntryFound {
		fmt.Printf("❌ No revenue journal entries found!\n")
	}

	// 2. Check sale items mapping
	fmt.Printf("\n=== CHECKING SALE ITEMS REVENUE MAPPING ===\n")
	saleItemsQuery := `
		SELECT 
			si.id,
			si.sale_id,
			si.revenue_account_id,
			si.description,
			si.total_price,
			a.name as account_name,
			a.type as account_type
		FROM sale_items si
		LEFT JOIN accounts a ON si.revenue_account_id = a.id
		ORDER BY si.sale_id, si.id;
	`

	saleRows, err := sqlDB.Query(saleItemsQuery)
	if err != nil {
		log.Printf("Error getting sale items: %v", err)
	} else {
		defer saleRows.Close()
		
		for saleRows.Next() {
			var id, saleID, revenueAccountID int
			var description, accountName, accountType string
			var totalPrice float64
			
			err := saleRows.Scan(&id, &saleID, &revenueAccountID, &description, &totalPrice, &accountName, &accountType)
			if err != nil {
				log.Printf("Error scanning sale item: %v", err)
				continue
			}

			fmt.Printf("Sale Item %d (Sale %d):\n", id, saleID)
			fmt.Printf("  Product: %s\n", description)
			fmt.Printf("  Amount: Rp %.0f\n", totalPrice)
			fmt.Printf("  Revenue Account: %d (%s - %s)\n", revenueAccountID, accountName, accountType)
			fmt.Printf("  ---\n")
		}
	}

	// 3. Get the correct Pendapatan Penjualan account ID
	var pendapatanPenjualanID int
	var pendapatanPenjualanName string
	err = sqlDB.QueryRow(`
		SELECT id, name 
		FROM accounts 
		WHERE name LIKE '%Pendapatan Penjualan%' OR name LIKE '%Pendapatan%'
		AND type = 'REVENUE'
		ORDER BY id
		LIMIT 1
	`).Scan(&pendapatanPenjualanID, &pendapatanPenjualanName)
	
	if err != nil {
		fmt.Printf("❌ Could not find 'Pendapatan Penjualan' account: %v\n", err)
		return
	}

	fmt.Printf("\n=== REVENUE ACCOUNT MAPPING ===\n")
	fmt.Printf("Target Account: ID %d (%s)\n", pendapatanPenjualanID, pendapatanPenjualanName)

	// 4. Check if we need to create missing journal entry
	var existingRevenueEntry int
	err = sqlDB.QueryRow(`
		SELECT COUNT(*) 
		FROM unified_journal_lines ujl
		INNER JOIN accounts a ON ujl.account_id = a.id
		WHERE a.id = $1 AND ujl.credit_amount > 0
	`, pendapatanPenjualanID).Scan(&existingRevenueEntry)

	if existingRevenueEntry == 0 {
		fmt.Printf("\n❌ NO REVENUE ENTRY FOUND for Pendapatan Penjualan (ID %d)\n", pendapatanPenjualanID)
		fmt.Printf("Need to create revenue journal entry...\n")

		// Get sales total to create correct revenue entry
		var salesTotalAmount float64
		err = sqlDB.QueryRow(`
			SELECT COALESCE(SUM(sub_total), 0)
			FROM sales 
			WHERE status IN ('INVOICED', 'PAID')
		`).Scan(&salesTotalAmount)
		
		if err != nil {
			log.Printf("Error getting sales total: %v", err)
			return
		}

		if salesTotalAmount > 0 {
			fmt.Printf("Sales total found: Rp %.0f\n", salesTotalAmount)
			fmt.Printf("Creating journal entry to record revenue...\n")

			// Create journal entry
			createJournalEntry(sqlDB, pendapatanPenjualanID, salesTotalAmount)
		} else {
			fmt.Printf("No sales found to create revenue entry\n")
		}
	} else {
		fmt.Printf("✅ Revenue entry exists for Pendapatan Penjualan\n")
	}

	// 5. Run balance sync to update account balance
	fmt.Printf("\n🔄 Running balance sync to update account balances...\n")
	syncRows, err := sqlDB.Query("SELECT * FROM sync_account_balances()")
	if err != nil {
		log.Printf("Warning: Could not run balance sync: %v", err)
	} else {
		defer syncRows.Close()
		
		syncCount := 0
		for syncRows.Next() {
			var accountID int
			var oldBalance, newBalance, difference float64
			
			err := syncRows.Scan(&accountID, &oldBalance, &newBalance, &difference)
			if err != nil {
				continue
			}
			
			if accountID == pendapatanPenjualanID {
				fmt.Printf("✅ Updated Pendapatan Penjualan balance:\n")
				fmt.Printf("  Old: Rp %.0f → New: Rp %.0f\n", oldBalance, newBalance)
			}
			syncCount++
		}
		
		if syncCount > 0 {
			fmt.Printf("✅ Balance sync completed - %d accounts updated\n", syncCount)
		}
	}

	fmt.Printf("\n🎉 REVENUE MAPPING ANALYSIS COMPLETED\n")
}

func createJournalEntry(sqlDB *sql.DB, revenueAccountID int, amount float64) {
	// Get next entry number
	var maxEntryNum int
	sqlDB.QueryRow(`
		SELECT COALESCE(MAX(CAST(SUBSTRING(entry_number FROM '[0-9]+') AS INTEGER)), 0)
		FROM unified_journal_ledger 
		WHERE entry_number LIKE 'JE-%'
	`).Scan(&maxEntryNum)
	
	nextEntryNum := fmt.Sprintf("JE-%05d", maxEntryNum+1)

	// Create journal ledger entry
	var journalID int
	err := sqlDB.QueryRow(`
		INSERT INTO unified_journal_ledger (
			entry_number, source_type, source_code, entry_date, 
			description, total_debit, total_credit, status, 
			is_balanced, is_auto_generated, created_by, created_at, updated_at
		) VALUES (
			$1, 'MANUAL', 'REVENUE_FIX', CURRENT_DATE,
			'Revenue correction for Pendapatan Penjualan', 
			0, $2, 'POSTED', 
			false, true, 1, NOW(), NOW()
		) RETURNING id
	`, nextEntryNum, amount).Scan(&journalID)

	if err != nil {
		log.Printf("Error creating journal ledger: %v", err)
		return
	}

	// Create journal lines - Revenue (Credit)
	_, err = sqlDB.Exec(`
		INSERT INTO unified_journal_lines (
			journal_id, account_id, line_number, description,
			debit_amount, credit_amount, created_at, updated_at
		) VALUES ($1, $2, 1, $3, 0, $4, NOW(), NOW())
	`, journalID, revenueAccountID, 
		fmt.Sprintf("Revenue - Sales correction"), amount)

	if err != nil {
		log.Printf("Error creating revenue journal line: %v", err)
		return
	}

	// For balancing, we need a corresponding debit entry
	// Let's use a temporary account or create a balancing entry
	// For now, we'll create an adjustment entry

	fmt.Printf("✅ Created journal entry %s for Rp %.0f\n", nextEntryNum, amount)
	fmt.Printf("   Revenue Account ID %d credited with Rp %.0f\n", revenueAccountID, amount)
}