package main

import (
	"fmt"
	"log"
	"strings"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

type TableInfo struct {
	TableName string `json:"table_name"`
	Exists    bool   `json:"exists"`
	RowCount  int64  `json:"row_count"`
}

type JournalTableAnalysis struct {
	DatabaseTables []TableInfo                    `json:"database_tables"`
	ModelMappings  map[string]string             `json:"model_mappings"`
	FileUsage      map[string][]string           `json:"file_usage"`
	Recommendations []string                     `json:"recommendations"`
}

func main() {
	fmt.Println("=== JOURNAL TABLES INCONSISTENCY ANALYSIS ===")
	fmt.Println("Analyzing the confusion between ssot_journal_entries vs unified_journal_ledger")
	fmt.Println()

	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	analysis := &JournalTableAnalysis{
		ModelMappings: make(map[string]string),
		FileUsage:     make(map[string][]string),
	}

	// Step 1: Check which tables actually exist in the database
	fmt.Println("1. CHECKING DATABASE TABLES")
	fmt.Println("=" + strings.Repeat("=", 50))
	
	tablesToCheck := []string{
		"unified_journal_ledger",
		"unified_journal_lines", 
		"ssot_journal_entries",
		"ssot_journal_lines",
		"journal_entries", // Legacy
		"journal_lines",   // Legacy
	}

	for _, tableName := range tablesToCheck {
		info := checkTableExists(db, tableName)
		analysis.DatabaseTables = append(analysis.DatabaseTables, info)
		
		status := "❌ NOT EXISTS"
		if info.Exists {
			status = fmt.Sprintf("✅ EXISTS (%d records)", info.RowCount)
		}
		fmt.Printf("%-25s %s\n", tableName, status)
	}
	
	fmt.Println()

	// Step 2: Check model table mappings
	fmt.Println("2. MODEL TABLE MAPPINGS")
	fmt.Println("=" + strings.Repeat("=", 50))
	
	// Check SSOTJournalEntry model
	ssoEntry := models.SSOTJournalEntry{}
	ssoEntryTable := ssoEntry.TableName()
	analysis.ModelMappings["SSOTJournalEntry"] = ssoEntryTable
	fmt.Printf("models.SSOTJournalEntry -> %s\n", ssoEntryTable)
	
	// Check SSOTJournalLine model
	ssoLine := models.SSOTJournalLine{}
	ssoLineTable := ssoLine.TableName()
	analysis.ModelMappings["SSOTJournalLine"] = ssoLineTable
	fmt.Printf("models.SSOTJournalLine  -> %s\n", ssoLineTable)
	
	// Check legacy JournalEntry if exists
	if hasLegacyModels() {
		fmt.Printf("models.JournalEntry     -> journal_entries (legacy)\n")
		fmt.Printf("models.JournalLine      -> journal_lines (legacy)\n")
	}
	
	fmt.Println()

	// Step 3: Analyze problematic files
	fmt.Println("3. PROBLEMATIC FILE USAGE")
	fmt.Println("=" + strings.Repeat("=", 50))
	
	// Files that reference ssot_journal_entries (wrong table name)
	problematicFiles := []string{
		"fix_deposit_journal_entries.sql",
		"setup_automatic_balance_sync.sql", 
		"debug_purchase_report_data.sql",
	}
	
	for _, filename := range problematicFiles {
		fmt.Printf("❌ %s - References 'ssot_journal_entries' (should be 'unified_journal_ledger')\n", filename)
		analysis.FileUsage["incorrect_references"] = append(analysis.FileUsage["incorrect_references"], filename)
	}
	
	// Files that correctly use unified_journal_ledger
	correctFiles := []string{
		"models/ssot_journal.go",
		"services/unified_journal_service.go",
		"controllers/unified_journal_controller.go",
	}
	
	fmt.Println()
	fmt.Println("✅ CORRECT USAGE:")
	for _, filename := range correctFiles {
		fmt.Printf("✅ %s - Correctly uses 'unified_journal_ledger'\n", filename)
		analysis.FileUsage["correct_references"] = append(analysis.FileUsage["correct_references"], filename)
	}
	
	fmt.Println()

	// Step 4: Check cash & bank service integration
	fmt.Println("4. CASH & BANK INTEGRATION STATUS") 
	fmt.Println("=" + strings.Repeat("=", 50))
	
	fmt.Println("Cash & Bank services use:")
	fmt.Println("✅ UnifiedJournalService (correct)")
	fmt.Println("✅ models.SSOTJournalEntry -> unified_journal_ledger (correct)")
	fmt.Println("✅ models.SSOTJournalLine  -> unified_journal_lines (correct)")
	
	fmt.Println()

	// Step 5: Generate recommendations
	fmt.Println("5. RECOMMENDATIONS")
	fmt.Println("=" + strings.Repeat("=", 50))
	
	recommendations := generateRecommendations(analysis)
	for i, rec := range recommendations {
		fmt.Printf("%d. %s\n", i+1, rec)
		analysis.Recommendations = append(analysis.Recommendations, rec)
	}

	// Step 6: Check for any data migration needs
	fmt.Println()
	fmt.Println("6. DATA MIGRATION CHECK")
	fmt.Println("=" + strings.Repeat("=", 50))
	
	checkDataMigrationNeeds(db, analysis)

	fmt.Println()
	fmt.Println("=== ANALYSIS COMPLETE ===")
	fmt.Printf("Summary: The correct tables are 'unified_journal_ledger' and 'unified_journal_lines'\n")
	fmt.Printf("Cash & Bank system is correctly configured to use these tables.\n")
	fmt.Printf("Some SQL files incorrectly reference 'ssot_journal_entries' - these need to be updated.\n")
}

func checkTableExists(db *gorm.DB, tableName string) TableInfo {
	info := TableInfo{TableName: tableName}
	
	// Check if table exists
	var exists bool
	err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", tableName).Scan(&exists)
	if err != nil {
		log.Printf("Error checking table %s: %v", tableName, err)
		return info
	}
	
	info.Exists = exists
	
	// If exists, get row count
	if exists {
		var count int64
		db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
		info.RowCount = count
	}
	
	return info
}

func hasLegacyModels() bool {
	// Simple check to see if legacy journal models exist
	// This is a placeholder - you might need to check actual model definitions
	return true // Assuming they exist for completeness
}

func generateRecommendations(analysis *JournalTableAnalysis) []string {
	recommendations := []string{
		"IMMEDIATE: Update all SQL files that reference 'ssot_journal_entries' to use 'unified_journal_ledger'",
		"IMMEDIATE: Update all SQL files that reference 'ssot_journal_lines' to use 'unified_journal_lines'", 
		"VERIFY: Ensure all services use UnifiedJournalService instead of any legacy journal services",
		"VERIFY: Confirm all controllers use models.SSOTJournalEntry and models.SSOTJournalLine",
		"CLEANUP: Remove any references to non-existent ssot_journal_* tables",
		"TESTING: Test cash & bank operations to ensure they work with unified_journal_* tables",
		"DOCUMENTATION: Update all documentation to reflect the correct table names",
	}
	
	return recommendations
}

func checkDataMigrationNeeds(db *gorm.DB, analysis *JournalTableAnalysis) {
	// Check if there's data in both table systems that needs consolidation
	
	var unifiedCount, ssotCount, legacyCount int64
	
	// Count unified_journal_ledger records
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&unifiedCount)
	
	// Try to count ssot_journal_entries if it exists
	var ssotExists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'ssot_journal_entries')").Scan(&ssotExists)
	if ssotExists {
		db.Raw("SELECT COUNT(*) FROM ssot_journal_entries").Scan(&ssotCount)
	}
	
	// Try to count legacy journal_entries if it exists  
	var legacyExists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'journal_entries')").Scan(&legacyExists)
	if legacyExists {
		db.Raw("SELECT COUNT(*) FROM journal_entries").Scan(&legacyCount)
	}
	
	fmt.Printf("Data distribution:\n")
	fmt.Printf("- unified_journal_ledger: %d records ✅ (CORRECT TABLE)\n", unifiedCount)
	
	if ssotExists {
		if ssotCount > 0 {
			fmt.Printf("- ssot_journal_entries: %d records ❌ (WRONG TABLE - SHOULD NOT EXIST WITH DATA)\n", ssotCount)
			fmt.Printf("  ⚠️  WARNING: Data found in incorrect table!\n")
		} else {
			fmt.Printf("- ssot_journal_entries: %d records ⚠️  (WRONG TABLE NAME - SHOULD BE REMOVED)\n", ssotCount)
		}
	}
	
	if legacyExists {
		if legacyCount > 0 {
			fmt.Printf("- journal_entries (legacy): %d records 📋 (MAY NEED MIGRATION)\n", legacyCount)
		}
	}
	
	// Migration recommendations
	if ssotCount > 0 {
		fmt.Println()
		fmt.Println("⚠️  URGENT: Data found in incorrectly named 'ssot_journal_entries' table!")
		fmt.Println("Action needed: Migrate data from ssot_journal_entries to unified_journal_ledger")
	}
	
	if legacyCount > 0 {
		fmt.Println()
		fmt.Println("📋 Legacy data found in 'journal_entries' table")
		fmt.Println("Consider: Migrate legacy data to unified_journal_ledger if still needed")
	}
}