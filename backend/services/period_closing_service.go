package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/utils"
	"gorm.io/gorm"
)

// PeriodClosingService handles flexible period closing operations
type PeriodClosingService struct {
	db     *gorm.DB
	logger *utils.JournalLogger
}

// NewPeriodClosingService creates a new period closing service
func NewPeriodClosingService(db *gorm.DB) *PeriodClosingService {
	return &PeriodClosingService{
		db:     db,
		logger: utils.NewJournalLogger(db),
	}
}

// GetLastClosingInfo returns information about the last closed period
func (pcs *PeriodClosingService) GetLastClosingInfo(ctx context.Context) (*models.LastClosingInfo, error) {
	info := &models.LastClosingInfo{
		HasPreviousClosing: false,
	}

	// Get the last closed period
	var lastPeriod models.AccountingPeriod
	err := pcs.db.Where("is_closed = ?", true).
		Order("end_date DESC").
		First(&lastPeriod).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query last closing period: %v", err)
	}

	if err == nil {
		// Found previous closing
		info.HasPreviousClosing = true
		info.LastClosingDate = &lastPeriod.EndDate
		
		// Next period starts the day after last closing
		nextStart := lastPeriod.EndDate.AddDate(0, 0, 1)
		info.NextStartDate = &nextStart
	} else {
		// No previous closing - find earliest transaction date
		var earliestJournal models.JournalEntry
		err := pcs.db.Where("status = ?", models.JournalStatusPosted).
			Order("entry_date ASC").
			First(&earliestJournal).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to find earliest transaction: %v", err)
		}

		if err == nil {
			info.PeriodStartDate = &earliestJournal.EntryDate
		}
	}

	return info, nil
}

