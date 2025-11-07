package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database using DATABASE_URL from .env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get underlying sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	fmt.Printf("📊 Auditing balance consistency across all accounts...\n\n")

	// Query to compare accounts table balance vs calculated balance from unified_journal_lines
	query := `
		WITH calculated_balances AS (
			SELECT 
				ujl.account_id,
				COALESCE(SUM(CAST(ujl.debit_amount AS DECIMAL)), 0) as calc_debit,
				COALESCE(SUM(CAST(ujl.credit_amount AS DECIMAL)), 0) as calc_credit,
				COALESCE(SUM(CAST(ujl.debit_amount AS DECIMAL)), 0) - COALESCE(SUM(CAST(ujl.credit_amount AS DECIMAL)), 0) as calc_balance
			FROM unified_journal_lines ujl
			GROUP BY ujl.account_id
		)
		SELECT 
			a.id,
			a.code,
			a.name,
			a.type,
			CAST(a.balance AS DECIMAL) as stored_balance,
			COALESCE(cb.calc_balance, 0) as calculated_balance,
			CAST(a.balance AS DECIMAL) - COALESCE(cb.calc_balance, 0) as difference,
			CASE 
				WHEN CAST(a.balance AS DECIMAL) != COALESCE(cb.calc_balance, 0) THEN '❌ MISMATCH'
				ELSE '✅ OK'
			END as status
		FROM accounts a
		LEFT JOIN calculated_balances cb ON a.id = cb.account_id
		WHERE a.is_active = true
		ORDER BY ABS(CAST(a.balance AS DECIMAL) - COALESCE(cb.calc_balance, 0)) DESC;
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		log.Fatalf("Failed to audit balances: %v", err)
	}
	defer rows.Close()

	var mismatchCount, totalAccounts int
	var criticalMismatches []map[string]interface{}

	fmt.Printf("%-4s %-10s %-25s %-12s %12s %12s %12s %s\n", 
		"ID", "Code", "Name", "Type", "Stored", "Calculated", "Difference", "Status")
	fmt.Println(strings.Repeat("-", 95))

	for rows.Next() {
		var id int
		var code, name, accountType, status string
		var storedBalance, calculatedBalance, difference float64

		err := rows.Scan(&id, &code, &name, &accountType, &storedBalance, &calculatedBalance, &difference, &status)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		totalAccounts++
		
		// Only show accounts with mismatches or significant balances
		if difference != 0 || storedBalance != 0 || calculatedBalance != 0 {
			fmt.Printf("%-4d %-10s %-25s %-12s %12.0f %12.0f %12.0f %s\n", 
				id, code, truncateString(name, 25), accountType, 
				storedBalance, calculatedBalance, difference, status)
		}

		if difference != 0 {
			mismatchCount++
			if storedBalance >= 1000000 || calculatedBalance >= 1000000 || difference >= 1000000 {
				criticalMismatches = append(criticalMismatches, map[string]interface{}{
					"id": id,
					"name": name,
					"stored": storedBalance,
					"calculated": calculatedBalance,
					"difference": difference,
				})
			}
		}
	}

	fmt.Printf("\n=== AUDIT SUMMARY ===\n")
	fmt.Printf("Total Active Accounts: %d\n", totalAccounts)
	fmt.Printf("Accounts with Balance Mismatches: %d\n", mismatchCount)
	fmt.Printf("Data Consistency Rate: %.1f%%\n", float64(totalAccounts-mismatchCount)/float64(totalAccounts)*100)

	if len(criticalMismatches) > 0 {
		fmt.Printf("\n🚨 CRITICAL MISMATCHES (> Rp 1,000,000):\n")
		for _, mismatch := range criticalMismatches {
			fmt.Printf("  - ID %d: %s\n", mismatch["id"], mismatch["name"])
			fmt.Printf("    Stored: Rp %.0f | Calculated: Rp %.0f | Diff: Rp %.0f\n", 
				mismatch["stored"], mismatch["calculated"], mismatch["difference"])
		}
	}

	// Check if the system is using both journal systems
	var regularJournalCount, unifiedJournalCount int
	sqlDB.QueryRow("SELECT COUNT(*) FROM journal_lines").Scan(&regularJournalCount)
	sqlDB.QueryRow("SELECT COUNT(*) FROM unified_journal_lines").Scan(&unifiedJournalCount)

	fmt.Printf("\n=== JOURNAL SYSTEM USAGE ===\n")
	fmt.Printf("Regular journal_lines entries: %d\n", regularJournalCount)
	fmt.Printf("Unified journal_lines entries: %d\n", unifiedJournalCount)

	if regularJournalCount > 0 && unifiedJournalCount > 0 {
		fmt.Printf("⚠️  WARNING: System is using BOTH journal systems!\n")
		fmt.Printf("   This dual-system architecture is likely the root cause of balance inconsistencies.\n")
	} else if unifiedJournalCount > 0 {
		fmt.Printf("✅ System is primarily using unified_journal_lines (recommended)\n")
	} else if regularJournalCount > 0 {
		fmt.Printf("ℹ️  System is using regular journal_lines only\n")
	}

	if mismatchCount > 0 {
		fmt.Printf("\n🔧 RECOMMENDATION:\n")
		fmt.Printf("   Run balance recalculation script for all mismatched accounts.\n")
		fmt.Printf("   Consider implementing automatic balance sync triggers.\n")
	} else {
		fmt.Printf("\n✅ All account balances are consistent!\n")
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}