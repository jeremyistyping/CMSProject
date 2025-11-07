package models

import (
	"time"
	"gorm.io/gorm"
)

// PaymentCodeSequence represents the payment code sequence table
type PaymentCodeSequence struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Prefix         string         `json:"prefix" gorm:"size:10;not null"`
	Year           int            `json:"year" gorm:"not null"`
	Month          int            `json:"month" gorm:"not null"`
	SequenceNumber int            `json:"sequence_number" gorm:"not null;default:0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`

	// Composite unique constraint
	// This is handled by the unique index in the migration
}

// BeforeCreate hook can be used for additional validation
func (p *PaymentCodeSequence) BeforeCreate(tx *gorm.DB) error {
	if p.SequenceNumber == 0 {
		p.SequenceNumber = 1
	}
	return nil
}
