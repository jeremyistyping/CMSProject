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
	Budget              float64        `json:"budget" gorm:"type:decimal(20,2);default:0"`
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

