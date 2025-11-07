package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("🔍 Checking Journal System Architecture")
	fmt.Println("======================================")

	// Check all tables with 'journal' in the name
	fmt.Println("\n📊 All Journal-related Tables:")
	tableRows, err := db.Raw(`
		SELECT table_name, table_type 
		FROM information_schema.tables 
		WHERE table_name LIKE '%journal%' 
			AND table_schema = 'public'
		ORDER BY table_name
	`).Rows()
	
	if err != nil {
		log.Printf("Error querying journal tables: %v", err)
	} else {
		defer tableRows.Close()
		for tableRows.Next() {
			var tableName, tableType string
			tableRows.Scan(&tableName, &tableType)
			fmt.Printf("   - %s (%s)\n", tableName, tableType)
		}
	}

	// Check SSOT-specific tables
	fmt.Println("\n🔄 SSOT Unified Journal Tables:")
	ssotTables := []string{
		"ssot_unified_journal_entries",
		"ssot_unified_journal_lines", 
		"unified_journal_entries",
		"unified_journal_lines",
		"ssot_journal_entries",
		"ssot_journal_lines",
	}

	for _, tableName := range ssotTables {
		var exists bool
		db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)", tableName).Scan(&exists)
		if exists {
			var count int64
			db.Raw("SELECT COUNT(*) FROM " + tableName).Scan(&count)
			fmt.Printf("   ✅ %s - %d records\n", tableName, count)
		} else {
			fmt.Printf("   ❌ %s - does not exist\n", tableName)
		}
	}

	// Check regular journal tables
	fmt.Println("\n📝 Regular Journal Tables:")
	regularTables := []string{"journal_entries", "journal_lines", "journals"}
	
	for _, tableName := range regularTables {
		var exists bool
		db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)", tableName).Scan(&exists)
		if exists {
			var count int64
			db.Raw("SELECT COUNT(*) FROM " + tableName).Scan(&count)
			fmt.Printf("   ✅ %s - %d records\n", tableName, count)

			// Check table structure
			fmt.Printf("     📋 Structure of %s:\n", tableName)
			columnRows, err := db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = ? ORDER BY ordinal_position", tableName).Rows()
			if err == nil {
				defer columnRows.Close()
				for columnRows.Next() {
					var columnName, dataType string
					columnRows.Scan(&columnName, &dataType)
					fmt.Printf("       - %s: %s\n", columnName, dataType)
				}
			}
		} else {
			fmt.Printf("   ❌ %s - does not exist\n", tableName)
		}
	}

	// Check archived tables
	fmt.Println("\n📦 Archived Journal Tables:")
	archivedTableRows, err := db.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_name LIKE '%journal%archived%' 
			AND table_schema = 'public'
		ORDER BY table_name
	`).Rows()
	
	if err != nil {
		log.Printf("Error querying archived tables: %v", err)
	} else {
		defer archivedTableRows.Close()
		for archivedTableRows.Next() {
			var tableName string
			archivedTableRows.Scan(&tableName)
			
			var count int64
			db.Raw("SELECT COUNT(*) FROM " + tableName).Scan(&count)
			fmt.Printf("   📦 %s - %d records\n", tableName, count)
		}
	}

	// Check for SSOT migration status
	fmt.Println("\n🔄 SSOT Migration Status:")
	
	// Check migration records
	migrationRows, err := db.Raw(`
		SELECT migration_id, description, applied_at 
		FROM migration_records 
		WHERE migration_id LIKE '%ssot%' OR description LIKE '%SSOT%'
		ORDER BY applied_at DESC
	`).Rows()
	
	if err != nil {
		fmt.Printf("   ⚠️ Could not check migration records: %v\n", err)
	} else {
		defer migrationRows.Close()
		found := false
		for migrationRows.Next() {
			found = true
			var migrationID, description, appliedAt string
			migrationRows.Scan(&migrationID, &description, &appliedAt)
			fmt.Printf("   ✅ %s: %s (Applied: %s)\n", migrationID, description, appliedAt)
		}
		if !found {
			fmt.Println("   ❌ No SSOT migration records found")
		}
	}

	// Check current configuration
	fmt.Println("\n⚙️ Current System Configuration:")
	
	// Check if there are any references to SSOT in tables
	var ssotRefs int64
	db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns 
		WHERE column_name LIKE '%ssot%' OR column_name LIKE '%unified%'
	`).Scan(&ssotRefs)
	fmt.Printf("   SSOT-related columns in database: %d\n", ssotRefs)

	// Summary
	fmt.Println("\n📊 Summary:")
	var totalJournalEntries, totalJournalLines int64
	
	// Count all journal entries across all tables
	db.Raw("SELECT COUNT(*) FROM journal_entries").Scan(&totalJournalEntries)
	db.Raw("SELECT COUNT(*) FROM journal_lines").Scan(&totalJournalLines)
	
	fmt.Printf("   Total Journal Entries: %d\n", totalJournalEntries)
	fmt.Printf("   Total Journal Lines: %d\n", totalJournalLines)
	
	// Check if SSOT system is active
	var ssotActive bool
	var ssotEntries, ssotLines int64
	
	// Check for SSOT unified tables
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'ssot_unified_journal_entries')").Scan(&ssotActive)
	if ssotActive {
		db.Raw("SELECT COUNT(*) FROM ssot_unified_journal_entries").Scan(&ssotEntries)
		db.Raw("SELECT COUNT(*) FROM ssot_unified_journal_lines").Scan(&ssotLines)
		fmt.Printf("   SSOT System: ✅ ACTIVE (%d entries, %d lines)\n", ssotEntries, ssotLines)
	} else {
		fmt.Printf("   SSOT System: ❌ NOT FOUND - Using regular journal tables\n")
	}

	fmt.Println("\n✅ Journal system analysis completed!")
}