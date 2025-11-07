package main

import (
	"fmt"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("📊 SSOT Database Status Check")
	fmt.Println("=============================")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully\n")

	// Check SSOT tables
	checkSSOTTables(db)

	// Check old tables (archived)
	checkArchivedTables(db)

	// Check data counts
	checkDataCounts(db)

	// Check migration logs
	checkMigrationLogs(db)

	// Check account balances view
	checkAccountBalances(db)

	fmt.Println("\n🎯 SSOT Status Summary")
	fmt.Println("=====================")
	printSSOTSummary(db)
}

func checkSSOTTables(db *gorm.DB) {
	fmt.Println("🔍 Checking SSOT Tables:")
	
	tables := []string{
		"unified_journal_ledger",
		"unified_journal_lines", 
		"journal_event_log",
	}

	for _, table := range tables {
		var exists bool
		db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists)
		
		if exists {
			var count int64
			db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			fmt.Printf("   ✅ %-25s exists (%d records)\n", table, count)
		} else {
			fmt.Printf("   ❌ %-25s missing\n", table)
		}
	}
}

func checkArchivedTables(db *gorm.DB) {
	fmt.Println("\n📚 Checking Archived Tables:")
	
	// Find archived tables
	var archivedTables []string
	db.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_name LIKE '%_archived_%'
		ORDER BY table_name
	`).Scan(&archivedTables)

	if len(archivedTables) == 0 {
		fmt.Println("   ℹ️  No archived tables found")
	} else {
		for _, table := range archivedTables {
			var count int64
			db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			fmt.Printf("   📦 %-35s (%d records)\n", table, count)
		}
	}
}

func checkDataCounts(db *gorm.DB) {
	fmt.Println("\n📈 Data Counts:")

	// SSOT entries
	var ssotEntries int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&ssotEntries)
	fmt.Printf("   📝 SSOT Journal Entries:      %d\n", ssotEntries)

	// SSOT lines
	var ssotLines int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_lines").Scan(&ssotLines)
	fmt.Printf("   📋 SSOT Journal Lines:       %d\n", ssotLines)

	// Event logs
	var eventLogs int64
	db.Raw("SELECT COUNT(*) FROM journal_event_log").Scan(&eventLogs)
	fmt.Printf("   📊 Event Logs:               %d\n", eventLogs)

	// Migrated entries
	var migratedCount int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger WHERE source_type = 'MIGRATED'").Scan(&migratedCount)
	fmt.Printf("   🔄 Migrated Entries:         %d\n", migratedCount)

	// Posted entries
	var postedCount int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger WHERE status = 'POSTED'").Scan(&postedCount)
	fmt.Printf("   ✅ Posted Entries:           %d\n", postedCount)

	// Draft entries
	var draftCount int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger WHERE status = 'DRAFT'").Scan(&draftCount)
	fmt.Printf("   📄 Draft Entries:            %d\n", draftCount)
}

func checkMigrationLogs(db *gorm.DB) {
	fmt.Println("\n📋 Migration Logs:")

	var migrationLogs []struct {
		MigrationName     string
		Status           string
		Message          string
		ExecutedAt       string
		ExecutionTimeMs  int64
	}

	err := db.Raw(`
		SELECT migration_name, status, message, 
		       TO_CHAR(executed_at, 'YYYY-MM-DD HH24:MI:SS') as executed_at,
		       execution_time_ms
		FROM migration_logs 
		WHERE migration_name LIKE '%ssot%' OR migration_name LIKE '%journal%'
		ORDER BY executed_at DESC
		LIMIT 10
	`).Scan(&migrationLogs).Error

	if err != nil {
		fmt.Printf("   ⚠️  Could not read migration logs: %v\n", err)
		return
	}

	if len(migrationLogs) == 0 {
		fmt.Println("   ℹ️  No SSOT migration logs found")
	} else {
		for _, log := range migrationLogs {
			status := "❌"
			if log.Status == "SUCCESS" {
				status = "✅"
			} else if log.Status == "PENDING" {
				status = "🟡"
			}
			
			fmt.Printf("   %s %-30s %s (%s)\n", 
				status, log.MigrationName, log.Status, log.ExecutedAt)
			if log.Message != "" {
				fmt.Printf("      └─ %s\n", log.Message)
			}
		}
	}
}

func checkAccountBalances(db *gorm.DB) {
	fmt.Println("\n💰 Account Balances:")

	// Check if account_balances exists
	var isView, isTable, isMaterializedView bool
	
	// Check if it's a view
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.views 
			WHERE table_name = 'account_balances'
		)
	`).Scan(&isView)

	// Check if it's a table
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'account_balances'
			AND table_type = 'BASE TABLE'
		)
	`).Scan(&isTable)

	// Check if it's a materialized view (PostgreSQL specific)
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_matviews
			WHERE matviewname = 'account_balances'
		)
	`).Scan(&isMaterializedView)

	if isMaterializedView {
		var count int64
		db.Raw("SELECT COUNT(*) FROM account_balances").Scan(&count)
		fmt.Printf("   ✅ account_balances exists as MATERIALIZED VIEW (%d accounts)\n", count)
	} else if isView {
		var count int64
		db.Raw("SELECT COUNT(*) FROM account_balances").Scan(&count)
		fmt.Printf("   🔍 account_balances exists as VIEW (%d accounts)\n", count)
	} else if isTable {
		var count int64
		db.Raw("SELECT COUNT(*) FROM account_balances").Scan(&count)
		fmt.Printf("   📋 account_balances exists as TABLE (%d accounts)\n", count)
		fmt.Println("   ⚠️  Should be MATERIALIZED VIEW for better performance")
	} else {
		fmt.Println("   ❌ account_balances does not exist")
	}
}

