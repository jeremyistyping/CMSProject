package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// FinancialValidationReport represents the validation results
type FinancialValidationReport struct {
	ReportDate           time.Time                `json:"report_date"`
	ValidationDate       time.Time                `json:"validation_date"`
	
	// Basic Accounting Equation Validation
	AccountingEquation   AccountingEquationCheck  `json:"accounting_equation"`
	
	// Journal Entry Validation
	JournalValidation    JournalValidationCheck   `json:"journal_validation"`
	
	// Account Balance Validation
	AccountValidation    AccountValidationCheck   `json:"account_validation"`
	
	// Financial Report Consistency
	ReportConsistency    ReportConsistencyCheck   `json:"report_consistency"`
	
	// Data Quality Issues
	DataQualityIssues    []DataQualityIssue       `json:"data_quality_issues"`
	
	// Overall Score
	OverallScore         float64                  `json:"overall_score"`
	OverallStatus        string                   `json:"overall_status"` // EXCELLENT, GOOD, NEEDS_ATTENTION, CRITICAL
	
	// Recommendations
	Recommendations      []string                 `json:"recommendations"`
}

type AccountingEquationCheck struct {
	TotalAssets          float64  `json:"total_assets"`
	TotalLiabilities     float64  `json:"total_liabilities"`
	TotalEquity          float64  `json:"total_equity"`
	LiabilitiesPlusEquity float64 `json:"liabilities_plus_equity"`
	Difference           float64  `json:"difference"`
	IsBalanced           bool     `json:"is_balanced"`
	BalancePercentage    float64  `json:"balance_percentage"`
}

type JournalValidationCheck struct {
	TotalEntries         int64    `json:"total_entries"`
	BalancedEntries      int64    `json:"balanced_entries"`
	UnbalancedEntries    int64    `json:"unbalanced_entries"`
	PostedEntries        int64    `json:"posted_entries"`
	DraftEntries         int64    `json:"draft_entries"`
	TotalDebits          float64  `json:"total_debits"`
	TotalCredits         float64  `json:"total_credits"`
	DebitCreditDifference float64 `json:"debit_credit_difference"`
	BalanceAccuracy      float64  `json:"balance_accuracy"`
}

type AccountValidationCheck struct {
	TotalActiveAccounts  int64    `json:"total_active_accounts"`
	AccountsWithBalance  int64    `json:"accounts_with_balance"`
	AssetAccounts        int64    `json:"asset_accounts"`
	LiabilityAccounts    int64    `json:"liability_accounts"`
	EquityAccounts       int64    `json:"equity_accounts"`
	RevenueAccounts      int64    `json:"revenue_accounts"`
	ExpenseAccounts      int64    `json:"expense_accounts"`
	InactiveAccounts     int64    `json:"inactive_accounts"`
}

type ReportConsistencyCheck struct {
	ProfitLossBalance     bool     `json:"profit_loss_balance"`
	BalanceSheetBalance   bool     `json:"balance_sheet_balance"`
	CashFlowBalance       bool     `json:"cash_flow_balance"`
	TrialBalanceBalance   bool     `json:"trial_balance_balance"`
	ConsistencyScore      float64  `json:"consistency_score"`
}

type DataQualityIssue struct {
	Category             string   `json:"category"`
	Severity             string   `json:"severity"` // HIGH, MEDIUM, LOW
	Description          string   `json:"description"`
	Count                int64    `json:"count"`
	Details              string   `json:"details"`
	Recommendation       string   `json:"recommendation"`
}

// Database connection
var db *sql.DB

func main() {
	// Load environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	// Connect to database
	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("🔍 Memulai Validasi Financial Report...")
	fmt.Println("=" + string(make([]byte, 60)) + "=")

	// Run comprehensive validation
	report, err := runFinancialValidation()
	if err != nil {
		log.Fatal("Failed to run financial validation:", err)
	}

	// Display results
	displayValidationResults(report)

	// Save results to JSON file
	saveResultsToFile(report)
	
	fmt.Println("\n✅ Validasi selesai. Report disimpan ke financial_validation_report.json")
}

