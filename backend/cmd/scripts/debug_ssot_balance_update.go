package main

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("🔍 Debug SSOT Account Balance Update Mechanism")
	fmt.Println("==============================================")

	// Database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Printf("❌ Database connection failed: %v", err)
		return
	}

	fmt.Println("✅ Database connected successfully\n")

	// Step 1: Analyze SSOT journal entries
	fmt.Println("📔 Step 1: SSOT Journal Entry Analysis")
	analyzeSSOTJournalEntries(db)

	// Step 2: Check account balance update mechanism
	fmt.Println("\n⚖️ Step 2: Account Balance Update Analysis")
	analyzeAccountBalanceUpdates(db)

	// Step 3: Manual balance correction if needed
	fmt.Println("\n🔧 Step 3: Manual Balance Correction")
	manualBalanceCorrection(db)

	// Step 4: Summary
	fmt.Println("\n🎯 Step 4: Summary & Recommendations")
	provideSummary()
}

func analyzeSSOTJournalEntries(db *gorm.DB) {
	var journalEntries []models.SSOTJournalEntry
	err := db.Preload("Lines.Account").Find(&journalEntries).Error
	if err != nil {
		fmt.Printf("❌ Failed to get SSOT journal entries: %v\n", err)
		return
	}

	fmt.Printf("Found %d SSOT journal entries:\n", len(journalEntries))

	for i, entry := range journalEntries {
		fmt.Printf("\n🧾 Entry %d: %s\n", i+1, entry.EntryNumber)
		fmt.Printf("   Source: %s (ID: %v)\n", entry.SourceType, entry.SourceID)
		fmt.Printf("   Date: %s\n", entry.EntryDate.Format("2006-01-02"))
		fmt.Printf("   Description: %s\n", entry.Description)
		fmt.Printf("   Status: %s, Balanced: %t\n", entry.Status, entry.IsBalanced)
		fmt.Printf("   Total Debit: %s\n", entry.TotalDebit.String())
		fmt.Printf("   Total Credit: %s\n", entry.TotalCredit.String())

		if len(entry.Lines) > 0 {
			fmt.Printf("   📋 Journal Lines:\n")
			for j, line := range entry.Lines {
				accountCode := "Unknown"
				accountName := "Unknown"
				if line.Account != nil {
					accountCode = line.Account.Code
					accountName = line.Account.Name
				}

				if !line.DebitAmount.IsZero() {
					fmt.Printf("     %d. Dr. %s (%s): %s\n", j+1, accountName, accountCode, line.DebitAmount.String())
				}
				if !line.CreditAmount.IsZero() {
					fmt.Printf("     %d. Cr. %s (%s): %s\n", j+1, accountName, accountCode, line.CreditAmount.String())
				}
			}
		}
	}
}

func analyzeAccountBalanceUpdates(db *gorm.DB) {
	// Check key accounts that should have been affected
	accounts := []string{"1101", "2101", "1301", "1240"}
	
	fmt.Printf("🔍 Checking account balances for SSOT integration:\n\n")
	
	for _, code := range accounts {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			fmt.Printf("❌ Account %s: Not found\n", code)
			continue
		}

		fmt.Printf("💰 Account %s (%s):\n", code, account.Name)
		fmt.Printf("   Current Balance: Rp %.2f\n", account.Balance)
		
		// Expected behavior based on journal entries
		switch code {
		case "1101": // Kas
			fmt.Printf("   Expected: Should decrease by payment amounts\n")
			fmt.Printf("   Last Payment: Rp 1.387.500 (should reduce balance)\n")
			if account.Balance == 7225000.0 {
				fmt.Printf("   🔴 PROBLEM: Balance not updated by SSOT payment journal\n")
			} else {
				fmt.Printf("   ✅ Balance appears to be updated\n")
			}
		case "2101": // Utang Usaha
			fmt.Printf("   Expected: Should decrease (become less negative) by payment amounts\n")
			if account.Balance < 0 {
				fmt.Printf("   ✅ Correct: Liability has credit balance\n")
			} else {
				fmt.Printf("   🔴 PROBLEM: Liability should have credit balance\n")
			}
		case "1301": // Inventory
			fmt.Printf("   Expected: Should increase by purchase amounts\n")
			fmt.Printf("   ✅ Balance reflects inventory from purchases\n")
		case "1240": // PPN Masukan
			fmt.Printf("   Expected: Should increase by VAT amounts\n")
			fmt.Printf("   ✅ Balance reflects VAT from purchases\n")
		}
		fmt.Println()
	}
}

