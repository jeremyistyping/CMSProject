package models

import (
	"time"
	"gorm.io/gorm"
)

// COA is an alias for Account (Chart of Accounts)
type COA = Account

type Account struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"not null;size:20;uniqueIndex:idx_accounts_code_active,where:deleted_at IS NULL"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Description string         `json:"description" gorm:"type:text"`
	Type        string         `json:"type" gorm:"not null;size:20"` // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
	Category    string         `json:"category" gorm:"size:50"`      // CURRENT_ASSET, FIXED_ASSET, etc.
	ParentID    *uint          `json:"parent_id" gorm:"index"`
	Level       int            `json:"level" gorm:"default:1"`
	IsHeader    bool           `json:"is_header" gorm:"default:false"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	IsSystemCritical bool      `json:"is_system_critical" gorm:"default:false"` // Lock critical accounts from modification
	Balance     float64        `json:"balance" gorm:"type:decimal(20,2);default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Calculated fields (not stored in DB)
	TotalBalance float64        `json:"total_balance" gorm:"-"` // Balance + sum of all children balances
	ChildCount   int            `json:"child_count" gorm:"-"`   // Number of child accounts

	// Relations
	Parent       *Account          `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children     []Account         `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Transactions []Transaction     `json:"-" gorm:"foreignKey:AccountID"`
	SaleItems    []SaleItem        `json:"-" gorm:"foreignKey:RevenueAccountID"`
	PurchaseItems []PurchaseItem   `json:"-" gorm:"foreignKey:ExpenseAccountID"`
	Assets       []Asset           `json:"-" gorm:"foreignKey:AssetAccountID"`
}


// Account Types Constants
const (
	AccountTypeAsset     = "ASSET"
	AccountTypeLiability = "LIABILITY"
	AccountTypeEquity    = "EQUITY"
	AccountTypeRevenue   = "REVENUE"
	AccountTypeExpense   = "EXPENSE"
)

// Normal Balance Types
type NormalBalanceType string

const (
	NormalBalanceDebit  NormalBalanceType = "DEBIT"  // Assets, Expenses
	NormalBalanceCredit NormalBalanceType = "CREDIT" // Liabilities, Equity, Revenue
)

// Account Categories Constants
const (
	// Balance Sheet Categories
	CategoryCurrentAsset      = "CURRENT_ASSET"
	CategoryFixedAsset        = "FIXED_ASSET"
	CategoryIntangibleAsset   = "INTANGIBLE_ASSET"
	CategoryInvestmentAsset   = "INVESTMENT_ASSET"
	CategoryCurrentLiability  = "CURRENT_LIABILITY"
	CategoryLongTermLiability = "LONG_TERM_LIABILITY"
	CategoryEquity            = "EQUITY"
	CategoryRetainedEarnings  = "RETAINED_EARNINGS"
	CategoryShareCapital      = "SHARE_CAPITAL"
	
	// Profit & Loss Categories - Revenue
	CategoryOperatingRevenue    = "OPERATING_REVENUE"
	CategoryServiceRevenue      = "SERVICE_REVENUE"
	CategorySalesRevenue        = "SALES_REVENUE"
	CategoryNonOperatingRevenue = "NON_OPERATING_REVENUE"
	CategoryInterestIncome      = "INTEREST_INCOME"
	CategoryDividendIncome      = "DIVIDEND_INCOME"
	CategoryGainOnSale          = "GAIN_ON_SALE"
	CategoryOtherIncome         = "OTHER_INCOME"
	
	// Profit & Loss Categories - Cost of Goods Sold
	CategoryCostOfGoodsSold        = "COST_OF_GOODS_SOLD"
	CategoryDirectMaterial         = "DIRECT_MATERIAL"
	CategoryDirectLabor            = "DIRECT_LABOR"
	CategoryManufacturingOverhead  = "MANUFACTURING_OVERHEAD"
	CategoryFreightIn              = "FREIGHT_IN"
	CategoryPurchaseReturns        = "PURCHASE_RETURNS"
	
	// Profit & Loss Categories - Operating Expenses
	CategoryOperatingExpense    = "OPERATING_EXPENSE"
	CategoryAdministrativeExp   = "ADMINISTRATIVE_EXPENSE"
	CategorySellingExpense      = "SELLING_EXPENSE"
	CategoryMarketingExpense    = "MARKETING_EXPENSE"
	CategoryGeneralExpense      = "GENERAL_EXPENSE"
	CategoryDepreciationExp     = "DEPRECIATION_EXPENSE"
	CategoryAmortizationExp     = "AMORTIZATION_EXPENSE"
	CategoryBadDebtExpense      = "BAD_DEBT_EXPENSE"
	
	// Profit & Loss Categories - Non-Operating Expenses
	CategoryNonOperatingExpense = "NON_OPERATING_EXPENSE"
	CategoryInterestExpense     = "INTEREST_EXPENSE"
	CategoryFinancialExpense    = "FINANCIAL_EXPENSE"
	CategoryLossOnSale          = "LOSS_ON_SALE"
	CategoryTaxExpense          = "TAX_EXPENSE"
	CategoryOtherExpense        = "OTHER_EXPENSE"
)

// AccountType enum
type AccountType string

// IsValidAccountType checks if account type is valid
func IsValidAccountType(accountType string) bool {
	types := []string{
		AccountTypeAsset,
		AccountTypeLiability,
		AccountTypeEquity,
		AccountTypeRevenue,
		AccountTypeExpense,
	}
	for _, t := range types {
		if t == accountType {
			return true
		}
	}
	return false
}

// GetNormalBalance returns the normal balance type for the account
func (a *Account) GetNormalBalance() NormalBalanceType {
	switch a.Type {
	case AccountTypeAsset, AccountTypeExpense:
		return NormalBalanceDebit
	case AccountTypeLiability, AccountTypeEquity, AccountTypeRevenue:
		return NormalBalanceCredit
	default:
		return NormalBalanceDebit // Default to debit
	}
}


// Request/Response structures
type AccountCreateRequest struct {
	Code           string      `json:"code" binding:"required,max=20"`
	Name           string      `json:"name" binding:"required,max=100"`
	Type           AccountType `json:"type" binding:"required"`
	Category       string      `json:"category"`
	ParentID       *uint       `json:"parent_id"`
	Description    string      `json:"description"`
	OpeningBalance float64     `json:"opening_balance"`
	IsHeader       *bool       `json:"is_header"` // Allow manual header creation
}

type AccountUpdateRequest struct {
	Code           string      `json:"code" binding:"max=20"`
	Name           string      `json:"name" binding:"required,max=100"`
	Type           AccountType `json:"type"`
	Description    string      `json:"description"`
	Category       string      `json:"category"`
	ParentID       *uint       `json:"parent_id"`
	IsActive       *bool       `json:"is_active"`
	OpeningBalance *float64    `json:"opening_balance"`
	IsHeader       *bool       `json:"is_header"` // Allow manual header update
}

type AccountImportRequest struct {
	Code           string      `json:"code" binding:"required"`
	Name           string      `json:"name" binding:"required"`
	Type           AccountType `json:"type" binding:"required"`
	Category       string      `json:"category"`
	ParentCode     string      `json:"parent_code"`
	Description    string      `json:"description"`
	OpeningBalance float64     `json:"opening_balance"`
}

type AccountSummaryResponse struct {
	Type           AccountType `json:"type"`
	TotalAccounts  int64       `json:"total_accounts"`
	TotalBalance   float64     `json:"total_balance"`
	ActiveAccounts int64       `json:"active_accounts"`
}

