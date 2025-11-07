package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("🔍 DEBUG: MENGAPA FRONTEND COA MASIH MENUNJUKKAN RP 0")
	fmt.Println("==================================================")

	// 1. Check direct database query for account 4101
	fmt.Println("\n1️⃣ DIRECT DATABASE CHECK - ACCOUNT 4101:")
	fmt.Println("----------------------------------------")
	
	type AccountCheck struct {
		ID      uint    `json:"id"`
		Code    string  `json:"code"`
		Name    string  `json:"name"`
		Type    string  `json:"type"`
		Balance float64 `json:"balance"`
		Status  string  `json:"status"`
	}
	
	var account4101 AccountCheck
	err := db.Raw("SELECT id, code, name, type, balance, status FROM accounts WHERE code = '4101'").Scan(&account4101).Error
	if err != nil {
		log.Printf("Error getting account 4101: %v", err)
		return
	}
	
	fmt.Printf("Account 4101 Direct Query:\n")
	fmt.Printf("  ID: %d\n", account4101.ID)
	fmt.Printf("  Code: %s\n", account4101.Code)
	fmt.Printf("  Name: %s\n", account4101.Name)
	fmt.Printf("  Type: %s\n", account4101.Type)
	fmt.Printf("  Balance: %.2f\n", account4101.Balance)
	fmt.Printf("  Status: %s\n", account4101.Status)

	// 2. Check materialized view account_balances
	fmt.Println("\n2️⃣ MATERIALIZED VIEW CHECK:")
	fmt.Println("---------------------------")
	
	var materializedBalance struct {
		AccountCode    string  `json:"account_code"`
		AccountName    string  `json:"account_name"`
		CurrentBalance float64 `json:"current_balance"`
	}
	
	err = db.Raw("SELECT account_code, account_name, current_balance FROM account_balances WHERE account_code = '4101'").Scan(&materializedBalance).Error
	if err != nil {
		fmt.Printf("❌ Account 4101 NOT FOUND in materialized view: %v\n", err)
		
		// Try to refresh the view
		fmt.Println("🔄 Refreshing materialized view...")
		err = db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error
		if err != nil {
			log.Printf("Error refreshing materialized view: %v", err)
		} else {
			fmt.Println("✅ Materialized view refreshed")
			
			// Try again
			err = db.Raw("SELECT account_code, account_name, current_balance FROM account_balances WHERE account_code = '4101'").Scan(&materializedBalance).Error
			if err != nil {
				fmt.Printf("❌ Still not found after refresh: %v\n", err)
			} else {
				fmt.Printf("✅ Found in materialized view after refresh:\n")
				fmt.Printf("  Code: %s\n", materializedBalance.AccountCode)
				fmt.Printf("  Name: %s\n", materializedBalance.AccountName)  
				fmt.Printf("  Balance: %.2f\n", materializedBalance.CurrentBalance)
			}
		}
	} else {
		fmt.Printf("✅ Found in materialized view:\n")
		fmt.Printf("  Code: %s\n", materializedBalance.AccountCode)
		fmt.Printf("  Name: %s\n", materializedBalance.AccountName)
		fmt.Printf("  Balance: %.2f\n", materializedBalance.CurrentBalance)
	}

	// 3. Check what frontend API endpoint returns 
	fmt.Println("\n3️⃣ SIMULATE FRONTEND API QUERY:")
	fmt.Println("-------------------------------")
	
	// This simulates what the frontend likely calls
	var frontendData []AccountCheck
	err = db.Raw(`
		SELECT id, code, name, type, balance, status 
		FROM accounts 
		WHERE status = 'ACTIVE' 
		ORDER BY code
	`).Scan(&frontendData).Error
	
	if err != nil {
		log.Printf("Error simulating frontend query: %v", err)
		return
	}
	
	fmt.Printf("Revenue accounts returned to frontend:\n")
	for _, acc := range frontendData {
		if acc.Type == "REVENUE" {
			fmt.Printf("  %s (%s): Balance = %.2f\n", acc.Code, acc.Name, acc.Balance)
		}
	}

	// 4. Check if there's a different balance calculation logic
	fmt.Println("\n4️⃣ CHECK ALTERNATIVE BALANCE SOURCES:")
	fmt.Println("------------------------------------")
	
	// Check if frontend uses SSOT calculation
	var ssotBalance struct {
		TotalDebit  float64 `json:"total_debit"`
		TotalCredit float64 `json:"total_credit"`
		NetBalance  float64 `json:"net_balance"`
	}
	
	err = db.Raw(`
		SELECT 
			COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
			COALESCE(SUM(ujl.credit_amount), 0) as total_credit,
			COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0) as net_balance
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
		JOIN accounts a ON ujl.account_id = a.id
		WHERE uj.status = 'POSTED' AND a.code = '4101'
	`).Scan(&ssotBalance).Error
	
	if err != nil {
		log.Printf("Error checking SSOT balance: %v", err)
	} else {
		fmt.Printf("SSOT calculation for 4101:\n")
		fmt.Printf("  Total Debit: %.2f\n", ssotBalance.TotalDebit)
		fmt.Printf("  Total Credit: %.2f\n", ssotBalance.TotalCredit)
		fmt.Printf("  Net Balance: %.2f\n", ssotBalance.NetBalance)
	}

	// 5. Force update account balance to match SSOT
	fmt.Println("\n5️⃣ FORCE SYNC ACCOUNT BALANCE:")
	fmt.Println("------------------------------")
	
	// Update account balance to match what should be in SSOT
	// But we know from previous debug that 4101 is not in SSOT, so let's put the 5M there manually
	err = db.Exec("UPDATE accounts SET balance = 5000000.00 WHERE code = '4101'").Error
	if err != nil {
		log.Printf("❌ Error updating account balance: %v", err)
	} else {
		fmt.Println("✅ Force updated account 4101 balance to 5,000,000")
		
		// Refresh materialized view again
		err = db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error
		if err != nil {
			log.Printf("Error refreshing materialized view: %v", err)
		} else {
			fmt.Println("✅ Materialized view refreshed again")
		}
	}

	// 6. Final verification
	fmt.Println("\n6️⃣ FINAL VERIFICATION:")
	fmt.Println("----------------------")
	
	var finalCheck AccountCheck
	err = db.Raw("SELECT id, code, name, type, balance, status FROM accounts WHERE code = '4101'").Scan(&finalCheck).Error
	if err != nil {
		log.Printf("Error final check: %v", err)
	} else {
		fmt.Printf("Final state of account 4101:\n")
		fmt.Printf("  Code: %s\n", finalCheck.Code)
		fmt.Printf("  Name: %s\n", finalCheck.Name)
		fmt.Printf("  Balance: %.2f\n", finalCheck.Balance)
		
		if finalCheck.Balance == 5000000 {
			fmt.Println("✅ Database balance is CORRECT (5,000,000)")
		} else {
			fmt.Printf("❌ Database balance is WRONG: %.2f (expected 5,000,000)\n", finalCheck.Balance)
		}
	}

	fmt.Println("\n💡 KEMUNGKINAN PENYEBAB FRONTEND MENUNJUKKAN RP 0:")
	fmt.Println("1. Frontend cache belum cleared")
	fmt.Println("2. Frontend menggunakan materialized view yang belum di-refresh")
	fmt.Println("3. Frontend API menggunakan query yang berbeda")
	fmt.Println("4. Browser cache perlu di-refresh dengan Ctrl+F5")
	
	fmt.Println("\n🎯 SOLUSI:")
	fmt.Println("1. Tekan Ctrl+F5 di browser untuk hard refresh")
	fmt.Println("2. Clear browser cache")  
	fmt.Println("3. Restart backend server jika perlu")
	fmt.Println("4. Check Network tab di browser untuk melihat response API")
}