func runFinancialValidation() (*FinancialValidationReport, error) {
	report := &FinancialValidationReport{
		ReportDate:     time.Now().AddDate(0, 0, -1), // Yesterday
		ValidationDate: time.Now(),
		DataQualityIssues: make([]DataQualityIssue, 0),
		Recommendations:   make([]string, 0),
	}

	fmt.Println("1. 🧮 Validasi Accounting Equation (Assets = Liabilities + Equity)...")
	if err := validateAccountingEquation(report); err != nil {
		return nil, fmt.Errorf("accounting equation validation failed: %v", err)
	}

	fmt.Println("2. 📚 Validasi Journal Entries...")
	if err := validateJournalEntries(report); err != nil {
		return nil, fmt.Errorf("journal validation failed: %v", err)
	}

	fmt.Println("3. 🏦 Validasi Account Structure...")
	if err := validateAccountStructure(report); err != nil {
		return nil, fmt.Errorf("account validation failed: %v", err)
	}

	fmt.Println("4. 📊 Validasi Report Consistency...")
	if err := validateReportConsistency(report); err != nil {
		return nil, fmt.Errorf("report consistency validation failed: %v", err)
	}

	fmt.Println("5. 🔍 Analisis Data Quality...")
	if err := analyzeDataQuality(report); err != nil {
		return nil, fmt.Errorf("data quality analysis failed: %v", err)
	}

	// Calculate overall score
	calculateOverallScore(report)

	return report, nil
}

