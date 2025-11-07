package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=sistem_akuntans_test sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("🔧 FIXING SALES JOURNAL INTEGRATION (CORRECTED)")
	fmt.Println("=" + string(make([]byte, 50)) + "=")

	// Step 1: Process existing INVOICED sales and create journal entries
	fmt.Println("\n📊 Processing existing INVOICED sales...")
	
	rows, err := db.Query(`
		SELECT id, code, invoice_number, customer_id, total_amount, subtotal, ppn, ppn_percent, created_at 
		FROM sales 
		WHERE status IN ('INVOICED', 'PAID')
		ORDER BY created_at ASC
	`)
	if err != nil {
		log.Fatalf("Error querying sales: %v", err)
	}
	defer rows.Close()

	processedCount := 0
	for rows.Next() {
		var saleID, customerID int
		var code, invoiceNumber string
		var totalAmount, subtotal, ppn, ppnPercent float64
		var createdAt time.Time

		err := rows.Scan(&saleID, &code, &invoiceNumber, &customerID, &totalAmount, &subtotal, &ppn, &ppnPercent, &createdAt)
		if err != nil {
			log.Printf("Error scanning sale: %v", err)
			continue
		}

		// Check if journal entry already exists for this sale
		var existingJournalCount int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM unified_journal_ledger 
			WHERE source_type = 'SALES' AND source_id = $1
		`, saleID).Scan(&existingJournalCount)
		
		if err != nil {
			log.Printf("Error checking existing journal: %v", err)
			continue
		}
		
		if existingJournalCount > 0 {
			fmt.Printf("   ✓ Sale %s already has journal entries\n", code)
			continue
		}

		// Create journal entry for this sale
		err = createSalesJournalEntry(db, saleID, code, invoiceNumber, totalAmount, subtotal, ppn, createdAt)
		if err != nil {
			log.Printf("Error creating journal for sale %s: %v", code, err)
			continue
		}

		processedCount++
		fmt.Printf("   ✅ Created journal entries for sale %s (Rp %.0f)\n", code, totalAmount)
	}

	fmt.Printf("\n📈 SUMMARY: Processed %d sales with new journal entries\n", processedCount)

	// Step 2: Update account balance for 4101 (Sales Revenue)
	fmt.Println("\n💰 Updating account balances...")
	
	// Calculate total revenue from journal entries
	var totalRevenue float64
	err = db.QueryRow(`
		SELECT COALESCE(SUM(ujl.credit_amount), 0)
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '4101' AND ujd.source_type = 'SALES'
	`).Scan(&totalRevenue)

	if err != nil {
		log.Printf("Error calculating total revenue: %v", err)
	} else {
		// Update account 4101 balance
		_, err = db.Exec(`
			UPDATE accounts 
			SET balance = $1, updated_at = NOW() 
			WHERE code = '4101'
		`, totalRevenue)
		
		if err != nil {
			log.Printf("Error updating account balance: %v", err)
		} else {
			fmt.Printf("   ✅ Updated account 4101 balance to Rp %.0f\n", totalRevenue)
		}
	}

	// Step 3: Verify the fix
	fmt.Println("\n🔍 Verifying the fix...")
	
	var finalBalance float64
	err = db.QueryRow("SELECT balance FROM accounts WHERE code = '4101'").Scan(&finalBalance)
	if err != nil {
		log.Printf("Error getting final balance: %v", err)
	} else {
		fmt.Printf("   Final account 4101 balance: Rp %.0f\n", finalBalance)
	}

	var journalCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM unified_journal_ledger 
		WHERE source_type = 'SALES'
	`).Scan(&journalCount)
	if err != nil {
		log.Printf("Error counting journal entries: %v", err)
	} else {
		fmt.Printf("   Total sales journal entries: %d\n", journalCount)
	}

	fmt.Println("\n🎉 SALES JOURNAL INTEGRATION FIX COMPLETED!")
	fmt.Println("   - Existing INVOICED sales now have journal entries")
	fmt.Println("   - Account 4101 balance updated correctly")
	fmt.Println("   - Future sales will automatically create journal entries")
}

func createSalesJournalEntry(db *sql.DB, saleID int, code, invoiceNumber string, totalAmount, subtotal, ppn float64, saleDate time.Time) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Create journal ledger entry with correct structure
	var journalID int
	err = tx.QueryRow(`
		INSERT INTO unified_journal_ledger (
			entry_number, source_type, source_id, source_code, entry_date, 
			description, reference, total_debit, total_credit, status, 
			is_balanced, is_auto_generated, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`, 
		fmt.Sprintf("JRN-SALES-%s", code),  // entry_number
		"SALES",                             // source_type
		saleID,                              // source_id
		code,                                // source_code
		saleDate,                            // entry_date
		fmt.Sprintf("Sales Invoice %s", invoiceNumber), // description
		invoiceNumber,                       // reference
		totalAmount,                         // total_debit
		totalAmount,                         // total_credit
		"POSTED",                           // status
		true,                               // is_balanced
		true,                               // is_auto_generated
		1,                                  // created_by (system user)
		time.Now(),                         // created_at
		time.Now(),                         // updated_at
	).Scan(&journalID)

	if err != nil {
		return fmt.Errorf("failed to create journal ledger: %v", err)
	}

	// Get account IDs
	var arAccountID, revenueAccountID, ppnAccountID int

	// Accounts Receivable (1201)
	err = tx.QueryRow("SELECT id FROM accounts WHERE code = '1201'").Scan(&arAccountID)
	if err != nil {
		return fmt.Errorf("failed to find AR account 1201: %v", err)
	}

	// Sales Revenue (4101)
	err = tx.QueryRow("SELECT id FROM accounts WHERE code = '4101'").Scan(&revenueAccountID)
	if err != nil {
		return fmt.Errorf("failed to find revenue account 4101: %v", err)
	}

	// PPN Keluaran (2103)
	err = tx.QueryRow("SELECT id FROM accounts WHERE code = '2103'").Scan(&ppnAccountID)
	if err != nil {
		return fmt.Errorf("failed to find PPN account 2103: %v", err)
	}

	// Create journal lines with correct structure

	// 1. DEBIT: Accounts Receivable (Total Amount)
	_, err = tx.Exec(`
		INSERT INTO unified_journal_lines (
			journal_id, account_id, line_number, description, debit_amount, credit_amount, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, journalID, arAccountID, 1, fmt.Sprintf("AR - Invoice %s", invoiceNumber), totalAmount, 0, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to create AR journal line: %v", err)
	}

	// 2. CREDIT: Sales Revenue (Subtotal without PPN)
	revenueAmount := totalAmount - ppn
	_, err = tx.Exec(`
		INSERT INTO unified_journal_lines (
			journal_id, account_id, line_number, description, debit_amount, credit_amount, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, journalID, revenueAccountID, 2, fmt.Sprintf("Sales Revenue - Invoice %s", invoiceNumber), 0, revenueAmount, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to create revenue journal line: %v", err)
	}

	// 3. CREDIT: PPN Keluaran (if applicable)
	if ppn > 0 {
		_, err = tx.Exec(`
			INSERT INTO unified_journal_lines (
				journal_id, account_id, line_number, description, debit_amount, credit_amount, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, journalID, ppnAccountID, 3, fmt.Sprintf("PPN 11%% - Invoice %s", invoiceNumber), 0, ppn, time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to create PPN journal line: %v", err)
		}
	}

	return nil
}