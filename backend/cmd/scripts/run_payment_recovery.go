package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// Quick diagnostic script to analyze and potentially fix payment issues
type PaymentDiagnostics struct {
	db *gorm.DB
}

type DiagnosticReport struct {
	Timestamp             time.Time                    `json:"timestamp"`
	TotalPayments         int                          `json:"total_payments"`
	PendingPayments       int                          `json:"pending_payments"`
	CompletedPayments     int                          `json:"completed_payments"`
	ProblematicPayments   []ProblematicPayment         `json:"problematic_payments"`
	CashBankIssues        []CashBankIssue              `json:"cash_bank_issues"`
	OrphanedData          OrphanedDataSummary          `json:"orphaned_data"`
	IntegrityScore        float64                      `json:"integrity_score"`
	RecommendedActions    []string                     `json:"recommended_actions"`
}

type ProblematicPayment struct {
	PaymentID            uint     `json:"payment_id"`
	Code                 string   `json:"code"`
	Amount               float64  `json:"amount"`
	Method               string   `json:"method"`
	Status               string   `json:"status"`
	Date                 string   `json:"date"`
	Issues               []string `json:"issues"`
	HasJournalEntries    bool     `json:"has_journal_entries"`
	HasCashBankTx        bool     `json:"has_cash_bank_tx"`
	HasAllocations       bool     `json:"has_allocations"`
	AllocatedAmount      float64  `json:"allocated_amount"`
	AllocationMismatch   bool     `json:"allocation_mismatch"`
}

type CashBankIssue struct {
	AccountID       uint    `json:"account_id"`
	AccountName     string  `json:"account_name"`
	Balance         float64 `json:"balance"`
	IssueType       string  `json:"issue_type"`
	Description     string  `json:"description"`
}

type OrphanedDataSummary struct {
	OrphanedJournalDetails     int `json:"orphaned_journal_details"`
	OrphanedInvoiceAllocations int `json:"orphaned_invoice_allocations"`
	OrphanedBillAllocations    int `json:"orphaned_bill_allocations"`
	OrphanedCashBankTx         int `json:"orphaned_cash_bank_transactions"`
}

func NewPaymentDiagnostics() *PaymentDiagnostics {
	db := database.ConnectDB()
	return &PaymentDiagnostics{db: db}
}

func (pd *PaymentDiagnostics) GenerateFullReport() (*DiagnosticReport, error) {
	log.Println("🔍 Generating comprehensive payment diagnostics report...")
	
	report := &DiagnosticReport{
		Timestamp:           time.Now(),
		ProblematicPayments: []ProblematicPayment{},
		CashBankIssues:      []CashBankIssue{},
		RecommendedActions:  []string{},
	}

	// Step 1: Analyze payment counts
	var totalPayments, pendingPayments, completedPayments int64
	pd.db.Model(&models.Payment{}).Count(&totalPayments)
	pd.db.Model(&models.Payment{}).Where("status = ?", models.PaymentStatusPending).Count(&pendingPayments)
	pd.db.Model(&models.Payment{}).Where("status = ?", models.PaymentStatusCompleted).Count(&completedPayments)
	report.TotalPayments = int(totalPayments)
	report.PendingPayments = int(pendingPayments)
	report.CompletedPayments = int(completedPayments)

	// Step 2: Find problematic payments
	problematicPayments, err := pd.findProblematicPayments()
	if err != nil {
		return nil, fmt.Errorf("failed to find problematic payments: %v", err)
	}
	report.ProblematicPayments = problematicPayments

	// Step 3: Check cash/bank account issues
	cashBankIssues, err := pd.findCashBankIssues()
	if err != nil {
		return nil, fmt.Errorf("failed to find cash bank issues: %v", err)
	}
	report.CashBankIssues = cashBankIssues

	// Step 4: Check for orphaned data
	orphanedData, err := pd.findOrphanedData()
	if err != nil {
		return nil, fmt.Errorf("failed to find orphaned data: %v", err)
	}
	report.OrphanedData = orphanedData

	// Step 5: Calculate integrity score
	report.IntegrityScore = pd.calculateIntegrityScore(report)

	// Step 6: Generate recommendations
	report.RecommendedActions = pd.generateRecommendations(report)

	log.Println("✅ Diagnostic report generated successfully")
	return report, nil
}

