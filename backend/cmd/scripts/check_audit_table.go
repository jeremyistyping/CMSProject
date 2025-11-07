package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔍 Checking audit_logs table structure...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Check if audit_logs table exists
	var tableExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'audit_logs'
		)
	`).Scan(&tableExists)
	
	fmt.Printf("audit_logs table exists: %v\n", tableExists)
	
	if tableExists {
		// Get column information
		var columns []struct {
			ColumnName string `db:"column_name"`
			DataType   string `db:"data_type"`
			IsNullable string `db:"is_nullable"`
		}
		
		db.Raw(`
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_name = 'audit_logs'
			ORDER BY ordinal_position
		`).Scan(&columns)
		
		fmt.Printf("\nTable structure (%d columns):\n", len(columns))
		for _, col := range columns {
			fmt.Printf("  - %s: %s (nullable: %s)\n", col.ColumnName, col.DataType, col.IsNullable)
		}
	}
	
	// Try alternative approach - create minimal reverse sync without notes
	fmt.Println("\n🔧 Creating simplified reverse sync trigger...")
	
	simpleTrigger := `
-- Simplified reverse sync trigger without notes column
CREATE OR REPLACE FUNCTION sync_coa_to_cashbank_simple()
RETURNS TRIGGER AS $$
BEGIN
    -- Only process if balance actually changed
    IF OLD.balance IS DISTINCT FROM NEW.balance THEN
        
        -- Update all linked CashBanks
        UPDATE cash_banks 
        SET balance = NEW.balance, updated_at = NOW()
        WHERE account_id = NEW.id 
        AND deleted_at IS NULL 
        AND is_active = true;
        
        -- Simple audit log without notes
        INSERT INTO audit_logs (
            table_name, 
            action, 
            record_id, 
            old_values, 
            new_values,
            created_at
        ) VALUES (
            'coa_to_cashbank_sync',
            'COA_BALANCE_CHANGED',
            NEW.id,
            json_build_object('old_balance', OLD.balance),
            json_build_object('new_balance', NEW.balance, 'account_code', NEW.code),
            NOW()
        );
        
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Replace trigger
DROP TRIGGER IF EXISTS trg_sync_coa_to_cashbank ON accounts;
CREATE TRIGGER trg_sync_coa_to_cashbank
    AFTER UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION sync_coa_to_cashbank_simple();
`
	
	err := db.Exec(simpleTrigger).Error
	if err != nil {
		fmt.Printf("❌ Failed to create simplified trigger: %v\n", err)
	} else {
		fmt.Println("✅ Simplified reverse sync trigger created successfully!")
	}
	
	// Test the simplified trigger
	fmt.Println("\n🧪 Testing simplified reverse sync...")
	
	// Find a CashBank to test with
	var testData struct {
		CashBankID      uint    `db:"cashbank_id"`
		CashBankName    string  `db:"cashbank_name"`
		CashBankBalance float64 `db:"cashbank_balance"`
		AccountID       uint    `db:"account_id"`
		AccountCode     string  `db:"account_code"`
		COABalance      float64 `db:"coa_balance"`
	}
	
	err = db.Raw(`
		SELECT 
			cb.id as cashbank_id,
			cb.name as cashbank_name, 
			cb.balance as cashbank_balance,
			a.id as account_id,
			a.code as account_code,
			a.balance as coa_balance
		FROM cash_banks cb 
		JOIN accounts a ON cb.account_id = a.id 
		WHERE cb.deleted_at IS NULL AND cb.is_active = true
		LIMIT 1
	`).Scan(&testData).Error
	
	if err != nil {
		fmt.Printf("❌ Error finding test data: %v\n", err)
		return
	}
	
	if testData.AccountID > 0 {
		fmt.Printf("📊 Testing with:\n")
		fmt.Printf("   CashBank: %s (balance: %.2f)\n", testData.CashBankName, testData.CashBankBalance)
		fmt.Printf("   COA: %s (balance: %.2f)\n", testData.AccountCode, testData.COABalance)
		
		// Test by changing COA balance
		newBalance := testData.COABalance + 100000
		fmt.Printf("\n🔧 Updating COA %s balance from %.2f to %.2f\n", testData.AccountCode, testData.COABalance, newBalance)
		
		err = db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", newBalance, testData.AccountID).Error
		if err != nil {
			fmt.Printf("❌ Failed to update COA: %v\n", err)
		} else {
			// Check if CashBank updated
			var updatedCashBankBalance float64
			db.Raw("SELECT balance FROM cash_banks WHERE id = ?", testData.CashBankID).Scan(&updatedCashBankBalance)
			
			if updatedCashBankBalance == newBalance {
				fmt.Printf("🎉 SUCCESS! CashBank balance synced to %.2f\n", updatedCashBankBalance)
			} else {
				fmt.Printf("❌ FAILED! CashBank balance is %.2f, expected %.2f\n", updatedCashBankBalance, newBalance)
			}
			
			// Restore original
			db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", testData.COABalance, testData.AccountID)
			fmt.Println("🔄 Original balance restored")
		}
	}
	
	fmt.Println("\n✅ Reverse sync setup completed!")
}
