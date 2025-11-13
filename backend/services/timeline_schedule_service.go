package services

import (
	"app-sistem-akuntansi/models"
	"errors"
	"strings"
	"gorm.io/gorm"
)

type TimelineScheduleService interface {
	GetSchedulesByProject(projectID uint) ([]models.TimelineSchedule, error)
	GetScheduleByID(projectID uint, scheduleID uint) (*models.TimelineSchedule, error)
	CreateSchedule(schedule *models.TimelineSchedule) error
	UpdateSchedule(schedule *models.TimelineSchedule) error
	DeleteSchedule(projectID uint, scheduleID uint) error
	UpdateScheduleStatus(projectID uint, scheduleID uint, status string) error
}

type timelineScheduleService struct {
	db *gorm.DB
}

// NewTimelineScheduleService creates a new timeline schedule service
func NewTimelineScheduleService(db *gorm.DB) TimelineScheduleService {
	return &timelineScheduleService{db: db}
}

// GetSchedulesByProject retrieves all timeline schedules for a project
func (s *timelineScheduleService) GetSchedulesByProject(projectID uint) ([]models.TimelineSchedule, error) {
	var schedules []models.TimelineSchedule
	result := s.db.Where("project_id = ?", projectID).
		Order("start_date ASC, created_at ASC").
		Find(&schedules)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return schedules, nil
}

// GetScheduleByID retrieves a single timeline schedule by ID
func (s *timelineScheduleService) GetScheduleByID(projectID uint, scheduleID uint) (*models.TimelineSchedule, error) {
	var schedule models.TimelineSchedule
	result := s.db.Where("id = ? AND project_id = ?", scheduleID, projectID).First(&schedule)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("schedule not found")
		}
		return nil, result.Error
	}
	
	return &schedule, nil
}

// CreateSchedule creates a new timeline schedule with validation
func (s *timelineScheduleService) CreateSchedule(schedule *models.TimelineSchedule) error {
	// Validate required fields
	if err := s.validateSchedule(schedule); err != nil {
		return err
	}
	
	// Check if project exists
	var project models.Project
	if err := s.db.First(&project, schedule.ProjectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("project not found")
		}
		return err
	}
	
	// Create the schedule
	result := s.db.Create(schedule)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// UpdateSchedule updates an existing timeline schedule with validation
func (s *timelineScheduleService) UpdateSchedule(schedule *models.TimelineSchedule) error {
	// Check if schedule exists
	var existing models.TimelineSchedule
	if err := s.db.First(&existing, schedule.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("schedule not found")
		}
		return err
	}
	
	// Verify it belongs to the correct project
	if existing.ProjectID != schedule.ProjectID {
		return errors.New("schedule does not belong to this project")
	}
	
	// Validate updated data
	if err := s.validateSchedule(schedule); err != nil {
		return err
	}
	
	// Update the schedule
	result := s.db.Model(&existing).Updates(schedule)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// DeleteSchedule deletes a timeline schedule (soft delete)
func (s *timelineScheduleService) DeleteSchedule(projectID uint, scheduleID uint) error {
	// Check if schedule exists and belongs to the project
	var schedule models.TimelineSchedule
	if err := s.db.Where("id = ? AND project_id = ?", scheduleID, projectID).First(&schedule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("schedule not found")
		}
		return err
	}
	
	// Soft delete
	result := s.db.Delete(&schedule)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// UpdateScheduleStatus updates the status of a timeline schedule
func (s *timelineScheduleService) UpdateScheduleStatus(projectID uint, scheduleID uint, status string) error {
	// Get schedule
	schedule, err := s.GetScheduleByID(projectID, scheduleID)
	if err != nil {
		return err
	}
	
	// Validate status
	validStatuses := []string{
		models.TimelineStatusNotStarted,
		models.TimelineStatusInProgress,
		models.TimelineStatusCompleted,
	}
	validStatus := false
	for _, s := range validStatuses {
		if status == s {
			validStatus = true
			break
		}
	}
	if !validStatus {
		return errors.New("invalid status value")
	}
	
	// Update status
	schedule.Status = status
	
	// Save
	if err := s.db.Save(schedule).Error; err != nil {
		return err
	}
	
	return nil
}

// validateSchedule validates timeline schedule data
func (s *timelineScheduleService) validateSchedule(schedule *models.TimelineSchedule) error {
	if schedule.ProjectID == 0 {
		return errors.New("project ID is required")
	}
	
	if strings.TrimSpace(schedule.WorkArea) == "" {
		return errors.New("work area is required")
	}
	
	if schedule.StartDate.IsZero() {
		return errors.New("start date is required")
	}
	
	if schedule.EndDate.IsZero() {
		return errors.New("end date is required")
	}
	
	if schedule.EndDate.Before(schedule.StartDate) {
		return errors.New("end date must be after or equal to start date")
	}
	
	// Validate status
	if schedule.Status != "" {
		validStatuses := []string{
			models.TimelineStatusNotStarted,
			models.TimelineStatusInProgress,
			models.TimelineStatusCompleted,
		}
		validStatus := false
		for _, s := range validStatuses {
			if schedule.Status == s {
				validStatus = true
				break
			}
		}
		if !validStatus {
			return errors.New("invalid status value")
		}
	}
	
	return nil
}

