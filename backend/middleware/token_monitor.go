package middleware

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gorm.io/gorm"
)

// TokenRefreshEvent represents a token refresh event
type TokenRefreshEvent struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"index"`
	Username     string    `json:"username" gorm:"size:100"`
	SessionID    string    `json:"session_id" gorm:"size:255;index"`
	IPAddress    string    `json:"ip_address" gorm:"size:45"`
	UserAgent    string    `json:"user_agent" gorm:"type:text"`
	Success      bool      `json:"success"`
	FailureReason string   `json:"failure_reason" gorm:"size:255"`
	TokenExpiry  time.Time `json:"token_expiry"`
	RefreshCount int       `json:"refresh_count"` // Number of times this session has been refreshed
	Timestamp    time.Time `json:"timestamp"`
	CreatedAt    time.Time `json:"created_at"`
}

// TokenUsageStats represents token usage statistics
type TokenUsageStats struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	UserID                uint      `json:"user_id" gorm:"index"`
	Date                  time.Time `json:"date" gorm:"type:date;index"`
	TotalRequests         int       `json:"total_requests"`
	SuccessfulRefreshes   int       `json:"successful_refreshes"`
	FailedRefreshes       int       `json:"failed_refreshes"`
	AverageSessionDuration int64    `json:"average_session_duration"` // in minutes
	PeakHourRequests      int       `json:"peak_hour_requests"`
	SuspiciousActivity    bool      `json:"suspicious_activity"`
	LastUpdated           time.Time `json:"last_updated"`
	CreatedAt             time.Time `json:"created_at"`
}

// TokenMonitor manages token refresh monitoring
type TokenMonitor struct {
	db     *gorm.DB
	logger *log.Logger
	mu     sync.RWMutex
	stats  map[uint]*dailyStats // userID -> daily stats
}

// dailyStats holds temporary daily statistics
type dailyStats struct {
	requests    int
	refreshes   int
	failures    int
	lastRefresh time.Time
	suspiciousPatterns []string
}

// NewTokenMonitor creates a new token monitor
func NewTokenMonitor(db *gorm.DB) *TokenMonitor {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, 0755)
	}

	// Create token monitoring log file
	logFile, err := os.OpenFile(
		filepath.Join(logDir, "token_monitor.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		log.Printf("Failed to create token monitor log file: %v", err)
		logFile = os.Stdout
	}

	logger := log.New(logFile, "[TOKEN_MONITOR] ", log.LstdFlags|log.Lmicroseconds)

	monitor := &TokenMonitor{
		db:     db,
		logger: logger,
		stats:  make(map[uint]*dailyStats),
	}

	// Start background processes
	go monitor.aggregateStats()
	go monitor.detectAnomalies()

	return monitor
}

// LogRefreshAttempt logs a token refresh attempt
func (tm *TokenMonitor) LogRefreshAttempt(userID uint, username, sessionID, ipAddress, userAgent string, success bool, failureReason string, tokenExpiry time.Time, refreshCount int) {
	event := &TokenRefreshEvent{
		UserID:        userID,
		Username:      username,
		SessionID:     sessionID,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Success:       success,
		FailureReason: failureReason,
		TokenExpiry:   tokenExpiry,
		RefreshCount:  refreshCount,
		Timestamp:     time.Now(),
	}

	// Log to database (non-blocking)
	go tm.logEventToDatabase(event)

	// Log to file
	tm.logEventToFile(event)

	// Update daily statistics
	tm.updateDailyStats(userID, success)

	// Check for suspicious activity
	if tm.isSuspiciousActivity(userID, ipAddress, refreshCount) {
		tm.logSecurityAlert(event)
	}
}

// logEventToDatabase saves refresh event to database
func (tm *TokenMonitor) logEventToDatabase(event *TokenRefreshEvent) {
	if tm.db != nil {
		if err := tm.db.Create(event).Error; err != nil {
			log.Printf("Failed to save token refresh event to database: %v", err)
		}
	}
}

// logEventToFile writes refresh event to log file
func (tm *TokenMonitor) logEventToFile(event *TokenRefreshEvent) {
	logData := map[string]interface{}{
		"user_id":        event.UserID,
		"username":       event.Username,
		"session_id":     event.SessionID,
		"ip_address":     event.IPAddress,
		"success":        event.Success,
		"failure_reason": event.FailureReason,
		"refresh_count":  event.RefreshCount,
		"timestamp":      event.Timestamp.Format(time.RFC3339),
	}

	if event.TokenExpiry.After(time.Now()) {
		logData["token_expiry"] = event.TokenExpiry.Format(time.RFC3339)
	}

	jsonData, _ := json.Marshal(logData)
	tm.logger.Printf("%s\n", string(jsonData))
}

// updateDailyStats updates daily statistics for a user
func (tm *TokenMonitor) updateDailyStats(userID uint, success bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	stats, exists := tm.stats[userID]
	if !exists {
		stats = &dailyStats{
			lastRefresh: time.Now(),
			suspiciousPatterns: make([]string, 0),
		}
		tm.stats[userID] = stats
	}

	stats.requests++
	if success {
		stats.refreshes++
		stats.lastRefresh = time.Now()
	} else {
		stats.failures++
	}
}

// isSuspiciousActivity checks if the refresh pattern is suspicious
func (tm *TokenMonitor) isSuspiciousActivity(userID uint, ipAddress string, refreshCount int) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	stats, exists := tm.stats[userID]
	if !exists {
		return false
	}

	// Check for various suspicious patterns
	suspicious := false
	reasons := make([]string, 0)

	// Too many refresh attempts in short time
	if stats.refreshes > 50 && time.Since(stats.lastRefresh) < time.Hour {
		suspicious = true
		reasons = append(reasons, "HIGH_FREQUENCY_REFRESH")
	}

	// High failure rate
	if stats.failures > 10 && float64(stats.failures)/float64(stats.requests) > 0.5 {
		suspicious = true
		reasons = append(reasons, "HIGH_FAILURE_RATE")
	}

	// Excessive refresh count for single session
	if refreshCount > 100 {
		suspicious = true
		reasons = append(reasons, "EXCESSIVE_SESSION_REFRESH")
	}

	if suspicious {
		stats.suspiciousPatterns = append(stats.suspiciousPatterns, reasons...)
	}

	return suspicious
}

// logSecurityAlert logs suspicious token refresh activity
func (tm *TokenMonitor) logSecurityAlert(event *TokenRefreshEvent) {
	tm.mu.RLock()
	patterns := tm.stats[event.UserID].suspiciousPatterns
	tm.mu.RUnlock()

	alert := map[string]interface{}{
		"type":             "SUSPICIOUS_TOKEN_ACTIVITY",
		"user_id":          event.UserID,
		"username":         event.Username,
		"session_id":       event.SessionID,
		"ip_address":       event.IPAddress,
		"refresh_count":    event.RefreshCount,
		"suspicious_patterns": patterns,
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(alert)
	tm.logger.Printf("SECURITY_ALERT: %s\n", string(jsonData))

	// Also log to general audit if available
	if GlobalAuditLogger != nil {
		GlobalAuditLogger.logger.Printf("TOKEN_SECURITY_ALERT: %s\n", string(jsonData))
	}
}

// aggregateStats aggregates daily statistics and saves them to database
func (tm *TokenMonitor) aggregateStats() {
	ticker := time.NewTicker(1 * time.Hour) // Aggregate every hour
	defer ticker.Stop()

	for range ticker.C {
		tm.mu.Lock()
		
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		for userID, stats := range tm.stats {
			if stats.requests == 0 {
				continue
			}

			// Calculate average session duration (simplified)
			avgDuration := int64(90) // Default 90 minutes
			if stats.refreshes > 0 {
				avgDuration = int64(stats.refreshes * 90) // 90 minutes per refresh
			}

			// Check for existing stats for today
			var existingStats TokenUsageStats
			err := tm.db.Where("user_id = ? AND date = ?", userID, today).First(&existingStats).Error

			usageStats := TokenUsageStats{
				UserID:                userID,
				Date:                  today,
				TotalRequests:         stats.requests,
				SuccessfulRefreshes:   stats.refreshes,
				FailedRefreshes:       stats.failures,
				AverageSessionDuration: avgDuration,
				PeakHourRequests:      stats.requests, // Simplified
				SuspiciousActivity:    len(stats.suspiciousPatterns) > 0,
				LastUpdated:           now,
			}

			if err == gorm.ErrRecordNotFound {
				// Create new record
				tm.db.Create(&usageStats)
			} else {
				// Update existing record
				existingStats.TotalRequests += stats.requests
				existingStats.SuccessfulRefreshes += stats.refreshes
				existingStats.FailedRefreshes += stats.failures
				existingStats.SuspiciousActivity = existingStats.SuspiciousActivity || len(stats.suspiciousPatterns) > 0
				existingStats.LastUpdated = now
				tm.db.Save(&existingStats)
			}

			// Reset daily stats
			tm.stats[userID] = &dailyStats{
				suspiciousPatterns: make([]string, 0),
			}
		}
		
		tm.mu.Unlock()
	}
}

// detectAnomalies runs anomaly detection on token usage patterns
func (tm *TokenMonitor) detectAnomalies() {
	ticker := time.NewTicker(6 * time.Hour) // Check every 6 hours
	defer ticker.Stop()

	for range ticker.C {
		tm.runAnomalyDetection()
	}
}

// runAnomalyDetection performs anomaly detection on recent token usage
func (tm *TokenMonitor) runAnomalyDetection() {
	if tm.db == nil {
		return
	}

	// Get token usage stats from the last 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	
	var stats []TokenUsageStats
	err := tm.db.Where("date >= ?", sevenDaysAgo).Find(&stats).Error
	if err != nil {
		log.Printf("Failed to get token usage stats for anomaly detection: %v", err)
		return
	}

	// Analyze patterns
	userStats := make(map[uint][]TokenUsageStats)
	for _, stat := range stats {
		userStats[stat.UserID] = append(userStats[stat.UserID], stat)
	}

	for userID, userUsage := range userStats {
		if len(userUsage) < 3 {
			continue // Need at least 3 days of data
		}

		// Calculate averages
		avgRequests := calculateAverage(userUsage, "requests")
		avgFailures := calculateAverage(userUsage, "failures")

		// Check for anomalies
		latestUsage := userUsage[len(userUsage)-1]
		
		// Anomaly: Sudden spike in requests (3x average)
		if float64(latestUsage.TotalRequests) > avgRequests*3 {
			tm.logAnomaly(userID, "SPIKE_IN_REQUESTS", map[string]interface{}{
				"current_requests": latestUsage.TotalRequests,
				"average_requests": avgRequests,
			})
		}

		// Anomaly: High failure rate
		if float64(latestUsage.FailedRefreshes) > avgFailures*2 && latestUsage.FailedRefreshes > 5 {
			tm.logAnomaly(userID, "HIGH_FAILURE_RATE", map[string]interface{}{
				"current_failures": latestUsage.FailedRefreshes,
				"average_failures": avgFailures,
			})
		}

		// Anomaly: Suspicious activity flag
		if latestUsage.SuspiciousActivity {
			tm.logAnomaly(userID, "FLAGGED_SUSPICIOUS_ACTIVITY", map[string]interface{}{
				"date": latestUsage.Date.Format("2006-01-02"),
			})
		}
	}
}

// calculateAverage calculates average for a specific field
func calculateAverage(stats []TokenUsageStats, field string) float64 {
	if len(stats) == 0 {
		return 0
	}

	total := 0.0
	for _, stat := range stats {
		switch field {
		case "requests":
			total += float64(stat.TotalRequests)
		case "refreshes":
			total += float64(stat.SuccessfulRefreshes)
		case "failures":
			total += float64(stat.FailedRefreshes)
		}
	}

	return total / float64(len(stats))
}

// logAnomaly logs detected anomaly
func (tm *TokenMonitor) logAnomaly(userID uint, anomalyType string, details map[string]interface{}) {
	anomaly := map[string]interface{}{
		"type":        "TOKEN_USAGE_ANOMALY",
		"user_id":     userID,
		"anomaly":     anomalyType,
		"details":     details,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(anomaly)
	tm.logger.Printf("ANOMALY: %s\n", string(jsonData))
}

// GetTokenStats returns token usage statistics for a user
func (tm *TokenMonitor) GetTokenStats(userID uint, days int) ([]TokenUsageStats, error) {
	if tm.db == nil {
		return nil, nil
	}

	startDate := time.Now().AddDate(0, 0, -days)
	
	var stats []TokenUsageStats
	err := tm.db.Where("user_id = ? AND date >= ?", userID, startDate).
		Order("date DESC").
		Find(&stats).Error

	return stats, err
}

// GetRecentRefreshEvents returns recent token refresh events
func (tm *TokenMonitor) GetRecentRefreshEvents(userID uint, limit int) ([]TokenRefreshEvent, error) {
	if tm.db == nil {
		return nil, nil
	}

	var events []TokenRefreshEvent
	query := tm.db.Order("timestamp DESC").Limit(limit)
	
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	err := query.Find(&events).Error
	return events, err
}

// Global token monitor instance
var GlobalTokenMonitor *TokenMonitor

// InitTokenMonitor initializes the global token monitor
func InitTokenMonitor(db *gorm.DB) {
	GlobalTokenMonitor = NewTokenMonitor(db)
	
	// Auto-migrate token monitoring tables
	if err := db.AutoMigrate(&TokenRefreshEvent{}, &TokenUsageStats{}); err != nil {
		log.Printf("Failed to migrate token monitoring tables: %v", err)
	}
}
