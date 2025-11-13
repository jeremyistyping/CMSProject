package models

import (
	"time"
	"gorm.io/gorm"
)

// WeeklyReport represents a weekly construction project report
type WeeklyReport struct {
	ID                  uint           `json:"id" gorm:"primaryKey"`
	ProjectID           uint           `json:"project_id" gorm:"not null;index"`
	Week                int            `json:"week" gorm:"not null"` // Week number (1-52)
	Year                int            `json:"year" gorm:"not null;index"`
	ProjectManager      string         `json:"project_manager" gorm:"size:200"`
	TotalWorkDays       int            `json:"total_work_days" gorm:"default:0"`
	WeatherDelays       int            `json:"weather_delays" gorm:"default:0"`
	TeamSize            int            `json:"team_size" gorm:"default:0"`
	Accomplishments     string         `json:"accomplishments" gorm:"type:text"`
	Challenges          string         `json:"challenges" gorm:"type:text"`
	NextWeekPriorities  string         `json:"next_week_priorities" gorm:"type:text"`
	GeneratedDate       time.Time      `json:"generated_date" gorm:"not null"`
	CreatedBy           string         `json:"created_by" gorm:"size:100"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relation
	Project             *Project       `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for WeeklyReport model
func (WeeklyReport) TableName() string {
	return "weekly_reports"
}

// BeforeCreate hook to set default values
func (wr *WeeklyReport) BeforeCreate(tx *gorm.DB) error {
	if wr.GeneratedDate.IsZero() {
		wr.GeneratedDate = time.Now()
	}
	if wr.CreatedBy == "" {
		wr.CreatedBy = "System"
	}
	return nil
}

// Validate checks if the weekly report data is valid
func (wr *WeeklyReport) Validate() error {
	if wr.ProjectID == 0 {
		return gorm.ErrRecordNotFound
	}
	if wr.Week < 1 || wr.Week > 53 {
		return gorm.ErrInvalidData
	}
	if wr.Year < 2000 {
		return gorm.ErrInvalidData
	}
	return nil
}

// GetWeekDateRange returns the start and end date for the report week
func (wr *WeeklyReport) GetWeekDateRange() (time.Time, time.Time) {
	// Calculate start date (Monday) of the week
	firstDayOfYear := time.Date(wr.Year, 1, 1, 0, 0, 0, 0, time.Local)
	daysToAdd := (wr.Week - 1) * 7
	weekStart := firstDayOfYear.AddDate(0, 0, daysToAdd)
	
	// Adjust to Monday if not already
	for weekStart.Weekday() != time.Monday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}
	
	weekEnd := weekStart.AddDate(0, 0, 6) // Sunday
	return weekStart, weekEnd
}

// WeeklyReportDTO for API responses
type WeeklyReportDTO struct {
	ID                 uint      `json:"id"`
	ProjectID          uint      `json:"project_id"`
	ProjectName        string    `json:"project_name"`
	ProjectCode        string    `json:"project_code,omitempty"`
	Week               int       `json:"week"`
	Year               int       `json:"year"`
	WeekLabel          string    `json:"week_label"` // e.g., "Week 45, 2025"
	ProjectManager     string    `json:"project_manager"`
	TotalWorkDays      int       `json:"total_work_days"`
	WeatherDelays      int       `json:"weather_delays"`
	TeamSize           int       `json:"team_size"`
	Accomplishments    string    `json:"accomplishments"`
	Challenges         string    `json:"challenges"`
	NextWeekPriorities string    `json:"next_week_priorities"`
	GeneratedDate      time.Time `json:"generated_date"`
	CreatedBy          string    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// WeeklyReportCreateRequest for creating a new weekly report
type WeeklyReportCreateRequest struct {
	ProjectID          uint   `json:"project_id" binding:"required"`
	Week               int    `json:"week" binding:"required,min=1,max=53"`
	Year               int    `json:"year" binding:"required,min=2000"`
	ProjectManager     string `json:"project_manager"`
	TotalWorkDays      int    `json:"total_work_days"`
	WeatherDelays      int    `json:"weather_delays"`
	TeamSize           int    `json:"team_size"`
	Accomplishments    string `json:"accomplishments"`
	Challenges         string `json:"challenges"`
	NextWeekPriorities string `json:"next_week_priorities"`
	CreatedBy          string `json:"created_by"`
}

// WeeklyReportUpdateRequest for updating an existing weekly report
type WeeklyReportUpdateRequest struct {
	ProjectManager     *string `json:"project_manager"`
	TotalWorkDays      *int    `json:"total_work_days"`
	WeatherDelays      *int    `json:"weather_delays"`
	TeamSize           *int    `json:"team_size"`
	Accomplishments    *string `json:"accomplishments"`
	Challenges         *string `json:"challenges"`
	NextWeekPriorities *string `json:"next_week_priorities"`
}

