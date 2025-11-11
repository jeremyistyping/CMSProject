package services

import (
	"app-sistem-akuntansi/models"
	"errors"
	"strings"
	"gorm.io/gorm"
)

type DailyUpdateService interface {
	GetDailyUpdatesByProject(projectID uint) ([]models.DailyUpdate, error)
	GetDailyUpdateByID(projectID uint, updateID uint) (*models.DailyUpdate, error)
	CreateDailyUpdate(dailyUpdate *models.DailyUpdate) error
	UpdateDailyUpdate(dailyUpdate *models.DailyUpdate) error
	DeleteDailyUpdate(projectID uint, updateID uint) error
	GetDailyUpdatesByDateRange(projectID uint, startDate, endDate string) ([]models.DailyUpdate, error)
}

type dailyUpdateService struct {
	db *gorm.DB
}

// NewDailyUpdateService creates a new daily update service
func NewDailyUpdateService(db *gorm.DB) DailyUpdateService {
	return &dailyUpdateService{db: db}
}

// GetDailyUpdatesByProject retrieves all daily updates for a project
func (s *dailyUpdateService) GetDailyUpdatesByProject(projectID uint) ([]models.DailyUpdate, error) {
	var updates []models.DailyUpdate
	result := s.db.Where("project_id = ?", projectID).
		Order("date DESC").
		Find(&updates)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return updates, nil
}

// GetDailyUpdateByID retrieves a single daily update by ID
func (s *dailyUpdateService) GetDailyUpdateByID(projectID uint, updateID uint) (*models.DailyUpdate, error) {
	var update models.DailyUpdate
	result := s.db.Where("id = ? AND project_id = ?", updateID, projectID).First(&update)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("daily update not found")
		}
		return nil, result.Error
	}
	
	return &update, nil
}

// CreateDailyUpdate creates a new daily update with validation
func (s *dailyUpdateService) CreateDailyUpdate(dailyUpdate *models.DailyUpdate) error {
	// Validate required fields
	if err := s.validateDailyUpdate(dailyUpdate); err != nil {
		return err
	}
	
	// Check if project exists
	var project models.Project
	if err := s.db.First(&project, dailyUpdate.ProjectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("project not found")
		}
		return err
	}
	
	// Initialize photos array if nil
	if dailyUpdate.Photos == nil {
		dailyUpdate.Photos = []string{}
	}
	
	// Create the daily update
	result := s.db.Create(dailyUpdate)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// UpdateDailyUpdate updates an existing daily update with validation
func (s *dailyUpdateService) UpdateDailyUpdate(dailyUpdate *models.DailyUpdate) error {
	// Check if daily update exists
	var existing models.DailyUpdate
	if err := s.db.First(&existing, dailyUpdate.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("daily update not found")
		}
		return err
	}
	
	// Verify it belongs to the correct project
	if existing.ProjectID != dailyUpdate.ProjectID {
		return errors.New("daily update does not belong to this project")
	}
	
	// Validate updated data
	if err := s.validateDailyUpdate(dailyUpdate); err != nil {
		return err
	}
	
	// Initialize photos array if nil
	if dailyUpdate.Photos == nil {
		dailyUpdate.Photos = []string{}
	}
	
	// Update the daily update
	result := s.db.Model(&existing).Updates(dailyUpdate)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// DeleteDailyUpdate deletes a daily update (soft delete)
func (s *dailyUpdateService) DeleteDailyUpdate(projectID uint, updateID uint) error {
	// Check if daily update exists and belongs to the project
	var update models.DailyUpdate
	if err := s.db.Where("id = ? AND project_id = ?", updateID, projectID).First(&update).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("daily update not found")
		}
		return err
	}
	
	// Soft delete
	result := s.db.Delete(&update)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// GetDailyUpdatesByDateRange retrieves daily updates within a date range
func (s *dailyUpdateService) GetDailyUpdatesByDateRange(projectID uint, startDate, endDate string) ([]models.DailyUpdate, error) {
	var updates []models.DailyUpdate
	
	query := s.db.Where("project_id = ?", projectID)
	
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}
	
	result := query.Order("date DESC").Find(&updates)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return updates, nil
}

// validateDailyUpdate validates daily update data
func (s *dailyUpdateService) validateDailyUpdate(update *models.DailyUpdate) error {
	if update.ProjectID == 0 {
		return errors.New("project ID is required")
	}
	
	if update.Date.IsZero() {
		return errors.New("date is required")
	}
	
	if strings.TrimSpace(update.WorkDescription) == "" {
		return errors.New("work description is required")
	}
	
	if update.WorkersPresent < 0 {
		return errors.New("workers present cannot be negative")
	}
	
	// Validate weather value
	validWeathers := []string{
		models.WeatherSunny,
		models.WeatherCloudy,
		models.WeatherRainy,
		models.WeatherStormy,
		models.WeatherPartlyCloudy,
	}
	validWeather := false
	for _, w := range validWeathers {
		if update.Weather == w {
			validWeather = true
			break
		}
	}
	if !validWeather {
		return errors.New("invalid weather value")
	}
	
	if strings.TrimSpace(update.CreatedBy) == "" {
		return errors.New("created by is required")
	}
	
	return nil
}

