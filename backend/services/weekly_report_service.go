package services

import (
	"app-sistem-akuntansi/models"
	"errors"
	"fmt"
	"time"
	"gorm.io/gorm"
)

type WeeklyReportService interface {
	GetWeeklyReportsByProject(projectID uint) ([]models.WeeklyReportDTO, error)
	GetWeeklyReportByID(projectID uint, reportID uint) (*models.WeeklyReportDTO, error)
	CreateWeeklyReport(request *models.WeeklyReportCreateRequest) (*models.WeeklyReportDTO, error)
	UpdateWeeklyReport(reportID uint, request *models.WeeklyReportUpdateRequest) (*models.WeeklyReportDTO, error)
	DeleteWeeklyReport(projectID uint, reportID uint) error
	GetWeeklyReportsByYear(projectID uint, year int) ([]models.WeeklyReportDTO, error)
	CheckReportExists(projectID uint, week int, year int) (bool, error)
}

type weeklyReportService struct {
	db *gorm.DB
}

// NewWeeklyReportService creates a new weekly report service
func NewWeeklyReportService(db *gorm.DB) WeeklyReportService {
	return &weeklyReportService{db: db}
}

// GetWeeklyReportsByProject retrieves all weekly reports for a project
func (s *weeklyReportService) GetWeeklyReportsByProject(projectID uint) ([]models.WeeklyReportDTO, error) {
	var reports []models.WeeklyReport
	result := s.db.Preload("Project").
		Where("project_id = ?", projectID).
		Order("year DESC, week DESC").
		Find(&reports)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return s.convertToDTO(reports), nil
}

// GetWeeklyReportByID retrieves a single weekly report by ID
func (s *weeklyReportService) GetWeeklyReportByID(projectID uint, reportID uint) (*models.WeeklyReportDTO, error) {
	var report models.WeeklyReport
	result := s.db.Preload("Project").
		Where("id = ? AND project_id = ?", reportID, projectID).
		First(&report)
	
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("weekly report not found")
		}
		return nil, result.Error
	}
	
	dto := s.convertSingleToDTO(&report)
	return &dto, nil
}

// CreateWeeklyReport creates a new weekly report with validation
func (s *weeklyReportService) CreateWeeklyReport(request *models.WeeklyReportCreateRequest) (*models.WeeklyReportDTO, error) {
	// Validate required fields
	if err := s.validateCreateRequest(request); err != nil {
		return nil, err
	}
	
	// Check if project exists
	var project models.Project
	if err := s.db.First(&project, request.ProjectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	
	// Check if report already exists for this week/year
	exists, err := s.CheckReportExists(request.ProjectID, request.Week, request.Year)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("weekly report already exists for week %d, %d", request.Week, request.Year)
	}
	
	// Create the weekly report
	report := &models.WeeklyReport{
		ProjectID:          request.ProjectID,
		Week:               request.Week,
		Year:               request.Year,
		ProjectManager:     request.ProjectManager,
		TotalWorkDays:      request.TotalWorkDays,
		WeatherDelays:      request.WeatherDelays,
		TeamSize:           request.TeamSize,
		Accomplishments:    request.Accomplishments,
		Challenges:         request.Challenges,
		NextWeekPriorities: request.NextWeekPriorities,
		GeneratedDate:      time.Now(),
		CreatedBy:          request.CreatedBy,
	}
	
	if err := report.Validate(); err != nil {
		return nil, err
	}
	
	result := s.db.Create(report)
	if result.Error != nil {
		return nil, result.Error
	}
	
	// Load the project relation
	s.db.Preload("Project").First(report, report.ID)
	
	dto := s.convertSingleToDTO(report)
	return &dto, nil
}

