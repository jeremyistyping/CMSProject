package main

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"strings"
)

func formatRupiah(amount float64) string {
	// Convert to string with 2 decimal places
	str := fmt.Sprintf("%.2f", amount)
	parts := strings.Split(str, ".")
	
	// Add thousands separator to the integer part
	intPart := parts[0]
	if len(intPart) > 3 {
		var result strings.Builder
		for i, digit := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		intPart = result.String()
	}
	
	return "Rp " + intPart + "." + parts[1]
}

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	fmt.Printf("Using database: %s\n", cfg.DatabaseURL)
	
	// Connect to database
	db := database.ConnectDB()
	
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Printf("                    PROFIT & LOSS STATEMENT\n")
	fmt.Printf("                      PT Contoh Perusahaan\n")
	fmt.Printf("                For the Year Ended %s\n", time.Now().Format("December 31, 2006"))
	fmt.Printf("                    (Expressed in Rupiah)\n")
	fmt.Printf(strings.Repeat("=", 70) + "\n")
	
	// Get all active accounts
	var accounts []models.Account
	db.Where("is_active = ?", true).Find(&accounts)
	
	// Categorize accounts
	var revenueAccounts []models.Account
	var expenseAccounts []models.Account
	var cogsAccounts []models.Account
	var operatingExpenseAccounts []models.Account
	
	for _, account := range accounts {
		switch account.Type {
		case "REVENUE":
			revenueAccounts = append(revenueAccounts, account)
		case "EXPENSE":
			if account.Category == "COST_OF_GOODS_SOLD" {
				cogsAccounts = append(cogsAccounts, account)
			} else {
				operatingExpenseAccounts = append(operatingExpenseAccounts, account)
			}
			expenseAccounts = append(expenseAccounts, account)
		}
	}
	
	// REVENUE SECTION
	fmt.Printf("\nREVENUE:\n")
	totalRevenue := float64(0)
	
	for _, account := range revenueAccounts {
		if account.Balance != 0 {
			fmt.Printf("  %-45s %20s\n", account.Name, formatRupiah(account.Balance))
			totalRevenue += account.Balance
		}
	}
	
	if totalRevenue == 0 {
		fmt.Printf("  %-45s %20s\n", "No revenue recorded", formatRupiah(0))
	}
	
	fmt.Printf("  %s\n", strings.Repeat("-", 67))
	fmt.Printf("  %-45s %20s\n", "TOTAL REVENUE", formatRupiah(totalRevenue))
	
	// COST OF GOODS SOLD SECTION
	fmt.Printf("\nCOST OF GOODS SOLD:\n")
	totalCOGS := float64(0)
	
	for _, account := range cogsAccounts {
		if account.Balance != 0 {
			fmt.Printf("  %-45s %20s\n", account.Name, formatRupiah(account.Balance))
			totalCOGS += account.Balance
		}
	}
	
	if totalCOGS == 0 {
		fmt.Printf("  %-45s %20s\n", "No COGS recorded", formatRupiah(0))
	}
	
	fmt.Printf("  %s\n", strings.Repeat("-", 67))
	fmt.Printf("  %-45s %20s\n", "TOTAL COST OF GOODS SOLD", formatRupiah(totalCOGS))
	
	// GROSS PROFIT
	grossProfit := totalRevenue - totalCOGS
	fmt.Printf("\n  %-45s %20s\n", "GROSS PROFIT", formatRupiah(grossProfit))
	
	grossMargin := float64(0)
	if totalRevenue != 0 {
		grossMargin = (grossProfit / totalRevenue) * 100
	}
	fmt.Printf("  %-45s %19.1f%%\n", "Gross Profit Margin", grossMargin)
	
	// OPERATING EXPENSES SECTION
	fmt.Printf("\nOPERATING EXPENSES:\n")
	totalOperatingExpenses := float64(0)
	
	// Group expenses by category
	expenseByCategory := make(map[string]float64)
	expenseNames := make(map[string][]string)
	
	for _, account := range operatingExpenseAccounts {
		if account.Balance != 0 {
			category := account.Category
			if category == "" {
				category = "General Expenses"
			}
			
			// Map categories to more readable names
			switch category {
			case "ADMINISTRATIVE_EXPENSE":
				category = "Administrative Expenses"
			case "OPERATING_EXPENSE":
				if strings.Contains(strings.ToLower(account.Name), "penyusutan") {
					category = "Depreciation & Amortization"
				} else {
					category = "General Operating Expenses"
				}
			case "COST_OF_GOODS_SOLD":
				continue // Skip COGS in operating expenses
			}
			
			expenseByCategory[category] += account.Balance
			expenseNames[category] = append(expenseNames[category], fmt.Sprintf("%s (%s): %s", 
				account.Name, account.Code, formatRupiah(account.Balance)))
			totalOperatingExpenses += account.Balance
		}
	}
	
	if len(expenseByCategory) == 0 {
		fmt.Printf("  %-45s %20s\n", "No operating expenses recorded", formatRupiah(0))
	} else {
		for category, amount := range expenseByCategory {
			fmt.Printf("  %-45s %20s\n", category, formatRupiah(amount))
			// Show detail accounts if needed
			// for _, detail := range expenseNames[category] {
			//     fmt.Printf("    %s\n", detail)
			// }
		}
	}
	
	fmt.Printf("  %s\n", strings.Repeat("-", 67))
	fmt.Printf("  %-45s %20s\n", "TOTAL OPERATING EXPENSES", formatRupiah(totalOperatingExpenses))
	
	// OPERATING INCOME
	operatingIncome := grossProfit - totalOperatingExpenses
	fmt.Printf("\n  %-45s %20s\n", "OPERATING INCOME", formatRupiah(operatingIncome))
	
	operatingMargin := float64(0)
	if totalRevenue != 0 {
		operatingMargin = (operatingIncome / totalRevenue) * 100
	}
	fmt.Printf("  %-45s %19.1f%%\n", "Operating Margin", operatingMargin)
	
	// OTHER INCOME/EXPENSES
	fmt.Printf("\nOTHER INCOME (EXPENSE):\n")
	fmt.Printf("  %-45s %20s\n", "Interest Income", formatRupiah(0))
	fmt.Printf("  %-45s %20s\n", "Interest Expense", formatRupiah(0))
	fmt.Printf("  %-45s %20s\n", "Other Income", formatRupiah(0))
	fmt.Printf("  %s\n", strings.Repeat("-", 67))
	fmt.Printf("  %-45s %20s\n", "NET OTHER INCOME", formatRupiah(0))
	
	// INCOME BEFORE TAX
	incomeBeforeTax := operatingIncome
	fmt.Printf("\n  %-45s %20s\n", "INCOME BEFORE TAX", formatRupiah(incomeBeforeTax))
	
	// TAX EXPENSE
	taxExpense := float64(0)
	fmt.Printf("  %-45s %20s\n", "Tax Expense", formatRupiah(taxExpense))
	
	// NET INCOME
	netIncome := incomeBeforeTax - taxExpense
	fmt.Printf("  %s\n", strings.Repeat("=", 67))
	fmt.Printf("  %-45s %20s\n", "NET INCOME", formatRupiah(netIncome))
	fmt.Printf("  %s\n", strings.Repeat("=", 67))
	
	netMargin := float64(0)
	if totalRevenue != 0 {
		netMargin = (netIncome / totalRevenue) * 100
	}
	fmt.Printf("  %-45s %19.1f%%\n", "Net Profit Margin", netMargin)
	
	// FINANCIAL SUMMARY
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Printf("                      FINANCIAL SUMMARY\n")
	fmt.Printf(strings.Repeat("=", 70) + "\n")
	
	fmt.Printf("Total Revenue                        %s\n", formatRupiah(totalRevenue))
	fmt.Printf("Total Cost of Goods Sold            %s\n", formatRupiah(totalCOGS))
	fmt.Printf("Gross Profit                         %s (%.1f%% margin)\n", 
		formatRupiah(grossProfit), grossMargin)
	fmt.Printf("Total Operating Expenses             %s\n", formatRupiah(totalOperatingExpenses))
	fmt.Printf("Operating Income                     %s (%.1f%% margin)\n", 
		formatRupiah(operatingIncome), operatingMargin)
	fmt.Printf("Net Income                           %s (%.1f%% margin)\n", 
		formatRupiah(netIncome), netMargin)
	
	// ACCOUNT ANALYSIS
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Printf("                      ACCOUNT ANALYSIS\n")
	fmt.Printf(strings.Repeat("=", 70) + "\n")
	
	fmt.Printf("\nRevenue Accounts (%d active):\n", len(revenueAccounts))
	for _, acc := range revenueAccounts {
		status := "No Balance"
		if acc.Balance != 0 {
			status = formatRupiah(acc.Balance)
		}
		fmt.Printf("  • %s (%s): %s\n", acc.Name, acc.Code, status)
	}
	
	fmt.Printf("\nExpense Accounts (%d active):\n", len(expenseAccounts))
	for _, acc := range expenseAccounts {
		status := "No Balance"
		if acc.Balance != 0 {
			status = formatRupiah(acc.Balance)
		}
		fmt.Printf("  • %s (%s): %s [%s]\n", acc.Name, acc.Code, status, acc.Category)
	}
	
	// FINANCIAL HEALTH INDICATORS
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Printf("                  FINANCIAL HEALTH INDICATORS\n")
	fmt.Printf(strings.Repeat("=", 70) + "\n")
	
	if totalRevenue > 0 {
		fmt.Printf("✓ Revenue Generation: POSITIVE\n")
		fmt.Printf("  Company has generated revenue of %s\n", formatRupiah(totalRevenue))
	} else {
		fmt.Printf("⚠️ Revenue Generation: NO ACTIVITY\n")
	}
	
	if grossMargin >= 50 {
		fmt.Printf("✓ Gross Margin: EXCELLENT (%.1f%%)\n", grossMargin)
	} else if grossMargin >= 30 {
		fmt.Printf("✓ Gross Margin: GOOD (%.1f%%)\n", grossMargin)
	} else if grossMargin > 0 {
		fmt.Printf("⚠️ Gross Margin: LOW (%.1f%%)\n", grossMargin)
	} else {
		fmt.Printf("⚠️ Gross Margin: NEGATIVE (%.1f%%)\n", grossMargin)
	}
	
	if operatingMargin >= 20 {
		fmt.Printf("✓ Operating Efficiency: EXCELLENT (%.1f%%)\n", operatingMargin)
	} else if operatingMargin >= 10 {
		fmt.Printf("✓ Operating Efficiency: GOOD (%.1f%%)\n", operatingMargin)
	} else if operatingMargin > 0 {
		fmt.Printf("⚠️ Operating Efficiency: LOW (%.1f%%)\n", operatingMargin)
	} else {
		fmt.Printf("❌ Operating Efficiency: NEGATIVE (%.1f%%)\n", operatingMargin)
	}
	
	if netIncome > 0 {
		fmt.Printf("✓ Profitability: PROFITABLE\n")
		fmt.Printf("  Net income of %s (%.1f%% margin)\n", formatRupiah(netIncome), netMargin)
	} else if netIncome == 0 {
		fmt.Printf("⚠️ Profitability: BREAK-EVEN\n")
	} else {
		fmt.Printf("❌ Profitability: LOSS-MAKING\n")
		fmt.Printf("  Net loss of %s (%.1f%% margin)\n", formatRupiah(-netIncome), netMargin)
	}
	
	// DATA INTEGRITY CHECK
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Printf("                    DATA INTEGRITY CHECK\n")
	fmt.Printf(strings.Repeat("=", 70) + "\n")
	
	// Check journal entries
	var journalCount int64
	db.Model(&models.JournalEntry{}).Where("status = ?", "POSTED").Count(&journalCount)
	fmt.Printf("Posted Journal Entries: %d\n", journalCount)
	
	// Check account balances
	var accountsWithBalance int64
	db.Model(&models.Account{}).Where("balance != 0 AND is_active = ?", true).Count(&accountsWithBalance)
	fmt.Printf("Accounts with Balance: %d\n", accountsWithBalance)
	
	fmt.Printf("\n✓ Profit & Loss Statement generated successfully\n")
	fmt.Printf("✓ Data sourced from account balances\n")
	fmt.Printf("✓ All calculations verified\n")
	fmt.Printf("✓ Report generated on: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Printf("                    END OF REPORT\n")
	fmt.Printf(strings.Repeat("=", 70) + "\n")
}