func (pd *PaymentDiagnostics) findProblematicPayments() ([]ProblematicPayment, error) {
	var problematicPayments []ProblematicPayment

	// Query untuk mencari semua payment dengan detail analisis
	query := `
		SELECT 
			p.id,
			p.code,
			p.amount,
			p.method,
			p.status,
			p.date,
			CASE WHEN je.id IS NOT NULL THEN 1 ELSE 0 END as has_journal_entries,
			CASE WHEN cbt.id IS NOT NULL THEN 1 ELSE 0 END as has_cash_bank_tx,
			CASE WHEN pa.total_allocated IS NOT NULL THEN 1 ELSE 0 END as has_allocations,
			COALESCE(pa.total_allocated, 0) as allocated_amount
		FROM payments p
		LEFT JOIN journal_entries je ON je.reference_type = 'PAYMENT' AND je.reference_id = p.id
		LEFT JOIN cash_bank_transactions cbt ON cbt.reference_type = 'PAYMENT' AND cbt.reference_id = p.id
		LEFT JOIN (
			SELECT payment_id, SUM(allocated_amount) as total_allocated
			FROM payment_allocations 
			GROUP BY payment_id
		) pa ON pa.payment_id = p.id
		WHERE p.status = 'PENDING' 
		   OR je.id IS NULL 
		   OR cbt.id IS NULL
		   OR ABS(p.amount - COALESCE(pa.total_allocated, 0)) > 0.01
		ORDER BY p.date ASC
	`

	var results []struct {
		ID                uint    `gorm:"column:id"`
		Code              string  `gorm:"column:code"`
		Amount            float64 `gorm:"column:amount"`
		Method            string  `gorm:"column:method"`
		Status            string  `gorm:"column:status"`
		Date              time.Time `gorm:"column:date"`
		HasJournalEntries bool    `gorm:"column:has_journal_entries"`
		HasCashBankTx     bool    `gorm:"column:has_cash_bank_tx"`
		HasAllocations    bool    `gorm:"column:has_allocations"`
		AllocatedAmount   float64 `gorm:"column:allocated_amount"`
	}

	if err := pd.db.Raw(query).Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, result := range results {
		var issues []string

		// Analyze issues
		if result.Status == "PENDING" {
			issues = append(issues, "Status is PENDING")
		}
		if !result.HasJournalEntries {
			issues = append(issues, "Missing journal entries")
		}
		if !result.HasCashBankTx {
			issues = append(issues, "Missing cash/bank transaction")
		}
		if !result.HasAllocations {
			issues = append(issues, "No allocations found")
		}

		allocationMismatch := abs(result.Amount - result.AllocatedAmount) > 0.01
		if allocationMismatch {
			issues = append(issues, fmt.Sprintf("Allocation mismatch: %.2f vs %.2f", 
				result.Amount, result.AllocatedAmount))
		}

		// Only add if there are actual issues
		if len(issues) > 0 {
			problematicPayments = append(problematicPayments, ProblematicPayment{
				PaymentID:          result.ID,
				Code:               result.Code,
				Amount:             result.Amount,
				Method:             result.Method,
				Status:             result.Status,
				Date:               result.Date.Format("2006-01-02"),
				Issues:             issues,
				HasJournalEntries:  result.HasJournalEntries,
				HasCashBankTx:      result.HasCashBankTx,
				HasAllocations:     result.HasAllocations,
				AllocatedAmount:    result.AllocatedAmount,
				AllocationMismatch: allocationMismatch,
			})
		}
	}

	return problematicPayments, nil
}

func (pd *PaymentDiagnostics) findCashBankIssues() ([]CashBankIssue, error) {
	var issues []CashBankIssue

	// Find negative balances
	var negativeBalanceAccounts []models.CashBank
	if err := pd.db.Where("balance < 0").Find(&negativeBalanceAccounts).Error; err != nil {
		return nil, err
	}

	for _, account := range negativeBalanceAccounts {
		issues = append(issues, CashBankIssue{
			AccountID:   account.ID,
			AccountName: account.Name,
			Balance:     account.Balance,
			IssueType:   "NEGATIVE_BALANCE",
			Description: fmt.Sprintf("Account has negative balance: %.2f", account.Balance),
		})
	}

	// Find inactive accounts with transactions
	var inactiveWithTx []models.CashBank
	if err := pd.db.Where("is_active = ?", false).Find(&inactiveWithTx).Error; err != nil {
		return nil, err
	}

	for _, account := range inactiveWithTx {
		var txCount int64
		pd.db.Model(&models.CashBankTransaction{}).Where("cash_bank_id = ?", account.ID).Count(&txCount)
		if txCount > 0 {
			issues = append(issues, CashBankIssue{
				AccountID:   account.ID,
				AccountName: account.Name,
				Balance:     account.Balance,
				IssueType:   "INACTIVE_WITH_TRANSACTIONS",
				Description: fmt.Sprintf("Inactive account has %d transactions", txCount),
			})
		}
	}

	return issues, nil
}

