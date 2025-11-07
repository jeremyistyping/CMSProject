package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 Fix SSOT Account Balances & Test Authentication")
	fmt.Println("=================================================")
	fmt.Println()

	// Initialize database
	db := database.ConnectDB()

	// Step 1: Check current SSOT status
	fmt.Println("📊 Step 1: Checking Current SSOT Status...")
	checkCurrentSSOTStatus(db)
	fmt.Println()

	// Step 2: Create account_balances view if missing
	fmt.Println("🏗️  Step 2: Creating/Refreshing Account Balances View...")
	createAccountBalancesView(db)
	fmt.Println()

	// Step 3: Test authentication with debug info
	fmt.Println("🔐 Step 3: Testing Authentication (Debug Mode)...")
	testAuthenticationDebug()
	fmt.Println()

	// Step 4: Test one API endpoint manually with auth
	fmt.Println("🧪 Step 4: Testing One API Endpoint...")
	testSingleEndpoint()
}

func checkCurrentSSOTStatus(db *gorm.DB) {
	// Check journal entries
	var journalCount int64
	db.Table("journal_entries").Count(&journalCount)
	fmt.Printf("✅ Journal Entries: %d entries\n", journalCount)

	if journalCount > 0 {
		// Get sample entries
		var entries []struct {
			ID          uint    `json:"id"`
			EntryDate   string  `json:"entry_date"`
			Reference   string  `json:"reference"`
			Description string  `json:"description"`
			TotalDebit  float64 `json:"total_debit"`
			TotalCredit float64 `json:"total_credit"`
		}
		db.Table("journal_entries").Select("id, DATE(entry_date) as entry_date, reference, description, total_debit, total_credit").Limit(5).Find(&entries)
		
		fmt.Println("📝 Recent Journal Entries:")
		for _, entry := range entries {
			fmt.Printf("   ID: %d, Date: %s, Desc: %s, Debit: %.2f, Credit: %.2f\n", 
				entry.ID, entry.EntryDate, entry.Description, entry.TotalDebit, entry.TotalCredit)
		}
	}

	// Check accounts
	var accountCount int64
	db.Table("accounts").Count(&accountCount)
	fmt.Printf("✅ Chart of Accounts: %d accounts\n", accountCount)

	// Check account_balances view
	var balanceCount int64
	err := db.Table("account_balances").Count(&balanceCount)
	if err != nil {
		fmt.Printf("❌ Account Balances View: Does not exist\n")
	} else {
		fmt.Printf("✅ Account Balances View: %d records\n", balanceCount)
	}
}

func createAccountBalancesView(db *gorm.DB) {
	fmt.Println("🔄 Creating account_balances materialized view...")

	// Drop existing view if exists
	dropQuery := "DROP VIEW IF EXISTS account_balances"
	result := db.Exec(dropQuery)
	if result.Error != nil {
		fmt.Printf("⚠️  Warning dropping view: %v\n", result.Error)
	}

	// Create account_balances view
	createViewQuery := `
	CREATE VIEW account_balances AS
	SELECT 
		a.id as account_id,
		a.code as account_code,
		a.name as account_name,
		a.account_type,
		COALESCE(SUM(jd.debit_amount), 0) as total_debit,
		COALESCE(SUM(jd.credit_amount), 0) as total_credit,
		CASE 
			WHEN a.account_type IN ('ASSET', 'EXPENSE') THEN 
				COALESCE(SUM(jd.debit_amount), 0) - COALESCE(SUM(jd.credit_amount), 0)
			ELSE 
				COALESCE(SUM(jd.credit_amount), 0) - COALESCE(SUM(jd.debit_amount), 0)
		END as balance,
		NOW() as last_updated
	FROM accounts a
	LEFT JOIN journal_details jd ON a.id = jd.account_id
	LEFT JOIN journal_entries je ON jd.journal_entry_id = je.id
	WHERE a.is_active = 1
	GROUP BY a.id, a.code, a.name, a.account_type
	`

	result = db.Exec(createViewQuery)
	if result.Error != nil {
		fmt.Printf("❌ Error creating view: %v\n", result.Error)
	} else {
		fmt.Println("✅ Account balances view created successfully!")

		// Check the new view
		var count int64
		db.Table("account_balances").Count(&count)
		fmt.Printf("✅ Account balances records: %d\n", count)

		// Show sample balances
		var balances []struct {
			AccountCode string  `json:"account_code"`
			AccountName string  `json:"account_name"`
			AccountType string  `json:"account_type"`
			Balance     float64 `json:"balance"`
		}
		db.Table("account_balances").Select("account_code, account_name, account_type, balance").Where("balance != 0").Limit(5).Find(&balances)
		
		if len(balances) > 0 {
			fmt.Println("💰 Sample Account Balances:")
			for _, bal := range balances {
				fmt.Printf("   %s - %s (%s): %.2f\n", 
					bal.AccountCode, bal.AccountName, bal.AccountType, bal.Balance)
			}
		}
	}
}

