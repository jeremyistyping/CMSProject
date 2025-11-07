package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
)

// ActivityLogController handles activity log related requests
type ActivityLogController struct{}

// NewActivityLogController creates a new activity log controller
func NewActivityLogController() *ActivityLogController {
	return &ActivityLogController{}
}

// GetActivityLogs retrieves activity logs with filters
// @Summary Get activity logs
// @Description Retrieve activity logs with optional filters (admin only)
// @Tags Activity Logs
// @Accept json
// @Produce json
// @Param user_id query int false "Filter by user ID"
// @Param username query string false "Filter by username"
// @Param role query string false "Filter by role"
// @Param method query string false "Filter by HTTP method"
// @Param path query string false "Filter by path"
// @Param resource query string false "Filter by resource"
// @Param status_code query int false "Filter by status code"
// @Param is_error query boolean false "Filter by error status"
// @Param ip_address query string false "Filter by IP address"
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param limit query int false "Limit results (default: 100)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} map[string]interface{} "Activity logs retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/activity-logs [get]
func (ctrl *ActivityLogController) GetActivityLogs(c *gin.Context) {
	// Build filter from query parameters
	filter := models.ActivityLogFilter{
		Username:  c.Query("username"),
		Role:      c.Query("role"),
		Method:    c.Query("method"),
		Path:      c.Query("path"),
		Resource:  c.Query("resource"),
		IPAddress: c.Query("ip_address"),
	}
	
	// Parse user_id
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uid := uint(userID)
			filter.UserID = &uid
		}
	}
	
	// Parse status_code
	if statusCodeStr := c.Query("status_code"); statusCodeStr != "" {
		if statusCode, err := strconv.Atoi(statusCodeStr); err == nil {
			filter.StatusCode = &statusCode
		}
	}
	
	// Parse is_error
	if isErrorStr := c.Query("is_error"); isErrorStr != "" {
		if isError, err := strconv.ParseBool(isErrorStr); err == nil {
			filter.IsError = &isError
		}
	}
	
	// Parse start_date
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}
	
	// Parse end_date
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}
	
	// Parse limit and offset
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 100 // Default limit
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}
	
	// Get activity logs
	logs, total, err := middleware.GetActivityLogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve activity logs",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"logs":   logs,
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

// GetMyActivityLogs retrieves activity logs for the current user
// @Summary Get my activity logs
// @Description Retrieve activity logs for the authenticated user
// @Tags Activity Logs
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param limit query int false "Limit results (default: 50)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} map[string]interface{} "Activity logs retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/activity-logs/me [get]
func (ctrl *ActivityLogController) GetMyActivityLogs(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}
	
	uid := userID.(uint)
	
	// Build filter
	filter := models.ActivityLogFilter{
		UserID: &uid,
	}
	
	// Parse dates
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}
	
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}
	
	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 50 // Default limit for personal logs
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}
	
	// Get activity logs
	logs, total, err := middleware.GetActivityLogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve activity logs",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"logs":   logs,
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

// GetActivitySummary retrieves activity summary
// @Summary Get activity summary
// @Description Get activity summary grouped by user and date (admin only)
// @Tags Activity Logs
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (RFC3339 format)"
// @Param end_date query string true "End date (RFC3339 format)"
// @Success 200 {object} map[string]interface{} "Activity summary retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid date parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/activity-logs/summary [get]
func (ctrl *ActivityLogController) GetActivitySummary(c *gin.Context) {
	// Parse dates
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	
	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_date and end_date are required",
		})
		return
	}
	
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid start_date format",
			"message": "Use RFC3339 format (e.g., 2024-01-01T00:00:00Z)",
		})
		return
	}
	
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid end_date format",
			"message": "Use RFC3339 format (e.g., 2024-12-31T23:59:59Z)",
		})
		return
	}
	
	// Get activity summary
	summary, err := middleware.GetActivitySummary(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve activity summary",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"summary":    summary,
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

// CleanupOldLogs removes old activity logs
// @Summary Cleanup old activity logs
// @Description Remove activity logs older than specified days (admin only)
// @Tags Activity Logs
// @Accept json
// @Produce json
// @Param days_to_keep query int true "Number of days to keep (default: 90)"
// @Success 200 {object} map[string]interface{} "Old logs cleaned up successfully"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/activity-logs/cleanup [post]
func (ctrl *ActivityLogController) CleanupOldLogs(c *gin.Context) {
	// Parse days_to_keep
	daysToKeep := 90 // Default: keep logs for 90 days
	if daysStr := c.Query("days_to_keep"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 {
			daysToKeep = days
		}
	}
	
	// Check if activity logger is initialized
	if middleware.GlobalActivityLogger == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Activity logger not initialized",
		})
		return
	}
	
	// Cleanup database logs
	deletedCount, err := middleware.GlobalActivityLogger.CleanupOldLogs(daysToKeep)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cleanup old logs",
			"message": err.Error(),
		})
		return
	}
	
	// Cleanup log files
	if err := middleware.GlobalActivityLogger.CleanupOldLogFiles(daysToKeep); err != nil {
		fmt.Printf("⚠️  Failed to cleanup old log files: %v\n", err)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Old activity logs cleaned up successfully",
		"data": gin.H{
			"deleted_count": deletedCount,
			"days_to_keep":  daysToKeep,
		},
	})
}

// GetActivityStats retrieves activity statistics
// @Summary Get activity statistics
// @Description Get activity statistics for dashboard (admin only)
// @Tags Activity Logs
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Activity statistics retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/activity-logs/stats [get]
func (ctrl *ActivityLogController) GetActivityStats(c *gin.Context) {
	// Get today's date range
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	// Get last 7 days
	last7Days := now.AddDate(0, 0, -7)
	
	// Get today's activity count
	todayFilter := models.ActivityLogFilter{
		StartDate: &startOfDay,
		EndDate:   &endOfDay,
	}
	_, todayCount, _ := middleware.GetActivityLogs(todayFilter)
	
	// Get last 7 days activity count
	weekFilter := models.ActivityLogFilter{
		StartDate: &last7Days,
		EndDate:   &now,
	}
	_, weekCount, _ := middleware.GetActivityLogs(weekFilter)
	
	// Get error count for today
	todayErrors := true
	todayErrorFilter := models.ActivityLogFilter{
		StartDate: &startOfDay,
		EndDate:   &endOfDay,
		IsError:   &todayErrors,
	}
	_, todayErrorCount, _ := middleware.GetActivityLogs(todayErrorFilter)
	
	// Get most active users today
	todaySummary, _ := middleware.GetActivitySummary(startOfDay, endOfDay)
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"today_total":       todayCount,
			"today_errors":      todayErrorCount,
			"last_7_days_total": weekCount,
			"most_active_today": todaySummary,
		},
	})
}
