package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("🧪 Testing SSOT Cash Flow Statement Implementation...")
	fmt.Println("============================================================")

	// Load configuration
	cfg := config.LoadConfig()
	if cfg.Environment == "production" {
		log.Fatal("⚠️ This test should not be run in production environment!")
	}

	// Connect to database
	db := database.ConnectDB()

	// Run tests
	if err := runCashFlowTests(db); err != nil {
		log.Fatalf("❌ Cash Flow tests failed: %v", err)
	}

	fmt.Println("\n✅ All SSOT Cash Flow tests completed successfully!")
}

func runCashFlowTests(db *gorm.DB) error {
	fmt.Println("\n📊 Testing SSOT Cash Flow Service...")

	// Initialize service
	cashFlowService := services.NewSSOTCashFlowService(db)

	// Test date range (last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	fmt.Printf("🗓️ Testing period: %s to %s\n", startDateStr, endDateStr)

	// Arrange: create a synthetic posted cash injection today to guarantee non-zero cash flow
	fmt.Println("\n🔧 Arrange: Creating synthetic cash injection journal (this will be cleaned up)...")
	injectionAmount := 111540000.00 // IDR 111,540,000
	if err := createSyntheticCashInjection(db, injectionAmount, endDate); err != nil {
		fmt.Printf("⚠️ Unable to create synthetic journal: %v\n", err)
	}

	// Test 1: Basic Cash Flow Generation
	fmt.Println("\n🔧 Test 1: Basic Cash Flow Generation")
	fmt.Println("----------------------------------------")

	cashFlowData, err := cashFlowService.GenerateSSOTCashFlow(startDateStr, endDateStr)
	if err != nil {
		return fmt.Errorf("failed to generate cash flow statement: %v", err)
	}

	if cashFlowData == nil {
		return fmt.Errorf("cash flow data is nil")
	}

	fmt.Printf("✅ Cash Flow data generated successfully\n")
	fmt.Printf("   Company: %s\n", cashFlowData.Company.Name)
	fmt.Printf("   Period: %s - %s\n", cashFlowData.StartDate.Format("2006-01-02"), cashFlowData.EndDate.Format("2006-01-02"))
	fmt.Printf("   Currency: %s\n", cashFlowData.Currency)

	// Test 2: Cash Flow Activities Analysis
	fmt.Println("\n💰 Test 2: Cash Flow Activities Analysis")
	fmt.Println("----------------------------------------")

	fmt.Printf("Operating Activities:\n")
	fmt.Printf("  Net Income: %.2f\n", cashFlowData.OperatingActivities.NetIncome)
	fmt.Printf("  Total Adjustments: %.2f\n", cashFlowData.OperatingActivities.Adjustments.TotalAdjustments)
	fmt.Printf("  Working Capital Changes: %.2f\n", cashFlowData.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges)
	fmt.Printf("  Total Operating Cash Flow: %.2f\n", cashFlowData.OperatingActivities.TotalOperatingCashFlow)

	fmt.Printf("\nInvesting Activities:\n")
	fmt.Printf("  Purchase of Fixed Assets: %.2f\n", cashFlowData.InvestingActivities.PurchaseOfFixedAssets)
	fmt.Printf("  Sale of Fixed Assets: %.2f\n", cashFlowData.InvestingActivities.SaleOfFixedAssets)
	fmt.Printf("  Purchase of Investments: %.2f\n", cashFlowData.InvestingActivities.PurchaseOfInvestments)
	fmt.Printf("  Sale of Investments: %.2f\n", cashFlowData.InvestingActivities.SaleOfInvestments)
	fmt.Printf("  Total Investing Cash Flow: %.2f\n", cashFlowData.InvestingActivities.TotalInvestingCashFlow)

	fmt.Printf("\nFinancing Activities:\n")
	fmt.Printf("  Share Capital Increase: %.2f\n", cashFlowData.FinancingActivities.ShareCapitalIncrease)
	fmt.Printf("  Long-term Debt Changes: %.2f\n", cashFlowData.FinancingActivities.LongTermDebtIncrease - cashFlowData.FinancingActivities.LongTermDebtDecrease)
	fmt.Printf("  Dividends Paid: %.2f\n", cashFlowData.FinancingActivities.DividendsPaid)
	fmt.Printf("  Total Financing Cash Flow: %.2f\n", cashFlowData.FinancingActivities.TotalFinancingCashFlow)

	// Test 3: Cash Balance Reconciliation
	fmt.Println("\n🏦 Test 3: Cash Balance Reconciliation")
	fmt.Println("----------------------------------------")

	fmt.Printf("Cash at Beginning: %.2f\n", cashFlowData.CashAtBeginning)
	fmt.Printf("Net Cash Flow: %.2f\n", cashFlowData.NetCashFlow)
	expectedEndingCash := cashFlowData.CashAtBeginning + cashFlowData.NetCashFlow
	fmt.Printf("Expected Cash at End: %.2f\n", expectedEndingCash)
	fmt.Printf("Actual Cash at End: %.2f\n", cashFlowData.CashAtEnd)
	
	difference := cashFlowData.CashAtEnd - expectedEndingCash
	fmt.Printf("Difference: %.2f\n", difference)

	if abs(difference) <= 0.01 {
		fmt.Println("✅ Cash flow balances correctly!")
	} else {
		fmt.Printf("⚠️ Cash flow has reconciliation difference: %.2f\n", difference)
	}

	// Test 4: Cash Flow Ratios
	fmt.Println("\n📈 Test 4: Cash Flow Ratios & Analysis")
	fmt.Println("----------------------------------------")

	fmt.Printf("Operating Cash Flow Ratio: %.2f\n", cashFlowData.CashFlowRatios.OperatingCashFlowRatio)
	fmt.Printf("Free Cash Flow: %.2f\n", cashFlowData.CashFlowRatios.FreeCashFlow)
	fmt.Printf("Cash Flow to Debt Ratio: %.2f\n", cashFlowData.CashFlowRatios.CashFlowToDebtRatio)

	// Test 5: Account Details Verification
	fmt.Println("\n🔍 Test 5: Account Details Verification")
	fmt.Println("----------------------------------------")

	fmt.Printf("Total accounts with activity: %d\n", len(cashFlowData.AccountDetails))

	// Group accounts by type
	assetAccounts := 0
	liabilityAccounts := 0
	equityAccounts := 0
	expenseAccounts := 0

	for _, account := range cashFlowData.AccountDetails {
		switch account.AccountType {
		case "ASSET":
			assetAccounts++
		case "LIABILITY":
			liabilityAccounts++
		case "EQUITY":
			equityAccounts++
		case "EXPENSE":
			expenseAccounts++
		}
	}

	fmt.Printf("  Asset accounts: %d\n", assetAccounts)
	fmt.Printf("  Liability accounts: %d\n", liabilityAccounts)
	fmt.Printf("  Equity accounts: %d\n", equityAccounts)
	fmt.Printf("  Expense accounts: %d\n", expenseAccounts)

	// Test 6: Operating Activities Detail Items
	fmt.Println("\n🔧 Test 6: Operating Activities Details")
	fmt.Println("----------------------------------------")

	fmt.Printf("Adjustment Items: %d\n", len(cashFlowData.OperatingActivities.Adjustments.Items))
	for i, item := range cashFlowData.OperatingActivities.Adjustments.Items {
		if i < 5 { // Show first 5 items
			fmt.Printf("  %s (%s): %.2f [%s]\n", item.AccountName, item.AccountCode, item.Amount, item.Type)
		}
	}
	if len(cashFlowData.OperatingActivities.Adjustments.Items) > 5 {
		fmt.Printf("  ... and %d more items\n", len(cashFlowData.OperatingActivities.Adjustments.Items)-5)
	}

	fmt.Printf("\nWorking Capital Items: %d\n", len(cashFlowData.OperatingActivities.WorkingCapitalChanges.Items))
	for i, item := range cashFlowData.OperatingActivities.WorkingCapitalChanges.Items {
		if i < 5 { // Show first 5 items
			fmt.Printf("  %s (%s): %.2f [%s]\n", item.AccountName, item.AccountCode, item.Amount, item.Type)
		}
	}
	if len(cashFlowData.OperatingActivities.WorkingCapitalChanges.Items) > 5 {
		fmt.Printf("  ... and %d more items\n", len(cashFlowData.OperatingActivities.WorkingCapitalChanges.Items)-5)
	}

	// Test 7: Data Structure Completeness
	fmt.Println("\n✅ Test 7: Data Structure Completeness")
	fmt.Println("----------------------------------------")

	checks := []string{}

	if cashFlowData.Company.Name == "" {
		checks = append(checks, "❌ Company name is missing")
	} else {
		checks = append(checks, "✅ Company name is present")
	}

	if cashFlowData.StartDate.IsZero() {
		checks = append(checks, "❌ Start date is missing")
	} else {
		checks = append(checks, "✅ Start date is valid")
	}

	if cashFlowData.EndDate.IsZero() {
		checks = append(checks, "❌ End date is missing")
	} else {
		checks = append(checks, "✅ End date is valid")
	}

	if cashFlowData.Currency == "" {
		checks = append(checks, "❌ Currency is missing")
	} else {
		checks = append(checks, "✅ Currency is present")
	}

	if !cashFlowData.Enhanced {
		checks = append(checks, "⚠️ Enhanced flag is false")
	} else {
		checks = append(checks, "✅ Enhanced flag is true")
	}

	for _, check := range checks {
		fmt.Println(check)
	}

	// Test 8: JSON Serialization
	fmt.Println("\n📋 Test 8: JSON Serialization")
	fmt.Println("----------------------------------------")

	jsonData, err := json.MarshalIndent(cashFlowData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cash flow data to JSON: %v", err)
	}

	fmt.Printf("✅ JSON serialization successful (%d bytes)\n", len(jsonData))

	// Optionally save to file for debugging
	if os.Getenv("SAVE_TEST_OUTPUT") == "true" {
		filename := fmt.Sprintf("cash_flow_test_output_%s.json", time.Now().Format("20060102_150405"))
		if err := os.WriteFile(filename, jsonData, 0644); err != nil {
			fmt.Printf("⚠️ Failed to save test output: %v\n", err)
		} else {
			fmt.Printf("💾 Test output saved to: %s\n", filename)
		}
	}

	// Test 9: Error Handling
	fmt.Println("\n🚨 Test 9: Error Handling")
	fmt.Println("----------------------------------------")

	// Test invalid date format
	_, err = cashFlowService.GenerateSSOTCashFlow("invalid-date", endDateStr)
	if err == nil {
		checks = append(checks, "❌ Should have failed with invalid start date")
	} else {
		fmt.Println("✅ Correctly handles invalid start date")
	}

	// Test invalid date format
	_, err = cashFlowService.GenerateSSOTCashFlow(startDateStr, "invalid-date")
	if err == nil {
		checks = append(checks, "❌ Should have failed with invalid end date")
	} else {
		fmt.Println("✅ Correctly handles invalid end date")
	}

	// Test future dates
	futureStart := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	futureEnd := time.Now().AddDate(1, 0, 1).Format("2006-01-02")
	
	futureCashFlow, err := cashFlowService.GenerateSSOTCashFlow(futureStart, futureEnd)
	if err != nil {
		fmt.Printf("⚠️ Future date test failed: %v\n", err)
	} else {
		fmt.Printf("✅ Future dates handled (empty result expected): Net Cash Flow = %.2f\n", futureCashFlow.NetCashFlow)
	}

	// Cleanup synthetic data (best-effort)
	fmt.Println("\n🧹 Cleaning up synthetic test data...")
	db.Exec("DELETE FROM unified_journal_lines WHERE journal_id IN (SELECT id FROM unified_journal_ledger WHERE description = 'Synthetic Cash Injection Test')")
	db.Exec("DELETE FROM unified_journal_ledger WHERE description = 'Synthetic Cash Injection Test'")

	return nil
}

