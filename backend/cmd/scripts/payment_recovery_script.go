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

// Payment Recovery Script untuk memperbaiki payment PENDING
type PaymentRecoveryService struct {
	db *gorm.DB
}

type RecoveryResult struct {
	PaymentID       uint      `json:"payment_id"`
	Status          string    `json:"status"`
	Action          string    `json:"action"`
	Details         string    `json:"details"`
	Success         bool      `json:"success"`
	ProcessedAt     time.Time `json:"processed_at"`
	ErrorMessage    string    `json:"error_message,omitempty"`
}

type RecoverySummary struct {
	TotalProcessed      int               `json:"total_processed"`
	TotalFixed          int               `json:"total_fixed"`
	TotalErrors         int               `json:"total_errors"`
	FixedPayments       []RecoveryResult  `json:"fixed_payments"`
	ErrorPayments       []RecoveryResult  `json:"error_payments"`
	BalanceCorrections  []BalanceCorrection `json:"balance_corrections"`
	ProcessingTime      time.Duration     `json:"processing_time"`
	StartedAt           time.Time         `json:"started_at"`
	CompletedAt         time.Time         `json:"completed_at"`
}

type BalanceCorrection struct {
	CashBankID      uint    `json:"cash_bank_id"`
	AccountName     string  `json:"account_name"`
	OldBalance      float64 `json:"old_balance"`
	NewBalance      float64 `json:"new_balance"`
	Adjustment      float64 `json:"adjustment"`
	Reason          string  `json:"reason"`
}

func NewPaymentRecoveryService() *PaymentRecoveryService {
	db := database.ConnectDB()
	return &PaymentRecoveryService{db: db}
}

// Main recovery function
func (s *PaymentRecoveryService) RecoverAllPendingPayments(dryRun bool) (*RecoverySummary, error) {
	startTime := time.Now()
	summary := &RecoverySummary{
		StartedAt:          startTime,
		FixedPayments:      []RecoveryResult{},
		ErrorPayments:      []RecoveryResult{},
		BalanceCorrections: []BalanceCorrection{},
	}

	log.Printf("🚀 Starting payment recovery process (Dry Run: %v)", dryRun)

	// Step 1: Find all pending payments
	pendingPayments, err := s.findPendingPayments()
	if err != nil {
		return nil, fmt.Errorf("failed to find pending payments: %v", err)
	}

	log.Printf("📋 Found %d pending payments to process", len(pendingPayments))
	summary.TotalProcessed = len(pendingPayments)

	// Step 2: Process each pending payment
	for _, payment := range pendingPayments {
		result := s.recoverPendingPayment(payment, dryRun)
		if result.Success {
			summary.TotalFixed++
			summary.FixedPayments = append(summary.FixedPayments, result)
		} else {
			summary.TotalErrors++
			summary.ErrorPayments = append(summary.ErrorPayments, result)
		}
	}

	// Step 3: Check and fix negative balances
	if !dryRun {
		balanceCorrections, err := s.fixNegativeBalances()
		if err != nil {
			log.Printf("⚠️ Warning: Failed to fix negative balances: %v", err)
		} else {
			summary.BalanceCorrections = balanceCorrections
		}
	}

	// Step 4: Cleanup orphaned data
	if !dryRun {
		s.cleanupOrphanedData()
	}

	summary.CompletedAt = time.Now()
	summary.ProcessingTime = summary.CompletedAt.Sub(startTime)

	log.Printf("🎉 Recovery completed: %d fixed, %d errors, took %.2f seconds",
		summary.TotalFixed, summary.TotalErrors, summary.ProcessingTime.Seconds())

	return summary, nil
}

// Find all pending payments that need recovery
func (s *PaymentRecoveryService) findPendingPayments() ([]models.Payment, error) {
	var pendingPayments []models.Payment
	
	// Query untuk mencari payment PENDING yang bermasalah
	query := `
		SELECT p.*
		FROM payments p
		LEFT JOIN journal_entries je ON je.reference_type = 'PAYMENT' AND je.reference_id = p.id
		LEFT JOIN cash_bank_transactions cbt ON cbt.reference_type = 'PAYMENT' AND cbt.reference_id = p.id
		WHERE p.status = 'PENDING'
		  AND (je.id IS NULL OR cbt.id IS NULL)
		ORDER BY p.date ASC
	`

	if err := s.db.Raw(query).Scan(&pendingPayments).Error; err != nil {
		return nil, err
	}

	// Load related data
	for i := range pendingPayments {
		s.db.Preload("Contact").Preload("User").
			Preload("InvoiceAllocations.Sale").
			Preload("BillAllocations.Purchase").
			Find(&pendingPayments[i])
	}

	return pendingPayments, nil
}

