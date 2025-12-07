package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardController struct {
	DB                       *gorm.DB
	dashboardService         *services.DashboardService
	employeeDashboardService *services.EmployeeDashboardService
}

func NewDashboardController(db *gorm.DB, _ interface{}) *DashboardController {
	return &DashboardController{
		DB:                       db,
		dashboardService:         services.NewDashboardService(db),
		employeeDashboardService: services.NewEmployeeDashboardService(db),
	}
}

// GetAnalytics returns dashboard analytics data
func (dc *DashboardController) GetAnalytics(c *gin.Context) {
	analytics, err := dc.dashboardService.GetDashboardAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetEmployeeDashboardData returns dashboard data for employees
func (dc *DashboardController) GetEmployeeDashboardData(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	userRole := c.GetString("user_role")

	data, err := dc.employeeDashboardService.GetEmployeeDashboardData(userIDUint, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetEmployeeApprovalWorkflows returns approval workflows for employees
func (dc *DashboardController) GetEmployeeApprovalWorkflows(c *gin.Context) {
	userRole := c.GetString("user_role")

	workflows, err := dc.employeeDashboardService.GetEmployeeApprovalWorkflows(userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    workflows,
	})
}

// GetEmployeePurchaseRequests returns purchase requests for employees
func (dc *DashboardController) GetEmployeePurchaseRequests(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	requests, err := dc.employeeDashboardService.GetPurchaseRequestsForEmployee(userIDUint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    requests,
	})
}

// GetEmployeeApprovalNotifications returns approval notifications for employees
func (dc *DashboardController) GetEmployeeApprovalNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	notifications, err := dc.employeeDashboardService.GetApprovalNotifications(userIDUint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    notifications,
	})
}

// GetEmployeePurchaseApprovalStatus returns purchase approval status for employees
func (dc *DashboardController) GetEmployeePurchaseApprovalStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	status, err := dc.employeeDashboardService.GetPurchaseApprovalStatus(userIDUint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// GetEmployeeNotificationsSummary returns notifications summary for employees
func (dc *DashboardController) GetEmployeeNotificationsSummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get recent notifications as summary
	notifications, err := dc.employeeDashboardService.GetRecentNotifications(userIDUint, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    notifications,
	})
}

// MarkNotificationAsRead marks a notification as read
func (dc *DashboardController) MarkNotificationAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	// Update notification
	now := time.Now()
	result := dc.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userIDUint).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification marked as read",
	})
}

// GetQuickStats returns quick statistics for dashboard widgets
func (dc *DashboardController) GetQuickStats(c *gin.Context) {
	userRole := c.GetString("user_role")
	userRoleLower := strings.ToLower(userRole)

	// Only for authorized roles
	if userRoleLower != "admin" && userRoleLower != "director" && userRoleLower != "finance" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view quick stats"})
		return
	}

	stats, err := dc.dashboardService.GetQuickStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
