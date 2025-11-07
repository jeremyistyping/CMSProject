package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	// Load configuration
	_ = config.LoadConfig()
	fmt.Printf("✅ Configuration loaded successfully\n")

	// Connect to database
	db := database.ConnectDB()
	fmt.Printf("✅ Database connected successfully\n")

	// Check if database triggers exist
	fmt.Printf("\n=== Checking Database Triggers ===\n")
	var triggerCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.triggers 
		WHERE trigger_name LIKE '%sync_cashbank%'
	`).Scan(&triggerCount)
	
	if triggerCount > 0 {
		fmt.Printf("✅ Found %d CashBank sync trigger(s)\n", triggerCount)
		
		// Get trigger details
		var triggers []struct {
			TriggerName string `db:"trigger_name"`
			TableName   string `db:"event_object_table"`
			Timing      string `db:"action_timing"`
			Event       string `db:"event_manipulation"`
		}
		
		db.Raw(`
			SELECT trigger_name, event_object_table, action_timing, event_manipulation 
			FROM information_schema.triggers 
			WHERE trigger_name LIKE '%sync_cashbank%'
		`).Scan(&triggers)
		
		for _, trigger := range triggers {
			fmt.Printf("   - Trigger: %s on %s (%s %s)\n", 
				trigger.TriggerName, trigger.TableName, trigger.Timing, trigger.Event)
		}
	} else {
		fmt.Printf("❌ No CashBank sync triggers found - triggers not installed\n")
	}

	// Check if audit_logs table exists
	fmt.Printf("\n=== Checking Audit Logs Table ===\n")
	var auditTableExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'audit_logs'
		)
	`).Scan(&auditTableExists)
	
	if auditTableExists {
		fmt.Printf("✅ audit_logs table exists\n")
	} else {
		fmt.Printf("❌ audit_logs table does not exist\n")
	}

	// Check CashBank-COA linking status
	fmt.Printf("\n=== Checking CashBank-COA Linking ===\n")
	var totalCashBanks, linkedCashBanks int64
	
	db.Model(&struct{}{}).Table("cash_banks").
		Where("deleted_at IS NULL AND is_active = true").
		Count(&totalCashBanks)
	
	db.Model(&struct{}{}).Table("cash_banks").
		Where("deleted_at IS NULL AND is_active = true AND account_id > 0").
		Count(&linkedCashBanks)
	
	fmt.Printf("Total CashBanks: %d\n", totalCashBanks)
	fmt.Printf("Linked CashBanks: %d\n", linkedCashBanks)
	fmt.Printf("Unlinked CashBanks: %d\n", totalCashBanks-linkedCashBanks)
	
	if linkedCashBanks == totalCashBanks && totalCashBanks > 0 {
		fmt.Printf("✅ All CashBanks are linked to COA accounts\n")
	} else if linkedCashBanks > 0 {
		fmt.Printf("⚠️  Some CashBanks are not linked to COA accounts\n")
	} else {
		fmt.Printf("❌ No CashBanks are linked to COA accounts\n")
	}

	// Check for any sync discrepancies
	fmt.Printf("\n=== Checking Sync Discrepancies ===\n")
	var discrepancyCount int64
	db.Raw(`
		SELECT COUNT(*)
		FROM cash_banks cb
		LEFT JOIN accounts a ON cb.account_id = a.id AND a.deleted_at IS NULL
		LEFT JOIN (
			SELECT 
				cash_bank_id,
				SUM(amount) as transaction_sum
			FROM cash_bank_transactions 
			WHERE deleted_at IS NULL 
			GROUP BY cash_bank_id
		) tx_sum ON cb.id = tx_sum.cash_bank_id
		WHERE cb.deleted_at IS NULL 
		  AND cb.is_active = true
		  AND (
		      cb.account_id = 0 OR cb.account_id IS NULL OR
		      cb.balance != COALESCE(a.balance, 0) OR
		      COALESCE(tx_sum.transaction_sum, 0) != cb.balance
		  )
	`).Scan(&discrepancyCount)
	
	if discrepancyCount == 0 {
		fmt.Printf("✅ No sync discrepancies found\n")
	} else {
		fmt.Printf("⚠️  Found %d sync discrepancies\n", discrepancyCount)
	}

	// Final assessment
	fmt.Printf("\n=== PHASE 1 IMPLEMENTATION STATUS ===\n")
	
	score := 0
	maxScore := 4
	
	if triggerCount > 0 {
		score++
		fmt.Printf("✅ Database triggers: IMPLEMENTED\n")
	} else {
		fmt.Printf("❌ Database triggers: NOT IMPLEMENTED\n")
	}
	
	if auditTableExists {
		score++
		fmt.Printf("✅ Audit logging: IMPLEMENTED\n")
	} else {
		fmt.Printf("❌ Audit logging: NOT IMPLEMENTED\n")
	}
	
	if linkedCashBanks > 0 {
		score++
		fmt.Printf("✅ CashBank-COA linking: PARTIALLY IMPLEMENTED (%d/%d linked)\n", linkedCashBanks, totalCashBanks)
	} else {
		fmt.Printf("❌ CashBank-COA linking: NOT IMPLEMENTED\n")
	}
	
	if discrepancyCount == 0 {
		score++
		fmt.Printf("✅ Sync validation: PASSED\n")
	} else {
		fmt.Printf("⚠️  Sync validation: ISSUES FOUND\n")
	}
	
	fmt.Printf("\nOVERALL SCORE: %d/%d\n", score, maxScore)
	
	if score == maxScore {
		fmt.Printf("🎉 PHASE 1 FULLY IMPLEMENTED AND WORKING!\n")
	} else if score >= maxScore/2 {
		fmt.Printf("⚠️  PHASE 1 PARTIALLY IMPLEMENTED - needs attention\n")
	} else {
		fmt.Printf("❌ PHASE 1 NOT PROPERLY IMPLEMENTED\n")
	}
}
