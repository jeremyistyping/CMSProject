package services

import (
	"context"
	"fmt"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/utils"
	"gorm.io/gorm"
)

// FiscalYearClosingService handles fiscal year-end closing operations
type FiscalYearClosingService struct {
	db     *gorm.DB
	logger *utils.JournalLogger
}

// NewFiscalYearClosingService creates a new fiscal year closing service
func NewFiscalYearClosingService(db *gorm.DB) *FiscalYearClosingService {
	return &FiscalYearClosingService{
		db:     db,
		logger: utils.NewJournalLogger(db),
	}
}

// FiscalYearClosingPreview contains preview data for closing
type FiscalYearClosingPreview struct {
	FiscalYearEnd      time.Time                `json:"fiscal_year_end"`
	TotalRevenue       float64                  `json:"total_revenue"`
	TotalExpense       float64                  `json:"total_expense"`
	NetIncome          float64                  `json:"net_income"`
	RetainedEarningsID uint                     `json:"retained_earnings_id"`
	RevenueAccounts    []FiscalAccountBalance   `json:"revenue_accounts"`
	ExpenseAccounts    []FiscalAccountBalance   `json:"expense_accounts"`
	ClosingEntries     []ClosingEntryPreview    `json:"closing_entries"`
	CanClose           bool                     `json:"can_close"`
	ValidationMessages []string                 `json:"validation_messages"`
}

