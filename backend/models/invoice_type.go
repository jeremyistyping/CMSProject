package models

import (
	"time"
	"gorm.io/gorm"
)

// InvoiceType represents different types of invoices with unique numbering sequences
type InvoiceType struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:100"`          // e.g., "Corporate Sales", "Retail Sales"
	Code        string         `json:"code" gorm:"unique;not null;size:20"`    // e.g., "STA-C", "STA-B"
	Description string         `json:"description" gorm:"type:text"`           // Optional description
	IsActive    bool           `json:"is_active" gorm:"default:true"`          // Active status
	CreatedBy   uint           `json:"created_by" gorm:"index"`                // User who created this type
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Creator User `json:"creator" gorm:"foreignKey:CreatedBy"`
}

// InvoiceCounter tracks the invoice numbering sequence per type per year
type InvoiceCounter struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	InvoiceTypeID uint      `json:"invoice_type_id" gorm:"not null;index"`     // Foreign key to InvoiceType
	Year          int       `json:"year" gorm:"not null;index"`                // Year for the counter (e.g., 2025)
	Counter       int       `json:"counter" gorm:"not null;default:0"`         // Current counter value
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relations
	InvoiceType InvoiceType `json:"invoice_type" gorm:"foreignKey:InvoiceTypeID"`
}

// Ensure unique constraint on InvoiceTypeID + Year
func (InvoiceCounter) TableName() string {
	return "invoice_counters"
}

// Request/Response DTOs for InvoiceType
type InvoiceTypeCreateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Code        string `json:"code" binding:"required,min=1,max=20"`
	Description string `json:"description" binding:"max=500"`
}

type InvoiceTypeUpdateRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Code        *string `json:"code" binding:"omitempty,min=1,max=20"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active"`
}

type InvoiceTypeResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Creator     User      `json:"creator,omitempty"`
}

// Invoice numbering helper
type InvoiceNumberRequest struct {
	InvoiceTypeID uint      `json:"invoice_type_id" binding:"required"`
	Date          time.Time `json:"date" binding:"required"`
}

type InvoiceNumberResponse struct {
	InvoiceNumber string `json:"invoice_number"`
	Counter       int    `json:"counter"`
	Year          int    `json:"year"`
	Month         string `json:"month_roman"`
	TypeCode      string `json:"type_code"`
}

// Roman numeral conversion helper (for months)
var romanNumerals = map[int]string{
	1: "I", 2: "II", 3: "III", 4: "IV", 5: "V", 6: "VI",
	7: "VII", 8: "VIII", 9: "IX", 10: "X", 11: "XI", 12: "XII",
}

// GetRomanMonth returns roman numeral for given month (1-12)
func GetRomanMonth(month int) string {
	if roman, exists := romanNumerals[month]; exists {
		return roman
	}
	return "I" // fallback
}