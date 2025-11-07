package main

import (
	"fmt"
	"log"
	"strings"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
)

func main() {
	// Initialize database connection
	dsn := "postgres://postgres:password@localhost:5432/accounting_system?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔍 Diagnosing Chart of Accounts Balance Issues")
	fmt.Println(strings.Repeat("=", 60))

	// 1. Check AccountResolver mappings
	resolver := services.NewAccountResolver(db)
	
	fmt.Println("\n1. Testing AccountResolver Mappings:")
	testAccountResolver(resolver)

	// 2. Check actual account balances vs expected
	fmt.Println("\n2. Analyzing Account Balances:")
	analyzeAccountBalances(db)

	// 3. Check recent journal entries
	fmt.Println("\n3. Recent Journal Entries Analysis:")
	analyzeRecentJournalEntries(db)

	// 4. Check SSOT entries
	fmt.Println("\n4. SSOT Journal Entries Analysis:")
	analyzeSSOTEntries(db)

	fmt.Println("\n✅ Diagnosis complete. See analysis above.")
}

func testAccountResolver(resolver *services.AccountResolver) {
	testCases := []services.AccountType{
		services.AccountTypeAccountsReceivable,
		services.AccountTypeSalesRevenue,
		services.AccountTypeBank,
		services.AccountTypeCash,
		services.AccountTypePPNPayable,
	}

	for _, accountType := range testCases {
		account, err := resolver.GetAccount(accountType)
		if err != nil {
			fmt.Printf("❌ %s: ERROR - %v\n", accountType, err)
		} else {
			fmt.Printf("✅ %s: %s (%s) - Balance: %.2f\n", 
				accountType, account.Name, account.Code, account.Balance)
		}
	}
}

func analyzeAccountBalances(db *gorm.DB) {
	// Key accounts to check
	accountCodes := []string{"1201", "4101", "1102", "1104"} // AR, Sales Revenue, Bank accounts
	
	for _, code := range accountCodes {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			fmt.Printf("❌ Account %s: Not found\n", code)
			continue
		}

		// Check if balance makes sense for account type
		normalBalance := account.GetNormalBalance()
		balanceStatus := "✅ CORRECT"
		
		if normalBalance == models.NormalBalanceDebit {
			if account.Balance < 0 {
				balanceStatus = "❌ INCORRECT (should be positive for debit normal balance)"
			}
		} else {
			if account.Balance > 0 {
				balanceStatus = "❌ INCORRECT (should be negative for credit normal balance)"
			}
		}

		fmt.Printf("Account %s (%s): Balance %.2f, Type: %s, Normal: %s - %s\n",
			code, account.Name, account.Balance, account.Type, normalBalance, balanceStatus)
	}
}

func analyzeRecentJournalEntries(db *gorm.DB) {
	var entries []models.JournalEntry
	err := db.Preload("JournalLines.Account").
		Where("reference_type = ? AND created_at > NOW() - INTERVAL '7 days'", 
			models.JournalRefSale).
		Order("created_at DESC").
		Limit(5).
		Find(&entries).Error

	if err != nil {
		fmt.Printf("❌ Error fetching journal entries: %v\n", err)
		return
	}

	fmt.Printf("Found %d recent sales journal entries:\n", len(entries))
	for _, entry := range entries {
		fmt.Printf("\nEntry ID %d - %s (Status: %s)\n", entry.ID, entry.Description, entry.Status)
		fmt.Printf("  Total Debit: %.2f, Total Credit: %.2f, Balanced: %v\n", 
			entry.TotalDebit, entry.TotalCredit, entry.IsBalanced)
		
		for _, line := range entry.JournalLines {
			fmt.Printf("  - %s (%s): Debit %.2f, Credit %.2f\n",
				line.Account.Name, line.Account.Code, line.DebitAmount, line.CreditAmount)
		}
	}
}

func analyzeSSOTEntries(db *gorm.DB) {
	var ssotEntries []models.SSOTJournalEntry
	err := db.Preload("Lines.Account").
		Where("created_at > NOW() - INTERVAL '7 days'").
		Order("created_at DESC").
		Limit(5).
		Find(&ssotEntries).Error

	if err != nil {
		fmt.Printf("❌ Error fetching SSOT entries: %v\n", err)
		return
	}

	fmt.Printf("Found %d recent SSOT journal entries:\n", len(ssotEntries))
	for _, entry := range ssotEntries {
		totalDebit, _ := entry.TotalDebit.Float64()
		totalCredit, _ := entry.TotalCredit.Float64()
		fmt.Printf("\nSSOT Entry ID %d - %s (Status: %s)\n", 
			entry.ID, entry.Description, entry.Status)
		fmt.Printf("  Total Debit: %.2f, Total Credit: %.2f\n", totalDebit, totalCredit)
		
		for _, line := range entry.Lines {
			debitAmount, _ := line.DebitAmount.Float64()
			creditAmount, _ := line.CreditAmount.Float64()
			fmt.Printf("  - %s (%s): Debit %.2f, Credit %.2f\n",
				line.Account.Name, line.Account.Code, debitAmount, creditAmount)
		}
	}
}
