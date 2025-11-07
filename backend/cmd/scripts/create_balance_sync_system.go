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

	fmt.Printf("🔗 Connecting to database...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get underlying sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	fmt.Printf("🔧 Creating Balance Synchronization System...\n\n")

	// 1. Create stored procedure for balance recalculation
	createSyncProcedure := `
	CREATE OR REPLACE FUNCTION sync_account_balances() 
	RETURNS TABLE(account_id INT, old_balance NUMERIC, new_balance NUMERIC, difference NUMERIC) 
	LANGUAGE plpgsql AS $$
	DECLARE
		account_rec RECORD;
		calc_balance NUMERIC;
		old_bal NUMERIC;
	BEGIN
		-- Loop through all active accounts
		FOR account_rec IN 
			SELECT id, balance FROM accounts WHERE is_active = true
		LOOP
			old_bal := account_rec.balance;
			
			-- Calculate correct balance from unified_journal_lines
			SELECT 
				COALESCE(SUM(CAST(ujl.debit_amount AS DECIMAL)), 0) - 
				COALESCE(SUM(CAST(ujl.credit_amount AS DECIMAL)), 0)
			INTO calc_balance
			FROM unified_journal_lines ujl
			WHERE ujl.account_id = account_rec.id;
			
			-- If calculated balance is null (no journal entries), set to 0
			IF calc_balance IS NULL THEN
				calc_balance := 0;
			END IF;
			
			-- Update balance if different
			IF old_bal != calc_balance THEN
				UPDATE accounts 
				SET balance = calc_balance, updated_at = NOW() 
				WHERE id = account_rec.id;
				
				-- Return the change
				account_id := account_rec.id;
				old_balance := old_bal;
				new_balance := calc_balance;
				difference := calc_balance - old_bal;
				RETURN NEXT;
			END IF;
		END LOOP;
		
		RETURN;
	END;
	$$;
	`

	fmt.Printf("📝 Creating sync_account_balances() stored procedure...\n")
	_, err = sqlDB.Exec(createSyncProcedure)
	if err != nil {
		log.Printf("Warning: Failed to create stored procedure: %v\n", err)
	} else {
		fmt.Printf("✅ Stored procedure created successfully\n")
	}

	// 2. Create trigger function for automatic balance updates
	createTriggerFunction := `
	CREATE OR REPLACE FUNCTION update_account_balance_trigger()
	RETURNS TRIGGER AS $$
	DECLARE
		affected_account_id INT;
		new_balance NUMERIC;
	BEGIN
		-- Determine which account was affected
		IF TG_OP = 'DELETE' THEN
			affected_account_id := OLD.account_id;
		ELSE
			affected_account_id := NEW.account_id;
		END IF;
		
		-- Recalculate balance for the affected account
		SELECT 
			COALESCE(SUM(CAST(debit_amount AS DECIMAL)), 0) - 
			COALESCE(SUM(CAST(credit_amount AS DECIMAL)), 0)
		INTO new_balance
		FROM unified_journal_lines
		WHERE account_id = affected_account_id;
		
		-- Update the account balance
		UPDATE accounts 
		SET balance = COALESCE(new_balance, 0), updated_at = NOW()
		WHERE id = affected_account_id;
		
		-- Return the appropriate row
		IF TG_OP = 'DELETE' THEN
			RETURN OLD;
		ELSE
			RETURN NEW;
		END IF;
	END;
	$$ LANGUAGE plpgsql;
	`

	fmt.Printf("📝 Creating automatic balance update trigger function...\n")
	_, err = sqlDB.Exec(createTriggerFunction)
	if err != nil {
		log.Printf("Warning: Failed to create trigger function: %v\n", err)
	} else {
		fmt.Printf("✅ Trigger function created successfully\n")
	}

	// 3. Create the actual trigger
	createTrigger := `
	DROP TRIGGER IF EXISTS balance_sync_trigger ON unified_journal_lines;
	CREATE TRIGGER balance_sync_trigger
		AFTER INSERT OR UPDATE OR DELETE ON unified_journal_lines
		FOR EACH ROW
		EXECUTE FUNCTION update_account_balance_trigger();
	`

	fmt.Printf("📝 Creating automatic balance sync trigger...\n")
	_, err = sqlDB.Exec(createTrigger)
	if err != nil {
		log.Printf("Warning: Failed to create trigger: %v\n", err)
	} else {
		fmt.Printf("✅ Automatic balance sync trigger created successfully\n")
	}

	// 4. Create a view for monitoring balance consistency
	createMonitoringView := `
	CREATE OR REPLACE VIEW account_balance_monitoring AS
	WITH calculated_balances AS (
		SELECT 
			account_id,
			COALESCE(SUM(CAST(debit_amount AS DECIMAL)), 0) - 
			COALESCE(SUM(CAST(credit_amount AS DECIMAL)), 0) as calculated_balance
		FROM unified_journal_lines
		GROUP BY account_id
	)
	SELECT 
		a.id,
		a.code,
		a.name,
		a.type,
		CAST(a.balance AS DECIMAL) as stored_balance,
		COALESCE(cb.calculated_balance, 0) as calculated_balance,
		CAST(a.balance AS DECIMAL) - COALESCE(cb.calculated_balance, 0) as difference,
		CASE 
			WHEN CAST(a.balance AS DECIMAL) != COALESCE(cb.calculated_balance, 0) THEN 'MISMATCH'
			ELSE 'OK'
		END as status,
		a.updated_at as last_balance_update
	FROM accounts a
	LEFT JOIN calculated_balances cb ON a.id = cb.account_id
	WHERE a.is_active = true
	ORDER BY ABS(CAST(a.balance AS DECIMAL) - COALESCE(cb.calculated_balance, 0)) DESC;
	`

	fmt.Printf("📝 Creating monitoring view...\n")
	_, err = sqlDB.Exec(createMonitoringView)
	if err != nil {
		log.Printf("Warning: Failed to create monitoring view: %v\n", err)
	} else {
		fmt.Printf("✅ Balance monitoring view created successfully\n")
	}

	// 5. Run the initial sync to fix current issues
	fmt.Printf("\n🔄 Running initial balance synchronization...\n")
	rows, err := sqlDB.Query("SELECT * FROM sync_account_balances()")
	if err != nil {
		log.Printf("Warning: Failed to run initial sync: %v\n", err)
	} else {
		defer rows.Close()
		
		syncCount := 0
		fmt.Printf("Fixed accounts:\n")
		for rows.Next() {
			var accountID int
			var oldBalance, newBalance, difference float64
			err := rows.Scan(&accountID, &oldBalance, &newBalance, &difference)
			if err != nil {
				log.Printf("Error scanning sync result: %v", err)
				continue
			}
			
			fmt.Printf("  - Account ID %d: Rp %.0f → Rp %.0f (diff: Rp %.0f)\n", 
				accountID, oldBalance, newBalance, difference)
			syncCount++
		}
		
		if syncCount == 0 {
			fmt.Printf("  No accounts needed fixing.\n")
		} else {
			fmt.Printf("  Total accounts fixed: %d\n", syncCount)
		}
	}

	fmt.Printf("\n🎉 Balance Synchronization System Setup Complete!\n\n")
	
	fmt.Printf("📋 FEATURES CREATED:\n")
	fmt.Printf("  1. ✅ sync_account_balances() - Manual sync function\n")
	fmt.Printf("  2. ✅ Automatic trigger on unified_journal_lines changes\n")
	fmt.Printf("  3. ✅ account_balance_monitoring view for health checks\n\n")
	
	fmt.Printf("💡 HOW TO USE:\n")
	fmt.Printf("  • Manual sync:     SELECT * FROM sync_account_balances();\n")
	fmt.Printf("  • Health check:    SELECT * FROM account_balance_monitoring WHERE status='MISMATCH';\n")
	fmt.Printf("  • Monitor all:     SELECT * FROM account_balance_monitoring;\n\n")
	
	fmt.Printf("🚀 PREVENTION ACHIEVED:\n")
	fmt.Printf("  • Future journal entries will auto-update account balances\n")
	fmt.Printf("  • Manual sync available for batch corrections\n")
	fmt.Printf("  • Real-time monitoring for early detection\n")
}