func testAuthenticationDebug() {
	fmt.Println("🔍 Testing login with debug information...")

	loginData := map[string]string{
		"email":    "admin@company.com",
		"password": "password123",
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		fmt.Printf("❌ JSON Marshal error: %v\n", err)
		return
	}

	fmt.Printf("📤 Sending login request to: http://localhost:8080/api/v1/auth/login\n")
	fmt.Printf("📤 Request payload: %s\n", string(jsonData))

	resp, err := http.Post("http://localhost:8080/api/v1/auth/login", 
		"application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ HTTP request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📥 Response Status: %d\n", resp.StatusCode)
	fmt.Printf("📥 Response Headers: %v\n", resp.Header)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("📥 Raw Response Body: %s\n", string(body))

	// Try to parse as JSON
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err == nil {
		fmt.Println("📊 Parsed Response:")
		prettyJSON, _ := json.MarshalIndent(responseData, "", "  ")
		fmt.Println(string(prettyJSON))

		// Check for token in different possible locations
		if token, exists := responseData["token"]; exists && token != nil {
			fmt.Printf("✅ Token found: %v\n", token)
		} else if data, exists := responseData["data"]; exists {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if token, exists := dataMap["token"]; exists {
					fmt.Printf("✅ Token found in data: %v\n", token)
				}
			}
		} else {
			fmt.Println("⚠️  No token found in response")
		}
	} else {
		fmt.Printf("❌ JSON parse error: %v\n", err)
	}
}

func testSingleEndpoint() {
	fmt.Println("🧪 Testing Profit & Loss endpoint directly...")

	// First try without authentication
	fmt.Println("1. Testing without authentication...")
	url := "http://localhost:8080/api/v1/reports/profit-loss?start_date=2024-01-01&end_date=2025-12-31&format=json"
	
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("📥 Status: %d\n", resp.StatusCode)
	fmt.Printf("📥 Response Size: %d bytes\n", len(body))
	
	if resp.StatusCode == 401 {
		fmt.Println("✅ Correctly requires authentication")
		
		// Check if we can see the auth requirement details
		fmt.Printf("📥 Auth Error: %s\n", string(body))
	} else if resp.StatusCode == 200 {
		fmt.Println("⚠️  No authentication required - checking data...")
		
		// Check for data
		bodyStr := string(body)
		if strings.Contains(bodyStr, "No P&L relevant transactions found") {
			fmt.Println("❌ Response: No P&L data found")
		} else if strings.Contains(bodyStr, "total") || strings.Contains(bodyStr, "revenue") {
			fmt.Println("✅ Response contains financial data!")
		} else {
			fmt.Printf("📋 Response preview: %s...\n", bodyStr[:min(200, len(bodyStr))])
		}
	}

	// Test with admin user existence
	fmt.Println("\n2. Testing admin user existence...")
	testAdminUserExists()
}

func testAdminUserExists() {
	db := database.ConnectDB()
	
	var userCount int64
	db.Table("users").Where("email = ?", "admin@company.com").Count(&userCount)
	
	if userCount > 0 {
		fmt.Println("✅ Admin user exists in database")
		
		// Get user details
		var user struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Role     string `json:"role"`
			IsActive bool   `json:"is_active"`
		}
		db.Table("users").Where("email = ?", "admin@company.com").First(&user)
		fmt.Printf("   User: %s (ID: %d, Role: %s, Active: %t)\n", 
			user.Username, user.ID, user.Role, user.IsActive)
	} else {
		fmt.Println("❌ Admin user does not exist!")
		fmt.Println("💡 You may need to create the admin user first")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}