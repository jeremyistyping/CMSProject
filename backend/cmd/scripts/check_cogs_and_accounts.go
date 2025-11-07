package main

import (
	"fmt"
	"log"
	"strings"

	"app-sistem-akuntansi/models"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("CHECKING COGS JOURNALS & EXPENSE ACCOUNTS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// 1. Check COGS journals
	fmt.Println("1. COGS JOURNALS:")
	fmt.Println(strings.Repeat("-", 80))
	var cogsJournals []models.SimpleSSOTJournal
	db.Where("transaction_type = ?", "COGS").Find(&cogsJournals)
	
	if len(cogsJournals) == 0 {
		fmt.Println("❌ NO COGS journals found!")
	} else {
		fmt.Printf("✅ Found %d COGS journals:\n\n", len(cogsJournals))
		for _, j := range cogsJournals {
			fmt.Printf("   ID: %d | Entry: %s | Amount: Rp %.2f | Date: %s\n",
				j.ID, j.EntryNumber, j.TotalAmount, j.Date.Format("2006-01-02"))
			
			// Get journal items
			var items []models.SimpleSSOTJournalItem
			db.Where("journal_id = ?", j.ID).Find(&items)
			for _, item := range items {
				fmt.Printf("      %s (%s): Dr %.2f / Cr %.2f\n",
					item.AccountName, item.AccountCode, item.Debit, item.Credit)
			}
			fmt.Println()
		}
	}

	// 2. Check expense accounts (5xxx and 6xxx)
	fmt.Println()
	fmt.Println("2. EXPENSE & COGS ACCOUNTS:")
	fmt.Println(strings.Repeat("-", 80))
	
	var accounts []models.Account
	db.Where("code LIKE ? OR code LIKE ?", "5%", "6%").
		Where("is_active = ?", true).
		Order("code").
		Find(&accounts)
	
	if len(accounts) == 0 {
		fmt.Println("❌ NO expense/COGS accounts found!")
		fmt.Println()
		fmt.Println("⚠️  THIS IS THE PROBLEM!")
		fmt.Println("   You need to create COGS account (5101) first!")
		fmt.Println()
		fmt.Println("SQL to create account:")
		fmt.Println("   INSERT INTO accounts (code, name, type, is_active, created_at, updated_at)")
		fmt.Println("   VALUES ('5101', 'HARGA POKOK PENJUALAN', 'EXPENSE', true, NOW(), NOW());")
	} else {
		fmt.Printf("✅ Found %d expense accounts:\n\n", len(accounts))
		
		hasCOGS := false
		for _, acc := range accounts {
			fmt.Printf("   %s - %s (Balance: Rp %.2f)\n", acc.Code, acc.Name, acc.Balance)
			if strings.HasPrefix(acc.Code, "51") {
				hasCOGS = true
			}
		}
		
		fmt.Println()
		if !hasCOGS {
			fmt.Println("⚠️  No COGS account (51xx) found!")
			fmt.Println("   Create account 5101 - HARGA POKOK PENJUALAN")
		}
	}

	// 3. Check all journal entries for this period
	fmt.Println()
	fmt.Println("3. ALL JOURNAL ENTRIES (October 2025):")
	fmt.Println(strings.Repeat("-", 80))
	
	var allJournals []models.SimpleSSOTJournal
	db.Where("date >= ? AND date <= ?", "2025-10-01", "2025-10-31").
		Order("date, id").
		Find(&allJournals)
	
	fmt.Printf("Found %d journal entries\n\n", len(allJournals))
	for _, j := range allJournals {
		fmt.Printf("   %s | %s | %s | Rp %.2f\n",
			j.Date.Format("2006-01-02"), j.TransactionType, j.EntryNumber, j.TotalAmount)
	}

	// 4. Summary for P&L
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("SUMMARY FOR P&L")
	fmt.Println(strings.Repeat("=", 80))
	
	// Calculate totals from journal items
	var items []struct {
		AccountCode string
		AccountName string
		TotalDebit  float64
		TotalCredit float64
	}
	
	db.Raw(`
		SELECT 
			account_code,
			account_name,
			SUM(debit) as total_debit,
			SUM(credit) as total_credit
		FROM simple_ssot_journal_items
		WHERE journal_id IN (
			SELECT id FROM simple_ssot_journals 
			WHERE date >= '2025-10-01' AND date <= '2025-10-31'
			AND deleted_at IS NULL
		)
		GROUP BY account_code, account_name
		ORDER BY account_code
	`).Scan(&items)
	
	var totalRevenue, totalCOGS, totalExpenses float64
	
	fmt.Println()
	fmt.Println("Account Balances for October 2025:")
	fmt.Println()
	
	for _, item := range items {
		netAmount := item.TotalDebit - item.TotalCredit
		
		if netAmount != 0 {
			fmt.Printf("   %s - %s: Rp %.2f\n", item.AccountCode, item.AccountName, netAmount)
			
			// Categorize
			if strings.HasPrefix(item.AccountCode, "4") {
				totalRevenue += netAmount
			} else if strings.HasPrefix(item.AccountCode, "51") {
				totalCOGS += netAmount
			} else if strings.HasPrefix(item.AccountCode, "5") || strings.HasPrefix(item.AccountCode, "6") {
				totalExpenses += netAmount
			}
		}
	}
	
	fmt.Println()
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total Revenue (4xxx):    Rp %12.2f\n", totalRevenue)
	fmt.Printf("Total COGS (51xx):       Rp %12.2f\n", totalCOGS)
	fmt.Printf("Total Expenses (5/6xxx): Rp %12.2f\n", totalExpenses)
	fmt.Println(strings.Repeat("=", 80))
	
	if totalExpenses == 0 && totalCOGS == 0 {
		fmt.Println()
		fmt.Println("❌ PROBLEM IDENTIFIED:")
		fmt.Println("   No COGS or Expense accounts have transactions!")
		fmt.Println()
		fmt.Println("POSSIBLE CAUSES:")
		fmt.Println("   1. COGS journals not created yet")
		fmt.Println("   2. COGS account (5101) doesn't exist")
		fmt.Println("   3. Purchase recorded to Inventory but no COGS journal when sold")
		fmt.Println()
		fmt.Println("SOLUTION:")
		fmt.Println("   1. Create account 5101 if missing")
		fmt.Println("   2. Run migration to create COGS journals")
		fmt.Println("   3. Restart backend server")
	}
}

