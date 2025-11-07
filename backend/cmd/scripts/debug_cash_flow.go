package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/services"
)

func main() {
	fmt.Println("🔍 Debugging Cash Flow Statement Calculation...")
	fmt.Println("============================================================")

	// Initialize database
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create service
	cfService := services.NewSSOTCashFlowService(db)

	// Use a recent period (last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0)

	fmt.Printf("\n📅 Period: %s to %s\n\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Generate cash flow
	cfData, err := cfService.GenerateSSOTCashFlow(startDate, endDate)
	if err != nil {
		log.Fatalf("Failed to generate cash flow: %v", err)
	}

	// Print detailed breakdown
	fmt.Println("🏢 OPERATING ACTIVITIES:")
	fmt.Println("----------------------------------------")
	fmt.Printf("Net Income: %.2f\n", cfData.OperatingActivities.NetIncome)
	
	fmt.Println("\n📊 Adjustments:")
	fmt.Printf("  Depreciation: %.2f\n", cfData.OperatingActivities.Adjustments.Depreciation)
	fmt.Printf("  Amortization: %.2f\n", cfData.OperatingActivities.Adjustments.Amortization)
	fmt.Printf("  Bad Debt: %.2f\n", cfData.OperatingActivities.Adjustments.BadDebtExpense)
	fmt.Printf("  Other: %.2f\n", cfData.OperatingActivities.Adjustments.OtherNonCashItems)
	fmt.Printf("  Total Adjustments: %.2f\n", cfData.OperatingActivities.Adjustments.TotalAdjustments)
	
	fmt.Println("\n🔄 Working Capital Changes:")
	fmt.Printf("  AR Change: %.2f\n", cfData.OperatingActivities.WorkingCapitalChanges.AccountsReceivableChange)
	fmt.Printf("  Inventory Change: %.2f\n", cfData.OperatingActivities.WorkingCapitalChanges.InventoryChange)
	fmt.Printf("  Prepaid Change: %.2f\n", cfData.OperatingActivities.WorkingCapitalChanges.PrepaidExpensesChange)
	fmt.Printf("  AP Change: %.2f\n", cfData.OperatingActivities.WorkingCapitalChanges.AccountsPayableChange)
	fmt.Printf("  Accrued Change: %.2f\n", cfData.OperatingActivities.WorkingCapitalChanges.AccruedLiabilitiesChange)
	fmt.Printf("  Other WC Change: %.2f\n", cfData.OperatingActivities.WorkingCapitalChanges.OtherWorkingCapitalChange)
	fmt.Printf("  Total WC Changes: %.2f\n", cfData.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges)
	
	fmt.Println("\n📋 Working Capital Items Details:")
	for i, item := range cfData.OperatingActivities.WorkingCapitalChanges.Items {
		fmt.Printf("  %d. %s (%s): %.2f [%s]\n", 
			i+1, item.AccountName, item.AccountCode, item.Amount, item.Type)
	}
	
	fmt.Printf("\n💰 TOTAL OPERATING CASH FLOW: %.2f\n", cfData.OperatingActivities.TotalOperatingCashFlow)
	
	fmt.Println("\n\n🏦 FINANCING ACTIVITIES:")
	fmt.Println("----------------------------------------")
	fmt.Printf("Share Capital Increase: %.2f\n", cfData.FinancingActivities.ShareCapitalIncrease)
	fmt.Printf("Total Financing Cash Flow: %.2f\n", cfData.FinancingActivities.TotalFinancingCashFlow)
	
	fmt.Println("\n\n💵 CASH SUMMARY:")
	fmt.Println("----------------------------------------")
	fmt.Printf("Beginning Cash: %.2f\n", cfData.CashAtBeginning)
	fmt.Printf("Net Cash Flow: %.2f\n", cfData.NetCashFlow)
	fmt.Printf("Ending Cash: %.2f\n", cfData.CashAtEnd)
	
	fmt.Println("\n\n📄 JSON Output (for API verification):")
	fmt.Println("----------------------------------------")
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"operating_activities": map[string]interface{}{
			"net_income": cfData.OperatingActivities.NetIncome,
			"working_capital_changes": map[string]interface{}{
				"total_working_capital_changes": cfData.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges,
				"items_count": len(cfData.OperatingActivities.WorkingCapitalChanges.Items),
			},
			"total_operating_cash_flow": cfData.OperatingActivities.TotalOperatingCashFlow,
		},
	}, "", "  ")
	fmt.Println(string(jsonData))
}
