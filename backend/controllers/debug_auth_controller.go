package controllers

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
)

type DebugAuthController struct{}

func NewDebugAuthController() *DebugAuthController {
	return &DebugAuthController{}
}

// DebugAuthHeader logs and returns detailed auth header information
func (dac *DebugAuthController) DebugAuthHeader(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	
	// Get all headers for debugging
	allHeaders := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			allHeaders[key] = values[0]
		}
	}
	
	// Parse token if present
	var tokenInfo map[string]interface{}
	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		tokenInfo = map[string]interface{}{
			"has_bearer_prefix": strings.HasPrefix(authHeader, "Bearer "),
			"token_length":      len(tokenString),
			"token_preview":     tokenString[:min(20, len(tokenString))] + "...",
			"full_header":       authHeader,
		}
	}
	
	response := gin.H{
		"success": true,
		"debug_info": gin.H{
			"auth_header": authHeader,
			"token_info":  tokenInfo,
			"all_headers": allHeaders,
			"request_info": gin.H{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"ip":     c.ClientIP(),
			},
		},
	}
	
	c.JSON(http.StatusOK, response)
}

// DebugSessionValidation tests session validation
func (dac *DebugAuthController) DebugSessionValidation(c *gin.Context) {
	// This endpoint should be protected by JWT middleware
	// If we reach here, JWT validation passed
	
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	sessionID, _ := c.Get("session_id")
	permissions, _ := c.Get("permissions")
	
	response := gin.H{
		"success": true,
		"message": "JWT validation passed - this endpoint is protected",
		"user_info": gin.H{
			"user_id":    userID,
			"username":   username,
			"email":      email,
			"role":       role,
			"session_id": sessionID,
			"permissions": permissions,
		},
	}
	
	c.JSON(http.StatusOK, response)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
