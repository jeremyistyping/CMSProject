package services

import (
	"fmt"
	"time"

	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// ReportValidationService provides data validation and quality checks for reports
type ReportValidationService struct {
	db *gorm.DB
}

// ValidationIssue represents a data quality issue found during validation
type ValidationIssue struct {
	Level       string    `json:"level"`        // INFO, WARNING, ERROR, CRITICAL
	Category    string    `json:"category"`     // BALANCE, TRANSACTION, ACCOUNT, PERIOD
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`       // LOW, MEDIUM, HIGH
	Suggestion  string    `json:"suggestion"`
	AccountCode string    `json:"account_code,omitempty"`
	Amount      float64   `json:"amount,omitempty"`
	Count       int64     `json:"count,omitempty"`
	Date        time.Time `json:"date,omitempty"`
}

// ValidationReport contains the overall validation results
type ValidationReport struct {
	OverallHealth    string            `json:"overall_health"`    // EXCELLENT, GOOD, FAIR, POOR, CRITICAL
	HealthScore      float64           `json:"health_score"`      // 0-100
	TotalIssues      int               `json:"total_issues"`
	IssuesByLevel    map[string]int    `json:"issues_by_level"`
	IssuesByCategory map[string]int    `json:"issues_by_category"`
	Issues           []ValidationIssue `json:"issues"`
	Recommendations  []string          `json:"recommendations"`
	ValidationDate   time.Time         `json:"validation_date"`
}

// NewReportValidationService creates a new validation service
func NewReportValidationService(db *gorm.DB) *ReportValidationService {
	return &ReportValidationService{
		db: db,
	}
}

// ValidateReportData performs comprehensive validation checks
func (rvs *ReportValidationService) ValidateReportData(startDate, endDate time.Time) (*ValidationReport, error) {
	report := &ValidationReport{
		IssuesByLevel:    make(map[string]int),
		IssuesByCategory: make(map[string]int),
		ValidationDate:   time.Now(),
	}

	var issues []ValidationIssue

	// 1. Check for unbalanced journal entries
	unbalancedIssues, err := rvs.checkUnbalancedJournalEntries(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check unbalanced entries: %v", err)
	}
	issues = append(issues, unbalancedIssues...)

	// 2. Check for missing account mappings
	mappingIssues, err := rvs.checkMissingAccountMappings()
	if err != nil {
		return nil, fmt.Errorf("failed to check account mappings: %v", err)
	}
	issues = append(issues, mappingIssues...)

	// 3. Check for suspicious transactions
	suspiciousIssues, err := rvs.checkSuspiciousTransactions(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check suspicious transactions: %v", err)
	}
	issues = append(issues, suspiciousIssues...)

	// 4. Check account balance consistency
	balanceIssues, err := rvs.checkAccountBalanceConsistency(endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check balance consistency: %v", err)
	}
	issues = append(issues, balanceIssues...)

	// 5. Check for missing journal references
	referenceIssues, err := rvs.checkMissingReferences(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check references: %v", err)
	}
	issues = append(issues, referenceIssues...)

	// 6. Check for future-dated transactions
	futureDateIssues, err := rvs.checkFutureDatedTransactions(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check future dates: %v", err)
	}
	issues = append(issues, futureDateIssues...)

	// Compile results
	report.Issues = issues
	report.TotalIssues = len(issues)
	
	// Count issues by level and category
	for _, issue := range issues {
		report.IssuesByLevel[issue.Level]++
		report.IssuesByCategory[issue.Category]++
	}

	// Calculate health score and overall health
	report.HealthScore = rvs.calculateHealthScore(issues)
	report.OverallHealth = rvs.determineOverallHealth(report.HealthScore)
	report.Recommendations = rvs.generateRecommendations(issues)

	return report, nil
}

// checkUnbalancedJournalEntries checks for journal entries where debits != credits
func (rvs *ReportValidationService) checkUnbalancedJournalEntries(startDate, endDate time.Time) ([]ValidationIssue, error) {
	var issues []ValidationIssue
	
	var unbalancedEntries []models.JournalEntry
	err := rvs.db.Where("entry_date BETWEEN ? AND ? AND is_balanced = false AND status = 'POSTED'", 
		startDate, endDate).Find(&unbalancedEntries).Error
	if err != nil {
		return nil, err
	}

	if len(unbalancedEntries) > 0 {
		totalAmount := float64(0)
		for _, entry := range unbalancedEntries {
			totalAmount += entry.TotalDebit
		}

		issues = append(issues, ValidationIssue{
			Level:       "ERROR",
			Category:    "BALANCE",
			Title:       "Unbalanced Journal Entries Found",
			Description: fmt.Sprintf("Found %d journal entries where debits do not equal credits", len(unbalancedEntries)),
			Impact:      "HIGH",
			Suggestion:  "Review and correct these journal entries to ensure proper double-entry bookkeeping",
			Count:       int64(len(unbalancedEntries)),
			Amount:      totalAmount,
		})
	}

	return issues, nil
}

// checkMissingAccountMappings checks for accounts without proper type or category
func (rvs *ReportValidationService) checkMissingAccountMappings() ([]ValidationIssue, error) {
	var issues []ValidationIssue
	
	var problemAccounts []models.Account
	err := rvs.db.Where("(type = '' OR category = '') AND is_active = true").Find(&problemAccounts).Error
	if err != nil {
		return nil, err
	}

	if len(problemAccounts) > 0 {
		issues = append(issues, ValidationIssue{
			Level:       "WARNING",
			Category:    "ACCOUNT",
			Title:       "Accounts Missing Type or Category",
			Description: fmt.Sprintf("Found %d active accounts without proper type or category assignment", len(problemAccounts)),
			Impact:      "MEDIUM",
			Suggestion:  "Assign proper account types and categories to ensure accurate financial reporting",
			Count:       int64(len(problemAccounts)),
		})
	}

	return issues, nil
}

// checkSuspiciousTransactions checks for potentially problematic transactions
func (rvs *ReportValidationService) checkSuspiciousTransactions(startDate, endDate time.Time) ([]ValidationIssue, error) {
	var issues []ValidationIssue

	// Check for very large transactions (over 1 billion IDR)
	var largeTransactions []models.JournalEntry
	err := rvs.db.Where("entry_date BETWEEN ? AND ? AND total_debit > ? AND status = 'POSTED'", 
		startDate, endDate, 1000000000).Find(&largeTransactions).Error
	if err != nil {
		return nil, err
	}

	if len(largeTransactions) > 0 {
		issues = append(issues, ValidationIssue{
			Level:       "WARNING",
			Category:    "TRANSACTION",
			Title:       "Large Transaction Amounts Detected",
			Description: fmt.Sprintf("Found %d transactions over 1 billion IDR", len(largeTransactions)),
			Impact:      "MEDIUM",
			Suggestion:  "Review these large transactions to ensure accuracy",
			Count:       int64(len(largeTransactions)),
		})
	}

	// Check for round number transactions (potentially manual entries)
	var roundNumbers int64
	err = rvs.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ? AND status = 'POSTED' AND MOD(total_debit, 1000000) = 0 AND total_debit > 0", 
			startDate, endDate).Count(&roundNumbers).Error
	if err != nil {
		return nil, err
	}

	if roundNumbers > 10 { // Only warn if there are many round numbers
		issues = append(issues, ValidationIssue{
			Level:       "INFO",
			Category:    "TRANSACTION",
			Title:       "Many Round Number Transactions",
			Description: fmt.Sprintf("Found %d transactions with perfectly round amounts", roundNumbers),
			Impact:      "LOW",
			Suggestion:  "This may indicate manual entries; verify these are legitimate business transactions",
			Count:       roundNumbers,
		})
	}

	return issues, nil
}

// checkAccountBalanceConsistency verifies account balances are consistent
func (rvs *ReportValidationService) checkAccountBalanceConsistency(asOfDate time.Time) ([]ValidationIssue, error) {
	var issues []ValidationIssue

	// Check if total debits equal total credits across all accounts
	type BalanceSum struct {
		TotalDebits  float64
		TotalCredits float64
	}

	var balanceSum BalanceSum
	err := rvs.db.Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN jl.debit_amount > 0 THEN jl.debit_amount ELSE 0 END), 0) as total_debits,
			COALESCE(SUM(CASE WHEN jl.credit_amount > 0 THEN jl.credit_amount ELSE 0 END), 0) as total_credits
		FROM journal_lines jl
		JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE je.entry_date <= ? AND je.status = 'POSTED'
	`, asOfDate).Scan(&balanceSum).Error

	if err != nil {
		return nil, err
	}

	difference := balanceSum.TotalDebits - balanceSum.TotalCredits
	if difference > 0.01 || difference < -0.01 { // Allow for small rounding differences
		level := "WARNING"
		impact := "MEDIUM"
		if difference > 1000000 || difference < -1000000 { // Large differences are critical
			level = "CRITICAL"
			impact = "HIGH"
		}

		issues = append(issues, ValidationIssue{
			Level:       level,
			Category:    "BALANCE",
			Title:       "Total Debits and Credits Imbalance",
			Description: fmt.Sprintf("System-wide imbalance detected. Difference: %.2f", difference),
			Impact:      impact,
			Suggestion:  "Investigate and correct journal entries causing this imbalance",
			Amount:      difference,
		})
	}

	return issues, nil
}

