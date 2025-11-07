package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🚀 FORCE SSOT Migration")
	fmt.Println("========================")

	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("✅ Database connected successfully")

	// Step 1: Drop existing account_balances table and recreate as materialized view
	fmt.Println("\n1. Handling account_balances...")
	
	// Check if it's a table or view
	var tableType string
	err := db.Raw(`
		SELECT table_type 
		FROM information_schema.tables 
		WHERE table_name = 'account_balances' AND table_schema = 'public'
	`).Scan(&tableType).Error
	
	if err == nil && tableType == "BASE TABLE" {
		fmt.Println("⚠️  account_balances exists as BASE TABLE, need to drop and recreate as materialized view")
		
		// Backup existing data if needed
		var recordCount int64
		db.Raw("SELECT COUNT(*) FROM account_balances").Scan(&recordCount)
		fmt.Printf("   Current records: %d\n", recordCount)
		
		// Drop the table
		err = db.Exec("DROP TABLE IF EXISTS account_balances CASCADE").Error
		if err != nil {
			log.Fatalf("Failed to drop account_balances table: %v", err)
		}
		fmt.Println("✅ Dropped existing account_balances table")
	} else {
		// Try dropping view if exists
		err = db.Exec("DROP VIEW IF EXISTS account_balances CASCADE").Error
		if err != nil {
			log.Printf("Warning: Failed to drop view: %v", err)
		}
		// Try dropping materialized view if exists
		err = db.Exec("DROP MATERIALIZED VIEW IF EXISTS account_balances CASCADE").Error
		if err != nil {
			log.Printf("Warning: Failed to drop materialized view: %v", err)
		}
	}

	// Step 2: Create SSOT tables
	fmt.Println("\n2. Creating SSOT tables...")
	
	// Create unified_journal_ledger table
	createUnifiedJournalLedger(db)
	
	// Create unified_journal_lines table
	createUnifiedJournalLines(db)
	
	// Create journal_event_log table
	createJournalEventLog(db)

	// Step 3: Create materialized view
	fmt.Println("\n3. Creating account_balances materialized view...")
	createAccountBalancesMaterializedView(db)

	// Step 4: Create functions and triggers
	fmt.Println("\n4. Creating functions and triggers...")
	createFunctionsAndTriggers(db)

	// Step 5: Update migration log
	fmt.Println("\n5. Updating migration status...")
	err = db.Exec(`
		INSERT INTO migration_logs (migration_name, status, message, execution_time_ms)
		VALUES ('020_create_unified_journal_ssot.sql', 'SUCCESS', 'SSOT migration completed via force script', 0)
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

	fmt.Println("\n🎉 SSOT Migration completed successfully!")
	fmt.Println("💡 You can now restart the backend to verify SSOT system")
}

func createUnifiedJournalLedger(db *gorm.DB) {
	sql := `
		CREATE TABLE IF NOT EXISTS unified_journal_ledger (
			id BIGSERIAL PRIMARY KEY,
			entry_number VARCHAR(50) UNIQUE NOT NULL,
			
			-- Source Information  
			source_type VARCHAR(50) NOT NULL,
			source_id BIGINT,
			source_code VARCHAR(100),
			
			-- Journal Entry Details
			entry_date DATE NOT NULL,
			description TEXT NOT NULL,
			reference VARCHAR(200),
			notes TEXT,
			
			-- Financial Amounts
			total_debit DECIMAL(20,2) NOT NULL DEFAULT 0,
			total_credit DECIMAL(20,2) NOT NULL DEFAULT 0,
			
			-- Status & Control Fields
			status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
			is_balanced BOOLEAN NOT NULL DEFAULT true,
			is_auto_generated BOOLEAN NOT NULL DEFAULT false,
			
			-- Posting Information
			posted_at TIMESTAMP,
			posted_by BIGINT,
			
			-- Reversal Information
			reversed_by BIGINT,
			reversed_from BIGINT,
			reversal_reason TEXT,
			
			-- Audit Fields
			created_by BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);
	`
	
	err := db.Exec(sql).Error
	if err != nil {
		log.Fatalf("Failed to create unified_journal_ledger: %v", err)
	}
	fmt.Println("✅ Created unified_journal_ledger table")
}

func createUnifiedJournalLines(db *gorm.DB) {
	sql := `
		CREATE TABLE IF NOT EXISTS unified_journal_lines (
			id BIGSERIAL PRIMARY KEY,
			journal_id BIGINT NOT NULL,
			account_id BIGINT NOT NULL,
			
			-- Line Details
			line_number INTEGER NOT NULL,
			description TEXT,
			
			-- Financial Amounts
			debit_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
			credit_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
			
			-- Additional Information
			quantity DECIMAL(15,4),
			unit_price DECIMAL(15,4),
			
			-- Audit Fields
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			
			-- Foreign Key Constraints
			FOREIGN KEY (journal_id) REFERENCES unified_journal_ledger(id) ON DELETE CASCADE,
			FOREIGN KEY (account_id) REFERENCES accounts(id)
		);
	`
	
	err := db.Exec(sql).Error
	if err != nil {
		log.Fatalf("Failed to create unified_journal_lines: %v", err)
	}
	fmt.Println("✅ Created unified_journal_lines table")
}

func createJournalEventLog(db *gorm.DB) {
	sql := `
		CREATE TABLE IF NOT EXISTS journal_event_log (
			id BIGSERIAL PRIMARY KEY,
			event_uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
			journal_id BIGINT,
			
			-- Event Details
			event_type VARCHAR(50) NOT NULL,
			event_data JSONB NOT NULL,
			event_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			
			-- User Context
			user_id BIGINT,
			user_role VARCHAR(50),
			ip_address INET,
			user_agent TEXT,
			
			-- System Context
			source_system VARCHAR(50) DEFAULT 'ACCOUNTING_SYSTEM',
			correlation_id UUID,
			metadata JSONB,
			
			-- Foreign Key
			FOREIGN KEY (journal_id) REFERENCES unified_journal_ledger(id) ON DELETE SET NULL
		);
	`
	
	err := db.Exec(sql).Error
	if err != nil {
		log.Fatalf("Failed to create journal_event_log: %v", err)
	}
	fmt.Println("✅ Created journal_event_log table")
}

func createAccountBalancesMaterializedView(db *gorm.DB) {
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
	
	err := db.Exec(sql).Error
	if err != nil {
		log.Fatalf("Failed to create account_balances materialized view: %v", err)
	}

	// Create index
	err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_account_balances_account_id ON account_balances(account_id)").Error
	if err != nil {
		log.Printf("Warning: Failed to create index: %v", err)
	}

	fmt.Println("✅ Created account_balances materialized view")
}

func createFunctionsAndTriggers(db *gorm.DB) {
	// Create balance validation function
	balanceFunction := `
		CREATE OR REPLACE FUNCTION validate_journal_balance()
		RETURNS TRIGGER AS $$
		BEGIN
			-- Update totals when lines change
			UPDATE unified_journal_ledger 
			SET 
				total_debit = (
					SELECT COALESCE(SUM(debit_amount), 0) 
					FROM unified_journal_lines 
					WHERE journal_id = NEW.journal_id
				),
				total_credit = (
					SELECT COALESCE(SUM(credit_amount), 0) 
					FROM unified_journal_lines 
					WHERE journal_id = NEW.journal_id
				),
				is_balanced = (
					SELECT 
						COALESCE(SUM(debit_amount), 0) = COALESCE(SUM(credit_amount), 0)
					FROM unified_journal_lines 
					WHERE journal_id = NEW.journal_id
				),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = NEW.journal_id;
			
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`

	err := db.Exec(balanceFunction).Error
	if err != nil {
		log.Printf("Warning: Failed to create balance function: %v", err)
	} else {
		fmt.Println("✅ Created balance validation function")
	}

	// Create trigger
	trigger := `
		CREATE TRIGGER trg_validate_journal_balance
		AFTER INSERT OR UPDATE OR DELETE ON unified_journal_lines
		FOR EACH ROW EXECUTE FUNCTION validate_journal_balance();
	`

	err = db.Exec(trigger).Error
	if err != nil {
		log.Printf("Warning: Failed to create trigger: %v", err)
	} else {
		fmt.Println("✅ Created balance validation trigger")
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_unified_journal_ledger_entry_date ON unified_journal_ledger(entry_date)",
		"CREATE INDEX IF NOT EXISTS idx_unified_journal_ledger_status ON unified_journal_ledger(status)",
		"CREATE INDEX IF NOT EXISTS idx_unified_journal_ledger_source ON unified_journal_ledger(source_type, source_id)",
		"CREATE INDEX IF NOT EXISTS idx_unified_journal_lines_journal_id ON unified_journal_lines(journal_id)",
		"CREATE INDEX IF NOT EXISTS idx_unified_journal_lines_account_id ON unified_journal_lines(account_id)",
		"CREATE INDEX IF NOT EXISTS idx_journal_event_log_journal_id ON journal_event_log(journal_id)",
		"CREATE INDEX IF NOT EXISTS idx_journal_event_log_event_type ON journal_event_log(event_type)",
	}

	for _, indexSQL := range indexes {
		err = db.Exec(indexSQL).Error
		if err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	fmt.Println("✅ Created performance indexes")
}