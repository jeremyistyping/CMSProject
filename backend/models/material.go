package models

import (
	"time"

	"gorm.io/gorm"
)

// Material represents a master material/product item
type Material struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Code         string         `json:"code" gorm:"unique;not null;size:50"`
	Name         string         `json:"name" gorm:"not null;size:200"`
	Description  string         `json:"description" gorm:"type:text"`
	CategoryID   *uint          `json:"category_id" gorm:"index"`
	Unit         string         `json:"unit" gorm:"not null;size:20"` // kg, m, m2, m3, unit, pcs, etc.
	UnitPrice    float64        `json:"unit_price" gorm:"type:decimal(15,2);default:0"`
	MinStock     float64        `json:"min_stock" gorm:"type:decimal(15,2);default:0"`
	MaxStock     float64        `json:"max_stock" gorm:"type:decimal(15,2);default:0"`
	CurrentStock float64        `json:"current_stock" gorm:"type:decimal(15,2);default:0"`
	COAAccountID *uint          `json:"coa_account_id" gorm:"index"` // Link to COA for accounting
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	CreatedBy    uint           `json:"created_by" gorm:"index"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Category   *MaterialCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	COAAccount *COAAccount       `json:"coa_account,omitempty" gorm:"foreignKey:COAAccountID"`
	Creator    *User             `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// TableName specifies the table name for Material
func (Material) TableName() string {
	return "materials"
}

// MaterialCategory represents a category for materials
type MaterialCategory struct {
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
	Parent   *MaterialCategory  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []MaterialCategory `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// TableName specifies the table name for MaterialCategory
func (MaterialCategory) TableName() string {
	return "material_categories"
}

// UnitOfMeasure represents a unit of measurement
type UnitOfMeasure struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Code         string         `json:"code" gorm:"unique;not null;size:10"`
	Name         string         `json:"name" gorm:"not null;size:50"`
	Symbol       string         `json:"symbol" gorm:"size:10"`
	Category     string         `json:"category" gorm:"size:30"` // Length, Area, Volume, Weight, Quantity
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for UnitOfMeasure
func (UnitOfMeasure) TableName() string {
	return "unit_of_measures"
}

// UoM Category Constants
const (
	UoMCategoryLength   = "Length"
	UoMCategoryArea     = "Area"
	UoMCategoryVolume   = "Volume"
	UoMCategoryWeight   = "Weight"
	UoMCategoryQuantity = "Quantity"
)

// CreateMaterialDTO represents the data transfer object for creating a material
type CreateMaterialDTO struct {
	Code         string  `json:"code" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	CategoryID   *uint   `json:"category_id"`
	Unit         string  `json:"unit" binding:"required"`
	UnitPrice    float64 `json:"unit_price"`
	MinStock     float64 `json:"min_stock"`
	MaxStock     float64 `json:"max_stock"`
	COAAccountID *uint   `json:"coa_account_id"`
}

// UpdateMaterialDTO represents the data transfer object for updating a material
type UpdateMaterialDTO struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	CategoryID   *uint   `json:"category_id"`
	Unit         string  `json:"unit"`
	UnitPrice    float64 `json:"unit_price"`
	MinStock     float64 `json:"min_stock"`
	MaxStock     float64 `json:"max_stock"`
	COAAccountID *uint   `json:"coa_account_id"`
	IsActive     *bool   `json:"is_active"`
}

// CreateMaterialCategoryDTO represents the data transfer object for creating a material category
type CreateMaterialCategoryDTO struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ParentID    *uint  `json:"parent_id"`
}

// MaterialSummary represents a summary of material for reporting
type MaterialSummary struct {
	TotalMaterials   int64   `json:"total_materials"`
	ActiveMaterials  int64   `json:"active_materials"`
	LowStockCount    int64   `json:"low_stock_count"`
	TotalStockValue  float64 `json:"total_stock_value"`
	TotalCategories  int64   `json:"total_categories"`
}
