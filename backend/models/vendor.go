package models

import (
	"time"

	"gorm.io/gorm"
)

// Vendor represents a supplier/vendor master data
type Vendor struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Code          string         `json:"code" gorm:"unique;not null;size:20"`
	Name          string         `json:"name" gorm:"not null;size:200"`
	ContactPerson string         `json:"contact_person" gorm:"size:100"`
	Email         string         `json:"email" gorm:"size:100"`
	Phone         string         `json:"phone" gorm:"size:20"`
	Address       string         `json:"address" gorm:"type:text"`
	City          string         `json:"city" gorm:"size:100"`
	Province      string         `json:"province" gorm:"size:100"`
	PostalCode    string         `json:"postal_code" gorm:"size:10"`
	NPWP          string         `json:"npwp" gorm:"size:30"`           // Tax ID
	BankName      string         `json:"bank_name" gorm:"size:100"`
	BankAccount   string         `json:"bank_account" gorm:"size:50"`
	BankBranch    string         `json:"bank_branch" gorm:"size:100"`
	PaymentTerms  int            `json:"payment_terms" gorm:"default:30"` // Days
	CategoryID    *uint          `json:"category_id" gorm:"index"`
	Rating        float64        `json:"rating" gorm:"type:decimal(3,2);default:0"` // 0-5 rating
	Notes         string         `json:"notes" gorm:"type:text"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	CreatedBy     uint           `json:"created_by" gorm:"index"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Category *VendorCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Creator  *User           `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// TableName specifies the table name for Vendor
func (Vendor) TableName() string {
	return "vendors"
}

// VendorCategory represents a category for vendors
type VendorCategory struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"unique;not null;size:20"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Description string         `json:"description" gorm:"type:text"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for VendorCategory
func (VendorCategory) TableName() string {
	return "vendor_categories"
}

// VendorMaterial represents the relationship between vendor and materials they supply
type VendorMaterial struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	VendorID     uint           `json:"vendor_id" gorm:"not null;index"`
	MaterialID   uint           `json:"material_id" gorm:"not null;index"`
	UnitPrice    float64        `json:"unit_price" gorm:"type:decimal(15,2);default:0"`
	LeadTimeDays int            `json:"lead_time_days" gorm:"default:7"`
	MinOrderQty  float64        `json:"min_order_qty" gorm:"type:decimal(15,2);default:1"`
	IsPreferred  bool           `json:"is_preferred" gorm:"default:false"`
	Notes        string         `json:"notes" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Vendor   *Vendor   `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	Material *Material `json:"material,omitempty" gorm:"foreignKey:MaterialID"`
}

// TableName specifies the table name for VendorMaterial
func (VendorMaterial) TableName() string {
	return "vendor_materials"
}

// CreateVendorDTO represents the data transfer object for creating a vendor
type CreateVendorDTO struct {
	Code          string `json:"code" binding:"required"`
	Name          string `json:"name" binding:"required"`
	ContactPerson string `json:"contact_person"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	City          string `json:"city"`
	Province      string `json:"province"`
	PostalCode    string `json:"postal_code"`
	NPWP          string `json:"npwp"`
	BankName      string `json:"bank_name"`
	BankAccount   string `json:"bank_account"`
	BankBranch    string `json:"bank_branch"`
	PaymentTerms  int    `json:"payment_terms"`
	CategoryID    *uint  `json:"category_id"`
	Notes         string `json:"notes"`
}

// UpdateVendorDTO represents the data transfer object for updating a vendor
type UpdateVendorDTO struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	ContactPerson string  `json:"contact_person"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	Address       string  `json:"address"`
	City          string  `json:"city"`
	Province      string  `json:"province"`
	PostalCode    string  `json:"postal_code"`
	NPWP          string  `json:"npwp"`
	BankName      string  `json:"bank_name"`
	BankAccount   string  `json:"bank_account"`
	BankBranch    string  `json:"bank_branch"`
	PaymentTerms  int     `json:"payment_terms"`
	CategoryID    *uint   `json:"category_id"`
	Rating        float64 `json:"rating"`
	Notes         string  `json:"notes"`
	IsActive      *bool   `json:"is_active"`
}

// CreateVendorCategoryDTO represents the data transfer object for creating a vendor category
type CreateVendorCategoryDTO struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// VendorSummary represents a summary of vendors for reporting
type VendorSummary struct {
	TotalVendors    int64   `json:"total_vendors"`
	ActiveVendors   int64   `json:"active_vendors"`
	TotalCategories int64   `json:"total_categories"`
	AverageRating   float64 `json:"average_rating"`
}

// VendorPerformance represents vendor performance metrics
type VendorPerformance struct {
	VendorID         uint    `json:"vendor_id"`
	VendorName       string  `json:"vendor_name"`
	TotalOrders      int64   `json:"total_orders"`
	TotalAmount      float64 `json:"total_amount"`
	OnTimeDelivery   float64 `json:"on_time_delivery"` // Percentage
	QualityRating    float64 `json:"quality_rating"`
	AverageLeadTime  float64 `json:"average_lead_time"`
}
