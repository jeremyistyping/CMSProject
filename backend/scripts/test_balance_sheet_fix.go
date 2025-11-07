package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
	"fmt"
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Connect to database
	dsn := "accounting_user:Bismillah2024!@tcp(localhost:3306)/accounting_system?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔧 BALANCE SHEET FIX TEST")
	fmt.Println("=========================")

	// Create balance sheet service
	bsService := services.NewSSOTBalanceSheetService(db)
	
	// Generate balance sheet with the fix
	asOfDate := "2025-09-22"
	balanceSheet, err := bsService.GenerateSSOTBalanceSheet(asOfDate)
	if err != nil {
		log.Printf("❌ Error generating balance sheet: %v", err)
		return
	}

	fmt.Printf("\n📊 BALANCE SHEET RESULTS (as of %s):\n", asOfDate)
	fmt.Println("==========================================")
	
	// Assets
	fmt.Printf("ASSETS:\n")
	fmt.Printf("  Current Assets:\n")
	fmt.Printf("    Cash:                 %.2f\n", balanceSheet.Assets.CurrentAssets.Cash)
	fmt.Printf("    Receivables:          %.2f\n", balanceSheet.Assets.CurrentAssets.Receivables)
	fmt.Printf("    Other Current Assets: %.2f\n", balanceSheet.Assets.CurrentAssets.OtherCurrentAssets)
	fmt.Printf("  TOTAL ASSETS:           %.2f\n", balanceSheet.Assets.TotalAssets)
	
	fmt.Println()
	
	// Liabilities
	fmt.Printf("LIABILITIES:\n")
	fmt.Printf("  Current Liabilities:\n")
	fmt.Printf("    Accounts Payable:     %.2f\n", balanceSheet.Liabilities.CurrentLiabilities.AccountsPayable)
	fmt.Printf("    Tax Payable:          %.2f\n", balanceSheet.Liabilities.CurrentLiabilities.TaxPayable)
	fmt.Printf("    Other Current Liab:   %.2f\n", balanceSheet.Liabilities.CurrentLiabilities.OtherCurrentLiabilities)
	fmt.Printf("  TOTAL LIABILITIES:      %.2f\n", balanceSheet.Liabilities.TotalLiabilities)
	
	fmt.Println()
	
	// Equity
	fmt.Printf("EQUITY:\n")
	fmt.Printf("  Share Capital:          %.2f\n", balanceSheet.Equity.ShareCapital)
	fmt.Printf("  Retained Earnings:      %.2f  ← SHOULD INCLUDE NET INCOME\n", balanceSheet.Equity.RetainedEarnings)
	fmt.Printf("  Other Equity:           %.2f\n", balanceSheet.Equity.OtherEquity)
	fmt.Printf("  TOTAL EQUITY:           %.2f\n", balanceSheet.Equity.TotalEquity)
	
	fmt.Println()
	fmt.Printf("TOTAL LIABILITIES + EQUITY: %.2f\n", balanceSheet.TotalLiabilitiesAndEquity)
	
	fmt.Println()
	fmt.Println("🔍 BALANCE CHECK:")
	fmt.Println("==================")
	
	balance_diff := balanceSheet.Assets.TotalAssets - balanceSheet.TotalLiabilitiesAndEquity
	fmt.Printf("Assets:           %.2f\n", balanceSheet.Assets.TotalAssets)
	fmt.Printf("Liabilities + Eq: %.2f\n", balanceSheet.TotalLiabilitiesAndEquity)
	fmt.Printf("Difference:       %.2f\n", balance_diff)
	
	if balanceSheet.IsBalanced {
		fmt.Println("✅ BALANCE SHEET IS BALANCED!")
	} else {
		fmt.Printf("❌ BALANCE SHEET NOT BALANCED (Diff: %.2f)\n", balanceSheet.BalanceDifference)
	}
	
	fmt.Println()
	fmt.Println("📋 EQUITY LINE ITEMS:")
	fmt.Println("=====================")
	for _, item := range balanceSheet.Equity.Items {
		fmt.Printf("  %s - %s: %.2f\n", item.AccountCode, item.AccountName, item.Amount)
	}
	
	// Test net income calculation separately
	fmt.Println()
	fmt.Println("💰 NET INCOME VERIFICATION:")
	fmt.Println("============================")
	
	// Check revenue accounts
	var revenue float64
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'REVENUE'`).Scan(&revenue)
	fmt.Printf("Total Revenue:    %.2f\n", revenue)
	
	// Check expense accounts  
	var expense float64
	db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EXPENSE'`).Scan(&expense)
	fmt.Printf("Total Expenses:   %.2f\n", expense)
	
	netIncome := revenue - expense
	fmt.Printf("Net Income:       %.2f\n", netIncome)
	
	fmt.Printf("\n✅ Net Income should be included in Retained Earnings: %.2f\n", balanceSheet.Equity.RetainedEarnings)
	
	fmt.Println()
	fmt.Println("🎯 EXPECTED OUTCOME:")
	fmt.Println("====================")
	fmt.Println("With Net Income included in Retained Earnings,")
	fmt.Println("the Balance Sheet should now be balanced!")
	
	if balanceSheet.IsBalanced {
		fmt.Println("🎉 SUCCESS: Balance Sheet Fix Working!")
	} else {
		fmt.Println("⚠️  Still needs more investigation...")
	}
}