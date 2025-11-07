package models

import (
	"time"
	"gorm.io/gorm"
)

type Budget struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"unique;not null;size:20"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Description string         `json:"description" gorm:"type:text"`
	Year        int            `json:"year" gorm:"not null"`
	Status      string         `json:"status" gorm:"not null;size:20;default:'DRAFT'"` // DRAFT, APPROVED, ACTIVE, CLOSED
	TotalBudget float64        `json:"total_budget" gorm:"type:decimal(20,2);default:0"`
	TotalActual float64        `json:"total_actual" gorm:"type:decimal(20,2);default:0"`
	Variance    float64        `json:"variance" gorm:"type:decimal(20,2);default:0"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	ApprovedBy  *uint          `json:"approved_by" gorm:"index"`
	ApprovedAt  *time.Time     `json:"approved_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User        User         `json:"user" gorm:"foreignKey:UserID"`
	Approver    *User        `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
	BudgetItems []BudgetItem `json:"budget_items" gorm:"foreignKey:BudgetID"`
}

type BudgetItem struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	BudgetID      uint           `json:"budget_id" gorm:"not null;index"`
	AccountID     uint           `json:"account_id" gorm:"not null;index"`
	Month         int            `json:"month" gorm:"not null"` // 1-12
	BudgetAmount  float64        `json:"budget_amount" gorm:"type:decimal(20,2);default:0"`
	ActualAmount  float64        `json:"actual_amount" gorm:"type:decimal(20,2);default:0"`
	Variance      float64        `json:"variance" gorm:"type:decimal(20,2);default:0"`
	VariancePercent float64      `json:"variance_percent" gorm:"type:decimal(5,2);default:0"`
	Notes         string         `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Budget  Budget  `json:"budget" gorm:"foreignKey:BudgetID"`
	Account Account `json:"account" gorm:"foreignKey:AccountID"`
}

type BudgetComparison struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Description string         `json:"description" gorm:"type:text"`
	Period      string         `json:"period" gorm:"not null;size:20"` // MONTHLY, QUARTERLY, YEARLY
	Year        int            `json:"year" gorm:"not null"`
	Month       *int           `json:"month"` // For monthly comparison
	Quarter     *int           `json:"quarter"` // For quarterly comparison
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// Budget Status Constants
const (
	BudgetStatusDraft    = "DRAFT"
	BudgetStatusApproved = "APPROVED"
	BudgetStatusActive   = "ACTIVE"
	BudgetStatusClosed   = "CLOSED"
)

// Budget Comparison Period Constants
const (
	ComparisonPeriodMonthly    = "MONTHLY"
	ComparisonPeriodQuarterly  = "QUARTERLY"
	ComparisonPeriodYearly     = "YEARLY"
)
