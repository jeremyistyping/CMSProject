package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"gorm.io/gorm"
)

func main() {
	log.Println("🔐 Testing Enhanced Security System...")
	
	// Connect to database
	db := database.ConnectDB()
	
	// Initialize database with security tables
	log.Println("📦 Running database migrations...")
	database.InitializeDatabase(db)
	
	// Test security service
	testSecurityService(db)
	
	// Test security models
	testSecurityModels(db)
	
	// Test security metrics
	testSecurityMetrics(db)
	
	log.Println("✅ Security system tests completed!")
}

func testSecurityService(db *gorm.DB) {
	log.Println("\n🔧 Testing SecurityService...")
	
	securityService := services.NewSecurityService(db)
	
	// Test logging security incident
	err := securityService.LogSecurityIncident(
		models.IncidentTypeSuspiciousRequest,
		models.SeverityHigh,
		"Test suspicious request from automated testing",
		"127.0.0.1",
		"test-user-agent/1.0",
		"GET",
		"/test-suspicious-path?union=select",
		"",
		nil,
		"test-session-123",
	)
	
	if err != nil {
		log.Printf("❌ Failed to log security incident: %v", err)
	} else {
		log.Println("✅ Security incident logged successfully")
	}
	
	// Test system alert
	err = securityService.CreateAlert(
		models.AlertTypeSecurityBreach,
		models.AlertLevelCritical,
		"Test Security Alert",
		"This is a test security alert from automated testing",
	)
	
	if err != nil {
		log.Printf("❌ Failed to create system alert: %v", err)
	} else {
		log.Println("✅ System alert created successfully")
	}
	
	// Test suspicious pattern detection
	headers := map[string]string{
		"User-Agent": "sqlmap/1.0",
		"X-Custom": "test-header",
	}
	
	isSuspicious, reason := securityService.DetectSuspiciousPattern(
		"GET", 
		"/api/users?id=1' UNION SELECT * FROM users--", 
		"sqlmap/1.0", 
		"192.168.1.100", 
		headers,
	)
	
	if isSuspicious {
		log.Printf("✅ Suspicious pattern detected: %s", reason)
	} else {
		log.Println("❌ Failed to detect suspicious pattern")
	}
	
	// Test IP whitelisting (should return false for test IP)
	isWhitelisted := securityService.IsIPWhitelisted("192.168.1.100")
	log.Printf("🔍 IP whitelist check for 192.168.1.100: %t", isWhitelisted)
}

func testSecurityModels(db *gorm.DB) {
	log.Println("\n📊 Testing Security Models...")
	
	// Create test admin user if not exists
	var adminUser models.User
	result := db.Where("username = ?", "admin").First(&adminUser)
	if result.Error == gorm.ErrRecordNotFound {
		adminUser = models.User{
			Username: "admin",
			Email:    "admin@test.com",
			Role:     "admin",
			IsActive: true,
		}
		db.Create(&adminUser)
		log.Println("✅ Test admin user created")
	}
	
	// Test SecurityIncident model
	incident := models.SecurityIncident{
		IncidentType:   models.IncidentTypeIPWhitelistViolation,
		Severity:       models.SeverityMedium,
		Description:    "Test IP whitelist violation",
		ClientIP:       "10.0.0.1",
		UserAgent:      "curl/7.68.0",
		RequestMethod:  "POST",
		RequestPath:    "/api/admin/users",
		RequestHeaders: `{"Authorization": "Bearer invalid-token"}`,
		SessionID:      "test-session-456",
		Resolved:       false,
	}
	
	if err := db.Create(&incident).Error; err != nil {
		log.Printf("❌ Failed to create security incident: %v", err)
	} else {
		log.Printf("✅ Security incident created with ID: %d", incident.ID)
	}
	
	// Test SystemAlert model
	alert := models.SystemAlert{
		AlertType:    models.AlertTypeSuspiciousActivity,
		Level:        models.AlertLevelWarning,
		Title:        "Test Suspicious Activity",
		Message:      "Multiple failed login attempts detected",
		Count:        3,
		FirstSeen:    time.Now(),
		LastSeen:     time.Now(),
		Acknowledged: false,
	}
	
	if err := db.Create(&alert).Error; err != nil {
		log.Printf("❌ Failed to create system alert: %v", err)
	} else {
		log.Printf("✅ System alert created with ID: %d", alert.ID)
	}
	
	// Test IpWhitelist model
	whitelist := models.IpWhitelist{
		IpAddress:   "127.0.0.1",
		Environment: "development",
		Description: "Localhost for development",
		IsActive:    true,
		AddedBy:     adminUser.ID,
	}
	
	if err := db.Create(&whitelist).Error; err != nil {
		log.Printf("❌ Failed to create IP whitelist entry: %v", err)
	} else {
		log.Printf("✅ IP whitelist entry created with ID: %d", whitelist.ID)
	}
	
	// Test SecurityConfig model
	config := models.SecurityConfig{
		Key:               "max_login_attempts",
		Value:             "5",
		DataType:          "int",
		Environment:       "all",
		Description:       "Maximum login attempts before lockout",
		IsEncrypted:       false,
		LastModifiedBy:    adminUser.ID,
	}
	
	if err := db.Create(&config).Error; err != nil {
		log.Printf("❌ Failed to create security config: %v", err)
	} else {
		log.Printf("✅ Security config created with ID: %d", config.ID)
	}
}

