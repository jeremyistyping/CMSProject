package services

import (
	"encoding/json"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

type NotificationService struct {
	notificationRepo *repositories.NotificationRepository
	smartService     *SmartNotificationService
	db               *gorm.DB
}

func NewNotificationService(db *gorm.DB, notificationRepo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		smartService:     NewSmartNotificationService(db, notificationRepo),
		db:               db,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(notification *models.Notification) error {
	return s.notificationRepo.Create(notification)
}

// GetUserNotifications gets notifications for a user
func (s *NotificationService) GetUserNotifications(userID uint, page, limit int, onlyUnread bool) ([]models.Notification, int64, error) {
	return s.notificationRepo.GetUserNotifications(userID, page, limit, onlyUnread)
}

// GetNotificationsByType gets notifications by type for a user
func (s *NotificationService) GetNotificationsByType(userID uint, notificationType string, page, limit int) ([]models.Notification, int64, error) {
	return s.notificationRepo.GetNotificationsByType(userID, notificationType, page, limit)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID, userID uint) error {
	return s.notificationRepo.MarkAsRead(notificationID, userID)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID uint) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}

// GetUnreadCount gets count of unread notifications
func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	return s.notificationRepo.GetUnreadCount(userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(notificationID, userID uint) error {
	return s.notificationRepo.Delete(notificationID, userID)
}

// CreateApprovalNotification creates approval-related notifications
func (s *NotificationService) CreateApprovalNotification(userID uint, notificationType, title, message string, data interface{}) error {
	var dataString string
	if data != nil {
		dataBytes, err := json.Marshal(data)
		if err == nil {
			dataString = string(dataBytes)
		}
	}

	notification := &models.Notification{
		UserID:   userID,
		Type:     notificationType,
		Title:    title,
		Message:  message,
		Data:     dataString,
		Priority: s.getNotificationPriority(notificationType),
	}

	return s.CreateNotification(notification)
}

// CreatePurchaseRequestSubmissionNotification notifies when purchase request is submitted
func (s *NotificationService) CreatePurchaseRequestSubmissionNotification(pr *models.PurchaseRequest) error {
	return s.smartService.CreatePurchaseRequestNotification(pr, "SUBMITTED", nil)
}

// CreatePurchaseRequestApprovedNotification notifies when purchase request is approved
func (s *NotificationService) CreatePurchaseRequestApprovedNotification(pr *models.PurchaseRequest, approverID uint) error {
	return s.smartService.CreatePurchaseRequestNotification(pr, "APPROVED", map[string]interface{}{
		"approver_id": approverID,
	})
}

// CreatePurchaseRequestRejectedNotification notifies when purchase request is rejected
func (s *NotificationService) CreatePurchaseRequestRejectedNotification(pr *models.PurchaseRequest, approverID uint, reason string) error {
	return s.smartService.CreatePurchaseRequestNotification(pr, "REJECTED", map[string]interface{}{
		"approver_id": approverID,
		"reason":      reason,
	})
}

// SendBulkNotification sends notification to multiple users
func (s *NotificationService) SendBulkNotification(userIDs []uint, notificationType, title, message string, data interface{}) error {
	for _, userID := range userIDs {
		err := s.CreateApprovalNotification(userID, notificationType, title, message, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *NotificationService) getNotificationPriority(notificationType string) string {
	switch notificationType {
	case models.NotificationTypeApprovalPending:
		return models.NotificationPriorityHigh
	case models.NotificationTypeApprovalRejected:
		return models.NotificationPriorityHigh
	case models.NotificationTypeApprovalApproved:
		return models.NotificationPriorityNormal
	default:
		return models.NotificationPriorityNormal
	}
}
