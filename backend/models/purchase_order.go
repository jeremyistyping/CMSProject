package models

import (
	"time"

	"gorm.io/gorm"
)

// Purchase Order Status Constants
const (
	POStatusDraft           = "DRAFT"
	POStatusSent            = "SENT"
	POStatusPartialReceived = "PARTIAL_RECEIVED"
	POStatusCompleted       = "COMPLETED"
	POStatusCancelled       = "CANCELLED"
)

// Goods Receipt Status Constants
const (
	GRStatusPending   = "PENDING"
	GRStatusInspected = "INSPECTED"
	GRStatusAccepted  = "ACCEPTED"
	GRStatusRejected  = "REJECTED"
)

type PurchaseOrder struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	Code              string     `json:"code" gorm:"unique;not null;size:30"`
	PurchaseRequestID uint       `json:"purchase_request_id" gorm:"not null;index"`
	ProjectID         uint       `json:"project_id" gorm:"not null;index"`
	VendorID          *uint      `json:"vendor_id" gorm:"index"`

	OrderDate            time.Time  `json:"order_date" gorm:"type:date;not null"`
	ExpectedDeliveryDate *time.Time `json:"expected_delivery_date" gorm:"type:date"`
	DeliveryAddress      string     `json:"delivery_address" gorm:"type:text"`
	PaymentTerms         string     `json:"payment_terms" gorm:"size:100"`
	Notes                string     `json:"notes" gorm:"type:text"`

	Subtotal       float64 `json:"subtotal" gorm:"type:decimal(15,2);default:0"`
	TaxAmount      float64 `json:"tax_amount" gorm:"type:decimal(15,2);default:0"`
	DiscountAmount float64 `json:"discount_amount" gorm:"type:decimal(15,2);default:0"`
	TotalAmount    float64 `json:"total_amount" gorm:"type:decimal(15,2);default:0"`

	Status string `json:"status" gorm:"size:20;default:'DRAFT'"`

	CreatedBy  uint       `json:"created_by" gorm:"not null;index"`
	ApprovedBy *uint      `json:"approved_by" gorm:"index"`
	ApprovedAt *time.Time `json:"approved_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	PurchaseRequest PurchaseRequest     `json:"purchase_request" gorm:"foreignKey:PurchaseRequestID"`
	Project         Project             `json:"project" gorm:"foreignKey:ProjectID"`
	Vendor          *Vendor             `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	Creator         User                `json:"creator" gorm:"foreignKey:CreatedBy"`
	Approver        *User               `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
	Items           []PurchaseOrderItem `json:"items" gorm:"foreignKey:PurchaseOrderID"`
	GoodsReceipts   []GoodsReceipt      `json:"goods_receipts" gorm:"foreignKey:PurchaseOrderID"`
}

type PurchaseOrderItem struct {
	ID              uint   `json:"id" gorm:"primaryKey"`
	PurchaseOrderID uint   `json:"purchase_order_id" gorm:"not null;index"`
	PRItemID        *uint  `json:"pr_item_id" gorm:"index"`
	MaterialID      *uint  `json:"material_id" gorm:"index"`
	ItemName        string `json:"item_name" gorm:"not null;size:255"`
	Description     string `json:"description" gorm:"type:text"`

	Quantity   float64 `json:"quantity" gorm:"type:decimal(10,2);not null"`
	Unit       string  `json:"unit" gorm:"size:50"`
	UnitPrice  float64 `json:"unit_price" gorm:"type:decimal(15,2);default:0"`
	TotalPrice float64 `json:"total_price" gorm:"type:decimal(15,2);default:0"`

	ReceivedQuantity float64 `json:"received_quantity" gorm:"type:decimal(10,2);default:0"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	PurchaseOrder PurchaseOrder `json:"-" gorm:"foreignKey:PurchaseOrderID"`
	PRItem        *PurchaseRequestItem `json:"pr_item,omitempty" gorm:"foreignKey:PRItemID"`
	Material      *Material     `json:"material,omitempty" gorm:"foreignKey:MaterialID"`
}