// Recover individual pending payment
func (s *PaymentRecoveryService) recoverPendingPayment(payment models.Payment, dryRun bool) RecoveryResult {
	result := RecoveryResult{
		PaymentID:   payment.ID,
		Status:      "PROCESSING",
		ProcessedAt: time.Now(),
	}

	log.Printf("🔧 Processing Payment ID: %d, Code: %s, Amount: %.2f", 
		payment.ID, payment.Code, payment.Amount)

	// Analyze what's missing
	missingComponents := s.analyzeMissingComponents(payment)
	result.Details = fmt.Sprintf("Missing components: %s", strings.Join(missingComponents, ", "))

	if dryRun {
		result.Action = "DRY_RUN_ANALYSIS"
		result.Success = true
		result.Status = "ANALYZED"
		log.Printf("📊 DRY RUN - Would fix: %s", result.Details)
		return result
	}

	// Start database transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("PANIC: %v", r)
			result.Status = "FAILED"
		}
	}()

	var actions []string

	// Step 1: Create missing journal entries
	if s.hasMissingJournalEntries(payment) {
		if err := s.createMissingJournalEntries(tx, payment); err != nil {
			tx.Rollback()
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("Failed to create journal entries: %v", err)
			result.Status = "FAILED"
			return result
		}
		actions = append(actions, "Created journal entries")
		log.Printf("✅ Created journal entries for payment %d", payment.ID)
	}

	// Step 2: Create missing cash/bank transactions
	if s.hasMissingCashBankTransaction(payment) {
		if err := s.createMissingCashBankTransaction(tx, payment); err != nil {
			tx.Rollback()
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("Failed to create cash/bank transaction: %v", err)
			result.Status = "FAILED"
			return result
		}
		actions = append(actions, "Created cash/bank transaction")
		log.Printf("✅ Created cash/bank transaction for payment %d", payment.ID)
	}

	// Step 3: Update payment status to COMPLETED
	if err := tx.Model(&payment).Update("status", models.PaymentStatusCompleted).Error; err != nil {
		tx.Rollback()
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("Failed to update payment status: %v", err)
		result.Status = "FAILED"
		return result
	}
	actions = append(actions, "Updated status to COMPLETED")

	// Step 4: Update related documents (sales/purchases outstanding amounts)
	if err := s.updateRelatedDocumentsOutstanding(tx, payment); err != nil {
		log.Printf("⚠️ Warning: Failed to update related documents for payment %d: %v", payment.ID, err)
		// Don't fail the whole transaction for this
	} else {
		actions = append(actions, "Updated related documents")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("Failed to commit transaction: %v", err)
		result.Status = "FAILED"
		return result
	}

	result.Success = true
	result.Status = "RECOVERED"
	result.Action = strings.Join(actions, "; ")
	log.Printf("🎉 Successfully recovered payment %d: %s", payment.ID, result.Action)

	return result
}

// Analyze what components are missing for a payment
func (s *PaymentRecoveryService) analyzeMissingComponents(payment models.Payment) []string {
	var missing []string

	if s.hasMissingJournalEntries(payment) {
		missing = append(missing, "Journal Entries")
	}

	if s.hasMissingCashBankTransaction(payment) {
		missing = append(missing, "Cash/Bank Transaction")
	}

	if len(missing) == 0 {
		missing = append(missing, "None (Status issue only)")
	}

	return missing
}

// Check if payment has missing journal entries
func (s *PaymentRecoveryService) hasMissingJournalEntries(payment models.Payment) bool {
	var count int64
	s.db.Model(&models.JournalEntry{}).
		Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).
		Count(&count)
	return count == 0
}

// Check if payment has missing cash/bank transaction
func (s *PaymentRecoveryService) hasMissingCashBankTransaction(payment models.Payment) bool {
	var count int64
	s.db.Model(&models.CashBankTransaction{}).
		Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).
		Count(&count)
	return count == 0
}

