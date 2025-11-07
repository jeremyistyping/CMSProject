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
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("CHECK PURCHASE JOURNAL DETAIL: PO/2025/10/0003")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// 1. Get purchase
	var purchase models.Purchase
	err = db.Preload("Vendor").
		Where("code = ?", "PO/2025/10/0003").
		First(&purchase).Error
	
	if err != nil {
		log.Fatalf("Failed to fetch purchase: %v", err)
	}

	fmt.Printf("Purchase ID: %d\n", purchase.ID)
	fmt.Printf("Purchase Code: %s\n", purchase.Code)
	fmt.Printf("Vendor: %s\n", purchase.Vendor.Name)
	fmt.Printf("Date: %s\n", purchase.Date.Format("2006-01-02"))
	fmt.Printf("Total Amount: Rp %.2f\n", purchase.TotalAmount)
	fmt.Printf("PPN Amount: Rp %.2f\n", purchase.PPNAmount)
	fmt.Printf("Status: %s\n", purchase.Status)
	fmt.Printf("Approval Status: %s\n", purchase.ApprovalStatus)
	fmt.Println()

	// 2. Get journal entries
	var journals []models.SimpleSSOTJournal
	err = db.Where("transaction_type = ? AND transaction_id = ?", "PURCHASE", purchase.ID).
		Find(&journals).Error
	
	if err != nil {
		log.Fatalf("Failed to fetch journals: %v", err)
	}

	if len(journals) == 0 {
		fmt.Println("❌ NO JOURNAL ENTRIES FOUND!")
		return
	}

	fmt.Printf("✅ Found %d journal entry/entries:\n", len(journals))
	fmt.Println()

	for _, journal := range journals {
		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("Journal ID: %d\n", journal.ID)
		fmt.Printf("Entry Number: %s\n", journal.EntryNumber)
		fmt.Printf("Date: %s\n", journal.Date.Format("2006-01-02"))
		fmt.Printf("Total Amount: Rp %.2f\n", journal.TotalAmount)
		fmt.Printf("Status: %s\n", journal.Status)
		fmt.Println()

		// Get journal items
		var items []models.SimpleSSOTJournalItem
		err = db.Where("journal_id = ?", journal.ID).Find(&items).Error
		if err != nil {
			fmt.Printf("⚠️  Failed to fetch journal items: %v\n", err)
			continue
		}

		fmt.Printf("Journal Items (%d lines):\n", len(items))
		fmt.Println()

		var totalDebit, totalCredit float64

		for i, item := range items {
			fmt.Printf("%d. Account: %s - %s\n", i+1, item.AccountCode, item.AccountName)
			fmt.Printf("   Debit:  Rp %12.2f\n", item.Debit)
			fmt.Printf("   Credit: Rp %12.2f\n", item.Credit)
			fmt.Printf("   Desc:   %s\n", item.Description)
			fmt.Println()

			totalDebit += item.Debit
			totalCredit += item.Credit

			// Analyze account categorization
			code := item.AccountCode
			if len(code) > 0 {
				switch {
				case strings.HasPrefix(code, "1"):
					fmt.Printf("   📊 Category: ASSET\n")
				case strings.HasPrefix(code, "2"):
					fmt.Printf("   📊 Category: LIABILITY\n")
				case strings.HasPrefix(code, "4"):
					fmt.Printf("   📊 Category: REVENUE\n")
				case strings.HasPrefix(code, "51"):
					fmt.Printf("   📊 Category: COGS (Cost of Goods Sold)\n")
				case strings.HasPrefix(code, "52") || strings.HasPrefix(code, "53") || strings.HasPrefix(code, "54") ||
					strings.HasPrefix(code, "55") || strings.HasPrefix(code, "56") || strings.HasPrefix(code, "57") ||
					strings.HasPrefix(code, "58") || strings.HasPrefix(code, "59"):
					fmt.Printf("   📊 Category: OPERATING EXPENSES (52-59)\n")
				case strings.HasPrefix(code, "60") || strings.HasPrefix(code, "61"):
					fmt.Printf("   📊 Category: OPERATING EXPENSES (60-61) ✅ FIXED\n")
				case strings.HasPrefix(code, "6"):
					fmt.Printf("   📊 Category: OTHER EXPENSES (62-69)\n")
				case strings.HasPrefix(code, "7"):
					fmt.Printf("   📊 Category: OTHER INCOME\n")
				default:
					fmt.Printf("   📊 Category: UNKNOWN\n")
				}
			}
			fmt.Println()
		}

		fmt.Println(strings.Repeat("-", 40))
		fmt.Printf("Total Debit:  Rp %12.2f\n", totalDebit)
		fmt.Printf("Total Credit: Rp %12.2f\n", totalCredit)
		fmt.Printf("Balanced: %v\n", totalDebit == totalCredit)
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("ANALYSIS & RECOMMENDATION")
	fmt.Println(strings.Repeat("=", 80))
	
	// Analyze the journal items
	var hasExpense, hasCOGS, hasInventory bool
	var expenseAmount, cogsAmount float64
	
	for _, journal := range journals {
		var items []models.SimpleSSOTJournalItem
		db.Where("journal_id = ?", journal.ID).Find(&items)
		
		for _, item := range items {
			code := item.AccountCode
			if item.Debit > 0 {  // Only check debit side
				if strings.HasPrefix(code, "51") {
					hasCOGS = true
					cogsAmount += item.Debit
				} else if strings.HasPrefix(code, "5") || strings.HasPrefix(code, "6") {
					hasExpense = true
					expenseAmount += item.Debit
				} else if strings.HasPrefix(code, "13") || strings.HasPrefix(code, "15") {
					hasInventory = true
				}
			}
		}
	}
	
	fmt.Println()
	if hasCOGS {
		fmt.Printf("✅ Purchase recorded as COGS (5xxx): Rp %.2f\n", cogsAmount)
		fmt.Println("   This SHOULD appear in P&L report under COGS section")
	}
	
	if hasExpense {
		fmt.Printf("✅ Purchase recorded as Expense (5xxx/6xxx): Rp %.2f\n", expenseAmount)
		fmt.Println("   This SHOULD appear in P&L report under Operating Expenses")
		fmt.Println("   ✅ With the latest fix, account 60xx-61xx now included in Operating Expenses")
	}
	
	if hasInventory {
		fmt.Println("⚠️  Purchase recorded as Inventory (13xx/15xx)")
		fmt.Println("   This will NOT appear in P&L until goods are sold")
		fmt.Println("   Need to create COGS entry when goods are sold:")
		fmt.Println("   Dr. COGS (5001)")
		fmt.Println("       Cr. Inventory (13xx)")
	}
	
	if !hasCOGS && !hasExpense && hasInventory {
		fmt.Println()
		fmt.Println("❌ ISSUE IDENTIFIED:")
		fmt.Println("   Purchase is recorded to Inventory, NOT to Expense/COGS")
		fmt.Println("   This is why P&L shows Rp 0 for expenses")
		fmt.Println()
		fmt.Println("📋 RECOMMENDED ACTION:")
		fmt.Println("   For trading/retail business: Create COGS entry when selling")
		fmt.Println("   For service business: Change purchase to expense directly (6001 or 5001)")
	}
	
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
}

