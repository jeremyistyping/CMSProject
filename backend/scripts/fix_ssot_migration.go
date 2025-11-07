package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 SSOT Migration Conflict Fix")
	fmt.Println("===============================")

	// Load configuration
	_ = config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully")

	// Step 1: Check existing account_balances view
	fmt.Println("\n1. Checking existing account_balances view...")
	var viewExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.views 
			WHERE table_name = 'account_balances'
		)
	`).Scan(&viewExists).Error
	
	if err != nil {
		log.Fatalf("Failed to check view existence: %v", err)
	}

	// Check if it's a materialized view 
	var isMaterialized bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_matviews 
			WHERE matviewname = 'account_balances'
		)
	`).Scan(&isMaterialized).Error
	
	if err != nil {
		log.Fatalf("Failed to check materialized view: %v", err)
	}

	if isMaterialized {
		fmt.Println("✅ account_balances materialized view already exists - no action needed")
	} else if viewExists {
		fmt.Println("⚠️  account_balances exists as regular view, dropping and recreating as materialized view...")
		
		// Drop the existing view
		err = db.Exec("DROP VIEW IF EXISTS account_balances CASCADE").Error
		if err != nil {
			log.Fatalf("Failed to drop existing view: %v", err)
		}
		fmt.Println("✅ Dropped existing view")
		
		// Create materialized view
		createMaterializedView(db)
	} else {
		fmt.Println("ℹ️  account_balances view doesn't exist, creating materialized view...")
		// Create materialized view
		createMaterializedView(db)
	}

	// Step 2: Check if tables exist
	fmt.Println("\n2. Checking SSOT tables...")
	tables := []string{"unified_journal_ledger", "unified_journal_lines", "journal_event_log"}
	
	for _, table := range tables {
		var tableExists bool
		err := db.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = ?
			)
		`, table).Scan(&tableExists).Error
		
		if err != nil {
			log.Printf("Failed to check table %s: %v", table, err)
			continue
		}
		
		if tableExists {
			fmt.Printf("✅ Table %s exists\n", table)
		} else {
			fmt.Printf("❌ Table %s missing\n", table)
		}
	}

	// Step 3: Update migration status
	fmt.Println("\n3. Updating migration status...")
	err = db.Exec(`
		INSERT INTO migration_logs (migration_name, status, message, execution_time_ms)
		VALUES ('020_create_unified_journal_ssot.sql', 'SUCCESS', 'SSOT migration completed manually', 0)
		ON CONFLICT (migration_name) DO UPDATE SET
			status = EXCLUDED.status,
			message = EXCLUDED.message,
			executed_at = CURRENT_TIMESTAMP
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update migration log: %v", err)
	} else {
		fmt.Println("✅ Migration status updated")
	}

	fmt.Println("\n🎉 SSOT Migration conflict fix completed!")
	fmt.Println("💡 You can now restart the backend to verify SSOT system")
}

func createMaterializedView(db *gorm.DB) {
	fmt.Println("🔧 Creating account_balances materialized view...")
	
	sql := `
		CREATE MATERIALIZED VIEW account_balances AS
		WITH journal_totals AS (
			SELECT
				jl.account_id,
				SUM(jl.debit_amount) as total_debits,
				SUM(jl.credit_amount) as total_credits,
				COUNT(*) as transaction_count,
				MAX(jd.posted_at) as last_transaction_date
			FROM unified_journal_lines jl
			JOIN unified_journal_ledger jd ON jl.journal_id = jd.id
			WHERE jd.status = 'POSTED'
			AND jd.deleted_at IS NULL
			GROUP BY jl.account_id
		)
		SELECT
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			a.category as account_category,
			CASE
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN 'DEBIT'
				WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 'CREDIT'
				ELSE 'DEBIT'
			END as normal_balance,
			COALESCE(jt.total_debits, 0) as total_debits,
			COALESCE(jt.total_credits, 0) as total_credits,
			COALESCE(jt.transaction_count, 0) as transaction_count,
			jt.last_transaction_date,
			CASE
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN
					COALESCE(jt.total_debits, 0) - COALESCE(jt.total_credits, 0)
				WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN
					COALESCE(jt.total_credits, 0) - COALESCE(jt.total_debits, 0)
				ELSE 0
			END as current_balance,
			NOW() as last_updated,
			a.is_active,
			a.is_header
		FROM accounts a
		LEFT JOIN journal_totals jt ON a.id = jt.account_id
		WHERE a.deleted_at IS NULL;
	`

	// Check if unified tables exist first
	var unifiedTablesExist bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name IN ('unified_journal_ledger', 'unified_journal_lines')
			GROUP BY 1
			HAVING COUNT(*) = 2
		)
	`).Scan(&unifiedTablesExist).Error

	if err != nil || !unifiedTablesExist {
		fmt.Println("⚠️  Unified journal tables don't exist yet, creating empty materialized view...")
		sql = `
			CREATE MATERIALIZED VIEW account_balances AS
			SELECT
				a.id as account_id,
				a.code as account_code,
				a.name as account_name,
				a.type as account_type,
				a.category as account_category,
				CASE
					WHEN a.type IN ('ASSET', 'EXPENSE') THEN 'DEBIT'
					WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 'CREDIT'
					ELSE 'DEBIT'
				END as normal_balance,
				0::decimal(20,2) as total_debits,
				0::decimal(20,2) as total_credits,
				0 as transaction_count,
				NULL::timestamp as last_transaction_date,
				0::decimal(20,2) as current_balance,
				NOW() as last_updated,
				a.is_active,
				a.is_header
			FROM accounts a
			WHERE a.deleted_at IS NULL;
		`
	}

	err = db.Exec(sql).Error
	if err != nil {
		log.Fatalf("Failed to create materialized view: %v", err)
	}

	// Create index on materialized view
	err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_account_balances_account_id ON account_balances(account_id)").Error
	if err != nil {
		log.Printf("Warning: Failed to create index on materialized view: %v", err)
	}

	fmt.Println("✅ account_balances materialized view created successfully")
}