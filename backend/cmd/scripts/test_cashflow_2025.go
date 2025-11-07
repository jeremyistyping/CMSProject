package main

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
	"fmt"
	"log"
)

func main() {
	fmt.Println("🧪 Testing Cash Flow for Period 2025-01-01 to 2025-12-31")
	fmt.Println("============================================================")

	// Load configuration
	cfg := config.LoadConfig()
	if cfg.Environment == "production" {
		log.Fatal("⚠️ This test should not be run in production environment!")
	}

	// Connect to database
	db := database.ConnectDB()

	// Create service
	cashFlowService := services.NewSSOTCashFlowService(db)

	// Generate cash flow
	cfData, err := cashFlowService.GenerateSSOTCashFlow("2025-01-01", "2025-12-31")
	if err != nil {
		log.Fatalf("Failed to generate cash flow: %v", err)
	}

	// Print results
	fmt.Println("\n💰 OPERATING ACTIVITIES")
	fmt.Println("----------------------------------------")
	fmt.Printf("Net Income: %.2f\n", cfData.OperatingActivities.NetIncome)
	fmt.Printf("Total Adjustments: %.2f\n", cfData.OperatingActivities.Adjustments.TotalAdjustments)
	fmt.Printf("Total Working Capital Changes: %.2f\n", cfData.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges)
	fmt.Printf("TOTAL OPERATING CASH FLOW: %.2f\n", cfData.OperatingActivities.TotalOperatingCashFlow)

	fmt.Println("\nWorking Capital Changes Detail:")
	for _, item := range cfData.OperatingActivities.WorkingCapitalChanges.Items {
		fmt.Printf("  %s (%s): %.2f [%s]\n", item.AccountName, item.AccountCode, item.Amount, item.Type)
	}

	fmt.Println("\n🏭 INVESTING ACTIVITIES")
	fmt.Println("----------------------------------------")
	fmt.Printf("TOTAL INVESTING CASH FLOW: %.2f\n", cfData.InvestingActivities.TotalInvestingCashFlow)

	fmt.Println("\n💵 FINANCING ACTIVITIES")
	fmt.Println("----------------------------------------")
	fmt.Printf("Share Capital Increase: %.2f\n", cfData.FinancingActivities.ShareCapitalIncrease)
	fmt.Printf("TOTAL FINANCING CASH FLOW: %.2f\n", cfData.FinancingActivities.TotalFinancingCashFlow)

	fmt.Println("\n📊 SUMMARY")
	fmt.Println("----------------------------------------")
	fmt.Printf("Cash at Beginning: %.2f\n", cfData.CashAtBeginning)
	fmt.Printf("Operating CF: %.2f\n", cfData.OperatingActivities.TotalOperatingCashFlow)
	fmt.Printf("Investing CF: %.2f\n", cfData.InvestingActivities.TotalInvestingCashFlow)
	fmt.Printf("Financing CF: %.2f\n", cfData.FinancingActivities.TotalFinancingCashFlow)
	fmt.Printf("NET CASH FLOW: %.2f\n", cfData.NetCashFlow)
	fmt.Printf("Cash at End: %.2f\n", cfData.CashAtEnd)

	fmt.Println("\n✅ Test completed!")

	// Verify calculation
	expectedOperating := cfData.OperatingActivities.NetIncome + 
		cfData.OperatingActivities.Adjustments.TotalAdjustments + 
		cfData.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges

	fmt.Println("\n🔍 VERIFICATION")
	fmt.Println("----------------------------------------")
	fmt.Printf("Expected Operating CF (Net Income + Adjustments + WC Changes): %.2f\n", expectedOperating)
	fmt.Printf("Actual Operating CF: %.2f\n", cfData.OperatingActivities.TotalOperatingCashFlow)
	fmt.Printf("Match: %v\n", expectedOperating == cfData.OperatingActivities.TotalOperatingCashFlow)
}
