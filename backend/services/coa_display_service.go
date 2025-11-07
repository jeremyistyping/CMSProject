package services

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// COADisplayAccount represents an account with display-friendly balance
type COADisplayAccount struct {
	ID             uint    `json:"id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Category       string  `json:"category"`
	RawBalance     float64 `json:"raw_balance"`     // Balance dari database (asli)
	DisplayBalance float64 `json:"display_balance"` // Balance untuk ditampilkan (sudah dikoreksi)
	IsActive       bool    `json:"is_active"`
	ParentID       *uint   `json:"parent_id,omitempty"`
	HasChildren    bool    `json:"has_children"`
	Level          int     `json:"level"`
}

// COADisplayService handles Chart of Accounts display logic
type COADisplayService struct {
	db *gorm.DB
}

// NewCOADisplayService creates a new COA display service
func NewCOADisplayService(db *gorm.DB) *COADisplayService {
	return &COADisplayService{db: db}
}

// GetCOAForDisplay returns accounts formatted for COA display with correct balance signs
// Only includes active accounts with non-zero balances
func (s *COADisplayService) GetCOAForDisplay() ([]COADisplayAccount, error) {
	var accounts []models.Account
	
	// Get active accounts with non-zero balances (what should appear in COA)
	err := s.db.Where("is_active = ? AND balance != 0", true).
		Order("code ASC").
		Find(&accounts).Error
		
	if err != nil {
		return nil, err
	}
		
	var displayAccounts []COADisplayAccount
	
	for _, acc := range accounts {
		displayAccount := COADisplayAccount{
			ID:             acc.ID,
			Code:           acc.Code,
			Name:           acc.Name,
			Type:           acc.Type,
			Category:       acc.Category,
			RawBalance:     acc.Balance,
			DisplayBalance: s.getDisplayBalance(acc.Balance, acc.Type),
			IsActive:       acc.IsActive,
			ParentID:       acc.ParentID,
			HasChildren:    s.hasChildren(acc.ID),
			Level:          s.getAccountLevel(acc.Code),
		}
		
		displayAccounts = append(displayAccounts, displayAccount)
	}
	
	return displayAccounts, nil
}

// GetAllAccountsForDisplay returns ALL accounts (including zero balance) for account management
func (s *COADisplayService) GetAllAccountsForDisplay() ([]COADisplayAccount, error) {
	var accounts []models.Account
	
	err := s.db.Where("is_active = ?", true).
		Order("code ASC").
		Find(&accounts).Error
		
	if err != nil {
		return nil, err
	}
		
	var displayAccounts []COADisplayAccount
	
	for _, acc := range accounts {
		displayAccount := COADisplayAccount{
			ID:             acc.ID,
			Code:           acc.Code,
			Name:           acc.Name,
			Type:           acc.Type,
			Category:       acc.Category,
			RawBalance:     acc.Balance,
			DisplayBalance: s.getDisplayBalance(acc.Balance, acc.Type),
			IsActive:       acc.IsActive,
			ParentID:       acc.ParentID,
			HasChildren:    s.hasChildren(acc.ID),
			Level:          s.getAccountLevel(acc.Code),
		}
		
		displayAccounts = append(displayAccounts, displayAccount)
	}
	
	return displayAccounts, nil
}

// getDisplayBalance converts raw balance to display balance based on account type
func (s *COADisplayService) getDisplayBalance(rawBalance float64, accountType string) float64 {
	switch accountType {
	case "ASSET", "EXPENSE":
		// ASSET dan EXPENSE: tampilkan as-is
		// Positive balance = positive display (normal)
		return rawBalance
		
	case "REVENUE", "LIABILITY", "EQUITY":
		// REVENUE, LIABILITY, EQUITY: flip sign untuk display
		// Negative balance (normal for these account types) → positive display
		// Ini karena di accounting system, credit increases these accounts
		// tapi user expect melihat positive numbers di reports
		return rawBalance * -1
		
	default:
		// Default: tampilkan as-is
		return rawBalance
	}
}

// hasChildren checks if an account has child accounts
func (s *COADisplayService) hasChildren(accountID uint) bool {
	var count int64
	s.db.Model(&models.Account{}).
		Where("parent_id = ? AND is_active = ?", accountID, true).
		Count(&count)
	return count > 0
}

// getAccountLevel determines account hierarchy level based on code
func (s *COADisplayService) getAccountLevel(code string) int {
	switch len(code) {
	case 1, 2, 3:
		return 0 // Header level
	case 4:
		return 1 // Main account level  
	case 5, 6:
		return 2 // Sub account level
	default:
		return 3 // Detail level
	}
}

// GetAccountsByType returns accounts grouped by type for structured display
func (s *COADisplayService) GetAccountsByType() (map[string][]COADisplayAccount, error) {
	accounts, err := s.GetCOAForDisplay()
	if err != nil {
		return nil, err
	}
	
	groupedAccounts := make(map[string][]COADisplayAccount)
	
	for _, account := range accounts {
		groupedAccounts[account.Type] = append(groupedAccounts[account.Type], account)
	}
	
	return groupedAccounts, nil
}

// GetBalanceSheetAccounts returns accounts formatted for Balance Sheet
func (s *COADisplayService) GetBalanceSheetAccounts() (map[string][]COADisplayAccount, error) {
	accounts, err := s.GetCOAForDisplay()
	if err != nil {
		return nil, err
	}
	
	balanceSheet := make(map[string][]COADisplayAccount)
	balanceSheet["ASSETS"] = []COADisplayAccount{}
	balanceSheet["LIABILITIES"] = []COADisplayAccount{}
	balanceSheet["EQUITY"] = []COADisplayAccount{}
	
	for _, account := range accounts {
		switch account.Type {
		case "ASSET":
			balanceSheet["ASSETS"] = append(balanceSheet["ASSETS"], account)
		case "LIABILITY":
			balanceSheet["LIABILITIES"] = append(balanceSheet["LIABILITIES"], account)
		case "EQUITY":
			balanceSheet["EQUITY"] = append(balanceSheet["EQUITY"], account)
		}
	}
	
	return balanceSheet, nil
}

// GetIncomeStatementAccounts returns accounts formatted for Profit & Loss
func (s *COADisplayService) GetIncomeStatementAccounts() (map[string][]COADisplayAccount, error) {
	accounts, err := s.GetCOAForDisplay()
	if err != nil {
		return nil, err
	}
	
	incomeStatement := make(map[string][]COADisplayAccount)
	incomeStatement["REVENUE"] = []COADisplayAccount{}
	incomeStatement["EXPENSE"] = []COADisplayAccount{}
	
	for _, account := range accounts {
		switch account.Type {
		case "REVENUE":
			incomeStatement["REVENUE"] = append(incomeStatement["REVENUE"], account)
		case "EXPENSE":
			incomeStatement["EXPENSE"] = append(incomeStatement["EXPENSE"], account)
		}
	}
	
	return incomeStatement, nil
}