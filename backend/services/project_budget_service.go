package services

import (
	"app-sistem-akuntansi/models"
	"errors"

	"gorm.io/gorm"
)

// ProjectBudgetService defines operations for managing project-level COA budgets
// This service works directly with the project_budgets table created via SQL migration.
type ProjectBudgetService interface {
	GetBudgetsByProject(projectID uint) ([]models.ProjectBudget, error)
	UpsertBudgets(projectID uint, items []models.ProjectBudgetUpsertRequest) error
	DeleteBudget(projectID uint, budgetID uint) error
}

type projectBudgetService struct {
	db *gorm.DB
}

// NewProjectBudgetService creates a new ProjectBudgetService
func NewProjectBudgetService(db *gorm.DB) ProjectBudgetService {
	return &projectBudgetService{db: db}
}

// GetBudgetsByProject returns all active budget rows for a project
func (s *projectBudgetService) GetBudgetsByProject(projectID uint) ([]models.ProjectBudget, error) {
	var budgets []models.ProjectBudget
	if err := s.db.
		Where("project_id = ?", projectID).
		Order("account_id").
		Find(&budgets).Error; err != nil {
		return nil, err
	}
	return budgets, nil
}

// UpsertBudgets creates or updates budget rows per (project_id, account_id)
// If a row exists, its estimated_amount is updated; otherwise a new row is inserted.
func (s *projectBudgetService) UpsertBudgets(projectID uint, items []models.ProjectBudgetUpsertRequest) error {
	if projectID == 0 {
		return errors.New("project_id is required")
	}
	if len(items) == 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if item.AccountID == 0 {
				return errors.New("account_id is required")
			}
			if item.EstimatedAmount < 0 {
				return errors.New("estimated_amount cannot be negative")
			}

			var budget models.ProjectBudget
			err := tx.
				Where("project_id = ? AND account_id = ? AND deleted_at IS NULL", projectID, item.AccountID).
				First(&budget).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new
				budget = models.ProjectBudget{
					ProjectID:       projectID,
					AccountID:       item.AccountID,
					EstimatedAmount: item.EstimatedAmount,
				}
				if err := tx.Create(&budget).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				// Update existing
				budget.EstimatedAmount = item.EstimatedAmount
				if err := tx.Save(&budget).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// DeleteBudget performs a soft delete of a budget row, scoped by project for safety.
func (s *projectBudgetService) DeleteBudget(projectID uint, budgetID uint) error {
	if projectID == 0 || budgetID == 0 {
		return errors.New("project_id and budget_id are required")
	}

	var budget models.ProjectBudget
	if err := s.db.
		Where("id = ? AND project_id = ?", budgetID, projectID).
		First(&budget).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("project budget not found")
		}
		return err
	}

	return s.db.Delete(&budget).Error
}

