package models

import (
	"time"
	"gorm.io/gorm"
)

// SecurityIncident represents a security incident in the system
type SecurityIncident struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	IncidentType  string         `json:"incident_type" gorm:"not null;size:50;index"` // IP_WHITELIST_VIOLATION, SUSPICIOUS_REQUEST, etc.
	Severity      string         `json:"severity" gorm:"not null;size:20;default:'medium'"` // low, medium, high, critical
	Description   string         `json:"description" gorm:"type:text"`
	ClientIP      string         `json:"client_ip" gorm:"size:45;index"`
	UserAgent     string         `json:"user_agent" gorm:"type:text"`
	RequestMethod string         `json:"request_method" gorm:"size:10"`
	RequestPath   string         `json:"request_path" gorm:"size:500"`
	RequestHeaders string        `json:"request_headers" gorm:"type:text"`
	UserID        *uint          `json:"user_id" gorm:"index"`
	SessionID     string         `json:"session_id" gorm:"size:255;index"`
	Resolved      bool           `json:"resolved" gorm:"default:false"`
	ResolvedAt    *time.Time     `json:"resolved_at"`
	ResolvedBy    *uint          `json:"resolved_by"`
	Notes         string         `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	User       User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ResolvedByUser User `json:"resolved_by_user,omitempty" gorm:"foreignKey:ResolvedBy"`
}

// SystemAlert represents system-wide security alerts
type SystemAlert struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	AlertType   string         `json:"alert_type" gorm:"not null;size:50;index"` // RATE_LIMIT, SUSPICIOUS_ACTIVITY, etc.
	Level       string         `json:"level" gorm:"not null;size:20"` // info, warning, error, critical
	Title       string         `json:"title" gorm:"not null;size:200"`
	Message     string         `json:"message" gorm:"type:text"`
	Count       int            `json:"count" gorm:"default:1"`
	FirstSeen   time.Time      `json:"first_seen"`
	LastSeen    time.Time      `json:"last_seen"`
	Acknowledged bool          `json:"acknowledged" gorm:"default:false"`
	AcknowledgedAt *time.Time  `json:"acknowledged_at"`
	AcknowledgedBy *uint       `json:"acknowledged_by"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	AcknowledgedByUser User `json:"acknowledged_by_user,omitempty" gorm:"foreignKey:AcknowledgedBy"`
}

// RequestLog represents detailed request logging for security monitoring
type RequestLog struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Method          string         `json:"method" gorm:"not null;size:10;index"`
	Path            string         `json:"path" gorm:"not null;size:500;index"`
	ClientIP        string         `json:"client_ip" gorm:"not null;size:45;index"`
	UserAgent       string         `json:"user_agent" gorm:"type:text"`
	StatusCode      int            `json:"status_code" gorm:"index"`
	ResponseTime    int64          `json:"response_time"` // milliseconds
	RequestSize     int64          `json:"request_size"`
	ResponseSize    int64          `json:"response_size"`
	UserID          *uint          `json:"user_id" gorm:"index"`
	SessionID       string         `json:"session_id" gorm:"size:255;index"`
	IsSuspicious    bool           `json:"is_suspicious" gorm:"default:false;index"`
	SuspiciousReason string        `json:"suspicious_reason" gorm:"size:200"`
	Timestamp       time.Time      `json:"timestamp" gorm:"index"`
	CreatedAt       time.Time      `json:"created_at"`
	
	// Relations  
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// IpWhitelist represents allowed IPs for specific environments
type IpWhitelist struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	IpAddress    string         `json:"ip_address" gorm:"not null;size:45;index"`
	IpRange      string         `json:"ip_range" gorm:"size:50"` // CIDR notation if range
	Environment  string         `json:"environment" gorm:"not null;size:20;index"` // production, development, staging
	Description  string         `json:"description" gorm:"size:200"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	AddedBy      uint           `json:"added_by" gorm:"not null"`
	ExpiresAt    *time.Time     `json:"expires_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	AddedByUser User `json:"added_by_user,omitempty" gorm:"foreignKey:AddedBy"`
}

// SecurityConfig represents system security configuration
type SecurityConfig struct {
	ID                    uint           `json:"id" gorm:"primaryKey"`
	Key                   string         `json:"key" gorm:"unique;not null;size:100"` // config key
	Value                 string         `json:"value" gorm:"type:text"`
	DataType              string         `json:"data_type" gorm:"size:20"` // string, int, bool, json
	Environment           string         `json:"environment" gorm:"size:20"` // production, development, all
	Description           string         `json:"description" gorm:"type:text"`
	IsEncrypted           bool           `json:"is_encrypted" gorm:"default:false"`
	LastModifiedBy        uint           `json:"last_modified_by" gorm:"not null"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	LastModifiedByUser User `json:"last_modified_by_user,omitempty" gorm:"foreignKey:LastModifiedBy"`
}

// SecurityMetrics represents aggregated security metrics
type SecurityMetrics struct {
	ID                     uint      `json:"id" gorm:"primaryKey"`
	Date                   time.Time `json:"date" gorm:"type:date;index;unique"`
	TotalRequests          int64     `json:"total_requests"`
	AuthSuccessRate        float64   `json:"auth_success_rate"`
	SuspiciousRequestCount int64     `json:"suspicious_request_count"`
	BlockedIpCount         int64     `json:"blocked_ip_count"`
	RateLimitViolations    int64     `json:"rate_limit_violations"`
	TokenRefreshCount      int64     `json:"token_refresh_count"`
	SecurityIncidentCount  int64     `json:"security_incident_count"`
	AvgResponseTime        float64   `json:"avg_response_time"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// Incident severity levels
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium" 
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Incident types
const (
	IncidentTypeIPWhitelistViolation = "IP_WHITELIST_VIOLATION"
	IncidentTypeSuspiciousRequest    = "SUSPICIOUS_REQUEST"
	IncidentTypeRateLimitExceeded    = "RATE_LIMIT_EXCEEDED"
	IncidentTypeMultipleFailedAuth   = "MULTIPLE_FAILED_AUTH"
	IncidentTypeUnauthorizedAccess   = "UNAUTHORIZED_ACCESS"
	IncidentTypeTokenAbuseDetected   = "TOKEN_ABUSE_DETECTED"
	IncidentTypeSQLInjectionAttempt  = "SQL_INJECTION_ATTEMPT"
	IncidentTypeXSSAttempt           = "XSS_ATTEMPT"
	IncidentTypeDirectoryTraversal   = "DIRECTORY_TRAVERSAL"
)

// Alert types
const (
	AlertTypeRateLimit         = "RATE_LIMIT"
	AlertTypeSuspiciousActivity = "SUSPICIOUS_ACTIVITY"
	AlertTypeSecurityBreach    = "SECURITY_BREACH"
	AlertTypeSystemHealth      = "SYSTEM_HEALTH"
	AlertTypePerformance       = "PERFORMANCE"
)

// Alert levels
const (
	AlertLevelInfo     = "info"
	AlertLevelWarning  = "warning"
	AlertLevelError    = "error"
	AlertLevelCritical = "critical"
)