// Create missing journal entries for payment
func (s *PaymentRecoveryService) createMissingJournalEntries(tx *gorm.DB, payment models.Payment) error {
	
	// Generate journal entry code
	code, err := s.generateJournalCode(tx, payment.Date)
	if err != nil {
		return fmt.Errorf("failed to generate journal code: %v", err)
	}

	// Create journal header
	journal := &models.JournalEntry{
		Code:          code,
		ReferenceType: "PAYMENT",
		ReferenceID:   &payment.ID,
		EntryDate:     payment.Date,
		Description:   fmt.Sprintf("Payment %s - %s", payment.Code, payment.Method),
		UserID:        payment.UserID,
		Status:        "POSTED",
		TotalDebit:    payment.Amount,
		TotalCredit:   payment.Amount,
		IsBalanced:    true,
		IsAutoGenerated: true,
	}

	if err := tx.Create(journal).Error; err != nil {
		return fmt.Errorf("failed to create journal entry: %v", err)
	}

	// Create journal lines based on payment type
	var lines []models.JournalLine
	if payment.Method == "RECEIVABLE" {
		lines = s.createReceivableJournalLines(payment, journal.ID)
	} else if payment.Method == "PAYABLE" {
		lines = s.createPayableJournalLines(payment, journal.ID)
	}

	// Save journal lines
	for _, line := range lines {
		if err := tx.Create(&line).Error; err != nil {
			return fmt.Errorf("failed to create journal line: %v", err)
		}
	}

	return nil
}

// Create journal lines for customer payment (receivable)
func (s *PaymentRecoveryService) createReceivableJournalLines(payment models.Payment, journalEntryID uint) []models.JournalLine {
	var lines []models.JournalLine

	// Debit: Cash/Bank
	cashAccountID := s.getCashBankAccountID(payment)
	lines = append(lines, models.JournalLine{
		JournalEntryID: journalEntryID,
		AccountID:      cashAccountID,
		Description:    fmt.Sprintf("Cash received from %s", payment.Contact.Name),
		DebitAmount:    payment.Amount,
		CreditAmount:   0,
		LineNumber:     1,
	})

	// Credit: Accounts Receivable
	arAccountID := s.getAccountIDByCode("1201") // Accounts Receivable
	lines = append(lines, models.JournalLine{
		JournalEntryID: journalEntryID,
		AccountID:      arAccountID,
		Description:    fmt.Sprintf("Payment received from %s", payment.Contact.Name),
		DebitAmount:    0,
		CreditAmount:   payment.Amount,
		LineNumber:     2,
	})

	return lines
}

// Create journal lines for vendor payment (payable)
func (s *PaymentRecoveryService) createPayableJournalLines(payment models.Payment, journalEntryID uint) []models.JournalLine {
	var lines []models.JournalLine

	// Debit: Accounts Payable
	apAccountID := s.getAccountIDByCode("2101") // Accounts Payable
	lines = append(lines, models.JournalLine{
		JournalEntryID: journalEntryID,
		AccountID:      apAccountID,
		Description:    fmt.Sprintf("Payment to %s", payment.Contact.Name),
		DebitAmount:    payment.Amount,
		CreditAmount:   0,
		LineNumber:     1,
	})

	// Credit: Cash/Bank
	cashAccountID := s.getCashBankAccountID(payment)
	lines = append(lines, models.JournalLine{
		JournalEntryID: journalEntryID,
		AccountID:      cashAccountID,
		Description:    fmt.Sprintf("Payment to %s", payment.Contact.Name),
		DebitAmount:    0,
		CreditAmount:   payment.Amount,
		LineNumber:     2,
	})

	return lines
}

// Create missing cash/bank transaction
func (s *PaymentRecoveryService) createMissingCashBankTransaction(tx *gorm.DB, payment models.Payment) error {
	// Find appropriate cash/bank account
	cashBankID, err := s.determineCashBankAccount(payment)
	if err != nil {
		return fmt.Errorf("failed to determine cash/bank account: %v", err)
	}

	// Get cash/bank account for balance update
	var cashBank models.CashBank
	if err := tx.First(&cashBank, cashBankID).Error; err != nil {
		return fmt.Errorf("cash/bank account not found: %v", err)
	}

	// Calculate amount (positive for incoming, negative for outgoing)
	amount := payment.Amount
	if payment.Method == "PAYABLE" {
		amount = -payment.Amount
	}

	// Update balance
	newBalance := cashBank.Balance + amount
	cashBank.Balance = newBalance

	if err := tx.Save(&cashBank).Error; err != nil {
		return fmt.Errorf("failed to update cash/bank balance: %v", err)
	}

	// Create transaction record
	transaction := &models.CashBankTransaction{
		CashBankID:      cashBankID,
		ReferenceType:   "PAYMENT",
		ReferenceID:     payment.ID,
		Amount:          amount,
		BalanceAfter:    newBalance,
		TransactionDate: payment.Date,
		Notes:           fmt.Sprintf("Payment %s - %s", payment.Code, payment.Method),
	}

	if err := tx.Create(transaction).Error; err != nil {
		return fmt.Errorf("failed to create cash/bank transaction: %v", err)
	}

	return nil
}

