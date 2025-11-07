package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// UnifiedPeriodClosingService handles period closing using unified journal system (SSOT)
type UnifiedPeriodClosingService struct {
	db                    *gorm.DB
	unifiedJournalService *UnifiedJournalService
	logger                *utils.JournalLogger
}

// NewUnifiedPeriodClosingService creates a new unified period closing service
func NewUnifiedPeriodClosingService(db *gorm.DB) *UnifiedPeriodClosingService {
	return &UnifiedPeriodClosingService{
		db:                    db,
		unifiedJournalService: NewUnifiedJournalService(db),
		logger:                utils.NewJournalLogger(db),
	}
}

// ExecutePeriodClosing performs period closing using unified journal system
func (s *UnifiedPeriodClosingService) ExecutePeriodClosing(ctx context.Context, startDate, endDate time.Time, description string, userID uint64) error {
	log.Printf("[UNIFIED CLOSING] Starting period closing: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get retained earnings account
		var retainedEarnings models.Account
		if err := tx.Where("code = ? AND type = ?", "3201", "EQUITY").First(&retainedEarnings).Error; err != nil {
			return fmt.Errorf("retained earnings account (3201) not found: %v", err)
		}

		// 2. Get all revenue accounts with non-zero balances
		var revenueAccounts []models.Account
		if err := tx.Where("type = ? AND ABS(balance) > 0.01 AND is_header = false", "REVENUE").
			Find(&revenueAccounts).Error; err != nil {
			return fmt.Errorf("failed to get revenue accounts: %v", err)
		}

		// 3. Get all expense accounts with non-zero balances
		var expenseAccounts []models.Account
		if err := tx.Where("type = ? AND ABS(balance) > 0.01 AND is_header = false", "EXPENSE").
			Find(&expenseAccounts).Error; err != nil {
			return fmt.Errorf("failed to get expense accounts: %v", err)
		}

		if len(revenueAccounts) == 0 && len(expenseAccounts) == 0 {
			log.Println("[UNIFIED CLOSING] No revenue or expense accounts to close")
			return nil
		}

		// 4. Calculate totals
		// IMPORTANT: Revenue accounts have NEGATIVE balances (credit normal)
		// Expense accounts have POSITIVE balances (debit normal)
		var totalRevenue, totalExpense decimal.Decimal
		for _, acc := range revenueAccounts {
			// Revenue balance is negative, so negate to get positive amount
			totalRevenue = totalRevenue.Add(decimal.NewFromFloat(-acc.Balance))
		}
		for _, acc := range expenseAccounts {
			// Expense balance is positive
			totalExpense = totalExpense.Add(decimal.NewFromFloat(acc.Balance))
		}

		netIncome := totalRevenue.Sub(totalExpense)

		log.Printf("[UNIFIED CLOSING] Total Revenue: %.2f, Total Expense: %.2f, Net Income: %.2f",
			totalRevenue.InexactFloat64(), totalExpense.InexactFloat64(), netIncome.InexactFloat64())

		// 5. Create unified journal entry for closing
		var journalLines []models.SSOTJournalLine
		lineNum := 1

		// Close Revenue accounts (Debit Revenue, Credit Retained Earnings)
		// Revenue has NEGATIVE balance, we debit it to make it zero
		for _, acc := range revenueAccounts {
			if acc.Balance < -0.01 { // Revenue balance is negative
				journalLines = append(journalLines, models.SSOTJournalLine{
					AccountID:    uint64(acc.ID),
					LineNumber:   lineNum,
					Description:  fmt.Sprintf("Close revenue account: %s", acc.Name),
					DebitAmount:  decimal.NewFromFloat(-acc.Balance), // Negate to get positive debit
					CreditAmount: decimal.Zero,
				})
				lineNum++
			}
		}

		// Credit Retained Earnings with total revenue
		if totalRevenue.GreaterThan(decimal.Zero) {
			journalLines = append(journalLines, models.SSOTJournalLine{
				AccountID:    uint64(retainedEarnings.ID),
				LineNumber:   lineNum,
				Description:  "Transfer revenue to retained earnings",
				DebitAmount:  decimal.Zero,
				CreditAmount: totalRevenue,
			})
			lineNum++
		}

		// Debit Retained Earnings with total expense
		if totalExpense.GreaterThan(decimal.Zero) {
			journalLines = append(journalLines, models.SSOTJournalLine{
				AccountID:    uint64(retainedEarnings.ID),
				LineNumber:   lineNum,
				Description:  "Transfer expense from retained earnings",
				DebitAmount:  totalExpense,
				CreditAmount: decimal.Zero,
			})
			lineNum++
		}

		// Close Expense accounts (Debit Retained Earnings, Credit Expense)
		for _, acc := range expenseAccounts {
			if acc.Balance > 0.01 {
				journalLines = append(journalLines, models.SSOTJournalLine{
					AccountID:    uint64(acc.ID),
					LineNumber:   lineNum,
					Description:  fmt.Sprintf("Close expense account: %s", acc.Name),
					DebitAmount:  decimal.Zero,
					CreditAmount: decimal.NewFromFloat(acc.Balance),
				})
				lineNum++
			}
		}

		// 6. Create unified journal entry
		closingEntry := &models.SSOTJournalEntry{
			SourceType:      "CLOSING",
			EntryDate:       endDate,
			Description:     description,
			TotalDebit:      totalRevenue.Add(totalExpense),
			TotalCredit:     totalRevenue.Add(totalExpense),
			Status:          "POSTED",
			IsBalanced:      true,
			IsAutoGenerated: true,
			CreatedBy:       userID,
			Lines:           journalLines,
		}

		now := time.Now()
		closingEntry.PostedAt = &now
		closingEntry.PostedBy = &userID

		// Create journal entry in unified system (this will automatically update balances)
		if err := tx.Create(closingEntry).Error; err != nil {
			return fmt.Errorf("failed to create unified closing journal: %v", err)
		}

		log.Printf("[UNIFIED CLOSING] Created unified journal entry ID: %d with %d lines", closingEntry.ID, len(journalLines))

		// 7. Update account balances based on the journal entry
		// Revenue accounts: balance becomes 0 (debited by their balance)
		// Expense accounts: balance becomes 0 (credited by their balance)
		// Retained Earnings: increases by net income
		for _, line := range journalLines {
			var account models.Account
			if err := tx.First(&account, line.AccountID).Error; err != nil {
				return fmt.Errorf("failed to find account %d: %v", line.AccountID, err)
			}

			var balanceChange float64
			if account.Type == "ASSET" || account.Type == "EXPENSE" {
				// Debit normal: debit increases, credit decreases
				balanceChange = line.DebitAmount.InexactFloat64() - line.CreditAmount.InexactFloat64()
			} else {
				// Credit normal: credit increases, debit decreases
				balanceChange = line.CreditAmount.InexactFloat64() - line.DebitAmount.InexactFloat64()
			}

			if err := tx.Model(&models.Account{}).
				Where("id = ?", line.AccountID).
				Update("balance", gorm.Expr("balance + ?", balanceChange)).Error; err != nil {
				return fmt.Errorf("failed to update account %d balance: %v", line.AccountID, err)
			}

			log.Printf("[UNIFIED CLOSING] Updated account %s (%s): balance change = %.2f",
				account.Code, account.Name, balanceChange)
		}

		// 8. Create accounting period record
		userIDUint := uint(userID)
		accountingPeriod := models.AccountingPeriod{
			StartDate:    startDate,
			EndDate:      endDate,
			Description:  description,
			IsClosed:     true,
			IsLocked:     true,
			ClosedBy:     &userIDUint,
			ClosedAt:     &now,
			TotalRevenue: totalRevenue.InexactFloat64(),
			TotalExpense: totalExpense.InexactFloat64(),
			NetIncome:    netIncome.InexactFloat64(),
		}

		if err := tx.Create(&accountingPeriod).Error; err != nil {
			return fmt.Errorf("failed to create accounting period: %v", err)
		}

		log.Printf("[UNIFIED CLOSING] ✅ Period closing completed successfully")
		log.Printf("[UNIFIED CLOSING] Net Income: %.2f transferred to Retained Earnings", netIncome.InexactFloat64())

		return nil
	})
}

