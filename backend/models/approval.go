package models

import (
	"time"
	"gorm.io/gorm"
)

// ApprovalWorkflow defines the approval workflow configuration
type ApprovalWorkflow struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Name            string         `json:"name" gorm:"not null;size:100"`
	Module          string         `json:"module" gorm:"not null;size:50"` // SALES, PURCHASE, EXPENSE, etc
	MinAmount       float64        `json:"min_amount" gorm:"type:decimal(15,2);default:0"`
	MaxAmount       float64        `json:"max_amount" gorm:"type:decimal(15,2)"`
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	RequireDirector bool           `json:"require_director" gorm:"default:false"`
	RequireFinance  bool           `json:"require_finance" gorm:"default:false"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Steps []ApprovalStep `json:"steps" gorm:"foreignKey:WorkflowID"`
}

// ApprovalStep defines individual steps in approval workflow
type ApprovalStep struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	WorkflowID   uint           `json:"workflow_id" gorm:"not null;index"`
	StepOrder    int            `json:"step_order" gorm:"not null"`
	StepName     string         `json:"step_name" gorm:"not null;size:100"`
	ApproverRole string         `json:"approver_role" gorm:"not null;size:50"` // finance, director, admin
	IsOptional   bool           `json:"is_optional" gorm:"default:false"`
	IsParallel   bool           `json:"is_parallel" gorm:"default:false"` // Allow parallel approval with other steps in same order
	TimeLimit    int            `json:"time_limit" gorm:"default:24"` // hours
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Workflow ApprovalWorkflow `json:"workflow" gorm:"foreignKey:WorkflowID"`
}

// ApprovalRequest represents an approval request for a transaction
type ApprovalRequest struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	RequestCode      string         `json:"request_code" gorm:"unique;not null;size:50"`
	WorkflowID       uint           `json:"workflow_id" gorm:"not null;index"`
	RequesterID      uint           `json:"requester_id" gorm:"not null;index"`
	EntityType       string         `json:"entity_type" gorm:"not null;size:20"` // SALE, PURCHASE
	EntityID         uint           `json:"entity_id" gorm:"not null;index"`
	Amount           float64        `json:"amount" gorm:"type:decimal(15,2);not null"`
	Status           string         `json:"status" gorm:"size:20;default:'PENDING'"` // PENDING, APPROVED, REJECTED, CANCELLED
	Priority         string         `json:"priority" gorm:"size:20;default:'NORMAL'"` // LOW, NORMAL, HIGH, URGENT
	RequestTitle     string         `json:"request_title" gorm:"not null;size:200"`
	RequestMessage   string         `json:"request_message" gorm:"type:text"`
	RejectReason     string         `json:"reject_reason" gorm:"type:text"`
	CompletedAt      *time.Time     `json:"completed_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Workflow       ApprovalWorkflow `json:"workflow" gorm:"foreignKey:WorkflowID"`
	Requester      User             `json:"requester" gorm:"foreignKey:RequesterID"`
	ApprovalSteps  []ApprovalAction `json:"approval_steps" gorm:"foreignKey:RequestID"`
	ApprovalHistory []ApprovalHistory `json:"approval_history" gorm:"foreignKey:RequestID"`
}

// ApprovalAction represents individual approval actions
type ApprovalAction struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	RequestID    uint           `json:"request_id" gorm:"not null;index"`
	StepID       uint           `json:"step_id" gorm:"not null;index"`
	ApproverID   *uint          `json:"approver_id" gorm:"index"`
	Status       string         `json:"status" gorm:"size:20;default:'PENDING'"` // PENDING, APPROVED, REJECTED, SKIPPED
	Comments     string         `json:"comments" gorm:"type:text"`
	ActionDate   *time.Time     `json:"action_date"`
	IsActive     bool           `json:"is_active" gorm:"default:false"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Request  ApprovalRequest `json:"request" gorm:"foreignKey:RequestID"`
	Step     ApprovalStep    `json:"step" gorm:"foreignKey:StepID"`
	Approver *User           `json:"approver" gorm:"foreignKey:ApproverID"`
}

// ApprovalHistory stores complete approval history
type ApprovalHistory struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	RequestID  uint           `json:"request_id" gorm:"not null;index"`
	UserID     uint           `json:"user_id" gorm:"not null;index"`
	Action     string         `json:"action" gorm:"not null;size:50"` // CREATED, APPROVED, REJECTED, CANCELLED
	Comments   string         `json:"comments" gorm:"type:text"`
	Metadata   string         `json:"metadata" gorm:"type:jsonb"` // Additional context data
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Request ApprovalRequest `json:"request" gorm:"foreignKey:RequestID"`
	User    User             `json:"user" gorm:"foreignKey:UserID"`
}

// Approval Status Constants
const (
	ApprovalStatusPending   = "PENDING"
	ApprovalStatusApproved  = "APPROVED"
	ApprovalStatusRejected  = "REJECTED"
	ApprovalStatusCancelled = "CANCELLED"
)

// Approval Priority Constants  
const (
	ApprovalPriorityLow    = "LOW"
	ApprovalPriorityNormal = "NORMAL"
	ApprovalPriorityHigh   = "HIGH"
	ApprovalPriorityUrgent = "URGENT"
)

// Approval Module Constants
const (
	ApprovalModuleSales    = "SALES"
	ApprovalModulePurchase = "PURCHASE"
	ApprovalModuleExpense  = "EXPENSE"
	ApprovalModuleAsset    = "ASSET"
)

// Approval Action Constants
const (
	ApprovalActionCreated   = "CREATED"
	ApprovalActionApproved  = "APPROVED"
	ApprovalActionRejected  = "REJECTED"
	ApprovalActionCancelled = "CANCELLED"
	ApprovalActionSkipped   = "SKIPPED"
)

// Entity Type Constants
const (
	EntityTypeSale     = "SALE"
	EntityTypePurchase = "PURCHASE"
	EntityTypeExpense  = "EXPENSE"
	EntityTypeAsset    = "ASSET"
)

// DTOs for API requests/responses
type CreateApprovalWorkflowRequest struct {
	Name            string                        `json:"name" binding:"required"`
	Module          string                        `json:"module" binding:"required"`
	MinAmount       float64                       `json:"min_amount"`
	MaxAmount       float64                       `json:"max_amount"`
	RequireDirector bool                          `json:"require_director"`
	RequireFinance  bool                          `json:"require_finance"`
	Steps           []CreateApprovalStepRequest   `json:"steps"`
}

type CreateApprovalStepRequest struct {
	StepOrder    int    `json:"step_order" binding:"required"`
	StepName     string `json:"step_name" binding:"required"`
	ApproverRole string `json:"approver_role" binding:"required"`
	IsOptional   bool   `json:"is_optional"`
	TimeLimit    int    `json:"time_limit"`
}

type CreateApprovalRequestDTO struct {
	EntityType     string  `json:"entity_type" binding:"required"`
	EntityID       uint    `json:"entity_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	Priority       string  `json:"priority"`
	RequestTitle   string  `json:"request_title" binding:"required"`
	RequestMessage string  `json:"request_message"`
}

type ApprovalActionDTO struct {
	Action   string `json:"action" binding:"required"` // APPROVE, REJECT
	Comments string `json:"comments"`
}
