package repositories

import (
	"app-sistem-akuntansi/models"
	"errors"

	"gorm.io/gorm"
)

type ProjectBudgetRepository interface {
	GetByProjectAndAccount(projectID, accountID uint) (*models.ProjectBudget, error)
	GetByProject(projectID uint) ([]models.ProjectBudget, error)
	Create(budget *models.ProjectBudget) error
	Update(budget *models.ProjectBudget) error
	Upsert(budget *models.ProjectBudget) error
	Delete(id uint) error
}

type projectBudgetRepository struct {
	db *gorm.DB
}

func NewProjectBudgetRepository(db *gorm.DB) ProjectBudgetRepository {
	return &projectBudgetRepository{db: db}
}

func (r *projectBudgetRepository) GetByProjectAndAccount(projectID, accountID uint) (*models.ProjectBudget, error) {
	var budget models.ProjectBudget
	err := r.db.Where("project_id = ? AND account_id = ?", projectID, accountID).
		First(&budget).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &budget, nil
}

func (r *projectBudgetRepository) GetByProject(projectID uint) ([]models.ProjectBudget, error) {
	var budgets []models.ProjectBudget
	err := r.db.Where("project_id = ?", projectID).
		Preload("Project").
		Find(&budgets).Error
	return budgets, err
}

func (r *projectBudgetRepository) Create(budget *models.ProjectBudget) error {
	return r.db.Create(budget).Error
}

func (r *projectBudgetRepository) Update(budget *models.ProjectBudget) error {
	return r.db.Save(budget).Error
}

func (r *projectBudgetRepository) Upsert(budget *models.ProjectBudget) error {
	// Check if exists
	existing, err := r.GetByProjectAndAccount(budget.ProjectID, budget.AccountID)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing
		budget.ID = existing.ID
		return r.Update(budget)
	}

	// Create new
	return r.Create(budget)
}

func (r *projectBudgetRepository) Delete(id uint) error {
	return r.db.Delete(&models.ProjectBudget{}, id).Error
}
