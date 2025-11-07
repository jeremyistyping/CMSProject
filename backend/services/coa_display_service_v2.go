package services

import (
	"fmt"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// COADisplayServiceV2 handles COA balance display logic for frontend
type COADisplayServiceV2 struct {
	db *gorm.DB
}

// NewCOADisplayServiceV2 creates a new instance of COADisplayServiceV2
func NewCOADisplayServiceV2(db *gorm.DB) *COADisplayServiceV2 {
	return &COADisplayServiceV2{
		db: db,
	}
}

// COADisplay represents COA data formatted for frontend display
type COADisplay struct {
	ID            uint    `json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Category      string  `json:"category"`
	Balance       float64 `json:"balance"`         // Raw balance from database
	DisplayBalance float64 `json:"display_balance"` // Formatted balance for display
	IsPositive    bool    `json:"is_positive"`     // Whether to display as positive
	ParentID      *uint   `json:"parent_id"`
	IsHeader      bool    `json:"is_header"`
	Level         int     `json:"level"`
}

// GetCOAForDisplay retrieves COA with proper balance formatting for frontend
func (s *COADisplayServiceV2) GetCOAForDisplay() ([]COADisplay, error) {
	var coas []models.COA
	if err := s.db.Find(&coas).Error; err != nil {
		return nil, fmt.Errorf("failed to get COA: %v", err)
	}

	var displayCOAs []COADisplay
	for _, coa := range coas {
		displayCOA := COADisplay{
			ID:       coa.ID,
			Code:     coa.Code,
			Name:     coa.Name,
			Type:     coa.Type,
			Category: coa.Category,
			Balance:  coa.Balance,
			ParentID: coa.ParentID,
			IsHeader: coa.IsHeader,
			Level:    coa.Level,
		}

		// Apply display logic based on account type
		displayCOA.DisplayBalance, displayCOA.IsPositive = s.getDisplayBalance(coa)
		
		displayCOAs = append(displayCOAs, displayCOA)
	}

	return displayCOAs, nil
}

// GetSingleCOAForDisplay retrieves a single COA with proper balance formatting
func (s *COADisplayServiceV2) GetSingleCOAForDisplay(accountID uint) (*COADisplay, error) {
	var coa models.COA
	if err := s.db.First(&coa, accountID).Error; err != nil {
		return nil, fmt.Errorf("COA not found: %v", err)
	}

	displayBalance, isPositive := s.getDisplayBalance(coa)
	
	return &COADisplay{
		ID:             coa.ID,
		Code:           coa.Code,
		Name:           coa.Name,
		Type:           coa.Type,
		Category:       coa.Category,
		Balance:        coa.Balance,
		DisplayBalance: displayBalance,
		IsPositive:     isPositive,
		ParentID:       coa.ParentID,
		IsHeader:       coa.IsHeader,
		Level:          coa.Level,
	}, nil
}

// getDisplayBalance determines how to display the balance based on account type
func (s *COADisplayServiceV2) getDisplayBalance(coa models.COA) (displayBalance float64, isPositive bool) {
	// RULES FOR DISPLAY:
	// 1. REVENUE accounts (4xxx) - Show as POSITIVE (convert negative to positive)
	// 2. LIABILITY accounts like PPN Keluaran (2103) - Show as POSITIVE
	// 3. ASSET and EXPENSE - Show as stored (normally positive)
	// 4. EQUITY - Show as POSITIVE (convert negative to positive)

	switch coa.Type {
	case "REVENUE":
		// Revenue accounts are stored as negative but should display as positive
		if coa.Balance < 0 {
			displayBalance = -coa.Balance // Convert to positive
			isPositive = true
		} else {
			displayBalance = coa.Balance
			isPositive = true
		}

	case "LIABILITY":
		// Liability accounts (like PPN Keluaran) should display as positive
		if coa.Balance < 0 {
			displayBalance = -coa.Balance // Convert to positive
			isPositive = true
		} else {
			displayBalance = coa.Balance
			isPositive = true
		}

	case "EQUITY":
		// Equity accounts should display as positive
		if coa.Balance < 0 {
			displayBalance = -coa.Balance // Convert to positive
			isPositive = true
		} else {
			displayBalance = coa.Balance
			isPositive = true
		}

	case "ASSET":
		// Asset accounts display as stored (normally positive)
		displayBalance = coa.Balance
		isPositive = coa.Balance >= 0

	case "EXPENSE":
		// Expense accounts display as stored (normally positive)
		displayBalance = coa.Balance
		isPositive = coa.Balance >= 0

	default:
		// Default: show as stored
		displayBalance = coa.Balance
		isPositive = coa.Balance >= 0
	}

	// Ensure we never show negative zero
	if displayBalance == -0 {
		displayBalance = 0
	}

	return displayBalance, isPositive
}

// GetAccountBalancesByType retrieves account balances grouped by type with display formatting
func (s *COADisplayServiceV2) GetAccountBalancesByType() (map[string][]COADisplay, error) {
	accountTypes := []string{"ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE"}
	result := make(map[string][]COADisplay)

	for _, accountType := range accountTypes {
		var coas []models.COA
		if err := s.db.Where("type = ? AND is_header = ?", accountType, false).Find(&coas).Error; err != nil {
			return nil, fmt.Errorf("failed to get %s accounts: %v", accountType, err)
		}

		var displayCOAs []COADisplay
		for _, coa := range coas {
			displayBalance, isPositive := s.getDisplayBalance(coa)
			displayCOA := COADisplay{
				ID:             coa.ID,
				Code:           coa.Code,
				Name:           coa.Name,
				Type:           coa.Type,
				Category:       coa.Category,
				Balance:        coa.Balance,
				DisplayBalance: displayBalance,
				IsPositive:     isPositive,
				ParentID:       coa.ParentID,
				IsHeader:       coa.IsHeader,
				Level:          coa.Level,
			}
			displayCOAs = append(displayCOAs, displayCOA)
		}
		result[accountType] = displayCOAs
	}

	return result, nil
}

// UpdateCOABalanceFromJournal updates COA balance from journal entries (internal use)
func (s *COADisplayServiceV2) UpdateCOABalanceFromJournal(accountID uint, debit, credit float64, tx *gorm.DB) error {
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	var coa models.COA
	if err := dbToUse.First(&coa, accountID).Error; err != nil {
		return fmt.Errorf("COA not found: %v", err)
	}

	// Calculate net change based on account type
	netChange := debit - credit
	
	switch coa.Type {
	case "ASSET", "EXPENSE":
		// Assets and Expenses increase with debit
		coa.Balance += netChange
	case "LIABILITY", "EQUITY", "REVENUE":
		// Liabilities, Equity, and Revenue increase with credit
		// Store as negative for accounting convention
		coa.Balance -= netChange
	}

	return dbToUse.Save(&coa).Error
}

// GetSpecificAccountsForDisplay gets specific accounts for sales display
func (s *COADisplayServiceV2) GetSpecificAccountsForDisplay(accountCodes []string) (map[string]COADisplay, error) {
	result := make(map[string]COADisplay)
	
	for _, code := range accountCodes {
		var coa models.COA
		if err := s.db.Where("code = ?", code).First(&coa).Error; err != nil {
			continue // Skip if not found
		}
		
		displayBalance, isPositive := s.getDisplayBalance(coa)
		result[code] = COADisplay{
			ID:             coa.ID,
			Code:           coa.Code,
			Name:           coa.Name,
			Type:           coa.Type,
			Category:       coa.Category,
			Balance:        coa.Balance,
			DisplayBalance: displayBalance,
			IsPositive:     isPositive,
			ParentID:       coa.ParentID,
			IsHeader:       coa.IsHeader,
			Level:          coa.Level,
		}
	}
	
	return result, nil
}