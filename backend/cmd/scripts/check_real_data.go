package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
)

func main() {
	fmt.Println("🔍 Checking Real vs Synthetic SSOT Journal Data...")
	fmt.Println("==================================================")
	
	// Load config and connect
	cfg := config.LoadConfig()
	if cfg.Environment == "production" {
		log.Fatal("⚠️ This should not be run in production!")
	}
	
	db := database.ConnectDB()
	
	// Clean up any synthetic data first
	fmt.Println("🧹 Cleaning up synthetic test data...")
	db.Exec("DELETE FROM unified_journal_lines WHERE journal_id IN (SELECT id FROM unified_journal_ledger WHERE description = 'Synthetic Cash Injection Test')")
	db.Exec("DELETE FROM unified_journal_ledger WHERE description = 'Synthetic Cash Injection Test'")
	
	// Check real SSOT journal data
	fmt.Println("\n📊 Checking REAL SSOT Journal Entries...")
	var entries []struct {
		ID          uint64    `json:"id"`
		EntryNumber string    `json:"entry_number"`
		EntryDate   time.Time `json:"entry_date"`
		Description string    `json:"description"`
		TotalDebit  float64   `json:"total_debit"`
		TotalCredit float64   `json:"total_credit"`
		Status      string    `json:"status"`
		SourceType  string    `json:"source_type"`
	}
	
	db.Raw(`
		SELECT id, entry_number, entry_date, description, total_debit, total_credit, status, source_type
		FROM unified_journal_ledger 
		WHERE status = 'POSTED' 
		ORDER BY entry_date DESC, id DESC
		LIMIT 20
	`).Scan(&entries)
	
	totalReal := 0.0
	fmt.Printf("Found %d real POSTED journal entries:\n", len(entries))
	for i, entry := range entries {
		fmt.Printf("%d. [%s] %s - %s (Debit: %.0f, Credit: %.0f) - %s\n", 
			i+1, entry.EntryDate.Format("2006-01-02"), entry.EntryNumber, 
			entry.Description, entry.TotalDebit, entry.TotalCredit, entry.SourceType)
		totalReal += entry.TotalDebit
	}
	
	fmt.Printf("\n💰 Total Real Debits: IDR %.0f\n", totalReal)
	
	// Check cash account balances
	fmt.Println("\n🏦 Current Cash Account Balances...")
	var cashAccounts []struct {
		ID      uint64  `json:"id"`
		Code    string  `json:"code"`
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
	}
	
	db.Raw(`
		SELECT a.id, a.code, a.name, 
		       COALESCE(a.balance, 0) as balance
		FROM accounts a
		WHERE a.code LIKE '110%' 
		  AND a.is_active = true
		ORDER BY a.code
	`).Scan(&cashAccounts)
	
	totalCashBalance := 0.0
	for _, acc := range cashAccounts {
		fmt.Printf("  %s - %s: IDR %.0f\n", acc.Code, acc.Name, acc.Balance)
		totalCashBalance += acc.Balance
	}
	fmt.Printf("📊 Total Cash Balance: IDR %.0f\n", totalCashBalance)
	
	// Now generate SSOT Cash Flow with ONLY real data
	fmt.Println("\n🔄 Testing SSOT Cash Flow with REAL DATA ONLY...")
	cashFlowService := services.NewSSOTCashFlowService(db)
	
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)
	
	cashFlowData, err := cashFlowService.GenerateSSOTCashFlow(
		startDate.Format("2006-01-02"), 
		endDate.Format("2006-01-02"))
	if err != nil {
		log.Printf("❌ Error: %v", err)
		return
	}
	
	fmt.Printf("\n📈 REAL Cash Flow Results:\n")
	fmt.Printf("  Net Cash Flow: IDR %.0f\n", cashFlowData.NetCashFlow)
	fmt.Printf("  Cash at Beginning: IDR %.0f\n", cashFlowData.CashAtBeginning) 
	fmt.Printf("  Cash at End: IDR %.0f\n", cashFlowData.CashAtEnd)
	fmt.Printf("  Operating Activities: IDR %.0f\n", cashFlowData.OperatingActivities.TotalOperatingCashFlow)
	fmt.Printf("  Financing Activities: IDR %.0f\n", cashFlowData.FinancingActivities.TotalFinancingCashFlow)
	
	// Check account details breakdown
	fmt.Printf("\n🔍 Account Details (%d accounts):\n", len(cashFlowData.AccountDetails))
	for _, detail := range cashFlowData.AccountDetails {
		if detail.NetBalance != 0 {
			fmt.Printf("  %s - %s: IDR %.0f\n", detail.AccountCode, detail.AccountName, detail.NetBalance)
		}
	}
	
	// Summary
	fmt.Println("\n📋 Summary:")
	if cashFlowData.NetCashFlow == 0 {
		fmt.Println("❌ Cash Flow is ZERO - this might indicate:")
		fmt.Println("   - No real transactions in SSOT journal")
		fmt.Println("   - Account mapping issues") 
		fmt.Println("   - Date range problems")
	} else {
		fmt.Printf("✅ Cash Flow shows real data: IDR %.0f\n", cashFlowData.NetCashFlow)
	}
	
	if totalCashBalance != cashFlowData.CashAtEnd {
		fmt.Printf("⚠️  Account Balance (%.0f) vs Cash Flow End (%.0f) mismatch\n", 
			totalCashBalance, cashFlowData.CashAtEnd)
	} else {
		fmt.Println("✅ Account balances match cash flow calculations")
	}
}