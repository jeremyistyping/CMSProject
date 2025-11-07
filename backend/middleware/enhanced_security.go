package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Enhanced Security Middleware
type EnhancedSecurityMiddleware struct {
	db                 *gorm.DB
	securityService    *services.SecurityService
	allowedIPs         []string
	trustedProxies     []string
	maxRequestsPerMin  int
	securityHeaders    map[string]string
}

// Initialize enhanced security middleware
func NewEnhancedSecurityMiddleware(db *gorm.DB) *EnhancedSecurityMiddleware {
	// Default security headers
	securityHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY", 
		"X-XSS-Protection":       "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'",
	}

	// Get allowed IPs from environment or use defaults
	allowedIPs := getSecurityAllowedIPs()
	trustedProxies := getTrustedProxies()

	return &EnhancedSecurityMiddleware{
		db:                db,
		securityService:   services.NewSecurityService(db),
		allowedIPs:        allowedIPs,
		trustedProxies:    trustedProxies,
		maxRequestsPerMin: 60,
		securityHeaders:   securityHeaders,
	}
}

// IP Whitelisting middleware for development endpoints
func (esm *EnhancedSecurityMiddleware) IPWhitelist() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get real IP (considering proxies)
		clientIP := esm.getRealClientIP(c)
		
		// Check database whitelist first, then static IPs
		if !esm.securityService.IsIPWhitelisted(clientIP) && !esm.isIPAllowed(clientIP) {
			log.Printf("🚨 SECURITY: Blocked access from unauthorized IP: %s to %s", clientIP, c.Request.URL.Path)
			
			// Get user info for logging
			var userID *uint
			var sessionID string
			if user, exists := c.Get("user"); exists {
				if u, ok := user.(models.User); ok {
					userID = &u.ID
				}
			}
			if session, exists := c.Get("session_id"); exists {
				if s, ok := session.(string); ok {
					sessionID = s
				}
			}
			
			// Log security incident with enhanced logging
			esm.securityService.LogSecurityIncident(
				models.IncidentTypeIPWhitelistViolation,
				models.SeverityHigh,
				fmt.Sprintf("Unauthorized IP access attempt: %s to %s", clientIP, c.Request.URL.Path),
				clientIP, c.GetHeader("User-Agent"), c.Request.Method, c.Request.URL.Path, "",
				userID, sessionID,
			)
			
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access forbidden: IP not whitelisted",
				"code":  "IP_NOT_ALLOWED",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// Enhanced security headers middleware
func (esm *EnhancedSecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For Swagger and docs endpoints, let the Swagger-specific middleware control CSP.
		path := c.Request.URL.Path
		skipCSP := strings.HasPrefix(path, "/swagger") || strings.HasPrefix(path, "/docs") || strings.HasPrefix(path, "/openapi")

		// Add security headers (optionally skip CSP so it can be set by swaggerCSPMiddleware)
		for header, value := range esm.securityHeaders {
			if skipCSP && strings.EqualFold(header, "Content-Security-Policy") {
				// Ensure any previously-set CSP is cleared for these paths so Swagger override applies cleanly
				c.Writer.Header().Del("Content-Security-Policy")
				continue
			}
			c.Header(header, value)
		}

		// Remove sensitive server headers
		c.Header("Server", "") // Hide server information

		c.Next()
	}
}

// Request monitoring middleware with enhanced security logging
func (esm *EnhancedSecurityMiddleware) RequestMonitoring() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Get request details
		clientIP := esm.getRealClientIP(c)
		userAgent := c.GetHeader("User-Agent")
		method := c.Request.Method
		path := c.Request.URL.Path
		
		// Get user info if authenticated
		var userID *uint
		var sessionID string
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(models.User); ok {
				userID = &u.ID
			}
		}
		if session, exists := c.Get("session_id"); exists {
			if s, ok := session.(string); ok {
				sessionID = s
			}
		}
		
		// Collect headers for analysis
		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		
		// Check for suspicious patterns using security service
		isSuspicious, suspiciousReason := esm.securityService.DetectSuspiciousPattern(method, path, userAgent, clientIP, headers)
		
		// Block highly suspicious requests (but allow localhost and whitelisted IPs)
		if isSuspicious && !esm.securityService.IsIPWhitelisted(clientIP) && !esm.isIPAllowed(clientIP) {
			if strings.Contains(suspiciousReason, "SQL_INJECTION") ||
			   strings.Contains(suspiciousReason, "DIRECTORY_TRAVERSAL") ||
			   strings.Contains(suspiciousReason, "XSS_ATTEMPT") {
				
				// Log critical security incident
				esm.securityService.LogSecurityIncident(
					models.IncidentTypeSuspiciousRequest,
					models.SeverityCritical,
					fmt.Sprintf("Blocked malicious request: %s", suspiciousReason),
					clientIP, userAgent, method, path, "",
					userID, sessionID,
				)
				
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Request blocked for security reasons",
					"code":  "SECURITY_VIOLATION",
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
		
		// Calculate response metrics
		duration := time.Since(start)
		status := c.Writer.Status()
		requestSize := c.Request.ContentLength
		responseSize := int64(c.Writer.Size())
		
		// Log suspicious requests to database
		if isSuspicious {
			esm.securityService.LogSuspiciousRequest(
				method, path, clientIP, userAgent, status, duration.Milliseconds(),
				userID, sessionID, suspiciousReason,
			)
		} else {
			// Log normal requests if detailed logging is enabled
			esm.securityService.LogRequest(
				method, path, clientIP, userAgent, status,
				duration.Milliseconds(), requestSize, responseSize, userID, sessionID,
			)
		}
		
		// Log slow requests
		if duration > 5*time.Second {
			log.Printf("⚠️ PERFORMANCE: Slow request detected: %s %s (%v) from %s", 
				method, path, duration, clientIP)
				
			// Create performance alert
			esm.securityService.CreateAlert(
				models.AlertTypePerformance,
				models.AlertLevelWarning,
				"Slow Request Detected",
				fmt.Sprintf("Request %s %s took %v from %s", method, path, duration, clientIP),
			)
		}
		
		// Log and track security-related status codes
		if status == 401 || status == 403 || status == 429 {
			log.Printf("🔐 SECURITY: Access denied %d for %s %s from %s (UA: %s)", 
				status, method, path, clientIP, userAgent)
				
			// Log as security incident for unauthorized access
			if status == 401 || status == 403 {
				esm.securityService.LogSecurityIncident(
					models.IncidentTypeUnauthorizedAccess,
					models.SeverityMedium,
					fmt.Sprintf("Unauthorized access attempt (HTTP %d)", status),
					clientIP, userAgent, method, path, "",
					userID, sessionID,
				)
			}
		}
	}
}

