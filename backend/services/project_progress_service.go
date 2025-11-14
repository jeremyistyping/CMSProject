package services

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ProjectProgressService defines operations for managing project physical progress history
// backed by the project_progress table.
type ProjectProgressService interface {
	GetProgressHistory(projectID uint, startDate, endDate *time.Time) ([]models.ProjectProgress, error)
	UpsertProgressSnapshot(projectID uint, date time.Time, physicalProgress float64, volumeAchieved *float64, remarks string) (*models.ProjectProgress, error)
}

type projectProgressService struct {
	db *gorm.DB
}

// NewProjectProgressService creates a new ProjectProgressService
func NewProjectProgressService(db *gorm.DB) ProjectProgressService {
	return &projectProgressService{db: db}
}

// GetProgressHistory returns progress snapshots for a project within an optional date range.
func (s *projectProgressService) GetProgressHistory(projectID uint, startDate, endDate *time.Time) ([]models.ProjectProgress, error) {
	var records []models.ProjectProgress

	query := s.db.Where("project_id = ?", projectID)
	if startDate != nil {
		query = query.Where("date >= ?", startDate.Format("2006-01-02"))
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate.Format("2006-01-02"))
	}

	if err := query.Order("date ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query project progress history: %w", err)
	}

	return records, nil
}

// UpsertProgressSnapshot creates or updates a progress snapshot per (project_id, date).
// If a record exists for that date it will be updated, otherwise a new one is created.
func (s *projectProgressService) UpsertProgressSnapshot(projectID uint, date time.Time, physicalProgress float64, volumeAchieved *float64, remarks string) (*models.ProjectProgress, error) {
	if projectID == 0 {
		return nil, fmt.Errorf("project_id is required")
	}

	// Clamp progress between 0 and 100 for safety
	if physicalProgress < 0 {
		physicalProgress = 0
	}
	if physicalProgress > 100 {
		physicalProgress = 100
	}

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var existing models.ProjectProgress
		queryDate := date.Format("2006-01-02")

		err := tx.Where("project_id = ? AND date = ?", projectID, queryDate).
			First(&existing).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if err == gorm.ErrRecordNotFound {
			// Create new snapshot
			newRecord := models.ProjectProgress{
				ProjectID:               projectID,
				Date:                    date,
				PhysicalProgressPercent: physicalProgress,
				Remarks:                 remarks,
			}
			if volumeAchieved != nil {
				newRecord.VolumeAchieved = volumeAchieved
			}

			if err := tx.Create(&newRecord).Error; err != nil {
				return err
			}
			// Overwrite existing to return back to caller
			existing = newRecord
		} else {
			// Update existing snapshot
			existing.PhysicalProgressPercent = physicalProgress
			existing.Remarks = remarks
			if volumeAchieved != nil {
				existing.VolumeAchieved = volumeAchieved
			}

			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if txErr != nil {
		return nil, fmt.Errorf("failed to upsert project progress: %w", txErr)
	}

	// Re-fetch latest snapshot for the given date to return consistent data
	var result models.ProjectProgress
	if err := s.db.Where("project_id = ? AND date = ?", projectID, date.Format("2006-01-02")).First(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to reload project progress: %w", err)
	}

	return &result, nil
}