// Update outstanding amounts for related documents
func (s *PaymentRecoveryService) updateRelatedDocumentsOutstanding(tx *gorm.DB, payment models.Payment) error {
	// Load allocations
	var paymentAllocations []models.PaymentAllocation
	tx.Where("payment_id = ?", payment.ID).Find(&paymentAllocations)

	// Update invoice outstanding amounts
	for _, alloc := range paymentAllocations {
		if alloc.InvoiceID != nil {
			var sale models.Sale
			if err := tx.First(&sale, *alloc.InvoiceID).Error; err == nil {
				sale.OutstandingAmount -= alloc.AllocatedAmount
				if sale.OutstandingAmount <= 0.01 { // Consider floating point precision
					sale.OutstandingAmount = 0
					sale.Status = models.SaleStatusPaid
				}
				tx.Save(&sale)
			}
		}

		// Update bill outstanding amounts
		if alloc.BillID != nil {
			var purchase models.Purchase
			if err := tx.First(&purchase, *alloc.BillID).Error; err == nil {
				purchase.OutstandingAmount -= alloc.AllocatedAmount
				if purchase.OutstandingAmount <= 0.01 { // Consider floating point precision
					purchase.OutstandingAmount = 0
					purchase.Status = models.PurchaseStatusPaid
				}
				tx.Save(&purchase)
			}
		}
	}

	return nil
}

// Fix negative balances
func (s *PaymentRecoveryService) fixNegativeBalances() ([]BalanceCorrection, error) {
	var corrections []BalanceCorrection

	// Find accounts with negative balances
	var negativeAccounts []models.CashBank
	if err := s.db.Where("balance < 0").Find(&negativeAccounts).Error; err != nil {
		return nil, err
	}

	for _, account := range negativeAccounts {
		oldBalance := account.Balance
		
		// For this fix, we'll set negative balances to 0
		// In production, you might want more sophisticated logic
		account.Balance = 0
		
		correction := BalanceCorrection{
			CashBankID:  account.ID,
			AccountName: account.Name,
			OldBalance:  oldBalance,
			NewBalance:  0,
			Adjustment:  -oldBalance,
			Reason:      "Corrected negative balance to zero",
		}

		if err := s.db.Save(&account).Error; err != nil {
			log.Printf("⚠️ Failed to correct balance for account %d: %v", account.ID, err)
			continue
		}

		corrections = append(corrections, correction)
		log.Printf("💰 Corrected negative balance for %s: %.2f -> %.2f", 
			account.Name, oldBalance, 0.0)
	}

	return corrections, nil
}

// Cleanup orphaned data
func (s *PaymentRecoveryService) cleanupOrphanedData() {
	log.Printf("🧹 Starting orphaned data cleanup...")

	// Clean orphaned journal lines (not journal_details)
	result := s.db.Exec(`
		DELETE FROM journal_lines 
		WHERE journal_entry_id NOT IN (SELECT id FROM journal_entries)
	`)
	if result.Error == nil && result.RowsAffected > 0 {
		log.Printf("🗑️ Cleaned %d orphaned journal details", result.RowsAffected)
	}

	// Clean orphaned payment allocations
	result = s.db.Exec(`
		DELETE FROM payment_allocations 
		WHERE payment_id NOT IN (SELECT id FROM payments)
	`)
	if result.Error == nil && result.RowsAffected > 0 {
		log.Printf("🗑️ Cleaned %d orphaned payment allocations", result.RowsAffected)
	}
}

// Helper functions
func (s *PaymentRecoveryService) generateJournalCode(tx *gorm.DB, date time.Time) (string, error) {
	year := date.Year()
	month := int(date.Month())
	
	// Get next sequence number
	var maxSeq int
	tx.Model(&models.JournalEntry{}).
		Where("YEAR(date) = ? AND MONTH(date) = ?", year, month).
		Select("COALESCE(MAX(CAST(SUBSTRING_INDEX(code, '/', -1) AS UNSIGNED)), 0)").
		Scan(&maxSeq)
	
	seq := maxSeq + 1
	return fmt.Sprintf("JE-%04d/%02d/%04d", year, month, seq), nil
}

