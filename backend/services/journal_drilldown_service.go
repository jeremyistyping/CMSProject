package services

import (
	"context"
	"fmt"
	"log"
	"time"
	"strings"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

type JournalDrilldownService struct {
	journalRepo repositories.JournalEntryRepository
	accountRepo repositories.AccountRepository
	db          *gorm.DB
}

func NewJournalDrilldownService(db *gorm.DB) *JournalDrilldownService {
	return &JournalDrilldownService{
		journalRepo: repositories.NewJournalEntryRepository(db),
		accountRepo: repositories.NewAccountRepository(db),
		db:          db,
	}
}

// JournalDrilldownRequest represents the request for journal drill-down
type JournalDrilldownRequest struct {
	AccountCodes  []string  `json:"account_codes"`  // Specific account codes to filter
	AccountIDs    []uint    `json:"account_ids"`    // Specific account IDs to filter
	StartDate     time.Time `json:"start_date"`     // Date range start
	EndDate       time.Time `json:"end_date"`       // Date range end
	ReportType    string    `json:"report_type"`    // BALANCE_SHEET, PROFIT_LOSS, CASH_FLOW
	LineItemName  string    `json:"line_item_name"` // Name of the line item clicked
	MinAmount     *float64  `json:"min_amount"`     // Minimum amount filter
	MaxAmount     *float64  `json:"max_amount"`     // Maximum amount filter
	TransactionTypes []string `json:"transaction_types"` // SALE, PURCHASE, PAYMENT, etc.
	Page          int       `json:"page"`
	Limit         int       `json:"limit"`
}

// JournalDrilldownResponse represents the response with journal entries
type JournalDrilldownResponse struct {
	JournalEntries []models.JournalEntry `json:"journal_entries"`
	Total          int64                 `json:"total"`
	Summary        JournalEntrySummary   `json:"summary"`
	Metadata       DrilldownMetadata     `json:"metadata"`
}

// JournalEntrySummary provides summary information
type JournalEntrySummary struct {
	TotalDebit       float64 `json:"total_debit"`
	TotalCredit      float64 `json:"total_credit"`
	NetAmount        float64 `json:"net_amount"`
	EntryCount       int64   `json:"entry_count"`
	DateRangeStart   time.Time `json:"date_range_start"`
	DateRangeEnd     time.Time `json:"date_range_end"`
	AccountsInvolved []string  `json:"accounts_involved"`
}

// DrilldownMetadata provides context about the drill-down
type DrilldownMetadata struct {
	ReportType     string    `json:"report_type"`
	LineItemName   string    `json:"line_item_name"`
	FilterCriteria string    `json:"filter_criteria"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// GetJournalEntriesForDrilldown retrieves journal entries for financial report drill-down
func (s *JournalDrilldownService) GetJournalEntriesForDrilldown(ctx context.Context, req *JournalDrilldownRequest) (*JournalDrilldownResponse, error) {
	log.Printf("📊 Journal Drilldown Request: %+v", req)

	// Build the base query
	query := s.db.WithContext(ctx).Model(&models.JournalEntry{}).
		Preload("Account").
		Preload("JournalLines").
		Preload("JournalLines.Account").
		Preload("Creator")

	// Apply date range filter
	query = query.Where("entry_date >= ? AND entry_date <= ?", req.StartDate, req.EndDate)

	// Apply account filters
	if len(req.AccountCodes) > 0 || len(req.AccountIDs) > 0 {
		// Get journal entries that have lines with the specified accounts
		subQuery := s.db.Model(&models.JournalLine{}).Select("journal_entry_id")
		
		if len(req.AccountCodes) > 0 {
			// Get account IDs from codes
			var accountIDs []uint
			s.db.Model(&models.Account{}).Where("code IN ?", req.AccountCodes).Pluck("id", &accountIDs)
			if len(accountIDs) > 0 {
				subQuery = subQuery.Where("account_id IN ?", accountIDs)
			}
		}
		
		if len(req.AccountIDs) > 0 {
			subQuery = subQuery.Or("account_id IN ?", req.AccountIDs)
		}
		
		query = query.Where("id IN (?)", subQuery)
	}

	// Apply amount filters
	if req.MinAmount != nil {
		query = query.Where("(total_debit >= ? OR total_credit >= ?)", *req.MinAmount, *req.MinAmount)
	}
	if req.MaxAmount != nil {
		query = query.Where("(total_debit <= ? OR total_credit <= ?)", *req.MaxAmount, *req.MaxAmount)
	}

	// Apply transaction type filters
	if len(req.TransactionTypes) > 0 {
		query = query.Where("reference_type IN ?", req.TransactionTypes)
	}

	// Only show posted entries for drill-down
	query = query.Where("status = ?", models.JournalStatusPosted)

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count journal entries: %w", err)
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Prevent excessive data load
	}

	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Order by entry date (newest first) and then by creation time
	query = query.Order("entry_date DESC, created_at DESC")

	// Execute query
	var journalEntries []models.JournalEntry
	if err := query.Find(&journalEntries).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch journal entries: %w", err)
	}

	// Calculate summary
	summary, err := s.calculateDrilldownSummary(ctx, req)
	if err != nil {
		log.Printf("⚠️ Failed to calculate summary: %v", err)
		// Continue without summary rather than failing
	}

	// Build metadata
	metadata := DrilldownMetadata{
		ReportType:     req.ReportType,
		LineItemName:   req.LineItemName,
		FilterCriteria: s.buildFilterCriteriaText(req),
		GeneratedAt:    time.Now(),
	}

	response := &JournalDrilldownResponse{
		JournalEntries: journalEntries,
		Total:          total,
		Summary:        summary,
		Metadata:       metadata,
	}

	log.Printf("✅ Journal Drilldown completed: %d entries found", len(journalEntries))
	return response, nil
}

// calculateDrilldownSummary calculates summary statistics for the drill-down
func (s *JournalDrilldownService) calculateDrilldownSummary(ctx context.Context, req *JournalDrilldownRequest) (JournalEntrySummary, error) {
	query := s.db.WithContext(ctx).Model(&models.JournalEntry{})

	// Apply same filters as main query
	query = query.Where("entry_date >= ? AND entry_date <= ?", req.StartDate, req.EndDate)
	query = query.Where("status = ?", models.JournalStatusPosted)

	if len(req.AccountCodes) > 0 || len(req.AccountIDs) > 0 {
		subQuery := s.db.Model(&models.JournalLine{}).Select("journal_entry_id")
		
		if len(req.AccountCodes) > 0 {
			var accountIDs []uint
			s.db.Model(&models.Account{}).Where("code IN ?", req.AccountCodes).Pluck("id", &accountIDs)
			if len(accountIDs) > 0 {
				subQuery = subQuery.Where("account_id IN ?", accountIDs)
			}
		}
		
		if len(req.AccountIDs) > 0 {
			subQuery = subQuery.Or("account_id IN ?", req.AccountIDs)
		}
		
		query = query.Where("id IN (?)", subQuery)
	}

	if req.MinAmount != nil {
		query = query.Where("(total_debit >= ? OR total_credit >= ?)", *req.MinAmount, *req.MinAmount)
	}
	if req.MaxAmount != nil {
		query = query.Where("(total_debit <= ? OR total_credit <= ?)", *req.MaxAmount, *req.MaxAmount)
	}

	if len(req.TransactionTypes) > 0 {
		query = query.Where("reference_type IN ?", req.TransactionTypes)
	}

	// Calculate summary
	var summary struct {
		TotalDebit  float64
		TotalCredit float64
		Count       int64
	}

	if err := query.Select("COALESCE(SUM(total_debit), 0) as total_debit, COALESCE(SUM(total_credit), 0) as total_credit, COUNT(*) as count").Scan(&summary).Error; err != nil {
		return JournalEntrySummary{}, err
	}

	// Get involved account names
	var accountNames []string
	if len(req.AccountCodes) > 0 {
		s.db.Model(&models.Account{}).Where("code IN ?", req.AccountCodes).Pluck("name", &accountNames)
	} else if len(req.AccountIDs) > 0 {
		s.db.Model(&models.Account{}).Where("id IN ?", req.AccountIDs).Pluck("name", &accountNames)
	}

	return JournalEntrySummary{
		TotalDebit:       summary.TotalDebit,
		TotalCredit:      summary.TotalCredit,
		NetAmount:        summary.TotalDebit - summary.TotalCredit,
		EntryCount:       summary.Count,
		DateRangeStart:   req.StartDate,
		DateRangeEnd:     req.EndDate,
		AccountsInvolved: accountNames,
	}, nil
}

// buildFilterCriteriaText creates human-readable filter criteria text
func (s *JournalDrilldownService) buildFilterCriteriaText(req *JournalDrilldownRequest) string {
	var criteria []string

	if len(req.AccountCodes) > 0 {
		criteria = append(criteria, fmt.Sprintf("Account Codes: %s", strings.Join(req.AccountCodes, ", ")))
	}

	if len(req.TransactionTypes) > 0 {
		criteria = append(criteria, fmt.Sprintf("Transaction Types: %s", strings.Join(req.TransactionTypes, ", ")))
	}

	if req.MinAmount != nil || req.MaxAmount != nil {
		amountFilter := "Amount: "
		if req.MinAmount != nil && req.MaxAmount != nil {
			amountFilter += fmt.Sprintf("%.2f - %.2f", *req.MinAmount, *req.MaxAmount)
		} else if req.MinAmount != nil {
			amountFilter += fmt.Sprintf(">= %.2f", *req.MinAmount)
		} else if req.MaxAmount != nil {
			amountFilter += fmt.Sprintf("<= %.2f", *req.MaxAmount)
		}
		criteria = append(criteria, amountFilter)
	}

	criteria = append(criteria, fmt.Sprintf("Date Range: %s - %s", 
		req.StartDate.Format("2006-01-02"), 
		req.EndDate.Format("2006-01-02")))

	if len(criteria) == 0 {
		return "No specific filters applied"
	}

	return strings.Join(criteria, " | ")
}

// GetJournalEntryDetail gets detailed information for a specific journal entry
func (s *JournalDrilldownService) GetJournalEntryDetail(ctx context.Context, journalEntryID uint) (*models.JournalEntry, error) {
	entry, err := s.journalRepo.FindByID(ctx, journalEntryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get journal entry detail: %w", err)
	}

	return entry, nil
}

// GetAccountsForPeriod gets all accounts that have journal entries in a specific period
func (s *JournalDrilldownService) GetAccountsForPeriod(ctx context.Context, startDate, endDate time.Time) ([]models.Account, error) {
	var accounts []models.Account
	
	query := s.db.WithContext(ctx).Model(&models.Account{}).
		Joins("JOIN journal_lines ON accounts.id = journal_lines.account_id").
		Joins("JOIN journal_entries ON journal_lines.journal_entry_id = journal_entries.id").
		Where("journal_entries.entry_date >= ? AND journal_entries.entry_date <= ?", startDate, endDate).
		Where("journal_entries.status = ?", models.JournalStatusPosted).
		Group("accounts.id").
		Order("accounts.code")

	if err := query.Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get accounts for period: %w", err)
	}

	return accounts, nil
}
