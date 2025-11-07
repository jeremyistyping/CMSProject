package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("🧪 Testing SSOT Journal Migration")
	fmt.Println("==================================")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Database connected successfully")

	// Test 1: Check if migration was applied
	fmt.Println("\n1. Checking migration status...")
	if err := testMigrationStatus(db); err != nil {
		log.Printf("❌ Migration status check failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ Migration status check passed")

	// Test 2: Test table structure
	fmt.Println("\n2. Checking table structures...")
	if err := testTableStructures(db); err != nil {
		log.Printf("❌ Table structure check failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ Table structures check passed")

	// Test 3: Test triggers and functions
	fmt.Println("\n3. Checking triggers and functions...")
	if err := testTriggersAndFunctions(db); err != nil {
		log.Printf("❌ Triggers and functions check failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ Triggers and functions check passed")

	// Test 4: Test basic CRUD operations
	fmt.Println("\n4. Testing basic CRUD operations...")
	if err := testCRUDOperations(db); err != nil {
		log.Printf("❌ CRUD operations test failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ CRUD operations test passed")

	// Test 5: Test materialized view
	fmt.Println("\n5. Testing materialized view...")
	if err := testMaterializedView(db); err != nil {
		log.Printf("❌ Materialized view test failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ Materialized view test passed")

	// Test 6: Test balance calculations
	fmt.Println("\n6. Testing balance calculations...")
	if err := testBalanceCalculations(db); err != nil {
		log.Printf("❌ Balance calculations test failed: %v", err)
		os.Exit(1)
	}
	fmt.Println("✅ Balance calculations test passed")

	fmt.Println("\n🎉 All SSOT migration tests passed successfully!")
	fmt.Println("✅ The unified journal system is ready to use.")
}

func testMigrationStatus(db *gorm.DB) error {
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM migration_logs WHERE migration_name LIKE '%unified_journal_ssot%' AND status = 'SUCCESS'").Scan(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check migration status: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("SSOT migration not found or not successful")
	}

	fmt.Printf("   SSOT migration found and successful (count: %d)", count)
	return nil
}

func testTableStructures(db *gorm.DB) error {
	// Test unified_journal_ledger table
	var tableCount int64
	err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'unified_journal_ledger'").Scan(&tableCount).Error
	if err != nil {
		return fmt.Errorf("failed to check unified_journal_ledger table: %v", err)
	}
	if tableCount == 0 {
		return fmt.Errorf("unified_journal_ledger table not found")
	}

	// Test unified_journal_lines table  
	err = db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'unified_journal_lines'").Scan(&tableCount).Error
	if err != nil {
		return fmt.Errorf("failed to check unified_journal_lines table: %v", err)
	}
	if tableCount == 0 {
		return fmt.Errorf("unified_journal_lines table not found")
	}

	// Test journal_event_log table
	err = db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'journal_event_log'").Scan(&tableCount).Error
	if err != nil {
		return fmt.Errorf("failed to check journal_event_log table: %v", err)
	}
	if tableCount == 0 {
		return fmt.Errorf("journal_event_log table not found")
	}

	// Test account_balances materialized view
	err = db.Raw("SELECT COUNT(*) FROM information_schema.views WHERE table_name = 'account_balances'").Scan(&tableCount).Error
	if err != nil {
		return fmt.Errorf("failed to check account_balances view: %v", err)
	}
	if tableCount == 0 {
		return fmt.Errorf("account_balances materialized view not found")
	}

	fmt.Printf("   All required tables and views exist")
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

	fmt.Printf("   Found %d triggers", triggerCount)

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

	fmt.Printf("   Found %d functions", functionCount)
	return nil
}

func testCRUDOperations(db *gorm.DB) error {
	// Create a test journal entry
	entry := &models.UnifiedJournalEntry{
		SourceType:  models.SourceTypeManual,
		EntryDate:   time.Now(),
		Description: "Test SSOT Journal Entry",
		CreatedBy:   1,
	}

	// Test insert
	if err := db.Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create journal entry: %v", err)
	}
	fmt.Printf("   Created journal entry with ID: %d", entry.ID)

	// Create test lines
	lines := []models.UnifiedJournalLine{
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
	for _, line := range lines {
		if err := db.Create(&line).Error; err != nil {
			return fmt.Errorf("failed to create journal line: %v", err)
		}
	}
	fmt.Printf("   Created %d journal lines", len(lines))

	// Test read
	var readEntry models.UnifiedJournalEntry
	if err := db.Preload("Lines").First(&readEntry, entry.ID).Error; err != nil {
		return fmt.Errorf("failed to read journal entry: %v", err)
	}
	fmt.Printf("   Read journal entry with %d lines", len(readEntry.Lines))

	// Test update entry totals (this should trigger the validation function)
	if err := db.Model(&readEntry).Updates(map[string]interface{}{
		"total_debit":  100.00,
		"total_credit": 100.00,
		"is_balanced":  true,
	}).Error; err != nil {
		return fmt.Errorf("failed to update journal entry: %v", err)
	}
	fmt.Printf("   Updated journal entry totals")

	// Test delete (cleanup)
	if err := db.Delete(&readEntry).Error; err != nil {
		return fmt.Errorf("failed to delete journal entry: %v", err)
	}
	fmt.Printf("   Cleaned up test data")

	return nil
}

func testMaterializedView(db *gorm.DB) error {
	// Test if we can query the materialized view
	var balances []models.SSOTAccountBalance
	err := db.Limit(5).Find(&balances).Error
	if err != nil {
		return fmt.Errorf("failed to query account_balances view: %v", err)
	}

	fmt.Printf("   Found %d account balances", len(balances))

	// Test refresh materialized view
	err = db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error
	if err != nil {
		return fmt.Errorf("failed to refresh materialized view: %v", err)
	}
	fmt.Printf("   Refreshed materialized view successfully")

	return nil
}

func testBalanceCalculations(db *gorm.DB) error {
	// Create a test entry with specific amounts to verify balance calculation
	entry := &models.UnifiedJournalEntry{
		SourceType:  models.SourceTypeManual,
		EntryDate:   time.Now(),
		Description: "Balance Test Entry",
		CreatedBy:   1,
	}

	// Insert entry
	if err := db.Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create test entry: %v", err)
	}

	// Create balanced lines
	testAmount := decimal.NewFromFloat(250.50)
	lines := []models.UnifiedJournalLine{
		{
			JournalID:    entry.ID,
			AccountID:    1,
			LineNumber:   1,
			Description:  "Balance Test Debit",
			DebitAmount:  testAmount,
			CreditAmount: decimal.Zero,
		},
		{
			JournalID:    entry.ID,
			AccountID:    2,
			LineNumber:   2,
			Description:  "Balance Test Credit",
			DebitAmount:  decimal.Zero,
			CreditAmount: testAmount,
		},
	}

	// Insert lines
	for _, line := range lines {
		if err := db.Create(&line).Error; err != nil {
			// Cleanup and return error
			db.Delete(entry)
			return fmt.Errorf("failed to create balance test line: %v", err)
		}
	}

	// Test if trigger calculates balance correctly
	var updatedEntry models.UnifiedJournalEntry
	if err := db.First(&updatedEntry, entry.ID).Error; err != nil {
		db.Delete(entry)
		return fmt.Errorf("failed to read updated entry: %v", err)
	}

	// Verify balance calculation
	expectedDebit := testAmount
	expectedCredit := testAmount
	
	if !updatedEntry.TotalDebit.Equal(expectedDebit) {
		db.Delete(entry)
		return fmt.Errorf("total debit mismatch: expected %s, got %s", 
			expectedDebit.String(), updatedEntry.TotalDebit.String())
	}

	if !updatedEntry.TotalCredit.Equal(expectedCredit) {
		db.Delete(entry)
		return fmt.Errorf("total credit mismatch: expected %s, got %s", 
			expectedCredit.String(), updatedEntry.TotalCredit.String())
	}

	if !updatedEntry.IsBalanced {
		db.Delete(entry)
		return fmt.Errorf("entry should be balanced but is_balanced = false")
	}

	fmt.Printf("   Balance calculation verified: D=%s, C=%s, Balanced=%t", 
		updatedEntry.TotalDebit.String(), 
		updatedEntry.TotalCredit.String(), 
		updatedEntry.IsBalanced)

	// Cleanup
	if err := db.Delete(entry).Error; err != nil {
		return fmt.Errorf("failed to cleanup balance test data: %v", err)
	}

	return nil
}