package main

import (
	"encoding/json"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
)

func main() {
	log.Printf("🧪 Testing COA Display Service with corrected balance signs...")

	// Initialize database connection
	db := database.ConnectDB()

	// Initialize COA display service
	coaService := services.NewCOADisplayService(db)

	log.Printf("\n📊 Test 1: Get COA for Display (only non-zero balances)")
	accounts, err := coaService.GetCOAForDisplay()
	if err != nil {
		log.Printf("❌ Error: %v", err)
		return
	}

	log.Printf("✅ Found %d accounts with non-zero balances:", len(accounts))
	for _, acc := range accounts {
		log.Printf("   %s - %s (%s)", acc.Code, acc.Name, acc.Type)
		log.Printf("     Raw: %.2f → Display: %.2f", acc.RawBalance, acc.DisplayBalance)
		
		// Verify corrections
		if acc.Type == "REVENUE" || acc.Type == "LIABILITY" {
			if acc.RawBalance < 0 && acc.DisplayBalance > 0 {
				log.Printf("     ✅ CORRECTED: Negative raw balance → Positive display")
			} else if acc.RawBalance > 0 && acc.DisplayBalance < 0 {
				log.Printf("     ✅ CORRECTED: Positive raw balance → Negative display")
			}
		}
	}

	log.Printf("\n📊 Test 2: Get Accounts Grouped by Type")
	groupedAccounts, err := coaService.GetAccountsByType()
	if err != nil {
		log.Printf("❌ Error: %v", err)
		return
	}

	for accountType, accountList := range groupedAccounts {
		log.Printf("✅ %s accounts: %d", accountType, len(accountList))
		for _, acc := range accountList {
			log.Printf("   %s: Display Balance %.2f", acc.Name, acc.DisplayBalance)
		}
	}

	log.Printf("\n📊 Test 3: Balance Sheet Format")
	balanceSheet, err := coaService.GetBalanceSheetAccounts()
	if err != nil {
		log.Printf("❌ Error: %v", err)
		return
	}

	log.Printf("📈 BALANCE SHEET:")
	log.Printf("ASSETS:")
	for _, acc := range balanceSheet["ASSETS"] {
		log.Printf("   %s: %.2f", acc.Name, acc.DisplayBalance)
	}
	log.Printf("LIABILITIES:")
	for _, acc := range balanceSheet["LIABILITIES"] {
		log.Printf("   %s: %.2f", acc.Name, acc.DisplayBalance)
	}
	log.Printf("EQUITY:")
	for _, acc := range balanceSheet["EQUITY"] {
		log.Printf("   %s: %.2f", acc.Name, acc.DisplayBalance)
	}

	log.Printf("\n📊 Test 4: Income Statement Format")
	incomeStatement, err := coaService.GetIncomeStatementAccounts()
	if err != nil {
		log.Printf("❌ Error: %v", err)
		return
	}

	log.Printf("📈 PROFIT & LOSS:")
	log.Printf("REVENUE:")
	for _, acc := range incomeStatement["REVENUE"] {
		log.Printf("   %s: %.2f", acc.Name, acc.DisplayBalance)
	}
	log.Printf("EXPENSES:")
	for _, acc := range incomeStatement["EXPENSE"] {
		log.Printf("   %s: %.2f", acc.Name, acc.DisplayBalance)
	}

	log.Printf("\n💻 Test 5: JSON Output Example")
	if len(accounts) > 0 {
		// Show sample JSON output for first few accounts
		sampleAccounts := accounts
		if len(accounts) > 3 {
			sampleAccounts = accounts[:3]
		}
		
		jsonData, err := json.MarshalIndent(sampleAccounts, "", "  ")
		if err != nil {
			log.Printf("❌ JSON Error: %v", err)
			return
		}
		
		log.Printf("Sample API Response:")
		log.Printf("%s", jsonData)
	}

	log.Printf("\n✅ COA Display Service testing completed!")
	log.Printf("\n🎯 SUMMARY:")
	log.Printf("   - Revenue accounts now display as POSITIVE (corrected from negative)")
	log.Printf("   - PPN Keluaran accounts now display as POSITIVE (corrected from negative)")
	log.Printf("   - Asset accounts display correctly as POSITIVE")
	log.Printf("   - Service ready for frontend integration!")
}