func printSSOTSummary(db *gorm.DB) {
	// Overall status check
	var hasAllTables bool = true
	var tablesStatus []string

	tables := []string{"unified_journal_ledger", "unified_journal_lines", "journal_event_log"}
	for _, table := range tables {
		var exists bool
		db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists)
		if !exists {
			hasAllTables = false
			tablesStatus = append(tablesStatus, fmt.Sprintf("❌ %s missing", table))
		}
	}

	// Check account balances
	var hasAccountBalances bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'account_balances'
		) OR EXISTS (
			SELECT 1 FROM pg_matviews
			WHERE matviewname = 'account_balances'
		)
	`).Scan(&hasAccountBalances)

	// Count data
	var totalEntries, totalLines int64
	db.Raw("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&totalEntries)
	db.Raw("SELECT COUNT(*) FROM unified_journal_lines").Scan(&totalLines)

	fmt.Println("Overall Status:")
	if hasAllTables && hasAccountBalances && totalEntries > 0 {
		fmt.Println("   🎉 SSOT is FULLY OPERATIONAL")
		fmt.Printf("   📊 %d journal entries, %d lines\n", totalEntries, totalLines)
		fmt.Println("   ✅ Ready for production use")
	} else {
		fmt.Println("   ⚠️  SSOT setup is INCOMPLETE")
		if !hasAllTables {
			for _, status := range tablesStatus {
				fmt.Printf("   %s\n", status)
			}
		}
		if !hasAccountBalances {
			fmt.Println("   ❌ account_balances missing")
		}
		if totalEntries == 0 {
			fmt.Println("   ⚠️  No journal entries found")
		}
		fmt.Println("   💡 Run migration scripts to complete setup")
	}

	fmt.Println("\nNext Steps:")
	if hasAllTables && hasAccountBalances && totalEntries > 0 {
		fmt.Println("   • Test API endpoints: make test-ssot")
		fmt.Println("   • Update frontend to use SSOT endpoints")
		fmt.Println("   • Monitor system performance") 
		fmt.Println("   • Archive old tables after verification")
	} else {
		fmt.Println("   • Run: make migrate-ssot")
		fmt.Println("   • Run: make cleanup-models") 
		fmt.Println("   • Run: make update-routes")
		fmt.Println("   • Or run all: make full-migration")
	}
}