package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("")
	fmt.Println("============================================")
	fmt.Println("EMERGENCY: Dropping All Problematic Triggers")
	fmt.Println("============================================")
	fmt.Println("")

	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// List of triggers to drop
	triggers := []struct {
		name string
		table string
	}{
		{"trigger_sync_cashbank_coa", "cash_banks"},
		{"trigger_recalc_cashbank_balance_insert", "cash_bank_transactions"},
		{"trigger_recalc_cashbank_balance_update", "cash_bank_transactions"},
		{"trigger_recalc_cashbank_balance_delete", "cash_bank_transactions"},
		{"trigger_validate_account_balance", "accounts"},
		{"trigger_update_parent_balances", "accounts"},
		{"trigger_update_parent_account_balances", "accounts"},
		{"trigger_sync_account_balance", "accounts"},
		{"trigger_update_account_balance", "accounts"},
	}

	fmt.Println("Dropping triggers...")
	for _, trigger := range triggers {
		sql := fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", trigger.name, trigger.table)
		if err := db.Exec(sql).Error; err != nil {
			fmt.Printf("⚠️  Warning: Failed to drop %s on %s: %v\n", trigger.name, trigger.table, err)
		} else {
			fmt.Printf("✅ Dropped trigger: %s on %s\n", trigger.name, trigger.table)
		}
	}

	// List remaining triggers
	fmt.Println("\nChecking remaining triggers...")
	var remaining []struct {
		TriggerName string `gorm:"column:trigger_name"`
		TableName   string `gorm:"column:event_object_table"`
		Action      string `gorm:"column:action_statement"`
	}

	if err := db.Raw(`
		SELECT 
			trigger_name, 
			event_object_table, 
			action_statement
		FROM information_schema.triggers 
		WHERE event_object_schema = 'public'
		AND event_object_table IN ('accounts', 'cash_banks', 'cash_bank_transactions')
		ORDER BY event_object_table, trigger_name
	`).Scan(&remaining).Error; err != nil {
		fmt.Printf("⚠️  Warning: Failed to check remaining triggers: %v\n", err)
	} else {
		if len(remaining) == 0 {
			fmt.Println("✅ No problematic triggers remaining!")
		} else {
			fmt.Printf("⚠️  %d triggers still exist:\n", len(remaining))
			for _, t := range remaining {
				fmt.Printf("   - %s on %s\n", t.TriggerName, t.TableName)
			}
		}
	}

	fmt.Println("")
	fmt.Println("✅ Emergency trigger cleanup completed!")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  1. Restart backend (it will run migration automatically)")
	fmt.Println("  2. Test invoice creation again")
	fmt.Println("")
}
