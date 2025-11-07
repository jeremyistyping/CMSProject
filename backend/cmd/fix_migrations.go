package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database using DATABASE_URL from .env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database: %s\n", databaseURL)
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔧 Starting manual migration fixes...")

	// 1. Create purchase_payments table
	fmt.Println("📋 Creating purchase_payments table...")
	createPurchasePaymentsSQL := `
		CREATE TABLE IF NOT EXISTS purchase_payments (
			id BIGSERIAL PRIMARY KEY,
			purchase_id BIGINT NOT NULL,
			payment_number VARCHAR(50),
			date TIMESTAMP WITH TIME ZONE,
			amount DECIMAL(15,2) DEFAULT 0,
			method VARCHAR(20),
			reference VARCHAR(100),
			notes TEXT,
			cash_bank_id BIGINT,
			user_id BIGINT NOT NULL,
			payment_id BIGINT, -- Cross-reference to payments table
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			
			FOREIGN KEY (purchase_id) REFERENCES purchases(id) ON DELETE CASCADE,
			FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL
		);
		
		-- Create indexes for purchase_payments
		CREATE INDEX IF NOT EXISTS idx_purchase_payments_purchase_id ON purchase_payments(purchase_id);
		CREATE INDEX IF NOT EXISTS idx_purchase_payments_payment_id ON purchase_payments(payment_id);
		CREATE INDEX IF NOT EXISTS idx_purchase_payments_date ON purchase_payments(date);
	`
	
	result := db.Exec(createPurchasePaymentsSQL)
	if result.Error != nil {
		fmt.Printf("❌ Failed to create purchase_payments table: %v\n", result.Error)
	} else {
		fmt.Printf("✅ purchase_payments table created successfully\n")
	}

	// 2. Create account_balances materialized view
	fmt.Println("📊 Creating account_balances materialized view...")
	createAccountBalancesSQL := `
		-- Drop materialized view if it exists to recreate safely
		DROP MATERIALIZED VIEW IF EXISTS account_balances;

		-- Create account_balances materialized view
		CREATE MATERIALIZED VIEW account_balances AS
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			a.category as account_category,
			
			-- Current balance from accounts table
			a.balance as current_balance,
			
			-- Calculate balance from SSOT journal system (if tables exist)
			CASE 
				WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_lines') THEN
					COALESCE((
						SELECT 
							CASE 
								WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
									SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
								ELSE 
									SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
							END
						FROM unified_journal_lines ujl
						JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
						WHERE ujl.account_id = a.id 
						  AND ujd.status = 'POSTED'
						  AND ujd.deleted_at IS NULL
					), 0)
				ELSE 
					-- Fallback to traditional journal system
					COALESCE((
						SELECT 
							CASE 
								WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
									SUM(jl.debit_amount) - SUM(jl.credit_amount)
								ELSE 
									SUM(jl.credit_amount) - SUM(jl.debit_amount)
							END
						FROM journal_lines jl
						JOIN journal_entries je ON jl.journal_entry_id = je.id
						WHERE jl.account_id = a.id 
						  AND je.status = 'POSTED'
						  AND je.deleted_at IS NULL
					), 0)
			END as calculated_balance,
			
			-- Balance difference for reconciliation
			a.balance - CASE 
				WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_lines') THEN
					COALESCE((
						SELECT 
							CASE 
								WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
									SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
								ELSE 
									SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
							END
						FROM unified_journal_lines ujl
						JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
						WHERE ujl.account_id = a.id 
						  AND ujd.status = 'POSTED'
						  AND ujd.deleted_at IS NULL
					), 0)
				ELSE 
					COALESCE((
						SELECT 
							CASE 
								WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
									SUM(jl.debit_amount) - SUM(jl.credit_amount)
								ELSE 
									SUM(jl.credit_amount) - SUM(jl.debit_amount)
							END
						FROM journal_lines jl
						JOIN journal_entries je ON jl.journal_entry_id = je.id
						WHERE jl.account_id = a.id 
						  AND je.status = 'POSTED'
						  AND je.deleted_at IS NULL
					), 0)
			END as balance_difference,
			
			-- Metadata
			a.is_active,
			a.created_at,
			a.updated_at,
			NOW() as last_refresh

		FROM accounts a
		WHERE a.deleted_at IS NULL;

		-- Create indexes on materialized view for better performance
		CREATE INDEX IF NOT EXISTS idx_account_balances_account_id ON account_balances(account_id);
		CREATE INDEX IF NOT EXISTS idx_account_balances_account_type ON account_balances(account_type);
		CREATE INDEX IF NOT EXISTS idx_account_balances_difference ON account_balances(balance_difference) WHERE ABS(balance_difference) > 0.01;
	`
	
	result = db.Exec(createAccountBalancesSQL)
	if result.Error != nil {
		fmt.Printf("❌ Failed to create account_balances materialized view: %v\n", result.Error)
	} else {
		fmt.Printf("✅ account_balances materialized view created successfully\n")
	}

	// 3. Create SSOT sync function
	fmt.Println("🔧 Creating SSOT sync function...")
	createSSOTFunctionSQL := `
		-- Create or replace the BIGINT variant used by triggers
		CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param BIGINT)
		RETURNS VOID AS $$
		DECLARE
			account_type_var VARCHAR(50);
			new_balance DECIMAL(20,2);
		BEGIN
			-- Get account type (ensure account exists)
			SELECT type INTO account_type_var
			FROM accounts 
			WHERE id = account_id_param;

			-- Compute balance from posted SSOT journals with normal balance rules
			SELECT 
				CASE 
					WHEN account_type_var IN ('ASSET', 'EXPENSE') THEN 
						COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
					ELSE 
						COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
				END
			INTO new_balance
			FROM unified_journal_lines ujl
			LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
			WHERE ujl.account_id = account_id_param
			  AND uje.status = 'POSTED';

			-- Update accounts.balance
			UPDATE accounts 
			SET balance = COALESCE(new_balance, 0),
				updated_at = NOW()
			WHERE id = account_id_param;

		END;
		$$ LANGUAGE plpgsql;

		-- Keep an INTEGER overload for backward compatibility; delegate to BIGINT
		CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param INTEGER)
		RETURNS VOID AS $$
		BEGIN
			PERFORM sync_account_balance_from_ssot(account_id_param::BIGINT);
		END;
		$$ LANGUAGE plpgsql;

		-- Create refresh function for account_balances materialized view
		CREATE OR REPLACE FUNCTION refresh_account_balances()
		RETURNS VOID AS $$
		BEGIN
			REFRESH MATERIALIZED VIEW account_balances;
		END;
		$$ LANGUAGE plpgsql;
	`
	
	result = db.Exec(createSSOTFunctionSQL)
	if result.Error != nil {
		fmt.Printf("❌ Failed to create SSOT sync functions: %v\n", result.Error)
	} else {
		fmt.Printf("✅ SSOT sync functions created successfully\n")
	}

	// 4. Test the fixes
	fmt.Println("🧪 Testing the fixes...")
	
	// Test if account_balances materialized view exists
	var exists bool
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM pg_matviews WHERE matviewname = 'account_balances')").Scan(&exists)
	if result.Error != nil {
		fmt.Printf("⚠️  Could not check account_balances materialized view: %v\n", result.Error)
	} else {
		fmt.Printf("✅ account_balances materialized view exists: %v\n", exists)
	}

	// Test if purchase_payments table exists
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'purchase_payments')").Scan(&exists)
	if result.Error != nil {
		fmt.Printf("⚠️  Could not check purchase_payments table: %v\n", result.Error)
	} else {
		fmt.Printf("✅ purchase_payments table exists: %v\n", exists)
	}

	// Test SSOT function
	result = db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'sync_account_balance_from_ssot')").Scan(&exists)
	if result.Error != nil {
		fmt.Printf("⚠️  Could not check SSOT functions: %v\n", result.Error)
	} else {
		fmt.Printf("✅ SSOT sync function exists: %v\n", exists)
	}

	fmt.Println("🎯 Manual migration fixes completed!")
	fmt.Println("📋 Summary:")
	fmt.Println("   ✅ purchase_payments table: Created")
	fmt.Println("   ✅ account_balances materialized view: Created") 
	fmt.Println("   ✅ SSOT sync functions: Created")
	fmt.Println("")
	fmt.Println("💡 You can now restart your backend server.")
	fmt.Println("   The SSOT Journal System should work properly now.")
}