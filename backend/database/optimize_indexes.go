package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// OptimizeDatabaseIndexes adds missing indexes to improve query performance
func OptimizeDatabaseIndexes(db *gorm.DB) error {
	log.Println("🔧 Starting database index optimization...")

	optimizations := []indexOptimization{
		// Projects table optimizations
		{
			table:       "projects",
			indexName:   "idx_projects_status",
			columns:     []string{"status"},
			description: "Optimize project status queries",
		},

		// Daily updates table optimizations
		{
			table:       "daily_updates",
			indexName:   "idx_daily_updates_project",
			columns:     []string{"project_id"},
			description: "Optimize daily update queries by project",
		},

		// Milestones table optimizations
		{
			table:       "milestones",
			indexName:   "idx_milestones_project",
			columns:     []string{"project_id"},
			description: "Optimize milestone queries by project",
		},
		{
			table:       "milestones",
			indexName:   "idx_milestones_status",
			columns:     []string{"status"},
			description: "Optimize milestone status queries",
		},

		// Weekly reports table optimizations
		{
			table:       "weekly_reports",
			indexName:   "idx_weekly_reports_project",
			columns:     []string{"project_id"},
			description: "Optimize weekly report queries by project",
		},

		// Purchase requests table optimizations
		{
			table:       "purchase_requests",
			indexName:   "idx_purchase_requests_project",
			columns:     []string{"project_id"},
			description: "Optimize purchase request queries by project",
		},
		{
			table:       "purchase_requests",
			indexName:   "idx_purchase_requests_status",
			columns:     []string{"status"},
			description: "Optimize purchase request status queries",
		},

		// CBS nodes table optimizations
		{
			table:       "cbs_nodes",
			indexName:   "idx_cbs_nodes_project",
			columns:     []string{"project_id"},
			description: "Optimize CBS node queries by project",
		},
		{
			table:       "cbs_nodes",
			indexName:   "idx_cbs_nodes_parent",
			columns:     []string{"parent_id"},
			description: "Optimize CBS node hierarchy queries",
		},

		// Notifications table optimizations
		{
			table:       "notifications",
			indexName:   "idx_notifications_user",
			columns:     []string{"user_id"},
			description: "Optimize notification queries by user",
		},
		{
			table:       "notifications",
			indexName:   "idx_notifications_read",
			columns:     []string{"is_read"},
			description: "Optimize unread notification queries",
		},

		// Users table optimizations
		{
			table:       "users",
			indexName:   "idx_users_role",
			columns:     []string{"role"},
			description: "Optimize user role queries",
		},
	}

	successCount := 0
	for _, opt := range optimizations {
		if err := createIndexIfNotExists(db, opt); err != nil {
			// Silently skip if table doesn't exist
			continue
		} else {
			successCount++
		}

		// Small delay to prevent overwhelming the database
		time.Sleep(50 * time.Millisecond)
	}

	if successCount > 0 {
		log.Printf("🎯 Database optimization complete: %d indexes verified/created", successCount)
	}
	return nil
}

// indexOptimization represents a database index optimization
type indexOptimization struct {
	table       string
	indexName   string
	columns     []string
	description string
}

// createIndexIfNotExists creates an index if it doesn't already exist (PostgreSQL)
func createIndexIfNotExists(db *gorm.DB, opt indexOptimization) error {
	// Check if table exists first
	var tableExists bool
	tableCheckQuery := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)`
	if err := db.Raw(tableCheckQuery, opt.table).Scan(&tableExists).Error; err != nil {
		return err
	}
	if !tableExists {
		return nil // Table doesn't exist, skip silently
	}

	// Build column list
	columns := ""
	for i, col := range opt.columns {
		if i > 0 {
			columns += ", "
		}
		columns += col
	}

	// Create index if not exists (PostgreSQL syntax)
	createIndexSQL := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", opt.indexName, opt.table, columns)

	if err := db.Exec(createIndexSQL).Error; err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// OptimizeDatabaseSettings applies general database optimization settings
func OptimizeDatabaseSettings(db *gorm.DB) error {
	// Get the underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying DB: %w", err)
	}

	// Set connection pool settings for better performance
	sqlDB.SetMaxOpenConns(25)                   // Limit total connections
	sqlDB.SetMaxIdleConns(10)                   // Keep some connections idle
	sqlDB.SetConnMaxLifetime(5 * time.Minute)   // Rotate connections

	return nil
}

// RunFullDatabaseOptimization runs all optimization procedures
func RunFullDatabaseOptimization(db *gorm.DB) error {
	start := time.Now()
	log.Println("🚀 Starting full database optimization...")

	// Step 1: Optimize indexes
	if err := OptimizeDatabaseIndexes(db); err != nil {
		return fmt.Errorf("index optimization failed: %w", err)
	}

	// Step 2: Optimize database settings
	if err := OptimizeDatabaseSettings(db); err != nil {
		log.Printf("⚠️ Warning: Database settings optimization failed: %v", err)
	}

	duration := time.Since(start)
	log.Printf("🎯 Full database optimization completed in %v", duration)

	return nil
}
