package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// SecurityMonitor handles security monitoring and alerting
type SecurityMonitor struct {
	DB           *gorm.DB
	Config       *config.Config
	alertQueue   chan SecurityAlert
	metricsStore *SecurityMetrics
	mu           sync.RWMutex
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
	Title       string    `json:"title"`
	Description string    `json:"description"`
	UserID      uint      `json:"user_id,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	Endpoint    string    `json:"endpoint,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Resolved    bool      `json:"resolved"`
}

// SecurityMetrics stores security metrics
type SecurityMetrics struct {
	TotalLoginAttempts      int64
	FailedLoginAttempts     int64
	TokensIssued            int64
	TokensRefreshed         int64
	TokensRevoked           int64
	RateLimitViolations     int64
	SuspiciousActivities    int64
	LastAlertTime           time.Time
	ActiveSessions          int64
	UniqueIPsLastHour       int64
	BlockedIPs              []string
}

// AlertType constants
const (
	AlertTypeAuthentication  = "AUTHENTICATION"
	AlertTypeAuthorization   = "AUTHORIZATION"
	AlertTypeRateLimit       = "RATE_LIMIT"
	AlertTypeSessionAnomaly  = "SESSION_ANOMALY"
	AlertTypeSuspiciousActivity = "SUSPICIOUS_ACTIVITY"
	AlertTypeSystemSecurity  = "SYSTEM_SECURITY"
)

// Severity levels
const (
	SeverityLow      = "LOW"
	SeverityMedium   = "MEDIUM"
	SeverityHigh     = "HIGH"
	SeverityCritical = "CRITICAL"
)

// NewSecurityMonitor creates a new security monitor
func NewSecurityMonitor(db *gorm.DB, cfg *config.Config) *SecurityMonitor {
	sm := &SecurityMonitor{
		DB:           db,
		Config:       cfg,
		alertQueue:   make(chan SecurityAlert, 100),
		metricsStore: &SecurityMetrics{},
	}
	
	// Start monitoring goroutines
	if cfg.EnableMonitoring {
		go sm.processAlerts()
		go sm.collectMetrics()
		go sm.periodicHealthCheck()
	}
	
	return sm
}

// LogSecurityEvent logs a security event
func (sm *SecurityMonitor) LogSecurityEvent(eventType, description string, metadata map[string]interface{}) {
	event := SecurityAlert{
		Type:        eventType,
		Severity:    sm.determineSeverity(eventType, metadata),
		Title:       sm.generateTitle(eventType),
		Description: description,
		Metadata:    metadata,
		Timestamp:   time.Now(),
		Resolved:    false,
	}
	
	// Extract additional info from metadata
	if userID, ok := metadata["user_id"].(uint); ok {
		event.UserID = userID
	}
	if ip, ok := metadata["ip_address"].(string); ok {
		event.IPAddress = ip
	}
	if endpoint, ok := metadata["endpoint"].(string); ok {
		event.Endpoint = endpoint
	}
	
	// Send to alert queue
	select {
	case sm.alertQueue <- event:
	default:
		log.Printf("Alert queue full, dropping event: %s", eventType)
	}
	
	// Store in database
	sm.DB.Create(&event)
	
	// Update metrics
	sm.updateMetrics(eventType)
}

// MonitorFailedLogin monitors failed login attempts
func (sm *SecurityMonitor) MonitorFailedLogin(username, ipAddress string, reason string) {
	sm.mu.Lock()
	sm.metricsStore.FailedLoginAttempts++
	sm.mu.Unlock()
	
	// Check for brute force attempts
	var failedCount int64
	sm.DB.Model(&models.AuthAttempt{}).
		Where("ip_address = ? AND success = false AND attempted_at > ?", 
			ipAddress, time.Now().Add(-15*time.Minute)).
		Count(&failedCount)
	
	if failedCount >= 5 {
		sm.LogSecurityEvent(AlertTypeAuthentication, 
			fmt.Sprintf("Multiple failed login attempts from IP: %s for user: %s", ipAddress, username),
			map[string]interface{}{
				"username": username,
				"ip_address": ipAddress,
				"failed_count": failedCount,
				"reason": reason,
			})
		
		// Consider blocking the IP
		if failedCount >= 10 {
			sm.blockIP(ipAddress)
		}
	}
}

