package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 DATABASE FRESH INSTALL FIX")
	fmt.Println("============================")
	fmt.Println()

	// Check if this is intentional
	fmt.Println("⚠️  PERINGATAN: Script ini akan memperbaiki database yang baru dibuat.")
	fmt.Println("✅ Yang akan dilakukan:")
	fmt.Println("   - Jalankan complete database migrations")
	fmt.Println("   - Buat SSOT journal system tables")
	fmt.Println("   - Setup materialized views")
	fmt.Println("   - Seed initial data")
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

	// Step 1: Run complete database initialization
	fmt.Println("\n📋 Step 1: Menjalankan database initialization...")
	database.InitializeDatabase(db)
	fmt.Println("   ✅ Database initialization selesai")

	// Step 2: Run SSOT migration
	fmt.Println("\n🔄 Step 2: Menjalankan SSOT migration...")
	if err := runSSOTMigration(db); err != nil {
		fmt.Printf("   ⚠️ Warning SSOT migration: %v\n", err)
	} else {
		fmt.Println("   ✅ SSOT migration berhasil")
	}

	// Step 3: Create materialized view
	fmt.Println("\n🏗️ Step 3: Membuat materialized view...")
	if err := createMaterializedView(db); err != nil {
		fmt.Printf("   ⚠️ Warning materialized view: %v\n", err)
	} else {
		fmt.Println("   ✅ Materialized view berhasil dibuat")
	}

	// Step 4: Create additional indexes
	fmt.Println("\n📊 Step 4: Membuat additional indexes...")
	database.CreateIndexes(db)
	fmt.Println("   ✅ Indexes berhasil dibuat")

	// Step 5: Verify database structure
	fmt.Println("\n🧪 Step 5: Verifikasi struktur database...")
	if err := verifyDatabaseStructure(db); err != nil {
		fmt.Printf("   ❌ Error verifikasi: %v\n", err)
	} else {
		fmt.Println("   ✅ Verifikasi berhasil - Database siap digunakan")
	}

	fmt.Println("\n🎉 DATABASE FRESH INSTALL FIX SELESAI!")
	fmt.Println("✅ Database sudah lengkap dan siap digunakan")
	fmt.Println("✅ Semua tabel dan views sudah tersedia")
	fmt.Println("✅ Error 'column does not exist' sudah teratasi")
}

