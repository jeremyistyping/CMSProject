package main

import (
	"fmt"
	"log"
	"time"

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

	fmt.Println("🚀 Creating Sample Financial Data for Reports")
	fmt.Println("==============================================")

	// Create sample journal entries directly
	createSampleJournalEntries(db)

	fmt.Println("\n✅ Sample data created successfully!")
	fmt.Println("\n📊 Now you can test the reports with actual data:")
	fmt.Println("   - Navigate to http://localhost:3000/reports")
	fmt.Println("   - Try generating Balance Sheet or Profit & Loss reports")
	fmt.Println("   - You should now see real financial data!")
}

func createSampleJournalEntries(db *gorm.DB) {
	fmt.Println("\n📝 Creating Sample Journal Entries...")

	// Get necessary accounts - using exact codes from database
	var cashAccount, revenueAccount, expenseAccount, cogsAccount models.Account
	
	db.Where("code = ?", "1101").First(&cashAccount)  // Kas
	db.Where("code = ?", "4101").First(&revenueAccount)  // Pendapatan Penjualan
	db.Where("code = ?", "5101").First(&cogsAccount)  // Harga Pokok Penjualan
	db.Where("code = ?", "5202").First(&expenseAccount)  // Beban Listrik

	if cashAccount.ID == 0 || revenueAccount.ID == 0 {
		fmt.Println("   ⚠️  Warning: Required accounts (1101, 4101) not found")
		// Show available accounts
		var accounts []models.Account
		db.Where("is_active = ?", true).Limit(10).Find(&accounts)
		fmt.Println("   Available accounts:")
		for _, acc := range accounts {
			fmt.Printf("     %s - %s\n", acc.Code, acc.Name)
		}
		return
	}

	fmt.Printf("   Found accounts: Cash=%s, Revenue=%s\n", cashAccount.Code, revenueAccount.Code)

	// Journal Entry 1: Sales Revenue
	journalEntry1 := models.JournalEntry{
		Code:        "JE001-SALES",
		EntryDate:   time.Now().AddDate(0, 0, -30),
		Description: "Monthly sales revenue",
		Status:      models.JournalStatusPosted,
		TotalDebit:  10000000,
		TotalCredit: 10000000,
		IsBalanced:  true,
	}
	
	result := db.Create(&journalEntry1)
	if result.Error != nil {
		fmt.Printf("   Error creating journal entry 1: %v\n", result.Error)
		return
	}

	// Journal lines for sales
	lines1 := []models.JournalLine{
		{
			JournalEntryID: journalEntry1.ID,
			AccountID:      cashAccount.ID,
			DebitAmount:    10000000,
			CreditAmount:   0,
			Description:    "Cash received from sales",
		},
		{
			JournalEntryID: journalEntry1.ID,
			AccountID:      revenueAccount.ID,
			DebitAmount:    0,
			CreditAmount:   10000000,
			Description:    "Sales revenue",
		},
	}
	
	for _, line := range lines1 {
		db.Create(&line)
	}
	
	fmt.Printf("   ✓ Created sales entry: %s (Debit: %.0f)\n", journalEntry1.Code, journalEntry1.TotalDebit)

	// Journal Entry 2: Cost of Goods Sold (if COGS account exists)
	if cogsAccount.ID > 0 {
		journalEntry2 := models.JournalEntry{
			Code:        "JE002-COGS",
			EntryDate:   time.Now().AddDate(0, 0, -28),
			Description: "Cost of goods sold",
			Status:      models.JournalStatusPosted,
			TotalDebit:  6000000,
			TotalCredit: 6000000,
			IsBalanced:  true,
		}
		
		result := db.Create(&journalEntry2)
		if result.Error != nil {
			fmt.Printf("   Error creating journal entry 2: %v\n", result.Error)
		} else {
			// Journal lines for COGS
			lines2 := []models.JournalLine{
				{
					JournalEntryID: journalEntry2.ID,
					AccountID:      cogsAccount.ID,
					DebitAmount:    6000000,
					CreditAmount:   0,
					Description:    "Cost of goods sold",
				},
				{
					JournalEntryID: journalEntry2.ID,
					AccountID:      cashAccount.ID,
					DebitAmount:    0,
					CreditAmount:   6000000,
					Description:    "Cash paid for inventory",
				},
			}
			
			for _, line := range lines2 {
				db.Create(&line)
			}
			
			fmt.Printf("   ✓ Created COGS entry: %s (Debit: %.0f)\n", journalEntry2.Code, journalEntry2.TotalDebit)
		}
	}

	// Journal Entry 3: Operating Expense (if expense account exists)
	if expenseAccount.ID > 0 {
		journalEntry3 := models.JournalEntry{
			Code:        "JE003-EXPENSE",
			EntryDate:   time.Now().AddDate(0, 0, -15),
			Description: "Monthly operating expenses",
			Status:      models.JournalStatusPosted,
			TotalDebit:  500000,
			TotalCredit: 500000,
			IsBalanced:  true,
		}
		
		result := db.Create(&journalEntry3)
		if result.Error != nil {
			fmt.Printf("   Error creating journal entry 3: %v\n", result.Error)
		} else {
			// Journal lines for expense
			lines3 := []models.JournalLine{
				{
					JournalEntryID: journalEntry3.ID,
					AccountID:      expenseAccount.ID,
					DebitAmount:    500000,
					CreditAmount:   0,
					Description:    "Electricity expense",
				},
				{
					JournalEntryID: journalEntry3.ID,
					AccountID:      cashAccount.ID,
					DebitAmount:    0,
					CreditAmount:   500000,
					Description:    "Cash paid for electricity",
				},
			}
			
			for _, line := range lines3 {
				db.Create(&line)
			}
			
			fmt.Printf("   ✓ Created expense entry: %s (Debit: %.0f)\n", journalEntry3.Code, journalEntry3.TotalDebit)
		}
	}

	// Journal Entry 4: Additional revenue for better balance sheet
	journalEntry4 := models.JournalEntry{
		Code:        "JE004-REVENUE2",
		EntryDate:   time.Now().AddDate(0, 0, -10),
		Description: "Additional sales revenue",
		Status:      models.JournalStatusPosted,
		TotalDebit:  5000000,
		TotalCredit: 5000000,
		IsBalanced:  true,
	}
	
	result = db.Create(&journalEntry4)
	if result.Error != nil {
		fmt.Printf("   Error creating journal entry 4: %v\n", result.Error)
	} else {
		// Journal lines for additional revenue
		lines4 := []models.JournalLine{
			{
				JournalEntryID: journalEntry4.ID,
				AccountID:      cashAccount.ID,
				DebitAmount:    5000000,
				CreditAmount:   0,
				Description:    "Cash from additional sales",
			},
			{
				JournalEntryID: journalEntry4.ID,
				AccountID:      revenueAccount.ID,
				DebitAmount:    0,
				CreditAmount:   5000000,
				Description:    "Additional sales revenue",
			},
		}
		
		for _, line := range lines4 {
			db.Create(&line)
		}
		
		fmt.Printf("   ✓ Created additional revenue: %s (Debit: %.0f)\n", journalEntry4.Code, journalEntry4.TotalDebit)
	}

	fmt.Println("\n📊 Summary of created journal entries:")
	
	// Check totals
	var totalEntries int64
	db.Model(&models.JournalEntry{}).Where("status = ?", models.JournalStatusPosted).Count(&totalEntries)
	
	var totalDebit, totalCredit float64
	db.Table("journal_lines").
		Joins("JOIN journal_entries ON journal_lines.journal_entry_id = journal_entries.id").
		Where("journal_entries.status = ?", models.JournalStatusPosted).
		Select("COALESCE(SUM(debit_amount), 0) as total_debit, COALESCE(SUM(credit_amount), 0) as total_credit").
		Row().Scan(&totalDebit, &totalCredit)
	
	fmt.Printf("   Posted Journal Entries: %d\n", totalEntries)
	fmt.Printf("   Total Debits: Rp %.0f\n", totalDebit)
	fmt.Printf("   Total Credits: Rp %.0f\n", totalCredit)
	fmt.Printf("   Balanced: %v\n", totalDebit == totalCredit)
}