package database

import (
	"log"
	"gorm.io/gorm"
)

// FixConcurrentRefreshError removes the problematic trigger that causes SQLSTATE 55000
// This is a critical fix that runs on every startup to ensure the trigger is removed
func FixConcurrentRefreshError(db *gorm.DB) {
	log.Println("🔧 [CRITICAL FIX] Checking for concurrent refresh trigger...")
	
	// Check if trigger exists
	var triggerExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_trigger 
			WHERE tgname = 'trg_refresh_account_balances'
		)
	`).Scan(&triggerExists).Error
	
	if err != nil {
		log.Printf("⚠️  Failed to check trigger existence: %v", err)
		return
	}
	
	if !triggerExists {
		log.Println("✅ Trigger already removed - no action needed")
		return
	}
	
	log.Println("⚠️  Found problematic trigger - removing now...")
	
	// Drop the trigger
	if err := db.Exec("DROP TRIGGER IF EXISTS trg_refresh_account_balances ON unified_journal_lines").Error; err != nil {
		log.Printf("❌ Failed to drop trigger: %v", err)
		return
	}
	
	log.Println("✅ Trigger removed successfully")
	
	// Create helper functions if they don't exist
	log.Println("🔧 Creating manual refresh helper functions...")
	
	// Create manual refresh function
	createManualRefreshSQL := `
		CREATE OR REPLACE FUNCTION manual_refresh_account_balances()
		RETURNS TABLE(
			success BOOLEAN,
			message TEXT,
			refreshed_at TIMESTAMPTZ
		) AS $$
		DECLARE
			start_time TIMESTAMPTZ;
			end_time TIMESTAMPTZ;
			duration INTERVAL;
		BEGIN
			start_time := clock_timestamp();
			
			-- Refresh the materialized view
			REFRESH MATERIALIZED VIEW account_balances;
			
			end_time := clock_timestamp();
			duration := end_time - start_time;
			
			RETURN QUERY SELECT 
				TRUE as success,
				format('Account balances refreshed in %s', duration) as message,
				end_time as refreshed_at;
				
		EXCEPTION WHEN OTHERS THEN
			RETURN QUERY SELECT 
				FALSE as success,
				format('Refresh failed: %s', SQLERRM) as message,
				clock_timestamp() as refreshed_at;
		END;
		$$ LANGUAGE plpgsql;
	`
	
	if err := db.Exec(createManualRefreshSQL).Error; err != nil {
		log.Printf("⚠️  Failed to create manual_refresh_account_balances function: %v", err)
	} else {
		log.Println("✅ manual_refresh_account_balances() function created")
	}
	
	// Create freshness check function
	createFreshnessCheckSQL := `
		CREATE OR REPLACE FUNCTION check_account_balances_freshness()
		RETURNS TABLE(
			last_updated TIMESTAMPTZ,
			age_minutes INTEGER,
			needs_refresh BOOLEAN
		) AS $$
		DECLARE
			last_update TIMESTAMPTZ;
			age_mins INTEGER;
		BEGIN
			-- Get the last_updated timestamp from the materialized view
			SELECT MAX(ab.last_updated) INTO last_update
			FROM account_balances ab;
			
			-- Calculate age in minutes
			age_mins := EXTRACT(EPOCH FROM (NOW() - last_update)) / 60;
			
			RETURN QUERY SELECT 
				last_update,
				age_mins,
				age_mins > 60 as needs_refresh; -- Suggest refresh if older than 1 hour
		END;
		$$ LANGUAGE plpgsql;
	`
	
	if err := db.Exec(createFreshnessCheckSQL).Error; err != nil {
		log.Printf("⚠️  Failed to create check_account_balances_freshness function: %v", err)
	} else {
		log.Println("✅ check_account_balances_freshness() function created")
	}
	
	log.Println("✅ [CRITICAL FIX] Concurrent refresh error fix completed")
	log.Println("ℹ️  Balance sync is handled by setup_automatic_balance_sync.sql triggers")
	log.Println("ℹ️  For materialized view refresh, use: SELECT * FROM manual_refresh_account_balances();")
}
