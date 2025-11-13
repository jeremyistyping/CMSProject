package models

import (
	"time"
	"gorm.io/gorm"
)

// TimelineSchedule represents a project timeline schedule item
type TimelineSchedule struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	ProjectID     uint           `json:"project_id" gorm:"not null;index"`
	WorkArea      string         `json:"work_area" gorm:"not null;size:200;index"` // e.g., Site Preparation, Foundation Work, etc.
	AssignedTeam  string         `json:"assigned_team" gorm:"size:200"`             // Team or contractor name
	StartDate     time.Time      `json:"start_date" gorm:"not null;index"`
	EndDate       time.Time      `json:"end_date" gorm:"not null;index"`
	StartTime     string         `json:"start_time" gorm:"size:10;default:'08:00'"` // HH:MM format
	EndTime       string         `json:"end_time" gorm:"size:10;default:'17:00'"`   // HH:MM format
	Notes         string         `json:"notes" gorm:"type:text"`
	Status        string         `json:"status" gorm:"not null;size:20;default:'not-started';index"` // not-started, in-progress, completed
	Duration      int            `json:"duration" gorm:"-"`                         // Calculated field (days)
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relation
	Project       *Project       `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for TimelineSchedule model
func (TimelineSchedule) TableName() string {
	return "timeline_schedules"
}

// Timeline Schedule Status Constants
const (
	TimelineStatusNotStarted = "not-started"
	TimelineStatusInProgress = "in-progress"
	TimelineStatusCompleted  = "completed"
)

// Common Work Areas
const (
	WorkAreaSitePreparation     = "Site Preparation"
	WorkAreaFoundationWork      = "Foundation Work"
	WorkAreaStructuralWork      = "Structural Work"
	WorkAreaRoofingWork         = "Roofing Work"
	WorkAreaMechanicalWork      = "Mechanical Work"
	WorkAreaElectricalWork      = "Electrical Work"
	WorkAreaPlumbingWork        = "Plumbing Work"
	WorkAreaInteriorWork        = "Interior Work"
	WorkAreaExteriorFinishes    = "Exterior Finishes"
	WorkAreaLandscaping         = "Landscaping"
	WorkAreaFinalInspection     = "Final Inspection"
)

// BeforeCreate hook to set default values
func (ts *TimelineSchedule) BeforeCreate(tx *gorm.DB) error {
	if ts.Status == "" {
		ts.Status = TimelineStatusNotStarted
	}
	if ts.StartTime == "" {
		ts.StartTime = "08:00"
	}
	if ts.EndTime == "" {
		ts.EndTime = "17:00"
	}
	return nil
}

// AfterFind hook to calculate duration
func (ts *TimelineSchedule) AfterFind(tx *gorm.DB) error {
	ts.Duration = ts.CalculateDuration()
	return nil
}

// CalculateDuration calculates the duration in days between start and end date
func (ts *TimelineSchedule) CalculateDuration() int {
	if ts.EndDate.Before(ts.StartDate) {
		return 0
	}
	duration := int(ts.EndDate.Sub(ts.StartDate).Hours()/24) + 1 // +1 to include both start and end date
	return duration
}

// IsActive checks if the schedule is currently active (today is between start and end date)
func (ts *TimelineSchedule) IsActive() bool {
	now := time.Now()
	return (now.After(ts.StartDate) || now.Equal(ts.StartDate)) && now.Before(ts.EndDate)
}

// IsOverdue checks if the schedule is past end date and not completed
func (ts *TimelineSchedule) IsOverdue() bool {
	return time.Now().After(ts.EndDate) && ts.Status != TimelineStatusCompleted
}

// DaysRemaining calculates days remaining until end date
func (ts *TimelineSchedule) DaysRemaining() int {
	if ts.Status == TimelineStatusCompleted {
		return 0
	}
	days := int(time.Until(ts.EndDate).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// Validate checks if the timeline schedule data is valid
func (ts *TimelineSchedule) Validate() error {
	if ts.ProjectID == 0 {
		return gorm.ErrRecordNotFound
	}
	if ts.WorkArea == "" {
		return gorm.ErrInvalidData
	}
	if ts.StartDate.IsZero() || ts.EndDate.IsZero() {
		return gorm.ErrInvalidData
	}
	if ts.EndDate.Before(ts.StartDate) {
		return gorm.ErrInvalidData
	}
	return nil
}

// GetStatusColor returns color scheme for UI display
func (ts *TimelineSchedule) GetStatusColor() string {
	switch ts.Status {
	case TimelineStatusNotStarted:
		return "gray"
	case TimelineStatusInProgress:
		return "yellow"
	case TimelineStatusCompleted:
		return "green"
	default:
		return "gray"
	}
}

// DTO for Timeline Schedule with additional computed fields
type TimelineScheduleResponse struct {
	ID            uint      `json:"id"`
	ProjectID     uint      `json:"project_id"`
	WorkArea      string    `json:"work_area"`
	AssignedTeam  string    `json:"assigned_team"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	Notes         string    `json:"notes"`
	Status        string    `json:"status"`
	Duration      int       `json:"duration"`
	DaysRemaining int       `json:"days_remaining"`
	IsActive      bool      `json:"is_active"`
	IsOverdue     bool      `json:"is_overdue"`
	StatusColor   string    `json:"status_color"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToResponse converts TimelineSchedule to TimelineScheduleResponse
func (ts *TimelineSchedule) ToResponse() *TimelineScheduleResponse {
	return &TimelineScheduleResponse{
		ID:            ts.ID,
		ProjectID:     ts.ProjectID,
		WorkArea:      ts.WorkArea,
		AssignedTeam:  ts.AssignedTeam,
		StartDate:     ts.StartDate,
		EndDate:       ts.EndDate,
		StartTime:     ts.StartTime,
		EndTime:       ts.EndTime,
		Notes:         ts.Notes,
		Status:        ts.Status,
		Duration:      ts.CalculateDuration(),
		DaysRemaining: ts.DaysRemaining(),
		IsActive:      ts.IsActive(),
		IsOverdue:     ts.IsOverdue(),
		StatusColor:   ts.GetStatusColor(),
		CreatedAt:     ts.CreatedAt,
		UpdatedAt:     ts.UpdatedAt,
	}
}

