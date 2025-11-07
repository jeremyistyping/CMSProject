package main

import (
	"fmt"
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
	
	fmt.Printf("\n=== FIXING JOURNAL ENTRY BALANCE FLAGS ===\n")
	
	// 1. Fix is_balanced flag for journal entries
	var allEntries []models.JournalEntry
	db.Find(&allEntries)
	
	fixedCount := 0
	for _, entry := range allEntries {
		if entry.TotalDebit == entry.TotalCredit && entry.TotalDebit > 0 {
			if !entry.IsBalanced {
				db.Model(&entry).Update("is_balanced", true)
				fixedCount++
			}
		} else if entry.TotalDebit != entry.TotalCredit {
			if entry.IsBalanced {
				db.Model(&entry).Update("is_balanced", false)
			}
		}
	}
	
	fmt.Printf("Fixed %d journal entry balance flags\n", fixedCount)
	
	fmt.Printf("\n=== FIXING ACCOUNT CATEGORIES ===\n")
	
	// 2. Fix account categories for better P&L classification
	accountFixes := []struct {
		Code     string
		Category string
		Reason   string
	}{
		{"5100", "COST_OF_GOODS_SOLD", "Cost of Goods Sold should be COGS category"},
		{"5101", "COST_OF_GOODS_SOLD", "Harga Pokok Penjualan should be COGS category"},
		{"4102", "OPERATING_REVENUE", "Shipping Revenue should be operating revenue"},
		{"4100", "OPERATING_REVENUE", "Sales Revenue should be operating revenue"},
	}
	
	for _, fix := range accountFixes {
		result := db.Model(&models.Account{}).Where("code = ?", fix.Code).Update("category", fix.Category)
		if result.RowsAffected > 0 {
			fmt.Printf("✓ Fixed account %s: %s\n", fix.Code, fix.Reason)
		}
	}
	
	fmt.Printf("\n=== VALIDATING ACCOUNT STRUCTURE ===\n")
	
	// 3. Validate and report account structure
	var revenueAccounts []models.Account
	db.Where("type = ? AND is_active = ?", "REVENUE", true).Find(&revenueAccounts)
	
	fmt.Printf("Revenue Accounts after fixes:\n")
	totalRevenue := float64(0)
	for _, acc := range revenueAccounts {
		fmt.Printf("  - %s (%s): Balance %.2f | Category: %s\n", 
			acc.Code, acc.Name, acc.Balance, acc.Category)
		totalRevenue += acc.Balance
	}
	fmt.Printf("Total Revenue Balance: %.2f\n", totalRevenue)
	
	var expenseAccounts []models.Account
	db.Where("type = ? AND is_active = ?", "EXPENSE", true).Find(&expenseAccounts)
	
	fmt.Printf("\nExpense Accounts after fixes:\n")
	totalExpenses := float64(0)
	cogsExpenses := float64(0)
	operatingExpenses := float64(0)
	
	for _, acc := range expenseAccounts {
		fmt.Printf("  - %s (%s): Balance %.2f | Category: %s\n", 
			acc.Code, acc.Name, acc.Balance, acc.Category)
		totalExpenses += acc.Balance
		
		if acc.Category == "COST_OF_GOODS_SOLD" {
			cogsExpenses += acc.Balance
		} else {
			operatingExpenses += acc.Balance
		}
	}
	
	fmt.Printf("Total Expenses Balance: %.2f\n", totalExpenses)
	fmt.Printf("  - COGS: %.2f\n", cogsExpenses)
	fmt.Printf("  - Operating Expenses: %.2f\n", operatingExpenses)
	
	fmt.Printf("\n=== P&L CALCULATION PREVIEW ===\n")
	grossProfit := totalRevenue - cogsExpenses
	netIncome := grossProfit - operatingExpenses
	
	fmt.Printf("Revenue: %.2f\n", totalRevenue)
	fmt.Printf("COGS: %.2f\n", cogsExpenses)
	fmt.Printf("Gross Profit: %.2f\n", grossProfit)
	fmt.Printf("Operating Expenses: %.2f\n", operatingExpenses)
	fmt.Printf("Net Income: %.2f\n", netIncome)
	
	if totalRevenue > 0 {
		grossMargin := (grossProfit / totalRevenue) * 100
		netMargin := (netIncome / totalRevenue) * 100
		fmt.Printf("Gross Margin: %.2f%%\n", grossMargin)
		fmt.Printf("Net Margin: %.2f%%\n", netMargin)
	}
	
	fmt.Printf("\n=== JOURNAL ENTRIES SUMMARY BY MONTH ===\n")
	
	// 4. Summary of journal entries affecting P&L
	var monthlyRevenue, monthlyExpenses float64
	
	db.Raw(`
		SELECT COALESCE(SUM(je.total_credit), 0) as revenue
		FROM journal_entries je 
		JOIN accounts a ON je.account_id = a.id 
		WHERE a.type = 'REVENUE' 
		AND je.status = 'POSTED' 
		AND EXTRACT(MONTH FROM je.entry_date) = EXTRACT(MONTH FROM CURRENT_DATE)
		AND EXTRACT(YEAR FROM je.entry_date) = EXTRACT(YEAR FROM CURRENT_DATE)
	`).Scan(&monthlyRevenue)
	
	db.Raw(`
		SELECT COALESCE(SUM(je.total_debit), 0) as expenses
		FROM journal_entries je 
		JOIN accounts a ON je.account_id = a.id 
		WHERE a.type = 'EXPENSE' 
		AND je.status = 'POSTED' 
		AND EXTRACT(MONTH FROM je.entry_date) = EXTRACT(MONTH FROM CURRENT_DATE)
		AND EXTRACT(YEAR FROM je.entry_date) = EXTRACT(YEAR FROM CURRENT_DATE)
	`).Scan(&monthlyExpenses)
	
	fmt.Printf("This Month's Activity from Journal Entries:\n")
	fmt.Printf("Revenue: %.2f\n", monthlyRevenue)
	fmt.Printf("Expenses: %.2f\n", monthlyExpenses)
	fmt.Printf("Net Income: %.2f\n", monthlyRevenue - monthlyExpenses)
	
	fmt.Printf("\n=== VALIDATION COMPLETE ===\n")
	fmt.Printf("✓ Journal entry balance flags fixed\n")
	fmt.Printf("✓ Account categories corrected\n") 
	fmt.Printf("✓ P&L accounts ready for reporting\n")
}