// PreviewPeriodClosing generates preview of period closing
func (pcs *PeriodClosingService) PreviewPeriodClosing(ctx context.Context, startDate, endDate time.Time) (*models.PeriodClosingPreview, error) {
	preview := &models.PeriodClosingPreview{
		StartDate:          startDate,
		EndDate:            endDate,
		ValidationMessages: []string{},
		CanClose:           true,
		PeriodDays:         int(endDate.Sub(startDate).Hours() / 24),
	}

	// Validate date range
	if endDate.Before(startDate) {
		preview.CanClose = false
		preview.ValidationMessages = append(preview.ValidationMessages, "End date must be after start date")
		return preview, nil
	}

	// Check if period is already closed
	var existingPeriod models.AccountingPeriod
	err := pcs.db.Where("start_date = ? AND end_date = ? AND is_closed = ?", startDate, endDate, true).
		First(&existingPeriod).Error
	
	if err == nil {
		preview.CanClose = false
		preview.ValidationMessages = append(preview.ValidationMessages, 
			fmt.Sprintf("Period from %s to %s is already closed", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")))
		return preview, nil
	}

	// Check for overlapping closed periods
	var overlappingCount int64
	pcs.db.Model(&models.AccountingPeriod{}).
		Where("is_closed = ? AND ((start_date BETWEEN ? AND ?) OR (end_date BETWEEN ? AND ?) OR (start_date <= ? AND end_date >= ?))",
			true, startDate, endDate, startDate, endDate, startDate, endDate).
		Count(&overlappingCount)
	
	if overlappingCount > 0 {
		preview.CanClose = false
		preview.ValidationMessages = append(preview.ValidationMessages, 
			fmt.Sprintf("Found %d overlapping closed periods. Cannot close overlapping periods.", overlappingCount))
	}

	// Validate period continuity - check if there's a gap from last closing
	lastInfo, err := pcs.GetLastClosingInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get last closing info: %v", err)
	}

	if lastInfo.HasPreviousClosing {
		expectedStart := lastInfo.NextStartDate
		if !startDate.Equal(*expectedStart) {
			preview.ValidationMessages = append(preview.ValidationMessages,
				fmt.Sprintf("⚠️ Warning: Period start (%s) does not match expected start (%s) after last closing",
					startDate.Format("2006-01-02"), expectedStart.Format("2006-01-02")))
		}
		preview.LastClosingDate = lastInfo.LastClosingDate
	}

	// Get retained earnings account
	var retainedEarningsAccount models.Account
	err = pcs.db.Where("code = ? AND type = ? AND is_active = true", "3201", models.AccountTypeEquity).
		First(&retainedEarningsAccount).Error
	
	if err != nil {
		preview.CanClose = false
		if err == gorm.ErrRecordNotFound {
			preview.ValidationMessages = append(preview.ValidationMessages, 
				"Retained Earnings account (code 3201) not found. Please create it first.")
		} else {
			preview.ValidationMessages = append(preview.ValidationMessages, 
				fmt.Sprintf("Error finding Retained Earnings account: %v", err))
		}
		return preview, nil
	}

	preview.RetainedEarningsID = retainedEarningsAccount.ID

	// Count transactions in period
	pcs.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ? AND status = ?", startDate, endDate, models.JournalStatusPosted).
		Count(&preview.TransactionCount)

	// IMPORTANT: Only get accounts with NON-ZERO balances
	// This prevents double-closing accounts that were already closed in previous periods
	// Period closing should ONLY close accounts with actual balances
	var revenueAccounts []models.Account
	if err := pcs.db.Where("type = ? AND ABS(balance) > 0.01 AND is_active = true AND is_header = false", models.AccountTypeRevenue).
		Find(&revenueAccounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get revenue accounts: %v", err)
	}

	// Get all expense accounts with NON-ZERO balances
	var expenseAccounts []models.Account
	if err := pcs.db.Where("type = ? AND ABS(balance) > 0.01 AND is_active = true AND is_header = false", models.AccountTypeExpense).
		Find(&expenseAccounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get expense accounts: %v", err)
	}

	// Calculate totals from current account balances
	for _, acc := range revenueAccounts {
		preview.TotalRevenue += acc.Balance
		preview.RevenueAccounts = append(preview.RevenueAccounts, models.PeriodAccountBalance{
			ID:      acc.ID,
			Code:    acc.Code,
			Name:    acc.Name,
			Balance: acc.Balance,
			Type:    string(acc.Type),
		})
	}

	for _, acc := range expenseAccounts {
		preview.TotalExpense += acc.Balance
		preview.ExpenseAccounts = append(preview.ExpenseAccounts, models.PeriodAccountBalance{
			ID:      acc.ID,
			Code:    acc.Code,
			Name:    acc.Name,
			Balance: acc.Balance,
			Type:    string(acc.Type),
		})
	}

	preview.NetIncome = preview.TotalRevenue - preview.TotalExpense

	// Generate closing entry previews
	if len(revenueAccounts) > 0 {
		preview.ClosingEntries = append(preview.ClosingEntries, models.ClosingEntryPreview{
			Description:   "Close Revenue Accounts to Retained Earnings",
			DebitAccount:  "Revenue Accounts (Total)",
			CreditAccount: fmt.Sprintf("%s - %s", retainedEarningsAccount.Code, retainedEarningsAccount.Name),
			Amount:        preview.TotalRevenue,
		})
	}

	if len(expenseAccounts) > 0 {
		preview.ClosingEntries = append(preview.ClosingEntries, models.ClosingEntryPreview{
			Description:   "Close Expense Accounts to Retained Earnings",
			DebitAccount:  fmt.Sprintf("%s - %s", retainedEarningsAccount.Code, retainedEarningsAccount.Name),
			CreditAccount: "Expense Accounts (Total)",
			Amount:        preview.TotalExpense,
		})
	}

	// Check for unbalanced entries in the period
	var unbalancedCount int64
	pcs.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ? AND is_balanced = false", startDate, endDate).
		Count(&unbalancedCount)

	if unbalancedCount > 0 {
		preview.CanClose = false
		preview.ValidationMessages = append(preview.ValidationMessages,
			fmt.Sprintf("❌ Found %d unbalanced journal entries in this period. Please fix them before closing.", unbalancedCount))
	}

	// Check for draft entries in the period
	var draftCount int64
	pcs.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ? AND status = ?", startDate, endDate, models.JournalStatusDraft).
		Count(&draftCount)

	if draftCount > 0 {
		preview.CanClose = false
		preview.ValidationMessages = append(preview.ValidationMessages,
			fmt.Sprintf("❌ Found %d draft journal entries in this period. Please post or delete them before closing.", draftCount))
	}

	// Summary messages
	if preview.TransactionCount == 0 {
		preview.ValidationMessages = append(preview.ValidationMessages, "⚠️ No transactions found in this period")
		
		// Add warning if there are balances but no transactions in period
		if len(revenueAccounts) > 0 || len(expenseAccounts) > 0 {
			preview.ValidationMessages = append(preview.ValidationMessages,
				"⚠️ Warning: Revenue/Expense accounts have balances from transactions OUTSIDE this period")
			preview.ValidationMessages = append(preview.ValidationMessages,
				"⚠️ Period closing will reset ALL temporary accounts regardless of period")
		}
	} else {
		preview.ValidationMessages = append(preview.ValidationMessages,
			fmt.Sprintf("✅ Found %d posted transactions in this period", preview.TransactionCount))
	}

	if len(revenueAccounts) > 0 || len(expenseAccounts) > 0 {
		preview.ValidationMessages = append(preview.ValidationMessages,
			fmt.Sprintf("📊 Will close %d revenue accounts and %d expense accounts", len(revenueAccounts), len(expenseAccounts)))
		preview.ValidationMessages = append(preview.ValidationMessages,
			"💰 Net Income will be transferred to Retained Earnings")
	} else {
		preview.ValidationMessages = append(preview.ValidationMessages, "ℹ️ No revenue or expense accounts to close")
	}

	return preview, nil
}

// ExecutePeriodClosing performs the actual period closing
func (pcs *PeriodClosingService) ExecutePeriodClosing(ctx context.Context, req models.PeriodClosingRequest) error {
	// Extract user ID from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("user context required: %v", err)
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start date format: %v", err)
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return fmt.Errorf("invalid end date format: %v", err)
	}

	// Get preview first to validate
	preview, err := pcs.PreviewPeriodClosing(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to generate preview: %v", err)
	}

	if !preview.CanClose {
		return fmt.Errorf("cannot close period: %v", preview.ValidationMessages)
	}

	// Execute closing in transaction
	return pcs.db.Transaction(func(tx *gorm.DB) error {
		// Get retained earnings account
		var retainedEarningsAccount models.Account
		if err := tx.First(&retainedEarningsAccount, preview.RetainedEarningsID).Error; err != nil {
			return fmt.Errorf("retained earnings account not found: %v", err)
		}

		// Generate closing journal entry code
		closingJournalCode := fmt.Sprintf("CLO-%s-%s", startDate.Format("2006-01"), endDate.Format("01-02"))

		// Create description
		description := req.Description
		if description == "" {
			description = fmt.Sprintf("Period Closing: %s to %s", startDate.Format("Jan 02, 2006"), endDate.Format("Jan 02, 2006"))
		}

		// Create main closing journal entry
		now := time.Now()
		closingJournal := models.JournalEntry{
			Code:            closingJournalCode,
			Description:     description,
			Reference:       fmt.Sprintf("PERIOD-CLOSING-%s-%s", startDate.Format("20060102"), endDate.Format("20060102")),
			ReferenceType:   models.JournalRefClosing,
			EntryDate:       endDate, // Closing entry dated at end of period
			UserID:          userID,
			Status:          models.JournalStatusPosted,
			PostingDate:     &now,
			PostedBy:        &userID,
			IsAutoGenerated: true,
			IsBalanced:      true,
		}

		var journalLines []models.JournalLine
		lineNumber := 1

		// Close Revenue Accounts (TEMPORARY)
		// IMPORTANT: Only close accounts with actual balances (> 0.01)
		for _, revAccount := range preview.RevenueAccounts {
			if revAccount.Balance > 0.01 {
				// Debit Revenue Account (to zero it out)
				journalLines = append(journalLines, models.JournalLine{
					AccountID:    revAccount.ID,
					Description:  fmt.Sprintf("Close revenue account: %s", revAccount.Name),
					DebitAmount:  revAccount.Balance,
					CreditAmount: 0,
					LineNumber:   lineNumber,
				})
				lineNumber++
			}
		}

		// Credit Retained Earnings with total revenue
		if preview.TotalRevenue > 0 {
			journalLines = append(journalLines, models.JournalLine{
				AccountID:    retainedEarningsAccount.ID,
				Description:  "Transfer total revenue to retained earnings",
				DebitAmount:  0,
				CreditAmount: preview.TotalRevenue,
				LineNumber:   lineNumber,
			})
			lineNumber++
		}

		// Debit Retained Earnings with total expense
		if preview.TotalExpense > 0 {
			journalLines = append(journalLines, models.JournalLine{
				AccountID:    retainedEarningsAccount.ID,
				Description:  "Transfer total expenses from retained earnings",
				DebitAmount:  preview.TotalExpense,
				CreditAmount: 0,
				LineNumber:   lineNumber,
			})
			lineNumber++
		}

		// Close Expense Accounts (TEMPORARY)
		// IMPORTANT: Only close accounts with actual balances (> 0.01)
		for _, expAccount := range preview.ExpenseAccounts {
			if expAccount.Balance > 0.01 {
				// Credit Expense Account (to zero it out)
				journalLines = append(journalLines, models.JournalLine{
					AccountID:    expAccount.ID,
					Description:  fmt.Sprintf("Close expense account: %s", expAccount.Name),
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
		if err := pcs.updateAccountBalancesInTx(tx, &closingJournal); err != nil {
			return fmt.Errorf("failed to update account balances: %v", err)
		}

		netIncome := preview.TotalRevenue - preview.TotalExpense

		// Create AccountingPeriod record
		accountingPeriod := models.AccountingPeriod{
			StartDate:        startDate,
			EndDate:          endDate,
			Description:      description,
			IsClosed:         true,
			IsLocked:         true,
			ClosedBy:         &userID,
			ClosedAt:         &now,
			TotalRevenue:     preview.TotalRevenue,
			TotalExpense:     preview.TotalExpense,
			NetIncome:        netIncome,
			ClosingJournalID: &closingJournal.ID,
			Notes:            req.Notes,
		}

		if err := tx.Create(&accountingPeriod).Error; err != nil {
			return fmt.Errorf("failed to create accounting period record: %v", err)
		}

		// Auto-update fiscal year start in settings to next period start date
		// This ensures fiscal year is always synchronized with the active accounting period
		nextPeriodStart := endDate.AddDate(0, 0, 1)
		newFiscalYearStart := formatFiscalYearStart(nextPeriodStart)
		
		log.Printf("[PERIOD CLOSING] Auto-updating fiscal_year_start to: %s (from %s)", newFiscalYearStart, nextPeriodStart.Format("2006-01-02"))
		
		var settings models.Settings
		if err := tx.First(&settings).Error; err != nil {
			log.Printf("[PERIOD CLOSING] Warning: failed to get settings for fiscal year update: %v", err)
			// Don't fail the entire closing if settings update fails
		} else {
			oldFiscalYearStart := settings.FiscalYearStart
			settings.FiscalYearStart = newFiscalYearStart
			settings.UpdatedBy = userID
			
			if err := tx.Save(&settings).Error; err != nil {
				log.Printf("[PERIOD CLOSING] Warning: failed to update fiscal_year_start: %v", err)
				// Don't fail the entire closing if settings update fails
			} else {
				log.Printf("[PERIOD CLOSING] ✅ Fiscal year start updated: %s → %s", oldFiscalYearStart, newFiscalYearStart)
			}
		}

		// Log the closing activity
		pcs.logger.LogProcessingInfo(ctx, "Period Closing Completed", map[string]interface{}{
			"start_date":           startDate,
			"end_date":             endDate,
			"net_income":           netIncome,
			"total_revenue":        preview.TotalRevenue,
			"total_expense":        preview.TotalExpense,
			"journal_entry_code":   closingJournalCode,
			"journal_entry_id":     closingJournal.ID,
			"accounting_period_id": accountingPeriod.ID,
			"user_id":              userID,
			"new_fiscal_year_start": newFiscalYearStart,
		})

	return nil
	})
}

// updateAccountBalancesInTx updates account balances based on journal lines
// This mirrors the logic in journal_entry_repository.go
func (pcs *PeriodClosingService) updateAccountBalancesInTx(tx *gorm.DB, entry *models.JournalEntry) error {
	log.Printf("[PERIOD CLOSING] Updating account balances for %d journal lines", len(entry.JournalLines))
	
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

		log.Printf("[PERIOD CLOSING] Processing %s (%s) - Type: %s, Current Balance: %.2f", 
			account.Code, account.Name, account.Type, account.Balance)

		// Calculate balance change based on account type (normal balance)
		var balanceChange float64
		if account.Type == models.AccountTypeAsset || account.Type == models.AccountTypeExpense {
			// Debit normal accounts: debit increases, credit decreases
			balanceChange = line.DebitAmount - line.CreditAmount
		} else {
			// Credit normal accounts (LIABILITY, EQUITY, REVENUE): credit increases, debit decreases
			balanceChange = line.CreditAmount - line.DebitAmount
		}

		log.Printf("[PERIOD CLOSING] Line: Debit=%.2f, Credit=%.2f, Balance Change=%.2f", 
			line.DebitAmount, line.CreditAmount, balanceChange)

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
		
		log.Printf("[PERIOD CLOSING] ✅ Updated %s - Rows affected: %d", account.Code, result.RowsAffected)
	}

	log.Println("[PERIOD CLOSING] All account balances updated successfully")
	return nil
}

// GetClosingHistory returns history of period closings
func (pcs *PeriodClosingService) GetClosingHistory(ctx context.Context, limit int) ([]models.AccountingPeriod, error) {
	var periods []models.AccountingPeriod
	
	if limit <= 0 {
		limit = 20
	}

	err := pcs.db.Preload("ClosedByUser").
		Preload("ClosingJournal").
		Where("is_closed = ?", true).
		Order("end_date DESC").
		Limit(limit).
		Find(&periods).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get closing history: %v", err)
	}

	return periods, nil
}

// IsDateInClosedPeriod checks if a given date falls within a closed period
func (pcs *PeriodClosingService) IsDateInClosedPeriod(ctx context.Context, date time.Time) (bool, error) {
	var count int64
	err := pcs.db.Model(&models.AccountingPeriod{}).
		Where("is_closed = ? AND ? BETWEEN start_date AND end_date", true, date).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check closed period: %v", err)
	}

	return count > 0, nil
}

// GetPeriodInfoForDate returns period information for a specific date
func (pcs *PeriodClosingService) GetPeriodInfoForDate(ctx context.Context, date time.Time) map[string]interface{} {
	var period models.AccountingPeriod
	err := pcs.db.Preload("ClosedByUser").Where("? BETWEEN start_date AND end_date AND is_closed = ?", date, true).
		First(&period).Error
	
	if err != nil {
		return nil
	}
	
	info := map[string]interface{}{
		"start_date":  period.StartDate,
		"end_date":    period.EndDate,
		"description": period.Description,
		"is_locked":   period.IsLocked,
	}
	
	if period.ClosedBy != nil {
		info["closed_by"] = *period.ClosedBy
	}
	
	if period.ClosedByUser.ID != 0 {
		info["closed_by_name"] = period.ClosedByUser.GetDisplayName()
	}
	
	return info
}

// formatFiscalYearStart converts a date to fiscal year start format (e.g., "January 1")
func formatFiscalYearStart(date time.Time) string {
	months := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
	month := months[date.Month()-1]
	day := date.Day()
	return fmt.Sprintf("%s %d", month, day)
}

// ReopenPeriod reopens a closed period (if not locked)
func (pcs *PeriodClosingService) ReopenPeriod(ctx context.Context, startDate, endDate time.Time, reason string) error {
	// Extract user ID from context
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("user context required: %v", err)
	}

	return pcs.db.Transaction(func(tx *gorm.DB) error {
		// Find the period
		var period models.AccountingPeriod
		err := tx.Where("start_date = ? AND end_date = ? AND is_closed = ?", 
			startDate, endDate, true).First(&period).Error
		
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("closed period not found for dates %s to %s", 
					startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
			}
			return fmt.Errorf("failed to find period: %v", err)
		}

		// Check if period is locked (cannot reopen)
		if period.IsLocked {
			return fmt.Errorf("period is permanently locked and cannot be reopened")
		}

		// Check if there's a newer closed period (cannot reopen if newer periods exist)
		var newerCount int64
		tx.Model(&models.AccountingPeriod{}).
			Where("is_closed = ? AND start_date > ?", true, period.EndDate).
			Count(&newerCount)
		
		if newerCount > 0 {
			return fmt.Errorf("cannot reopen period: newer closed periods exist")
		}

		// Get the closing journal entry
		if period.ClosingJournalID != nil {
			// Delete or reverse the closing journal entry
			var closingJournal models.JournalEntry
			if err := tx.First(&closingJournal, *period.ClosingJournalID).Error; err == nil {
				// Create reversal entry
				now := time.Now()
				reversalJournal := models.JournalEntry{
					Code:            fmt.Sprintf("REV-%s", closingJournal.Code),
					Description:     fmt.Sprintf("Reversal: %s - Reason: %s", closingJournal.Description, reason),
					Reference:       closingJournal.Reference + "-REVERSAL",
					ReferenceType:   models.JournalRefClosing,
					EntryDate:       time.Now(),
					UserID:          userID,
					Status:          models.JournalStatusPosted,
					PostingDate:     &now,
					PostedBy:        &userID,
					IsAutoGenerated: true,
					IsBalanced:      true,
					TotalDebit:      closingJournal.TotalCredit,
					TotalCredit:     closingJournal.TotalDebit,
				}
				
				if err := tx.Create(&reversalJournal).Error; err != nil {
					return fmt.Errorf("failed to create reversal journal: %v", err)
				}

				// Create reversal lines (swap debit/credit)
				var originalLines []models.JournalLine
				tx.Where("journal_entry_id = ?", closingJournal.ID).Find(&originalLines)
				
				for _, line := range originalLines {
					reversalLine := models.JournalLine{
						JournalEntryID: reversalJournal.ID,
						AccountID:      line.AccountID,
						Description:    "Reversal: " + line.Description,
						DebitAmount:    line.CreditAmount,
						CreditAmount:   line.DebitAmount,
						LineNumber:     line.LineNumber,
					}
					tx.Create(&reversalLine)

					// Get account type to determine correct balance calculation
					var account models.Account
					if err := tx.Select("id, code, name, type").First(&account, line.AccountID).Error; err != nil {
						return fmt.Errorf("failed to get account details for reversal: %v", err)
					}
					
					// Calculate balance change for REVERSAL based on account type
					// Since we're reversing: debit becomes credit, credit becomes debit
					var balanceChange float64
					if account.Type == models.AccountTypeAsset || account.Type == models.AccountTypeExpense {
						// Debit normal accounts: debit increases, credit decreases
						// For reversal: swap the original amounts
						balanceChange = line.CreditAmount - line.DebitAmount
					} else {
						// Credit normal accounts (LIABILITY, EQUITY, REVENUE): credit increases, debit decreases
						// For reversal: swap the original amounts
						balanceChange = line.DebitAmount - line.CreditAmount
					}
					
					tx.Model(&models.Account{}).
						Where("id = ?", line.AccountID).
						Update("balance", gorm.Expr("balance + ?", balanceChange))
				}
			}
		}

		// Reopen the period
		now := time.Now()
		updates := map[string]interface{}{
			"is_closed":  false,
			"closed_by":  nil,
			"closed_at":  nil,
			"updated_at": now,
			"notes":      fmt.Sprintf("%s\n[REOPENED %s by UserID %d]: %s", 
				period.Notes, now.Format("2006-01-02 15:04:05"), userID, reason),
		}
		
		if err := tx.Model(&period).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to reopen period: %v", err)
		}

		// Log the reopen activity
		pcs.logger.LogProcessingInfo(ctx, "Period Reopened", map[string]interface{}{
			"period_id":  period.ID,
			"start_date": startDate,
			"end_date":   endDate,
			"reason":     reason,
			"user_id":    userID,
		})

		return nil
	})
}
