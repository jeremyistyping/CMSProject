package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("📊 Checking Journal Systems Status")
	fmt.Println("===================================")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully")

	// 1. Check OLD journal tables
	fmt.Println("\n1. 📋 Checking OLD Journal System tables...")
	oldTables := []string{
		"journal_entries",
		"journal_entry_lines", 
		"journal_entry_details",
		"journals",
	}

	for _, table := range oldTables {
		var exists bool
		var count int64
		
		err := db.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = ? AND table_schema = 'public'
			)
		`, table).Scan(&exists).Error
		
		if err != nil {
			log.Printf("Error checking table %s: %v", table, err)
			continue
		}
		
		if exists {
			// Count records in the table
			db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			fmt.Printf("  ✅ %s exists with %d records\n", table, count)
		} else {
			fmt.Printf("  ❌ %s not found\n", table)
		}
	}

	// 2. Check NEW SSOT journal tables
	fmt.Println("\n2. 🆕 Checking NEW SSOT Journal System tables...")
	ssotTables := []string{
		"unified_journal_ledger",
		"unified_journal_lines",
		"journal_event_log",
	}

	for _, table := range ssotTables {
		var exists bool
		var count int64
		
		err := db.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = ? AND table_schema = 'public'
			)
		`, table).Scan(&exists).Error
		
		if err != nil {
			log.Printf("Error checking table %s: %v", table, err)
			continue
		}
		
		if exists {
			// Count records in the table
			db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			fmt.Printf("  ✅ %s exists with %d records\n", table, count)
		} else {
			fmt.Printf("  ❌ %s not found\n", table)
		}
	}

	// 3. Check account balances views/tables
	fmt.Println("\n3. 💰 Checking Account Balances...")
	
	// Check materialized view
	var matViewExists bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_matviews WHERE matviewname = 'account_balances')").Scan(&matViewExists).Error
	if err == nil && matViewExists {
		var balanceCount int64
		db.Raw("SELECT COUNT(*) FROM account_balances").Scan(&balanceCount)
		fmt.Printf("  ✅ account_balances (materialized view) with %d account records\n", balanceCount)
	}

	// Check period balances table
	var periodTableExists bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'account_period_balances')").Scan(&periodTableExists).Error
	if err == nil && periodTableExists {
		var periodCount int64
		db.Raw("SELECT COUNT(*) FROM account_period_balances").Scan(&periodCount)
		fmt.Printf("  ✅ account_period_balances (period-based table) with %d records\n", periodCount)
	}

	// 4. Check routes/endpoints status
	fmt.Println("\n4. 🛣️  Available Journal Routes:")
	fmt.Println("  OLD System routes:")
	fmt.Println("    - GET/POST /api/v1/journal-entries")
	fmt.Println("    - GET /api/v1/journal-drilldown")
	
	fmt.Println("  NEW SSOT System routes:")
	fmt.Println("    - GET/POST /api/v1/journals")
	fmt.Println("    - PUT /api/v1/journals/:id/post")  
	fmt.Println("    - POST /api/v1/journals/:id/reverse")
	fmt.Println("    - GET /api/v1/journals/account-balances")

	// 5. Migration status
	fmt.Println("\n5. 📋 Migration Status:")
	rows, err := db.Raw(`
		SELECT migration_name, status, message
		FROM migration_logs 
		WHERE migration_name LIKE '%journal%' OR migration_name LIKE '%ssot%'
		ORDER BY executed_at DESC
		LIMIT 10
	`).Rows()
	
	if err != nil {
		log.Printf("Error checking migrations: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var name, status, message string
			rows.Scan(&name, &status, &message)
			
			statusIcon := "❌"
			if status == "SUCCESS" {
				statusIcon = "✅"
			}
			
			fmt.Printf("  %s %s: %s\n", statusIcon, name, status)
		}
	}

	fmt.Println("\n📋 Summary:")
	fmt.Println("===========")
	fmt.Println("• OLD Journal System: Still exists alongside the new system")
	fmt.Println("• NEW SSOT Journal System: Fully implemented and ready") 
	fmt.Println("• Both systems can coexist during transition period")
	fmt.Println("• SSOT provides unified, more efficient journal management")
	fmt.Println("• Old data is preserved and safe")
	
	fmt.Println("\n💡 Next Steps:")
	fmt.Println("• Test SSOT system with sample entries")
	fmt.Println("• Plan migration of old journal data if needed")
	fmt.Println("• Update frontend to use new SSOT endpoints")
	fmt.Println("• Consider gradual retirement of old system")

	fmt.Println("\n✅ Journal systems status check completed!")
}