// Environment-specific route enabler
func (esm *EnhancedSecurityMiddleware) EnvironmentGate(allowedEnvs ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentEnv := getEnvironment()
		
		allowed := false
		for _, env := range allowedEnvs {
			if strings.ToLower(env) == currentEnv {
				allowed = true
				break
			}
		}
		
		if !allowed {
			log.Printf("🚨 SECURITY: Blocked access to environment-restricted endpoint in %s environment: %s", currentEnv, c.Request.URL.Path)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Endpoint not available in current environment",
				"code":  "ENV_RESTRICTED",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// Helper functions

func getSecurityAllowedIPs() []string {
	allowedIPs := os.Getenv("SECURITY_ALLOWED_IPS")
	if allowedIPs == "" {
		// Default safe IPs for development
		return []string{"127.0.0.1", "::1", "localhost"}
	}
	return strings.Split(allowedIPs, ",")
}

func getTrustedProxies() []string {
	proxies := os.Getenv("TRUSTED_PROXIES")
	if proxies == "" {
		return []string{"127.0.0.1", "::1"}
	}
	return strings.Split(proxies, ",")
}

func getEnvironment() string {
	env := strings.ToLower(os.Getenv("ENV"))
	if env == "" {
		env = strings.ToLower(os.Getenv("GO_ENV"))
	}
	if env == "" {
		env = strings.ToLower(os.Getenv("ENVIRONMENT"))
	}
	if env == "" {
		env = "development" // default
	}
	return env
}

func (esm *EnhancedSecurityMiddleware) getRealClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (most common)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}
	
	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}
	
	// Fall back to remote address
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	
	return ip
}

func (esm *EnhancedSecurityMiddleware) isIPAllowed(clientIP string) bool {
	// Parse client IP
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}
	
	for _, allowedIP := range esm.allowedIPs {
		// Handle CIDR notation
		if strings.Contains(allowedIP, "/") {
			_, network, err := net.ParseCIDR(allowedIP)
			if err == nil && network.Contains(ip) {
				return true
			}
		} else {
			// Direct IP comparison
			if allowedIP == clientIP || allowedIP == "localhost" && (clientIP == "127.0.0.1" || clientIP == "::1") {
				return true
			}
		}
	}
	
	return false
}

func (esm *EnhancedSecurityMiddleware) isSuspiciousRequest(c *gin.Context) bool {
	userAgent := c.GetHeader("User-Agent")
	path := c.Request.URL.Path
	
	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"sql", "union", "select", "drop", "delete", "insert", "update",
		"script", "javascript", "vbscript", "onload", "onerror",
		"<script>", "</script>", "eval(", "document.cookie",
		"../", "..\\", "/etc/passwd", "/proc/", "cmd=", "exec=",
	}
	
	// Check URL path for suspicious patterns
	lowerPath := strings.ToLower(path)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}
	
	// Check User-Agent for suspicious patterns
	if userAgent != "" {
		lowerUA := strings.ToLower(userAgent)
		suspiciousUA := []string{"sqlmap", "nikto", "nmap", "masscan", "curl", "wget", "python-requests"}
		for _, pattern := range suspiciousUA {
			if strings.Contains(lowerUA, pattern) {
				return true
			}
		}
	}
	
	return false
}

// Legacy method - now uses SecurityService for enhanced incident logging
func (esm *EnhancedSecurityMiddleware) logSecurityIncident(c *gin.Context, incidentType string, details string) {
	clientIP := esm.getRealClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	method := c.Request.Method
	path := c.Request.URL.Path
	
	// Get user info if available
	var userID *uint
	var sessionID string
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(models.User); ok {
			userID = &u.ID
		}
	}
	if session, exists := c.Get("session_id"); exists {
		if s, ok := session.(string); ok {
			sessionID = s
		}
	}
	
	// Log using SecurityService with proper incident type mapping
	var mappedIncidentType string
	var severity string
	
	switch incidentType {
	case "IP_WHITELIST_VIOLATION":
		mappedIncidentType = models.IncidentTypeIPWhitelistViolation
		severity = models.SeverityHigh
	case "SUSPICIOUS_REQUEST":
		mappedIncidentType = models.IncidentTypeSuspiciousRequest
		severity = models.SeverityMedium
	default:
		mappedIncidentType = models.IncidentTypeSuspiciousRequest
		severity = models.SeverityMedium
	}
	
	esm.securityService.LogSecurityIncident(
		mappedIncidentType, severity, details,
		clientIP, userAgent, method, path, "",
		userID, sessionID,
	)
}
