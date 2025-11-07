package models

import (
	"time"
	"gorm.io/gorm"
)

type Report struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"unique;not null;size:20"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Type        string         `json:"type" gorm:"not null;size:30"` // BALANCE_SHEET, INCOME_STATEMENT, CASH_FLOW, etc.
	Period      string         `json:"period" gorm:"not null;size:20"` // MONTHLY, QUARTERLY, YEARLY, CUSTOM
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	Status      string         `json:"status" gorm:"not null;size:20;default:'DRAFT'"` // DRAFT, GENERATED, PUBLISHED
	FilePath    string         `json:"file_path" gorm:"size:500"`
	FileFormat  string         `json:"file_format" gorm:"size:10"` // PDF, EXCEL, CSV
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	GeneratedAt *time.Time     `json:"generated_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User User `json:"user" gorm:"foreignKey:UserID"`
}

type ReportTemplate struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Type        string         `json:"type" gorm:"not null;size:30"`
	Description string         `json:"description" gorm:"type:text"`
	Template    string         `json:"template" gorm:"type:text"` // JSON template structure
	IsDefault   bool           `json:"is_default" gorm:"default:false"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User User `json:"user" gorm:"foreignKey:UserID"`
}

type FinancialRatio struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"not null;size:100"`
	Category     string         `json:"category" gorm:"not null;size:30"` // LIQUIDITY, PROFITABILITY, LEVERAGE, etc.
	Formula      string         `json:"formula" gorm:"not null;type:text"`
	Value        float64        `json:"value" gorm:"type:decimal(15,4)"`
	Period       string         `json:"period" gorm:"not null;size:7"` // YYYY-MM format
	CompanyID    uint           `json:"company_id" gorm:"not null;index"`
	CalculatedAt time.Time      `json:"calculated_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type AccountPeriodBalance struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	AccountID uint           `json:"account_id" gorm:"not null;index"`
	Period    string         `json:"period" gorm:"not null;size:7"` // YYYY-MM format
	Balance   float64        `json:"balance" gorm:"type:decimal(20,2);default:0"`
	DebitTotal  float64      `json:"debit_total" gorm:"type:decimal(20,2);default:0"`
	CreditTotal float64      `json:"credit_total" gorm:"type:decimal(20,2);default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Account Account `json:"account" gorm:"foreignKey:AccountID"`
}

type CompanyProfile struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Name            string         `json:"name" gorm:"not null;size:200"`
	LegalName       string         `json:"legal_name" gorm:"size:200"`
	TaxNumber       string         `json:"tax_number" gorm:"size:50"`
	RegistrationNumber string      `json:"registration_number" gorm:"size:50"`
	Industry        string         `json:"industry" gorm:"size:100"`
	Address         string         `json:"address" gorm:"type:text"`
	City            string         `json:"city" gorm:"size:100"`
	State           string         `json:"state" gorm:"size:100"`
	PostalCode      string         `json:"postal_code" gorm:"size:20"`
	Country         string         `json:"country" gorm:"size:100;default:'Indonesia'"`
	Phone           string         `json:"phone" gorm:"size:20"`
	Email           string         `json:"email" gorm:"size:100"`
	Website         string         `json:"website" gorm:"size:100"`
	Logo            string         `json:"logo" gorm:"size:500"`
	FiscalYearStart string         `json:"fiscal_year_start" gorm:"size:5;default:'01-01'"` // MM-DD format
	Currency        string         `json:"currency" gorm:"size:3;default:'IDR'"`
	SharesOutstanding float64      `json:"shares_outstanding" gorm:"type:decimal(20,4);default:0"` // Number of shares outstanding
	ParValuePerShare  float64      `json:"par_value_per_share" gorm:"type:decimal(15,4);default:1000"` // Par value per share in currency
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// Report Metadata moved to financial_report.go to avoid conflicts

// Report Response DTO
type ReportResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Period      string    `json:"period"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Format      string    `json:"format"`
	Data        interface{} `json:"data,omitempty"`
	FileURL     string    `json:"file_url,omitempty"`
	FileName    string    `json:"file_name,omitempty"`
	FileSize    int64     `json:"file_size,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
	GeneratedBy uint      `json:"generated_by"`
	Status      string    `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Report Template Request DTO
type ReportTemplateRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Type        string `json:"type" binding:"required,max=30"`
	Description string `json:"description"`
	Template    string `json:"template" binding:"required"`
	IsDefault   bool   `json:"is_default"`
}

// Report Types Constants - Basic report types (Financial report types in financial_report.go)
const (
	ReportTypeIncomeStatement = "INCOME_STATEMENT"
	ReportTypeAccountsReceivable = "ACCOUNTS_RECEIVABLE"
	ReportTypeAccountsPayable = "ACCOUNTS_PAYABLE"
	ReportTypeInventory       = "INVENTORY"
	ReportTypeTaxReport       = "TAX_REPORT"
	ReportTypeBudgetComparison = "BUDGET_COMPARISON"
	// Note: Common financial report types moved to financial_report.go
)

// Report Status Constants
const (
	ReportStatusDraft     = "DRAFT"
	ReportStatusGenerated = "GENERATED"
	ReportStatusPublished = "PUBLISHED"
)

// Report Period Constants
const (
	ReportPeriodMonthly   = "MONTHLY"
	ReportPeriodQuarterly = "QUARTERLY"
	ReportPeriodYearly    = "YEARLY"
	ReportPeriodCustom    = "CUSTOM"
)

// Financial Ratio Categories Constants
const (
	RatioCategoryLiquidity     = "LIQUIDITY"
	RatioCategoryProfitability = "PROFITABILITY"
	RatioCategoryLeverage      = "LEVERAGE"
	RatioCategoryEfficiency    = "EFFICIENCY"
	RatioCategoryMarket        = "MARKET"
)
