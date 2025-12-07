package database

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"

	"gorm.io/gorm"
)

// InitializeDatabase runs migrations and seeds initial data
func InitializeDatabase(db *gorm.DB) {
	log.Println("🚀 UNIPRO PROJECT MANAGEMENT - DATABASE INITIALIZATION")
	log.Println("Initializing database...")

	// Run GORM auto migrations for basic models
	RunMigrations(db)

	// Run SQL-based auto migrations
	if err := RunAutoMigrations(db); err != nil {
		log.Printf("⚠️  Auto migration warning: %v", err)
	}

	// Seed initial data
	SeedData(db)

	// Run notification system migrations
	MigrateNotificationConfig(db)

	// Clean up duplicate notifications
	CleanupDuplicateNotificationsMigration(db)

	log.Println("✅ Database initialization completed")
}

// RunMigrations creates all tables based on models
func RunMigrations(db *gorm.DB) {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		// Core models
		&models.User{},

		// Auth models
		&models.RefreshToken{},
		&models.UserSession{},
		&models.BlacklistedToken{},
		&models.AuthAttempt{},
		&models.RateLimitRecord{},
		&models.Permission{},
		&models.RolePermission{},

		// Approval models
		&models.ApprovalWorkflow{},
		&models.ApprovalStep{},
		&models.ApprovalRequest{},
		&models.ApprovalAction{},
		&models.ApprovalHistory{},

		// Project models
		&models.Project{},
		&models.DailyUpdate{},
		&models.Milestone{},
		&models.WeeklyReport{},
		&models.TimelineSchedule{},
		&models.ProjectBudget{},
		&models.ProjectProgress{},
		&models.ProjectActualCost{},

		// Cost Control models
		&models.CBSNode{},
		&models.PRCBSMapping{},
		&models.PurchaseRequest{},
		&models.PurchaseRequestItem{},

		// Budget models
		&models.Budget{},
		&models.BudgetItem{},
		&models.BudgetComparison{},

		// Audit models
		&models.AuditLog{},
		&models.ActivityLog{},

		// Security models
		&models.SecurityIncident{},
		&models.SystemAlert{},
		&models.RequestLog{},
		&models.IpWhitelist{},
		&models.SecurityConfig{},
		&models.SecurityMetrics{},

		// Migration tracking
		&models.MigrationRecord{},

		// Notification models
		&models.Notification{},
		&models.StockAlert{},
		&models.NotificationRule{},
		&models.NotificationPreference{},
		&models.NotificationBatch{},
		&models.NotificationQueue{},

		// Settings models
		&models.Settings{},
		&models.SettingsHistory{},
	)

	if err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	log.Println("✅ Database migrations completed successfully")
}

// CreateIndexes creates additional database indexes for performance optimization
func CreateIndexes(db *gorm.DB) {
	log.Println("Creating additional database indexes...")

	// Project indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at)")

	// Daily Update indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_daily_updates_project_id ON daily_updates(project_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_daily_updates_date ON daily_updates(date)")

	// Milestone indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_milestones_project_id ON milestones(project_id)")

	// Purchase Request indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_purchase_requests_project_id ON purchase_requests(project_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_purchase_requests_status ON purchase_requests(status)")

	// CBS indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_cbs_nodes_project_id ON cbs_nodes(project_id)")

	// Notification indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read)")

	// Audit Log indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_table_record ON audit_logs(table_name, record_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action ON audit_logs(user_id, action)")

	// Security indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_security_incidents_type_severity ON security_incidents(incident_type, severity)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_security_incidents_client_ip ON security_incidents(client_ip, created_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_system_alerts_type_level ON system_alerts(alert_type, level)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_request_logs_client_ip ON request_logs(client_ip, timestamp)")

	log.Println("✅ Additional database indexes created successfully")
}

// MigrateNotificationConfig migrates notification configuration
func MigrateNotificationConfig(db *gorm.DB) {
	log.Println("Migrating notification configuration...")

	err := db.AutoMigrate(
		&models.NotificationRule{},
		&models.NotificationPreference{},
		&models.NotificationBatch{},
		&models.NotificationQueue{},
	)

	if err != nil {
		log.Printf("⚠️  Notification config migration warning: %v", err)
	} else {
		log.Println("✅ Notification configuration migrated successfully")
	}
}

// CleanupDuplicateNotificationsMigration removes duplicate notifications
func CleanupDuplicateNotificationsMigration(db *gorm.DB) {
	migrationID := "cleanup_duplicate_notifications_v1.0"

	// Check if this migration has already been applied
	var existing models.MigrationRecord
	err := db.Where("migration_id = ?", migrationID).First(&existing).Error
	if err == nil {
		return // Already applied
	}

	log.Println("Running duplicate notification cleanup migration...")

	// Find and remove duplicate notifications
	var duplicates []struct {
		UserID uint
		Type   string
		Count  int64
	}

	err = db.Raw(`
		SELECT 
			user_id,
			type,
			COUNT(*) as count
		FROM notifications 
		WHERE created_at >= NOW() - INTERVAL '7 days'
		GROUP BY user_id, type, title
		HAVING COUNT(*) > 1
	`).Scan(&duplicates).Error

	if err != nil {
		log.Printf("Error finding duplicate notifications: %v", err)
		return
	}

	totalRemoved := 0
	for _, dup := range duplicates {
		// Keep the latest notification, remove older ones
		result := db.Exec(`
			DELETE FROM notifications 
			WHERE id NOT IN (
				SELECT DISTINCT ON (user_id, type, title) id
				FROM notifications 
				WHERE user_id = ? AND type = ?
				AND created_at >= NOW() - INTERVAL '7 days'
				ORDER BY user_id, type, title, created_at DESC
			)
			AND user_id = ? AND type = ?
			AND created_at >= NOW() - INTERVAL '7 days'
		`, dup.UserID, dup.Type, dup.UserID, dup.Type)

		if result.Error == nil {
			totalRemoved += int(result.RowsAffected)
		}
	}

	// Record this migration as completed
	migrationRecord := models.MigrationRecord{
		MigrationID: migrationID,
		Description: fmt.Sprintf("Cleanup duplicate notifications - removed %d duplicates", totalRemoved),
		Version:     "1.0",
		AppliedAt:   time.Now(),
	}

	db.Create(&migrationRecord)

	log.Printf("✅ Duplicate notification cleanup completed - removed %d duplicates", totalRemoved)
}

// FixConcurrentRefreshError fixes concurrent refresh token issues
func FixConcurrentRefreshError(db *gorm.DB) {
	log.Println("Checking for concurrent refresh token issues...")

	// Clean up expired refresh tokens
	result := db.Exec(`
		DELETE FROM refresh_tokens 
		WHERE expires_at < NOW() OR is_revoked = true
	`)

	if result.Error != nil {
		log.Printf("⚠️  Error cleaning up refresh tokens: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("✅ Cleaned up %d expired/revoked refresh tokens", result.RowsAffected)
	}
}