// MonitorTokenActivity monitors JWT token activities
func (sm *SecurityMonitor) MonitorTokenActivity(activity string, userID uint, metadata map[string]interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	switch activity {
	case "issued":
		sm.metricsStore.TokensIssued++
	case "refreshed":
		sm.metricsStore.TokensRefreshed++
	case "revoked":
		sm.metricsStore.TokensRevoked++
	case "blacklisted":
		sm.metricsStore.TokensRevoked++
		
		// Log suspicious token revocation
		sm.LogSecurityEvent(AlertTypeSessionAnomaly,
			fmt.Sprintf("Token blacklisted for user ID: %d", userID),
			metadata)
	}
}

// MonitorRateLimit monitors rate limit violations
func (sm *SecurityMonitor) MonitorRateLimit(ipAddress, endpoint string, violations int) {
	sm.mu.Lock()
	sm.metricsStore.RateLimitViolations++
	sm.mu.Unlock()
	
	if violations >= 3 {
		sm.LogSecurityEvent(AlertTypeRateLimit,
			fmt.Sprintf("Repeated rate limit violations from IP: %s on endpoint: %s", ipAddress, endpoint),
			map[string]interface{}{
				"ip_address": ipAddress,
				"endpoint": endpoint,
				"violations": violations,
			})
	}
}

// CheckSessionAnomaly checks for session anomalies
func (sm *SecurityMonitor) CheckSessionAnomaly(userID uint, currentIP, previousIP string, deviceChange bool) {
	if currentIP != previousIP {
		metadata := map[string]interface{}{
			"user_id": userID,
			"current_ip": currentIP,
			"previous_ip": previousIP,
			"device_change": deviceChange,
		}
		
		sm.LogSecurityEvent(AlertTypeSessionAnomaly,
			fmt.Sprintf("IP address change detected for user ID: %d", userID),
			metadata)
		
		// Check for impossible travel (simplified version)
		if sm.isImpossibleTravel(previousIP, currentIP) {
			sm.LogSecurityEvent(AlertTypeSuspiciousActivity,
				fmt.Sprintf("Possible account compromise - impossible travel detected for user ID: %d", userID),
				metadata)
		}
	}
}

// processAlerts processes queued alerts
func (sm *SecurityMonitor) processAlerts() {
	for alert := range sm.alertQueue {
		// Log to console
		log.Printf("[SECURITY ALERT] %s: %s", alert.Type, alert.Description)
		
		// Send notifications based on severity
		switch alert.Severity {
		case SeverityCritical:
			sm.sendImmediateNotification(alert)
		case SeverityHigh:
			sm.sendEmailNotification(alert)
		case SeverityMedium:
			sm.logToFile(alert)
		default:
			// Just log for low severity
		}
		
		sm.mu.Lock()
		sm.metricsStore.LastAlertTime = alert.Timestamp
		sm.mu.Unlock()
	}
}

// collectMetrics collects security metrics periodically
func (sm *SecurityMonitor) collectMetrics() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		sm.mu.Lock()
		
		// Count active sessions
		var activeSessions int64
		sm.DB.Model(&models.UserSession{}).Where("is_active = true").Count(&activeSessions)
		sm.metricsStore.ActiveSessions = activeSessions
		
		// Count unique IPs in last hour
		var uniqueIPs int64
		sm.DB.Model(&models.AuthAttempt{}).
			Where("attempted_at > ?", time.Now().Add(-1*time.Hour)).
			Distinct("ip_address").
			Count(&uniqueIPs)
		sm.metricsStore.UniqueIPsLastHour = uniqueIPs
		
		sm.mu.Unlock()
		
		// Check for anomalies
		if activeSessions > 1000 {
			sm.LogSecurityEvent(AlertTypeSystemSecurity,
				fmt.Sprintf("Unusually high number of active sessions: %d", activeSessions),
				map[string]interface{}{"active_sessions": activeSessions})
		}
	}
}

// periodicHealthCheck performs periodic security health checks
func (sm *SecurityMonitor) periodicHealthCheck() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		sm.performHealthCheck()
	}
}

// performHealthCheck performs a security health check
func (sm *SecurityMonitor) performHealthCheck() {
	sm.mu.RLock()
	metrics := *sm.metricsStore
	sm.mu.RUnlock()
	
	// Check failed login ratio
	if metrics.TotalLoginAttempts > 0 {
		failureRate := float64(metrics.FailedLoginAttempts) / float64(metrics.TotalLoginAttempts)
		if failureRate > 0.3 { // More than 30% failure rate
			sm.LogSecurityEvent(AlertTypeSuspiciousActivity,
				fmt.Sprintf("High login failure rate detected: %.2f%%", failureRate*100),
				map[string]interface{}{
					"failure_rate": failureRate,
					"total_attempts": metrics.TotalLoginAttempts,
					"failed_attempts": metrics.FailedLoginAttempts,
				})
		}
	}
	
	// Check rate limit violations
	if metrics.RateLimitViolations > 100 {
		sm.LogSecurityEvent(AlertTypeRateLimit,
			fmt.Sprintf("High number of rate limit violations: %d", metrics.RateLimitViolations),
			map[string]interface{}{"violations": metrics.RateLimitViolations})
	}
}

