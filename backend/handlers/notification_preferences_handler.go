package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationPreferencesHandler struct {
	smartService *services.SmartNotificationService
	db           *gorm.DB
}

func NewNotificationPreferencesHandler(db *gorm.DB, notificationRepo *repositories.NotificationRepository) *NotificationPreferencesHandler {
	return &NotificationPreferencesHandler{
		smartService: services.NewSmartNotificationService(db, notificationRepo),
		db:           db,
	}
}

// GetPreferences gets user notification preferences
// GET /api/notifications/preferences
func (h *NotificationPreferencesHandler) GetPreferences(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}

	preferences, err := h.smartService.GetUserNotificationPreferences(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"preferences": preferences,
	})
}

// UpdatePreferences updates user notification preferences
// PUT /api/notifications/preferences
func (h *NotificationPreferencesHandler) UpdatePreferences(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}

	var request struct {
		EmailEnabled          *bool   `json:"email_enabled"`
		InAppEnabled          *bool   `json:"in_app_enabled"`
		PushEnabled           *bool   `json:"push_enabled"`
		BatchNotifications    *bool   `json:"batch_notifications"`
		MaxDailyNotifications *int    `json:"max_daily_notifications"`
		QuietHoursStart       *string `json:"quiet_hours_start"`
		QuietHoursEnd         *string `json:"quiet_hours_end"`
		ApprovalPending       *bool   `json:"approval_pending"`
		ApprovalApproved      *bool   `json:"approval_approved"`
		ApprovalRejected      *bool   `json:"approval_rejected"`
		ApprovalEscalated     *bool   `json:"approval_escalated"`
		StockAlerts           *bool   `json:"stock_alerts"`
		PaymentReminders      *bool   `json:"payment_reminders"`
		SystemAlerts          *bool   `json:"system_alerts"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update map with only provided fields
	updates := make(map[string]interface{})
	
	if request.EmailEnabled != nil {
		updates["email_enabled"] = *request.EmailEnabled
	}
	if request.InAppEnabled != nil {
		updates["in_app_enabled"] = *request.InAppEnabled
	}
	if request.PushEnabled != nil {
		updates["push_enabled"] = *request.PushEnabled
	}
	if request.BatchNotifications != nil {
		updates["batch_notifications"] = *request.BatchNotifications
	}
	if request.MaxDailyNotifications != nil {
		updates["max_daily_notifications"] = *request.MaxDailyNotifications
	}
	if request.QuietHoursStart != nil {
		updates["quiet_hours_start"] = *request.QuietHoursStart
	}
	if request.QuietHoursEnd != nil {
		updates["quiet_hours_end"] = *request.QuietHoursEnd
	}
	if request.ApprovalPending != nil {
		updates["approval_pending"] = *request.ApprovalPending
	}
	if request.ApprovalApproved != nil {
		updates["approval_approved"] = *request.ApprovalApproved
	}
	if request.ApprovalRejected != nil {
		updates["approval_rejected"] = *request.ApprovalRejected
	}
	if request.ApprovalEscalated != nil {
		updates["approval_escalated"] = *request.ApprovalEscalated
	}
	if request.StockAlerts != nil {
		updates["stock_alerts"] = *request.StockAlerts
	}
	if request.PaymentReminders != nil {
		updates["payment_reminders"] = *request.PaymentReminders
	}
	if request.SystemAlerts != nil {
		updates["system_alerts"] = *request.SystemAlerts
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	err = h.smartService.UpdateUserNotificationPreferences(userID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated preferences
	preferences, _ := h.smartService.GetUserNotificationPreferences(userID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Preferences updated successfully",
		"preferences": preferences,
	})
}

// GetNotificationRules gets notification rules for current user's role
// GET /api/notifications/rules
func (h *NotificationPreferencesHandler) GetNotificationRules(c *gin.Context) {
	role, err := utils.GetUserRoleFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}

	var rules []models.NotificationRule
	err = h.db.Where("role = ? AND is_active = ?", role, true).Find(&rules).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Also get the notification matrix for this role
	matrix := models.GetNotificationMatrix()
	roleMatrix, exists := matrix[strings.ToLower(role)]
	
	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"matrix": roleMatrix,
		"matrix_exists": exists,
	})
}

// GetNotificationStats gets notification statistics
// GET /api/notifications/stats
func (h *NotificationPreferencesHandler) GetNotificationStats(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}

	// Get various statistics
	stats := make(map[string]interface{})

	// Total notifications today
	var todayCount int64
	today := time.Now().Truncate(24 * time.Hour)
	h.db.Model(&models.Notification{}).
		Where("user_id = ? AND created_at >= ?", userID, today).
		Count(&todayCount)
	stats["today_count"] = todayCount

	// Unread count
	var unreadCount int64
	h.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&unreadCount)
	stats["unread_count"] = unreadCount

	// Count by type
	var typeCounts []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	h.db.Model(&models.Notification{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Scan(&typeCounts)
	stats["by_type"] = typeCounts

	// Count by priority
	var priorityCounts []struct {
		Priority string `json:"priority"`
		Count    int64  `json:"count"`
	}
	h.db.Model(&models.Notification{}).
		Select("priority, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("priority").
		Scan(&priorityCounts)
	stats["by_priority"] = priorityCounts

	// Get preferences
	preferences, _ := h.smartService.GetUserNotificationPreferences(userID)
	if preferences != nil {
		stats["daily_limit"] = preferences.MaxDailyNotifications
		stats["remaining_today"] = preferences.MaxDailyNotifications - int(todayCount)
		stats["batch_enabled"] = preferences.BatchNotifications
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// TestNotification sends a test notification
// POST /api/notifications/test
func (h *NotificationPreferencesHandler) TestNotification(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}

	role, _ := utils.GetUserRoleFromToken(c)

	// Create a test notification
	notification := &models.Notification{
		UserID:   userID,
		Type:     "TEST_NOTIFICATION",
		Title:    "Test Notification",
		Message:  fmt.Sprintf("This is a test notification for role: %s", role),
		Priority: models.NotificationPriorityNormal,
	}

	err = h.db.Create(notification).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test notification sent successfully",
		"notification": notification,
	})
}
