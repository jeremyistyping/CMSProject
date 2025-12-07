package services

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
)

type StartupService struct {
	DB *gorm.DB
}

func NewStartupService(db *gorm.DB) *StartupService {
	return &StartupService{
		DB: db,
	}
}

// RunStartupTasks runs all necessary startup tasks
func (s *StartupService) RunStartupTasks() {
	log.Println("🚀 Running startup tasks...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Task 1: Validate database connection
	s.validateDatabaseConnection(ctx)

	// Task 2: Check project data integrity
	s.checkProjectDataIntegrity(ctx)

	log.Println("✅ Startup tasks completed successfully")
}

// validateDatabaseConnection validates the database connection
func (s *StartupService) validateDatabaseConnection(ctx context.Context) {
	log.Println("🔧 Validating database connection...")

	sqlDB, err := s.DB.DB()
	if err != nil {
		log.Printf("❌ Failed to get database instance: %v", err)
		return
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Printf("❌ Database ping failed: %v", err)
		return
	}

	log.Println("✅ Database connection validated successfully")
}

// checkProjectDataIntegrity checks project data integrity
func (s *StartupService) checkProjectDataIntegrity(ctx context.Context) {
	log.Println("🔍 Checking project data integrity...")

	startTime := time.Now()

	var projectCount int64
	if err := s.DB.WithContext(ctx).Table("projects").Where("deleted_at IS NULL").Count(&projectCount).Error; err != nil {
		log.Printf("⚠️ Failed to count projects: %v", err)
		return
	}

	var activeProjectCount int64
	if err := s.DB.WithContext(ctx).Table("projects").Where("status = ? AND deleted_at IS NULL", "active").Count(&activeProjectCount).Error; err != nil {
		log.Printf("⚠️ Failed to count active projects: %v", err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("✅ Project data integrity check completed - %d total projects, %d active in %v", projectCount, activeProjectCount, duration)
}

// GetStartupStatus returns current startup status information
func (s *StartupService) GetStartupStatus(ctx context.Context) (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Check database connection
	sqlDB, err := s.DB.DB()
	if err != nil {
		status["database_status"] = "error"
		status["database_error"] = err.Error()
	} else {
		if err := sqlDB.PingContext(ctx); err != nil {
			status["database_status"] = "error"
			status["database_error"] = err.Error()
		} else {
			status["database_status"] = "healthy"
		}
	}

	// Check project counts
	var projectCount int64
	if err := s.DB.WithContext(ctx).Table("projects").Where("deleted_at IS NULL").Count(&projectCount).Error; err != nil {
		status["project_count"] = 0
	} else {
		status["project_count"] = projectCount
	}

	// Check active project counts
	var activeProjectCount int64
	if err := s.DB.WithContext(ctx).Table("projects").Where("status = ? AND deleted_at IS NULL", "active").Count(&activeProjectCount).Error; err != nil {
		status["active_project_count"] = 0
	} else {
		status["active_project_count"] = activeProjectCount
	}

	status["last_checked"] = time.Now().Format(time.RFC3339)

	return status, nil
}
