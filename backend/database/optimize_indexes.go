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
		// Sales table optimizations
		{
			table:       "sales",
			indexName:   "idx_sales_status_customer",
			columns:     []string{"status", "customer_id"},
			description: "Optimize payment queries by status and customer",
		},
		{
			table:       "sales",
			indexName:   "idx_sales_outstanding",
			columns:     []string{"outstanding_amount"},
			description: "Optimize searches for unpaid invoices",
		},
		{
			table:       "sales",
			indexName:   "idx_sales_date_status",
			columns:     []string{"invoice_date", "status"},
			description: "Optimize date-based reports with status filter",
		},

		// Payments table optimizations
		{
			table:       "payments",
			indexName:   "idx_payments_contact_date",
			columns:     []string{"contact_id", "date"},
			description: "Optimize payment history queries",
		},
		{
			table:       "payments",
			indexName:   "idx_payments_status_method",
			columns:     []string{"status", "method"},
			description: "Optimize payment reporting by status and method",
		},

		// Cash banks optimizations
		{
			table:       "cash_banks",
			indexName:   "idx_cash_banks_active",
			columns:     []string{"is_active", "account_type"},
			description: "Optimize cash bank dropdown queries",
		},

		// Journal entries optimizations
		{
			table:       "ssot_journal_entries",
			indexName:   "idx_journal_source",
			columns:     []string{"source_type", "source_id"},
			description: "Optimize SSOT journal source lookups",
		},
		{
			table:       "ssot_journal_entries",
			indexName:   "idx_journal_date_status",
			columns:     []string{"entry_date", "status"},
			description: "Optimize journal date and status queries",
		},

		// Products table optimizations
		{
			table:       "products",
			indexName:   "idx_products_category_active",
			columns:     []string{"category_id", "is_active"},
			description: "Optimize product category queries",
		},
		{
			table:       "products",
			indexName:   "idx_products_stock",
			columns:     []string{"current_stock", "min_stock"},
			description: "Optimize stock level monitoring",
		},

		// Accounts table optimizations
		{
			table:       "accounts",
			indexName:   "idx_accounts_code_parent",
			columns:     []string{"code", "parent_id"},
			description: "Optimize account hierarchy queries",
		},
		{
			table:       "accounts",
			indexName:   "idx_accounts_type_active",
			columns:     []string{"account_type", "is_active"},
			description: "Optimize account type filtering",
		},

		// Contacts table optimizations
		{
			table:       "contacts",
			indexName:   "idx_contacts_type_name",
			columns:     []string{"contact_type", "name"},
			description: "Optimize contact searches by type and name",
		},

		// Purchases table optimizations
		{
			table:       "purchases",
			indexName:   "idx_purchases_vendor_status",
			columns:     []string{"vendor_id", "status"},
			description: "Optimize vendor purchase queries",
		},
		{
			table:       "purchases",
			indexName:   "idx_purchases_date_status",
			columns:     []string{"purchase_date", "status"},
			description: "Optimize purchase date reports",
		},
	}

	successCount := 0
	for _, opt := range optimizations {
		if err := createIndexIfNotExists(db, opt); err != nil {
			log.Printf("⚠️ Warning: Failed to create index %s on %s: %v", opt.indexName, opt.table, err)
		} else {
			successCount++
			log.Printf("✅ Created index %s on %s: %s", opt.indexName, opt.table, opt.description)
		}
		
		// Small delay to prevent overwhelming the database
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("🎯 Database optimization complete: %d/%d indexes created successfully", successCount, len(optimizations))
	return nil
}

// indexOptimization represents a database index optimization
type indexOptimization struct {
	table       string
	indexName   string
	columns     []string
	description string
}

// createIndexIfNotExists creates an index if it doesn't already exist
func createIndexIfNotExists(db *gorm.DB, opt indexOptimization) error {
	// Check if index already exists
	var count int64
	indexCheckQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		AND table_name = '%s' 
		AND index_name = '%s'
	`, opt.table, opt.indexName)

	if err := db.Raw(indexCheckQuery).Scan(&count).Error; err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if count > 0 {
		log.Printf("⏭️ Index %s already exists on %s", opt.indexName, opt.table)
		return nil
	}

	// Create the index
	columns := ""
	for i, col := range opt.columns {
		if i > 0 {
			columns += ", "
		}
		columns += col
	}

	createIndexSQL := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", opt.indexName, opt.table, columns)
	
	if err := db.Exec(createIndexSQL).Error; err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// OptimizeDatabaseSettings applies general database optimization settings
func OptimizeDatabaseSettings(db *gorm.DB) error {
	log.Println("⚙️ Applying database optimization settings...")

	// Get the underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying DB: %w", err)
	}

	// Set connection pool settings for better performance
	sqlDB.SetMaxOpenConns(25)      // Limit total connections
	sqlDB.SetMaxIdleConns(10)      // Keep some connections idle
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Rotate connections

	log.Println("✅ Database connection pool optimized")

	// Try to apply MySQL-specific optimizations
	optimizationQueries := []struct {
		query       string
		description string
	}{
		{
			query:       "SET SESSION query_cache_type = ON",
			description: "Enable query cache for session",
		},
		{
			query:       "SET SESSION query_cache_size = 67108864", // 64MB
			description: "Set query cache size to 64MB",
		},
	}

	for _, opt := range optimizationQueries {
		if err := db.Exec(opt.query).Error; err != nil {
			log.Printf("⚠️ Warning: %s failed: %v", opt.description, err)
		} else {
			log.Printf("✅ %s", opt.description)
		}
	}

	return nil
}

// AnalyzeSlowQueries identifies potentially slow queries
func AnalyzeSlowQueries(db *gorm.DB) error {
	log.Println("🔍 Analyzing database for slow query patterns...")

	// Common slow query patterns to check
	slowQueryChecks := []struct {
		query       string
		description string
		suggestion  string
	}{
		{
			query: `
				SELECT TABLE_NAME, ROUND((DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024, 2) AS 'Size (MB)'
				FROM information_schema.TABLES 
				WHERE table_schema = DATABASE() 
				ORDER BY (DATA_LENGTH + INDEX_LENGTH) DESC 
				LIMIT 10
			`,
			description: "Largest tables by size",
			suggestion:  "Consider archiving old data or partitioning large tables",
		},
		{
			query: `
				SELECT CONCAT(table_schema,'.',table_name) AS 'Table Name', 
				       table_rows AS 'Number of Rows',
				       ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size in MB'
				FROM information_schema.TABLES
				WHERE table_schema = DATABASE()
				AND table_rows > 10000
				ORDER BY table_rows DESC
			`,
			description: "Tables with high row counts",
			suggestion:  "Consider indexing frequently queried columns",
		},
	}

	for _, check := range slowQueryChecks {
		log.Printf("📊 %s:", check.description)
		
		var results []map[string]interface{}
		if err := db.Raw(check.query).Find(&results).Error; err != nil {
			log.Printf("⚠️ Query failed: %v", err)
			continue
		}

		if len(results) > 0 {
			log.Printf("💡 %s", check.suggestion)
			for i, result := range results {
				if i >= 5 { // Limit output
					break
				}
				log.Printf("   - %+v", result)
			}
		} else {
			log.Printf("✅ No issues found")
		}
	}

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

	// Step 3: Analyze slow queries
	if err := AnalyzeSlowQueries(db); err != nil {
		log.Printf("⚠️ Warning: Slow query analysis failed: %v", err)
	}

	duration := time.Since(start)
	log.Printf("🎯 Full database optimization completed in %v", duration)
	
	return nil
}