// checkMissingReferences checks for journal entries without proper references
func (rvs *ReportValidationService) checkMissingReferences(startDate, endDate time.Time) ([]ValidationIssue, error) {
	var issues []ValidationIssue
	
	var missingRefs int64
	err := rvs.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ? AND (reference = '' OR reference IS NULL) AND status = 'POSTED'", 
			startDate, endDate).Count(&missingRefs).Error
	if err != nil {
		return nil, err
	}

	if missingRefs > 0 {
		issues = append(issues, ValidationIssue{
			Level:       "WARNING",
			Category:    "TRANSACTION",
			Title:       "Journal Entries Missing References",
			Description: fmt.Sprintf("Found %d journal entries without reference information", missingRefs),
			Impact:      "MEDIUM",
			Suggestion:  "Add reference information to improve audit trail",
			Count:       missingRefs,
		})
	}

	return issues, nil
}

// checkFutureDatedTransactions checks for transactions dated in the future
func (rvs *ReportValidationService) checkFutureDatedTransactions(startDate, endDate time.Time) ([]ValidationIssue, error) {
	var issues []ValidationIssue
	
	var futureEntries int64
	err := rvs.db.Model(&models.JournalEntry{}).
		Where("entry_date > ? AND status = 'POSTED'", time.Now().AddDate(0, 0, 1)).Count(&futureEntries).Error
	if err != nil {
		return nil, err
	}

	if futureEntries > 0 {
		issues = append(issues, ValidationIssue{
			Level:       "ERROR",
			Category:    "TRANSACTION",
			Title:       "Future-Dated Transactions Found",
			Description: fmt.Sprintf("Found %d transactions dated more than 1 day in the future", futureEntries),
			Impact:      "HIGH",
			Suggestion:  "Review and correct transaction dates to ensure they reflect actual business activity",
			Count:       futureEntries,
		})
	}

	return issues, nil
}