// FiscalAccountBalance represents account balance summary for fiscal year closing
type FiscalAccountBalance struct {
	ID      uint    `json:"id"`
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

// ClosingEntryPreview represents a preview of closing journal entry
type ClosingEntryPreview struct {
	Description  string  `json:"description"`
	DebitAccount string  `json:"debit_account"`
	CreditAccount string `json:"credit_account"`
	Amount       float64 `json:"amount"`
}

// PreviewFiscalYearClosing generates preview of fiscal year-end closing
func (fycs *FiscalYearClosingService) PreviewFiscalYearClosing(ctx context.Context, fiscalYearEnd time.Time) (*FiscalYearClosingPreview, error) {
	preview := &FiscalYearClosingPreview{
		FiscalYearEnd:      fiscalYearEnd,
		ValidationMessages: []string{},
		CanClose:           true,
	}

	// Get retained earnings account by code (3201 = LABA DITAHAN)
	// Using code instead of ID ensures consistency across databases
	var retainedEarningsAccount models.Account
	err := fycs.db.Where("code = ? AND type = ? AND is_active = true", "3201", models.AccountTypeEquity).First(&retainedEarningsAccount).Error
	if err != nil {
		preview.CanClose = false
		if err == gorm.ErrRecordNotFound {
			preview.ValidationMessages = append(preview.ValidationMessages, "Retained Earnings account (code 3201) not found in database. Please create it first.")
		} else {
			preview.ValidationMessages = append(preview.ValidationMessages, fmt.Sprintf("Error finding Retained Earnings account: %v", err))
		}
		return preview, nil
	}

	preview.RetainedEarningsID = retainedEarningsAccount.ID

	// Validate that retained earnings is an equity account (permanent account)
	if retainedEarningsAccount.Type != models.AccountTypeEquity {
		preview.CanClose = false
		preview.ValidationMessages = append(preview.ValidationMessages,
			fmt.Sprintf("Retained Earnings must be an Equity account, but found: %s", retainedEarningsAccount.Type))
		return preview, nil
	}

	// Calculate fiscal year start
	fiscalYearStart := fiscalYearEnd.AddDate(-1, 0, 1)

	// Get all TEMPORARY accounts (Revenue & Expense) with balances
	// These are the only accounts that should be reset to zero at year-end
	
	// Get all revenue accounts with NON-ZERO balances (TEMPORARY accounts)
	// IMPORTANT: Use ABS(balance) > 0.01 to avoid double-closing already closed accounts
	var revenueAccounts []models.Account
	if err := fycs.db.Where("type = ? AND ABS(balance) > 0.01 AND is_active = true AND is_header = false", models.AccountTypeRevenue).
		Find(&revenueAccounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get revenue accounts: %v", err)
	}

	// Get all expense accounts with NON-ZERO balances (TEMPORARY accounts)
	var expenseAccounts []models.Account
	if err := fycs.db.Where("type = ? AND ABS(balance) > 0.01 AND is_active = true AND is_header = false", models.AccountTypeExpense).
		Find(&expenseAccounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get expense accounts: %v", err)
	}

	// Calculate totals
	for _, acc := range revenueAccounts {
		preview.TotalRevenue += acc.Balance
		preview.RevenueAccounts = append(preview.RevenueAccounts, FiscalAccountBalance{
			ID:      acc.ID,
			Code:    acc.Code,
			Name:    acc.Name,
			Balance: acc.Balance,
		})
	}

	for _, acc := range expenseAccounts {
		preview.TotalExpense += acc.Balance
		preview.ExpenseAccounts = append(preview.ExpenseAccounts, FiscalAccountBalance{
			ID:      acc.ID,
			Code:    acc.Code,
			Name:    acc.Name,
			Balance: acc.Balance,
		})
	}

	preview.NetIncome = preview.TotalRevenue - preview.TotalExpense

	// Generate closing entry previews
	if len(revenueAccounts) > 0 {
		preview.ClosingEntries = append(preview.ClosingEntries, ClosingEntryPreview{
			Description:   "Close Revenue Accounts to Retained Earnings",
			DebitAccount:  "Revenue Accounts (Total)",
			CreditAccount: fmt.Sprintf("%s - %s", retainedEarningsAccount.Code, retainedEarningsAccount.Name),
			Amount:        preview.TotalRevenue,
		})
	}

	if len(expenseAccounts) > 0 {
		preview.ClosingEntries = append(preview.ClosingEntries, ClosingEntryPreview{
			Description:   "Close Expense Accounts to Retained Earnings",
			DebitAccount:  fmt.Sprintf("%s - %s", retainedEarningsAccount.Code, retainedEarningsAccount.Name),
			CreditAccount: "Expense Accounts (Total)",
			Amount:        preview.TotalExpense,
		})
	}

	// Validation checks
	if preview.TotalRevenue == 0 && preview.TotalExpense == 0 {
		preview.ValidationMessages = append(preview.ValidationMessages, "No revenue or expense to close")
	}

	// Check for unbalanced entries in the fiscal year
	var unbalancedCount int64
	fycs.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ? AND is_balanced = false", fiscalYearStart, fiscalYearEnd).
		Count(&unbalancedCount)

	if unbalancedCount > 0 {
		preview.CanClose = false
		preview.ValidationMessages = append(preview.ValidationMessages,
			fmt.Sprintf("Found %d unbalanced journal entries in fiscal year", unbalancedCount))
	}

	// Validate that no permanent accounts (Assets, Liabilities, Equity except distributions) will be affected
	// Only temporary accounts (Revenue & Expense) should be closed
	if len(revenueAccounts) > 0 || len(expenseAccounts) > 0 {
		preview.ValidationMessages = append(preview.ValidationMessages,
			fmt.Sprintf("Will close %d revenue accounts and %d expense accounts (temporary accounts)", len(revenueAccounts), len(expenseAccounts)))
		preview.ValidationMessages = append(preview.ValidationMessages,
			"Permanent accounts (Assets, Liabilities, Equity) will remain unchanged")
	}

	return preview, nil
}

// ExecuteFiscalYearClosing performs the actual fiscal year-end closing
func (fycs *FiscalYearClosingService) ExecuteFiscalYearClosing(ctx context.Context, fiscalYearEnd time.Time, notes string) error {
	// Extract user ID from context (supports both gin.Context and standard context.Context)
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("user context required: %v", err)
	}

	// Get preview first to validate
	preview, err := fycs.PreviewFiscalYearClosing(ctx, fiscalYearEnd)
	if err != nil {
		return fmt.Errorf("failed to generate preview: %v", err)
	}

	if !preview.CanClose {
		return fmt.Errorf("cannot close fiscal year: %v", preview.ValidationMessages)
	}

	// Execute closing in transaction
	return fycs.db.Transaction(func(tx *gorm.DB) error {
		// Get retained earnings account
		var retainedEarningsAccount models.Account
		if err := tx.First(&retainedEarningsAccount, preview.RetainedEarningsID).Error; err != nil {
			return fmt.Errorf("retained earnings account not found: %v", err)
		}

		// Generate closing journal entry code
		closingJournalCode := fmt.Sprintf("CLO-%d-12-31", fiscalYearEnd.Year())

		// Create main closing journal entry
		closingJournal := models.JournalEntry{
			Code:            closingJournalCode,
			Description:     fmt.Sprintf("Fiscal Year-End Closing %d - %s", fiscalYearEnd.Year(), notes),
			Reference:       fmt.Sprintf("FY%d-CLOSING", fiscalYearEnd.Year()),
			ReferenceType:   models.JournalRefClosing,
			EntryDate:       fiscalYearEnd,
			UserID:          userID,
			Status:          models.JournalStatusPosted,
			IsAutoGenerated: true,
			IsBalanced:      true,
		}

		var journalLines []models.JournalLine
		lineNumber := 1

		// Close TEMPORARY ACCOUNTS - Revenue Accounts
		// Revenue accounts are temporary and must be reset to zero each fiscal year
		// IMPORTANT: Only close accounts with actual balances (> 0.01) to prevent double posting
		for _, revAccount := range preview.RevenueAccounts {
			if revAccount.Balance > 0.01 {
				// Debit Revenue Account (to zero it out - normal balance is credit)
				journalLines = append(journalLines, models.JournalLine{
					AccountID:    revAccount.ID,
					Description:  fmt.Sprintf("Close temporary account: %s to retained earnings", revAccount.Name),
					DebitAmount:  revAccount.Balance,
					CreditAmount: 0,
					LineNumber:   lineNumber,
				})
				lineNumber++
			}
		}

		// Credit Retained Earnings (PERMANENT ACCOUNT) with total revenue
		// Retained Earnings is a permanent equity account that carries forward
		if preview.TotalRevenue > 0 {
			journalLines = append(journalLines, models.JournalLine{
				AccountID:    retainedEarningsAccount.ID,
				Description:  "Transfer total revenue to retained earnings (permanent account)",
				DebitAmount:  0,
				CreditAmount: preview.TotalRevenue,
				LineNumber:   lineNumber,
			})
			lineNumber++
		}

		// Close TEMPORARY ACCOUNTS - Expense Accounts
		// Expense accounts are temporary and must be reset to zero each fiscal year
		if preview.TotalExpense > 0 {
			// Debit Retained Earnings (PERMANENT ACCOUNT) to reduce by expenses
			journalLines = append(journalLines, models.JournalLine{
				AccountID:    retainedEarningsAccount.ID,
				Description:  "Transfer total expenses from retained earnings (permanent account)",
				DebitAmount:  preview.TotalExpense,
				CreditAmount: 0,
				LineNumber:   lineNumber,
			})
			lineNumber++
		}

	// IMPORTANT: Only close accounts with actual balances (> 0.01) to prevent double posting
	for _, expAccount := range preview.ExpenseAccounts {
		if expAccount.Balance > 0.01 {
			// Credit Expense Account (to zero it out - normal balance is debit)
			journalLines = append(journalLines, models.JournalLine{
				AccountID:    expAccount.ID,
				Description:  fmt.Sprintf("Close temporary account: %s to retained earnings", expAccount.Name),
				DebitAmount:  0,
				CreditAmount: expAccount.Balance,
				LineNumber:   lineNumber,
			})
			lineNumber++
		}
	}

		// Calculate totals for journal entry
		closingJournal.TotalDebit = preview.TotalRevenue + preview.TotalExpense
		closingJournal.TotalCredit = preview.TotalRevenue + preview.TotalExpense

		// Create journal entry
		if err := tx.Create(&closingJournal).Error; err != nil {
			return fmt.Errorf("failed to create closing journal entry: %v", err)
		}

		// Create journal lines and associate with journal entry
		for i := range journalLines {
			journalLines[i].JournalEntryID = closingJournal.ID
			if err := tx.Create(&journalLines[i]).Error; err != nil {
				return fmt.Errorf("failed to create journal line: %v", err)
			}
		}

		// Attach lines to journal entry for balance update
		closingJournal.JournalLines = journalLines

		// Update account balances based on journal lines
		// This will properly update Revenue/Expense accounts to 0 and update Retained Earnings
		if err := fycs.updateAccountBalancesInTx(tx, &closingJournal); err != nil {
			return fmt.Errorf("failed to update account balances: %v", err)
		}

		netIncome := preview.TotalRevenue - preview.TotalExpense

		// Log the closing activity
		fycs.logger.LogProcessingInfo(ctx, "Fiscal Year-End Closing Completed", map[string]interface{}{
			"fiscal_year_end":     fiscalYearEnd,
			"net_income":          netIncome,
			"total_revenue":       preview.TotalRevenue,
			"total_expense":       preview.TotalExpense,
			"journal_entry_code":  closingJournalCode,
			"journal_entry_id":    closingJournal.ID,
			"retained_earnings_id": retainedEarningsAccount.ID,
			"user_id":             userID,
		})

		return nil
	})
}

