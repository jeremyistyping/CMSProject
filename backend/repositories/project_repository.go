package repositories

import (
	"app-sistem-akuntansi/models"
	"errors"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	GetAll() ([]models.Project, error)
	GetByID(id uint) (*models.Project, error)
	Create(project *models.Project) error
	Update(project *models.Project) error
	Delete(id uint) error
	Archive(id uint) error
	GetByStatus(status string) ([]models.Project, error)
	GetActiveProjects() ([]models.Project, error)
}

type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// GetAll retrieves all projects (excluding soft deleted)
func (r *projectRepository) GetAll() ([]models.Project, error) {
	var projects []models.Project
	if err := r.db.Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// GetByID retrieves a single project by ID
func (r *projectRepository) GetByID(id uint) (*models.Project, error) {
	var project models.Project
	if err := r.db.First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	return &project, nil
}

// Create creates a new project
func (r *projectRepository) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

// Update updates an existing project
func (r *projectRepository) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

// Delete hard deletes a project (not recommended, use Archive instead)
func (r *projectRepository) Delete(id uint) error {
	return r.db.Delete(&models.Project{}, id).Error
}

// Archive soft deletes a project by setting status to archived
func (r *projectRepository) Archive(id uint) error {
	return r.db.Model(&models.Project{}).Where("id = ?", id).Update("status", models.ProjectStatusArchived).Error
}

// GetByStatus retrieves projects by status
func (r *projectRepository) GetByStatus(status string) ([]models.Project, error) {
	var projects []models.Project
	if err := r.db.Where("status = ?", status).Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// GetActiveProjects retrieves all active projects (not archived or deleted)
func (r *projectRepository) GetActiveProjects() ([]models.Project, error) {
	var projects []models.Project
	if err := r.db.Where("status != ?", models.ProjectStatusArchived).Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

