package models

import (
	"time"

	"gorm.io/gorm"
)

// ProjectProgress represents a time-series snapshot of physical progress for a project.
type ProjectProgress struct {
	ID                      uint           `json:"id" gorm:"primaryKey"`
	ProjectID               uint           `json:"project_id" gorm:"not null;index"`
	Date                    time.Time      `json:"date" gorm:"not null;index"`
	PhysicalProgressPercent float64        `json:"physical_progress_percent" gorm:"type:decimal(5,2);default:0"`
	VolumeAchieved          *float64       `json:"volume_achieved,omitempty" gorm:"type:decimal(20,4)"`
	Remarks                 string         `json:"remarks" gorm:"type:text"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// TableName specifies the table name for ProjectProgress model
func (ProjectProgress) TableName() string {
	return "project_progress"
}
