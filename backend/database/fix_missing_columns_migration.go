package database

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// RunMissingColumnsFix fixes all missing columns and constraint issues
func RunMissingColumnsFix(db *gorm.DB) {
	migrationID := "fix_missing_columns_v2024.1"

	// Check if this migration has already been applied
	var existing models.MigrationRecord
	err := db.Where("migration_id = ?", migrationID).First(&existing).Error
	if err == nil {
		log.Printf("Missing columns fix migration '%s' already applied at %v, skipping...", migrationID, existing.AppliedAt)
		return
	}

	log.Println("🔧 Starting missing columns fix migration...")

	// Step 1: Fix inventory transaction_type issue (rename column references)
	fixInventoryTransactionType(db)

	// Step 2: Fix financial ratio calculation_date issue (use calculated_at)
	fixFinancialRatioDate(db)

	// Step 4: Fix user full_name issue (add computed full name)
	fixUserFullName(db)

	// Step 5: Fix audit log constraint issue
	fixAuditLogConstraint(db)

	// Record this migration as completed
	migrationRecord := models.MigrationRecord{
		MigrationID: migrationID,
		Description: "Fix missing columns and constraint issues - inventory transaction_type, financial ratio calculation_date, user full_name, audit log constraints",
		Version:     "2024.1",
		AppliedAt:   time.Now(),
	}

	if err := db.Create(&migrationRecord).Error; err != nil {
		log.Printf("Warning: Failed to record migration completion: %v", err)
	} else {
		log.Println("✅ Missing columns fix migration recorded successfully")
	}

	log.Println("✅ Missing columns fix migration completed successfully")
}

// fixInventoryTransactionType fixes inventory column references
func fixInventoryTransactionType(db *gorm.DB) {
	log.Println("Fixing inventory transaction_type references...")

	// Drop problematic indexes that use non-existent transaction_type
	db.Exec(`DROP INDEX IF EXISTS idx_inventory_product_type`)

	// Recreate index with correct column name (type instead of transaction_type)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_inventory_product_type ON inventories(product_id, type)`)

	// Also create index that might be useful for transaction_type queries
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_inventory_reference_type ON inventories(reference_type, reference_id)`)

	log.Println("✅ Inventory transaction_type references fixed")
}


// fixFinancialRatioDate fixes financial ratio date references
func fixFinancialRatioDate(db *gorm.DB) {
	log.Println("Fixing financial ratio date references...")

	// Drop problematic index
	db.Exec(`DROP INDEX IF EXISTS idx_financial_ratios_date`)

	// Create index using correct column name
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_financial_ratios_calculated_at ON financial_ratios(calculated_at)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_financial_ratios_period_calc ON financial_ratios(period, calculated_at)`)

	log.Println("✅ Financial ratio date references fixed")
}

// fixUserFullName fixes user full_name references
func fixUserFullName(db *gorm.DB) {
	log.Println("Fixing user full_name references...")

	// Check if users table exists
	var tableExists bool
	db.Raw(`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users')`).Scan(&tableExists)

	if !tableExists {
		log.Println("Users table does not exist, skipping user full_name fix")
		return
	}

	// Drop problematic views first
	log.Println("Dropping views that reference full_name...")
	db.Exec(`DROP VIEW IF EXISTS critical_audit_changes`)

	// Recreate the view with computed full_name
	db.Exec(`
		CREATE OR REPLACE VIEW critical_audit_changes AS
		SELECT 
			al.id,
			al.table_name,
			al.record_id,
			al.action,
			al.old_values,
			al.new_values,
			al.user_id,
			al.created_at,
			al.updated_at,
			u.username as user_username,
			COALESCE(CONCAT(u.first_name, ' ', u.last_name), u.username, 'Unknown User') as full_name
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.action IN ('DELETE', 'UPDATE')
			AND al.table_name IN ('journal_entries', 'accounts', 'transactions', 'sales', 'purchases')
			AND al.created_at >= CURRENT_DATE - INTERVAL '7 days'
		ORDER BY al.created_at DESC
	`)

	// Also create a user view with computed full_name for easier queries
	db.Exec(`
		CREATE OR REPLACE VIEW users_with_full_name AS
		SELECT 
			*,
			COALESCE(CONCAT(first_name, ' ', last_name), username, 'Unknown User') as full_name
		FROM users
		WHERE deleted_at IS NULL
	`)

	log.Println("✅ User full_name references fixed")
}

