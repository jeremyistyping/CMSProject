package models

import (
	"time"
	"gorm.io/gorm"
)

// Milestone represents a project milestone or deliverable
type Milestone struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ProjectID       uint           `json:"project_id" gorm:"not null;index"`
	Title           string         `json:"title" gorm:"not null;size:200"`
	Description     string         `json:"description" gorm:"type:text"`
	WorkArea        string         `json:"work_area" gorm:"size:100;index"` // Site Preparation, Foundation Work, etc.
	Priority        string         `json:"priority" gorm:"size:20;default:'medium';index"` // low, medium, high
	AssignedTeam    string         `json:"assigned_team" gorm:"size:200"`
	TargetDate      time.Time      `json:"target_date" gorm:"not null;index"`
	CompletionDate  *time.Time     `json:"completion_date" gorm:"index"`
	Status          string         `json:"status" gorm:"not null;size:20;default:'pending';index"` // pending, in-progress, completed, delayed
	Progress        float64        `json:"progress" gorm:"type:decimal(5,2);default:0"` // 0-100
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relation
	Project         *Project       `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for Milestone model
func (Milestone) TableName() string {
	return "milestones"
}

// Milestone Status Constants
const (
	MilestoneStatusPending    = "pending"
	MilestoneStatusInProgress = "in-progress"
	MilestoneStatusCompleted  = "completed"
	MilestoneStatusDelayed    = "delayed"
)

// Milestone Priority Constants
const (
	MilestonePriorityLow    = "low"
	MilestonePriorityMedium = "medium"
	MilestonePriorityHigh   = "high"
)

// BeforeCreate hook to set default values
func (m *Milestone) BeforeCreate(tx *gorm.DB) error {
	if m.Status == "" {
		m.Status = MilestoneStatusPending
	}
	if m.Priority == "" {
		m.Priority = MilestonePriorityMedium
	}
	if m.Progress == 0 {
		m.Progress = 0
	}
	return nil
}

// BeforeSave hook to auto-update status based on data
func (m *Milestone) BeforeSave(tx *gorm.DB) error {
	// Auto-complete if completion date is set
	if m.CompletionDate != nil && m.Status != MilestoneStatusCompleted {
		m.Status = MilestoneStatusCompleted
		m.Progress = 100
	}
	
	// Auto-set to delayed if past target date and not completed
	if m.Status != MilestoneStatusCompleted && time.Now().After(m.TargetDate) {
		m.Status = MilestoneStatusDelayed
	}
	
	// Set progress to 100 if completed
	if m.Status == MilestoneStatusCompleted && m.Progress != 100 {
		m.Progress = 100
	}
	
	return nil
}

// IsOverdue checks if milestone is past target date and not completed
func (m *Milestone) IsOverdue() bool {
	return time.Now().After(m.TargetDate) && m.Status != MilestoneStatusCompleted
}

// DaysUntilTarget calculates days remaining until target date
func (m *Milestone) DaysUntilTarget() int {
	if m.Status == MilestoneStatusCompleted {
		return 0
	}
	days := int(time.Until(m.TargetDate).Hours() / 24)
	return days
}

// DaysDelayed calculates how many days the milestone is delayed
func (m *Milestone) DaysDelayed() int {
	if !m.IsOverdue() {
		return 0
	}
	days := int(time.Since(m.TargetDate).Hours() / 24)
	return days
}

// Complete marks the milestone as completed
func (m *Milestone) Complete() {
	now := time.Now()
	m.CompletionDate = &now
	m.Status = MilestoneStatusCompleted
	m.Progress = 100
}

// Validate checks if the milestone data is valid
func (m *Milestone) Validate() error {
	if m.ProjectID == 0 {
		return gorm.ErrRecordNotFound
	}
	if m.Title == "" {
		return gorm.ErrInvalidData
	}
	if m.TargetDate.IsZero() {
		return gorm.ErrInvalidData
	}
	return nil
}

// GetStatusColor returns color scheme for UI display
func (m *Milestone) GetStatusColor() string {
	switch m.Status {
	case MilestoneStatusPending:
		return "gray"
	case MilestoneStatusInProgress:
		return "blue"
	case MilestoneStatusCompleted:
		return "green"
	case MilestoneStatusDelayed:
		return "red"
	default:
		return "gray"
	}
}