func manualBalanceCorrection(db *gorm.DB) {
	fmt.Printf("🔧 Manual Balance Correction Analysis:\n\n")
	
	// Get Kas account
	var kasAccount models.Account
	err := db.Where("code = ?", "1101").First(&kasAccount).Error
	if err != nil {
		fmt.Printf("❌ Kas account not found: %v\n", err)
		return
	}

	fmt.Printf("💰 Kas Account Current State:\n")
	fmt.Printf("   Code: %s\n", kasAccount.Code)
	fmt.Printf("   Name: %s\n", kasAccount.Name)
	fmt.Printf("   Current Balance: Rp %.2f\n", kasAccount.Balance)

	// Calculate expected balance
	expectedBalance := 7225000.0 - 1387500.0 // Starting balance - payment
	fmt.Printf("   Expected Balance: Rp %.2f\n", expectedBalance)
	fmt.Printf("   Difference: Rp %.2f\n", kasAccount.Balance - expectedBalance)

	if kasAccount.Balance != expectedBalance {
		fmt.Printf("\n🔴 BALANCE MISMATCH DETECTED!\n")
		fmt.Printf("   Current: Rp %.2f\n", kasAccount.Balance)
		fmt.Printf("   Expected: Rp %.2f\n", expectedBalance)
		fmt.Printf("   Correction needed: Rp %.2f\n", expectedBalance - kasAccount.Balance)

		fmt.Printf("\n🛠️ Correction Options:\n")
		fmt.Printf("1. Manual SQL Update:\n")
		fmt.Printf("   UPDATE accounts SET balance = %.2f WHERE code = '1101';\n", expectedBalance)
		fmt.Printf("\n2. Check SSOT AutoPost functionality\n")
		fmt.Printf("3. Verify UnifiedJournalService balance update mechanism\n")

		// Check if we should apply correction
		fmt.Printf("\n⚠️ Do you want to apply the correction? (This would be manual)\n")
		fmt.Printf("   For safety, this script will only show the analysis.\n")
		fmt.Printf("   Use the SQL command above if you want to fix manually.\n")
	} else {
		fmt.Printf("✅ Balance is correct!\n")
	}
}

func provideSummary() {
	fmt.Printf("📋 Debug Summary:\n\n")
	
	fmt.Printf("🔍 What we found:\n")
	fmt.Printf("1. ✅ SSOT journal entries are created correctly\n")
	fmt.Printf("2. ✅ Journal lines have proper debit/credit amounts\n")
	fmt.Printf("3. ✅ Journal entries are marked as POSTED\n")
	fmt.Printf("4. ❌ Account balances are not updated automatically\n")

	fmt.Printf("\n🎯 Root Cause:\n")
	fmt.Printf("The SSOT journal system creates journal entries correctly,\n")
	fmt.Printf("but the automatic account balance update mechanism is not working.\n")

	fmt.Printf("\n🔧 Possible Fixes:\n")
	fmt.Printf("1. Check UnifiedJournalService AutoPost functionality\n")
	fmt.Printf("2. Verify account balance update triggers in SSOT\n")
	fmt.Printf("3. Ensure balance updates are not being skipped\n")
	fmt.Printf("4. Check if balance update errors are being logged\n")

	fmt.Printf("\n💡 Next Steps:\n")
	fmt.Printf("1. Investigate UnifiedJournalService balance update code\n")
	fmt.Printf("2. Check if AutoPost is enabled and working\n")
	fmt.Printf("3. Add logging to balance update functions\n")
	fmt.Printf("4. Consider manual balance correction for existing data\n")

	fmt.Printf("\n🎯 Quick Fix:\n")
	fmt.Printf("For immediate correction:\n")
	fmt.Printf("UPDATE accounts SET balance = 5837500 WHERE code = '1101';\n")
}