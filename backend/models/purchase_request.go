package models

import (
	"time"

	"gorm.io/gorm"
)

type PurchaseRequest struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Code         string    `json:"code" gorm:"unique;not null;size:20"`
	ProjectID    uint      `json:"project_id" gorm:"not null;index"`
	RequestDate  time.Time `json:"request_date" gorm:"not null"`
	RequiredDate time.Time `json:"required_date"`
	VendorID     *uint     `json:"vendor_id" gorm:"index"` // Optional suggested vendor
	Notes        string    `json:"notes" gorm:"type:text"`
	Status       string    `json:"status" gorm:"size:20;default:'PENDING'"` // PENDING, APPROVED, REJECTED, REVISION, PO_CREATED
	TotalAmount  float64   `json:"total_amount" gorm:"type:decimal(15,2);default:0"`

	// Approval fields
	ApprovedBy      *uint      `json:"approved_by" gorm:"index"`
	ApprovedAt      *time.Time `json:"approved_at"`
	RejectionReason string     `json:"rejection_reason" gorm:"type:text"`

	CreatedBy uint           `json:"created_by" gorm:"not null;index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Project   Project               `json:"project" gorm:"foreignKey:ProjectID"`
	Vendor    *Contact              `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	Requester User                  `json:"requester" gorm:"foreignKey:CreatedBy"`
	Approver  *User                 `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
	Items     []PurchaseRequestItem `json:"items" gorm:"foreignKey:PurchaseRequestID"`
}

type PurchaseRequestItem struct {
	ID                uint    `json:"id" gorm:"primaryKey"`
	PurchaseRequestID uint    `json:"purchase_request_id" gorm:"not null;index"`
	ItemName          string  `json:"item_name" gorm:"not null;size:255"` // Can be free text or product name
	ProductID         *uint   `json:"product_id" gorm:"index"`            // Optional link to product catalog
	Quantity          float64 `json:"quantity" gorm:"type:decimal(10,2);not null"`
	Unit              string  `json:"unit" gorm:"size:50"`
	EstimatedPrice    float64 `json:"estimated_price" gorm:"type:decimal(15,2);default:0"`
	TotalPrice        float64 `json:"total_price" gorm:"type:decimal(15,2);default:0"`
	Notes             string  `json:"notes" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	PurchaseRequest PurchaseRequest `json:"purchase_request" gorm:"foreignKey:PurchaseRequestID"`
	Product         *Product        `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

// Purchase Request Status Constants
const (
	PRStatusPending   = "PENDING"
	PRStatusApproved  = "APPROVED"
	PRStatusRejected  = "REJECTED"
	PRStatusRevision  = "REVISION"
	PRStatusPOCreated = "PO_CREATED"
)
