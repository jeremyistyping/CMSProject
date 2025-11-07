package main

import (
	"fmt"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔄 Migrate Old Journal to SSOT (OPTIONAL)")
	fmt.Println("==========================================")
	
	fmt.Println("⚠️  WARNING: This script will migrate old journal_entries to the new SSOT system")
	fmt.Println("⚠️  This is OPTIONAL - old data will remain safe even if you don't run this")
	fmt.Println("⚠️  Only run this if you want to consolidate all journal data in SSOT")
	
	fmt.Println("\n🔍 Preview of migration plan:")
	fmt.Println("1. Read old journal_entries table")
	fmt.Println("2. Create corresponding entries in unified_journal_ledger")
	fmt.Println("3. Convert lines to unified_journal_lines format")
	fmt.Println("4. Create audit trail in journal_event_log")
	fmt.Println("5. Original data remains untouched for safety")
	
	fmt.Println("\n💡 Do you want to proceed? (This is just a preview script)")
	fmt.Println("   To actually run migration, uncomment the execution code below")
	
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Check old journal entries
	var oldJournalCount int64
	db.Raw("SELECT COUNT(*) FROM journal_entries").Scan(&oldJournalCount)
	fmt.Printf("\n📊 Found %d old journal entries to potentially migrate\n", oldJournalCount)
	
	if oldJournalCount > 0 {
		fmt.Println("\nSample of old journal entries:")
		rows, err := db.Raw(`
			SELECT id, description, date, total_amount, status 
			FROM journal_entries 
			ORDER BY date DESC 
			LIMIT 5
		`).Rows()
		
		if err == nil {
			defer rows.Close()
			fmt.Println("ID | Description | Date | Amount | Status")
			fmt.Println("---|-------------|------|--------|-------")
			for rows.Next() {
				var id uint64
				var desc, status string
				var date time.Time  
				var amount float64
				rows.Scan(&id, &desc, &date, &amount, &status)
				fmt.Printf("%d | %s | %s | %.2f | %s\n", 
					id, desc, date.Format("2006-01-02"), amount, status)
			}
		}
	}
	
	fmt.Println("\n📋 Migration Strategy Options:")
	fmt.Println("===============================")
	fmt.Println("Option 1: Keep Both Systems (RECOMMENDED)")
	fmt.Println("  ✅ Old data stays in journal_entries for historical reference") 
	fmt.Println("  ✅ New transactions use SSOT unified_journal_ledger")
	fmt.Println("  ✅ No risk, easy rollback")
	fmt.Println("  ✅ Reports can query both if needed")
	
	fmt.Println("\nOption 2: Migrate All Data (ADVANCED)")
	fmt.Println("  ⚠️  Move old journal_entries → unified_journal_ledger")
	fmt.Println("  ⚠️  Requires careful mapping of fields")
	fmt.Println("  ⚠️  Need to update reports to use new schema")
	fmt.Println("  ⚠️  More complex but fully unified")
	
	fmt.Println("\nOption 3: Hybrid Reporting (BALANCED)")
	fmt.Println("  ✅ Keep old data as-is")
	fmt.Println("  ✅ Use SSOT for new entries")
	fmt.Println("  ✅ Create unified reporting views that combine both")
	fmt.Println("  ✅ Best of both worlds")
	
	/* 
	UNCOMMENT THIS SECTION TO ACTUALLY RUN MIGRATION:
	
	fmt.Println("\n🔄 Running actual migration...")
	
	// Example migration logic:
	rows, err := db.Raw("SELECT * FROM journal_entries").Rows()
	if err != nil {
		log.Fatalf("Failed to read old entries: %v", err)
	}
	defer rows.Close()
	
	migratedCount := 0
	for rows.Next() {
		// Map old journal_entries fields to unified_journal_ledger
		// Create corresponding records in SSOT tables
		// This would require careful field mapping
		migratedCount++
	}
	
	fmt.Printf("✅ Migrated %d journal entries to SSOT system\n", migratedCount)
	*/
	
	fmt.Println("\n💡 RECOMMENDATION:")
	fmt.Println("For now, keep both systems running. This gives you:")
	fmt.Println("• Zero risk to existing data")
	fmt.Println("• Time to test SSOT thoroughly") 
	fmt.Println("• Ability to compare old vs new system")
	fmt.Println("• Easy rollback if needed")
	
	fmt.Println("\n✅ Migration planning completed!")
	fmt.Println("📱 Next: Test SSOT with new journal entries via API")
}