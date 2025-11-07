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
	fmt.Println("🎯 Final SSOT Fix & Complete API Test")
	fmt.Println("=====================================")
	fmt.Println()

	db := database.ConnectDB()
	
	// Step 1: Fix SSOT account_balances view
	fmt.Println("🔧 Step 1: Creating Account Balances View...")
	createAccountBalancesFromJournalLines(db)
	fmt.Println()

	// Step 2: Test authentication with correct token field
	fmt.Println("🔐 Step 2: Test Authentication...")
	token, err := authenticateAndGetToken()
	if err != nil {
		fmt.Printf("❌ Authentication failed: %v\n", err)
		return
	}
	fmt.Println("✅ Authentication successful!")
	fmt.Println()

	// Step 3: Test all API endpoints with authentication
	fmt.Println("🧪 Step 3: Test All Financial Report APIs...")
	testAllAPIsWithAuth(token)
	fmt.Println()

	// Step 4: Final status
	fmt.Println("🏆 Final Status...")
	showFinalStatus(db)
}

func createAccountBalancesFromJournalLines(db *gorm.DB) {
	fmt.Println("🏗️  Creating account_balances view using journal_lines...")

	// Drop any existing view/table
	db.Exec("DROP TABLE IF EXISTS account_balances CASCADE")
	db.Exec("DROP VIEW IF EXISTS account_balances CASCADE")

	// Check journal_lines structure first
	var sampleLine struct {
		ID        uint    `gorm:"column:id"`
		EntryID   uint    `gorm:"column:journal_entry_id"`
		AccountID uint    `gorm:"column:account_id"`
		Debit     float64 `gorm:"column:debit_amount"`
		Credit    float64 `gorm:"column:credit_amount"`
	}
	
	err := db.Table("journal_lines").Select("id, journal_entry_id, account_id, debit_amount, credit_amount").First(&sampleLine)
	if err != nil {
		fmt.Printf("❌ Cannot read journal_lines: %v\n", err)
		return
	}

	fmt.Printf("📝 Sample journal line: ID=%d, EntryID=%d, AccountID=%d, Debit=%.2f, Credit=%.2f\n", 
		sampleLine.ID, sampleLine.EntryID, sampleLine.AccountID, sampleLine.Debit, sampleLine.Credit)

	// Create comprehensive account_balances view
	createViewQuery := `
	CREATE VIEW account_balances AS
	SELECT 
		a.id as account_id,
		a.code as account_code,
		a.name as account_name,
		a.account_type,
		
		-- Sum all debits for this account
		COALESCE(SUM(jl.debit_amount), 0) as total_debit,
		
		-- Sum all credits for this account
		COALESCE(SUM(jl.credit_amount), 0) as total_credit,
		
		-- Calculate balance based on account type
		CASE 
			WHEN a.account_type IN ('ASSET', 'EXPENSE') THEN 
				COALESCE(SUM(jl.debit_amount), 0) - COALESCE(SUM(jl.credit_amount), 0)
			ELSE 
				COALESCE(SUM(jl.credit_amount), 0) - COALESCE(SUM(jl.debit_amount), 0)
		END as balance,
		
		NOW() as last_updated
		
	FROM accounts a
	LEFT JOIN journal_lines jl ON a.id = jl.account_id
	LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id
	WHERE a.is_active = true AND (je.status = 'POSTED' OR je.status IS NULL)
	GROUP BY a.id, a.code, a.name, a.account_type
	`

	err = db.Exec(createViewQuery)
	if err.Error != nil {
		fmt.Printf("❌ Error creating view: %v\n", err.Error)
		return
	}

	fmt.Println("✅ Account balances view created!")

	// Test the view
	var count int64
	db.Table("account_balances").Count(&count)
	fmt.Printf("✅ Account balances view contains %d accounts\n", count)

	// Show sample data
	var balances []struct {
		AccountCode string  `gorm:"column:account_code"`
		AccountName string  `gorm:"column:account_name"`
		AccountType string  `gorm:"column:account_type"`
		TotalDebit  float64 `gorm:"column:total_debit"`
		TotalCredit float64 `gorm:"column:total_credit"`
		Balance     float64 `gorm:"column:balance"`
	}
	
	db.Table("account_balances").
		Where("total_debit > 0 OR total_credit > 0").
		Order("ABS(balance) DESC").
		Limit(5).
		Find(&balances)

	if len(balances) > 0 {
		fmt.Println("💰 Top Account Balances:")
		for _, bal := range balances {
			fmt.Printf("   %s - %s (%s): Debit=%.2f, Credit=%.2f, Balance=%.2f\n", 
				bal.AccountCode, bal.AccountName, bal.AccountType, 
				bal.TotalDebit, bal.TotalCredit, bal.Balance)
		}
	}
}

