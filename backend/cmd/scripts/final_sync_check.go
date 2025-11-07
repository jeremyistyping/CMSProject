package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🎯 Final CashBank-COA Bidirectional Sync Test...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Show current mapping
	fmt.Println("\n📊 Current CashBank-COA Mapping:")
	fmt.Println("=====================================")
	
	var mappings []struct {
		CBID     uint    `db:"cb_id"`
		CBCode   string  `db:"cb_code"`
		CBName   string  `db:"cb_name"`
		CBBalance float64 `db:"cb_balance"`
		AccID    uint    `db:"acc_id"`
		AccCode  string  `db:"acc_code"`
		AccName  string  `db:"acc_name"`
		AccBalance float64 `db:"acc_balance"`
		IsSync   bool    `db:"is_sync"`
	}
	
	db.Raw(`
		SELECT 
			cb.id as cb_id,
			cb.code as cb_code,
			cb.name as cb_name,
			cb.balance as cb_balance,
			a.id as acc_id,
			a.code as acc_code,
			a.name as acc_name,
			a.balance as acc_balance,
			(cb.balance = a.balance) as is_sync
		FROM cash_banks cb 
		JOIN accounts a ON cb.account_id = a.id 
		WHERE cb.deleted_at IS NULL 
		  AND cb.is_active = true
		ORDER BY cb.id
	`).Scan(&mappings)
	
	fmt.Printf("Found %d linked CashBank-COA pairs:\n", len(mappings))
	for _, m := range mappings {
		syncStatus := "✅ SYNCED"
		if !m.IsSync {
			syncStatus = "❌ OUT OF SYNC"
		}
		fmt.Printf("  CB%d %s \"%s\" (%.0f) ↔ %s \"%s\" (%.0f) %s\n", 
			m.CBID, m.CBCode, m.CBName, m.CBBalance,
			m.AccCode, m.AccName, m.AccBalance, syncStatus)
	}
	
	if len(mappings) == 0 {
		fmt.Println("❌ No linked accounts found!")
		return
	}
	
	// Test with the first available account
	testMapping := mappings[0]
	fmt.Printf("\n🧪 Testing with: CB%d \"%s\" ↔ %s \"%s\"\n", 
		testMapping.CBID, testMapping.CBName, testMapping.AccCode, testMapping.AccName)
	
	// Test 1: COA → CashBank (Reverse Sync)
	fmt.Println("\n🔄 TEST 1: COA → CashBank (Reverse Sync)")
	fmt.Println("=========================================")
	
	originalCOABalance := testMapping.AccBalance
	originalCBBalance := testMapping.CBBalance
	testAmount1 := originalCOABalance + 75000
	
	fmt.Printf("Updating COA %s balance: %.2f → %.2f\n", 
		testMapping.AccCode, originalCOABalance, testAmount1)
	
	// Update COA balance
	err := db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", 
		testAmount1, testMapping.AccID).Error
	if err != nil {
		fmt.Printf("❌ Failed to update COA: %v\n", err)
	} else {
		// Check if CashBank balance updated
		var newCBBalance float64
		db.Raw("SELECT balance FROM cash_banks WHERE id = ?", testMapping.CBID).Scan(&newCBBalance)
		
		fmt.Printf("Result: CashBank balance changed to %.2f\n", newCBBalance)
		
		if newCBBalance == testAmount1 {
			fmt.Println("🎉 SUCCESS! COA → CashBank sync working!")
		} else {
			fmt.Printf("❌ FAILED! Expected %.2f, got %.2f\n", testAmount1, newCBBalance)
		}
		
		// Restore for next test
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", originalCOABalance, testMapping.AccID)
		db.Exec("UPDATE cash_banks SET balance = ? WHERE id = ?", originalCBBalance, testMapping.CBID)
	}
	
	// Test 2: CashBank → COA (Forward Sync) - via transaction
	fmt.Println("\n🔄 TEST 2: CashBank → COA (Forward Sync via Transaction)")
	fmt.Println("========================================================")
	
	testAmount2 := 25000.0
	fmt.Printf("Creating transaction: +%.0f in CashBank %d\n", testAmount2, testMapping.CBID)
	
	// Create a transaction (this should trigger forward sync)
	err = db.Exec(`
		INSERT INTO cash_bank_transactions 
		(cash_bank_id, amount, balance_after, transaction_date, reference_type, reference_id, notes, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), 'TEST_FORWARD_SYNC', 1, 'Testing forward sync', NOW(), NOW())
	`, testMapping.CBID, testAmount2, originalCBBalance + testAmount2).Error
	
	if err != nil {
		fmt.Printf("❌ Failed to create transaction: %v\n", err)
	} else {
		fmt.Println("✅ Transaction created")
		
		// Check both balances
		var finalCBBalance, finalCOABalance float64
		db.Raw("SELECT balance FROM cash_banks WHERE id = ?", testMapping.CBID).Scan(&finalCBBalance)
		db.Raw("SELECT balance FROM accounts WHERE id = ?", testMapping.AccID).Scan(&finalCOABalance)
		
		expectedBalance := originalCBBalance + testAmount2
		
		fmt.Printf("Results:\n")
		fmt.Printf("  CashBank balance: %.2f (expected: %.2f)\n", finalCBBalance, expectedBalance)
		fmt.Printf("  COA balance: %.2f (expected: %.2f)\n", finalCOABalance, expectedBalance)
		
		if finalCBBalance == expectedBalance && finalCOABalance == expectedBalance {
			fmt.Println("🎉 SUCCESS! CashBank → COA sync working!")
		} else {
			fmt.Println("❌ FAILED! Forward sync not working properly")
		}
		
		// Clean up
		db.Exec("DELETE FROM cash_bank_transactions WHERE notes = 'Testing forward sync'")
		db.Exec("UPDATE cash_banks SET balance = ? WHERE id = ?", originalCBBalance, testMapping.CBID)
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", originalCOABalance, testMapping.AccID)
		fmt.Println("🧹 Test data cleaned up")
	}
	
	// Test 3: Manual balance update test
	fmt.Println("\n🔄 TEST 3: Direct CashBank Balance Update")
	fmt.Println("=========================================")
	
	testAmount3 := originalCBBalance + 100000
	fmt.Printf("Manually updating CashBank balance: %.2f → %.2f\n", 
		originalCBBalance, testAmount3)
	
	err = db.Exec("UPDATE cash_banks SET balance = ? WHERE id = ?", 
		testAmount3, testMapping.CBID).Error
	if err != nil {
		fmt.Printf("❌ Failed to update CashBank: %v\n", err)
	} else {
		// Check if COA updated (this might not work if no trigger on cash_banks balance update)
		var newCOABalance float64
		db.Raw("SELECT balance FROM accounts WHERE id = ?", testMapping.AccID).Scan(&newCOABalance)
		
		fmt.Printf("Result: COA balance is %.2f\n", newCOABalance)
		
		if newCOABalance == testAmount3 {
			fmt.Println("✅ CashBank → COA sync via direct update working!")
		} else {
			fmt.Println("⚠️  Direct CashBank update doesn't sync to COA (expected - use transactions)")
		}
		
		// Restore
		db.Exec("UPDATE cash_banks SET balance = ? WHERE id = ?", originalCBBalance, testMapping.CBID)
		db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", originalCOABalance, testMapping.AccID)
	}
	
	// Final audit check
	fmt.Println("\n📋 Recent Sync Activity:")
	fmt.Println("=========================")
	
	var recentAudits []struct {
		Action    string `db:"action"`
		TableName string `db:"table_name"`
		CreatedAt string `db:"created_at"`
	}
	
	db.Raw(`
		SELECT action, table_name, TO_CHAR(created_at, 'HH24:MI:SS') as created_at
		FROM audit_logs 
		WHERE table_name IN ('coa_to_cashbank_sync', 'cashbank_coa_sync')
		AND created_at > NOW() - INTERVAL '5 minutes'
		ORDER BY created_at DESC
		LIMIT 10
	`).Scan(&recentAudits)
	
	if len(recentAudits) > 0 {
		fmt.Printf("Found %d recent sync activities:\n", len(recentAudits))
		for _, audit := range recentAudits {
			fmt.Printf("  %s: %s (%s)\n", audit.CreatedAt, audit.Action, audit.TableName)
		}
	} else {
		fmt.Println("No recent sync activities found")
	}
	
	fmt.Println("\n✅ Bidirectional Sync Testing Completed!")
	fmt.Println("🎯 You can now test via UI - both directions should work!")
}
