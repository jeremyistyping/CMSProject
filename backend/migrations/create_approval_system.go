package migrations

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// MigrateApprovalSystem creates approval system tables
func MigrateApprovalSystem(db *gorm.DB) error {
	// Auto migrate approval system models
	err := db.AutoMigrate(
		&models.ApprovalWorkflow{},
		&models.ApprovalStep{},
		&models.ApprovalRequest{},
		&models.ApprovalAction{},
		&models.ApprovalHistory{},
	)
	if err != nil {
		return err
	}

	// Add indexes for better performance
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_approval_requests_status ON approval_requests(status)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_approval_requests_entity ON approval_requests(entity_type, entity_id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_approval_actions_active ON approval_actions(is_active, status)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_approval_history_request ON approval_history(request_id, created_at)").Error; err != nil {
		return err
	}

	return nil
}

// SeedDefaultApprovalWorkflows creates default approval workflows
func SeedDefaultApprovalWorkflows(db *gorm.DB) error {
	// Check if workflows already exist
	var count int64
	db.Model(&models.ApprovalWorkflow{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	// Sales approval workflows
	salesWorkflows := []models.ApprovalWorkflow{
		{
			Name:            "Sales Approval - Small (< 10M)",
			Module:          models.ApprovalModuleSales,
			MinAmount:       0,
			MaxAmount:       10000000,
			RequireFinance:  true,
			RequireDirector: false,
			IsActive:        true,
			Steps: []models.ApprovalStep{
				{
					StepOrder:    1,
					StepName:     "Finance Review",
					ApproverRole: "finance",
					IsOptional:   false,
					TimeLimit:    24,
				},
			},
		},
		{
			Name:            "Sales Approval - Medium (10M - 50M)",
			Module:          models.ApprovalModuleSales,
			MinAmount:       10000000,
			MaxAmount:       50000000,
			RequireFinance:  true,
			RequireDirector: true,
			IsActive:        true,
			Steps: []models.ApprovalStep{
				{
					StepOrder:    1,
					StepName:     "Finance Review",
					ApproverRole: "finance",
					IsOptional:   false,
					TimeLimit:    24,
				},
				{
					StepOrder:    2,
					StepName:     "Director Approval",
					ApproverRole: "director",
					IsOptional:   false,
					TimeLimit:    48,
				},
			},
		},
		{
			Name:            "Sales Approval - Large (> 50M)",
			Module:          models.ApprovalModuleSales,
			MinAmount:       50000000,
			MaxAmount:       0, // No upper limit
			RequireFinance:  true,
			RequireDirector: true,
			IsActive:        true,
			Steps: []models.ApprovalStep{
				{
					StepOrder:    1,
					StepName:     "Finance Review",
					ApproverRole: "finance",
					IsOptional:   false,
					TimeLimit:    12,
				},
				{
					StepOrder:    2,
					StepName:     "Director Approval",
					ApproverRole: "director",
					IsOptional:   false,
					TimeLimit:    24,
				},
			},
		},
	}

	// Purchase approval workflows
	purchaseWorkflows := []models.ApprovalWorkflow{
		{
			Name:            "Purchase Approval - Small (< 5M)",
			Module:          models.ApprovalModulePurchase,
			MinAmount:       0,
			MaxAmount:       5000000,
			RequireFinance:  true,
			RequireDirector: false,
			IsActive:        true,
			Steps: []models.ApprovalStep{
				{
					StepOrder:    1,
					StepName:     "Finance Review",
					ApproverRole: "finance",
					IsOptional:   false,
					TimeLimit:    24,
				},
			},
		},
		{
			Name:            "Purchase Approval - Medium (5M - 25M)",
			Module:          models.ApprovalModulePurchase,
			MinAmount:       5000000,
			MaxAmount:       25000000,
			RequireFinance:  true,
			RequireDirector: true,
			IsActive:        true,
			Steps: []models.ApprovalStep{
				{
					StepOrder:    1,
					StepName:     "Finance Review",
					ApproverRole: "finance",
					IsOptional:   false,
					TimeLimit:    24,
				},
				{
					StepOrder:    2,
					StepName:     "Director Approval",
					ApproverRole: "director",
					IsOptional:   false,
					TimeLimit:    48,
				},
			},
		},
		{
			Name:            "Purchase Approval - Large (> 25M)",
			Module:          models.ApprovalModulePurchase,
			MinAmount:       25000000,
			MaxAmount:       0, // No upper limit
			RequireFinance:  true,
			RequireDirector: true,
			IsActive:        true,
			Steps: []models.ApprovalStep{
				{
					StepOrder:    1,
					StepName:     "Finance Review",
					ApproverRole: "finance",
					IsOptional:   false,
					TimeLimit:    12,
				},
				{
					StepOrder:    2,
					StepName:     "Director Approval",
					ApproverRole: "director",
					IsOptional:   false,
					TimeLimit:    24,
				},
			},
		},
	}

	// Create sales workflows
	for _, workflow := range salesWorkflows {
		if err := db.Create(&workflow).Error; err != nil {
			return err
		}
	}

	// Create purchase workflows
	for _, workflow := range purchaseWorkflows {
		if err := db.Create(&workflow).Error; err != nil {
			return err
		}
	}

	return nil
}
