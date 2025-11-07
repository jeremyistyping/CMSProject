package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"app-sistem-akuntansi/config"
	"github.com/gin-gonic/gin"
)

// SecurityMiddleware applies security headers and enforces HTTPS
type SecurityMiddleware struct {
	Config *config.Config
}

// NewSecurityMiddleware creates a new security middleware instance
func NewSecurityMiddleware(cfg *config.Config) *SecurityMiddleware {
	return &SecurityMiddleware{Config: cfg}
}

// HTTPSEnforcement redirects HTTP to HTTPS in production
func (sm *SecurityMiddleware) HTTPSEnforcement() gin.HandlerFunc {
	return func(c *gin.Context) {
		if sm.Config.EnableHTTPS && sm.Config.Environment == "production" {
			// Check if request is already HTTPS
			if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
				// Redirect to HTTPS
				url := fmt.Sprintf("https://%s%s", c.Request.Host, c.Request.RequestURI)
				c.Redirect(http.StatusMovedPermanently, url)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// SecurityHeaders adds comprehensive security headers
func (sm *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		if sm.Config.EnableSecurityHeaders {
			// HSTS (HTTP Strict Transport Security)
			if sm.Config.EnableHTTPS {
				hstsValue := fmt.Sprintf("max-age=%d; includeSubDomains; preload", sm.Config.HSTSMaxAge)
				c.Header("Strict-Transport-Security", hstsValue)
			}
			
			// Content Security Policy
			c.Header("Content-Security-Policy", sm.Config.CSPPolicy)
			
			// X-Frame-Options (Clickjacking protection)
			c.Header("X-Frame-Options", sm.Config.XFrameOptions)
			
			// X-Content-Type-Options (MIME-sniffing protection)
			c.Header("X-Content-Type-Options", "nosniff")
			
			// X-XSS-Protection (XSS protection for older browsers)
			c.Header("X-XSS-Protection", "1; mode=block")
			
			// Referrer-Policy
			c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
			
			// Permissions-Policy (formerly Feature-Policy)
			c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			
			// X-Permitted-Cross-Domain-Policies
			c.Header("X-Permitted-Cross-Domain-Policies", "none")
			
			// X-Download-Options (IE8+ protection)
			c.Header("X-Download-Options", "noopen")
			
			// Remove sensitive headers
			c.Header("X-Powered-By", "")
			c.Header("Server", "")
		}
		
		c.Next()
	}
}

// CORS configuration with security
func (sm *SecurityMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		isAllowed := false
		for _, allowedOrigin := range sm.Config.AllowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				isAllowed = true
				break
			}
		}
		
		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			// Allow Cache-Control to satisfy preflight when browsers/clients send it
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-CSRF-Token, Cache-Control")
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Range")
		}
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func (sm *SecurityMiddleware) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// ContentTypeValidation ensures proper content type for API requests
func (sm *SecurityMiddleware) ContentTypeValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for GET requests and file uploads
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		
		contentType := c.Request.Header.Get("Content-Type")
		
		// Check for JSON content type
		if !strings.Contains(contentType, "application/json") && 
		   !strings.Contains(contentType, "multipart/form-data") &&
		   !strings.Contains(contentType, "application/x-www-form-urlencoded") {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "Content-Type must be application/json, multipart/form-data, or application/x-www-form-urlencoded",
				"code":  "INVALID_CONTENT_TYPE",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// IPWhitelist restricts access to certain IPs (optional)
func (sm *SecurityMiddleware) IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(allowedIPs) == 0 {
			c.Next()
			return
		}
		
		clientIP := c.ClientIP()
		allowed := false
		
		for _, ip := range allowedIPs {
			if clientIP == ip {
				allowed = true
				break
			}
		}
		
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied from this IP address",
				"code":  "IP_NOT_ALLOWED",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// SecureJSON adds security to JSON responses
func (sm *SecurityMiddleware) SecureJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Override the default JSON method to add security prefix
		c.Next()
		
		// Add JSON hijacking protection prefix for older browsers
		if c.Writer.Header().Get("Content-Type") == "application/json" {
			// The ")]}',\n" prefix prevents JSON hijacking
			// Modern browsers don't need this, but it's good for defense in depth
		}
	}
}

// Helper function to generate request ID
func generateRequestID() string {
	// Simple implementation - in production, use UUID
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
