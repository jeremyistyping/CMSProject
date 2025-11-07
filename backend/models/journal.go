package models

import (
	"time"
	"gorm.io/gorm"
)

type Journal struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Code         string         `json:"code" gorm:"unique;not null;size:30"`
	Date         time.Time      `json:"date"`
	Description  string         `json:"description" gorm:"not null;type:text"`
	ReferenceType string        `json:"reference_type" gorm:"size:50"` // MANUAL, SALE, PURCHASE, PAYMENT, etc.
	ReferenceID   *uint         `json:"reference_id" gorm:"index"`
	UserID       uint           `json:"user_id" gorm:"not null;index"`
	Status       string         `json:"status" gorm:"not null;size:20;default:'PENDING'"` // PENDING, POSTED, CANCELLED
	TotalDebit   float64        `json:"total_debit" gorm:"type:decimal(20,2);default:0"`
	TotalCredit  float64        `json:"total_credit" gorm:"type:decimal(20,2);default:0"`
	IsAdjusting  bool           `json:"is_adjusting" gorm:"default:false"`
	Period       string         `json:"period" gorm:"size:7"` // YYYY-MM format
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User         User            `json:"user" gorm:"foreignKey:UserID"`
	// JournalEntries relation removed - see journal_entry.go for new structure
}

// JournalEntry moved to journal_entry.go to avoid conflicts

// Journal Status Constants (different from JournalEntry status)
const (
	JournalStatusPending   = "PENDING"
	JournalStatusCancelled = "CANCELLED"
	// Note: JournalStatusPosted moved to journal_entry.go to avoid conflicts
)

// Journal Reference Types Constants
const (
	JournalRefTypeManual   = "MANUAL"
	JournalRefTypeSale     = "SALE"
	JournalRefTypePurchase = "PURCHASE"
	JournalRefTypePayment  = "PAYMENT"
	JournalRefTypeExpense  = "EXPENSE"
	JournalRefTypeAsset    = "ASSET"
	JournalRefTypeAdjustment = "ADJUSTMENT"
)
