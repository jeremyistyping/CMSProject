package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"errors"
	"strings"
	"time"
)

type ProjectService interface {
	GetAllProjects() ([]models.Project, error)
	GetProjectByID(id uint) (*models.Project, error)
	CreateProject(project *models.Project) error
	UpdateProject(project *models.Project) error
	DeleteProject(id uint) error
	ArchiveProject(id uint) error
	UpdateProgress(id uint, progressData map[string]float64) error
	GetProjectsByStatus(status string) ([]models.Project, error)
	GetActiveProjects() ([]models.Project, error)
}

type projectService struct {
	repo repositories.ProjectRepository
}

// NewProjectService creates a new project service
func NewProjectService(repo repositories.ProjectRepository) ProjectService {
	return &projectService{repo: repo}
}

// GetAllProjects retrieves all projects
func (s *projectService) GetAllProjects() ([]models.Project, error) {
	return s.repo.GetAll()
}

// GetProjectByID retrieves a project by ID
func (s *projectService) GetProjectByID(id uint) (*models.Project, error) {
	return s.repo.GetByID(id)
}

// CreateProject creates a new project with validation
func (s *projectService) CreateProject(project *models.Project) error {
	// Validate required fields
	if err := s.validateProject(project); err != nil {
		return err
	}
	
	// Set default status if not provided
	if project.Status == "" {
		project.Status = models.ProjectStatusActive
	}
	
	// Ensure progress values are within valid range (0-100)
	s.normalizeProgressValues(project)
	
	return s.repo.Create(project)
}

// UpdateProject updates an existing project with validation
func (s *projectService) UpdateProject(project *models.Project) error {
	// Check if project exists
	existing, err := s.repo.GetByID(project.ID)
	if err != nil {
		return err
	}
	
	// Validate updated data
	if err := s.validateProject(project); err != nil {
		return err
	}
	
	// Ensure progress values are within valid range
	s.normalizeProgressValues(project)
	
	// If project is being marked as completed, set completion date
	if project.Status == models.ProjectStatusCompleted && existing.Status != models.ProjectStatusCompleted {
		now := time.Now()
		project.CompletionDate = &now
	}
	
	return s.repo.Update(project)
}

// DeleteProject deletes a project (soft delete via GORM)
func (s *projectService) DeleteProject(id uint) error {
	// Check if project exists
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	
	return s.repo.Delete(id)
}

// ArchiveProject archives a project by changing status
func (s *projectService) ArchiveProject(id uint) error {
	// Check if project exists
	if _, err := s.repo.GetByID(id); err != nil {
		return err
	}
	
	return s.repo.Archive(id)
}

// UpdateProgress updates project progress values
func (s *projectService) UpdateProgress(id uint, progressData map[string]float64) error {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	
	// Update progress values if provided
	if val, ok := progressData["overall_progress"]; ok {
		project.OverallProgress = s.clampProgress(val)
	}
	if val, ok := progressData["foundation_progress"]; ok {
		project.FoundationProgress = s.clampProgress(val)
	}
	if val, ok := progressData["utilities_progress"]; ok {
		project.UtilitiesProgress = s.clampProgress(val)
	}
	if val, ok := progressData["interior_progress"]; ok {
		project.InteriorProgress = s.clampProgress(val)
	}
	if val, ok := progressData["equipment_progress"]; ok {
		project.EquipmentProgress = s.clampProgress(val)
	}
	
	// If overall progress reaches 100%, mark as completed
	if project.OverallProgress >= 100 && project.Status != models.ProjectStatusCompleted {
		project.Status = models.ProjectStatusCompleted
		now := time.Now()
		project.CompletionDate = &now
	}
	
	return s.repo.Update(project)
}

// GetProjectsByStatus retrieves projects by status
func (s *projectService) GetProjectsByStatus(status string) ([]models.Project, error) {
	return s.repo.GetByStatus(status)
}

// GetActiveProjects retrieves all active projects
func (s *projectService) GetActiveProjects() ([]models.Project, error) {
	return s.repo.GetActiveProjects()
}

// validateProject validates project data
func (s *projectService) validateProject(project *models.Project) error {
	if strings.TrimSpace(project.ProjectName) == "" {
		return errors.New("project name is required")
	}
	
	if strings.TrimSpace(project.Customer) == "" {
		return errors.New("customer is required")
	}
	
	if strings.TrimSpace(project.City) == "" {
		return errors.New("city is required")
	}
	
	if project.Budget < 0 {
		return errors.New("budget cannot be negative")
	}
	
	// Validate project type
	validTypes := []string{
		models.ProjectTypeNewBuild,
		models.ProjectTypeRenovation,
		models.ProjectTypeExpansion,
		models.ProjectTypeMaintenance,
	}
	validType := false
	for _, t := range validTypes {
		if project.ProjectType == t {
			validType = true
			break
		}
	}
	if !validType {
		return errors.New("invalid project type")
	}
	
	// Validate status if provided
	if project.Status != "" {
		validStatuses := []string{
			models.ProjectStatusActive,
			models.ProjectStatusCompleted,
			models.ProjectStatusOnHold,
			models.ProjectStatusArchived,
		}
		validStatus := false
		for _, st := range validStatuses {
			if project.Status == st {
				validStatus = true
				break
			}
		}
		if !validStatus {
			return errors.New("invalid project status")
		}
	}
	
	return nil
}

// normalizeProgressValues ensures all progress values are between 0 and 100
func (s *projectService) normalizeProgressValues(project *models.Project) {
	project.OverallProgress = s.clampProgress(project.OverallProgress)
	project.FoundationProgress = s.clampProgress(project.FoundationProgress)
	project.UtilitiesProgress = s.clampProgress(project.UtilitiesProgress)
	project.InteriorProgress = s.clampProgress(project.InteriorProgress)
	project.EquipmentProgress = s.clampProgress(project.EquipmentProgress)
}

// clampProgress ensures progress value is between 0 and 100
func (s *projectService) clampProgress(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

