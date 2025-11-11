package models

import (
	"time"
	"gorm.io/gorm"
)

// Project represents a construction project
type Project struct {
	ID                  uint           `json:"id" gorm:"primaryKey"`
	ProjectName         string         `json:"project_name" gorm:"not null;size:200"`
	ProjectDescription  string         `json:"project_description" gorm:"type:text"`
	Customer            string         `json:"customer" gorm:"not null;size:200"`
	City                string         `json:"city" gorm:"not null;size:100"`
	Address             string         `json:"address" gorm:"type:text"`
	ProjectType         string         `json:"project_type" gorm:"not null;size:50"` // New Build, Renovation, Expansion, Maintenance
	
	// Budget & Cost Tracking
	Budget              float64        `json:"budget" gorm:"type:decimal(20,2);default:0"`
	ActualCost          float64        `json:"actual_cost" gorm:"type:decimal(20,2);default:0"` // Total actual spending
	MaterialCost        float64        `json:"material_cost" gorm:"type:decimal(20,2);default:0"`
	LaborCost           float64        `json:"labor_cost" gorm:"type:decimal(20,2);default:0"`
	EquipmentCost       float64        `json:"equipment_cost" gorm:"type:decimal(20,2);default:0"`
	OverheadCost        float64        `json:"overhead_cost" gorm:"type:decimal(20,2);default:0"`
	Variance            float64        `json:"variance" gorm:"type:decimal(20,2);default:0"` // Budget - Actual
	VariancePercent     float64        `json:"variance_percent" gorm:"type:decimal(5,2);default:0"` // (Variance/Budget)*100
	
	Deadline            time.Time      `json:"deadline"`
	Status              string         `json:"status" gorm:"not null;size:20;default:'active'"` // active, completed, on_hold, archived
	
	// Progress tracking
	OverallProgress     float64        `json:"overall_progress" gorm:"type:decimal(5,2);default:0"` // 0-100
	FoundationProgress  float64        `json:"foundation_progress" gorm:"type:decimal(5,2);default:0"`
	UtilitiesProgress   float64        `json:"utilities_progress" gorm:"type:decimal(5,2);default:0"`
	InteriorProgress    float64        `json:"interior_progress" gorm:"type:decimal(5,2);default:0"`
	EquipmentProgress   float64        `json:"equipment_progress" gorm:"type:decimal(5,2);default:0"`
	
	// Additional fields for future features
	StartDate           *time.Time     `json:"start_date"`
	CompletionDate      *time.Time     `json:"completion_date"`
	Notes               string         `json:"notes" gorm:"type:text"`
	
	// Timestamps
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Purchases           []Purchase     `json:"purchases,omitempty" gorm:"foreignKey:ProjectID"`
}

// TableName specifies the table name for Project model
func (Project) TableName() string {
	return "projects"
}

// Project Status Constants
const (
	ProjectStatusActive    = "active"
	ProjectStatusCompleted = "completed"
	ProjectStatusOnHold    = "on_hold"
	ProjectStatusArchived  = "archived"
)

// Project Type Constants
const (
	ProjectTypeNewBuild    = "New Build"
	ProjectTypeRenovation  = "Renovation"
	ProjectTypeExpansion   = "Expansion"
	ProjectTypeMaintenance = "Maintenance"
)

// BeforeCreate hook to set default values
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.Status == "" {
		p.Status = ProjectStatusActive
	}
	return nil
}

// IsOverdue checks if project is past deadline
func (p *Project) IsOverdue() bool {
	return time.Now().After(p.Deadline) && p.Status != ProjectStatusCompleted
}

// DaysUntilDeadline calculates days remaining until deadline
func (p *Project) DaysUntilDeadline() int {
	if p.Status == ProjectStatusCompleted {
		return 0
	}
	days := int(time.Until(p.Deadline).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// UpdateCostTracking updates project cost tracking fields
func (p *Project) UpdateCostTracking(materialCost, laborCost, equipmentCost, overheadCost float64) {
	p.MaterialCost = materialCost
	p.LaborCost = laborCost
	p.EquipmentCost = equipmentCost
	p.OverheadCost = overheadCost
	p.ActualCost = materialCost + laborCost + equipmentCost + overheadCost
	p.Variance = p.Budget - p.ActualCost
	if p.Budget > 0 {
		p.VariancePercent = (p.Variance / p.Budget) * 100
	} else {
		p.VariancePercent = 0
	}
}

// IsOverBudget checks if project exceeded budget
func (p *Project) IsOverBudget() bool {
	return p.ActualCost > p.Budget
}

// GetBudgetUtilization returns budget utilization percentage
func (p *Project) GetBudgetUtilization() float64 {
	if p.Budget == 0 {
		return 0
	}
	return (p.ActualCost / p.Budget) * 100
}

// GetRemainingBudget returns remaining budget
func (p *Project) GetRemainingBudget() float64 {
	return p.Budget - p.ActualCost
}

// DTOs for Project Cost Tracking
type ProjectCostSummary struct {
	ProjectID          uint    `json:"project_id"`
	ProjectName        string  `json:"project_name"`
	Budget             float64 `json:"budget"`
	ActualCost         float64 `json:"actual_cost"`
	MaterialCost       float64 `json:"material_cost"`
	LaborCost          float64 `json:"labor_cost"`
	EquipmentCost      float64 `json:"equipment_cost"`
	OverheadCost       float64 `json:"overhead_cost"`
	Variance           float64 `json:"variance"`
	VariancePercent    float64 `json:"variance_percent"`
	BudgetUtilization  float64 `json:"budget_utilization"`
	RemainingBudget    float64 `json:"remaining_budget"`
	IsOverBudget       bool    `json:"is_over_budget"`
	TotalPurchases     int64   `json:"total_purchases"`
	OverallProgress    float64 `json:"overall_progress"`
	Status             string  `json:"status"`
}

type ProjectPurchaseSummary struct {
	ProjectID       uint    `json:"project_id"`
	ProjectName     string  `json:"project_name"`
	TotalPurchases  int64   `json:"total_purchases"`
	TotalAmount     float64 `json:"total_amount"`
	ApprovedAmount  float64 `json:"approved_amount"`
	PendingAmount   float64 `json:"pending_amount"`
}

