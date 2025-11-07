package database

import (
	"fmt"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// MigrateNotificationConfig creates tables for smart notification system
func MigrateNotificationConfig(db *gorm.DB) error {
	// Create notification configuration tables
	err := db.AutoMigrate(
		&models.NotificationRule{},
		&models.NotificationPreference{},
		&models.NotificationBatch{},
		&models.NotificationQueue{},
	)
	
	if err != nil {
		return fmt.Errorf("failed to migrate notification config tables: %v", err)
	}

	fmt.Println("✅ Notification configuration tables migrated successfully")
	
	// Create default notification rules for each role
	if err := createDefaultNotificationRules(db); err != nil {
		return fmt.Errorf("failed to create default notification rules: %v", err)
	}
	
	return nil
}

func createDefaultNotificationRules(db *gorm.DB) error {
	// Check if rules already exist
	var count int64
	db.Model(&models.NotificationRule{}).Count(&count)
	if count > 0 {
		fmt.Println("Notification rules already exist, skipping creation")
		return nil
	}

	rules := []models.NotificationRule{
		// Employee rules
		{
			Name:        "Employee Own Purchase Status",
			Description: "Employees receive notifications about their own purchase status",
			Role:        "employee",
			MinAmount:   0,
			MaxAmount:   0,
			NotificationTypes: []string{
				models.NotificationTypeApprovalApproved,
				models.NotificationTypeApprovalRejected,
			},
			Priority: models.NotificationPriorityNormal,
			IsActive: true,
		},
		// Finance rules
		{
			Name:        "Finance Standard Approval",
			Description: "Finance receives approval requests for purchases up to 25M",
			Role:        "finance",
			MinAmount:   0,
			MaxAmount:   25000000,
			NotificationTypes: []string{
				models.NotificationTypeApprovalPending,
				models.NotificationTypeApprovalEscalated,
			},
			Priority: models.NotificationPriorityHigh,
			IsActive: true,
		},
		{
			Name:        "Finance Stock Alerts",
			Description: "Finance receives stock and inventory alerts",
			Role:        "finance",
			MinAmount:   0,
			MaxAmount:   0,
			NotificationTypes: []string{
				models.NotificationTypeLowStock,
				models.NotificationTypeStockOut,
			},
			Priority: models.NotificationPriorityNormal,
			IsActive: true,
		},
		{
			Name:        "Finance Payment Reminders",
			Description: "Finance receives payment due reminders",
			Role:        "finance",
			MinAmount:   0,
			MaxAmount:   0,
			NotificationTypes: []string{
				models.NotificationTypePaymentDue,
			},
			Priority: models.NotificationPriorityHigh,
			IsActive: true,
		},
		// Director rules
		{
			Name:        "Director High-Value Approval",
			Description: "Director receives approval requests for purchases above 25M",
			Role:        "director",
			MinAmount:   25000001,
			MaxAmount:   0,
			NotificationTypes: []string{
				models.NotificationTypeApprovalPending,
				models.NotificationTypeHighValuePurchase,
			},
			Priority: models.NotificationPriorityUrgent,
			IsActive: true,
		},
		{
			Name:        "Director Escalations",
			Description: "Director receives escalated approvals",
			Role:        "director",
			MinAmount:   0,
			MaxAmount:   0,
			NotificationTypes: []string{
				models.NotificationTypeApprovalEscalated,
			},
			Priority: models.NotificationPriorityUrgent,
			IsActive: true,
		},
		{
			Name:        "Director Critical Alerts",
			Description: "Director receives critical business alerts",
			Role:        "director",
			MinAmount:   0,
			MaxAmount:   0,
			NotificationTypes: []string{
				models.NotificationTypeCriticalAlert,
				models.NotificationTypeBudgetExceeded,
			},
			Priority: models.NotificationPriorityUrgent,
			IsActive: true,
		},
	}

	for _, rule := range rules {
		if err := db.Create(&rule).Error; err != nil {
			fmt.Printf("Failed to create rule %s: %v\n", rule.Name, err)
			continue
		}
		fmt.Printf("✅ Created notification rule: %s\n", rule.Name)
	}

	return nil
}