func (s *PaymentRecoveryService) getCashBankAccountID(payment models.Payment) uint {
	// Try to find the actual cash/bank account used
	var transaction models.CashBankTransaction
	if err := s.db.Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).
		First(&transaction).Error; err == nil {
		var cashBank models.CashBank
		if err := s.db.First(&cashBank, transaction.CashBankID).Error; err == nil {
			if cashBank.AccountID > 0 {
				return cashBank.AccountID
			}
		}
	}
	
	// Default: try to find cash account by code
	return s.getAccountIDByCode("1101") // Cash account
}

func (s *PaymentRecoveryService) getAccountIDByCode(code string) uint {
	var account models.Account
	if err := s.db.Where("code = ?", code).First(&account).Error; err == nil {
		return account.ID
	}
	// Return 1 as fallback (usually the first account)
	return 1
}

func (s *PaymentRecoveryService) determineCashBankAccount(payment models.Payment) (uint, error) {
	// Try to find existing transaction first
	var existingTx models.CashBankTransaction
	if err := s.db.Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).
		First(&existingTx).Error; err == nil {
		return existingTx.CashBankID, nil
	}

	// Find appropriate account based on payment type
	var cashBank models.CashBank
	accountType := "CASH"
	if payment.Method == "PAYABLE" {
		accountType = "BANK" // Prefer bank for vendor payments
	}

	if err := s.db.Where("type = ? AND is_active = ?", accountType, true).
		Order("balance DESC").First(&cashBank).Error; err != nil {
		// Fallback to any active account
		if err := s.db.Where("is_active = ?", true).
			Order("balance DESC").First(&cashBank).Error; err != nil {
			return 0, fmt.Errorf("no active cash/bank account found")
		}
	}

	return cashBank.ID, nil
}

// Main function untuk menjalankan script
func main() {
	service := NewPaymentRecoveryService()

	// Run in dry-run mode first
	log.Println("======================================")
	log.Println("🔍 PAYMENT RECOVERY - DRY RUN MODE")
	log.Println("======================================")

	dryRunSummary, err := service.RecoverAllPendingPayments(true)
	if err != nil {
		log.Fatalf("❌ Dry run failed: %v", err)
	}

	// Print dry run results
	printSummary(dryRunSummary, true)

	// Ask for confirmation to proceed with actual recovery
	fmt.Print("\n🤔 Do you want to proceed with actual recovery? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		log.Println("🛑 Recovery cancelled by user")
		return
	}

	// Run actual recovery
	log.Println("\n======================================")
	log.Println("🚀 PAYMENT RECOVERY - EXECUTION MODE")
	log.Println("======================================")

	actualSummary, err := service.RecoverAllPendingPayments(false)
	if err != nil {
		log.Fatalf("❌ Recovery execution failed: %v", err)
	}

	// Print actual results
	printSummary(actualSummary, false)

	// Save results to file
	saveResultsToFile(actualSummary)
}

func printSummary(summary *RecoverySummary, isDryRun bool) {
	mode := "EXECUTION"
	if isDryRun {
		mode = "DRY RUN"
	}

	fmt.Printf("\n📊 %s SUMMARY:\n", mode)
	fmt.Printf("  Total Processed: %d\n", summary.TotalProcessed)
	fmt.Printf("  Successfully Fixed: %d\n", summary.TotalFixed)
	fmt.Printf("  Errors: %d\n", summary.TotalErrors)
	fmt.Printf("  Processing Time: %.2f seconds\n", summary.ProcessingTime.Seconds())

	if len(summary.FixedPayments) > 0 {
		fmt.Println("\n✅ Successfully Fixed Payments:")
		for _, result := range summary.FixedPayments {
			fmt.Printf("  - Payment ID %d: %s\n", result.PaymentID, result.Action)
		}
	}

	if len(summary.ErrorPayments) > 0 {
		fmt.Println("\n❌ Failed Payments:")
		for _, result := range summary.ErrorPayments {
			fmt.Printf("  - Payment ID %d: %s\n", result.PaymentID, result.ErrorMessage)
		}
	}

	if len(summary.BalanceCorrections) > 0 {
		fmt.Println("\n💰 Balance Corrections:")
		for _, correction := range summary.BalanceCorrections {
			fmt.Printf("  - %s: %.2f -> %.2f (adj: %.2f)\n", 
				correction.AccountName, correction.OldBalance, 
				correction.NewBalance, correction.Adjustment)
		}
	}
}

func saveResultsToFile(summary *RecoverySummary) {
	filename := fmt.Sprintf("payment_recovery_results_%s.json", 
		time.Now().Format("20060102_150405"))
	
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		log.Printf("⚠️ Failed to marshal results: %v", err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Printf("⚠️ Failed to save results to file: %v", err)
		return
	}

	log.Printf("💾 Results saved to: %s", filename)
}