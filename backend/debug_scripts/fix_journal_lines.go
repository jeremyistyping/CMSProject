package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("🔧 Fixing Journal Lines and Balances")
	fmt.Println("===================================")

	// First, let's see what journal entries exist
	var entries []models.JournalEntry
	db.Where("status = ?", models.JournalStatusPosted).Find(&entries)

	if len(entries) == 0 {
		fmt.Println("   ⚠️  No posted journal entries found!")
		return
	}

	fmt.Printf("   Found %d posted journal entries\n", len(entries))

	// Get accounts
	var cashAccount, revenueAccount, cogsAccount, expenseAccount models.Account
	db.Where("code = ?", "1101").First(&cashAccount)  // Kas
	db.Where("code = ?", "4101").First(&revenueAccount)  // Pendapatan Penjualan
	db.Where("code = ?", "5101").First(&cogsAccount)  // Harga Pokok Penjualan
	db.Where("code = ?", "5202").First(&expenseAccount)  // Beban Listrik

	if cashAccount.ID == 0 || revenueAccount.ID == 0 {
		fmt.Println("   ⚠️  Required accounts not found")
		return
	}

	// Clear existing journal lines for these entries to avoid duplicates
	for _, entry := range entries {
		db.Where("journal_entry_id = ?", entry.ID).Delete(&models.JournalLine{})
	}

	// Now create journal lines for each entry
	for _, entry := range entries {
		fmt.Printf("\n   📝 Creating lines for entry: %s\n", entry.Code)

		switch entry.Code {
		case "JE001-SALES":
			// Sales transaction
			lines := []models.JournalLine{
				{
					JournalEntryID: entry.ID,
					AccountID:      cashAccount.ID,
					DebitAmount:    entry.TotalDebit,
					CreditAmount:   0,
					Description:    "Cash received from sales",
					LineNumber:     1,
				},
				{
					JournalEntryID: entry.ID,
					AccountID:      revenueAccount.ID,
					DebitAmount:    0,
					CreditAmount:   entry.TotalCredit,
					Description:    "Sales revenue",
					LineNumber:     2,
				},
			}
			createJournalLines(db, lines)

		case "JE002-COGS":
			// COGS transaction
			if cogsAccount.ID > 0 {
				lines := []models.JournalLine{
					{
						JournalEntryID: entry.ID,
						AccountID:      cogsAccount.ID,
						DebitAmount:    entry.TotalDebit,
						CreditAmount:   0,
						Description:    "Cost of goods sold",
						LineNumber:     1,
					},
					{
						JournalEntryID: entry.ID,
						AccountID:      cashAccount.ID,
						DebitAmount:    0,
						CreditAmount:   entry.TotalCredit,
						Description:    "Cash paid for inventory",
						LineNumber:     2,
					},
				}
				createJournalLines(db, lines)
			}

		case "JE003-EXPENSE":
			// Expense transaction
			if expenseAccount.ID > 0 {
				lines := []models.JournalLine{
					{
						JournalEntryID: entry.ID,
						AccountID:      expenseAccount.ID,
						DebitAmount:    entry.TotalDebit,
						CreditAmount:   0,
						Description:    "Electricity expense",
						LineNumber:     1,
					},
					{
						JournalEntryID: entry.ID,
						AccountID:      cashAccount.ID,
						DebitAmount:    0,
						CreditAmount:   entry.TotalCredit,
						Description:    "Cash paid for electricity",
						LineNumber:     2,
					},
				}
				createJournalLines(db, lines)
			}

		case "JE004-REVENUE2":
			// Additional revenue
			lines := []models.JournalLine{
				{
					JournalEntryID: entry.ID,
					AccountID:      cashAccount.ID,
					DebitAmount:    entry.TotalDebit,
					CreditAmount:   0,
					Description:    "Cash from additional sales",
					LineNumber:     1,
				},
				{
					JournalEntryID: entry.ID,
					AccountID:      revenueAccount.ID,
					DebitAmount:    0,
					CreditAmount:   entry.TotalCredit,
					Description:    "Additional sales revenue",
					LineNumber:     2,
				},
			}
			createJournalLines(db, lines)
		}
	}

	// Verify the results
	fmt.Println("\n📊 Verification:")
	
	var totalLines int64
	db.Table("journal_lines").
		Joins("JOIN journal_entries ON journal_lines.journal_entry_id = journal_entries.id").
		Where("journal_entries.status = ?", models.JournalStatusPosted).
		Count(&totalLines)
	
	var totalDebit, totalCredit float64
	db.Table("journal_lines").
		Joins("JOIN journal_entries ON journal_lines.journal_entry_id = journal_entries.id").
		Where("journal_entries.status = ?", models.JournalStatusPosted).
		Select("COALESCE(SUM(debit_amount), 0) as total_debit, COALESCE(SUM(credit_amount), 0) as total_credit").
		Row().Scan(&totalDebit, &totalCredit)
	
	fmt.Printf("   Journal lines in posted entries: %d\n", totalLines)
	fmt.Printf("   Total Debits: Rp %.0f\n", totalDebit)
	fmt.Printf("   Total Credits: Rp %.0f\n", totalCredit)
	fmt.Printf("   Balanced: %v\n", totalDebit == totalCredit)

	if totalDebit > 0 {
		fmt.Println("\n✅ Success! Journal lines created successfully.")
		fmt.Println("   Now your reports should show actual financial data!")
		fmt.Println("   Try the Balance Sheet and Profit & Loss reports.")
	} else {
		fmt.Println("\n❌ Issue: Journal lines were not created properly.")
	}
}

func createJournalLines(db *gorm.DB, lines []models.JournalLine) {
	for _, line := range lines {
		result := db.Create(&line)
		if result.Error != nil {
			fmt.Printf("     ❌ Error creating line: %v\n", result.Error)
		} else {
			fmt.Printf("     ✓ Created line: %.0f/%.0f - %s\n", line.DebitAmount, line.CreditAmount, line.Description)
		}
	}
}