package models

import (
	"time"
	"gorm.io/gorm"
)

// Quote status constants
const (
	QuoteStatusDraft     = "DRAFT"
	QuoteStatusSent      = "SENT"
	QuoteStatusAccepted  = "ACCEPTED"
	QuoteStatusRejected  = "REJECTED"
	QuoteStatusExpired   = "EXPIRED"
	QuoteStatusCancelled = "CANCELLED"
)

// Quote represents a sales quotation
type Quote struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Basic Information
	Code         string    `json:"code" gorm:"uniqueIndex;not null"`
	CustomerID   uint      `json:"customer_id" gorm:"not null"`
	Customer     Contact   `json:"customer" gorm:"foreignKey:CustomerID"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	User         User      `json:"user" gorm:"foreignKey:UserID"`
	Date         time.Time `json:"date" gorm:"not null"`
	ValidUntil   time.Time `json:"valid_until" gorm:"not null"`
	
	// Financial Information
	SubtotalBeforeDiscount float64 `json:"subtotal_before_discount" gorm:"type:decimal(15,2);default:0"`
	Discount               float64 `json:"discount" gorm:"type:decimal(15,2);default:0"`
	SubtotalAfterDiscount  float64 `json:"subtotal_after_discount" gorm:"type:decimal(15,2);default:0"`
	TaxAmount              float64 `json:"tax_amount" gorm:"type:decimal(15,2);default:0"`
	TotalAmount            float64 `json:"total_amount" gorm:"type:decimal(15,2);default:0"`
	
	// Tax Information
	PPNRate               *float64 `json:"ppn_rate" gorm:"type:decimal(5,2)"`
	PPh21Rate             *float64 `json:"pph21_rate" gorm:"type:decimal(5,2)"`
	PPh23Rate             *float64 `json:"pph23_rate" gorm:"type:decimal(5,2)"`
	OtherTaxAdditions     *float64 `json:"other_tax_additions" gorm:"type:decimal(15,2)"`
	OtherTaxDeductions    *float64 `json:"other_tax_deductions" gorm:"type:decimal(15,2)"`
	
	// Status and Notes
	Status string `json:"status" gorm:"default:'DRAFT'"`
	Notes  string `json:"notes"`
	Terms  string `json:"terms"` // Terms and conditions
	
	// Conversion tracking
	ConvertedToInvoice bool  `json:"converted_to_invoice" gorm:"default:false"`
	InvoiceID          *uint `json:"invoice_id,omitempty"`
	Invoice            *Invoice `json:"invoice,omitempty" gorm:"foreignKey:InvoiceID"`
	
	// Items
	QuoteItems []QuoteItem `json:"quote_items" gorm:"foreignKey:QuoteID;constraint:OnDelete:CASCADE"`
}

// QuoteItem represents an item in a quotation
type QuoteItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	QuoteID       uint    `json:"quote_id" gorm:"not null"`
	Quote         Quote   `json:"quote" gorm:"foreignKey:QuoteID"`
	ProductID     uint    `json:"product_id" gorm:"not null"`
	Product       Product `json:"product" gorm:"foreignKey:ProductID"`
	Quantity      int     `json:"quantity" gorm:"not null"`
	UnitPrice     float64 `json:"unit_price" gorm:"type:decimal(15,2);not null"`
	TotalPrice    float64 `json:"total_price" gorm:"type:decimal(15,2);not null"`
	Description   string  `json:"description"`
}

// TableName specifies the table name for Quote
func (Quote) TableName() string {
	return "quotes"
}

// TableName specifies the table name for QuoteItem
func (QuoteItem) TableName() string {
	return "quote_items"
}

// QuoteCreateRequest represents the request to create a new quote
type QuoteCreateRequest struct {
	CustomerID         uint                      `json:"customer_id" binding:"required"`
	Date               time.Time                 `json:"date" binding:"required"`
	ValidUntil         time.Time                 `json:"valid_until" binding:"required"`
	Discount           float64                   `json:"discount"`
	PPNRate            *float64                  `json:"ppn_rate"`
	PPh21Rate          *float64                  `json:"pph21_rate"`
	PPh23Rate          *float64                  `json:"pph23_rate"`
	OtherTaxAdditions  *float64                  `json:"other_tax_additions"`
	OtherTaxDeductions *float64                  `json:"other_tax_deductions"`
	Notes              string                    `json:"notes"`
	Terms              string                    `json:"terms"`
	Items              []QuoteItemCreateRequest  `json:"items" binding:"required,min=1"`
}

// QuoteItemCreateRequest represents the request to create a new quote item
type QuoteItemCreateRequest struct {
	ProductID   uint    `json:"product_id" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" binding:"required,min=0"`
	Description string  `json:"description"`
}

// QuoteUpdateRequest represents the request to update a quote
type QuoteUpdateRequest struct {
	CustomerID         *uint                      `json:"customer_id"`
	Date               *time.Time                 `json:"date"`
	ValidUntil         *time.Time                 `json:"valid_until"`
	Discount           *float64                   `json:"discount"`
	PPNRate            *float64                   `json:"ppn_rate"`
	PPh21Rate          *float64                   `json:"pph21_rate"`
	PPh23Rate          *float64                   `json:"pph23_rate"`
	OtherTaxAdditions  *float64                   `json:"other_tax_additions"`
	OtherTaxDeductions *float64                   `json:"other_tax_deductions"`
	Notes              *string                    `json:"notes"`
	Terms              *string                    `json:"terms"`
	Items              []QuoteItemCreateRequest   `json:"items"`
}

// QuoteFilter represents filter parameters for quote queries
type QuoteFilter struct {
	Status     string `json:"status"`
	CustomerID string `json:"customer_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Search     string `json:"search"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

// QuoteSummary represents summary statistics for quotes
type QuoteSummary struct {
	TotalQuotes     int64   `json:"total_quotes"`
	TotalAmount     float64 `json:"total_amount"`
	DraftCount      int64   `json:"draft_count"`
	SentCount       int64   `json:"sent_count"`
	AcceptedCount   int64   `json:"accepted_count"`
	RejectedCount   int64   `json:"rejected_count"`
	ExpiredCount    int64   `json:"expired_count"`
	ConversionRate  float64 `json:"conversion_rate"`
}