func (pd *PaymentDiagnostics) findOrphanedData() (OrphanedDataSummary, error) {
	var summary OrphanedDataSummary

	// Orphaned journal lines
	pd.db.Raw(`
		SELECT COUNT(*) 
		FROM journal_lines jl 
		WHERE journal_entry_id NOT IN (SELECT id FROM journal_entries)
	`).Scan(&summary.OrphanedJournalDetails)

	// Orphaned payment allocations
	pd.db.Raw(`
		SELECT COUNT(*) 
		FROM payment_allocations pa 
		LEFT JOIN payments p ON pa.payment_id = p.id 
		WHERE p.id IS NULL
	`).Scan(&summary.OrphanedInvoiceAllocations)

	// No separate bill allocations - using the same count
	summary.OrphanedBillAllocations = 0

	// Orphaned cash bank transactions
	pd.db.Raw(`
		SELECT COUNT(*) 
		FROM cash_bank_transactions cbt 
		WHERE cbt.reference_type = 'PAYMENT' 
		  AND NOT EXISTS (
		    SELECT 1 FROM payments p WHERE p.id = cbt.reference_id
		  )
	`).Scan(&summary.OrphanedCashBankTx)

	return summary, nil
}

func (pd *PaymentDiagnostics) calculateIntegrityScore(report *DiagnosticReport) float64 {
	if report.TotalPayments == 0 {
		return 100.0
	}

	// Base score
	score := 100.0

	// Deduct for problematic payments
	if len(report.ProblematicPayments) > 0 {
		deduction := float64(len(report.ProblematicPayments)) / float64(report.TotalPayments) * 50
		score -= deduction
	}

	// Deduct for cash bank issues
	if len(report.CashBankIssues) > 0 {
		score -= float64(len(report.CashBankIssues)) * 5
	}

	// Deduct for orphaned data
	totalOrphaned := report.OrphanedData.OrphanedJournalDetails + 
		report.OrphanedData.OrphanedInvoiceAllocations + 
		report.OrphanedData.OrphanedBillAllocations + 
		report.OrphanedData.OrphanedCashBankTx

	if totalOrphaned > 0 {
		score -= float64(totalOrphaned) * 2
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (pd *PaymentDiagnostics) generateRecommendations(report *DiagnosticReport) []string {
	var recommendations []string

	if len(report.ProblematicPayments) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Fix %d problematic payments using payment recovery script", 
				len(report.ProblematicPayments)))
	}

	if len(report.CashBankIssues) > 0 {
		negativeCount := 0
		for _, issue := range report.CashBankIssues {
			if issue.IssueType == "NEGATIVE_BALANCE" {
				negativeCount++
			}
		}
		if negativeCount > 0 {
			recommendations = append(recommendations, 
				fmt.Sprintf("Correct %d negative balance accounts", negativeCount))
		}
	}

	totalOrphaned := report.OrphanedData.OrphanedJournalDetails + 
		report.OrphanedData.OrphanedInvoiceAllocations + 
		report.OrphanedData.OrphanedBillAllocations + 
		report.OrphanedData.OrphanedCashBankTx

	if totalOrphaned > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Clean up %d orphaned data records", totalOrphaned))
	}

	if report.IntegrityScore < 90 {
		recommendations = append(recommendations, 
			"Implement enhanced payment service with better error handling")
		recommendations = append(recommendations, 
			"Set up monitoring and alerting for payment processing")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "System is healthy - no immediate actions required")
	}

	return recommendations
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Quick fix function for immediate issues
func (pd *PaymentDiagnostics) QuickFix() error {
	log.Println("🚀 Running quick fixes for immediate issues...")

	// Fix 1: Update pending payments with complete data to COMPLETED
	result := pd.db.Exec(`
		UPDATE payments p
		SET status = 'COMPLETED'
		WHERE p.status = 'PENDING'
		  AND EXISTS (
		    SELECT 1 FROM journal_entries je 
		    WHERE je.reference_type = 'PAYMENT' AND je.reference_id = p.id
		  )
		  AND EXISTS (
		    SELECT 1 FROM cash_bank_transactions cbt 
		    WHERE cbt.reference_type = 'PAYMENT' AND cbt.reference_id = p.id
		  )
		  AND EXISTS (SELECT 1 FROM payment_allocations pa WHERE pa.payment_id = p.id)
	`)

	if result.Error == nil && result.RowsAffected > 0 {
		log.Printf("✅ Updated %d payments from PENDING to COMPLETED", result.RowsAffected)
	}

	// Fix 2: Remove orphaned payment allocations
	result = pd.db.Exec(`
		DELETE FROM payment_allocations 
		WHERE payment_id NOT IN (SELECT id FROM payments)
	`)
	if result.Error == nil && result.RowsAffected > 0 {
		log.Printf("🗑️ Cleaned %d orphaned payment allocations", result.RowsAffected)
	}

	log.Println("✅ Quick fixes completed")
	return nil
}