// calculateHealthScore calculates an overall health score based on issues
func (rvs *ReportValidationService) calculateHealthScore(issues []ValidationIssue) float64 {
	if len(issues) == 0 {
		return 100.0
	}

	totalPenalty := float64(0)
	
	for _, issue := range issues {
		penalty := float64(0)
		
		// Penalty based on level
		switch issue.Level {
		case "CRITICAL":
			penalty += 25.0
		case "ERROR":
			penalty += 15.0
		case "WARNING":
			penalty += 8.0
		case "INFO":
			penalty += 2.0
		}
		
		// Additional penalty based on impact
		switch issue.Impact {
		case "HIGH":
			penalty *= 1.5
		case "MEDIUM":
			penalty *= 1.2
		case "LOW":
			penalty *= 1.0
		}
		
		totalPenalty += penalty
	}

	score := 100.0 - totalPenalty
	if score < 0 {
		score = 0
	}
	
	return score
}

// determineOverallHealth determines overall health status based on score
func (rvs *ReportValidationService) determineOverallHealth(score float64) string {
	switch {
	case score >= 90:
		return "EXCELLENT"
	case score >= 75:
		return "GOOD"
	case score >= 60:
		return "FAIR"
	case score >= 40:
		return "POOR"
	default:
		return "CRITICAL"
	}
}

// generateRecommendations creates actionable recommendations based on issues
func (rvs *ReportValidationService) generateRecommendations(issues []ValidationIssue) []string {
	recommendations := []string{}
	
	hasUnbalanced := false
	hasLargeTransactions := false
	hasMissingRefs := false
	hasFutureDates := false
	
	for _, issue := range issues {
		switch issue.Category {
		case "BALANCE":
			if issue.Level == "ERROR" || issue.Level == "CRITICAL" {
				hasUnbalanced = true
			}
		case "TRANSACTION":
			if issue.Title == "Large Transaction Amounts Detected" {
				hasLargeTransactions = true
			}
			if issue.Title == "Journal Entries Missing References" {
				hasMissingRefs = true
			}
			if issue.Title == "Future-Dated Transactions Found" {
				hasFutureDates = true
			}
		}
	}
	
	if hasUnbalanced {
		recommendations = append(recommendations, "Priority 1: Fix unbalanced journal entries immediately to ensure data integrity")
	}
	if hasFutureDates {
		recommendations = append(recommendations, "Priority 2: Correct future-dated transactions before generating final reports")
	}
	if hasLargeTransactions {
		recommendations = append(recommendations, "Review large transaction amounts for accuracy")
	}
	if hasMissingRefs {
		recommendations = append(recommendations, "Improve documentation by adding references to journal entries")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Data quality is good. Continue monitoring for any changes.")
	}
	
	return recommendations
}