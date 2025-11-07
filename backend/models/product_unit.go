package models

import (
	"time"
	"gorm.io/gorm"
)

// ProductUnit represents product unit of measurement
type ProductUnit struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"uniqueIndex;size:20;not null"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Symbol      string    `json:"symbol" gorm:"size:10"`
	Type        string    `json:"type" gorm:"size:50"` // e.g., "Weight", "Volume", "Count", "Length"
	Description string    `json:"description" gorm:"size:255"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for ProductUnit model
func (ProductUnit) TableName() string {
	return "product_units"
}

// Common ProductUnit types
const (
	UnitTypeWeight = "Weight"
	UnitTypeVolume = "Volume"
	UnitTypeCount  = "Count"
	UnitTypeLength = "Length"
	UnitTypeArea   = "Area"
	UnitTypeTime   = "Time"
)
