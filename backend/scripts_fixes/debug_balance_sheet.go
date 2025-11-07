package main

import (
	"fmt"
	"log"
	"os"
	
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
)

func main() {
	fmt.Println("=== SSOT Balance Sheet Debug Test ===")
	
	// Initialize database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}
	
	fmt.Println("✅ Connected to database successfully")
	
	// Initialize SSOT Balance Sheet service
	balanceSheetService := services.NewSSOTBalanceSheetService(db)
	
	// Test date - use current date
	testDate := "2025-01-15"
	fmt.Printf("🔍 Testing Balance Sheet for date: %s\n", testDate)
	
	// Generate Balance Sheet
	result, err := balanceSheetService.GenerateSSOTBalanceSheet(testDate)
	if err != nil {
		fmt.Printf("❌ Error generating balance sheet: %v\n", err)
		os.Exit(1)
	}
	
	// Print results
	fmt.Printf("\n📊 Balance Sheet Results:\n")
	fmt.Printf("- Generated at: %s\n", result.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("- As of date: %s\n", result.AsOfDate.Format("2006-01-02"))
	fmt.Printf("- Total Assets: %.2f\n", result.Assets.TotalAssets)
	fmt.Printf("- Total Liabilities: %.2f\n", result.Liabilities.TotalLiabilities)
	fmt.Printf("- Total Equity: %.2f\n", result.Equity.TotalEquity)
	fmt.Printf("- Is Balanced: %v\n", result.IsBalanced)
	fmt.Printf("- Balance Difference: %.2f\n", result.BalanceDifference)
	
	// Current Assets breakdown
	fmt.Printf("\n💰 Current Assets:\n")
	fmt.Printf("- Cash: %.2f\n", result.Assets.CurrentAssets.Cash)
	fmt.Printf("- Receivables: %.2f\n", result.Assets.CurrentAssets.Receivables)
	fmt.Printf("- Inventory: %.2f\n", result.Assets.CurrentAssets.Inventory)
	fmt.Printf("- Total Current Assets: %.2f\n", result.Assets.CurrentAssets.TotalCurrentAssets)
	fmt.Printf("- Items count: %d\n", len(result.Assets.CurrentAssets.Items))
	
	// Print first few current asset items
	if len(result.Assets.CurrentAssets.Items) > 0 {
		fmt.Printf("\n📋 Current Asset Items:\n")
		for i, item := range result.Assets.CurrentAssets.Items {
			if i < 5 { // Show first 5 items
				fmt.Printf("  - %s (%s): %.2f\n", item.AccountName, item.AccountCode, item.Amount)
			}
		}
	} else {
		fmt.Printf("\n⚠️  No current asset items found\n")
	}
	
	// Account Details
	if len(result.AccountDetails) > 0 {
		fmt.Printf("\n📋 Account Details (first 5):\n")
		for i, detail := range result.AccountDetails {
			if i < 5 {
				fmt.Printf("  - %s (%s): Debit=%.2f, Credit=%.2f, Net=%.2f\n", 
					detail.AccountName, detail.AccountCode, detail.DebitTotal, detail.CreditTotal, detail.NetBalance)
			}
		}
	} else {
		fmt.Printf("\n⚠️  No account details found\n")
	}
	
	fmt.Println("\n✅ Balance Sheet test completed")
}