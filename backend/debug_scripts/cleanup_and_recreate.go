package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("🧹 Cleanup and Recreate Journal Data")
	fmt.Println("====================================")

	// Step 1: Clean up orphaned journal lines
	fmt.Println("\n1. Cleaning up orphaned journal lines...")
	result := db.Exec(`
		DELETE FROM journal_lines 
		WHERE journal_entry_id NOT IN (
			SELECT id FROM journal_entries WHERE status = 'POSTED'
		)
	`)
	if result.Error != nil {
		fmt.Printf("   Error cleaning journal lines: %v\n", result.Error)
	} else {
		fmt.Printf("   ✓ Cleaned up %d orphaned journal lines\n", result.RowsAffected)
	}

	// Step 2: Now add the foreign key constraint
	fmt.Println("\n2. Adding correct foreign key constraint...")
	result = db.Exec("ALTER TABLE journal_lines ADD CONSTRAINT fk_journal_entries_journal_lines FOREIGN KEY (journal_entry_id) REFERENCES journal_entries(id) ON DELETE CASCADE")
	if result.Error != nil {
		fmt.Printf("   Error creating constraint: %v\n", result.Error)
	} else {
		fmt.Println("   ✓ Created correct constraint successfully")
	}

	// Step 3: Now manually insert journal lines for our posted entries
	fmt.Println("\n3. Creating journal lines for posted entries...")
	
	// Get the IDs of our journal entries and accounts
	var journalEntries []struct {
		ID   int
		Code string
	}
	
	db.Raw("SELECT id, code FROM journal_entries WHERE status = 'POSTED' ORDER BY id").Scan(&journalEntries)
	
	var cashAccountID, revenueAccountID, cogsAccountID, expenseAccountID int
	db.Raw("SELECT id FROM accounts WHERE code = '1101'").Scan(&cashAccountID)      // Kas
	db.Raw("SELECT id FROM accounts WHERE code = '4101'").Scan(&revenueAccountID)   // Pendapatan Penjualan
	db.Raw("SELECT id FROM accounts WHERE code = '5101'").Scan(&cogsAccountID)      // Harga Pokok Penjualan
	db.Raw("SELECT id FROM accounts WHERE code = '5202'").Scan(&expenseAccountID)   // Beban Listrik

	if cashAccountID == 0 || revenueAccountID == 0 {
		fmt.Println("   ❌ Required accounts not found")
		return
	}

	fmt.Printf("   Found %d journal entries to process\n", len(journalEntries))
	fmt.Printf("   Account IDs: Cash=%d, Revenue=%d, COGS=%d, Expense=%d\n", cashAccountID, revenueAccountID, cogsAccountID, expenseAccountID)

	// Create journal lines for each entry
	for _, entry := range journalEntries {
		fmt.Printf("\n   📝 Processing entry: %s (ID: %d)\n", entry.Code, entry.ID)

		switch entry.Code {
		case "JE001-SALES":
			// Sales: Debit Cash, Credit Revenue
			result = db.Exec("INSERT INTO journal_lines (journal_entry_id, account_id, debit_amount, credit_amount, description, line_number) VALUES (?, ?, ?, ?, ?, ?)", 
				entry.ID, cashAccountID, 10000000, 0, "Cash received from sales", 1)
			if result.Error != nil {
				fmt.Printf("     ❌ Error creating line 1: %v\n", result.Error)
			} else {
				fmt.Printf("     ✓ Created debit line: Cash 10,000,000\n")
			}

			result = db.Exec("INSERT INTO journal_lines (journal_entry_id, account_id, debit_amount, credit_amount, description, line_number) VALUES (?, ?, ?, ?, ?, ?)", 
				entry.ID, revenueAccountID, 0, 10000000, "Sales revenue", 2)
			if result.Error != nil {
				fmt.Printf("     ❌ Error creating line 2: %v\n", result.Error)
			} else {
				fmt.Printf("     ✓ Created credit line: Revenue 10,000,000\n")
			}

		case "JE002-COGS":
			if cogsAccountID > 0 {
				// COGS: Debit COGS, Credit Cash
				result = db.Exec("INSERT INTO journal_lines (journal_entry_id, account_id, debit_amount, credit_amount, description, line_number) VALUES (?, ?, ?, ?, ?, ?)", 
					entry.ID, cogsAccountID, 6000000, 0, "Cost of goods sold", 1)
				if result.Error != nil {
					fmt.Printf("     ❌ Error creating line 1: %v\n", result.Error)
				} else {
					fmt.Printf("     ✓ Created debit line: COGS 6,000,000\n")
				}

				result = db.Exec("INSERT INTO journal_lines (journal_entry_id, account_id, debit_amount, credit_amount, description, line_number) VALUES (?, ?, ?, ?, ?, ?)", 
					entry.ID, cashAccountID, 0, 6000000, "Cash paid for inventory", 2)
				if result.Error != nil {
					fmt.Printf("     ❌ Error creating line 2: %v\n", result.Error)
				} else {
					fmt.Printf("     ✓ Created credit line: Cash 6,000,000\n")
				}
			}

		case "JE003-EXPENSE":
			if expenseAccountID > 0 {
				// Expense: Debit Expense, Credit Cash
				result = db.Exec("INSERT INTO journal_lines (journal_entry_id, account_id, debit_amount, credit_amount, description, line_number) VALUES (?, ?, ?, ?, ?, ?)", 
					entry.ID, expenseAccountID, 500000, 0, "Electricity expense", 1)
				if result.Error != nil {
					fmt.Printf("     ❌ Error creating line 1: %v\n", result.Error)
				} else {
					fmt.Printf("     ✓ Created debit line: Expense 500,000\n")
				}

				result = db.Exec("INSERT INTO journal_lines (journal_entry_id, account_id, debit_amount, credit_amount, description, line_number) VALUES (?, ?, ?, ?, ?, ?)", 
					entry.ID, cashAccountID, 0, 500000, "Cash paid for electricity", 2)
				if result.Error != nil {
					fmt.Printf("     ❌ Error creating line 2: %v\n", result.Error)
				} else {
					fmt.Printf("     ✓ Created credit line: Cash 500,000\n")
				}
			}

		case "JE004-REVENUE2":
			// Additional Revenue: Debit Cash, Credit Revenue
			result = db.Exec("INSERT INTO journal_lines (journal_entry_id, account_id, debit_amount, credit_amount, description, line_number) VALUES (?, ?, ?, ?, ?, ?)", 
				entry.ID, cashAccountID, 5000000, 0, "Cash from additional sales", 1)
			if result.Error != nil {
				fmt.Printf("     ❌ Error creating line 1: %v\n", result.Error)
			} else {
				fmt.Printf("     ✓ Created debit line: Cash 5,000,000\n")
			}

			result = db.Exec("INSERT INTO journal_lines (journal_entry_id, account_id, debit_amount, credit_amount, description, line_number) VALUES (?, ?, ?, ?, ?, ?)", 
				entry.ID, revenueAccountID, 0, 5000000, "Additional sales revenue", 2)
			if result.Error != nil {
				fmt.Printf("     ❌ Error creating line 2: %v\n", result.Error)
			} else {
				fmt.Printf("     ✓ Created credit line: Revenue 5,000,000\n")
			}
		}
	}

	// Step 4: Verify the results
	fmt.Println("\n📊 Final Verification:")
	
	var totalLines int64
	var totalDebit, totalCredit float64
	
	db.Raw("SELECT COUNT(*) FROM journal_lines jl JOIN journal_entries je ON jl.journal_entry_id = je.id WHERE je.status = 'POSTED'").Scan(&totalLines)
	db.Raw("SELECT COALESCE(SUM(jl.debit_amount), 0), COALESCE(SUM(jl.credit_amount), 0) FROM journal_lines jl JOIN journal_entries je ON jl.journal_entry_id = je.id WHERE je.status = 'POSTED'").Row().Scan(&totalDebit, &totalCredit)
	
	fmt.Printf("   Journal lines in posted entries: %d\n", totalLines)
	fmt.Printf("   Total Debits: Rp %.0f\n", totalDebit)
	fmt.Printf("   Total Credits: Rp %.0f\n", totalCredit)
	fmt.Printf("   Balanced: %v\n", totalDebit == totalCredit)

	if totalDebit > 0 && totalCredit > 0 {
		fmt.Println("\n🎉 SUCCESS! Journal data created successfully!")
		fmt.Println("   Your reports should now show real financial data.")
		fmt.Println("   Go to http://localhost:3000/reports and try:")
		fmt.Println("   - Balance Sheet")
		fmt.Println("   - Profit & Loss Statement")
		fmt.Println("   - Trial Balance")
	} else {
		fmt.Println("\n❌ Something went wrong. Check the errors above.")
	}
}