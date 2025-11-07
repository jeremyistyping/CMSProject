package services

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"

	"gorm.io/gorm"
)

// BalanceIntegrityValidator ensures balance consistency across all related tables
type BalanceIntegrityValidator struct {
	db *gorm.DB
}

// NewBalanceIntegrityValidator creates a new balance integrity validator
func NewBalanceIntegrityValidator(db *gorm.DB) *BalanceIntegrityValidator {
	return &BalanceIntegrityValidator{db: db}
}

// ValidationResult contains results of balance integrity validation
type ValidationResult struct {
	Valid               bool
	TotalChecks         int
	FailedChecks        int
	Inconsistencies     []BalanceInconsistency
	ValidationTime      time.Time
	ValidationDuration  time.Duration
}

// BalanceInconsistency represents a balance mismatch
type BalanceInconsistency struct {
	Type            string  // "CASHBANK_TRANSACTION", "CASHBANK_ACCOUNT", "JOURNAL_BALANCE"
	EntityID        uint
	EntityName      string
	ExpectedBalance float64
	ActualBalance   float64
	Difference      float64
	Severity        string  // "CRITICAL", "HIGH", "MEDIUM", "LOW"
	Description     string
}

// ValidateAllBalances performs comprehensive balance validation
func (v *BalanceIntegrityValidator) ValidateAllBalances() (*ValidationResult, error) {
	startTime := time.Now()
	log.Println("🔍 Starting comprehensive balance integrity validation...")

	result := &ValidationResult{
		Valid:          true,
		TotalChecks:    0,
		FailedChecks:   0,
		Inconsistencies: []BalanceInconsistency{},
		ValidationTime: startTime,
	}

	// Step 1: Validate CashBank vs Transaction consistency
	log.Println("📋 Step 1: Validating CashBank vs Transaction consistency")
	if err := v.validateCashBankTransactionConsistency(result); err != nil {
		return nil, fmt.Errorf("cashbank transaction validation failed: %v", err)
	}

	// Step 2: Validate CashBank vs GL Account consistency
	log.Println("📋 Step 2: Validating CashBank vs GL Account consistency")
	if err := v.validateCashBankAccountConsistency(result); err != nil {
		return nil, fmt.Errorf("cashbank account validation failed: %v", err)
	}

	// Step 3: Validate Journal Entry balance consistency
	log.Println("📋 Step 3: Validating Journal Entry balance consistency")
	if err := v.validateJournalBalanceConsistency(result); err != nil {
		return nil, fmt.Errorf("journal balance validation failed: %v", err)
	}

	// Step 4: Validate against double posting patterns
	log.Println("📋 Step 4: Checking for double posting patterns")
	if err := v.validateDoublePostingPatterns(result); err != nil {
		return nil, fmt.Errorf("double posting validation failed: %v", err)
	}

	result.ValidationDuration = time.Since(startTime)
	result.Valid = result.FailedChecks == 0

	// Log results
	if result.Valid {
		log.Printf("✅ All balance validations passed (%d checks, %.2fs)", 
			result.TotalChecks, result.ValidationDuration.Seconds())
	} else {
		log.Printf("❌ Balance validation failed (%d/%d checks failed, %.2fs)", 
			result.FailedChecks, result.TotalChecks, result.ValidationDuration.Seconds())
		
		// Log inconsistencies by severity
		for _, inconsistency := range result.Inconsistencies {
			log.Printf("  %s [%s]: %s (Expected: %.2f, Actual: %.2f, Diff: %.2f)",
				inconsistency.Severity, inconsistency.Type, inconsistency.Description,
				inconsistency.ExpectedBalance, inconsistency.ActualBalance, inconsistency.Difference)
		}
	}

	return result, nil
}