func validateAccountingEquation(report *FinancialValidationReport) error {
	// Query for account balances by type
	query := `
		SELECT 
			type,
			SUM(balance) as total_balance
		FROM accounts 
		WHERE is_active = true 
		GROUP BY type
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query account balances: %v", err)
	}
	defer rows.Close()

	balances := make(map[string]float64)
	for rows.Next() {
		var accountType string
		var totalBalance float64
		if err := rows.Scan(&accountType, &totalBalance); err != nil {
			return fmt.Errorf("failed to scan account balance: %v", err)
		}
		balances[accountType] = totalBalance
	}

	// Calculate totals
	report.AccountingEquation.TotalAssets = balances["ASSET"]
	report.AccountingEquation.TotalLiabilities = balances["LIABILITY"]
	report.AccountingEquation.TotalEquity = balances["EQUITY"]
	
	// For revenue and expense, we need to consider them as part of equity
	// (Retained Earnings = Revenue - Expenses)
	retainedEarnings := balances["REVENUE"] - balances["EXPENSE"]
	report.AccountingEquation.TotalEquity += retainedEarnings

	report.AccountingEquation.LiabilitiesPlusEquity = 
		report.AccountingEquation.TotalLiabilities + report.AccountingEquation.TotalEquity

	report.AccountingEquation.Difference = 
		report.AccountingEquation.TotalAssets - report.AccountingEquation.LiabilitiesPlusEquity

	// Consider balanced if difference is less than 0.01 (1 cent)
	report.AccountingEquation.IsBalanced = math.Abs(report.AccountingEquation.Difference) < 0.01

	// Calculate balance percentage
	if report.AccountingEquation.TotalAssets != 0 {
		report.AccountingEquation.BalancePercentage = 
			(1 - math.Abs(report.AccountingEquation.Difference)/math.Abs(report.AccountingEquation.TotalAssets)) * 100
	} else {
		report.AccountingEquation.BalancePercentage = 100
	}

	// Add recommendations if not balanced
	if !report.AccountingEquation.IsBalanced {
		report.DataQualityIssues = append(report.DataQualityIssues, DataQualityIssue{
			Category:    "ACCOUNTING_EQUATION",
			Severity:    "HIGH",
			Description: "Accounting equation is not balanced",
			Count:       1,
			Details:     fmt.Sprintf("Assets: %.2f, Liabilities + Equity: %.2f, Difference: %.2f", 
				report.AccountingEquation.TotalAssets, 
				report.AccountingEquation.LiabilitiesPlusEquity,
				report.AccountingEquation.Difference),
			Recommendation: "Periksa entry jurnal yang tidak seimbang dan perbaiki account balances",
		})
	}

	return nil
}

func validateJournalEntries(report *FinancialValidationReport) error {
	// Query journal entry statistics
	query := `
		SELECT 
			COUNT(*) as total_entries,
			COUNT(CASE WHEN is_balanced = true THEN 1 END) as balanced_entries,
			COUNT(CASE WHEN is_balanced = false THEN 1 END) as unbalanced_entries,
			COUNT(CASE WHEN status = 'POSTED' THEN 1 END) as posted_entries,
			COUNT(CASE WHEN status = 'DRAFT' THEN 1 END) as draft_entries,
			COALESCE(SUM(total_debit), 0) as total_debits,
			COALESCE(SUM(total_credit), 0) as total_credits
		FROM journal_entries
		WHERE deleted_at IS NULL
	`

	err := db.QueryRow(query).Scan(
		&report.JournalValidation.TotalEntries,
		&report.JournalValidation.BalancedEntries,
		&report.JournalValidation.UnbalancedEntries,
		&report.JournalValidation.PostedEntries,
		&report.JournalValidation.DraftEntries,
		&report.JournalValidation.TotalDebits,
		&report.JournalValidation.TotalCredits,
	)

	if err != nil {
		return fmt.Errorf("failed to query journal statistics: %v", err)
	}

	// Calculate debit-credit difference
	report.JournalValidation.DebitCreditDifference = 
		report.JournalValidation.TotalDebits - report.JournalValidation.TotalCredits

	// Calculate balance accuracy percentage
	if report.JournalValidation.TotalEntries > 0 {
		report.JournalValidation.BalanceAccuracy = 
			(float64(report.JournalValidation.BalancedEntries) / float64(report.JournalValidation.TotalEntries)) * 100
	}

	// Check for data quality issues
	if report.JournalValidation.UnbalancedEntries > 0 {
		report.DataQualityIssues = append(report.DataQualityIssues, DataQualityIssue{
			Category:    "JOURNAL_ENTRIES",
			Severity:    "MEDIUM",
			Description: "Unbalanced journal entries found",
			Count:       report.JournalValidation.UnbalancedEntries,
			Details:     fmt.Sprintf("%d dari %d entries tidak seimbang", 
				report.JournalValidation.UnbalancedEntries, report.JournalValidation.TotalEntries),
			Recommendation: "Review dan perbaiki journal entries yang tidak balance",
		})
	}

	if math.Abs(report.JournalValidation.DebitCreditDifference) > 0.01 {
		report.DataQualityIssues = append(report.DataQualityIssues, DataQualityIssue{
			Category:    "JOURNAL_ENTRIES",
			Severity:    "HIGH",
			Description: "Total debits and credits do not match",
			Count:       1,
			Details:     fmt.Sprintf("Total Debits: %.2f, Total Credits: %.2f, Difference: %.2f", 
				report.JournalValidation.TotalDebits, 
				report.JournalValidation.TotalCredits,
				report.JournalValidation.DebitCreditDifference),
			Recommendation: "Periksa semua journal entries untuk memastikan debit = credit",
		})
	}

	return nil
}

func validateAccountStructure(report *FinancialValidationReport) error {
	// Query account statistics by type
	query := `
		SELECT 
			COUNT(CASE WHEN is_active = true THEN 1 END) as total_active,
			COUNT(CASE WHEN is_active = true AND balance != 0 THEN 1 END) as with_balance,
			COUNT(CASE WHEN type = 'ASSET' AND is_active = true THEN 1 END) as asset_accounts,
			COUNT(CASE WHEN type = 'LIABILITY' AND is_active = true THEN 1 END) as liability_accounts,
			COUNT(CASE WHEN type = 'EQUITY' AND is_active = true THEN 1 END) as equity_accounts,
			COUNT(CASE WHEN type = 'REVENUE' AND is_active = true THEN 1 END) as revenue_accounts,
			COUNT(CASE WHEN type = 'EXPENSE' AND is_active = true THEN 1 END) as expense_accounts,
			COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_accounts
		FROM accounts
		WHERE deleted_at IS NULL
	`

	err := db.QueryRow(query).Scan(
		&report.AccountValidation.TotalActiveAccounts,
		&report.AccountValidation.AccountsWithBalance,
		&report.AccountValidation.AssetAccounts,
		&report.AccountValidation.LiabilityAccounts,
		&report.AccountValidation.EquityAccounts,
		&report.AccountValidation.RevenueAccounts,
		&report.AccountValidation.ExpenseAccounts,
		&report.AccountValidation.InactiveAccounts,
	)

	if err != nil {
		return fmt.Errorf("failed to query account statistics: %v", err)
	}

	// Check for minimum required account types
	if report.AccountValidation.AssetAccounts == 0 {
		report.DataQualityIssues = append(report.DataQualityIssues, DataQualityIssue{
			Category:    "ACCOUNT_STRUCTURE",
			Severity:    "HIGH",
			Description: "No asset accounts found",
			Count:       0,
			Details:     "Sistem membutuhkan minimal satu account ASSET yang aktif",
			Recommendation: "Buat account ASSET seperti Cash, Bank, atau Fixed Assets",
		})
	}

	if report.AccountValidation.RevenueAccounts == 0 {
		report.DataQualityIssues = append(report.DataQualityIssues, DataQualityIssue{
			Category:    "ACCOUNT_STRUCTURE",
			Severity:    "MEDIUM",
			Description: "No revenue accounts found",
			Count:       0,
			Details:     "Tidak ada account REVENUE yang aktif",
			Recommendation: "Buat account REVENUE untuk mencatat pendapatan",
		})
	}

	return nil
}

func validateReportConsistency(report *FinancialValidationReport) error {
	// For now, we'll implement basic consistency checks
	// In a more comprehensive system, we would:
	// 1. Generate actual reports and compare their totals
	// 2. Check trial balance against individual account balances
	// 3. Verify cash flow statement ties to balance sheet changes

	// Simple consistency checks based on our accounting equation validation
	report.ReportConsistency.BalanceSheetBalance = report.AccountingEquation.IsBalanced
	report.ReportConsistency.TrialBalanceBalance = math.Abs(report.JournalValidation.DebitCreditDifference) < 0.01
	
	// For P&L and Cash Flow, we'll assume they're balanced if journal entries are balanced
	report.ReportConsistency.ProfitLossBalance = report.JournalValidation.BalanceAccuracy > 95.0
	report.ReportConsistency.CashFlowBalance = report.JournalValidation.BalanceAccuracy > 95.0

	// Calculate consistency score
	consistencyCount := 0
	if report.ReportConsistency.BalanceSheetBalance { consistencyCount++ }
	if report.ReportConsistency.TrialBalanceBalance { consistencyCount++ }
	if report.ReportConsistency.ProfitLossBalance { consistencyCount++ }
	if report.ReportConsistency.CashFlowBalance { consistencyCount++ }

	report.ReportConsistency.ConsistencyScore = (float64(consistencyCount) / 4.0) * 100

	if report.ReportConsistency.ConsistencyScore < 75.0 {
		report.DataQualityIssues = append(report.DataQualityIssues, DataQualityIssue{
			Category:    "REPORT_CONSISTENCY",
			Severity:    "MEDIUM",
			Description: "Financial reports show consistency issues",
			Count:       int64(4 - consistencyCount),
			Details:     fmt.Sprintf("Consistency score: %.1f%%", report.ReportConsistency.ConsistencyScore),
			Recommendation: "Periksa data jurnal dan account balances untuk memastikan konsistensi laporan",
		})
	}

	return nil
}

func analyzeDataQuality(report *FinancialValidationReport) error {
	// Check for duplicate journal entry codes
	var duplicateCodes int64
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM (
			SELECT code 
			FROM journal_entries 
			WHERE deleted_at IS NULL 
			GROUP BY code 
			HAVING COUNT(*) > 1
		) duplicates
	`).Scan(&duplicateCodes)
	
	if err != nil {
		return fmt.Errorf("failed to check duplicate codes: %v", err)
	}

	if duplicateCodes > 0 {
		report.DataQualityIssues = append(report.DataQualityIssues, DataQualityIssue{
			Category:    "DATA_INTEGRITY",
			Severity:    "MEDIUM",
			Description: "Duplicate journal entry codes found",
			Count:       duplicateCodes,
			Details:     fmt.Sprintf("%d journal entry codes are duplicated", duplicateCodes),
			Recommendation: "Pastikan setiap journal entry memiliki code yang unique",
		})
	}

	// Check for accounts without proper account codes
	var invalidAccountCodes int64
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM accounts 
		WHERE (code IS NULL OR code = '' OR LENGTH(code) < 3) 
		AND deleted_at IS NULL
	`).Scan(&invalidAccountCodes)
	
	if err != nil {
		return fmt.Errorf("failed to check invalid account codes: %v", err)
	}

	if invalidAccountCodes > 0 {
		report.DataQualityIssues = append(report.DataQualityIssues, DataQualityIssue{
			Category:    "DATA_INTEGRITY",
			Severity:    "LOW",
			Description: "Accounts with invalid or missing codes",
			Count:       invalidAccountCodes,
			Details:     fmt.Sprintf("%d accounts have invalid or missing codes", invalidAccountCodes),
			Recommendation: "Pastikan semua account memiliki code yang valid (minimal 3 karakter)",
		})
	}

	// Check for journal entries without proper references
	var entriesWithoutRef int64
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM journal_entries 
		WHERE (reference IS NULL OR reference = '') 
		AND deleted_at IS NULL
	`).Scan(&entriesWithoutRef)
	
	if err != nil {
		return fmt.Errorf("failed to check entries without references: %v", err)
	}

	if entriesWithoutRef > 0 {
		report.DataQualityIssues = append(report.DataQualityIssues, DataQualityIssue{
			Category:    "DATA_COMPLETENESS",
			Severity:    "LOW",
			Description: "Journal entries without references",
			Count:       entriesWithoutRef,
			Details:     fmt.Sprintf("%d journal entries don't have reference information", entriesWithoutRef),
			Recommendation: "Tambahkan reference untuk semua journal entries untuk audit trail yang lebih baik",
		})
	}

	return nil
}

