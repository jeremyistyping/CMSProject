package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("")
	fmt.Println("============================================")
	fmt.Println("Dropping Remaining Problematic Triggers")
	fmt.Println("============================================")
	fmt.Println("")

	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Drop remaining triggers
	remainingTriggers := []struct {
		name  string
		table string
	}{
		{"trg_sync_coa_to_cashbank", "accounts"},
		{"trigger_validate_account_balance_consistency", "accounts"},
		{"trg_sync_cashbank_coa", "cash_bank_transactions"},
		{"trg_sync_cashbank_balance_update", "cash_banks"},
	}

	fmt.Println("Dropping remaining triggers...")
	for _, trigger := range remainingTriggers {
		sql := fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", trigger.name, trigger.table)
		if err := db.Exec(sql).Error; err != nil {
			fmt.Printf("⚠️  Warning: Failed to drop %s on %s: %v\n", trigger.name, trigger.table, err)
		} else {
			fmt.Printf("✅ Dropped trigger: %s on %s\n", trigger.name, trigger.table)
		}
	}

	// Also drop any triggers with similar names
	additionalTriggers := []string{
		"DROP TRIGGER IF EXISTS trg_sync_cashbank_coa ON cash_bank_transactions",
		"DROP TRIGGER IF EXISTS trg_sync_cashbank_coa ON cash_bank_transactions",
		"DROP TRIGGER IF EXISTS trg_sync_cashbank_coa ON cash_bank_transactions",
	}

	for _, sql := range additionalTriggers {
		if err := db.Exec(sql).Error; err != nil {
			fmt.Printf("⚠️  Warning: Failed to execute %s: %v\n", sql, err)
		} else {
			fmt.Printf("✅ Executed: %s\n", sql)
		}
	}

	// Final check
	fmt.Println("\nFinal check - remaining triggers...")
	var remaining []struct {
		TriggerName string `gorm:"column:trigger_name"`
		TableName   string `gorm:"column:event_object_table"`
	}

	if err := db.Raw(`
		SELECT 
			trigger_name, 
			event_object_table
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
	fmt.Println("✅ All triggers cleanup completed!")
	fmt.Println("")
}