type GoodsReceipt struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Code            string    `json:"code" gorm:"unique;not null;size:30"`
	PurchaseOrderID uint      `json:"purchase_order_id" gorm:"not null;index"`
	ProjectID       uint      `json:"project_id" gorm:"not null;index"`
	ReceiptDate     time.Time `json:"receipt_date" gorm:"type:date;not null"`
	ReceivedBy      uint      `json:"received_by" gorm:"not null;index"`
	Notes           string    `json:"notes" gorm:"type:text"`
	Status          string    `json:"status" gorm:"size:20;default:'PENDING'"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	PurchaseOrder PurchaseOrder       `json:"purchase_order" gorm:"foreignKey:PurchaseOrderID"`
	Project       Project             `json:"project" gorm:"foreignKey:ProjectID"`
	Receiver      User                `json:"receiver" gorm:"foreignKey:ReceivedBy"`
	Items         []GoodsReceiptItem  `json:"items" gorm:"foreignKey:GoodsReceiptID"`
}

type GoodsReceiptItem struct {
	ID               uint    `json:"id" gorm:"primaryKey"`
	GoodsReceiptID   uint    `json:"goods_receipt_id" gorm:"not null;index"`
	POItemID         uint    `json:"po_item_id" gorm:"not null;index"`
	ReceivedQuantity float64 `json:"received_quantity" gorm:"type:decimal(10,2);not null"`
	AcceptedQuantity float64 `json:"accepted_quantity" gorm:"type:decimal(10,2);default:0"`
	RejectedQuantity float64 `json:"rejected_quantity" gorm:"type:decimal(10,2);default:0"`
	RejectionReason  string  `json:"rejection_reason" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	GoodsReceipt GoodsReceipt      `json:"-" gorm:"foreignKey:GoodsReceiptID"`
	POItem       PurchaseOrderItem `json:"po_item" gorm:"foreignKey:POItemID"`
}

// NullableDate is a custom type that handles empty string as nil
type NullableDate struct {
	time.Time
	Valid bool
}

// UnmarshalJSON implements json.Unmarshaler
func (nd *NullableDate) UnmarshalJSON(b []byte) error {
	s := string(b)
	// Handle null, empty string, or "0001-01-01"
	if s == "null" || s == `""` || s == `"0001-01-01"` {
		nd.Valid = false
		return nil
	}
	
	// Try to parse the date
	var t time.Time
	err := t.UnmarshalJSON(b)
	if err != nil {
		nd.Valid = false
		return nil // Don't return error, just mark as invalid
	}
	
	nd.Time = t
	nd.Valid = true
	return nil
}

// CreatePORequest represents the request to create a PO from PR
type CreatePORequest struct {
	PurchaseRequestID    uint         `json:"purchase_request_id" binding:"required"`
	VendorID             *uint        `json:"vendor_id"`
	ExpectedDeliveryDate NullableDate `json:"expected_delivery_date"`
	DeliveryAddress      string       `json:"delivery_address"`
	PaymentTerms         string       `json:"payment_terms"`
	Notes                string       `json:"notes"`
	Items                []CreatePOItemRequest `json:"items"`
}

type CreatePOItemRequest struct {
	PRItemID   uint    `json:"pr_item_id"`
	MaterialID *uint   `json:"material_id"`
	ItemName   string  `json:"item_name" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required"`
	Unit       string  `json:"unit"`
	UnitPrice  float64 `json:"unit_price" binding:"required"`
}

// CreateGRRequest represents the request to create a Goods Receipt
type CreateGRRequest struct {
	PurchaseOrderID uint                  `json:"purchase_order_id" binding:"required"`
	ReceiptDate     string                `json:"receipt_date"` // Accept as string to handle date-only format
	Notes           string                `json:"notes"`
	Items           []CreateGRItemRequest `json:"items" binding:"required"`
}

type CreateGRItemRequest struct {
	POItemID         uint    `json:"po_item_id" binding:"required"`
	ReceivedQuantity float64 `json:"received_quantity" binding:"required"`
	AcceptedQuantity float64 `json:"accepted_quantity"`
	RejectedQuantity float64 `json:"rejected_quantity"`
	RejectionReason  string  `json:"rejection_reason"`
}
