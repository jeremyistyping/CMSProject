package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := config.ConnectDB(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🔍 Balance Sheet Diagnostic Report")
	fmt.Println("=====================================")

	// Check 1: Verify accounts exist
	checkAccounts(db)

	// Check 2: Verify journal entries exist
	checkJournalEntries(db)

	// Check 3: Test balance sheet generation
	testBalanceSheetGeneration(db)

	// Check 4: Create sample data if needed
	createSampleDataIfNeeded(db)
}

func checkAccounts(db *gorm.DB) {
	fmt.Println("\n📊 Checking Accounts...")
	
	var totalAccounts int64
	db.Model(&models.Account{}).Count(&totalAccounts)
	fmt.Printf("Total accounts: %d\n", totalAccounts)

	// Count by type
	var assetCount, liabilityCount, equityCount int64
	db.Model(&models.Account{}).Where("type = ?", models.AccountTypeAsset).Count(&assetCount)
	db.Model(&models.Account{}).Where("type = ?", models.AccountTypeLiability).Count(&liabilityCount)
	db.Model(&models.Account{}).Where("type = ?", models.AccountTypeEquity).Count(&equityCount)

	fmt.Printf("- Asset accounts: %d\n", assetCount)
	fmt.Printf("- Liability accounts: %d\n", liabilityCount)
	fmt.Printf("- Equity accounts: %d\n", equityCount)

	if totalAccounts == 0 {
		fmt.Println("⚠️  No accounts found! Balance sheet will be empty.")
	}

	// Show some sample accounts
	var sampleAccounts []models.Account
	db.Where("type IN ?", []string{models.AccountTypeAsset, models.AccountTypeLiability, models.AccountTypeEquity}).
		Limit(10).Find(&sampleAccounts)
	
	if len(sampleAccounts) > 0 {
		fmt.Println("\n📋 Sample Balance Sheet Accounts:")
		for _, acc := range sampleAccounts {
			fmt.Printf("- %s (%s): %s [Balance: %.2f]\n", acc.Code, acc.Type, acc.Name, acc.Balance)
		}
	}
}

func checkJournalEntries(db *gorm.DB) {
	fmt.Println("\n📝 Checking Journal Entries...")
	
	var totalJournals, totalEntries int64
	db.Model(&models.Journal{}).Count(&totalJournals)
	db.Model(&models.JournalEntry{}).Count(&totalEntries)
	
	fmt.Printf("Total journals: %d\n", totalJournals)
	fmt.Printf("Total journal entries: %d\n", totalEntries)

	// Count posted journals
	var postedJournals int64
	db.Model(&models.Journal{}).Where("status = ?", "POSTED").Count(&postedJournals)
	fmt.Printf("Posted journals: %d\n", postedJournals)

	if totalJournals == 0 {
		fmt.Println("⚠️  No journal entries found! Balance sheet will show opening balances only.")
	}

	// Show recent journal entries affecting balance sheet accounts
	var recentEntries []struct {
		JournalCode    string
		JournalDate    time.Time
		AccountCode    string
		AccountName    string
		AccountType    string
		DebitAmount    float64
		CreditAmount   float64
		Status         string
	}

	db.Table("journal_entries").
		Select("j.code as journal_code, j.date as journal_date, a.code as account_code, a.name as account_name, a.type as account_type, journal_entries.debit_amount, journal_entries.credit_amount, j.status").
		Joins("JOIN journals j ON journal_entries.journal_id = j.id").
		Joins("JOIN accounts a ON journal_entries.account_id = a.id").
		Where("a.type IN ?", []string{models.AccountTypeAsset, models.AccountTypeLiability, models.AccountTypeEquity}).
		Where("j.status = ?", "POSTED").
		Order("j.date DESC").
		Limit(10).
		Find(&recentEntries)

	if len(recentEntries) > 0 {
		fmt.Println("\n📋 Recent Journal Entries (Balance Sheet Accounts):")
		for _, entry := range recentEntries {
			fmt.Printf("- %s [%s] %s (%s): Dr %.2f, Cr %.2f [%s]\n", 
				entry.JournalCode, 
				entry.JournalDate.Format("2006-01-02"), 
				entry.AccountCode, 
				entry.AccountType,
				entry.DebitAmount, 
				entry.CreditAmount,
				entry.Status)
		}
	}
}

