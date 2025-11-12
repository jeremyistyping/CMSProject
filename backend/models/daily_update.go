package models

import (
	"time"
	"gorm.io/gorm"
	"github.com/lib/pq"
)

// DailyUpdate represents a daily construction project update
type DailyUpdate struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	ProjectID        uint            `json:"project_id" gorm:"not null;index"`
	Date             time.Time       `json:"date" gorm:"not null;index"`
	Weather          string          `json:"weather" gorm:"not null;size:50"`
	WorkersPresent   int             `json:"workers_present" gorm:"not null;default:0"`
	WorkDescription  string          `json:"work_description" gorm:"type:text;not null"`
	MaterialsUsed    string          `json:"materials_used" gorm:"type:text"`
	Issues           string          `json:"issues" gorm:"type:text"`
	TomorrowsPlan    string          `json:"tomorrows_plan" gorm:"type:text"`
	Photos           pq.StringArray  `json:"photos" gorm:"type:text[]"` // PostgreSQL array of photo URLs/paths
	CreatedBy        string          `json:"created_by" gorm:"not null;size:100"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `json:"-" gorm:"index"`
	
	// Relation
	Project          *Project       `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for DailyUpdate model
func (DailyUpdate) TableName() string {
	return "daily_updates"
}

// Weather Constants
const (
	WeatherSunny        = "Sunny"
	WeatherCloudy       = "Cloudy"
	WeatherRainy        = "Rainy"
	WeatherStormy       = "Stormy"
	WeatherPartlyCloudy = "Partly Cloudy"
)

// BeforeCreate hook to set default values
func (du *DailyUpdate) BeforeCreate(tx *gorm.DB) error {
	if du.Weather == "" {
		du.Weather = WeatherSunny
	}
	if du.CreatedBy == "" {
		du.CreatedBy = "System"
	}
	return nil
}

// Validate checks if the daily update data is valid
func (du *DailyUpdate) Validate() error {
	if du.ProjectID == 0 {
		return gorm.ErrRecordNotFound
	}
	if du.WorkDescription == "" {
		return gorm.ErrInvalidData
	}
	return nil
}

