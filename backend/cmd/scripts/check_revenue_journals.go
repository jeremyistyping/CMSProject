package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/cmd/scripts/utils"
)

type RevenueJournalInfo struct {
	AccountID         int     `gorm:"column:account_id"`
	AccountName       string  `gorm:"column:account_name"`
	CurrentBalance    float64 `gorm:"column:current_balance"`
	JournalCount      int     `gorm:"column:journal_count"`
	CalculatedBalance float64 `gorm:"column:calculated_balance"`
}

type JournalEntry struct {
	ID           int     `gorm:"column:id"`
	AccountID    int     `gorm:"column:account_id"`
	DebitAmount  float64 `gorm:"column:debit_amount"`
	CreditAmount float64 `gorm:"column:credit_amount"`
	Description  string  `gorm:"column:description"`
	CreatedAt    string  `gorm:"column:created_at"`
}

func main() {
	fmt.Printf("🔍 CHECKING REVENUE JOURNAL ENTRIES...\n\n")
	
	// Load environment variables from .env file
	databaseURL, err := utils.GetDatabaseURL()
	if err != nil {
		log.Fatal(err)
	}
	
	// Print environment info (with masked sensitive data)
	utils.PrintEnvInfo()

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

	// Check revenue accounts with journal entries
	fmt.Printf("\n=== REVENUE ACCOUNTS JOURNAL SUMMARY ===\n")
	var revenueInfo []RevenueJournalInfo
	
	query := `
		SELECT 
			ujl.account_id,
			a.name as account_name,
			a.balance as current_balance,
			COUNT(ujl.id) as journal_count,
			SUM(ujl.credit_amount - ujl.debit_amount) as calculated_balance
		FROM unified_journal_lines ujl 
		LEFT JOIN accounts a ON ujl.account_id = a.id 
		WHERE a.type = 'REVENUE'
		GROUP BY ujl.account_id, a.name, a.balance 
		ORDER BY ujl.account_id
	`
	
	err = gormDB.Raw(query).Scan(&revenueInfo).Error
	if err != nil {
		log.Fatal("Failed to get revenue journal info:", err)
	}

	if len(revenueInfo) == 0 {
		fmt.Printf("❌ NO REVENUE JOURNAL ENTRIES FOUND!\n")
		fmt.Printf("This explains why revenue accounts show 0 balance.\n\n")
		
		// Check if revenue accounts exist but have no journals
		var revenueAccounts []struct {
			ID      int     `gorm:"column:id"`
			Name    string  `gorm:"column:name"`
			Balance float64 `gorm:"column:balance"`
		}
		
		err = gormDB.Raw("SELECT id, name, balance FROM accounts WHERE type = 'REVENUE' ORDER BY id").Scan(&revenueAccounts).Error
		if err != nil {
			log.Printf("Error checking revenue accounts: %v", err)
		} else {
			fmt.Printf("📋 REVENUE ACCOUNTS WITHOUT JOURNALS:\n")
			for _, acc := range revenueAccounts {
				fmt.Printf("  ID: %d | %s | Balance: Rp %.2f\n", acc.ID, acc.Name, acc.Balance)
			}
		}
	} else {
		fmt.Printf("📊 Found %d revenue accounts with journal entries:\n\n", len(revenueInfo))
		fmt.Printf("%-4s | %-30s | %15s | %8s | %15s | %s\n", "ID", "Account Name", "Current Balance", "Journals", "Calc Balance", "Status")
		fmt.Printf("%-4s-+-%30s-+-%15s-+-%8s-+-%15s-+-%s\n", "----", "------------------------------", "---------------", "--------", "---------------", "--------")
		
		for _, info := range revenueInfo {
			status := "✅ OK"
			if info.CurrentBalance != info.CalculatedBalance {
				status = "❌ MISMATCH"
			}
			
			fmt.Printf("%-4d | %-30s | %15.2f | %8d | %15.2f | %s\n", 
				info.AccountID, info.AccountName, info.CurrentBalance, 
				info.JournalCount, info.CalculatedBalance, status)
		}
	}

	// Check recent revenue journal entries
	fmt.Printf("\n=== RECENT REVENUE JOURNAL ENTRIES ===\n")
	var recentEntries []JournalEntry
	
	recentQuery := `
		SELECT ujl.id, ujl.account_id, ujl.debit_amount, ujl.credit_amount, ujl.description, ujl.created_at::text
		FROM unified_journal_lines ujl 
		JOIN accounts a ON ujl.account_id = a.id 
		WHERE a.type = 'REVENUE'
		ORDER BY ujl.created_at DESC 
		LIMIT 10
	`
	
	err = gormDB.Raw(recentQuery).Scan(&recentEntries).Error
	if err != nil {
		log.Printf("Error getting recent entries: %v", err)
	} else {
		if len(recentEntries) == 0 {
			fmt.Printf("❌ NO RECENT REVENUE JOURNAL ENTRIES FOUND!\n")
		} else {
			fmt.Printf("📝 Last %d revenue journal entries:\n\n", len(recentEntries))
			fmt.Printf("%-6s | %-4s | %12s | %12s | %-30s | %s\n", "Entry", "Acc", "Debit", "Credit", "Description", "Created")
			fmt.Printf("%-6s-+%-4s-+%12s-+%12s-+%-30s-+-%s\n", "------", "----", "------------", "------------", "------------------------------", "--------")
			
			for _, entry := range recentEntries {
				fmt.Printf("%-6d | %-4d | %12.2f | %12.2f | %-30s | %s\n",
					entry.ID, entry.AccountID, entry.DebitAmount, entry.CreditAmount, 
					truncateString(entry.Description, 30), entry.CreatedAt[:19])
			}
		}
	}

	// Check if there are any pending sales transactions
	fmt.Printf("\n=== CHECKING FOR PENDING SALES TRANSACTIONS ===\n")
	var pendingSales int
	err = gormDB.Raw("SELECT COUNT(*) FROM transactions WHERE transaction_type = 'SALE' AND status != 'COMPLETED'").Scan(&pendingSales).Error
	if err != nil {
		log.Printf("Error checking pending sales: %v", err)
	} else {
		fmt.Printf("📊 Pending sales transactions: %d\n", pendingSales)
		if pendingSales > 0 {
			fmt.Printf("⚠️  There are %d pending sales that may need to be processed to generate revenue journals.\n", pendingSales)
		}
	}

	fmt.Printf("\n🎯 ANALYSIS SUMMARY:\n")
	if len(revenueInfo) == 0 {
		fmt.Printf("❌ NO REVENUE JOURNAL ENTRIES found in unified_journal_lines\n")
		fmt.Printf("This explains why revenue accounts show 0 balance.\n")
		fmt.Printf("\n💡 NEXT STEPS:\n")
		fmt.Printf("1. Check if sales transactions exist but journals weren't created\n")
		fmt.Printf("2. Create script to generate missing revenue journal entries\n")
		fmt.Printf("3. Verify the journal creation process in the application\n")
	} else {
		fmt.Printf("✅ Found revenue journal entries, checking for balance mismatches...\n")
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}