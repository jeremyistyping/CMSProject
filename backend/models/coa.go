package models

import (
	"time"

	"gorm.io/gorm"
)

// COAAccount represents a Chart of Account entry
type COAAccount struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Code           string         `json:"code" gorm:"unique;not null;size:20"`
	Name           string         `json:"name" gorm:"not null;size:100"`
	Description    string         `json:"description" gorm:"type:text"`
	Type           string         `json:"type" gorm:"not null;size:30"`    // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
	Category       string         `json:"category" gorm:"size:50"`         // Material, Labor, Equipment, Overhead, Subcontractor
	BudgetCategory string         `json:"budget_category" gorm:"size:50"`  // LABOUR_BUDGET, OPERASIONAL_BUDGET, OTHER
	WorkPackage    string         `json:"work_package" gorm:"size:100"`    // Pekerjaan Persiapan, Pekerjaan Beton, etc
	ParentID       *uint          `json:"parent_id" gorm:"index"`
	Level          int            `json:"level" gorm:"default:1"`
	IsActive       bool           `json:"is_active" gorm:"default:true"`
	IsHeader       bool           `json:"is_header" gorm:"default:false"` // Header accounts cannot have transactions
	Balance        float64        `json:"balance" gorm:"type:decimal(20,2);default:0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Parent   *COAAccount  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []COAAccount `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// TableName specifies the table name for COAAccount
func (COAAccount) TableName() string {
	return "coa_accounts"
}

// COA Type Constants
const (
	COATypeAsset     = "ASSET"
	COATypeLiability = "LIABILITY"
	COATypeEquity    = "EQUITY"
	COATypeRevenue   = "REVENUE"
	COATypeExpense   = "EXPENSE"
)

// COA Category Constants for Construction
const (
	COACategoryMaterial      = "Material"
	COACategoryLabor         = "Labor"
	COACategoryEquipment     = "Equipment"
	COACategoryOverhead      = "Overhead"
	COACategorySubcontractor = "Subcontractor"
	COACategoryOther         = "Other"
)

// COA Budget Category Constants
const (
	COABudgetCategoryLabour       = "LABOUR_BUDGET"
	COABudgetCategoryOperasional  = "OPERASIONAL_BUDGET"
	COABudgetCategoryOther        = "OTHER"
)

// COATreeNode represents a node in the COA tree structure
type COATreeNode struct {
	ID             uint           `json:"id"`
	Code           string         `json:"code"`
	Name           string         `json:"name"`
	Type           string         `json:"type"`
	Category       string         `json:"category"`
	BudgetCategory string         `json:"budget_category"`
	WorkPackage    string         `json:"work_package"`
	Level          int            `json:"level"`
	IsActive       bool           `json:"is_active"`
	IsHeader       bool           `json:"is_header"`
	Balance        float64        `json:"balance"`
	Children       []*COATreeNode `json:"children,omitempty"`
}

// CreateCOADTO represents the data transfer object for creating a COA
type CreateCOADTO struct {
	Code           string `json:"code" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	Type           string `json:"type" binding:"required"`
	Category       string `json:"category"`
	BudgetCategory string `json:"budget_category"`
	WorkPackage    string `json:"work_package"`
	ParentID       *uint  `json:"parent_id"`
	IsHeader       bool   `json:"is_header"`
}

// UpdateCOADTO represents the data transfer object for updating a COA
type UpdateCOADTO struct {
	Code           string `json:"code"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Type           string `json:"type"`
	Category       string `json:"category"`
	BudgetCategory string `json:"budget_category"`
	WorkPackage    string `json:"work_package"`
	ParentID       *uint  `json:"parent_id"`
	IsActive       *bool  `json:"is_active"`
	IsHeader       *bool  `json:"is_header"`
}
