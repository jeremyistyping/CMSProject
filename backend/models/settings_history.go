package models

import (
	"time"
	"gorm.io/gorm"
)

// SettingsHistory tracks changes to system settings for audit purposes
type SettingsHistory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Reference to settings
	SettingsID uint     `json:"settings_id" gorm:"not null"`
	Settings   Settings `json:"settings,omitempty" gorm:"foreignKey:SettingsID"`

	// Change tracking
	Field     string `json:"field" gorm:"not null"`       // Field that was changed
	OldValue  string `json:"old_value"`                   // Previous value (JSON string)
	NewValue  string `json:"new_value"`                   // New value (JSON string)
	Action    string `json:"action" gorm:"default:'UPDATE'"` // UPDATE, RESET, CREATE

	// User tracking
	ChangedBy uint `json:"changed_by" gorm:"not null"`
	User      User `json:"user,omitempty" gorm:"foreignKey:ChangedBy"`

	// Additional context
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Reason    string `json:"reason,omitempty"` // Optional reason for change
}

// TableName specifies the table name for SettingsHistory
func (SettingsHistory) TableName() string {
	return "settings_history"
}

// SettingsHistoryResponse for API responses
type SettingsHistoryResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	Action    string    `json:"action"`
	ChangedBy uint      `json:"changed_by"`
	UserName  string    `json:"user_name,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	Reason    string    `json:"reason,omitempty"`
}

// ToResponse converts SettingsHistory to SettingsHistoryResponse
func (h *SettingsHistory) ToResponse() SettingsHistoryResponse {
	response := SettingsHistoryResponse{
		ID:        h.ID,
		CreatedAt: h.CreatedAt,
		Field:     h.Field,
		OldValue:  h.OldValue,
		NewValue:  h.NewValue,
		Action:    h.Action,
		ChangedBy: h.ChangedBy,
		IPAddress: h.IPAddress,
		Reason:    h.Reason,
	}

	// Add user name if user is loaded
	if h.User.ID != 0 {
		// Combine first name and last name
		userName := ""
		if h.User.FirstName != "" || h.User.LastName != "" {
			userName = h.User.FirstName + " " + h.User.LastName
		} else if h.User.Username != "" {
			userName = h.User.Username
		}
		response.UserName = userName
	}

	return response
}

// SettingsHistoryFilter represents filter parameters for settings history queries
type SettingsHistoryFilter struct {
	Field     string `json:"field"`
	Action    string `json:"action"`
	ChangedBy string `json:"changed_by"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
}