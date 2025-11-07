package models

import (
	"time"
	"gorm.io/gorm"
)

type Contact struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Code         string         `json:"code" gorm:"unique;not null;size:20"`
	Name         string         `json:"name" gorm:"not null;size:100"`
	Type         string         `json:"type" gorm:"not null;size:20"` // CUSTOMER, VENDOR, EMPLOYEE
	Category     string         `json:"category" gorm:"size:50"`      // RETAIL, WHOLESALE, etc.
	Email        string         `json:"email" gorm:"size:100"`
	Phone        string         `json:"phone" gorm:"size:20"`
	Mobile       string         `json:"mobile" gorm:"size:20"`
	Fax          string         `json:"fax" gorm:"size:20"`
	Website      string         `json:"website" gorm:"size:100"`
	TaxNumber    string         `json:"tax_number" gorm:"size:50"`
	CreditLimit  float64        `json:"credit_limit" gorm:"type:decimal(15,2);default:0"`
	PaymentTerms int            `json:"payment_terms" gorm:"default:30"` // Days
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	
	// Additional fields
	PICName      string         `json:"pic_name" gorm:"size:100"`        // Person In Charge (for Customer/Vendor)
	ExternalID   string         `json:"external_id" gorm:"size:50"`      // Employee ID, Vendor ID, Customer ID
	Address      string         `json:"address" gorm:"type:text"`        // Simple address field
	
	// Default Expense Account for purchases from this vendor
	DefaultExpenseAccountID *uint `json:"default_expense_account_id" gorm:"index"`
	
	Notes        string         `json:"notes" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Addresses []ContactAddress `json:"addresses,omitempty" gorm:"foreignKey:ContactID"`
	Sales     []Sale           `json:"-" gorm:"foreignKey:CustomerID"`
	Purchases []Purchase       `json:"-" gorm:"foreignKey:VendorID"`
	Payments  []Payment        `json:"-" gorm:"foreignKey:ContactID"`
}

type ContactAddress struct {
	ID         uint             `json:"id" gorm:"primaryKey"`
	ContactID  uint             `json:"contact_id" gorm:"not null;index"`
	Type       string           `json:"type" gorm:"not null;size:20"` // BILLING, SHIPPING, MAILING
	Address1   string           `json:"address1" gorm:"not null;size:255"`
	Address2   string           `json:"address2" gorm:"size:255"`
	City       string           `json:"city" gorm:"not null;size:100"`
	State      string           `json:"state" gorm:"size:100"`
	PostalCode string           `json:"postal_code" gorm:"size:20"`
	Country    string           `json:"country" gorm:"not null;size:100;default:'Indonesia'"`
	IsDefault  bool             `json:"is_default" gorm:"default:false"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	DeletedAt  gorm.DeletedAt   `json:"-" gorm:"index"`

	// Relations
	Contact Contact `json:"-" gorm:"foreignKey:ContactID"`
}

// Contact Types Constants
const (
	ContactTypeCustomer = "CUSTOMER"
	ContactTypeVendor   = "VENDOR"
	ContactTypeEmployee = "EMPLOYEE"
)

// Address Types Constants
const (
	AddressTypeBilling  = "BILLING"
	AddressTypeShipping = "SHIPPING"
	AddressTypeMailing  = "MAILING"
)

// Contact Categories Constants
const (
	CategoryRetail        = "RETAIL"
	CategoryWholesale     = "WHOLESALE"
	CategoryDistributor   = "DISTRIBUTOR"
	CategoryManufacturer  = "MANUFACTURER"
	CategoryServiceProvider = "SERVICE_PROVIDER"
)

// ContactHistory tracks changes made to contacts
type ContactHistory struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ContactID uint           `json:"contact_id" gorm:"not null;index"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	Action    string         `json:"action" gorm:"not null;size:50"` // CREATED, UPDATED, DELETED
	Field     string         `json:"field" gorm:"size:100"`          // Field that was changed
	OldValue  string         `json:"old_value" gorm:"type:text"`
	NewValue  string         `json:"new_value" gorm:"type:text"`
	Notes     string         `json:"notes" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Contact Contact `json:"-" gorm:"foreignKey:ContactID"`
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// CommunicationLog tracks communication with contacts
type CommunicationLog struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ContactID   uint           `json:"contact_id" gorm:"not null;index"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Type        string         `json:"type" gorm:"not null;size:50"` // EMAIL, PHONE, MEETING, etc.
	Subject     string         `json:"subject" gorm:"size:255"`
	Content     string         `json:"content" gorm:"type:text"`
	Direction   string         `json:"direction" gorm:"not null;size:20"` // INBOUND, OUTBOUND
	Status      string         `json:"status" gorm:"default:'COMPLETED';size:20"` // COMPLETED, SCHEDULED, CANCELLED
	ScheduledAt *time.Time     `json:"scheduled_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Contact Contact `json:"-" gorm:"foreignKey:ContactID"`
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Communication Types Constants
const (
	CommunicationTypeEmail   = "EMAIL"
	CommunicationTypePhone   = "PHONE"
	CommunicationTypeMeeting = "MEETING"
	CommunicationTypeSMS     = "SMS"
	CommunicationTypeChat    = "CHAT"
)

// Communication Direction Constants
const (
	DirectionInbound  = "INBOUND"
	DirectionOutbound = "OUTBOUND"
)

// Communication Status Constants
const (
	StatusCompleted = "COMPLETED"
	StatusScheduled = "SCHEDULED"
	StatusCancelled = "CANCELLED"
)
