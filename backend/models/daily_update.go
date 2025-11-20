package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// DailyUpdate represents a daily construction project update
type DailyUpdate struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ProjectID       uint           `json:"project_id"`
	Date            time.Time      `json:"date"`
	Weather         string         `json:"weather"`
	WorkersPresent  int            `json:"workers_present"`
	WorkDescription string         `json:"work_description"`
	MaterialsUsed   string         `json:"materials_used"`
	Issues          string         `json:"issues"`
	TomorrowsPlan   string         `json:"tomorrows_plan"`
	Photos          pq.StringArray `json:"photos" gorm:"type:text[]"`
	CreatedBy       string         `json:"created_by"`

	// Progress fields
	Progress           float64 `json:"progress"`
	FoundationProgress float64 `json:"foundation_progress"`
	UtilitiesProgress  float64 `json:"utilities_progress"`
	InteriorProgress   float64 `json:"interior_progress"`
	EquipmentProgress  float64 `json:"equipment_progress"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relation
	Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
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
