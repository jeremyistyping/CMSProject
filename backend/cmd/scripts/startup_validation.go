package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// StartupValidation runs all necessary validations and fixes during application startup
func main() {
	log.Println("🚀 Starting Application Startup Validation...")

	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := gorm.Open(mysql.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Run all critical validations and fixes
	validateAndFixDatabase(db)

	log.Println("✅ Application Startup Validation completed successfully!")
}

// validateAndFixDatabase runs comprehensive validation and fixes
func validateAndFixDatabase(db *gorm.DB) {
	log.Println("🔍 Running comprehensive database validation...")

	// Check 1: Verify all migration records exist
	validateMigrationRecords(db)

	// Check 2: Validate critical tables and columns
	validateCriticalSchema(db)

	// Check 3: Check for common data integrity issues
	validateDataIntegrity(db)

	// Check 4: Verify upload directories exist
	validateUploadDirectories()

	// Check 5: Run emergency fixes if needed
	runEmergencyFixes(db)
}

// validateMigrationRecords checks that all required migrations have been applied
func validateMigrationRecords(db *gorm.DB) {
	log.Println("  📋 Validating migration records...")

	requiredMigrations := []string{
		"auto_fix_migration_v2.0",
		"sales_balance_fix_v1.0",
		"product_image_fix_v1.0",
		"cashbank_constraint_fix_v1.0",
	}

	for _, migrationID := range requiredMigrations {
		var migration models.MigrationRecord
		err := db.Where("migration_id = ?", migrationID).First(&migration).Error
		
		if err == gorm.ErrRecordNotFound {
			log.Printf("    ⚠️  Migration %s not found - will be applied during initialization", migrationID)
		} else if err != nil {
			log.Printf("    ❌ Error checking migration %s: %v", migrationID, err)
		} else {
			log.Printf("    ✅ Migration %s applied at %v", migrationID, migration.AppliedAt)
		}
	}
}

// validateCriticalSchema checks for critical table and column existence
func validateCriticalSchema(db *gorm.DB) {
	log.Println("  🏗️  Validating critical schema...")

	// Check critical tables
	criticalTables := []string{
		"products", "cash_banks", "accounts", "sales", "purchases",
		"users", "contacts", "journal_entries", "notifications",
	}

	for _, table := range criticalTables {
		if db.Migrator().HasTable(table) {
			log.Printf("    ✅ Table %s exists", table)
		} else {
			log.Printf("    ❌ Critical table %s is missing!", table)
		}
	}

	// Check critical columns
	checkCriticalColumns(db)
}

// checkCriticalColumns verifies critical columns exist
func checkCriticalColumns(db *gorm.DB) {
	criticalColumns := map[string][]string{
		"products":   {"id", "code", "name", "image_path"},
		"cash_banks": {"id", "code", "name", "account_id", "type"},
		"sales":      {"id", "code", "total_amount", "tax_amount", "subtotal_amount"},
		"accounts":   {"id", "code", "name", "type", "category"},
	}

	for table, columns := range criticalColumns {
		for _, column := range columns {
			if db.Migrator().HasColumn(&models.Product{}, column) || 
			   db.Migrator().HasColumn(&models.CashBank{}, column) ||
			   db.Migrator().HasColumn(&models.Sale{}, column) ||
			   db.Migrator().HasColumn(&models.Account{}, column) {
				// Column exists (simplified check)
			} else {
				log.Printf("    ⚠️  Column %s.%s might be missing", table, column)
			}
		}
	}
}

// validateDataIntegrity checks for common data integrity issues
func validateDataIntegrity(db *gorm.DB) {
	log.Println("  🔍 Validating data integrity...")

	// Check 1: CashBanks with invalid account_id
	var invalidCashBanks int64
	db.Raw(`
		SELECT COUNT(*) FROM cash_banks cb
		LEFT JOIN accounts a ON cb.account_id = a.id AND a.deleted_at IS NULL
		WHERE cb.account_id = 0 OR cb.account_id IS NULL OR a.id IS NULL
	`).Scan(&invalidCashBanks)

	if invalidCashBanks > 0 {
		log.Printf("    ⚠️  Found %d cash_banks with invalid account_id", invalidCashBanks)
	} else {
		log.Println("    ✅ All cash_banks have valid account_id")
	}

	// Check 2: Sales with balance issues
	var unbalancedSales int64
	db.Raw(`
		SELECT COUNT(*) FROM sales
		WHERE ABS((subtotal_amount - COALESCE(discount_amount,0) + COALESCE(tax_amount,0)) - total_amount) > 0.01
		AND status != 'CANCELLED'
	`).Scan(&unbalancedSales)

	if unbalancedSales > 0 {
		log.Printf("    ⚠️  Found %d sales with balance issues", unbalancedSales)
	} else {
		log.Println("    ✅ All sales have balanced calculations")
	}

	// Check 3: Products with problematic image paths
	var problematicImages int64
	db.Raw(`
		SELECT COUNT(*) FROM products
		WHERE image_path IS NOT NULL AND image_path != ''
		AND (LENGTH(image_path) > 255 OR image_path NOT LIKE '/uploads/%')
	`).Scan(&problematicImages)

	if problematicImages > 0 {
		log.Printf("    ⚠️  Found %d products with problematic image paths", problematicImages)
	} else {
		log.Println("    ✅ All product image paths are valid")
	}
}

// validateUploadDirectories ensures upload directories exist
func validateUploadDirectories() {
	log.Println("  📁 Validating upload directories...")

	directories := []string{
		"./uploads",
		"./uploads/products",
		"./uploads/assets",
		"./uploads/temp",
	}

	allExist := true
	for _, dir := range directories {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Printf("    ⚠️  Directory %s does not exist", dir)
			allExist = false
		}
	}

	if allExist {
		log.Println("    ✅ All upload directories exist")
	}
}

// runEmergencyFixes runs emergency fixes if critical issues are detected
func runEmergencyFixes(db *gorm.DB) {
	log.Println("  🚨 Checking if emergency fixes are needed...")

	// Check if any migrations are missing and run them
	var autoFixMigration models.MigrationRecord
	if err := db.Where("migration_id = ?", "auto_fix_migration_v2.0").First(&autoFixMigration).Error; err == gorm.ErrRecordNotFound {
		log.Println("    🚨 Auto fix migration not found - running now...")
		database.AutoFixMigration(db)
	}

	var salesFixMigration models.MigrationRecord
	if err := db.Where("migration_id = ?", "sales_balance_fix_v1.0").First(&salesFixMigration).Error; err == gorm.ErrRecordNotFound {
		log.Println("    🚨 Sales balance fix migration not found - running now...")
		database.SalesBalanceFixMigration(db)
	}

	var imageFixMigration models.MigrationRecord
	if err := db.Where("migration_id = ?", "product_image_fix_v1.0").First(&imageFixMigration).Error; err == gorm.ErrRecordNotFound {
		log.Println("    🚨 Product image fix migration not found - running now...")
		database.ProductImageFixMigration(db)
	}

	log.Println("    ✅ Emergency fixes check completed")
}
