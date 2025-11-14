package models

import (
	"time"

	"gorm.io/gorm"
)

// ProjectBudget represents budget allocation per project per COA account
// Backed by the project_budgets table (see migrations/20251114_create_project_budgets_table.sql)
type ProjectBudget struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ProjectID       uint           `json:"project_id" gorm:"not null;index"`
	AccountID       uint           `json:"account_id" gorm:"not null;index"`
	EstimatedAmount float64        `json:"estimated_amount" gorm:"type:decimal(15,2);not null;default:0"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations (optional, useful for joins)
	Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Account *Account `json:"account,omitempty" gorm:"foreignKey:AccountID"`
}

// TableName overrides the default table name
func (ProjectBudget) TableName() string {
	return "project_budgets"
}

// ProjectBudgetUpsertRequest is used by the API to create/update budgets per project
// This is intentionally small and focused on the fields the client needs to send.
type ProjectBudgetUpsertRequest struct {
	AccountID       uint    `json:"account_id" binding:"required"`
	EstimatedAmount float64 `json:"estimated_amount" binding:"required"`
}