// validateCashBankTransactionConsistency checks if CashBank balance matches transaction sum
func (v *BalanceIntegrityValidator) validateCashBankTransactionConsistency(result *ValidationResult) error {
	var cashBanks []models.CashBank
	if err := v.db.Find(&cashBanks).Error; err != nil {
		return fmt.Errorf("failed to get cash banks: %v", err)
	}

	for _, cb := range cashBanks {
		result.TotalChecks++

		// Calculate expected balance from transactions
		var transactionSum float64
		err := v.db.Model(&models.CashBankTransaction{}).
			Where("cash_bank_id = ?", cb.ID).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&transactionSum).Error

		if err != nil {
			return fmt.Errorf("failed to calculate transaction sum for cash bank %d: %v", cb.ID, err)
		}

		// Check for inconsistency
		if cb.Balance != transactionSum {
			result.FailedChecks++
			result.Valid = false

			severity := "HIGH"
			if abs(cb.Balance-transactionSum) > 1000000 { // > 1M IDR
				severity = "CRITICAL"
			} else if abs(cb.Balance-transactionSum) < 1000 { // < 1K IDR
				severity = "LOW"
			}

			inconsistency := BalanceInconsistency{
				Type:            "CASHBANK_TRANSACTION",
				EntityID:        cb.ID,
				EntityName:      cb.Name,
				ExpectedBalance: transactionSum,
				ActualBalance:   cb.Balance,
				Difference:      cb.Balance - transactionSum,
				Severity:        severity,
				Description:     fmt.Sprintf("CashBank %s balance doesn't match transaction sum", cb.Name),
			}
			result.Inconsistencies = append(result.Inconsistencies, inconsistency)
		}
	}

	log.Printf("  📊 CashBank-Transaction: %d accounts checked", len(cashBanks))
	return nil
}

// validateCashBankAccountConsistency checks if CashBank balance matches GL account balance
func (v *BalanceIntegrityValidator) validateCashBankAccountConsistency(result *ValidationResult) error {
	query := `
		SELECT 
			cb.id, cb.name, cb.balance as cashbank_balance,
			a.id as account_id, a.code as account_code, a.name as account_name, a.balance as account_balance
		FROM cash_banks cb
		JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL AND a.deleted_at IS NULL
	`

	type ConsistencyCheck struct {
		ID              uint    `json:"id"`
		Name            string  `json:"name"`
		CashbankBalance float64 `json:"cashbank_balance"`
		AccountID       uint    `json:"account_id"`
		AccountCode     string  `json:"account_code"`
		AccountName     string  `json:"account_name"`
		AccountBalance  float64 `json:"account_balance"`
	}

	var checks []ConsistencyCheck
	if err := v.db.Raw(query).Scan(&checks).Error; err != nil {
		return fmt.Errorf("failed to get cashbank-account consistency data: %v", err)
	}

	for _, check := range checks {
		result.TotalChecks++

		// Check for inconsistency
		if check.CashbankBalance != check.AccountBalance {
			result.FailedChecks++
			result.Valid = false

			severity := "HIGH"
			diff := abs(check.CashbankBalance - check.AccountBalance)
			if diff > 1000000 { // > 1M IDR
				severity = "CRITICAL"
			} else if diff < 1000 { // < 1K IDR
				severity = "MEDIUM"
			}

			inconsistency := BalanceInconsistency{
				Type:            "CASHBANK_ACCOUNT",
				EntityID:        check.ID,
				EntityName:      check.Name,
				ExpectedBalance: check.CashbankBalance, // CashBank is source of truth
				ActualBalance:   check.AccountBalance,
				Difference:      check.AccountBalance - check.CashbankBalance,
				Severity:        severity,
				Description:     fmt.Sprintf("CashBank %s balance doesn't match GL Account %s", check.Name, check.AccountCode),
			}
			result.Inconsistencies = append(result.Inconsistencies, inconsistency)
		}
	}

	log.Printf("  📊 CashBank-Account: %d links checked", len(checks))
	return nil
}

// validateJournalBalanceConsistency checks journal entry debit/credit balance
func (v *BalanceIntegrityValidator) validateJournalBalanceConsistency(result *ValidationResult) error {
	// Check SSOT journal entries
	query := `
		SELECT 
			sje.id, sje.entry_number, sje.total_amount,
			COALESCE(SUM(sjl.debit_amount), 0) as total_debit,
			COALESCE(SUM(sjl.credit_amount), 0) as total_credit
		FROM ssot_journal_entries sje
		LEFT JOIN ssot_journal_lines sjl ON sje.id = sjl.journal_entry_id
		WHERE sje.deleted_at IS NULL AND sje.status = 'POSTED'
		GROUP BY sje.id, sje.entry_number, sje.total_amount
		HAVING ABS(total_debit - total_credit) > 0.01
	`

	type JournalCheck struct {
		ID          uint    `json:"id"`
		EntryNumber string  `json:"entry_number"`
		TotalAmount float64 `json:"total_amount"`
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
	}

	var unbalancedJournals []JournalCheck
	if err := v.db.Raw(query).Scan(&unbalancedJournals).Error; err != nil {
		return fmt.Errorf("failed to check journal balance consistency: %v", err)
	}

	for _, journal := range unbalancedJournals {
		result.TotalChecks++
		result.FailedChecks++
		result.Valid = false

		inconsistency := BalanceInconsistency{
			Type:            "JOURNAL_BALANCE",
			EntityID:        journal.ID,
			EntityName:      journal.EntryNumber,
			ExpectedBalance: journal.TotalDebit,
			ActualBalance:   journal.TotalCredit,
			Difference:      journal.TotalDebit - journal.TotalCredit,
			Severity:        "CRITICAL",
			Description:     fmt.Sprintf("Journal %s is not balanced (Debit: %.2f, Credit: %.2f)", journal.EntryNumber, journal.TotalDebit, journal.TotalCredit),
		}
		result.Inconsistencies = append(result.Inconsistencies, inconsistency)
	}

	log.Printf("  📊 Journal Balance: %d unbalanced entries found", len(unbalancedJournals))
	return nil
}