func testSecurityMetrics(db *gorm.DB) {
	log.Println("\n📈 Testing Security Metrics...")
	
	// Create sample security metrics
	today := time.Now().Format("2006-01-02")
	todayTime, _ := time.Parse("2006-01-02", today)
	
	metrics := models.SecurityMetrics{
		Date:                   todayTime,
		TotalRequests:          1500,
		AuthSuccessRate:        95.5,
		SuspiciousRequestCount: 12,
		BlockedIpCount:         3,
		RateLimitViolations:    8,
		TokenRefreshCount:      45,
		SecurityIncidentCount:  2,
		AvgResponseTime:        125.3,
	}
	
	if err := db.Create(&metrics).Error; err != nil {
		log.Printf("❌ Failed to create security metrics: %v", err)
	} else {
		log.Printf("✅ Security metrics created for date: %s", today)
	}
	
	// Test querying security data
	var incidentCount int64
	db.Model(&models.SecurityIncident{}).Where("created_at >= ?", time.Now().Add(-24*time.Hour)).Count(&incidentCount)
	log.Printf("📊 Security incidents in last 24h: %d", incidentCount)
	
	var alertCount int64
	db.Model(&models.SystemAlert{}).Where("acknowledged = ?", false).Count(&alertCount)
	log.Printf("🚨 Unacknowledged alerts: %d", alertCount)
	
	var whitelistCount int64
	db.Model(&models.IpWhitelist{}).Where("is_active = ?", true).Count(&whitelistCount)
	log.Printf("📝 Active IP whitelist entries: %d", whitelistCount)
}

// Test environment variables and configuration
func init() {
	// Set test environment variables if not already set
	if os.Getenv("APP_ENV") == "" {
		os.Setenv("APP_ENV", "development")
	}
	if os.Getenv("ENABLE_DEBUG_ROUTES") == "" {
		os.Setenv("ENABLE_DEBUG_ROUTES", "true")
	}
	if os.Getenv("SECURITY_ALLOWED_IPS") == "" {
		os.Setenv("SECURITY_ALLOWED_IPS", "127.0.0.1,::1,localhost")
	}
	if os.Getenv("DETAILED_REQUEST_LOGGING") == "" {
		os.Setenv("DETAILED_REQUEST_LOGGING", "true")
	}
	if os.Getenv("SECURITY_LOG_DIR") == "" {
		os.Setenv("SECURITY_LOG_DIR", "./logs/security")
	}
	
	log.Println("🔧 Test environment variables configured")
	
	// Print current configuration
	config := map[string]string{
		"APP_ENV":                   os.Getenv("APP_ENV"),
		"ENABLE_DEBUG_ROUTES":       os.Getenv("ENABLE_DEBUG_ROUTES"),
		"SECURITY_ALLOWED_IPS":      os.Getenv("SECURITY_ALLOWED_IPS"),
		"DETAILED_REQUEST_LOGGING":  os.Getenv("DETAILED_REQUEST_LOGGING"),
		"SECURITY_LOG_DIR":          os.Getenv("SECURITY_LOG_DIR"),
	}
	
	configJSON, _ := json.MarshalIndent(config, "", "  ")
	fmt.Printf("🔧 Current Security Configuration:\n%s\n", configJSON)
}
