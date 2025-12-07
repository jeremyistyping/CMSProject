package database

import (
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDB establishes database connection
func ConnectDB() *gorm.DB {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	log.Println("✅ Database connected successfully")
	return db
}

// AutoMigrate runs GORM auto-migration for all models
func AutoMigrate(db *gorm.DB) {
	log.Println("🔄 Running GORM AutoMigrate...")

	err := db.AutoMigrate(
		// Core models
		&models.User{},
		&models.AuditLog{},
		&models.ActivityLog{},

		// Projects
		&models.Project{},
		&models.DailyUpdate{},
		&models.Milestone{},
		&models.WeeklyReport{},
		&models.TimelineSchedule{},
		&models.ProjectBudget{},
		&models.ProjectProgress{},
		&models.ProjectActualCost{},

		// Cost Control
		&models.CBSNode{},
		&models.PRCBSMapping{},
		&models.PurchaseRequest{},
		&models.PurchaseRequestItem{},

		// Approvals
		&models.ApprovalWorkflow{},
		&models.ApprovalStep{},
		&models.ApprovalRequest{},
		&models.ApprovalAction{},
		&models.ApprovalHistory{},

		// Budgets
		&models.Budget{},
		&models.BudgetItem{},
		&models.BudgetComparison{},

		// Notifications
		&models.Notification{},
		&models.StockAlert{},
		&models.NotificationRule{},
		&models.NotificationPreference{},
		&models.NotificationBatch{},
		&models.NotificationQueue{},

		// Settings
		&models.Settings{},
		&models.SettingsHistory{},

		// Security & Auth
		&models.Permission{},
		&models.RolePermission{},
		&models.RefreshToken{},
		&models.UserSession{},
		&models.BlacklistedToken{},
		&models.AuthAttempt{},
		&models.RateLimitRecord{},
		&models.SecurityIncident{},
		&models.SystemAlert{},
		&models.RequestLog{},
		&models.IpWhitelist{},
		&models.SecurityConfig{},
		&models.SecurityMetrics{},
	)

	if err != nil {
		log.Printf("⚠️ AutoMigrate warning: %v", err)
	} else {
		log.Println("✅ GORM AutoMigrate completed")
	}
}


