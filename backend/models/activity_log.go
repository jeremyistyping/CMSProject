package models

import (
	"time"
	"gorm.io/gorm"
)

// ActivityLog represents a log entry for user activities
type ActivityLog struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      *uint          `json:"user_id" gorm:"index;default:null"` // Nullable for anonymous users
	Username    string         `json:"username" gorm:"size:50;index"`
	Role        string         `json:"role" gorm:"size:20;index"`
	
	// Request Information
	Method      string         `json:"method" gorm:"size:10;not null"` // GET, POST, PUT, DELETE, etc.
	Path        string         `json:"path" gorm:"size:500;not null;index"`
	Action      string         `json:"action" gorm:"size:100"` // login, create_product, update_sale, etc.
	Resource    string         `json:"resource" gorm:"size:50;index"` // users, products, sales, etc.
	
	// Request Details
	RequestBody  string        `json:"request_body,omitempty" gorm:"type:text"`
	QueryParams  string        `json:"query_params,omitempty" gorm:"type:text"`
	
	// Response Information
	StatusCode   int           `json:"status_code" gorm:"index"`
	ResponseBody string        `json:"response_body,omitempty" gorm:"type:text"`
	
	// Network Information
	IPAddress    string        `json:"ip_address" gorm:"size:45;index"` // IPv4 or IPv6
	UserAgent    string        `json:"user_agent" gorm:"type:text"`
	
	// Performance Metrics
	Duration     int64         `json:"duration" gorm:"comment:Request duration in milliseconds"`
	
	// Additional Context
	Description  string        `json:"description" gorm:"type:text"`
	Metadata     string        `json:"metadata,omitempty" gorm:"type:jsonb"` // Additional JSON data
	
	// Error Tracking
	IsError      bool          `json:"is_error" gorm:"default:false;index"`
	ErrorMessage string        `json:"error_message,omitempty" gorm:"type:text"`
	
	// Audit Trail
	CreatedAt    time.Time     `json:"created_at" gorm:"index"`
	
	// Soft delete
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for ActivityLog
func (ActivityLog) TableName() string {
	return "activity_logs"
}

// ActivityLogSummary represents a summary of activity logs for reporting
type ActivityLogSummary struct {
	Date         string `json:"date"`
	UserID       *uint  `json:"user_id"` // Nullable for anonymous users
	Username     string `json:"username"`
	Role         string `json:"role"`
	TotalActions int64  `json:"total_actions"`
	SuccessCount int64  `json:"success_count"`
	ErrorCount   int64  `json:"error_count"`
}

// ActivityLogFilter represents filter criteria for querying activity logs
type ActivityLogFilter struct {
	UserID     *uint      `json:"user_id"`
	Username   string     `json:"username"`
	Role       string     `json:"role"`
	Method     string     `json:"method"`
	Path       string     `json:"path"`
	Resource   string     `json:"resource"`
	StatusCode *int       `json:"status_code"`
	IsError    *bool      `json:"is_error"`
	IPAddress  string     `json:"ip_address"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}
