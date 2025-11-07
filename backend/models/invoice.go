package models

import (
	"time"
	"gorm.io/gorm"
)

// Invoice status constants
const (
	InvoiceStatusDraft     = "DRAFT"
	InvoiceStatusSent      = "SENT"
	InvoiceStatusPaid      = "PAID"
	InvoiceStatusPartially = "PARTIALLY_PAID"
	InvoiceStatusOverdue   = "OVERDUE"
	InvoiceStatusCancelled = "CANCELLED"
)

// Invoice payment method constants  
const (
	InvoicePaymentCash       = "CASH"
	InvoicePaymentBankTransfer = "BANK_TRANSFER" 
	InvoicePaymentCredit     = "CREDIT"
	InvoicePaymentDebit      = "DEBIT"
	InvoicePaymentCheck      = "CHECK"
)

// Invoice represents a sales invoice
type Invoice struct {
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
	DueDate      time.Time `json:"due_date" gorm:"not null"`
	
	// Financial Information
	SubtotalBeforeDiscount float64 `json:"subtotal_before_discount" gorm:"type:decimal(15,2);default:0"`
	Discount               float64 `json:"discount" gorm:"type:decimal(15,2);default:0"`
	SubtotalAfterDiscount  float64 `json:"subtotal_after_discount" gorm:"type:decimal(15,2);default:0"`
	TaxAmount              float64 `json:"tax_amount" gorm:"type:decimal(15,2);default:0"`
	TotalAmount            float64 `json:"total_amount" gorm:"type:decimal(15,2);default:0"`
	
	// Tax Information
	PPNRate                *float64 `json:"ppn_rate" gorm:"type:decimal(5,2)"`
	PPh21Rate              *float64 `json:"pph21_rate" gorm:"type:decimal(5,2)"`
	PPh23Rate              *float64 `json:"pph23_rate" gorm:"type:decimal(5,2)"`
	OtherTaxAdditions      *float64 `json:"other_tax_additions" gorm:"type:decimal(15,2)"`
	OtherTaxDeductions     *float64 `json:"other_tax_deductions" gorm:"type:decimal(15,2)"`
	
	// Payment Information
	PaymentMethod     string  `json:"payment_method" gorm:"default:'CREDIT'"`
	PaymentReference  string  `json:"payment_reference"`
	BankAccountID     *uint   `json:"bank_account_id"`
	BankAccount       *Account `json:"bank_account,omitempty" gorm:"foreignKey:BankAccountID"`
	
	// Payment Tracking
	PaidAmount        float64 `json:"paid_amount" gorm:"type:decimal(15,2);default:0"`
	OutstandingAmount float64 `json:"outstanding_amount" gorm:"type:decimal(15,2);default:0"`
	
	// Status and Notes
	Status string `json:"status" gorm:"default:'DRAFT'"`
	Notes  string `json:"notes"`
	
	// Items
	InvoiceItems []InvoiceItem `json:"invoice_items" gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE"`
}

// InvoiceItem represents an item in an invoice
type InvoiceItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	InvoiceID     uint    `json:"invoice_id" gorm:"not null"`
	Invoice       Invoice `json:"invoice" gorm:"foreignKey:InvoiceID"`
	ProductID     uint    `json:"product_id" gorm:"not null"`
	Product       Product `json:"product" gorm:"foreignKey:ProductID"`
	Quantity      int     `json:"quantity" gorm:"not null"`
	UnitPrice     float64 `json:"unit_price" gorm:"type:decimal(15,2);not null"`
	TotalPrice    float64 `json:"total_price" gorm:"type:decimal(15,2);not null"`
	Description   string  `json:"description"`
}

// TableName specifies the table name for Invoice
func (Invoice) TableName() string {
	return "invoices"
}

// TableName specifies the table name for InvoiceItem
func (InvoiceItem) TableName() string {
	return "invoice_items"
}

// InvoiceCreateRequest represents the request to create a new invoice
type InvoiceCreateRequest struct {
	CustomerID       uint                       `json:"customer_id" binding:"required"`
	Date             time.Time                  `json:"date" binding:"required"`
	DueDate          time.Time                  `json:"due_date" binding:"required"`
	Discount         float64                    `json:"discount"`
	PaymentMethod    string                     `json:"payment_method"`
	PaymentReference string                     `json:"payment_reference"`
	BankAccountID    *uint                      `json:"bank_account_id"`
	PPNRate          *float64                   `json:"ppn_rate"`
	PPh21Rate        *float64                   `json:"pph21_rate"`
	PPh23Rate        *float64                   `json:"pph23_rate"`
	OtherTaxAdditions *float64                  `json:"other_tax_additions"`
	OtherTaxDeductions *float64                 `json:"other_tax_deductions"`
	Notes            string                     `json:"notes"`
	Items            []InvoiceItemCreateRequest `json:"items" binding:"required,min=1"`
}

// InvoiceItemCreateRequest represents the request to create a new invoice item
type InvoiceItemCreateRequest struct {
	ProductID   uint    `json:"product_id" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" binding:"required,min=0"`
	Description string  `json:"description"`
}

// InvoiceUpdateRequest represents the request to update an invoice
type InvoiceUpdateRequest struct {
	CustomerID       *uint                       `json:"customer_id"`
	Date             *time.Time                  `json:"date"`
	DueDate          *time.Time                  `json:"due_date"`
	Discount         *float64                    `json:"discount"`
	PaymentMethod    *string                     `json:"payment_method"`
	PaymentReference *string                     `json:"payment_reference"`
	BankAccountID    *uint                       `json:"bank_account_id"`
	PPNRate          *float64                    `json:"ppn_rate"`
	PPh21Rate        *float64                    `json:"pph21_rate"`
	PPh23Rate        *float64                    `json:"pph23_rate"`
	OtherTaxAdditions *float64                   `json:"other_tax_additions"`
	OtherTaxDeductions *float64                  `json:"other_tax_deductions"`
	Notes            *string                     `json:"notes"`
	Items            []InvoiceItemCreateRequest  `json:"items"`
}

// InvoiceFilter represents filter parameters for invoice queries
type InvoiceFilter struct {
	Status     string `json:"status"`
	CustomerID string `json:"customer_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Search     string `json:"search"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

// InvoiceSummary represents summary statistics for invoices
type InvoiceSummary struct {
	TotalInvoices    int64   `json:"total_invoices"`
	TotalAmount      float64 `json:"total_amount"`
	PaidAmount       float64 `json:"paid_amount"`
	OutstandingAmount float64 `json:"outstanding_amount"`
	DraftCount       int64   `json:"draft_count"`
	SentCount        int64   `json:"sent_count"`
	PaidCount        int64   `json:"paid_count"`
	OverdueCount     int64   `json:"overdue_count"`
}