package middleware

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
)

// AuthRequired returns the enhanced JWT middleware from jwt.go
func AuthRequired() gin.HandlerFunc {
	// This will be set up with database instance in routes
	return func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT Manager not initialized"})
		c.Abort()
	}
}

// RoleRequired ensures that the user belongs to one of the specified roles
func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		// Convert to lowercase for case-insensitive comparison
		roleStrLower := strings.ToLower(roleStr)
		
		for _, role := range roles {
			roleLower := strings.ToLower(role)
			if roleStrLower == roleLower {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

// AuthMiddleware is an alias for AuthRequired (backward compatibility)
func AuthMiddleware() gin.HandlerFunc {
	return AuthRequired()
}

// RequireRoles is an alias for RoleRequired (backward compatibility)
func RequireRoles(roles ...string) gin.HandlerFunc {
	return RoleRequired(roles...)
}
