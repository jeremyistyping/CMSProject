package main

import (
	"fmt"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🔍 Debugging CashBank Data Integrity...")
	
	// Load configuration and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()
	
	// Check CashBank data
	fmt.Println("\n📊 CashBank Accounts Analysis:")
	fmt.Println("=====================================")
	
	var cashBanks []struct {
		ID        uint    `db:"id"`
		Code      string  `db:"code"`
		Name      string  `db:"name"`
		Balance   float64 `db:"balance"`
		AccountID *uint   `db:"account_id"`
		IsActive  bool    `db:"is_active"`
	}
	
	db.Raw(`
		SELECT id, code, name, balance, account_id, is_active
		FROM cash_banks 
		WHERE deleted_at IS NULL
		ORDER BY id
	`).Scan(&cashBanks)
	
	fmt.Printf("Found %d CashBank accounts:\n", len(cashBanks))
	for _, cb := range cashBanks {
		accountInfo := "NOT LINKED"
		if cb.AccountID != nil && *cb.AccountID > 0 {
			accountInfo = fmt.Sprintf("→ COA ID %d", *cb.AccountID)
		}
		activeStatus := "INACTIVE"
		if cb.IsActive {
			activeStatus = "ACTIVE"
		}
		
		fmt.Printf("  ID:%d %s \"%s\" Balance:%.2f %s [%s]\n", 
			cb.ID, cb.Code, cb.Name, cb.Balance, accountInfo, activeStatus)
	}
	
	// Check COA accounts that should be linked
	fmt.Println("\n📊 COA Bank Accounts Analysis:")
	fmt.Println("=====================================")
	
	var coaAccounts []struct {
		ID      uint    `db:"id"`
		Code    string  `db:"code"`
		Name    string  `db:"name"`
		Balance float64 `db:"balance"`
	}
	
	db.Raw(`
		SELECT id, code, name, balance
		FROM accounts 
		WHERE deleted_at IS NULL 
		  AND code IN ('1101', '1102', '1103', '1105')
		ORDER BY code
	`).Scan(&coaAccounts)
	
	fmt.Printf("Found %d COA bank accounts:\n", len(coaAccounts))
	for _, acc := range coaAccounts {
		fmt.Printf("  %s %s Balance:%.2f\n", acc.Code, acc.Name, acc.Balance)
	}
	
	// Find linking issues
	fmt.Println("\n🔧 Fixing CashBank Data Issues...")
	fmt.Println("=====================================")
	
	// Fix empty names
	emptyNameCount := 0
	for _, cb := range cashBanks {
		if cb.Name == "" {
			// Try to derive name from COA account
			if cb.AccountID != nil && *cb.AccountID > 0 {
				var coaName string
				db.Raw("SELECT name FROM accounts WHERE id = ?", *cb.AccountID).Scan(&coaName)
				if coaName != "" {
					newName := fmt.Sprintf("CashBank %s", coaName)
					db.Exec("UPDATE cash_banks SET name = ? WHERE id = ?", newName, cb.ID)
					fmt.Printf("  ✅ Updated CashBank ID %d name to: %s\n", cb.ID, newName)
					emptyNameCount++
				}
			} else {
				// Generic name based on code
				newName := fmt.Sprintf("CashBank %s", cb.Code)
				db.Exec("UPDATE cash_banks SET name = ? WHERE id = ?", newName, cb.ID)
				fmt.Printf("  ✅ Updated CashBank ID %d name to: %s\n", cb.ID, newName)
				emptyNameCount++
			}
		}
	}
	
	if emptyNameCount > 0 {
		fmt.Printf("Fixed %d CashBank accounts with empty names\n", emptyNameCount)
	} else {
		fmt.Println("All CashBank accounts have valid names")
	}
	
	// Test trigger with proper data
	fmt.Println("\n🔄 Testing Reverse Sync Trigger with Fixed Data:")
	fmt.Println("=====================================")
	
	// Get updated data
	var linkedAccount struct {
		CashBankID   uint    `db:"cashbank_id"`
		CashBankName string  `db:"cashbank_name"`
		AccountID    uint    `db:"account_id"`
		AccountCode  string  `db:"account_code"`
		COABalance   float64 `db:"coa_balance"`
		CBBalance    float64 `db:"cb_balance"`
	}
	
	err := db.Raw(`
		SELECT 
			cb.id as cashbank_id,
			cb.name as cashbank_name,
			a.id as account_id,
			a.code as account_code,
			a.balance as coa_balance,
			cb.balance as cb_balance
		FROM cash_banks cb 
		JOIN accounts a ON cb.account_id = a.id 
		WHERE cb.deleted_at IS NULL 
		  AND cb.is_active = true
		  AND cb.account_id IS NOT NULL
		  AND cb.account_id > 0
		  AND cb.id > 0
		LIMIT 1
	`).Scan(&linkedAccount).Error
	
	if err != nil || linkedAccount.AccountID == 0 {
		fmt.Println("❌ Still no valid linked accounts found")
		
		// Show what we have
		var debugData []struct {
			CBID    uint   `db:"cb_id"`
			CBName  string `db:"cb_name"`
			AccID   *uint  `db:"acc_id"`
			AccCode string `db:"acc_code"`
		}
		
		db.Raw(`
			SELECT 
				cb.id as cb_id,
				cb.name as cb_name,
				cb.account_id as acc_id,
				COALESCE(a.code, 'NULL') as acc_code
			FROM cash_banks cb 
			LEFT JOIN accounts a ON cb.account_id = a.id 
			WHERE cb.deleted_at IS NULL 
		`).Scan(&debugData)
		
		fmt.Printf("Debug - Current linking status:\n")
		for _, d := range debugData {
			accInfo := "NOT LINKED"
			if d.AccID != nil && *d.AccID > 0 {
				accInfo = fmt.Sprintf("→ %s", d.AccCode)
			}
			fmt.Printf("  CashBank %d \"%s\" %s\n", d.CBID, d.CBName, accInfo)
		}
		
	} else {
		fmt.Printf("✅ Found valid linked account:\n")
		fmt.Printf("   CashBank: ID %d \"%s\" (%.2f)\n", 
			linkedAccount.CashBankID, linkedAccount.CashBankName, linkedAccount.CBBalance)
		fmt.Printf("   COA: %s (%.2f)\n", 
			linkedAccount.AccountCode, linkedAccount.COABalance)
		
		// Test the trigger
		testBalance := linkedAccount.COABalance + 50000
		fmt.Printf("\n🔧 Testing: Update COA %s from %.2f to %.2f\n", 
			linkedAccount.AccountCode, linkedAccount.COABalance, testBalance)
		
		// Update COA balance - this should trigger sync
		err = db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", 
			testBalance, linkedAccount.AccountID).Error
		
		if err != nil {
			fmt.Printf("❌ COA update failed: %v\n", err)
		} else {
			fmt.Println("✅ COA balance updated")
			
			// Check CashBank balance
			var newCBBalance float64
			db.Raw("SELECT balance FROM cash_banks WHERE id = ?", 
				linkedAccount.CashBankID).Scan(&newCBBalance)
			
			fmt.Printf("Result: CashBank balance is now %.2f\n", newCBBalance)
			
			if newCBBalance == testBalance {
				fmt.Println("🎉 SUCCESS! Reverse sync (COA → CashBank) is working!")
			} else {
				fmt.Printf("❌ FAILED! Expected %.2f, got %.2f\n", testBalance, newCBBalance)
				
				// Manual debug - check audit logs for trigger activity
				var auditCount int64
				db.Raw(`
					SELECT COUNT(*) FROM audit_logs 
					WHERE table_name = 'coa_to_cashbank_sync' 
					AND created_at > NOW() - INTERVAL '1 minute'
				`).Scan(&auditCount)
				
				fmt.Printf("Recent audit entries for sync: %d\n", auditCount)
				
				if auditCount > 0 {
					fmt.Println("✅ Trigger fired but sync failed - check trigger logic")
				} else {
					fmt.Println("❌ Trigger did not fire - check trigger installation")
				}
			}
			
			// Restore original values
			db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", 
				linkedAccount.COABalance, linkedAccount.AccountID)
			db.Exec("UPDATE cash_banks SET balance = ? WHERE id = ?", 
				linkedAccount.CBBalance, linkedAccount.CashBankID)
			fmt.Println("🔄 Original values restored")
		}
	}
	
	fmt.Println("\n✅ CashBank data debugging completed!")
}
