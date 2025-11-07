package controllers

import (
	"net/http"
	"strconv"

	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
)

type MonitoringController struct{}

func NewMonitoringController() *MonitoringController {
	return &MonitoringController{}
}

// GetRateLimitStatus returns current rate limit status for the requesting client
func (mc *MonitoringController) GetRateLimitStatus(c *gin.Context) {
	status := middleware.GetRateLimitStatus(c)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
		"message": "Rate limit status retrieved successfully",
	})
}

// GetAuditLogs returns paginated audit logs
func (mc *MonitoringController) GetAuditLogs(c *gin.Context) {
	if middleware.GlobalAuditLogger == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Audit logging is not enabled",
		})
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	resource := c.Query("resource")

	offset := (page - 1) * limit

	// Convert userID to *uint (nil if 0)
	var userIDPtr *uint
	if userID > 0 {
		uid := uint(userID)
		userIDPtr = &uid
	}

	logs, total, err := middleware.GlobalAuditLogger.GetAuditLogs(userIDPtr, resource, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve audit logs",
			"details": err.Error(),
		})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":        logs,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
		"message": "Audit logs retrieved successfully",
	})
}

// GetTokenStats returns token usage statistics
func (mc *MonitoringController) GetTokenStats(c *gin.Context) {
	if middleware.GlobalTokenMonitor == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Token monitoring is not enabled",
		})
		return
	}

	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	stats, err := middleware.GlobalTokenMonitor.GetTokenStats(uint(userID), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve token statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id": userID,
			"days":    days,
			"stats":   stats,
		},
		"message": "Token statistics retrieved successfully",
	})
}

// GetRecentRefreshEvents returns recent token refresh events
func (mc *MonitoringController) GetRecentRefreshEvents(c *gin.Context) {
	if middleware.GlobalTokenMonitor == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Token monitoring is not enabled",
		})
		return
	}

	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	events, err := middleware.GlobalTokenMonitor.GetRecentRefreshEvents(uint(userID), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve refresh events",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id": userID,
			"limit":   limit,
			"events":  events,
		},
		"message": "Token refresh events retrieved successfully",
	})
}

// GetSystemSecurityStatus returns overall system security status
func (mc *MonitoringController) GetSystemSecurityStatus(c *gin.Context) {
	status := gin.H{
		"rate_limiting": gin.H{
			"enabled":             true,
			"payment_limit":       100,
			"auth_limit":         10,
			"general_limit":      200,
			"window_minutes":     1,
		},
		"audit_logging": gin.H{
			"enabled":     middleware.GlobalAuditLogger != nil,
			"log_to_file": true,
			"log_to_db":   true,
		},
		"token_monitoring": gin.H{
			"enabled":              middleware.GlobalTokenMonitor != nil,
			"refresh_monitoring":   true,
			"anomaly_detection":    true,
			"suspicious_tracking":  true,
		},
		"jwt_security": gin.H{
			"access_token_expiry":  "15 minutes",
			"refresh_token_expiry": "7 days",
			"session_tracking":     true,
			"token_blacklisting":   true,
		},
	}

	// Add current rate limit status
	if middleware.GlobalAuditLogger != nil || middleware.GlobalTokenMonitor != nil {
		rateLimitStatus := middleware.GetRateLimitStatus(c)
		status["current_rate_limits"] = rateLimitStatus
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
		"message": "System security status retrieved successfully",
	})
}

// GetUserSecuritySummary returns security summary for a specific user
func (mc *MonitoringController) GetUserSecuritySummary(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	summary := gin.H{
		"user_id": userID,
	}

	// Get audit summary if available
	if middleware.GlobalAuditLogger != nil {
		// Convert userID to *uint (nil if 0)
		var userIDPtr *uint
		if userID > 0 {
			uid := uint(userID)
			userIDPtr = &uid
		}
		auditLogs, total, err := middleware.GlobalAuditLogger.GetAuditLogs(userIDPtr, "", 10, 0)
		if err == nil {
			summary["recent_activities"] = gin.H{
				"total_logs":     total,
				"recent_actions": auditLogs,
			}
		}
	}

	// Get token stats if available
	if middleware.GlobalTokenMonitor != nil {
		tokenStats, err := middleware.GlobalTokenMonitor.GetTokenStats(uint(userID), 7)
		if err == nil {
			summary["token_usage"] = gin.H{
				"last_7_days": tokenStats,
			}
		}

		refreshEvents, err := middleware.GlobalTokenMonitor.GetRecentRefreshEvents(uint(userID), 20)
		if err == nil {
			summary["recent_refreshes"] = refreshEvents
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
		"message": "User security summary retrieved successfully",
	})
}

// GetSecurityAlerts returns recent security alerts
func (mc *MonitoringController) GetSecurityAlerts(c *gin.Context) {
	// This would typically read from security log files
	// For now, return a placeholder
	
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	
	alerts := []gin.H{
		{
			"type":      "INFO",
			"message":   "Security monitoring system is active",
			"timestamp": "2024-01-01T00:00:00Z",
		},
	}

	// In a real implementation, you would parse log files for security alerts
	// and return actual security events

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"alerts": alerts,
			"limit":  limit,
			"total":  len(alerts),
		},
		"message": "Security alerts retrieved successfully",
	})
}