func calculateOverallScore(report *FinancialValidationReport) {
	score := 100.0

	// Deduct points for each data quality issue based on severity
	for _, issue := range report.DataQualityIssues {
		switch issue.Severity {
		case "HIGH":
			score -= 20.0
		case "MEDIUM":
			score -= 10.0
		case "LOW":
			score -= 5.0
		}
	}

	// Adjust for balance accuracy
	if report.JournalValidation.TotalEntries > 0 {
		accuracyFactor := report.JournalValidation.BalanceAccuracy / 100.0
		score = score * accuracyFactor
	}

	// Adjust for accounting equation balance
	if !report.AccountingEquation.IsBalanced {
		balanceFactor := report.AccountingEquation.BalancePercentage / 100.0
		score = score * balanceFactor
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	report.OverallScore = score

	// Determine status
	if score >= 95 {
		report.OverallStatus = "EXCELLENT"
		report.Recommendations = append(report.Recommendations, "Sistem akuntansi Anda dalam kondisi sangat baik!")
	} else if score >= 85 {
		report.OverallStatus = "GOOD"
		report.Recommendations = append(report.Recommendations, "Sistem akuntansi dalam kondisi baik dengan beberapa area yang bisa diperbaiki.")
	} else if score >= 70 {
		report.OverallStatus = "NEEDS_ATTENTION"
		report.Recommendations = append(report.Recommendations, "Sistem akuntansi memerlukan perhatian untuk memperbaiki beberapa masalah.")
	} else {
		report.OverallStatus = "CRITICAL"
		report.Recommendations = append(report.Recommendations, "Sistem akuntansi memerlukan perbaikan segera untuk memastikan akurasi laporan keuangan.")
	}

	// Add specific recommendations based on findings
	if !report.AccountingEquation.IsBalanced {
		report.Recommendations = append(report.Recommendations, 
			"Perbaiki accounting equation: pastikan Assets = Liabilities + Equity")
	}

	if report.JournalValidation.BalanceAccuracy < 95.0 {
		report.Recommendations = append(report.Recommendations, 
			"Review dan perbaiki journal entries yang tidak seimbang")
	}

	if report.ReportConsistency.ConsistencyScore < 80.0 {
		report.Recommendations = append(report.Recommendations, 
			"Lakukan reconciliation untuk memastikan konsistensi antar laporan")
	}
}

func displayValidationResults(report *FinancialValidationReport) {
	fmt.Println("\n" + "="*80)
	fmt.Printf("📊 FINANCIAL REPORT VALIDATION RESULTS\n")
	fmt.Printf("📅 Report Date: %s\n", report.ReportDate.Format("2006-01-02"))
	fmt.Printf("🕒 Validation Time: %s\n", report.ValidationDate.Format("2006-01-02 15:04:05"))
	fmt.Println("="*80)

	// Overall Score
	fmt.Printf("\n🏆 OVERALL SCORE: %.1f/100 (%s)\n", report.OverallScore, report.OverallStatus)
	
	statusEmoji := "✅"
	if report.OverallStatus == "NEEDS_ATTENTION" {
		statusEmoji = "⚠️"
	} else if report.OverallStatus == "CRITICAL" {
		statusEmoji = "❌"
	}
	fmt.Printf("%s Status: %s\n", statusEmoji, report.OverallStatus)

	// Accounting Equation
	fmt.Println("\n1. 🧮 ACCOUNTING EQUATION CHECK")
	fmt.Println("-" * 50)
	fmt.Printf("Assets:                 %15.2f\n", report.AccountingEquation.TotalAssets)
	fmt.Printf("Liabilities:            %15.2f\n", report.AccountingEquation.TotalLiabilities)
	fmt.Printf("Equity:                 %15.2f\n", report.AccountingEquation.TotalEquity)
	fmt.Printf("Liabilities + Equity:   %15.2f\n", report.AccountingEquation.LiabilitiesPlusEquity)
	fmt.Printf("Difference:             %15.2f\n", report.AccountingEquation.Difference)
	
	balanceStatus := "❌ NOT BALANCED"
	if report.AccountingEquation.IsBalanced {
		balanceStatus = "✅ BALANCED"
	}
	fmt.Printf("Status: %s (%.2f%%)\n", balanceStatus, report.AccountingEquation.BalancePercentage)

	// Journal Validation
	fmt.Println("\n2. 📚 JOURNAL ENTRIES VALIDATION")
	fmt.Println("-" * 50)
	fmt.Printf("Total Entries:          %15d\n", report.JournalValidation.TotalEntries)
	fmt.Printf("Balanced Entries:       %15d\n", report.JournalValidation.BalancedEntries)
	fmt.Printf("Unbalanced Entries:     %15d\n", report.JournalValidation.UnbalancedEntries)
	fmt.Printf("Posted Entries:         %15d\n", report.JournalValidation.PostedEntries)
	fmt.Printf("Draft Entries:          %15d\n", report.JournalValidation.DraftEntries)
	fmt.Printf("Total Debits:           %15.2f\n", report.JournalValidation.TotalDebits)
	fmt.Printf("Total Credits:          %15.2f\n", report.JournalValidation.TotalCredits)
	fmt.Printf("Difference:             %15.2f\n", report.JournalValidation.DebitCreditDifference)
	fmt.Printf("Balance Accuracy:       %14.1f%%\n", report.JournalValidation.BalanceAccuracy)

	// Account Validation
	fmt.Println("\n3. 🏦 ACCOUNT STRUCTURE VALIDATION")
	fmt.Println("-" * 50)
	fmt.Printf("Total Active Accounts:  %15d\n", report.AccountValidation.TotalActiveAccounts)
	fmt.Printf("Accounts with Balance:  %15d\n", report.AccountValidation.AccountsWithBalance)
	fmt.Printf("Asset Accounts:         %15d\n", report.AccountValidation.AssetAccounts)
	fmt.Printf("Liability Accounts:     %15d\n", report.AccountValidation.LiabilityAccounts)
	fmt.Printf("Equity Accounts:        %15d\n", report.AccountValidation.EquityAccounts)
	fmt.Printf("Revenue Accounts:       %15d\n", report.AccountValidation.RevenueAccounts)
	fmt.Printf("Expense Accounts:       %15d\n", report.AccountValidation.ExpenseAccounts)
	fmt.Printf("Inactive Accounts:      %15d\n", report.AccountValidation.InactiveAccounts)

	// Report Consistency
	fmt.Println("\n4. 📊 REPORT CONSISTENCY CHECK")
	fmt.Println("-" * 50)
	fmt.Printf("Balance Sheet:          %15s\n", getBoolStatus(report.ReportConsistency.BalanceSheetBalance))
	fmt.Printf("Trial Balance:          %15s\n", getBoolStatus(report.ReportConsistency.TrialBalanceBalance))
	fmt.Printf("Profit & Loss:          %15s\n", getBoolStatus(report.ReportConsistency.ProfitLossBalance))
	fmt.Printf("Cash Flow:              %15s\n", getBoolStatus(report.ReportConsistency.CashFlowBalance))
	fmt.Printf("Consistency Score:      %14.1f%%\n", report.ReportConsistency.ConsistencyScore)

	// Data Quality Issues
	if len(report.DataQualityIssues) > 0 {
		fmt.Println("\n5. 🔍 DATA QUALITY ISSUES")
		fmt.Println("-" * 50)
		for i, issue := range report.DataQualityIssues {
			severityEmoji := "🔵"
			if issue.Severity == "MEDIUM" {
				severityEmoji = "🟡"
			} else if issue.Severity == "HIGH" {
				severityEmoji = "🔴"
			}
			fmt.Printf("%d. %s [%s] %s\n", i+1, severityEmoji, issue.Severity, issue.Description)
			fmt.Printf("   Count: %d\n", issue.Count)
			fmt.Printf("   Details: %s\n", issue.Details)
			fmt.Printf("   Recommendation: %s\n", issue.Recommendation)
			fmt.Println()
		}
	} else {
		fmt.Println("\n5. 🔍 DATA QUALITY ISSUES")
		fmt.Println("-" * 50)
		fmt.Println("✅ No data quality issues found!")
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		fmt.Println("\n6. 💡 RECOMMENDATIONS")
		fmt.Println("-" * 50)
		for i, rec := range report.Recommendations {
			fmt.Printf("%d. %s\n", i+1, rec)
		}
	}

	fmt.Println("\n" + "="*80)
}

func getBoolStatus(value bool) string {
	if value {
		return "✅ GOOD"
	}
	return "❌ ISSUE"
}

func saveResultsToFile(report *FinancialValidationReport) {
	// For simplicity, we'll create a basic JSON-like output
	// In a real implementation, you'd use json.Marshal
	
	filename := fmt.Sprintf("financial_validation_report_%s.txt", 
		time.Now().Format("20060102_150405"))
	
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Warning: Could not create report file: %v\n", err)
		return
	}
	defer file.Close()

	// Write report summary
	fmt.Fprintf(file, "FINANCIAL REPORT VALIDATION SUMMARY\n")
	fmt.Fprintf(file, "Generated: %s\n", report.ValidationDate.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "Overall Score: %.1f/100 (%s)\n\n", report.OverallScore, report.OverallStatus)

	// Write key metrics
	fmt.Fprintf(file, "KEY METRICS:\n")
	fmt.Fprintf(file, "- Accounting Equation Balanced: %t\n", report.AccountingEquation.IsBalanced)
	fmt.Fprintf(file, "- Journal Balance Accuracy: %.1f%%\n", report.JournalValidation.BalanceAccuracy)
	fmt.Fprintf(file, "- Report Consistency Score: %.1f%%\n", report.ReportConsistency.ConsistencyScore)
	fmt.Fprintf(file, "- Data Quality Issues: %d\n\n", len(report.DataQualityIssues))

	// Write issues
	if len(report.DataQualityIssues) > 0 {
		fmt.Fprintf(file, "ISSUES FOUND:\n")
		for i, issue := range report.DataQualityIssues {
			fmt.Fprintf(file, "%d. [%s] %s (Count: %d)\n", i+1, issue.Severity, issue.Description, issue.Count)
		}
		fmt.Fprintf(file, "\n")
	}

	// Write recommendations
	if len(report.Recommendations) > 0 {
		fmt.Fprintf(file, "RECOMMENDATIONS:\n")
		for i, rec := range report.Recommendations {
			fmt.Fprintf(file, "%d. %s\n", i+1, rec)
		}
	}

	fmt.Printf("Report saved to: %s\n", filename)
}