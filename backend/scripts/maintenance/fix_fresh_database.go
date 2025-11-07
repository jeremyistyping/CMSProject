package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 DATABASE FRESH INSTALL FIX V2")
	fmt.Println("===============================")
	fmt.Println()

	fmt.Println("⚠️  PERINGATAN: Script ini akan memperbaiki database yang baru dibuat.")
	fmt.Println("✅ Yang akan dilakukan:")
	fmt.Println("   - Fix missing columns in transactions table")
	fmt.Println("   - Create SSOT journal system tables (if needed)")
	fmt.Println("   - Setup materialized views")
	fmt.Println("   - Verify database structure")
	fmt.Println()

	fmt.Print("Lanjutkan? (ketik 'ya' untuk konfirmasi): ")
	var confirm string
	fmt.Scanln(&confirm)
	
	if confirm != "ya" && confirm != "y" {
		fmt.Println("Fix dibatalkan.")
		return
	}

	db := database.ConnectDB()
	if db == nil {
		log.Fatal("❌ Gagal koneksi ke database")
	}

	fmt.Println("🔗 Berhasil terhubung ke database")

	// Step 1: Fix transactions table structure (PRIORITY 1)
	fmt.Println("\n🔧 Step 1: Fixing transactions table structure...")
	if err := fixTransactionsTable(db); err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Println("   ✅ Transactions table berhasil diperbaiki")
	}

	// Step 2: Create SSOT tables if needed (with proper error handling)
	fmt.Println("\n🏗️ Step 2: Ensuring SSOT journal tables exist...")
	if err := createSSOTTablesIfNeeded(db); err != nil {
		fmt.Printf("   ⚠️ Warning: %v\n", err)
	} else {
		fmt.Println("   ✅ SSOT tables ready")
	}

	// Step 3: Create materialized view
	fmt.Println("\n📊 Step 3: Setting up materialized view...")
	if err := createMaterializedView(db); err != nil {
		fmt.Printf("   ⚠️ Warning: %v\n", err)
	} else {
		fmt.Println("   ✅ Materialized view ready")
	}

	// Step 4: Verify the fix
	fmt.Println("\n🧪 Step 4: Verifying database structure...")
	if err := verifyDatabaseFix(db); err != nil {
		fmt.Printf("   ❌ Verification failed: %v\n", err)
	} else {
		fmt.Println("   ✅ All structures verified")
	}

	fmt.Println("\n🎉 DATABASE FIX COMPLETED!")
	fmt.Println("✅ Error 'column does not exist' should be resolved")
	fmt.Println("✅ Database ready for use")
}