// UpdateWeeklyReport updates an existing weekly report
func (s *weeklyReportService) UpdateWeeklyReport(reportID uint, request *models.WeeklyReportUpdateRequest) (*models.WeeklyReportDTO, error) {
	// Check if weekly report exists
	var existing models.WeeklyReport
	if err := s.db.First(&existing, reportID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("weekly report not found")
		}
		return nil, err
	}
	
	// Apply updates
	updateMap := make(map[string]interface{})
	
	if request.ProjectManager != nil {
		updateMap["project_manager"] = *request.ProjectManager
	}
	if request.TotalWorkDays != nil {
		if *request.TotalWorkDays < 0 {
			return nil, errors.New("total work days cannot be negative")
		}
		updateMap["total_work_days"] = *request.TotalWorkDays
	}
	if request.WeatherDelays != nil {
		if *request.WeatherDelays < 0 {
			return nil, errors.New("weather delays cannot be negative")
		}
		updateMap["weather_delays"] = *request.WeatherDelays
	}
	if request.TeamSize != nil {
		if *request.TeamSize < 0 {
			return nil, errors.New("team size cannot be negative")
		}
		updateMap["team_size"] = *request.TeamSize
	}
	if request.Accomplishments != nil {
		updateMap["accomplishments"] = *request.Accomplishments
	}
	if request.Challenges != nil {
		updateMap["challenges"] = *request.Challenges
	}
	if request.NextWeekPriorities != nil {
		updateMap["next_week_priorities"] = *request.NextWeekPriorities
	}
	
	// Update the weekly report
	if len(updateMap) > 0 {
		result := s.db.Model(&existing).Updates(updateMap)
		if result.Error != nil {
			return nil, result.Error
		}
	}
	
	// Reload the report with project
	s.db.Preload("Project").First(&existing, reportID)
	
	dto := s.convertSingleToDTO(&existing)
	return &dto, nil
}

// DeleteWeeklyReport deletes a weekly report (soft delete)
func (s *weeklyReportService) DeleteWeeklyReport(projectID uint, reportID uint) error {
	// Check if weekly report exists and belongs to the project
	var report models.WeeklyReport
	if err := s.db.Where("id = ? AND project_id = ?", reportID, projectID).First(&report).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("weekly report not found")
		}
		return err
	}
	
	// Soft delete
	result := s.db.Delete(&report)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// GetWeeklyReportsByYear retrieves weekly reports for a specific year
func (s *weeklyReportService) GetWeeklyReportsByYear(projectID uint, year int) ([]models.WeeklyReportDTO, error) {
	var reports []models.WeeklyReport
	result := s.db.Preload("Project").
		Where("project_id = ? AND year = ?", projectID, year).
		Order("week DESC").
		Find(&reports)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return s.convertToDTO(reports), nil
}

// CheckReportExists checks if a report already exists for a given week/year
func (s *weeklyReportService) CheckReportExists(projectID uint, week int, year int) (bool, error) {
	var count int64
	result := s.db.Model(&models.WeeklyReport{}).
		Where("project_id = ? AND week = ? AND year = ?", projectID, week, year).
		Count(&count)
	
	if result.Error != nil {
		return false, result.Error
	}
	
	return count > 0, nil
}

// validateCreateRequest validates create request data
func (s *weeklyReportService) validateCreateRequest(request *models.WeeklyReportCreateRequest) error {
	if request.ProjectID == 0 {
		return errors.New("project ID is required")
	}
	
	if request.Week < 1 || request.Week > 53 {
		return errors.New("week must be between 1 and 53")
	}
	
	if request.Year < 2000 {
		return errors.New("year must be 2000 or later")
	}
	
	if request.TotalWorkDays < 0 {
		return errors.New("total work days cannot be negative")
	}
	
	if request.WeatherDelays < 0 {
		return errors.New("weather delays cannot be negative")
	}
	
	if request.TeamSize < 0 {
		return errors.New("team size cannot be negative")
	}
	
	return nil
}

// convertToDTO converts a slice of WeeklyReport to DTOs
func (s *weeklyReportService) convertToDTO(reports []models.WeeklyReport) []models.WeeklyReportDTO {
	dtos := make([]models.WeeklyReportDTO, len(reports))
	for i, report := range reports {
		dtos[i] = s.convertSingleToDTO(&report)
	}
	return dtos
}

// convertSingleToDTO converts a single WeeklyReport to DTO
func (s *weeklyReportService) convertSingleToDTO(report *models.WeeklyReport) models.WeeklyReportDTO {
	dto := models.WeeklyReportDTO{
		ID:                 report.ID,
		ProjectID:          report.ProjectID,
		Week:               report.Week,
		Year:               report.Year,
		WeekLabel:          fmt.Sprintf("Week %d, %d", report.Week, report.Year),
		ProjectManager:     report.ProjectManager,
		TotalWorkDays:      report.TotalWorkDays,
		WeatherDelays:      report.WeatherDelays,
		TeamSize:           report.TeamSize,
		Accomplishments:    report.Accomplishments,
		Challenges:         report.Challenges,
		NextWeekPriorities: report.NextWeekPriorities,
		GeneratedDate:      report.GeneratedDate,
		CreatedBy:          report.CreatedBy,
		CreatedAt:          report.CreatedAt,
		UpdatedAt:          report.UpdatedAt,
	}
	
	if report.Project != nil {
		dto.ProjectName = report.Project.ProjectName
		// You could add a project code field if it exists
	}
	
	return dto
}