func testBalanceSheetGeneration(db *gorm.DB) {
	fmt.Println("\n🧪 Testing Balance Sheet Generation...")

	// Create repositories
	accountRepo := repositories.NewAccountRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	purchaseRepo := repositories.NewPurchaseRepository(db)
	productRepo := repositories.NewProductRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)

	// Create service
	service := services.NewStandardizedReportService(
		db, accountRepo, salesRepo, purchaseRepo, 
		productRepo, contactRepo, paymentRepo, cashBankRepo,
	)

	// Test with current date
	asOfDate := time.Now()
	
	fmt.Printf("Generating balance sheet as of: %s\n", asOfDate.Format("2006-01-02"))
	
	response, err := service.GenerateStandardBalanceSheet(asOfDate, "json", false)
	if err != nil {
		fmt.Printf("❌ Error generating balance sheet: %v\n", err)
		return
	}

	fmt.Printf("✅ Balance sheet generated successfully!\n")
	fmt.Printf("- Total Assets: %.2f\n", response.Statement.Totals["total_assets"])
	fmt.Printf("- Total Liabilities: %.2f\n", response.Statement.Totals["total_liabilities"])
	fmt.Printf("- Total Equity: %.2f\n", response.Statement.Totals["total_equity"])
	fmt.Printf("- Sections: %d\n", len(response.Statement.Sections))

	// Check if balance sheet is empty
	hasData := false
	for _, section := range response.Statement.Sections {
		if len(section.Items) > 0 || len(section.Subsections) > 0 {
			hasData = true
			break
		}
		for _, subsection := range section.Subsections {
			if len(subsection.Items) > 0 {
				hasData = true
				break
			}
		}
	}

	if !hasData {
		fmt.Println("⚠️  Balance sheet is empty - no accounts have balances!")
		fmt.Println("This explains why the frontend shows 'No Data Available'")
	} else {
		fmt.Println("✅ Balance sheet contains data")
	}

	// Show section details
	for _, section := range response.Statement.Sections {
		fmt.Printf("\n📊 %s (Subtotal: %.2f)\n", section.Name, section.Subtotal)
		if len(section.Subsections) > 0 {
			for _, subsection := range section.Subsections {
				fmt.Printf("  └─ %s: %d items, Subtotal: %.2f\n", 
					subsection.Name, len(subsection.Items), subsection.Subtotal)
			}
		} else {
			fmt.Printf("  └─ Direct items: %d\n", len(section.Items))
		}
	}
}

func createSampleDataIfNeeded(db *gorm.DB) {
	fmt.Println("\n🛠️ Sample Data Creation...")

	// Check if we need sample data
	var accountCount int64
	db.Model(&models.Account{}).
		Where("type IN ? AND balance != 0", []string{models.AccountTypeAsset, models.AccountTypeLiability, models.AccountTypeEquity}).
		Count(&accountCount)

	if accountCount > 0 {
		fmt.Println("✅ Accounts with balances already exist, skipping sample data creation")
		return
	}

	fmt.Println("⚠️  No accounts with balances found. Creating sample data...")

	// Create sample balance sheet accounts with balances
	sampleAccounts := []models.Account{
		{
			Code:     "1-1-001",
			Name:     "Kas",
			Type:     models.AccountTypeAsset,
			Balance:  50000000, // 50 million IDR
			IsActive: true,
			Level:    3,
		},
		{
			Code:     "1-1-002",
			Name:     "Bank",
			Type:     models.AccountTypeAsset,
			Balance:  100000000, // 100 million IDR
			IsActive: true,
			Level:    3,
		},
		{
			Code:     "1-1-003",
			Name:     "Piutang Dagang",
			Type:     models.AccountTypeAsset,
			Balance:  75000000, // 75 million IDR
			IsActive: true,
			Level:    3,
		},
		{
			Code:     "1-2-001",
			Name:     "Peralatan Kantor",
			Type:     models.AccountTypeAsset,
			Balance:  25000000, // 25 million IDR
			IsActive: true,
			Level:    3,
		},
		{
			Code:     "2-1-001",
			Name:     "Utang Dagang",
			Type:     models.AccountTypeLiability,
			Balance:  40000000, // 40 million IDR
			IsActive: true,
			Level:    3,
		},
		{
			Code:     "2-1-002",
			Name:     "Utang Pajak",
			Type:     models.AccountTypeLiability,
			Balance:  10000000, // 10 million IDR
			IsActive: true,
			Level:    3,
		},
		{
			Code:     "3-1-001",
			Name:     "Modal Disetor",
			Type:     models.AccountTypeEquity,
			Balance:  150000000, // 150 million IDR
			IsActive: true,
			Level:    3,
		},
		{
			Code:     "3-2-001",
			Name:     "Laba Ditahan",
			Type:     models.AccountTypeEquity,
			Balance:  50000000, // 50 million IDR
			IsActive: true,
			Level:    3,
		},
	}

	for _, account := range sampleAccounts {
		var existingAccount models.Account
		if err := db.Where("code = ?", account.Code).First(&existingAccount).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new account
				if err := db.Create(&account).Error; err != nil {
					fmt.Printf("❌ Failed to create account %s: %v\n", account.Code, err)
				} else {
					fmt.Printf("✅ Created account: %s - %s (Balance: %.0f)\n", 
						account.Code, account.Name, account.Balance)
				}
			}
		} else {
			// Update existing account balance
			if err := db.Model(&existingAccount).Update("balance", account.Balance).Error; err != nil {
				fmt.Printf("❌ Failed to update account %s: %v\n", account.Code, err)
			} else {
				fmt.Printf("✅ Updated account: %s - %s (Balance: %.0f)\n", 
					account.Code, account.Name, account.Balance)
			}
		}
	}

	fmt.Println("\n📊 Sample data creation completed!")
	fmt.Println("🔄 Try generating the balance sheet again from the frontend.")
}