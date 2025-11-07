package main

import (
	"encoding/json"
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
)

func main() {
	// Initialize database connection
	db := database.ConnectDB()

	fmt.Println("🔍 ANALYZING ACTUAL FRONTEND ENDPOINTS")
	fmt.Println("=====================================")

	// 1. Check /api/v1/accounts/hierarchy endpoint (yang digunakan frontend)
	fmt.Println("\n1️⃣ ENDPOINT: /api/v1/accounts/hierarchy")
	fmt.Println("-------------------------------------")
	
	// Simulate GetAccountHierarchy method
	type AccountHierarchy struct {
		ID          uint    `json:"id"`
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		Type        string  `json:"type"`
		Balance     float64 `json:"balance"`
		IsHeader    bool    `json:"is_header"`
		IsActive    bool    `json:"is_active"`
		Level       int     `json:"level"`
	}
	
	var hierarchyAccounts []AccountHierarchy
	err := db.Raw(`
		SELECT id, code, name, type, balance, is_header, is_active, level
		FROM accounts 
		WHERE is_active = true 
		ORDER BY code
	`).Scan(&hierarchyAccounts).Error
	
	if err != nil {
		log.Printf("Error getting hierarchy: %v", err)
		return
	}
	
	fmt.Printf("accounts/hierarchy returns %d accounts\n", len(hierarchyAccounts))
	
	// Find account 4101
	var found4101 *AccountHierarchy
	for _, acc := range hierarchyAccounts {
		if acc.Code == "4101" {
			found4101 = &acc
			break
		}
	}
	
	if found4101 != nil {
		fmt.Printf("✅ Account 4101 in hierarchy:\n")
		fmt.Printf("   Balance: %.0f\n", found4101.Balance)
		fmt.Printf("   IsActive: %v\n", found4101.IsActive)
		fmt.Printf("   IsHeader: %v\n", found4101.IsHeader)
		
		if found4101.Balance == 0 {
			fmt.Println("❌ HIERARCHY ENDPOINT RETURNS 0 BALANCE!")
		} else {
			fmt.Println("✅ Hierarchy endpoint has correct balance")
		}
	} else {
		fmt.Println("❌ Account 4101 NOT found in hierarchy")
	}

	// 2. Check /api/v1/ssot-reports/account-balances endpoint
	fmt.Println("\n2️⃣ ENDPOINT: /api/v1/ssot-reports/account-balances")
	fmt.Println("------------------------------------------------")
	
	type SSOTBalance struct {
		AccountCode    string  `json:"account_code"`
		AccountName    string  `json:"account_name"`
		CurrentBalance float64 `json:"current_balance"`
		AccountType    string  `json:"account_type"`
	}
	
	var ssotBalances []SSOTBalance
	err = db.Raw(`
		SELECT 
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(a.balance, 0) as current_balance
		FROM accounts a
		WHERE a.is_active = true
		ORDER BY a.code
	`).Scan(&ssotBalances).Error
	
	if err != nil {
		log.Printf("Error getting SSOT balances: %v", err)
		return
	}
	
	fmt.Printf("ssot-reports/account-balances returns %d accounts\n", len(ssotBalances))
	
	// Find account 4101 in SSOT
	var foundSSOT4101 *SSOTBalance
	for _, acc := range ssotBalances {
		if acc.AccountCode == "4101" {
			foundSSOT4101 = &acc
			break
		}
	}
	
	if foundSSOT4101 != nil {
		fmt.Printf("✅ Account 4101 in SSOT endpoint:\n")
		fmt.Printf("   Balance: %.0f\n", foundSSOT4101.CurrentBalance)
		fmt.Printf("   Type: %s\n", foundSSOT4101.AccountType)
		
		if foundSSOT4101.CurrentBalance == 0 {
			fmt.Println("❌ SSOT ENDPOINT RETURNS 0 BALANCE!")
		} else {
			fmt.Println("✅ SSOT endpoint has correct balance")
		}
	} else {
		fmt.Println("❌ Account 4101 NOT found in SSOT endpoint")
	}

	// 3. Check actual database balance again
	fmt.Println("\n3️⃣ DIRECT DATABASE CHECK:")
	fmt.Println("-------------------------")
	
	var directBalance float64
	err = db.Raw("SELECT balance FROM accounts WHERE code = '4101' AND is_active = true").Scan(&directBalance).Error
	if err != nil {
		log.Printf("Error getting direct balance: %v", err)
	} else {
		fmt.Printf("Direct DB balance for 4101: %.0f\n", directBalance)
	}

	// 4. Check if there's any data transformation in the frontend
	fmt.Println("\n4️⃣ CHECKING FOR DATA TRANSFORMATION ISSUES:")
	fmt.Println("------------------------------------------")
	
	// Test JSON marshaling to see if there are any issues
	if found4101 != nil {
		jsonData, err := json.Marshal(found4101)
		if err != nil {
			fmt.Printf("❌ JSON marshal error: %v\n", err)
		} else {
			fmt.Printf("✅ JSON output for 4101: %s\n", string(jsonData))
		}
	}

	// 5. Final summary
	fmt.Println("\n5️⃣ SUMMARY:")
	fmt.Println("----------")
	
	fmt.Printf("Database balance: %.0f\n", directBalance)
	if found4101 != nil {
		fmt.Printf("Hierarchy endpoint: %.0f\n", found4101.Balance)
	}
	if foundSSOT4101 != nil {
		fmt.Printf("SSOT endpoint: %.0f\n", foundSSOT4101.CurrentBalance)
	}
	
	fmt.Println("\n🎯 DIAGNOSIS:")
	if directBalance > 0 && found4101 != nil && found4101.Balance == 0 {
		fmt.Println("❌ Issue: Database has balance but API endpoint returns 0")
		fmt.Println("   This indicates a problem in the repository/service layer")
	} else if directBalance > 0 && found4101 != nil && found4101.Balance > 0 {
		fmt.Println("✅ Backend is working correctly")
		fmt.Println("   Problem is likely in frontend data handling or display")
	} else {
		fmt.Println("❌ Issue: Database balance is actually 0")
		fmt.Println("   Need to check why balance is not being updated properly")
	}
	
	fmt.Println("\n📋 NEXT ACTIONS:")
	fmt.Println("1. Check frontend JavaScript for data processing issues")
	fmt.Println("2. Check if frontend is using the correct field for balance display") 
	fmt.Println("3. Check browser DevTools Network tab for API responses")
}