// validateDoublePostingPatterns checks for patterns that suggest double posting
func (v *BalanceIntegrityValidator) validateDoublePostingPatterns(result *ValidationResult) error {
	// Check for suspicious patterns where balance is exactly double the transaction sum
	query := `
		SELECT 
			cb.id, cb.name, cb.balance,
			COALESCE(SUM(cbt.amount), 0) as transaction_sum,
			COUNT(cbt.id) as transaction_count
		FROM cash_banks cb
		LEFT JOIN cash_bank_transactions cbt ON cb.id = cbt.cash_bank_id
		WHERE cb.deleted_at IS NULL
		GROUP BY cb.id, cb.name, cb.balance
		HAVING cb.balance > 0 AND ABS(cb.balance - COALESCE(SUM(cbt.amount), 0) * 2) < 0.01
	`

	type DoublePostingCheck struct {
		ID               uint    `json:"id"`
		Name             string  `json:"name"`
		Balance          float64 `json:"balance"`
		TransactionSum   float64 `json:"transaction_sum"`
		TransactionCount int     `json:"transaction_count"`
	}

	var suspiciousAccounts []DoublePostingCheck
	if err := v.db.Raw(query).Scan(&suspiciousAccounts).Error; err != nil {
		return fmt.Errorf("failed to check double posting patterns: %v", err)
	}

	for _, account := range suspiciousAccounts {
		result.TotalChecks++
		result.FailedChecks++
		result.Valid = false

		inconsistency := BalanceInconsistency{
			Type:            "DOUBLE_POSTING_PATTERN",
			EntityID:        account.ID,
			EntityName:      account.Name,
			ExpectedBalance: account.TransactionSum,
			ActualBalance:   account.Balance,
			Difference:      account.Balance - account.TransactionSum,
			Severity:        "CRITICAL",
			Description:     fmt.Sprintf("Account %s shows double posting pattern (Balance: %.2f is exactly 2x Transaction Sum: %.2f)", account.Name, account.Balance, account.TransactionSum),
		}
		result.Inconsistencies = append(result.Inconsistencies, inconsistency)
	}

	log.Printf("  🚨 Double Posting: %d suspicious patterns found", len(suspiciousAccounts))
	return nil
}

// FixInconsistencies attempts to fix detected inconsistencies
func (v *BalanceIntegrityValidator) FixInconsistencies(result *ValidationResult, dryRun bool) error {
	if result.Valid {
		log.Println("✅ No inconsistencies to fix")
		return nil
	}

	log.Printf("🔧 Attempting to fix %d inconsistencies (dry run: %v)", len(result.Inconsistencies), dryRun)

	if !dryRun {
		tx := v.db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				log.Printf("❌ Fix operation panicked: %v", r)
			}
		}()

		fixedCount := 0
		for _, inconsistency := range result.Inconsistencies {
			if err := v.fixSingleInconsistency(tx, inconsistency); err != nil {
				log.Printf("⚠️ Failed to fix inconsistency %s: %v", inconsistency.Description, err)
			} else {
				fixedCount++
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to commit fixes: %v", err)
		}

		log.Printf("✅ Fixed %d/%d inconsistencies", fixedCount, len(result.Inconsistencies))
	} else {
		log.Println("🔍 Dry run - no actual fixes applied")
	}

	return nil
}

