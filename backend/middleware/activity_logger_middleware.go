package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

// GlobalActivityLogger holds the global instance of activity logger service
var GlobalActivityLogger *services.ActivityLoggerService

// InitActivityLogger initializes the global activity logger
func InitActivityLogger(loggerService *services.ActivityLoggerService) {
	GlobalActivityLogger = loggerService
	fmt.Println("✅ Activity Logger initialized")
}

// ActivityLoggerMiddleware creates a middleware that logs all user activities
func ActivityLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health check and static files
		if shouldSkipLogging(c.Request.URL.Path) {
			c.Next()
			return
		}
		
		// Check if activity logger is initialized
		if GlobalActivityLogger == nil {
			c.Next()
			return
		}
		
		// Record start time
		startTime := time.Now()
		
		// Capture request body
		requestBody := captureRequestBody(c)
		
		// Create a custom response writer to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(startTime).Milliseconds()
		
		// Extract user information from context
		userID := extractUserID(c)
		username := extractUsername(c)
		role := extractRole(c)
		
		// Determine action and resource from path
		action, resource := determineActionAndResource(c.Request.Method, c.Request.URL.Path)
		
		// Capture response body, but skip binary content (PDF, images, etc.)
		responseBody := captureResponseBody(c, blw.body)
		
		// Create activity log entry
		activityLog := &models.ActivityLog{
			UserID:       userID,
			Username:     username,
			Role:         role,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			Action:       action,
			Resource:     resource,
			RequestBody:  requestBody,
			QueryParams:  c.Request.URL.RawQuery,
			StatusCode:   c.Writer.Status(),
			ResponseBody: responseBody,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Duration:     duration,
			IsError:      c.Writer.Status() >= 400,
			ErrorMessage: extractErrorMessage(c),
			Description:  generateDescription(c.Request.Method, action, resource, c.Writer.Status()),
			Metadata:     "{}", // Set valid empty JSON object for JSONB field
			CreatedAt:    time.Now(),
		}
		
		// Log activity asynchronously to avoid blocking the request
		go func() {
			if err := GlobalActivityLogger.LogActivity(activityLog); err != nil {
				fmt.Printf("⚠️  Failed to log activity: %v\n", err)
			}
		}()
	}
}

// bodyLogWriter is a custom response writer that captures the response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// shouldSkipLogging checks if the request path should skip logging
func shouldSkipLogging(path string) bool {
	skipPaths := []string{
		"/health",
		"/favicon.ico",
		"/api/v1/health",
		"/uploads/",
		"/templates/",
		"/swagger/",
		"/api/v1/notifications", // Skip automatic notification polling
	}
	
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	
	return false
}

// captureResponseBody captures the response body, but skips binary content
func captureResponseBody(c *gin.Context, body *bytes.Buffer) string {
	contentType := c.Writer.Header().Get("Content-Type")
	
	// Skip binary content types (PDF, images, Excel, ZIP, etc.)
	binaryTypes := []string{
		"application/pdf",
		"application/octet-stream",
		"image/",
		"video/",
		"audio/",
		"application/zip",
		"application/x-zip",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument",
	}
	
	for _, binaryType := range binaryTypes {
		if strings.Contains(contentType, binaryType) {
			// Return placeholder for binary content
			return fmt.Sprintf("[Binary content: %s, size: %d bytes]", contentType, body.Len())
		}
	}
	
	// For text/JSON content, log it (with size limit)
	responseStr := body.String()
	
	// Skip if response is too large (> 10KB)
	if len(responseStr) > 10*1024 {
		return fmt.Sprintf("[Response too large: %d bytes]", len(responseStr))
	}
	
	// Truncate to reasonable size
	return truncateString(responseStr, 1000)
}

// captureRequestBody captures the request body
func captureRequestBody(c *gin.Context) string {
	// Skip for GET requests
	if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
		return ""
	}
	
	// Skip for large requests
	if c.Request.ContentLength > 10*1024 { // Skip if larger than 10KB
		return "[Request body too large to log]"
	}
	
	// Skip for file uploads
	contentType := c.Request.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		return "[File upload - not logged]"
	}
	
	// Read request body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "[Failed to read request body]"
	}
	
	// Restore request body for the actual handler
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	
	// Sanitize sensitive data
	return sanitizeRequestBody(string(bodyBytes))
}

