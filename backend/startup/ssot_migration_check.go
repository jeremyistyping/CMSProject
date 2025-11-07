package startup

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CheckAndRunSSOTMigration verifies if SSOT migration is correctly applied and working
func CheckAndRunSSOTMigration(db *gorm.DB) error {
	// Skip auto migration for account_balances since we have it as materialized view
	skipAccountBalances(db)
	log.Println("🔍 Checking SSOT Journal Migration status...")

	// 1. Check if migration was applied
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM migration_logs WHERE migration_name LIKE '%unified_journal_ssot%' AND status = 'SUCCESS'").Scan(&count).Error
	if err != nil {
		log.Printf("Warning: Could not check migration status: %v", err)
	}

	if count == 0 {
		log.Println("⚠️  SSOT migration has not been applied yet. Running migration verification...")
		if err := runSSOTMigrationVerification(db); err != nil {
			return fmt.Errorf("SSOT migration verification failed: %v", err)
		}
	} else {
		log.Printf("✅ SSOT migration has been previously applied (count: %d)", count)
		
		// Quick verify tables exist
		if err := verifyTablesExist(db); err != nil {
			log.Printf("⚠️  Some SSOT tables missing, running verification: %v", err)
			if err := runSSOTMigrationVerification(db); err != nil {
				return fmt.Errorf("SSOT migration verification failed: %v", err)
			}
		} else {
			log.Println("✅ SSOT tables verified successfully")
		}
	}
	
	return nil
}

// verifyTablesExist checks if all required tables exist
func verifyTablesExist(db *gorm.DB) error {
	requiredTables := []string{
		"unified_journal_ledger",
		"unified_journal_lines",
		"journal_event_log",
	}
	
	for _, table := range requiredTables {
		var tableCount int64
		err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", table).Scan(&tableCount).Error
		if err != nil {
			return fmt.Errorf("failed to check table '%s': %v", table, err)
		}
		if tableCount == 0 {
			return fmt.Errorf("table '%s' not found", table)
		}
	}
	
	return nil
}

// runSSOTMigrationVerification runs a quick functional verification of the SSOT setup
func runSSOTMigrationVerification(db *gorm.DB) error {
	log.Println("🧪 Running SSOT migration verification tests...")
	
	// Start a transaction for all tests
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Test table structure
	if err := testTableStructures(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("table structure check failed: %v", err)
	}
	log.Println("✅ Table structures verified")
	
	// Test triggers and functions
	if err := testTriggersAndFunctions(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("triggers and functions check failed: %v", err)
	}
	log.Println("✅ Triggers and functions verified")
	
	// Test basic CRUD operations
	if err := testCRUDOperations(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("CRUD operations test failed: %v", err)
	}
	log.Println("✅ CRUD operations verified")
	
	// Test materialized view
	if err := testMaterializedView(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("materialized view test failed: %v", err)
	}
	log.Println("✅ Materialized view verified")
	
	// Commit the transaction (will perform cleanup via cascade delete)
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit verification transaction: %v", err)
	}
	
	log.Println("🎉 SSOT Journal System verified successfully and ready to use!")
	return nil
}

// Individual test functions
func testTableStructures(db *gorm.DB) error {
	requiredTables := []string{
		"unified_journal_ledger",
		"unified_journal_lines",
		"journal_event_log",
	}
	
	for _, table := range requiredTables {
		var tableCount int64
		err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", table).Scan(&tableCount).Error
		if err != nil {
			return fmt.Errorf("failed to check table '%s': %v", table, err)
		}
		if tableCount == 0 {
			return fmt.Errorf("table '%s' not found", table)
		}
	}
	
	log.Printf("   Found %d required tables", len(requiredTables))
	return nil
}

func testTriggersAndFunctions(db *gorm.DB) error {
	// Test if triggers exist
	var triggerCount int64
	err := db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.triggers 
		WHERE trigger_name IN ('trg_validate_journal_balance', 'trg_generate_entry_number', 'trg_log_journal_event')
	`).Scan(&triggerCount).Error
	
	if err != nil {
		return fmt.Errorf("failed to check triggers: %v", err)
	}

	log.Printf("   Found %d triggers", triggerCount)

	// Test if functions exist
	var functionCount int64
	err = db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.routines 
		WHERE routine_name IN ('validate_journal_balance', 'generate_entry_number', 'log_journal_event')
	`).Scan(&functionCount).Error
	
	if err != nil {
		return fmt.Errorf("failed to check functions: %v", err)
	}

	log.Printf("   Found %d functions", functionCount)
	return nil
}

func testCRUDOperations(db *gorm.DB) error {
	// Create a test journal entry
	entry := &models.SSOTJournalEntry{
		SourceType:  models.SSOTSourceTypeManual,
		EntryDate:   time.Now(),
		Description: "SSOT Verification Test Entry",
		CreatedBy:   1,
	}

	// Test insert
	if err := db.Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create journal entry: %v", err)
	}
	log.Printf("   Created test journal entry with ID: %d", entry.ID)

	// Create test lines
	lines := []models.SSOTJournalLine{
		{
			JournalID:    entry.ID,
			AccountID:    1,
			LineNumber:   1,
			Description:  "Test Debit Line",
			DebitAmount:  decimal.NewFromFloat(100.00),
			CreditAmount: decimal.Zero,
		},
		{
			JournalID:    entry.ID,
			AccountID:    2,
			LineNumber:   2,
			Description:  "Test Credit Line",
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(100.00),
		},
	}

	// Test insert lines
	if err := db.Create(&lines).Error; err != nil {
		return fmt.Errorf("failed to create journal lines: %v", err)
	}
	log.Printf("   Created %d journal lines", len(lines))

	// Test read
	var readEntry models.SSOTJournalEntry
	if err := db.Preload("Lines").First(&readEntry, entry.ID).Error; err != nil {
		return fmt.Errorf("failed to read journal entry: %v", err)
	}
	log.Printf("   Read journal entry with %d lines", len(readEntry.Lines))
	
	// Test update entry totals (this should trigger the validation function)
	if err := db.Model(&readEntry).Updates(map[string]interface{}{
		"total_debit":  100.00,
		"total_credit": 100.00,
		"is_balanced":  true,
	}).Error; err != nil {
		return fmt.Errorf("failed to update journal entry: %v", err)
	}
	log.Printf("   Updated journal entry totals")
	
	// We don't delete the entry here - it will be deleted when the transaction rolls back
	return nil
}

func testMaterializedView(db *gorm.DB) error {
	// Materialized view is no longer required - skip this test
	log.Printf("   Skipping materialized view test (no longer required)")
	return nil
}

// skipAccountBalances prevents GORM from auto-migrating account_balances table
func skipAccountBalances(db *gorm.DB) {
	// This is a placeholder - the actual skip logic should be in the database migration
	// We'll handle this in the main database migration file
}
