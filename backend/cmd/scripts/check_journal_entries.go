package main

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	fmt.Printf("Using database: %s\n", cfg.DatabaseURL)
	
	// Connect to database
	db := database.ConnectDB()
	
	// 1. Check total journal entries
	var totalJournalEntries int64
	db.Model(&models.JournalEntry{}).Count(&totalJournalEntries)
	fmt.Printf("\n=== JOURNAL ENTRIES ANALYSIS ===\n")
	fmt.Printf("Total Journal Entries: %d\n", totalJournalEntries)
	
	// 2. Check journal entries by status
	var draftEntries, postedEntries, reversedEntries int64
	db.Model(&models.JournalEntry{}).Where("status = ?", "DRAFT").Count(&draftEntries)
	db.Model(&models.JournalEntry{}).Where("status = ?", "POSTED").Count(&postedEntries)
	db.Model(&models.JournalEntry{}).Where("status = ?", "REVERSED").Count(&reversedEntries)
	
	fmt.Printf("  - Draft Entries: %d\n", draftEntries)
	fmt.Printf("  - Posted Entries: %d\n", postedEntries)
	fmt.Printf("  - Reversed Entries: %d\n", reversedEntries)
	
	// 3. Check journal entries by reference type
	var journalsByType []struct {
		ReferenceType string
		Count         int64
	}
	db.Model(&models.JournalEntry{}).
		Select("reference_type, COUNT(*) as count").
		Where("reference_type IS NOT NULL AND reference_type != ''").
		Group("reference_type").
		Find(&journalsByType)
	
	fmt.Printf("\nJournal Entries by Type:\n")
	for _, item := range journalsByType {
		fmt.Printf("  - %s: %d\n", item.ReferenceType, item.Count)
	}
	
	// 4. Check recent journal entries (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var recentEntries []models.JournalEntry
	db.Where("created_at >= ?", thirtyDaysAgo).
		Order("created_at DESC").
		Limit(10).
		Find(&recentEntries)
	
	fmt.Printf("\nRecent Journal Entries (Last 10):\n")
	for _, entry := range recentEntries {
		fmt.Printf("  - Code: %s | Status: %s | Type: %s | Date: %s | Amount: %.2f\n", 
			entry.Code, 
			entry.Status, 
			entry.ReferenceType, 
			entry.CreatedAt.Format("2006-01-02"),
			entry.TotalDebit)
	}
	
	// 5. Check journal lines
	var totalJournalLines int64
	db.Model(&models.JournalLine{}).Count(&totalJournalLines)
	fmt.Printf("\n=== JOURNAL LINES ANALYSIS ===\n")
	fmt.Printf("Total Journal Lines: %d\n", totalJournalLines)
	
	// 6. Check accounts that are used in P&L
	fmt.Printf("\n=== ACCOUNTS ANALYSIS FOR P&L ===\n")
	
	// Revenue accounts
	var revenueAccounts []models.Account
	db.Where("type = ? AND is_active = ?", "REVENUE", true).Find(&revenueAccounts)
	fmt.Printf("Revenue Accounts: %d\n", len(revenueAccounts))
	for _, acc := range revenueAccounts {
		fmt.Printf("  - %s (%s): %s | Balance: %.2f | Category: %s\n", 
			acc.Code, acc.Name, acc.Type, acc.Balance, acc.Category)
	}
	
	// Expense accounts
	var expenseAccounts []models.Account
	db.Where("type = ? AND is_active = ?", "EXPENSE", true).Find(&expenseAccounts)
	fmt.Printf("\nExpense Accounts: %d\n", len(expenseAccounts))
	for _, acc := range expenseAccounts {
		fmt.Printf("  - %s (%s): %s | Balance: %.2f | Category: %s\n", 
			acc.Code, acc.Name, acc.Type, acc.Balance, acc.Category)
	}
	
	// 7. Check journal entries for current month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	var currentMonthEntries int64
	db.Model(&models.JournalEntry{}).
		Where("entry_date >= ? AND status = ?", startOfMonth, "POSTED").
		Count(&currentMonthEntries)
	
	fmt.Printf("\n=== CURRENT MONTH ANALYSIS ===\n")
	fmt.Printf("Posted Journal Entries This Month: %d\n", currentMonthEntries)
	
	// 8. Check for potential COGS accounts
	var cogsAccounts []models.Account
	db.Where("(type = ? AND (category LIKE '%COST_OF_GOODS%' OR category LIKE '%COGS%' OR name ILIKE '%harga pokok%' OR name ILIKE '%cost of goods%')) AND is_active = ?", 
		"EXPENSE", true).Find(&cogsAccounts)
	
	fmt.Printf("\nPotential COGS Accounts: %d\n", len(cogsAccounts))
	for _, acc := range cogsAccounts {
		fmt.Printf("  - %s (%s): %s | Balance: %.2f | Category: %s\n", 
			acc.Code, acc.Name, acc.Type, acc.Balance, acc.Category)
	}
	
	// 9. Check if there are any balance inconsistencies
	fmt.Printf("\n=== BALANCE VALIDATION ===\n")
	
	// Check journal entries balance
	var unbalancedEntries []models.JournalEntry
	db.Where("is_balanced = false AND status = ?", "POSTED").Find(&unbalancedEntries)
	fmt.Printf("Unbalanced Posted Journal Entries: %d\n", len(unbalancedEntries))
	
	if len(unbalancedEntries) > 0 {
		fmt.Printf("Warning: Found unbalanced posted journal entries:\n")
		for _, entry := range unbalancedEntries {
			fmt.Printf("  - Code: %s | Debit: %.2f | Credit: %.2f | Difference: %.2f\n",
				entry.Code, entry.TotalDebit, entry.TotalCredit, 
				entry.TotalDebit-entry.TotalCredit)
		}
	}
	
	// 10. Summary for P&L preparation
	fmt.Printf("\n=== P&L READINESS SUMMARY ===\n")
	fmt.Printf("✓ Total Journal Entries: %d\n", totalJournalEntries)
	fmt.Printf("✓ Posted Entries: %d\n", postedEntries)
	fmt.Printf("✓ Revenue Accounts: %d\n", len(revenueAccounts))
	fmt.Printf("✓ Expense Accounts: %d\n", len(expenseAccounts))
	fmt.Printf("✓ COGS Accounts: %d\n", len(cogsAccounts))
	fmt.Printf("✓ Current Month Posted Entries: %d\n", currentMonthEntries)
	
	if len(unbalancedEntries) > 0 {
		fmt.Printf("⚠️  Warning: %d unbalanced posted entries need attention\n", len(unbalancedEntries))
	} else {
		fmt.Printf("✓ All posted journal entries are balanced\n")
	}
	
	fmt.Printf("\n=== READY TO GENERATE P&L STATEMENT ===\n")
}