func fixTransactionsTable(db *gorm.DB) error {
	// Check if debit_amount column exists
	var columnExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'transactions' 
			AND column_name = 'debit_amount'
		)
	`).Scan(&columnExists).Error

	if err != nil {
		return fmt.Errorf("failed to check transactions table: %v", err)
	}

	if !columnExists {
		fmt.Println("   🔧 Adding missing columns to transactions table...")
		
		// Add columns one by one to avoid transaction issues
		queries := []string{
			"ALTER TABLE transactions ADD COLUMN IF NOT EXISTS debit_amount DECIMAL(20,2) DEFAULT 0",
			"ALTER TABLE transactions ADD COLUMN IF NOT EXISTS credit_amount DECIMAL(20,2) DEFAULT 0",
		}

		for i, query := range queries {
			err := db.Exec(query).Error
			if err != nil {
				return fmt.Errorf("failed to add column (step %d): %v", i+1, err)
			}
		}
		fmt.Println("   ✅ Missing columns added successfully")
	} else {
		fmt.Println("   ✅ Transactions table already has required columns")
	}

	return nil
}

func createSSOTTablesIfNeeded(db *gorm.DB) error {
	// Check if unified_journal_ledger exists
	var tableExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'unified_journal_ledger'
		)
	`).Scan(&tableExists).Error

	if err != nil {
		return fmt.Errorf("failed to check SSOT tables: %v", err)
	}

	if tableExists {
		fmt.Println("   ℹ️ SSOT tables already exist, skipping creation")
		return nil
	}

	fmt.Println("   🏗️ Creating SSOT journal tables...")

	// Create tables one by one
	tables := map[string]string{
		"unified_journal_ledger": `
			CREATE TABLE unified_journal_ledger (
				id BIGSERIAL PRIMARY KEY,
				transaction_uuid UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
				entry_number VARCHAR(50),
				source_type VARCHAR(50) NOT NULL DEFAULT 'MANUAL',
				source_id BIGINT,
				source_code VARCHAR(100),
				entry_date DATE NOT NULL DEFAULT CURRENT_DATE,
				description TEXT NOT NULL DEFAULT '',
				reference VARCHAR(100),
				notes TEXT,
				total_debit DECIMAL(20,2) NOT NULL DEFAULT 0,
				total_credit DECIMAL(20,2) NOT NULL DEFAULT 0,
				status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
				is_balanced BOOLEAN NOT NULL DEFAULT TRUE,
				is_auto_generated BOOLEAN NOT NULL DEFAULT false,
				posted_at TIMESTAMPTZ,
				posted_by BIGINT,
				reversed_by BIGINT,
				reversed_from BIGINT,
				created_by BIGINT NOT NULL DEFAULT 1,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				deleted_at TIMESTAMPTZ
			)`,
		"unified_journal_lines": `
			CREATE TABLE unified_journal_lines (
				id BIGSERIAL PRIMARY KEY,
				journal_id BIGINT NOT NULL,
				account_id BIGINT NOT NULL,
				line_number SMALLINT NOT NULL DEFAULT 1,
				description TEXT,
				debit_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
				credit_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
				quantity DECIMAL(15,4),
				unit_price DECIMAL(15,4),
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
		"journal_event_log": `
			CREATE TABLE journal_event_log (
				id BIGSERIAL PRIMARY KEY,
				event_uuid UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
				journal_id BIGINT,
				event_type VARCHAR(50) NOT NULL DEFAULT 'CREATED',
				event_data JSONB NOT NULL DEFAULT '{}',
				event_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				user_id BIGINT,
				user_role VARCHAR(50),
				ip_address INET,
				user_agent TEXT,
				source_system VARCHAR(50) DEFAULT 'ACCOUNTING_SYSTEM',
				correlation_id UUID,
				metadata JSONB
			)`,
	}

	for tableName, sql := range tables {
		fmt.Printf("   🔧 Creating table: %s...\n", tableName)
		err := db.Exec(sql).Error
		if err != nil {
			return fmt.Errorf("failed to create table %s: %v", tableName, err)
		}
	}

	// Add constraints separately to avoid issues
	fmt.Println("   🔧 Adding table constraints...")
	constraints := []string{
		"ALTER TABLE unified_journal_lines ADD CONSTRAINT fk_journal_lines_journal FOREIGN KEY (journal_id) REFERENCES unified_journal_ledger(id) ON DELETE CASCADE",
		"ALTER TABLE unified_journal_lines ADD CONSTRAINT fk_journal_lines_account FOREIGN KEY (account_id) REFERENCES accounts(id)",
		"ALTER TABLE journal_event_log ADD CONSTRAINT fk_event_log_journal FOREIGN KEY (journal_id) REFERENCES unified_journal_ledger(id)",
	}

	for _, constraint := range constraints {
		err := db.Exec(constraint).Error
		if err != nil {
			fmt.Printf("   ⚠️ Warning adding constraint: %v\n", err)
			// Continue anyway - constraints are nice to have but not critical
		}
	}

	return nil
}

func createMaterializedView(db *gorm.DB) error {
	fmt.Println("   🔍 Checking existing materialized view...")
	
	// Drop existing view if exists
	db.Exec("DROP MATERIALIZED VIEW IF EXISTS account_balances CASCADE")
	db.Exec("DROP VIEW IF EXISTS account_balances CASCADE")

	// Check if SSOT tables exist
	var ssoTExists bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_ledger')").Scan(&ssoTExists).Error
	if err != nil {
		return fmt.Errorf("failed to check SSOT tables: %v", err)
	}

	var createQuery string
	if ssoTExists {
		fmt.Println("   📊 Creating SSOT-compatible materialized view...")
		createQuery = `
		CREATE MATERIALIZED VIEW account_balances AS
		WITH journal_totals AS (
			SELECT 
				jl.account_id,
				COALESCE(SUM(jl.debit_amount), 0) as total_debits,
				COALESCE(SUM(jl.credit_amount), 0) as total_credits,
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
			COALESCE(a.category, 'OTHER') as account_category,
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
			COALESCE(a.is_active, true) as is_active,
			COALESCE(a.is_header, false) as is_header
		FROM accounts a
		LEFT JOIN journal_totals jt ON a.id = jt.account_id
		WHERE a.deleted_at IS NULL
		`
	} else {
		fmt.Println("   📊 Creating fallback materialized view...")
		createQuery = `
		CREATE MATERIALIZED VIEW account_balances AS
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(a.category, 'OTHER') as account_category,
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN 'DEBIT'
				WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 'CREDIT'
				ELSE 'DEBIT'
			END as normal_balance,
			0 as total_debits,
			0 as total_credits,
			0 as transaction_count,
			NULL as last_transaction_date,
			COALESCE(a.balance, 0) as current_balance,
			NOW() as last_updated,
			COALESCE(a.is_active, true) as is_active,
			COALESCE(a.is_header, false) as is_header
		FROM accounts a
		WHERE a.deleted_at IS NULL
		`
	}

	err = db.Exec(createQuery).Error
	if err != nil {
		return fmt.Errorf("failed to create materialized view: %v", err)
	}

	// Create index
	err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_account_balances_account_id ON account_balances(account_id)").Error
	if err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}

	// Initial refresh
	err = db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error
	if err != nil {
		return fmt.Errorf("failed to refresh materialized view: %v", err)
	}

	return nil
}

func verifyDatabaseFix(db *gorm.DB) error {
	// Test 1: Check transactions table has required columns
	var debitExists, creditExists bool
	
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'transactions' AND column_name = 'debit_amount'
		)
	`).Scan(&debitExists).Error
	if err != nil {
		return fmt.Errorf("failed to verify debit_amount column: %v", err)
	}

	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'transactions' AND column_name = 'credit_amount'
		)
	`).Scan(&creditExists).Error
	if err != nil {
		return fmt.Errorf("failed to verify credit_amount column: %v", err)
	}

	if !debitExists || !creditExists {
		return fmt.Errorf("transactions table still missing required columns")
	}
	fmt.Println("   ✅ Transactions table has required columns")

	// Test 2: Check materialized view
	var viewCount int64
	err = db.Raw("SELECT COUNT(*) FROM account_balances").Scan(&viewCount).Error
	if err != nil {
		return fmt.Errorf("materialized view not working: %v", err)
	}
	fmt.Printf("   ✅ Materialized view ready: %d accounts\n", viewCount)

	// Test 3: Test the original query that was failing
	var testCount int64
	err = db.Raw("SELECT COUNT(*) FROM transactions WHERE account_id > 0").Scan(&testCount).Error
	if err != nil {
		return fmt.Errorf("basic transactions query failed: %v", err)
	}
	fmt.Printf("   ✅ Transactions table query works: %d records\n", testCount)

	// Test 4: Simulate the original balance calculation query
	err = db.Raw(`
		SELECT COALESCE(SUM(debit_amount), 0) as debit_sum, 
		       COALESCE(SUM(credit_amount), 0) as credit_sum 
		FROM transactions 
		WHERE account_id = 1 AND deleted_at IS NULL
	`).Scan(&struct {
		DebitSum  float64 `json:"debit_sum"`
		CreditSum float64 `json:"credit_sum"`
	}{}).Error

	if err != nil {
		return fmt.Errorf("balance calculation query failed: %v", err)
	}
	fmt.Println("   ✅ Balance calculation query works")

	return nil
}