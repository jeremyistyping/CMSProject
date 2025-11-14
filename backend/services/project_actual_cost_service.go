package services

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ProjectActualCostService provides read access to derived project_actual_costs view.
type ProjectActualCostService interface {
	GetActualCosts(projectID uint, startDate, endDate *time.Time) ([]models.ProjectActualCost, error)
}

type projectActualCostService struct {
	db *gorm.DB
}

// NewProjectActualCostService creates a new ProjectActualCostService
func NewProjectActualCostService(db *gorm.DB) ProjectActualCostService {
	return &projectActualCostService{db: db}
}

// GetActualCosts returns actual cost rows for a project within an optional date range.
func (s *projectActualCostService) GetActualCosts(projectID uint, startDate, endDate *time.Time) ([]models.ProjectActualCost, error) {
	var rows []models.ProjectActualCost

	if projectID == 0 {
		return nil, fmt.Errorf("project_id is required")
	}

	query := s.db.Where("project_id = ?", projectID)
	if startDate != nil {
		query = query.Where("date >= ?", startDate.Format("2006-01-02"))
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate.Format("2006-01-02"))
	}

	if err := query.Order("date ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to query project actual costs: %w", err)
	}

	return rows, nil
}
