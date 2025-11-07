package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService   *services.NotificationService
	stockMonitoringService *services.StockMonitoringService
}

func NewNotificationHandler(notificationService *services.NotificationService, stockMonitoringService *services.StockMonitoringService) *NotificationHandler {
	return &NotificationHandler{
		notificationService:    notificationService,
		stockMonitoringService: stockMonitoringService,
	}
}

// GetNotifications gets user notifications
// GET /api/notifications
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}

	// Proactively run minimum stock check so low-stock notifications appear without manual triggers
	h.runAutoStockCheck(c)
	
	page := 1
	limit := 20
	onlyUnread := false

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if ur := c.Query("unread_only"); ur == "true" {
		onlyUnread = true
	}

	notifications, total, err := h.notificationService.GetUserNotifications(userID, page, limit, onlyUnread)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total": total,
		"page": page,
		"limit": limit,
	})
}

// MarkNotificationAsRead marks notification as read
// PUT /api/notifications/:id/read
func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.notificationService.MarkAsRead(uint(notificationID), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

// MarkAllNotificationsAsRead marks all notifications as read
// PUT /api/notifications/read-all
func (h *NotificationHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	err = h.notificationService.MarkAllAsRead(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All notifications marked as read",
	})
}

// GetUnreadCount gets count of unread notifications
// GET /api/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}

	// Ensure stock alerts are created before counting
	h.runAutoStockCheck(c)
	
	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// DeleteNotification deletes a notification
// DELETE /api/notifications/:id
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.notificationService.DeleteNotification(uint(notificationID), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification deleted successfully",
	})
}

// GetNotificationsByType gets notifications by type
// GET /api/notifications/type/:type
func (h *NotificationHandler) GetNotificationsByType(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	notificationType := c.Param("type")

	// If requesting stock-related notifications, proactively run the stock check
	h.runAutoStockCheck(c)
	
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	notifications, total, err := h.notificationService.GetNotificationsByType(userID, notificationType, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total": total,
		"page": page,
		"limit": limit,
		"type": notificationType,
	})
}

// runAutoStockCheck checks min stock and resolves alerts so notifications appear without manual triggers
func (h *NotificationHandler) runAutoStockCheck(c *gin.Context) {
	// Only run for roles that should see stock alerts
	role, err := utils.GetUserRoleFromToken(c)
	if err != nil {
		return
	}
	switch strings.ToLower(role) {
	case "admin", "inventory_manager", "director":
		if h.stockMonitoringService != nil {
			_ = h.stockMonitoringService.CheckMinimumStock()
			_ = h.stockMonitoringService.ResolveStockAlerts()
		}
	}
}

// CreateNotification creates a new notification (admin only)
// POST /api/notifications
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	userRole, err := utils.GetUserRoleFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can create notifications"})
		return
	}

	var request struct {
		UserID   uint   `json:"user_id" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Title    string `json:"title" binding:"required"`
		Message  string `json:"message" binding:"required"`
		Priority string `json:"priority"`
		Data     string `json:"data"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification := &models.Notification{
		UserID:   request.UserID,
		Type:     request.Type,
		Title:    request.Title,
		Message:  request.Message,
		Priority: request.Priority,
		Data:     request.Data,
	}

	if notification.Priority == "" {
		notification.Priority = models.NotificationPriorityNormal
	}

	err = h.notificationService.CreateNotification(notification)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Notification created successfully",
		"notification": notification,
	})
}

// GetApprovalNotifications gets all approval-related notifications
// GET /api/notifications/approvals
func (h *NotificationHandler) GetApprovalNotifications(c *gin.Context) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing token"})
		return
	}
	
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Get approval notifications
	approvalTypes := []string{
		models.NotificationTypeApprovalPending,
		models.NotificationTypeApprovalApproved,
		models.NotificationTypeApprovalRejected,
	}

	allNotifications := []models.Notification{}
	totalCount := int64(0)

	for _, notifType := range approvalTypes {
		notifications, total, err := h.notificationService.GetNotificationsByType(userID, notifType, 1, 1000)
		if err == nil {
			allNotifications = append(allNotifications, notifications...)
			totalCount += total
		}
	}

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit
	if end > len(allNotifications) {
		end = len(allNotifications)
	}
	if start > len(allNotifications) {
		start = len(allNotifications)
	}

	paginatedNotifications := allNotifications[start:end]

	c.JSON(http.StatusOK, gin.H{
		"notifications": paginatedNotifications,
		"total": totalCount,
		"page": page,
		"limit": limit,
	})
}