// PreviewPeriodClosing generates preview of period closing
func (s *UnifiedPeriodClosingService) PreviewPeriodClosing(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Get retained earnings account
	var retainedEarnings models.Account
	if err := s.db.Where("code = ? AND type = ?", "3201", "EQUITY").First(&retainedEarnings).Error; err != nil {
		return nil, fmt.Errorf("retained earnings account (3201) not found: %v", err)
	}

	// Get revenue and expense accounts
	var revenueAccounts []models.Account
	s.db.Where("type = ? AND ABS(balance) > 0.01 AND is_header = false", "REVENUE").Find(&revenueAccounts)

	var expenseAccounts []models.Account
	s.db.Where("type = ? AND ABS(balance) > 0.01 AND is_header = false", "EXPENSE").Find(&expenseAccounts)

	// Calculate totals
	var totalRevenue, totalExpense float64
	for _, acc := range revenueAccounts {
		totalRevenue += acc.Balance
	}
	for _, acc := range expenseAccounts {
		totalExpense += acc.Balance
	}

	netIncome := totalRevenue - totalExpense

	return map[string]interface{}{
		"start_date":        startDate.Format("2006-01-02"),
		"end_date":          endDate.Format("2006-01-02"),
		"total_revenue":     totalRevenue,
		"total_expense":     totalExpense,
		"net_income":        netIncome,
		"revenue_accounts":  len(revenueAccounts),
		"expense_accounts":  len(expenseAccounts),
		"can_close":         len(revenueAccounts) > 0 || len(expenseAccounts) > 0,
		"retained_earnings": retainedEarnings.Name,
	}, nil
}

