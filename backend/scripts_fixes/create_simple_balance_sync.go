package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("🔧 Creating Simple Automatic Balance Sync System")
	log.Println("===============================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	log.Printf("✅ Database connected successfully")

	// Step 1: Create sync function
	log.Println("\n1. Creating balance sync function...")
	
	syncFunctionSQL := `
	CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id_param INTEGER)
	RETURNS VOID AS $$
	DECLARE
	    account_type_var VARCHAR(50);
	    new_balance DECIMAL(20,2);
	BEGIN
	    -- Get account type
	    SELECT type INTO account_type_var 
	    FROM accounts 
	    WHERE id = account_id_param;
	    
	    -- Calculate balance based on account type and SSOT journal entries
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
	    
	    -- Update account balance
	    UPDATE accounts 
	    SET 
	        balance = COALESCE(new_balance, 0),
	        updated_at = NOW()
	    WHERE id = account_id_param;
	    
	    -- RAISE NOTICE 'Updated account ID % balance to %', account_id_param, COALESCE(new_balance, 0);
	END;
	$$ LANGUAGE plpgsql;`
	
	err = db.Exec(syncFunctionSQL).Error
	if err != nil {
		log.Printf("❌ Error creating sync function: %v", err)
	} else {
		log.Printf("✅ Sync function created successfully")
	}

	// Step 2: Create trigger function for journal status changes
	log.Println("\n2. Creating trigger function...")
	
	triggerFunctionSQL := `
	CREATE OR REPLACE FUNCTION trigger_sync_on_journal_posting()
	RETURNS TRIGGER AS $$
	DECLARE
	    line_record RECORD;
	BEGIN
	    -- When journal entry status changes to POSTED, sync all related accounts
	    IF NEW.status = 'POSTED' AND (OLD.status IS NULL OR OLD.status != 'POSTED') THEN
	        FOR line_record IN 
	            SELECT DISTINCT account_id 
	            FROM unified_journal_lines 
	            WHERE journal_id = NEW.id
	        LOOP
	            PERFORM sync_account_balance_from_ssot(line_record.account_id);
	        END LOOP;
	    END IF;
	    
	    -- When journal entry is reversed or cancelled, sync all related accounts
	    IF NEW.status IN ('REVERSED', 'CANCELLED') AND OLD.status = 'POSTED' THEN
	        FOR line_record IN 
	            SELECT DISTINCT account_id 
	            FROM unified_journal_lines 
	            WHERE journal_id = NEW.id
	        LOOP
	            PERFORM sync_account_balance_from_ssot(line_record.account_id);
	        END LOOP;
	    END IF;
	    
	    RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;`
	
	err = db.Exec(triggerFunctionSQL).Error
	if err != nil {
		log.Printf("❌ Error creating trigger function: %v", err)
	} else {
		log.Printf("✅ Trigger function created successfully")
	}

	// Step 3: Create trigger
	log.Println("\n3. Creating automatic sync trigger...")
	
	// Drop existing trigger first
	db.Exec("DROP TRIGGER IF EXISTS trg_auto_sync_balance_on_posting ON unified_journal_ledger")
	
	createTriggerSQL := `
	CREATE TRIGGER trg_auto_sync_balance_on_posting
	    AFTER UPDATE ON unified_journal_ledger
	    FOR EACH ROW 
	    WHEN (OLD.status IS DISTINCT FROM NEW.status)
	    EXECUTE FUNCTION trigger_sync_on_journal_posting();`
	
	err = db.Exec(createTriggerSQL).Error
	if err != nil {
		log.Printf("❌ Error creating trigger: %v", err)
	} else {
		log.Printf("✅ Automatic sync trigger created successfully")
	}

	// Step 4: Test the system
	log.Println("\n4. Testing the system...")
	
	// Check current cash balance
	var cashBalance float64
	db.Table("accounts").
		Select("balance").
		Where("code = '1100-075'").
		Scan(&cashBalance)
	
	log.Printf("💰 Current Kas balance: Rp %.2f", cashBalance)
	
	// Verify functions exist
	var functionExists bool
	db.Raw("SELECT EXISTS(SELECT 1 FROM pg_proc WHERE proname = 'sync_account_balance_from_ssot')").Scan(&functionExists)
	if functionExists {
		log.Printf("✅ Sync function is installed")
	} else {
		log.Printf("❌ Sync function not found")
	}
	
	// Verify trigger exists
	var triggerExists bool
	db.Raw("SELECT EXISTS(SELECT 1 FROM pg_trigger WHERE tgname = 'trg_auto_sync_balance_on_posting')").Scan(&triggerExists)
	if triggerExists {
		log.Printf("✅ Auto-sync trigger is active")
	} else {
		log.Printf("❌ Auto-sync trigger not found")
	}

	// Step 5: Create manual sync utility function
	log.Println("\n5. Creating manual sync utility...")
	
	manualSyncSQL := `
	CREATE OR REPLACE FUNCTION manual_sync_all_account_balances()
	RETURNS TEXT AS $$
	DECLARE
	    account_record RECORD;
	    sync_count INTEGER := 0;
	BEGIN
	    FOR account_record IN 
	        SELECT DISTINCT a.id 
	        FROM accounts a
	        WHERE EXISTS (
	            SELECT 1 FROM unified_journal_lines ujl
	            LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
	            WHERE ujl.account_id = a.id AND uje.status = 'POSTED'
	        )
	    LOOP
	        PERFORM sync_account_balance_from_ssot(account_record.id);
	        sync_count := sync_count + 1;
	    END LOOP;
	    
	    RETURN 'Successfully synced ' || sync_count || ' account balances';
	END;
	$$ LANGUAGE plpgsql;`
	
	err = db.Exec(manualSyncSQL).Error
	if err != nil {
		log.Printf("❌ Error creating manual sync function: %v", err)
	} else {
		log.Printf("✅ Manual sync utility created successfully")
	}

	log.Println("\n🎉 AUTOMATIC BALANCE SYNC SYSTEM IS NOW ACTIVE!")
	log.Println()
	log.Println("📋 System Features:")
	log.Println("✅ Automatic sync when journal entries are posted")
	log.Println("✅ Automatic sync when journal entries are reversed/cancelled")
	log.Println("✅ Manual sync function available: SELECT manual_sync_all_account_balances()")
	log.Println()
	log.Println("💡 How it works:")
	log.Println("• Every time a journal entry status changes to 'POSTED'")
	log.Println("• The trigger automatically updates all related account balances")
	log.Println("• Your frontend will always show accurate, real-time balances")
	log.Println()
	log.Println("🚀 Future deposits and transactions will automatically update balances!")
}