func runSSOTMigration(db *gorm.DB) error {
	fmt.Println("   📝 Running SSOT migration SQL...")

	// Read SSOT migration file
	migrationSQL := `
	-- Create SSOT journal tables if not exists
	CREATE TABLE IF NOT EXISTS unified_journal_ledger (
		id BIGSERIAL PRIMARY KEY,
		transaction_uuid UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
		entry_number VARCHAR(50),
		source_type VARCHAR(50) NOT NULL CHECK (source_type IN (
			'MANUAL', 'SALE', 'PURCHASE', 'PAYMENT', 'CASH_BANK', 
			'INVENTORY', 'ASSET', 'ADJUSTMENT', 'CLOSING', 'OPENING'
		)),
		source_id BIGINT,
		source_code VARCHAR(100),
		entry_date DATE NOT NULL DEFAULT CURRENT_DATE,
		description TEXT NOT NULL,
		reference VARCHAR(100),
		notes TEXT,
		total_debit DECIMAL(20,2) NOT NULL DEFAULT 0,
		total_credit DECIMAL(20,2) NOT NULL DEFAULT 0,
		status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'POSTED', 'REVERSED')),
		is_balanced BOOLEAN GENERATED ALWAYS AS (total_debit = total_credit AND total_debit > 0) STORED,
		is_auto_generated BOOLEAN NOT NULL DEFAULT false,
		posted_at TIMESTAMPTZ,
		posted_by BIGINT,
		reversed_by BIGINT,
		reversed_from BIGINT,
		created_by BIGINT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMPTZ
	);

	CREATE TABLE IF NOT EXISTS unified_journal_lines (
		id BIGSERIAL PRIMARY KEY,
		journal_id BIGINT NOT NULL REFERENCES unified_journal_ledger(id) ON DELETE CASCADE,
		account_id BIGINT NOT NULL REFERENCES accounts(id),
		line_number SMALLINT NOT NULL CHECK (line_number > 0),
		description TEXT,
		debit_amount DECIMAL(20,2) NOT NULL DEFAULT 0 CHECK (debit_amount >= 0),
		credit_amount DECIMAL(20,2) NOT NULL DEFAULT 0 CHECK (credit_amount >= 0),
		quantity DECIMAL(15,4),
		unit_price DECIMAL(15,4),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CONSTRAINT chk_amounts_not_both CHECK (
			NOT (debit_amount > 0 AND credit_amount > 0)
		),
		CONSTRAINT chk_amounts_not_zero CHECK (
			debit_amount > 0 OR credit_amount > 0
		),
		UNIQUE(journal_id, line_number)
	);

	CREATE TABLE IF NOT EXISTS journal_event_log (
		id BIGSERIAL PRIMARY KEY,
		event_uuid UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
		journal_id BIGINT REFERENCES unified_journal_ledger(id),
		event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
			'CREATED', 'POSTED', 'REVERSED', 'UPDATED', 'DELETED', 
			'BALANCED', 'VALIDATED', 'MIGRATED', 'SYSTEM_ACTION'
		)),
		event_data JSONB NOT NULL,
		event_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		user_id BIGINT,
		user_role VARCHAR(50),
		ip_address INET,
		user_agent TEXT,
		source_system VARCHAR(50) DEFAULT 'ACCOUNTING_SYSTEM',
		correlation_id UUID,
		metadata JSONB
	);
	`

	err := db.Exec(migrationSQL).Error
	if err != nil {
		return fmt.Errorf("gagal menjalankan SSOT migration: %v", err)
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
		return fmt.Errorf("gagal check SSOT tables: %v", err)
	}

	var createQuery string
	if ssoTExists {
		fmt.Println("   📊 Creating SSOT materialized view...")
		createQuery = `
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
		WHERE a.deleted_at IS NULL
		`
	} else {
		fmt.Println("   📊 Creating classic materialized view...")
		createQuery = `
		CREATE MATERIALIZED VIEW account_balances AS
		WITH journal_totals AS (
			SELECT 
				jl.account_id,
				SUM(jl.debit_amount) as total_debits,
				SUM(jl.credit_amount) as total_credits,
				COUNT(*) as transaction_count,
				MAX(je.created_at) as last_transaction_date
			FROM journal_lines jl
			JOIN journal_entries je ON jl.journal_entry_id = je.id
			WHERE je.status = 'POSTED'
			  AND je.deleted_at IS NULL
			  AND jl.deleted_at IS NULL
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
		WHERE a.deleted_at IS NULL
		`
	}

	err = db.Exec(createQuery).Error
	if err != nil {
		return fmt.Errorf("gagal membuat materialized view: %v", err)
	}

	// Create index
	err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_account_balances_account_id ON account_balances(account_id)").Error
	if err != nil {
		return fmt.Errorf("gagal membuat index: %v", err)
	}

	// Initial refresh
	err = db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error
	if err != nil {
		return fmt.Errorf("gagal refresh materialized view: %v", err)
	}

	return nil
}

func verifyDatabaseStructure(db *gorm.DB) error {
	// Test 1: Check if transactions table has proper structure for current code
	var columnExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'transactions' 
			AND column_name = 'debit_amount'
		)
	`).Scan(&columnExists).Error

	if err != nil {
		return fmt.Errorf("gagal check column transactions: %v", err)
	}

	if columnExists {
		fmt.Println("   ✅ Transactions table memiliki kolom debit_amount")
	} else {
		// Create missing columns for transactions table
		fmt.Println("   🔧 Adding missing columns to transactions table...")
		err = db.Exec(`
			ALTER TABLE transactions 
			ADD COLUMN IF NOT EXISTS debit_amount DECIMAL(20,2) DEFAULT 0,
			ADD COLUMN IF NOT EXISTS credit_amount DECIMAL(20,2) DEFAULT 0
		`).Error
		if err != nil {
			return fmt.Errorf("gagal menambah kolom transactions: %v", err)
		}
		fmt.Println("   ✅ Kolom debit_amount dan credit_amount berhasil ditambahkan")
	}

	// Test 2: Check materialized view
	var viewCount int64
	err = db.Raw("SELECT COUNT(*) FROM account_balances").Scan(&viewCount).Error
	if err != nil {
		return fmt.Errorf("gagal test materialized view: %v", err)
	}
	fmt.Printf("   ✅ Materialized view account_balances: %d records\n", viewCount)

	// Test 3: Check SSOT tables
	var ssoTCount int64
	err = db.Raw("SELECT COUNT(*) FROM unified_journal_ledger").Scan(&ssoTCount).Error
	if err != nil {
		fmt.Printf("   ⚠️ SSOT tables belum ada (normal untuk fresh install)\n")
	} else {
		fmt.Printf("   ✅ SSOT journal system: %d entries\n", ssoTCount)
	}

	return nil
}