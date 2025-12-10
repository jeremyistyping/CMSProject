package models

import (
	"time"

	"gorm.io/gorm"
)

// ExpenseTransaction represents a project expense transaction
type ExpenseTransaction struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ProjectID       uint           `json:"project_id" gorm:"not null;index"`
	TransactionDate time.Time      `json:"transaction_date" gorm:"not null;index;type:date"`
	COAAccountID    uint           `json:"coa_account_id" gorm:"not null;index"`
	Description     string         `json:"description" gorm:"not null;size:500"`
	Amount          float64        `json:"amount" gorm:"type:decimal(15,2);not null;default:0"`
	Unit            string         `json:"unit" gorm:"size:20;default:'ls'"`
	Quantity        float64        `json:"quantity" gorm:"type:decimal(10,2);default:1"`
	TransactionType string         `json:"transaction_type" gorm:"size:30;index"` // LABOUR, MATERIAL, OPERATIONAL, OTHER
	ReferenceType   string         `json:"reference_type" gorm:"size:30"`         // PR, PO, MANUAL, CBS
	ReferenceID     *uint          `json:"reference_id"`
	ReferenceNo     string         `json:"reference_no" gorm:"size:50"`
	Notes           string         `json:"notes" gorm:"type:text"`
	CreatedBy       uint           `json:"created_by" gorm:"not null;index"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Project    Project     `json:"project" gorm:"foreignKey:ProjectID"`
	COAAccount COAAccount  `json:"coa_account" gorm:"foreignKey:COAAccountID"`
	Creator    User        `json:"creator" gorm:"foreignKey:CreatedBy"`
}

// TableName specifies the table name for ExpenseTransaction
func (ExpenseTransaction) TableName() string {
	return "expense_transactions"
}

// Transaction Type Constants
const (
	ExpenseTypeLabour      = "LABOUR"
	ExpenseTypeMaterial    = "MATERIAL"
	ExpenseTypeOperational = "OPERATIONAL"
	ExpenseTypeOther       = "OTHER"
)

// Reference Type Constants
const (
	ExpenseRefTypePR     = "PR"
	ExpenseRefTypePO     = "PO"
	ExpenseRefTypeManual = "MANUAL"
	ExpenseRefTypeCBS    = "CBS"
)

// CreateExpenseTransactionDTO represents the data transfer object for creating an expense transaction
type CreateExpenseTransactionDTO struct {
	ProjectID       uint    `json:"project_id" binding:"required"`
	TransactionDate string  `json:"transaction_date" binding:"required"` // Format: YYYY-MM-DD
	COAAccountID    uint    `json:"coa_account_id" binding:"required"`
	Description     string  `json:"description" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Unit            string  `json:"unit"`
	Quantity        float64 `json:"quantity"`
	TransactionType string  `json:"transaction_type"`
	ReferenceType   string  `json:"reference_type"`
	ReferenceID     *uint   `json:"reference_id"`
	ReferenceNo     string  `json:"reference_no"`
	Notes           string  `json:"notes"`
}

// UpdateExpenseTransactionDTO represents the data transfer object for updating an expense transaction
type UpdateExpenseTransactionDTO struct {
	TransactionDate *time.Time `json:"transaction_date"`
	COAAccountID    *uint      `json:"coa_account_id"`
	Description     string     `json:"description"`
	Amount          *float64   `json:"amount"`
	Unit            string     `json:"unit"`
	Quantity        *float64   `json:"quantity"`
	TransactionType string     `json:"transaction_type"`
	ReferenceType   string     `json:"reference_type"`
	ReferenceID     *uint      `json:"reference_id"`
	ReferenceNo     string     `json:"reference_no"`
	Notes           string     `json:"notes"`
}

// BudgetVsActualSummary represents budget vs actual summary per COA
type BudgetVsActualSummary struct {
	ProjectID          uint    `json:"project_id"`
	COAAccountID       uint    `json:"coa_account_id"`
	COACode            string  `json:"coa_code"`
	COAName            string  `json:"coa_name"`
	BudgetCategory     string  `json:"budget_category"`
	WorkPackage        string  `json:"work_package"`
	BudgetEstimation   float64 `json:"budget_estimation"`
	ActualAmount       float64 `json:"actual_amount"`
	Variance           float64 `json:"variance"`
	VariancePercentage float64 `json:"variance_percentage"`
}

// BudgetReportResponse represents the complete budget report response
type BudgetReportResponse struct {
	ProjectID     uint                      `json:"project_id"`
	ProjectName   string                    `json:"project_name"`
	ReportDate    time.Time                 `json:"report_date"`
	StartDate     time.Time                 `json:"start_date"`
	EndDate       time.Time                 `json:"end_date"`
	LabourBudget  *BudgetCategoryReport     `json:"labour_budget"`
	OperationalBudget *BudgetCategoryReport `json:"operasional_budget"`
	OtherBudget   *BudgetCategoryReport     `json:"other_budget"`
}

// BudgetCategoryReport represents budget report for a specific category
type BudgetCategoryReport struct {
	BudgetEstimation float64                    `json:"budget_estimation"`
	Actual           float64                    `json:"actual"`
	Variance         float64                    `json:"variance"`
	Transactions     []ExpenseTransactionDetail `json:"transactions"`
	ByWorkPackage    []WorkPackageSummary       `json:"by_work_package,omitempty"`
}

// ExpenseTransactionDetail represents detailed transaction for report
type ExpenseTransactionDetail struct {
	Date            time.Time `json:"date"`
	Description     string    `json:"description"`
	Unit            string    `json:"unit"`
	Quantity        float64   `json:"quantity"`
	TotalPrice      float64   `json:"total_price"`
	COACode         string    `json:"coa_code"`
	COAName         string    `json:"coa_name"`
	WorkPackage     string    `json:"work_package,omitempty"`
	ReferenceNo     string    `json:"reference_no,omitempty"`
}

// WorkPackageSummary represents summary per work package
type WorkPackageSummary struct {
	WorkPackage      string                     `json:"work_package"`
	BudgetEstimation float64                    `json:"budget_estimation"`
	Actual           float64                    `json:"actual"`
	Variance         float64                    `json:"variance"`
	Transactions     []ExpenseTransactionDetail `json:"transactions"`
}
