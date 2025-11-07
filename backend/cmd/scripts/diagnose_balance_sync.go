package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/cmd/scripts/utils"
)

type BalanceDetail struct {
	AccountID        int     `gorm:"column:account_id"`
	AccountName      string  `gorm:"column:account_name"`
	AccountType      string  `gorm:"column:account_type"`
	CurrentBalance   float64 `gorm:"column:current_balance"`
	TotalDebits      float64 `gorm:"column:total_debits"`
	TotalCredits     float64 `gorm:"column:total_credits"`
	NetAmount        float64 `gorm:"column:net_amount"`
	ExpectedBalance  float64 `gorm:"column:expected_balance"`
}

func main() {
	fmt.Printf("🔍 COMPREHENSIVE BALANCE SYNC DIAGNOSIS...\n\n")
	
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

	// Step 1: Get detailed balance analysis for all accounts with journals
	fmt.Printf("\n=== DETAILED BALANCE ANALYSIS ===\n")
	
	var balanceDetails []BalanceDetail
	query := `
		SELECT 
			a.id as account_id,
			a.name as account_name,
			a.type as account_type,
			a.balance as current_balance,
			COALESCE(SUM(ujl.debit_amount), 0) as total_debits,
			COALESCE(SUM(ujl.credit_amount), 0) as total_credits,
			COALESCE(SUM(ujl.credit_amount - ujl.debit_amount), 0) as net_amount,
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN COALESCE(SUM(ujl.debit_amount - ujl.credit_amount), 0)
				WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN COALESCE(SUM(ujl.credit_amount - ujl.debit_amount), 0)
				ELSE 0
			END as expected_balance
		FROM accounts a 
		LEFT JOIN unified_journal_lines ujl ON a.id = ujl.account_id
		WHERE EXISTS (SELECT 1 FROM unified_journal_lines WHERE account_id = a.id)
		GROUP BY a.id, a.name, a.type, a.balance 
		ORDER BY a.type, a.id
	`
	
	err = gormDB.Raw(query).Scan(&balanceDetails).Error
	if err != nil {
		log.Fatal("Failed to get balance details:", err)
	}

	fmt.Printf("📊 Accounts with journal entries:\n\n")
	fmt.Printf("%-4s | %-25s | %-8s | %12s | %12s | %12s | %12s | %12s | %s\n", 
		"ID", "Account Name", "Type", "Current Bal", "Tot Debits", "Tot Credits", "Net Amount", "Expected", "Status")
	fmt.Printf("%-4s-+%-25s-+%-8s-+%12s-+%12s-+%12s-+%12s-+%12s-+-%s\n", 
		"----", "-------------------------", "--------", "------------", "------------", "------------", "------------", "------------", "--------")
	
	var mismatchedAccounts []BalanceDetail
	
	for _, detail := range balanceDetails {
		status := "✅ OK"
		if detail.CurrentBalance != detail.ExpectedBalance {
			status = "❌ WRONG"
			mismatchedAccounts = append(mismatchedAccounts, detail)
		}
		
		fmt.Printf("%-4d | %-25s | %-8s | %12.2f | %12.2f | %12.2f | %12.2f | %12.2f | %s\n",
			detail.AccountID, detail.AccountName, detail.AccountType,
			detail.CurrentBalance, detail.TotalDebits, detail.TotalCredits,
			detail.NetAmount, detail.ExpectedBalance, status)
	}

	if len(mismatchedAccounts) == 0 {
		fmt.Printf("\n✅ All account balances are correct!\n")
		return
	}

	// Step 2: Show detailed journal entries for mismatched accounts
	fmt.Printf("\n=== JOURNAL ENTRIES FOR MISMATCHED ACCOUNTS ===\n")
	
	for _, acc := range mismatchedAccounts {
		fmt.Printf("\n🔍 Account %d (%s) - %s:\n", acc.AccountID, acc.AccountName, acc.AccountType)
		
		rows, err := sqlDB.Query(`
			SELECT id, debit_amount, credit_amount, description, created_at
			FROM unified_journal_lines 
			WHERE account_id = $1 
			ORDER BY created_at
		`, acc.AccountID)
		
		if err != nil {
			log.Printf("Error getting journals for account %d: %v", acc.AccountID, err)
			continue
		}
		
		fmt.Printf("   %-6s | %12s | %12s | %-40s | %s\n", "Entry", "Debit", "Credit", "Description", "Date")
		fmt.Printf("   %-6s-+%12s-+%12s-+%-40s-+-%s\n", "------", "------------", "------------", "----------------------------------------", "--------")
		
		for rows.Next() {
			var id int
			var debitAmount, creditAmount float64
			var description, createdAt string
			
			rows.Scan(&id, &debitAmount, &creditAmount, &description, &createdAt)
			
			fmt.Printf("   %-6d | %12.2f | %12.2f | %-40s | %s\n",
				id, debitAmount, creditAmount, 
				truncateString(description, 40), createdAt[:19])
		}
		rows.Close()
	}

	// Step 3: Test what the sync function would do
	fmt.Printf("\n=== TESTING BALANCE SYNC FUNCTION ===\n")
	
	// Test the sync function on specific accounts
	for _, acc := range mismatchedAccounts {
		fmt.Printf("🔧 Testing sync for account %d (%s):\n", acc.AccountID, acc.AccountName)
		
		// Get what the sync function calculates
		var calculatedBalance float64
		err := sqlDB.QueryRow(`
			SELECT 
				CASE 
					WHEN $2 IN ('ASSET', 'EXPENSE') THEN COALESCE(SUM(debit_amount - credit_amount), 0)
					WHEN $2 IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN COALESCE(SUM(credit_amount - debit_amount), 0)
					ELSE 0
				END
			FROM unified_journal_lines 
			WHERE account_id = $1
		`, acc.AccountID, acc.AccountType).Scan(&calculatedBalance)
		
		if err != nil {
			log.Printf("   Error calculating balance: %v", err)
			continue
		}
		
		fmt.Printf("   Current Balance: Rp %.2f\n", acc.CurrentBalance)
		fmt.Printf("   Calculated Balance: Rp %.2f\n", calculatedBalance)
		fmt.Printf("   Difference: Rp %.2f\n", calculatedBalance - acc.CurrentBalance)
		
		if calculatedBalance != acc.CurrentBalance {
			fmt.Printf("   Action: Would update to Rp %.2f\n", calculatedBalance)
		} else {
			fmt.Printf("   Action: No change needed\n")
		}
	}

	// Step 4: Show the balance sync trigger status
	fmt.Printf("\n=== BALANCE SYNC TRIGGER STATUS ===\n")
	
	var triggerExists bool
	err = sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.triggers 
			WHERE trigger_name = 'balance_sync_trigger'
		)
	`).Scan(&triggerExists)
	
	if err != nil {
		log.Printf("Error checking trigger: %v", err)
	} else {
		if triggerExists {
			fmt.Printf("✅ Balance sync trigger is installed and active\n")
		} else {
			fmt.Printf("❌ Balance sync trigger is missing\n")
		}
	}

	// Final recommendations
	fmt.Printf("\n💡 RECOMMENDATIONS:\n")
	fmt.Printf("1. The balance sync trigger might be working incorrectly\n")
	fmt.Printf("2. Consider temporarily disabling the trigger while fixing balances\n") 
	fmt.Printf("3. Run manual sync after fixing to ensure consistency\n")
	
	fmt.Printf("\nTo disable trigger temporarily:\n")
	fmt.Printf("   ALTER TABLE unified_journal_lines DISABLE TRIGGER balance_sync_trigger;\n")
	fmt.Printf("\nTo re-enable trigger:\n")
	fmt.Printf("   ALTER TABLE unified_journal_lines ENABLE TRIGGER balance_sync_trigger;\n")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}