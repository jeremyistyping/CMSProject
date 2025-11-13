package services

import (
	"app-sistem-akuntansi/models"
	"errors"
	"strings"
	"gorm.io/gorm"
)

type MilestoneService interface {
	GetMilestonesByProject(projectID uint) ([]models.Milestone, error)
	GetMilestoneByID(projectID uint, milestoneID uint) (*models.Milestone, error)
	CreateMilestone(milestone *models.Milestone) error
	UpdateMilestone(milestone *models.Milestone) error
	DeleteMilestone(projectID uint, milestoneID uint) error
	CompleteMilestone(projectID uint, milestoneID uint) error
}

type milestoneService struct {
	db *gorm.DB
}

// NewMilestoneService creates a new milestone service
func NewMilestoneService(db *gorm.DB) MilestoneService {
	return &milestoneService{db: db}
}

// GetMilestonesByProject retrieves all milestones for a project
func (s *milestoneService) GetMilestonesByProject(projectID uint) ([]models.Milestone, error) {
	var milestones []models.Milestone
	result := s.db.Where("project_id = ?", projectID).
		Order("target_date ASC").
		Find(&milestones)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return milestones, nil
}

// GetMilestoneByID retrieves a single milestone by ID
func (s *milestoneService) GetMilestoneByID(projectID uint, milestoneID uint) (*models.Milestone, error) {
	var milestone models.Milestone
	result := s.db.Where("id = ? AND project_id = ?", milestoneID, projectID).First(&milestone)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("milestone not found")
		}
		return nil, result.Error
	}
	
	return &milestone, nil
}

// CreateMilestone creates a new milestone with validation
func (s *milestoneService) CreateMilestone(milestone *models.Milestone) error {
	// Validate required fields
	if err := s.validateMilestone(milestone); err != nil {
		return err
	}
	
	// Check if project exists
	var project models.Project
	if err := s.db.First(&project, milestone.ProjectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("project not found")
		}
		return err
	}
	
	// Create the milestone
	result := s.db.Create(milestone)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// UpdateMilestone updates an existing milestone with validation
func (s *milestoneService) UpdateMilestone(milestone *models.Milestone) error {
	// Check if milestone exists
	var existing models.Milestone
	if err := s.db.First(&existing, milestone.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("milestone not found")
		}
		return err
	}
	
	// Verify it belongs to the correct project
	if existing.ProjectID != milestone.ProjectID {
		return errors.New("milestone does not belong to this project")
	}
	
	// Validate updated data
	if err := s.validateMilestone(milestone); err != nil {
		return err
	}
	
	// Update the milestone
	result := s.db.Model(&existing).Updates(milestone)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// DeleteMilestone deletes a milestone (soft delete)
func (s *milestoneService) DeleteMilestone(projectID uint, milestoneID uint) error {
	// Check if milestone exists and belongs to the project
	var milestone models.Milestone
	if err := s.db.Where("id = ? AND project_id = ?", milestoneID, projectID).First(&milestone).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("milestone not found")
		}
		return err
	}
	
	// Soft delete
	result := s.db.Delete(&milestone)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// CompleteMilestone marks a milestone as completed
func (s *milestoneService) CompleteMilestone(projectID uint, milestoneID uint) error {
	// Get milestone
	milestone, err := s.GetMilestoneByID(projectID, milestoneID)
	if err != nil {
		return err
	}
	
	// Mark as complete
	milestone.Complete()
	
	// Save
	if err := s.db.Save(milestone).Error; err != nil {
		return err
	}
	
	return nil
}

// validateMilestone validates milestone data
func (s *milestoneService) validateMilestone(milestone *models.Milestone) error {
	if milestone.ProjectID == 0 {
		return errors.New("project ID is required")
	}
	
	if strings.TrimSpace(milestone.Title) == "" {
		return errors.New("title is required")
	}
	
	if milestone.TargetDate.IsZero() {
		return errors.New("target date is required")
	}
	
	// Validate status
	validStatuses := []string{
		models.MilestoneStatusPending,
		models.MilestoneStatusInProgress,
		models.MilestoneStatusCompleted,
		models.MilestoneStatusDelayed,
	}
	validStatus := false
	for _, s := range validStatuses {
		if milestone.Status == s {
			validStatus = true
			break
		}
	}
	if !validStatus && milestone.Status != "" {
		return errors.New("invalid status value")
	}
	
	// Validate progress
	if milestone.Progress < 0 || milestone.Progress > 100 {
		return errors.New("progress must be between 0 and 100")
	}
	
	// Validate priority
	if milestone.Priority != "" {
		validPriorities := []string{
			models.MilestonePriorityLow,
			models.MilestonePriorityMedium,
			models.MilestonePriorityHigh,
		}
		validPriority := false
		for _, p := range validPriorities {
			if milestone.Priority == p {
				validPriority = true
				break
			}
		}
		if !validPriority {
			return errors.New("invalid priority value")
		}
	}
	
	return nil
}

