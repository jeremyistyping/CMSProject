package main

import (
	"fmt"
	"log"
	"os"
	
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize database
	db := database.ConnectDB()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	fmt.Println("=== DEBUGGING BALANCE SHEET DIFFERENCE ===")
	fmt.Println()

	// Create SSOT Balance Sheet Service
	bsService := services.NewSSOTBalanceSheetService(db)

	// Generate balance sheet for 2025-12-31
	asOfDate := "2025-12-31"
	fmt.Printf("Generating Balance Sheet as of %s...\n\n", asOfDate)

	bsData, err := bsService.GenerateSSOTBalanceSheet(asOfDate)
	if err != nil {
		log.Fatalf("Failed to generate balance sheet: %v", err)
	}

	fmt.Println("\n=== BALANCE SHEET SUMMARY ===")
	fmt.Printf("Total Assets:              Rp %.2f\n", bsData.Assets.TotalAssets)
	fmt.Printf("Total Liabilities:         Rp %.2f\n", bsData.Liabilities.TotalLiabilities)
	fmt.Printf("Total Equity:              Rp %.2f\n", bsData.Equity.TotalEquity)
	fmt.Printf("Total Liabilities+Equity:  Rp %.2f\n", bsData.TotalLiabilitiesAndEquity)
	fmt.Printf("Difference:                Rp %.2f\n", bsData.BalanceDifference)
	fmt.Printf("Is Balanced:               %t\n", bsData.IsBalanced)

	fmt.Println("\n=== ASSETS BREAKDOWN ===")
	fmt.Printf("Current Assets:\n")
	fmt.Printf("  Cash:                    Rp %.2f\n", bsData.Assets.CurrentAssets.Cash)
	fmt.Printf("  Receivables:             Rp %.2f\n", bsData.Assets.CurrentAssets.Receivables)
	fmt.Printf("  Inventory:               Rp %.2f\n", bsData.Assets.CurrentAssets.Inventory)
	fmt.Printf("  Prepaid Expenses:        Rp %.2f\n", bsData.Assets.CurrentAssets.PrepaidExpenses)
	fmt.Printf("  Other Current Assets:    Rp %.2f\n", bsData.Assets.CurrentAssets.OtherCurrentAssets)
	fmt.Printf("  Total Current Assets:    Rp %.2f\n", bsData.Assets.CurrentAssets.TotalCurrentAssets)
	
	fmt.Printf("\nCurrent Assets Items:\n")
	for _, item := range bsData.Assets.CurrentAssets.Items {
		fmt.Printf("  %s - %s: Rp %.2f\n", item.AccountCode, item.AccountName, item.Amount)
	}

	fmt.Printf("\nNon-Current Assets:\n")
	fmt.Printf("  Fixed Assets:            Rp %.2f\n", bsData.Assets.NonCurrentAssets.FixedAssets)
	fmt.Printf("  Total Non-Current:       Rp %.2f\n", bsData.Assets.NonCurrentAssets.TotalNonCurrentAssets)

	fmt.Println("\n=== LIABILITIES BREAKDOWN ===")
	fmt.Printf("Current Liabilities:\n")
	fmt.Printf("  Accounts Payable:        Rp %.2f\n", bsData.Liabilities.CurrentLiabilities.AccountsPayable)
	fmt.Printf("  Short Term Debt:         Rp %.2f\n", bsData.Liabilities.CurrentLiabilities.ShortTermDebt)
	fmt.Printf("  Accrued Liabilities:     Rp %.2f\n", bsData.Liabilities.CurrentLiabilities.AccruedLiabilities)
	fmt.Printf("  Tax Payable:             Rp %.2f\n", bsData.Liabilities.CurrentLiabilities.TaxPayable)
	fmt.Printf("  Other Current:           Rp %.2f\n", bsData.Liabilities.CurrentLiabilities.OtherCurrentLiabilities)
	fmt.Printf("  Total Current:           Rp %.2f\n", bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities)

	fmt.Printf("\nCurrent Liabilities Items:\n")
	for _, item := range bsData.Liabilities.CurrentLiabilities.Items {
		fmt.Printf("  %s - %s: Rp %.2f\n", item.AccountCode, item.AccountName, item.Amount)
	}

	fmt.Println("\n=== EQUITY BREAKDOWN ===")
	fmt.Printf("Share Capital:             Rp %.2f\n", bsData.Equity.ShareCapital)
	fmt.Printf("Retained Earnings:         Rp %.2f\n", bsData.Equity.RetainedEarnings)
	fmt.Printf("Other Equity:              Rp %.2f\n", bsData.Equity.OtherEquity)
	fmt.Printf("Total Equity:              Rp %.2f\n", bsData.Equity.TotalEquity)

	fmt.Printf("\nEquity Items:\n")
	for _, item := range bsData.Equity.Items {
		fmt.Printf("  %s - %s: Rp %.2f\n", item.AccountCode, item.AccountName, item.Amount)
	}

	// Manual calculation verification
	fmt.Println("\n=== MANUAL CALCULATION VERIFICATION ===")
	manualAssets := bsData.Assets.CurrentAssets.TotalCurrentAssets + bsData.Assets.NonCurrentAssets.TotalNonCurrentAssets
	manualLiabEq := bsData.Liabilities.TotalLiabilities + bsData.Equity.TotalEquity
	manualDiff := manualAssets - manualLiabEq

	fmt.Printf("Manual Assets Total:       Rp %.2f\n", manualAssets)
	fmt.Printf("Manual Liab+Eq Total:      Rp %.2f\n", manualLiabEq)
	fmt.Printf("Manual Difference:         Rp %.2f\n", manualDiff)

	// Check subcategory totals
	fmt.Println("\n=== SUBCATEGORY TOTALS CHECK ===")
	calcCurrentLiab := bsData.Liabilities.CurrentLiabilities.AccountsPayable +
		bsData.Liabilities.CurrentLiabilities.ShortTermDebt +
		bsData.Liabilities.CurrentLiabilities.AccruedLiabilities +
		bsData.Liabilities.CurrentLiabilities.TaxPayable +
		bsData.Liabilities.CurrentLiabilities.OtherCurrentLiabilities
	
	fmt.Printf("Calculated Current Liabilities: Rp %.2f\n", calcCurrentLiab)
	fmt.Printf("Stored Current Liabilities:     Rp %.2f\n", bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities)
	fmt.Printf("Subcategory Diff:               Rp %.2f\n", calcCurrentLiab - bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities)

	// Check if difference is related to PPN
	fmt.Println("\n=== PPN ACCOUNTS CHECK ===")
	var ppnMasukan, ppnKeluaran, ppnNet float64
	for _, item := range bsData.Assets.CurrentAssets.Items {
		if item.AccountCode == "PPN_NET" {
			ppnNet = item.Amount
			fmt.Printf("Found PPN_NET in Assets: Rp %.2f\n", item.Amount)
		}
	}
	for _, item := range bsData.Liabilities.CurrentLiabilities.Items {
		if item.AccountCode == "PPN_NET" {
			ppnNet = item.Amount
			fmt.Printf("Found PPN_NET in Liabilities: Rp %.2f\n", item.Amount)
		}
	}

	// Query raw PPN account balances from DB
	type PPNBalance struct {
		Code    string
		Name    string
		Balance float64
	}
	var ppnAccounts []PPNBalance
	query := `
		SELECT code, name, COALESCE(balance, 0) as balance
		FROM accounts
		WHERE (LOWER(name) LIKE '%ppn%' OR code LIKE '115%' OR code LIKE '215%')
		  AND is_active = true
		  AND COALESCE(is_header, false) = false
		ORDER BY code
	`
	if err := db.Raw(query).Scan(&ppnAccounts).Error; err == nil {
		fmt.Println("\nRaw PPN Account Balances from DB:")
		for _, acc := range ppnAccounts {
			fmt.Printf("  %s - %s: Rp %.2f\n", acc.Code, acc.Name, acc.Balance)
			if acc.Balance < 0 {
				ppnKeluaran += -acc.Balance
			} else {
				ppnMasukan += acc.Balance
			}
		}
		fmt.Printf("\nTotal PPN Masukan (from DB): Rp %.2f\n", ppnMasukan)
		fmt.Printf("Total PPN Keluaran (from DB): Rp %.2f\n", ppnKeluaran)
		fmt.Printf("Net PPN (Keluaran - Masukan): Rp %.2f\n", ppnKeluaran - ppnMasukan)
		fmt.Printf("PPN_NET shown in report: Rp %.2f\n", ppnNet)
	}

	fmt.Println("\n=== POTENTIAL CAUSES OF DIFFERENCE ===")
	if manualDiff > 0 {
		fmt.Printf("Assets are HIGHER by Rp %.2f\n", manualDiff)
		fmt.Println("Possible causes:")
		fmt.Println("  1. Some assets are counted but not classified properly")
		fmt.Println("  2. PPN Masukan adjustment error")
		fmt.Println("  3. Missing liability or equity accounts")
	} else if manualDiff < 0 {
		fmt.Printf("Liabilities+Equity are HIGHER by Rp %.2f\n", -manualDiff)
		fmt.Println("Possible causes:")
		fmt.Println("  1. Some liabilities/equity are double counted")
		fmt.Println("  2. PPN Keluaran adjustment error")
		fmt.Println("  3. Missing asset accounts")
	}

	os.Exit(0)
}