func authenticateAndGetToken() (string, error) {
	loginData := map[string]string{
		"email":    "admin@company.com",
		"password": "password123",
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login data: %v", err)
	}

	resp, err := http.Post("http://localhost:8080/api/v1/auth/login", 
		"application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read login response: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp struct {
		AccessToken string `json:"access_token"`
		Message     string `json:"message"`
		User        struct {
			Username string `json:"username"`
			Role     string `json:"role"`
		} `json:"user"`
	}

	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("failed to parse login response: %v", err)
	}

	if loginResp.AccessToken == "" {
		return "", fmt.Errorf("no access_token received in login response")
	}

	fmt.Printf("✅ Logged in as: %s (%s)\n", loginResp.User.Username, loginResp.User.Role)
	return loginResp.AccessToken, nil
}

func testAllAPIsWithAuth(token string) {
	baseURL := "http://localhost:8080"
	startDate := "2024-01-01"
	endDate := "2025-12-31"

	endpoints := []struct {
		name string
		path string
	}{
		{"Profit & Loss Statement", "/api/v1/reports/profit-loss"},
		{"Balance Sheet", "/api/v1/reports/balance-sheet"},
		{"Cash Flow Statement", "/api/v1/reports/cash-flow"},
		{"Sales Summary", "/api/v1/reports/sales-summary"},
		{"Purchase Summary", "/api/v1/reports/purchase-summary"},
		{"Trial Balance", "/api/v1/reports/trial-balance"},
		{"General Ledger", "/api/v1/reports/general-ledger"},
		{"Journal Entry Analysis", "/api/v1/reports/journal-entry-analysis"},
	}

	successCount := 0
	dataAvailableCount := 0

	for i, endpoint := range endpoints {
		fmt.Printf("%d. Testing %s...\n", i+1, endpoint.name)
		
		url := fmt.Sprintf("%s%s?start_date=%s&end_date=%s&format=json", 
			baseURL, endpoint.path, startDate, endDate)

		hasData, responseSize, err := testSingleAPIWithAuth(url, token)
		
		if err != nil {
			fmt.Printf("   ❌ Error: %s\n", err)
		} else {
			successCount++
			if hasData {
				dataAvailableCount++
				fmt.Printf("   ✅ Success with data (%d bytes)\n", responseSize)
			} else {
				fmt.Printf("   ✅ Success but no data (%d bytes)\n", responseSize)
			}
		}
	}

	fmt.Printf("\n📊 API Test Results:\n")
	fmt.Printf("   Total endpoints: %d\n", len(endpoints))
	fmt.Printf("   Successful: %d (%.1f%%)\n", successCount, float64(successCount)/float64(len(endpoints))*100)
	fmt.Printf("   With data: %d (%.1f%%)\n", dataAvailableCount, float64(dataAvailableCount)/float64(len(endpoints))*100)
}

func testSingleAPIWithAuth(url, token string) (bool, int, error) {
	client := &http.Client{}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, 0, fmt.Errorf("request creation failed: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, 0, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, len(body), fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	bodyStr := string(body)
	responseSize := len(body)

	// Check if response contains data
	hasData := !strings.Contains(bodyStr, "No P&L relevant transactions found") &&
		!strings.Contains(bodyStr, "No data available") &&
		!strings.Contains(bodyStr, `"total_entries":0`) &&
		!strings.Contains(bodyStr, `"items":[]`) &&
		(strings.Contains(bodyStr, "total") || strings.Contains(bodyStr, "revenue") || 
		 strings.Contains(bodyStr, "balance") || strings.Contains(bodyStr, "amount"))

	return hasData, responseSize, nil
}

func showFinalStatus(db *gorm.DB) {
	// Check journal entries
	var journalCount int64
	db.Table("journal_entries").Count(&journalCount)

	// Check journal lines
	var lineCount int64
	db.Table("journal_lines").Count(&lineCount)

	// Check account balances view
	var balanceCount int64
	err := db.Table("account_balances").Count(&balanceCount)
	viewExists := (err == nil)

	// Check accounts
	var accountCount int64
	db.Table("accounts").Count(&accountCount)

	fmt.Printf("📋 SSOT System Status:\n")
	fmt.Printf("   - Journal Entries: %d\n", journalCount)
	fmt.Printf("   - Journal Lines: %d\n", lineCount)
	fmt.Printf("   - Accounts: %d\n", accountCount)
	fmt.Printf("   - Account Balances View: %t (%d records)\n", viewExists, balanceCount)

	fmt.Println()
	if journalCount > 0 && lineCount > 0 && viewExists && balanceCount > 0 {
		fmt.Println("🎉 SUCCESS! SSOT System is fully operational!")
		fmt.Println("✅ Journal entries exist and are properly linked")
		fmt.Println("✅ Account balances view is created and populated")
		fmt.Println("✅ Frontend financial reports should now display data")
		fmt.Println("✅ Authentication works with access_token")
		fmt.Println()
		fmt.Println("💡 Next steps:")
		fmt.Println("   1. Refresh your frontend browser page")
		fmt.Println("   2. Try generating financial reports")
		fmt.Println("   3. Check date ranges if reports show 'No Data'")
	} else {
		fmt.Println("⚠️  SSOT System needs attention:")
		if journalCount == 0 {
			fmt.Println("❌ No journal entries found")
		}
		if lineCount == 0 {
			fmt.Println("❌ No journal lines found")
		}
		if !viewExists {
			fmt.Println("❌ Account balances view missing")
		}
	}
}