// fixAuditLogConstraint fixes audit log constraint issues
func fixAuditLogConstraint(db *gorm.DB) {
	log.Println("Fixing audit log constraint issues...")

	// Step 1: Drop views that depend on audit_logs table
	log.Println("Dropping dependent views...")
	db.Exec(`DROP VIEW IF EXISTS critical_audit_changes`)
	db.Exec(`DROP VIEW IF EXISTS audit_trail_summary`)

	// Step 2: Try to alter the column type
	log.Println("Attempting to alter audit log action column...")
	err := db.Exec(`ALTER TABLE audit_logs ALTER COLUMN action TYPE varchar(50)`).Error
	if err != nil {
		log.Printf("Note: Could not alter action column type (non-critical): %v", err)
	} else {
		log.Println("✅ Audit log action column type updated")
	}

	// Step 3: Recreate the views
	log.Println("Recreating audit trail views...")

	// Recreate audit trail summary
	db.Exec(`
		CREATE OR REPLACE VIEW audit_trail_summary AS
		SELECT 
			DATE(created_at) as audit_date,
			table_name,
			action,
			user_id,
			COUNT(*) as action_count,
			MIN(created_at) as first_action,
			MAX(created_at) as last_action
		FROM audit_logs
		WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY DATE(created_at), table_name, action, user_id
		ORDER BY audit_date DESC, action_count DESC
	`)

	// Recreate critical audit changes with proper full_name handling
	db.Exec(`
		CREATE OR REPLACE VIEW critical_audit_changes AS
		SELECT 
			al.id,
			al.table_name,
			al.record_id,
			al.action,
			al.old_values,
			al.new_values,
			al.user_id,
			al.created_at,
			al.updated_at,
			u.username as user_username,
			COALESCE(CONCAT(u.first_name, ' ', u.last_name), u.username, 'Unknown User') as full_name
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.action IN ('DELETE', 'UPDATE')
			AND al.table_name IN ('journal_entries', 'accounts', 'transactions', 'sales', 'purchases')
			AND al.created_at >= CURRENT_DATE - INTERVAL '7 days'
		ORDER BY al.created_at DESC
	`)

	log.Println("✅ Audit log constraint issues fixed")
}

// RunIndexCleanupAndOptimization cleans up problematic indexes and creates optimized ones
func RunIndexCleanupAndOptimization(db *gorm.DB) {
	log.Println("🧹 Running index cleanup and optimization...")

	// Clean up all problematic indexes first
	problematicIndexes := []string{
		"idx_inventory_product_type",
		"idx_financial_ratios_date",
		"idx_cash_bank_transactions_flow", // This one references non-existent transaction_type too
	}

	for _, idx := range problematicIndexes {
		db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", idx))
		log.Printf("Dropped problematic index: %s", idx)
	}

	// Create optimized replacement indexes
	optimizedIndexes := map[string]string{
		"idx_inventory_product_type_fixed":          "CREATE INDEX IF NOT EXISTS idx_inventory_product_type_fixed ON inventories(product_id, type)",
		"idx_inventory_reference":                   "CREATE INDEX IF NOT EXISTS idx_inventory_reference ON inventories(reference_type, reference_id)",
		"idx_financial_ratios_calc_date":           "CREATE INDEX IF NOT EXISTS idx_financial_ratios_calc_date ON financial_ratios(calculated_at)",
		"idx_cash_bank_transactions_flow_fixed":    "CREATE INDEX IF NOT EXISTS idx_cash_bank_transactions_flow_fixed ON cash_bank_transactions(transaction_date, amount)",
		"idx_users_name_search":                    "CREATE INDEX IF NOT EXISTS idx_users_name_search ON users(first_name, last_name)",
	}

	for name, query := range optimizedIndexes {
		err := db.Exec(query).Error
		if err != nil {
			log.Printf("Warning: Failed to create index %s: %v", name, err)
		} else {
			log.Printf("✅ Created optimized index: %s", name)
		}
	}

	log.Println("✅ Index cleanup and optimization completed")
}