// sanitizeRequestBody removes sensitive information from request body
func sanitizeRequestBody(body string) string {
	// Try to parse as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(body), &jsonData); err == nil {
		// Remove sensitive fields
		sensitiveFields := []string{"password", "token", "refresh_token", "secret", "api_key", "credit_card"}
		for _, field := range sensitiveFields {
			if _, exists := jsonData[field]; exists {
				jsonData[field] = "[REDACTED]"
			}
		}
		
		// Convert back to JSON
		sanitized, err := json.Marshal(jsonData)
		if err == nil {
			return string(sanitized)
		}
	}
	
	// If not JSON or parsing failed, return truncated body
	return truncateString(body, 500)
}

// extractUserID extracts user ID from context, returns nil for anonymous users
func extractUserID(c *gin.Context) *uint {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			if id > 0 {
				return &id
			}
		}
		if id, ok := userID.(float64); ok {
			if id > 0 {
				uid := uint(id)
				return &uid
			}
		}
	}
	return nil
}

// extractUsername extracts username from context
func extractUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	
	// Fallback: if no username, return "anonymous"
	userID := extractUserID(c)
	if userID == nil {
		return "anonymous"
	}
	
	return "user_" + fmt.Sprint(*userID)
}

// extractRole extracts user role from context
func extractRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		if r, ok := role.(string); ok {
			return r
		}
	}
	return "guest"
}

// extractErrorMessage extracts error message from context or response
func extractErrorMessage(c *gin.Context) string {
	// Check if there's an error in the context
	if err, exists := c.Get("error"); exists {
		if errStr, ok := err.(string); ok {
			return errStr
		}
		if errObj, ok := err.(error); ok {
			return errObj.Error()
		}
	}
	
	// Check for errors array
	if len(c.Errors) > 0 {
		return c.Errors.String()
	}
	
	return ""
}

// determineActionAndResource determines the action and resource from the request
func determineActionAndResource(method, path string) (string, string) {
	// Remove /api/v1 prefix
	path = strings.TrimPrefix(path, "/api/v1")
	path = strings.TrimPrefix(path, "/api")
	path = strings.TrimPrefix(path, "/")
	
	// Split path into segments
	segments := strings.Split(path, "/")
	if len(segments) == 0 {
		return method, "unknown"
	}
	
	resource := segments[0]
	action := ""
	
	// Determine action based on method and path
	switch method {
	case "GET":
		if len(segments) > 1 && segments[1] != "" {
			// Specific resource (e.g., /products/123)
			action = "view_" + resource
		} else {
			// List resources (e.g., /products)
			action = "list_" + resource
		}
	case "POST":
		if len(segments) > 1 {
			// Action on specific resource (e.g., /sales/123/confirm)
			action = segments[len(segments)-1] + "_" + resource
		} else {
			// Create resource
			action = "create_" + resource
		}
	case "PUT", "PATCH":
		action = "update_" + resource
	case "DELETE":
		action = "delete_" + resource
	default:
		action = strings.ToLower(method) + "_" + resource
	}
	
	// Special cases
	if strings.Contains(path, "login") {
		action = "login"
		resource = "auth"
	} else if strings.Contains(path, "logout") {
		action = "logout"
		resource = "auth"
	} else if strings.Contains(path, "register") {
		action = "register"
		resource = "auth"
	}
	
	return action, resource
}

// generateDescription generates a human-readable description
func generateDescription(method, action, resource string, statusCode int) string {
	var desc string
	
	switch {
	case statusCode >= 500:
		desc = fmt.Sprintf("Server error during %s on %s", action, resource)
	case statusCode >= 400:
		desc = fmt.Sprintf("Failed %s on %s", action, resource)
	case statusCode >= 300:
		desc = fmt.Sprintf("Redirected during %s on %s", action, resource)
	case statusCode >= 200:
		desc = fmt.Sprintf("Successfully performed %s on %s", action, resource)
	default:
		desc = fmt.Sprintf("Performed %s on %s", action, resource)
	}
	
	return desc
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetActivityLogs is a helper function to get activity logs (for use in controllers)
func GetActivityLogs(filter models.ActivityLogFilter) ([]models.ActivityLog, int64, error) {
	if GlobalActivityLogger == nil {
		return nil, 0, fmt.Errorf("activity logger not initialized")
	}
	return GlobalActivityLogger.GetActivityLogs(filter)
}

// GetActivitySummary is a helper function to get activity summary
func GetActivitySummary(startDate, endDate time.Time) ([]models.ActivityLogSummary, error) {
	if GlobalActivityLogger == nil {
		return nil, fmt.Errorf("activity logger not initialized")
	}
	return GlobalActivityLogger.GetActivitySummary(startDate, endDate)
}
