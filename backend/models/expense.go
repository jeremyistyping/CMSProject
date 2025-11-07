package models

import (
	"time"
	"gorm.io/gorm"
)

type Expense struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Code          string         `json:"code" gorm:"unique;not null;size:20"`
	Name          string         `json:"name" gorm:"not null;size:100"`
	Amount        float64        `json:"amount" gorm:"type:decimal(15,2);default:0"`
	Date          time.Time      `json:"date"`
	CategoryID    *uint          `json:"category_id" gorm:"index"`
	AccountID     uint           `json:"account_id" gorm:"not null;index"`
	ContactID     *uint          `json:"contact_id" gorm:"index"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`
	Notes         string         `json:"notes" gorm:"type:text"`
	ReceiptNumber string         `json:"receipt_number" gorm:"size:50"`
	IsRecurring   bool           `json:"is_recurring" gorm:"default:false"`
	Status        string         `json:"status" gorm:"size:20"` // PENDING, PAID
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Category *ExpenseCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Account  Account          `json:"account" gorm:"foreignKey:AccountID"`
	Contact  *Contact         `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	User     User             `json:"user" gorm:"foreignKey:UserID"`
}

type ExpenseCategory struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"unique;not null;size:20"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Description string         `json:"description" gorm:"type:text"`
	ParentID    *uint          `json:"parent_id" gorm:"index"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Parent   *ExpenseCategory `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []ExpenseCategory `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Expenses []Expense        `json:"-" gorm:"foreignKey:CategoryID"`
}

// Expense Status Constants
const (
	ExpenseStatusPending = "PENDING"
	ExpenseStatusPaid    = "PAID"
)