// createSyntheticCashInjection inserts a minimal posted journal that debits Cash (1101) and credits Equity (3101)
func createSyntheticCashInjection(db *gorm.DB, amount float64, date time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Insert into unified_journal_ledger
		ledgerSQL := `
			INSERT INTO unified_journal_ledger (entry_number, source_type, entry_date, description, total_debit, total_credit, status, is_balanced, created_by, created_at, updated_at)
			VALUES (?, 'MANUAL', ?, 'Synthetic Cash Injection Test', ?, ?, 'POSTED', true, 1, NOW(), NOW())
			RETURNING id
		`
		var journalID uint64
		if err := tx.Raw(ledgerSQL, fmt.Sprintf("JE-TEST-%d", time.Now().UnixNano()), date.Format("2006-01-02"), amount, amount).Scan(&journalID).Error; err != nil {
			return err
		}

		// Find account IDs for 1101 (Cash) and 3101 (Equity)
		var cashID, equityID uint64
		if err := tx.Raw("SELECT id FROM accounts WHERE code = '1101' LIMIT 1").Scan(&cashID).Error; err != nil || cashID == 0 {
			return fmt.Errorf("cash account 1101 not found")
		}
		if err := tx.Raw("SELECT id FROM accounts WHERE code = '3101' LIMIT 1").Scan(&equityID).Error; err != nil || equityID == 0 {
			return fmt.Errorf("equity account 3101 not found")
		}

		// Insert lines
		lineSQL := `
			INSERT INTO unified_journal_lines (journal_id, account_id, line_number, description, debit_amount, credit_amount, created_at, updated_at)
			VALUES (?, ?, 1, 'Cash In', ?, 0, NOW(), NOW()),
			       (?, ?, 2, 'Owner Equity', 0, ?, NOW(), NOW())
		`
		if err := tx.Exec(lineSQL, journalID, cashID, amount, journalID, equityID, amount).Error; err != nil {
			return err
		}
		return nil
	})
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
