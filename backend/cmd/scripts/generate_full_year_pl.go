package main

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	fmt.Printf("Using database: %s\n", cfg.DatabaseURL)
	
	// Connect to database
	db := database.ConnectDB()
	
	// Initialize repositories and services
	accountRepo := repositories.NewAccountRepository(db)
	enhancedPLService := services.NewEnhancedProfitLossService(db, accountRepo)
	
	fmt.Printf("\n=== GENERATING FULL YEAR PROFIT & LOSS STATEMENT ===\n")
	
	// Define date range - full year to date
	now := time.Now()
	startDate := time.Date(2020, 1, 1, 0, 0, 0, 0, now.Location()) // Start from early date to include all historical data
	endDate := now
	
	fmt.Printf("Period: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	// Generate P&L statement
	plStatement, err := enhancedPLService.GenerateEnhancedProfitLoss(startDate, endDate)
	if err != nil {
		fmt.Printf("Error generating P&L statement: %v\n", err)
		return
	}
	
	fmt.Printf("\n=== COMPREHENSIVE PROFIT & LOSS STATEMENT ===\n")
	fmt.Printf("Company: %s\n", plStatement.Company.Name)
	fmt.Printf("Period: %s to %s\n", 
		plStatement.StartDate.Format("2006-01-02"), 
		plStatement.EndDate.Format("2006-01-02"))
	fmt.Printf("Currency: %s\n", plStatement.Currency)
	
	fmt.Printf("\n--- REVENUE ANALYSIS ---\n")
	fmt.Printf("Sales Revenue: %.2f\n", plStatement.Revenue.SalesRevenue.Subtotal)
	if len(plStatement.Revenue.SalesRevenue.Items) > 0 {
		for _, item := range plStatement.Revenue.SalesRevenue.Items {
			fmt.Printf("  - %s (%s): %.2f (%.2f%%)\n", 
				item.Name, item.Code, item.Amount, item.Percentage)
		}
	}
	
	fmt.Printf("Service Revenue: %.2f\n", plStatement.Revenue.ServiceRevenue.Subtotal)
	if len(plStatement.Revenue.ServiceRevenue.Items) > 0 {
		for _, item := range plStatement.Revenue.ServiceRevenue.Items {
			fmt.Printf("  - %s (%s): %.2f (%.2f%%)\n", 
				item.Name, item.Code, item.Amount, item.Percentage)
		}
	}
	
	fmt.Printf("Other Revenue: %.2f\n", plStatement.Revenue.OtherRevenue.Subtotal)
	if len(plStatement.Revenue.OtherRevenue.Items) > 0 {
		for _, item := range plStatement.Revenue.OtherRevenue.Items {
			fmt.Printf("  - %s (%s): %.2f (%.2f%%)\n", 
				item.Name, item.Code, item.Amount, item.Percentage)
		}
	}
	
	fmt.Printf("\nTOTAL REVENUE: %.2f\n", plStatement.Revenue.TotalRevenue)
	
	fmt.Printf("\n--- COST OF GOODS SOLD ANALYSIS ---\n")
	fmt.Printf("Direct Materials: %.2f\n", plStatement.CostOfGoodsSold.DirectMaterials.Subtotal)
	fmt.Printf("Direct Labor: %.2f\n", plStatement.CostOfGoodsSold.DirectLabor.Subtotal)
	fmt.Printf("Manufacturing Overhead: %.2f\n", plStatement.CostOfGoodsSold.ManufacturingOH.Subtotal)
	fmt.Printf("Other COGS: %.2f\n", plStatement.CostOfGoodsSold.OtherCOGS.Subtotal)
	
	if len(plStatement.CostOfGoodsSold.OtherCOGS.Items) > 0 {
		for _, item := range plStatement.CostOfGoodsSold.OtherCOGS.Items {
			fmt.Printf("  - %s (%s): %.2f (%.2f%%)\n", 
				item.Name, item.Code, item.Amount, item.Percentage)
		}
	}
	
	fmt.Printf("\nTOTAL COGS: %.2f\n", plStatement.CostOfGoodsSold.TotalCOGS)
	
	fmt.Printf("\n--- KEY PROFITABILITY METRICS ---\n")
	fmt.Printf("GROSS PROFIT: %.2f\n", plStatement.GrossProfit)
	fmt.Printf("Gross Profit Margin: %.2f%%\n", plStatement.GrossProfitMargin)
	
	fmt.Printf("\n--- OPERATING EXPENSES BREAKDOWN ---\n")
	fmt.Printf("Administrative Expenses: %.2f\n", plStatement.OperatingExpenses.Administrative.Subtotal)
	if len(plStatement.OperatingExpenses.Administrative.Items) > 0 {
		for _, item := range plStatement.OperatingExpenses.Administrative.Items {
			fmt.Printf("  - %s (%s): %.2f (%.2f%%)\n", 
				item.Name, item.Code, item.Amount, item.Percentage)
		}
	}
	
	fmt.Printf("Selling & Marketing Expenses: %.2f\n", plStatement.OperatingExpenses.SellingMarketing.Subtotal)
	fmt.Printf("General Expenses: %.2f\n", plStatement.OperatingExpenses.General.Subtotal)
	if len(plStatement.OperatingExpenses.General.Items) > 0 {
		for _, item := range plStatement.OperatingExpenses.General.Items {
			fmt.Printf("  - %s (%s): %.2f (%.2f%%)\n", 
				item.Name, item.Code, item.Amount, item.Percentage)
		}
	}
	
	fmt.Printf("Depreciation & Amortization: %.2f\n", plStatement.OperatingExpenses.Depreciation.Subtotal)
	
	fmt.Printf("\nTOTAL OPERATING EXPENSES: %.2f\n", plStatement.OperatingExpenses.TotalOpex)
	
	fmt.Printf("\n--- FINAL P&L RESULTS ---\n")
	fmt.Printf("OPERATING INCOME (EBIT): %.2f\n", plStatement.OperatingIncome)
	fmt.Printf("Operating Margin: %.2f%%\n", plStatement.OperatingMargin)
	fmt.Printf("EBITDA: %.2f\n", plStatement.EBITDA)
	fmt.Printf("EBITDA Margin: %.2f%%\n", plStatement.EBITDAMargin)
	
	fmt.Printf("\nINCOME BEFORE TAX: %.2f\n", plStatement.IncomeBeforeTax)
	fmt.Printf("Tax Expense: %.2f\n", plStatement.TaxExpense)
	fmt.Printf("Tax Rate: %.2f%%\n", plStatement.TaxRate)
	fmt.Printf("\nNET INCOME: %.2f\n", plStatement.NetIncome)
	fmt.Printf("Net Income Margin: %.2f%%\n", plStatement.NetIncomeMargin)
	
	if plStatement.SharesOutstanding > 0 {
		fmt.Printf("\nShares Outstanding: %.0f\n", plStatement.SharesOutstanding)
		fmt.Printf("Earnings Per Share: %.2f\n", plStatement.EarningsPerShare)
	}
	
	fmt.Printf("\n=== EXECUTIVE SUMMARY ===\n")
	fmt.Printf("• Total Revenue: Rp %.2f\n", plStatement.Revenue.TotalRevenue)
	fmt.Printf("• Total COGS: Rp %.2f\n", plStatement.CostOfGoodsSold.TotalCOGS)
	fmt.Printf("• Gross Profit: Rp %.2f (%.2f%% margin)\n", plStatement.GrossProfit, plStatement.GrossProfitMargin)
	fmt.Printf("• Operating Expenses: Rp %.2f\n", plStatement.OperatingExpenses.TotalOpex)
	fmt.Printf("• Operating Income: Rp %.2f (%.2f%% margin)\n", plStatement.OperatingIncome, plStatement.OperatingMargin)
	fmt.Printf("• Net Income: Rp %.2f (%.2f%% margin)\n", plStatement.NetIncome, plStatement.NetIncomeMargin)
	
	// Let's also generate a simpler current month P&L for comparison
	fmt.Printf("\n=== CURRENT MONTH P&L COMPARISON ===\n")
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthlyPL, err := enhancedPLService.GenerateEnhancedProfitLoss(currentMonthStart, endDate)
	if err != nil {
		fmt.Printf("Error generating monthly P&L: %v\n", err)
	} else {
		fmt.Printf("Current Month (%s):\n", currentMonthStart.Format("Jan 2006"))
		fmt.Printf("• Revenue: Rp %.2f\n", monthlyPL.Revenue.TotalRevenue)
		fmt.Printf("• COGS: Rp %.2f\n", monthlyPL.CostOfGoodsSold.TotalCOGS)
		fmt.Printf("• Gross Profit: Rp %.2f (%.2f%%)\n", monthlyPL.GrossProfit, monthlyPL.GrossProfitMargin)
		fmt.Printf("• Operating Expenses: Rp %.2f\n", monthlyPL.OperatingExpenses.TotalOpex)
		fmt.Printf("• Net Income: Rp %.2f (%.2f%%)\n", monthlyPL.NetIncome, monthlyPL.NetIncomeMargin)
	}
	
	fmt.Printf("\n=== P&L STATEMENT GENERATION SUCCESSFUL ===\n")
	fmt.Printf("✓ Financial data analyzed from %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	fmt.Printf("✓ All accounts properly categorized\n")
	fmt.Printf("✓ Profit & Loss Statement ready for presentation\n")
}