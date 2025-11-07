package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"app-sistem-akuntansi/models"
)

// BalanceSummaryDebug holds debug information about balance calculations
type BalanceSummaryDebug struct {
	Accounts     []models.CashBank `json:"accounts"`
	TotalCash    float64           `json:"total_cash"`
	TotalBank    float64           `json:"total_bank"`
	TotalBalance float64           `json:"total_balance"`
	Individual   []AccountDebug    `json:"individual_calculations"`
}

// AccountDebug shows detailed calculation for each account
type AccountDebug struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Balance        float64 `json:"balance"`
	Currency       string  `json:"currency"`
	IsActive       bool    `json:"is_active"`
	IncludedInSum  bool    `json:"included_in_sum"`
}

func main() {
	// Read database config
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "app_sistem_akuntansi")
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASS", "root")

	// Create DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	fmt.Printf("🔍 Connecting to database: %s@%s:%s/%s\n", dbUser, dbHost, dbPort, dbName)

	// Connect to database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable SQL logging for cleaner output
	})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	fmt.Println("✅ Database connected successfully")

	// Debug cash bank balances
	debugCashBankBalances(db)
}

func debugCashBankBalances(db *gorm.DB) {
	fmt.Println("\n🔍 DEBUGGING CASH BANK BALANCES")
	fmt.Println("==================================================")

	// Get all accounts
	var accounts []models.CashBank
	if err := db.Find(&accounts).Error; err != nil {
		log.Fatal("❌ Failed to fetch accounts:", err)
	}

	fmt.Printf("📊 Found %d total accounts in database\n", len(accounts))

	// Initialize debug summary
	debug := BalanceSummaryDebug{
		Accounts:     accounts,
		TotalCash:    0,
		TotalBank:    0,
		TotalBalance: 0,
		Individual:   []AccountDebug{},
	}

	fmt.Println("\n📋 Account Details:")
	fmt.Println("ID | Name | Type | Balance | Currency | Active | Included")
	fmt.Println("-----------------------------------------------------------------")

	for _, account := range accounts {
		accountDebug := AccountDebug{
			ID:             account.ID,
			Name:           account.Name,
			Type:           account.Type,
			Balance:        account.Balance,
			Currency:       account.Currency,
			IsActive:       account.IsActive,
			IncludedInSum:  account.IsActive, // Only active accounts included in summary
		}

		// Print account details
		fmt.Printf("%2d | %-15s | %-4s | %12.2f | %3s | %5t | %5t\n",
			account.ID,
			truncate(account.Name, 15),
			account.Type,
			account.Balance,
			account.Currency,
			account.IsActive,
			accountDebug.IncludedInSum)

		// Add to summary calculation (mimicking service logic)
		if account.IsActive {
			if account.Type == models.CashBankTypeCash {
				debug.TotalCash += account.Balance
			} else if account.Type == models.CashBankTypeBank {
				debug.TotalBank += account.Balance
			}
			debug.TotalBalance += account.Balance
		}

		debug.Individual = append(debug.Individual, accountDebug)
	}

	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("💰 Total Cash Balance:  %12.2f IDR\n", debug.TotalCash)
	fmt.Printf("🏦 Total Bank Balance:  %12.2f IDR\n", debug.TotalBank)
	fmt.Printf("💸 Total Balance:       %12.2f IDR\n", debug.TotalBalance)
	fmt.Println("==================================================")

	// Check for potential double posting issues
	fmt.Println("\n🔍 DOUBLE POSTING ANALYSIS")
	fmt.Println("==================================================")

	expectedTotal := 10000000.0 // Expected total from your description
	actualTotal := debug.TotalBalance
	difference := actualTotal - expectedTotal

	fmt.Printf("Expected Total: %12.2f IDR\n", expectedTotal)
	fmt.Printf("Actual Total:   %12.2f IDR\n", actualTotal)
	fmt.Printf("Difference:     %12.2f IDR\n", difference)

	if difference > 0.01 { // Allow for small floating point differences
		fmt.Println("⚠️  POSSIBLE DOUBLE POSTING DETECTED!")
		fmt.Printf("    The actual balance (%.2f) is %.2f higher than expected (%.2f)\n", 
			actualTotal, difference, expectedTotal)
		
		// Check if any account has exactly double the expected amount
		if difference == expectedTotal {
			fmt.Println("🔍 This looks like a classic double posting issue where balances are doubled.")
		}
	} else {
		fmt.Println("✅ Balance totals look correct - no double posting detected.")
	}

	// Save debug info to file
	debugJSON, err := json.MarshalIndent(debug, "", "  ")
	if err == nil {
		if err := os.WriteFile("cash_bank_balance_debug.json", debugJSON, 0644); err == nil {
			fmt.Println("\n📄 Debug information saved to: cash_bank_balance_debug.json")
		}
	}

	// Check journal entries for the main account
	if len(accounts) > 0 {
		fmt.Println("\n🔍 CHECKING JOURNAL ENTRIES FOR ACCOUNT BNK-2025-0004")
		fmt.Println("==================================================")
		
		var targetAccount *models.CashBank
		for _, acc := range accounts {
			if acc.Code == "BNK-2025-0004" || acc.Name == "BANK UOB" {
				targetAccount = &acc
				break
			}
		}
		
		if targetAccount != nil {
			checkJournalEntries(db, targetAccount)
		} else {
			fmt.Println("❌ Account BNK-2025-0004 (BANK UOB) not found")
		}
	}
}

func checkJournalEntries(db *gorm.DB, account *models.CashBank) {
	// Find related GL account
	var glAccount models.Account
	if err := db.Where("id = ?", account.AccountID).First(&glAccount).Error; err != nil {
		fmt.Printf("❌ GL Account not found for cash bank account: %v\n", err)
		return
	}
	
	fmt.Printf("🔗 GL Account: %s - %s\n", glAccount.Code, glAccount.Name)
	
	// Count journal entries affecting this account
	var journalCount int64
	if err := db.Model(&models.JournalEntry{}).
		Where("account_id = ?", glAccount.ID).
		Count(&journalCount).Error; err != nil {
		fmt.Printf("❌ Failed to count journal entries: %v\n", err)
		return
	}
	
	fmt.Printf("📊 Total journal entries: %d\n", journalCount)
	
	// Get recent journal entries
	var recentEntries []models.JournalEntry
	if err := db.Where("account_id = ?", glAccount.ID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentEntries).Error; err != nil {
		fmt.Printf("❌ Failed to fetch recent journal entries: %v\n", err)
		return
	}
	
	fmt.Println("\n📋 Recent Journal Entries (Latest 5):")
	fmt.Println("ID | Date | Debit | Credit | Description")
	fmt.Println("------------------------------------------------------------")
	
	for _, entry := range recentEntries {
		fmt.Printf("%d | %s | %8.2f | %8.2f | %s\n",
			entry.ID,
			entry.EntryDate.Format("2006-01-02"),
			entry.TotalDebit,
			entry.TotalCredit,
			truncate(entry.Description, 20))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}