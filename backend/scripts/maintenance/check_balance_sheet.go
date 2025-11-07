package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database config from environment
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	
	// Construct DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database successfully!")

	// Check journal entries for purchases
	fmt.Println("\n=== Purchase Journal Entries ===")
	var journalEntries []models.JournalEntry
	err = db.Where("reference_type = ?", "PURCHASE").Order("created_at DESC").Find(&journalEntries).Error
	if err != nil {
		log.Fatalf("Failed to fetch journal entries: %v", err)
	}

	fmt.Printf("Found %d purchase journal entries:\n", len(journalEntries))
	for _, entry := range journalEntries {
		fmt.Printf("ID=%d, Code=%s, Status=%s, Debit=%.2f, Credit=%.2f, Ref_ID=%d, Date=%s\n", 
			entry.ID, entry.Code, entry.Status, entry.TotalDebit, entry.TotalCredit, 
			*entry.ReferenceID, entry.EntryDate.Format("2006-01-02"))
	}

	// Check balance sheet accounts
	fmt.Println("\n=== Balance Sheet - Assets ===")
	var assetAccounts []models.Account
	err = db.Where("type = ?", models.AccountTypeAsset).Order("code").Find(&assetAccounts).Error
	if err != nil {
		log.Fatalf("Failed to fetch asset accounts: %v", err)
	}

	totalAssets := 0.0
	for _, acc := range assetAccounts {
		if acc.Balance != 0 || acc.Code == "1105" { // Show non-zero balances and PPN account
			fmt.Printf("  %s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalAssets += acc.Balance
		}
	}
	fmt.Printf("TOTAL ASSETS: %.2f\n", totalAssets)

	fmt.Println("\n=== Balance Sheet - Liabilities ===")
	var liabilityAccounts []models.Account
	err = db.Where("type = ?", models.AccountTypeLiability).Order("code").Find(&liabilityAccounts).Error
	if err != nil {
		log.Fatalf("Failed to fetch liability accounts: %v", err)
	}

	totalLiabilities := 0.0
	for _, acc := range liabilityAccounts {
		if acc.Balance != 0 || acc.Code == "2101" { // Show non-zero balances and Hutang Usaha account
			fmt.Printf("  %s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalLiabilities += acc.Balance
		}
	}
	fmt.Printf("TOTAL LIABILITIES: %.2f\n", totalLiabilities)

	fmt.Println("\n=== Balance Sheet - Equity ===")
	var equityAccounts []models.Account
	err = db.Where("type = ?", models.AccountTypeEquity).Order("code").Find(&equityAccounts).Error
	if err != nil {
		log.Fatalf("Failed to fetch equity accounts: %v", err)
	}

	totalEquity := 0.0
	for _, acc := range equityAccounts {
		if acc.Balance != 0 {
			fmt.Printf("  %s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalEquity += acc.Balance
		}
	}
	fmt.Printf("TOTAL EQUITY: %.2f\n", totalEquity)

	fmt.Println("\n=== Income Statement - Expenses ===")
	var expenseAccounts []models.Account
	err = db.Where("type = ?", models.AccountTypeExpense).Order("code").Find(&expenseAccounts).Error
	if err != nil {
		log.Fatalf("Failed to fetch expense accounts: %v", err)
	}

	totalExpenses := 0.0
	for _, acc := range expenseAccounts {
		if acc.Balance != 0 {
			fmt.Printf("  %s - %s: %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalExpenses += acc.Balance
		}
	}
	fmt.Printf("TOTAL EXPENSES: %.2f\n", totalExpenses)

	// Balance check
	fmt.Println("\n=== Balance Verification ===")
	fmt.Printf("Total Assets: %.2f\n", totalAssets)
	fmt.Printf("Total Liabilities + Equity: %.2f\n", totalLiabilities + totalEquity)
	fmt.Printf("Difference: %.2f\n", totalAssets - (totalLiabilities + totalEquity))

	balanceIsCorrect := (totalAssets - (totalLiabilities + totalEquity)) < 0.01 && 
	                    (totalAssets - (totalLiabilities + totalEquity)) > -0.01
	
	if balanceIsCorrect {
		fmt.Println("✅ Balance Sheet is balanced!")
	} else {
		fmt.Println("❌ Balance Sheet is NOT balanced!")
	}

	// Check recent purchases status
	fmt.Println("\n=== Recent Purchase Status ===")
	var purchases []models.Purchase
	err = db.Preload("Vendor").Order("created_at DESC").Limit(5).Find(&purchases).Error
	if err != nil {
		log.Fatalf("Failed to fetch purchases: %v", err)
	}

	for _, purchase := range purchases {
		fmt.Printf("Purchase %s - %s: Status=%s, Approval=%s, Amount=%.2f\n", 
			purchase.Code, purchase.Vendor.Name, purchase.Status, 
			purchase.ApprovalStatus, purchase.TotalAmount)
	}

	fmt.Println("\n🎯 Analysis Summary:")
	fmt.Printf("• Found %d purchase journal entries\n", len(journalEntries))
	fmt.Printf("• Accounts Payable Balance: %.2f\n", func() float64 {
		for _, acc := range liabilityAccounts {
			if acc.Code == "2101" {
				return acc.Balance
			}
		}
		return 0.0
	}())
	fmt.Printf("• PPN Masukan Balance: %.2f\n", func() float64 {
		for _, acc := range assetAccounts {
			if acc.Code == "1105" {
				return acc.Balance
			}
		}
		return 0.0
	}())
	fmt.Printf("• Total Expense Balances: %.2f\n", totalExpenses)
	
	if balanceIsCorrect {
		fmt.Println("✅ Purchase accounting is working correctly!")
	} else {
		fmt.Println("⚠️  There might be some balance discrepancies to investigate")
	}
}
