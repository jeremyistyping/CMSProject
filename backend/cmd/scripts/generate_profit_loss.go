package main

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"encoding/json"
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
	
	fmt.Printf("\n=== GENERATING PROFIT & LOSS STATEMENT ===\n")
	
	// Define date range - current month
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := now
	
	fmt.Printf("Period: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	// Generate P&L statement
	plStatement, err := enhancedPLService.GenerateEnhancedProfitLoss(startDate, endDate)
	if err != nil {
		fmt.Printf("Error generating P&L statement: %v\n", err)
		return
	}
	
	fmt.Printf("\n=== PROFIT & LOSS STATEMENT ===\n")
	fmt.Printf("Company: %s\n", plStatement.Company.Name)
	fmt.Printf("Period: %s to %s\n", 
		plStatement.StartDate.Format("2006-01-02"), 
		plStatement.EndDate.Format("2006-01-02"))
	fmt.Printf("Currency: %s\n", plStatement.Currency)
	
	fmt.Printf("\n--- REVENUE ---\n")
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
	
	fmt.Printf("\n--- COST OF GOODS SOLD ---\n")
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
	
	fmt.Printf("\n--- PROFITABILITY METRICS ---\n")
	fmt.Printf("GROSS PROFIT: %.2f\n", plStatement.GrossProfit)
	fmt.Printf("Gross Profit Margin: %.2f%%\n", plStatement.GrossProfitMargin)
	
	fmt.Printf("\n--- OPERATING EXPENSES ---\n")
	fmt.Printf("Administrative Expenses: %.2f\n", plStatement.OperatingExpenses.Administrative.Subtotal)
	if len(plStatement.OperatingExpenses.Administrative.Items) > 0 {
		for _, item := range plStatement.OperatingExpenses.Administrative.Items {
			fmt.Printf("  - %s (%s): %.2f (%.2f%%)\n", 
				item.Name, item.Code, item.Amount, item.Percentage)
		}
	}
	
	fmt.Printf("Selling & Marketing Expenses: %.2f\n", plStatement.OperatingExpenses.SellingMarketing.Subtotal)
	if len(plStatement.OperatingExpenses.SellingMarketing.Items) > 0 {
		for _, item := range plStatement.OperatingExpenses.SellingMarketing.Items {
			fmt.Printf("  - %s (%s): %.2f (%.2f%%)\n", 
				item.Name, item.Code, item.Amount, item.Percentage)
		}
	}
	
	fmt.Printf("General Expenses: %.2f\n", plStatement.OperatingExpenses.General.Subtotal)
	if len(plStatement.OperatingExpenses.General.Items) > 0 {
		for _, item := range plStatement.OperatingExpenses.General.Items {
			fmt.Printf("  - %s (%s): %.2f (%.2f%%)\n", 
				item.Name, item.Code, item.Amount, item.Percentage)
		}
	}
	
	fmt.Printf("Depreciation & Amortization: %.2f\n", plStatement.OperatingExpenses.Depreciation.Subtotal)
	if len(plStatement.OperatingExpenses.Depreciation.Items) > 0 {
		for _, item := range plStatement.OperatingExpenses.Depreciation.Items {
			fmt.Printf("  - %s (%s): %.2f (%.2f%%)\n", 
				item.Name, item.Code, item.Amount, item.Percentage)
		}
	}
	
	fmt.Printf("\nTOTAL OPERATING EXPENSES: %.2f\n", plStatement.OperatingExpenses.TotalOpex)
	
	fmt.Printf("\n--- OPERATING PERFORMANCE ---\n")
	fmt.Printf("OPERATING INCOME (EBIT): %.2f\n", plStatement.OperatingIncome)
	fmt.Printf("Operating Margin: %.2f%%\n", plStatement.OperatingMargin)
	fmt.Printf("EBITDA: %.2f\n", plStatement.EBITDA)
	fmt.Printf("EBITDA Margin: %.2f%%\n", plStatement.EBITDAMargin)
	
	fmt.Printf("\n--- NON-OPERATING ITEMS ---\n")
	fmt.Printf("Interest Income: %.2f\n", plStatement.OtherIncomeExpense.InterestIncome.Subtotal)
	fmt.Printf("Interest Expense: %.2f\n", plStatement.OtherIncomeExpense.InterestExpense.Subtotal)
	fmt.Printf("Other Income: %.2f\n", plStatement.OtherIncomeExpense.OtherIncome.Subtotal)
	fmt.Printf("Other Expense: %.2f\n", plStatement.OtherIncomeExpense.OtherExpense.Subtotal)
	fmt.Printf("Net Other Income: %.2f\n", plStatement.OtherIncomeExpense.NetOtherIncome)
	
	fmt.Printf("\n--- FINAL RESULTS ---\n")
	fmt.Printf("INCOME BEFORE TAX: %.2f\n", plStatement.IncomeBeforeTax)
	fmt.Printf("Tax Expense: %.2f\n", plStatement.TaxExpense)
	fmt.Printf("Tax Rate: %.2f%%\n", plStatement.TaxRate)
	fmt.Printf("NET INCOME: %.2f\n", plStatement.NetIncome)
	fmt.Printf("Net Income Margin: %.2f%%\n", plStatement.NetIncomeMargin)
	
	if plStatement.SharesOutstanding > 0 {
		fmt.Printf("\nShares Outstanding: %.0f\n", plStatement.SharesOutstanding)
		fmt.Printf("Earnings Per Share: %.2f\n", plStatement.EarningsPerShare)
	}
	
	fmt.Printf("\n=== P&L STATEMENT SUMMARY ===\n")
	fmt.Printf("Total Revenue: %.2f\n", plStatement.Revenue.TotalRevenue)
	fmt.Printf("Total COGS: %.2f\n", plStatement.CostOfGoodsSold.TotalCOGS)
	fmt.Printf("Gross Profit: %.2f (%.2f%%)\n", plStatement.GrossProfit, plStatement.GrossProfitMargin)
	fmt.Printf("Total Operating Expenses: %.2f\n", plStatement.OperatingExpenses.TotalOpex)
	fmt.Printf("Operating Income: %.2f (%.2f%%)\n", plStatement.OperatingIncome, plStatement.OperatingMargin)
	fmt.Printf("Net Income: %.2f (%.2f%%)\n", plStatement.NetIncome, plStatement.NetIncomeMargin)
	
	fmt.Printf("\n=== JSON EXPORT ===\n")
	jsonData, err := json.MarshalIndent(plStatement, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
	} else {
		limit := min(1000, len(jsonData))
		fmt.Printf("P&L Statement JSON (first 1000 chars):\n%s...\n", string(jsonData[:limit]))
	}
	
	fmt.Printf("\n=== P&L GENERATION COMPLETE ===\n")
	fmt.Printf("✓ Enhanced Profit & Loss Statement generated successfully\n")
	fmt.Printf("✓ All financial metrics calculated\n")
	fmt.Printf("✓ Report ready for use\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}