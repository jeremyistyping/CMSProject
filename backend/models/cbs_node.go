package models

import (
	"time"
)

// CBSNode represents a Cost Breakdown Structure node
type CBSNode struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ProjectID    uint       `gorm:"not null" json:"project_id"`
	ParentID     *uint      `json:"parent_id"`
	Code         string     `gorm:"size:50;not null" json:"code"`
	Name         string     `gorm:"size:255;not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description,omitempty"`
	COAAccountID *uint      `json:"coa_account_id,omitempty"`
	BudgetAmount int64      `gorm:"default:0" json:"budget_amount"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Project  *Project  `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Parent   *CBSNode  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []CBSNode `gorm:"foreignKey:ParentID" json:"children,omitempty"`

	// Computed fields (not stored in DB)
	ActualCost int64 `gorm:"-" json:"actual_cost"`
	Level      int   `gorm:"-" json:"level"`
}

// TableName specifies the table name for CBSNode
func (CBSNode) TableName() string {
	return "cbs_nodes"
}

// PRCBSMapping represents the mapping between PR items and CBS nodes
type PRCBSMapping struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	PurchaseRequestID uint      `gorm:"not null" json:"purchase_request_id"`
	CBSNodeID         uint      `gorm:"not null" json:"cbs_node_id"`
	PRItemID          *uint     `json:"pr_item_id,omitempty"`
	AllocatedAmount   int64     `gorm:"not null" json:"allocated_amount"`
	Notes             string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedBy         *uint     `json:"created_by,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// Relationships
	PurchaseRequest *PurchaseRequest     `gorm:"foreignKey:PurchaseRequestID" json:"purchase_request,omitempty"`
	CBSNode         *CBSNode             `gorm:"foreignKey:CBSNodeID" json:"cbs_node,omitempty"`
	PRItem          *PurchaseRequestItem `gorm:"foreignKey:PRItemID" json:"pr_item,omitempty"`
	Creator         *User                `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`
}

// TableName specifies the table name for PRCBSMapping
func (PRCBSMapping) TableName() string {
	return "pr_cbs_mappings"
}

// CBSNodeSummary represents cost summary for a CBS node
type CBSNodeSummary struct {
	NodeID       uint  `json:"node_id"`
	BudgetAmount int64 `json:"budget_amount"`
	ActualCost   int64 `json:"actual_cost"`
	Variance     int64 `json:"variance"`
	ChildrenCost int64 `json:"children_cost"`
	TotalCost    int64 `json:"total_cost"`
}
