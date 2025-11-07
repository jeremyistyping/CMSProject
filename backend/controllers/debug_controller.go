package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type DebugController struct {}

func NewDebugController() *DebugController {
	return &DebugController{}
}

// TestJWTContext tests what's available in the JWT context
func (dc *DebugController) TestJWTContext(c *gin.Context) {
	// Get all context values set by JWT middleware
	userID, userIDExists := c.Get("user_id")
	username, usernameExists := c.Get("username")
	email, emailExists := c.Get("email")
	role, roleExists := c.Get("role")
	sessionID, sessionIDExists := c.Get("session_id")
	permissions, permissionsExists := c.Get("permissions")
	
	// Get the Authorization header too
	authHeader := c.GetHeader("Authorization")
	
	response := gin.H{
		"success": true,
		"message": "JWT context test successful",
		"context_data": gin.H{
			"user_id": gin.H{
				"value":  userID,
				"exists": userIDExists,
			},
			"username": gin.H{
				"value":  username,
				"exists": usernameExists,
			},
			"email": gin.H{
				"value":  email,
				"exists": emailExists,
			},
			"role": gin.H{
				"value":  role,
				"exists": roleExists,
			},
			"session_id": gin.H{
				"value":  sessionID,
				"exists": sessionIDExists,
			},
			"permissions": gin.H{
				"value":  permissions,
				"exists": permissionsExists,
			},
		},
		"auth_header": authHeader,
	}
	
	c.JSON(http.StatusOK, response)
}

// TestRolePermission tests role-based access
func (dc *DebugController) TestRolePermission(c *gin.Context) {
	role, roleExists := c.Get("role")
	
	if !roleExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Role not found in context",
			"code": "ROLE_NOT_FOUND",
		})
		return
	}
	
	roleStr := role.(string)
	
	// Check if user has finance role (required for payments)
	hasFinanceRole := false
	hasAdminRole := false
	
	if roleStr == "finance" {
		hasFinanceRole = true
	}
	if roleStr == "admin" {
		hasAdminRole = true
	}
	
	canAccessPayments := hasFinanceRole || hasAdminRole
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user_role": roleStr,
		"has_finance_role": hasFinanceRole,
		"has_admin_role": hasAdminRole,
		"can_access_payments": canAccessPayments,
		"message": "Role permission test completed",
	})
}

// TestCashBankPermission tests permission middleware for cash_bank
func (dc *DebugController) TestCashBankPermission(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "You have permission to view cash_bank!",
		"endpoint": "/debug/auth/test-cashbank-permission",
	})
}

// TestPaymentsPermission tests permission middleware for payments
func (dc *DebugController) TestPaymentsPermission(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "You have permission to view payments!",
		"endpoint": "/debug/auth/test-payments-permission",
	})
}