// Helper methods

func (sm *SecurityMonitor) determineSeverity(eventType string, metadata map[string]interface{}) string {
	switch eventType {
	case AlertTypeAuthentication:
		if count, ok := metadata["failed_count"].(int64); ok && count > 10 {
			return SeverityHigh
		}
		return SeverityMedium
	case AlertTypeRateLimit:
		if violations, ok := metadata["violations"].(int); ok && violations > 10 {
			return SeverityHigh
		}
		return SeverityMedium
	case AlertTypeSuspiciousActivity:
		return SeverityCritical
	case AlertTypeSystemSecurity:
		return SeverityHigh
	default:
		return SeverityLow
	}
}

func (sm *SecurityMonitor) generateTitle(eventType string) string {
	titles := map[string]string{
		AlertTypeAuthentication: "Authentication Security Alert",
		AlertTypeAuthorization: "Authorization Security Alert",
		AlertTypeRateLimit: "Rate Limit Violation Alert",
		AlertTypeSessionAnomaly: "Session Anomaly Detected",
		AlertTypeSuspiciousActivity: "Suspicious Activity Detected",
		AlertTypeSystemSecurity: "System Security Alert",
	}
	
	if title, ok := titles[eventType]; ok {
		return title
	}
	return "Security Alert"
}

func (sm *SecurityMonitor) updateMetrics(eventType string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	switch eventType {
	case AlertTypeAuthentication:
		sm.metricsStore.SuspiciousActivities++
	case AlertTypeRateLimit:
		// Already counted in MonitorRateLimit
	case AlertTypeSuspiciousActivity:
		sm.metricsStore.SuspiciousActivities++
	}
}

func (sm *SecurityMonitor) blockIP(ipAddress string) {
	sm.mu.Lock()
	sm.metricsStore.BlockedIPs = append(sm.metricsStore.BlockedIPs, ipAddress)
	sm.mu.Unlock()
	
	// Store in database
	// This would integrate with firewall or WAF in production
	log.Printf("IP %s has been blocked due to suspicious activity", ipAddress)
}

func (sm *SecurityMonitor) isImpossibleTravel(previousIP, currentIP string) bool {
	// Simplified check - in production, use GeoIP database
	// to calculate actual distance and time
	return false
}

func (sm *SecurityMonitor) sendImmediateNotification(alert SecurityAlert) {
	// Send webhook notification
	if sm.Config.AlertWebhookURL != "" {
		payload, _ := json.Marshal(alert)
		go func() {
			resp, err := http.Post(sm.Config.AlertWebhookURL, "application/json", bytes.NewBuffer(payload))
			if err != nil {
				log.Printf("Failed to send webhook notification: %v", err)
			} else {
				resp.Body.Close()
			}
		}()
	}
}

func (sm *SecurityMonitor) sendEmailNotification(alert SecurityAlert) {
	// Email notification implementation
	if sm.Config.AlertEmail != "" {
		// In production, use SMTP or email service
		log.Printf("Email alert would be sent to %s: %s", sm.Config.AlertEmail, alert.Description)
	}
}

func (sm *SecurityMonitor) logToFile(alert SecurityAlert) {
	if sm.Config.LogFile != "" {
		// Log to file implementation
		log.Printf("Security alert logged: %s", alert.Description)
	}
}

// GetMetrics returns current security metrics
func (sm *SecurityMonitor) GetMetrics() SecurityMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return *sm.metricsStore
}

// GetRecentAlerts returns recent security alerts
func (sm *SecurityMonitor) GetRecentAlerts(limit int) []SecurityAlert {
	var alerts []SecurityAlert
	sm.DB.Order("timestamp desc").Limit(limit).Find(&alerts)
	return alerts
}

// MarkAlertResolved marks an alert as resolved
func (sm *SecurityMonitor) MarkAlertResolved(alertID uint) error {
	return sm.DB.Model(&SecurityAlert{}).Where("id = ?", alertID).Update("resolved", true).Error
}