// fixSingleInconsistency attempts to fix a single balance inconsistency
func (v *BalanceIntegrityValidator) fixSingleInconsistency(tx *gorm.DB, inconsistency BalanceInconsistency) error {
	switch inconsistency.Type {
	case "CASHBANK_ACCOUNT":
		// Sync GL account balance with CashBank balance (CashBank is source of truth)
		return tx.Exec("UPDATE accounts SET balance = ?, updated_at = NOW() WHERE id = (SELECT account_id FROM cash_banks WHERE id = ?)",
			inconsistency.ExpectedBalance, inconsistency.EntityID).Error

	case "DOUBLE_POSTING_PATTERN":
		// Fix doubled balance by dividing by 2
		correctBalance := inconsistency.ActualBalance / 2
		if err := tx.Exec("UPDATE cash_banks SET balance = ?, updated_at = NOW() WHERE id = ?",
			correctBalance, inconsistency.EntityID).Error; err != nil {
			return err
		}
		
		// Create corrective transaction
		return tx.Exec(`
			INSERT INTO cash_bank_transactions 
			(cash_bank_id, reference_type, reference_id, amount, balance_after, transaction_date, notes, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, NOW(), ?, NOW(), NOW())
		`, inconsistency.EntityID, "BALANCE_CORRECTION", 0, -correctBalance, correctBalance,
			"Balance correction - Double posting fix").Error

	case "CASHBANK_TRANSACTION":
		// This requires more complex analysis - log for manual review
		log.Printf("⚠️ Manual review required for CashBank-Transaction inconsistency: %s", inconsistency.Description)
		return nil

	default:
		return fmt.Errorf("unknown inconsistency type: %s", inconsistency.Type)
	}
}

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetBalanceHealthReport generates a comprehensive balance health report
func (v *BalanceIntegrityValidator) GetBalanceHealthReport() (map[string]interface{}, error) {
	validation, err := v.ValidateAllBalances()
	if err != nil {
		return nil, err
	}

	// Additional statistics
	var stats struct {
		TotalCashBanks      int     `json:"total_cash_banks"`
		TotalBalance        float64 `json:"total_balance"`
		TotalTransactions   int     `json:"total_transactions"`
		TotalJournalEntries int     `json:"total_journal_entries"`
	}

	var countCashBanks, countTransactions, countJournals int64
	v.db.Model(&models.CashBank{}).Count(&countCashBanks)
	v.db.Model(&models.CashBank{}).Select("COALESCE(SUM(balance), 0)").Scan(&stats.TotalBalance)
	v.db.Model(&models.CashBankTransaction{}).Count(&countTransactions)
	v.db.Table("ssot_journal_entries").Where("deleted_at IS NULL").Count(&countJournals)
	
	stats.TotalCashBanks = int(countCashBanks)
	stats.TotalTransactions = int(countTransactions)
	stats.TotalJournalEntries = int(countJournals)

	report := map[string]interface{}{
		"validation_result": validation,
		"system_statistics": stats,
		"health_score": func() float64 {
			if validation.TotalChecks == 0 {
				return 0
			}
			return float64(validation.TotalChecks-validation.FailedChecks) / float64(validation.TotalChecks) * 100
		}(),
		"recommendations": v.generateRecommendations(validation),
		"generated_at":    time.Now().Format(time.RFC3339),
	}

	return report, nil
}

// generateRecommendations generates recommendations based on validation results
func (v *BalanceIntegrityValidator) generateRecommendations(result *ValidationResult) []string {
	recommendations := []string{}

	if !result.Valid {
		criticalCount := 0
		highCount := 0
		
		for _, inconsistency := range result.Inconsistencies {
			if inconsistency.Severity == "CRITICAL" {
				criticalCount++
			} else if inconsistency.Severity == "HIGH" {
				highCount++
			}
		}

		if criticalCount > 0 {
			recommendations = append(recommendations, 
				fmt.Sprintf("URGENT: Fix %d critical balance inconsistencies immediately", criticalCount))
		}
		
		if highCount > 0 {
			recommendations = append(recommendations, 
				fmt.Sprintf("Fix %d high priority balance inconsistencies", highCount))
		}

		recommendations = append(recommendations, "Run balance integrity validation daily")
		recommendations = append(recommendations, "Implement automated balance reconciliation")
		recommendations = append(recommendations, "Use Single Source Posting Service for all balance updates")
	} else {
		recommendations = append(recommendations, "All balance validations passed - system is healthy")
		recommendations = append(recommendations, "Continue using Single Source Posting Service")
		recommendations = append(recommendations, "Schedule regular balance health checks")
	}

	return recommendations
}