// GetFiscalYearClosingHistory returns history of fiscal year closings
func (fycs *FiscalYearClosingService) GetFiscalYearClosingHistory(ctx context.Context) ([]map[string]interface{}, error) {
	var closingEntries []models.JournalEntry

	err := fycs.db.Where("reference_type = ?", models.JournalRefClosing).
		Order("entry_date DESC").
		Limit(10).
		Find(&closingEntries).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get closing history: %v", err)
	}

	var history []map[string]interface{}
	for _, entry := range closingEntries {
		history = append(history, map[string]interface{}{
			"id":          entry.ID,
			"code":        entry.Code,
			"description": entry.Description,
			"entry_date":  entry.EntryDate,
			"created_at":  entry.CreatedAt,
			"total_debit": entry.TotalDebit,
		})
	}

	return history, nil
}

// updateAccountBalancesInTx updates account balances based on journal lines
// This mirrors the logic in journal_entry_repository.go
func (fycs *FiscalYearClosingService) updateAccountBalancesInTx(tx *gorm.DB, entry *models.JournalEntry) error {
	// Process journal lines
	for _, line := range entry.JournalLines {
		if line.AccountID == 0 {
			return fmt.Errorf("invalid account ID in journal line")
		}

		// Get account details to determine normal balance type
		var account models.Account
		if err := tx.Select("id, code, name, type, balance").First(&account, line.AccountID).Error; err != nil {
			return fmt.Errorf("failed to get account %d details: %v", line.AccountID, err)
		}

		// Calculate balance change based on account type (normal balance)
		var balanceChange float64
		if account.Type == models.AccountTypeAsset || account.Type == models.AccountTypeExpense {
			// Debit normal accounts: debit increases, credit decreases
			balanceChange = line.DebitAmount - line.CreditAmount
		} else {
			// Credit normal accounts (LIABILITY, EQUITY, REVENUE): credit increases, debit decreases
			balanceChange = line.CreditAmount - line.DebitAmount
		}

		// Update account balance atomically
		result := tx.Model(&models.Account{}).
			Where("id = ? AND deleted_at IS NULL", line.AccountID).
			Update("balance", gorm.Expr("balance + ?", balanceChange))

		if result.Error != nil {
			return fmt.Errorf("failed to update balance for account %d: %v", line.AccountID, result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("account %d not found or inactive", line.AccountID)
		}
	}

	return nil
}