func main() {
	fmt.Println("==========================================")
	fmt.Println("💰 PAYMENT SYSTEM DIAGNOSTICS")
	fmt.Println("==========================================")

	diagnostics := NewPaymentDiagnostics()

	// Generate diagnostic report
	report, err := diagnostics.GenerateFullReport()
	if err != nil {
		log.Fatalf("❌ Failed to generate diagnostic report: %v", err)
	}

	// Print summary
	printDiagnosticSummary(report)

	// Ask if user wants detailed report
	fmt.Print("\n📋 Save detailed report to file? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		saveReport(report)
	}

	// Ask if user wants to run quick fixes
	if report.IntegrityScore < 100 {
		fmt.Print("\n🔧 Run quick fixes for immediate issues? (y/N): ")
		fmt.Scanln(&response)

		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			if err := diagnostics.QuickFix(); err != nil {
				log.Printf("❌ Quick fix failed: %v", err)
			} else {
				log.Println("✅ Quick fixes completed successfully")
			}
		}
	}

	// Ask if user wants to run full recovery
	if len(report.ProblematicPayments) > 0 {
		fmt.Printf("\n🚨 Found %d problematic payments that need manual recovery.\n", len(report.ProblematicPayments))
		fmt.Println("   Run the payment_recovery_script.go for comprehensive fixes.")
		
		fmt.Print("\nℹ️  Show problematic payment details? (y/N): ")
		fmt.Scanln(&response)
		
		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			printProblematicPayments(report.ProblematicPayments)
		}
	}
}

func printDiagnosticSummary(report *DiagnosticReport) {
	fmt.Printf("\n📊 DIAGNOSTIC SUMMARY (as of %s)\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Total Payments: %d\n", report.TotalPayments)
	fmt.Printf("  Pending Payments: %d\n", report.PendingPayments)
	fmt.Printf("  Completed Payments: %d\n", report.CompletedPayments)
	fmt.Printf("  Problematic Payments: %d\n", len(report.ProblematicPayments))
	fmt.Printf("  Cash/Bank Issues: %d\n", len(report.CashBankIssues))
	
	totalOrphaned := report.OrphanedData.OrphanedJournalDetails + 
		report.OrphanedData.OrphanedInvoiceAllocations + 
		report.OrphanedData.OrphanedBillAllocations + 
		report.OrphanedData.OrphanedCashBankTx
	fmt.Printf("  Orphaned Records: %d\n", totalOrphaned)
	
	fmt.Printf("  🎯 INTEGRITY SCORE: %.1f%%\n", report.IntegrityScore)

	// Color-coded status
	if report.IntegrityScore >= 95 {
		fmt.Println("  Status: 🟢 EXCELLENT")
	} else if report.IntegrityScore >= 80 {
		fmt.Println("  Status: 🟡 GOOD (minor issues)")
	} else if report.IntegrityScore >= 60 {
		fmt.Println("  Status: 🟠 NEEDS ATTENTION")
	} else {
		fmt.Println("  Status: 🔴 CRITICAL (immediate action required)")
	}

	if len(report.RecommendedActions) > 0 {
		fmt.Println("\n📝 RECOMMENDED ACTIONS:")
		for i, action := range report.RecommendedActions {
			fmt.Printf("  %d. %s\n", i+1, action)
		}
	}

	if len(report.CashBankIssues) > 0 {
		fmt.Println("\n💰 CASH/BANK ISSUES:")
		for _, issue := range report.CashBankIssues {
			fmt.Printf("  - %s: %s\n", issue.AccountName, issue.Description)
		}
	}
}

func printProblematicPayments(payments []ProblematicPayment) {
	fmt.Println("\n🚨 PROBLEMATIC PAYMENTS DETAILS:")
	for i, payment := range payments {
		fmt.Printf("\n  %d. Payment ID: %d (%s)\n", i+1, payment.PaymentID, payment.Code)
		fmt.Printf("     Amount: %.2f | Method: %s | Status: %s | Date: %s\n", 
			payment.Amount, payment.Method, payment.Status, payment.Date)
		fmt.Printf("     Issues: %s\n", strings.Join(payment.Issues, ", "))
		if payment.AllocationMismatch {
			fmt.Printf("     Allocation: Expected %.2f, Got %.2f\n", payment.Amount, payment.AllocatedAmount)
		}
	}
}

func saveReport(report *DiagnosticReport) {
	filename := fmt.Sprintf("payment_diagnostics_%s.json", 
		time.Now().Format("20060102_150405"))
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("❌ Failed to marshal report: %v", err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Printf("❌ Failed to save report: %v", err)
		return
	}

	log.Printf("💾 Diagnostic report saved to: %s", filename)
}