// GetLastClosingInfo returns information about the last closed period
func (s *UnifiedPeriodClosingService) GetLastClosingInfo(ctx context.Context) (map[string]interface{}, error) {
	var lastPeriod models.AccountingPeriod
	err := s.db.Where("is_closed = ?", true).
		Order("end_date DESC").
		First(&lastPeriod).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return map[string]interface{}{
				"has_previous_closing": false,
			}, nil
		}
		return nil, fmt.Errorf("failed to query last closing period: %v", err)
	}

	nextStart := lastPeriod.EndDate.AddDate(0, 0, 1)

	return map[string]interface{}{
		"has_previous_closing": true,
		"last_closing_date":    lastPeriod.EndDate.Format("2006-01-02"),
		"next_start_date":      nextStart.Format("2006-01-02"),
		"last_net_income":      lastPeriod.NetIncome,
	}, nil
}

// IsDateInClosedPeriod checks if a given date falls within a closed period
func (s *UnifiedPeriodClosingService) IsDateInClosedPeriod(ctx context.Context, date time.Time) (bool, error) {
	var count int64
	err := s.db.Model(&models.AccountingPeriod{}).
		Where("is_closed = ? AND ? BETWEEN start_date AND end_date", true, date).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check closed period: %v", err)
	}

	return count > 0, nil
}

// GetPeriodInfoForDate returns period information for a specific date
func (s *UnifiedPeriodClosingService) GetPeriodInfoForDate(ctx context.Context, date time.Time) map[string]interface{} {
	var period models.AccountingPeriod
	err := s.db.Preload("ClosedByUser").Where("? BETWEEN start_date AND end_date AND is_closed = ?", date, true).
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
