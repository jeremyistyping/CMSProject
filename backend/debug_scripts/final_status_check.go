package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"app-sistem-akuntansi/database"
)

func main() {
	fmt.Println("🏆 Final SSOT & API Status Check")
	fmt.Println("================================")
	fmt.Println()

	db := database.ConnectDB()

	// Check SSOT components
	fmt.Println("📊 SSOT System Status:")
	
	var journalEntries, journalLines, accounts int64
	db.Table("journal_entries").Count(&journalEntries)
	db.Table("journal_lines").Count(&journalLines)
	db.Table("accounts").Where("is_active = true").Count(&accounts)
	
	fmt.Printf("   ✅ Journal Entries: %d\n", journalEntries)
	fmt.Printf("   ✅ Journal Lines: %d\n", journalLines)
	fmt.Printf("   ✅ Active Accounts: %d\n", accounts)

	// Check account_balances view
	var balanceCount int64
	err := db.Table("account_balances").Count(&balanceCount)
	viewExists := (err == nil)
	
	if viewExists {
		fmt.Printf("   ✅ Account Balances View: %d records\n", balanceCount)
		
		// Show some active balances
		var activeBalances int64
		db.Table("account_balances").Where("total_debit > 0 OR total_credit > 0").Count(&activeBalances)
		fmt.Printf("   ✅ Accounts with Activity: %d\n", activeBalances)
	} else {
		fmt.Printf("   ❌ Account Balances View: Missing\n")
	}

	fmt.Println()

	// Test Authentication
	fmt.Println("🔐 Authentication Test:")
	token, authErr := testAuthentication()
	if authErr != nil {
		fmt.Printf("   ❌ Authentication: Failed - %v\n", authErr)
		return
	}
	fmt.Printf("   ✅ Authentication: Working\n")
	fmt.Println()

	// Test Key API Endpoints
	fmt.Println("🧪 API Endpoints Test:")
	endpoints := []struct {
		name string
		path string
	}{
		{"Profit & Loss", "/api/v1/reports/profit-loss"},
		{"Balance Sheet", "/api/v1/reports/balance-sheet"},
		{"Trial Balance", "/api/v1/reports/trial-balance"},
		{"Sales Summary", "/api/v1/reports/sales-summary"},
	}

	workingEndpoints := 0
	endpointsWithData := 0

	for _, endpoint := range endpoints {
		hasData, err := testAPIEndpoint(endpoint.path, token)
		if err != nil {
			fmt.Printf("   ❌ %s: %v\n", endpoint.name, err)
		} else {
			workingEndpoints++
			if hasData {
				endpointsWithData++
				fmt.Printf("   ✅ %s: Working with data\n", endpoint.name)
			} else {
				fmt.Printf("   ⚠️  %s: Working but no data\n", endpoint.name)
			}
		}
	}

	fmt.Println()
	fmt.Printf("📋 Summary:\n")
	fmt.Printf("   API Success Rate: %d/%d (%.1f%%)\n", 
		workingEndpoints, len(endpoints), float64(workingEndpoints)/float64(len(endpoints))*100)
	fmt.Printf("   APIs with Data: %d/%d (%.1f%%)\n", 
		endpointsWithData, len(endpoints), float64(endpointsWithData)/float64(len(endpoints))*100)
	
	fmt.Println()
	
	// Final recommendation
	if journalEntries > 0 && journalLines > 0 && viewExists && workingEndpoints == len(endpoints) {
		fmt.Println("🎉 SUCCESS! Everything is working perfectly!")
		fmt.Println("✅ SSOT Journal System is operational")
		fmt.Println("✅ Account Balances view is created")
		fmt.Println("✅ API endpoints are responding")
		fmt.Println("✅ Authentication is working")
		fmt.Println()
		fmt.Println("💡 Your frontend should now display financial data!")
		fmt.Println("   1. Refresh your browser")
		fmt.Println("   2. Try generating reports")
		fmt.Println("   3. If still no data, check date ranges in the frontend")
	} else {
		fmt.Println("⚠️  Some issues remain:")
		if journalEntries == 0 {
			fmt.Println("   - No journal entries found")
		}
		if !viewExists {
			fmt.Println("   - Account balances view is missing")
		}
		if workingEndpoints < len(endpoints) {
			fmt.Println("   - Some API endpoints are not working")
		}
	}
}

func testAuthentication() (string, error) {
	loginData := map[string]string{
		"email":    "admin@company.com",
		"password": "password123",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post("http://localhost:8080/api/v1/auth/login", 
		"application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var loginResp struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(body, &loginResp)
	
	if loginResp.AccessToken == "" {
		return "", fmt.Errorf("no access token")
	}

	return loginResp.AccessToken, nil
}

func testAPIEndpoint(path, token string) (bool, error) {
	url := fmt.Sprintf("http://localhost:8080%s?start_date=2024-01-01&end_date=2025-12-31&format=json", path)
	
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	bodyStr := string(body)
	hasData := len(bodyStr) > 100 && // Response is substantial
		!(bodyStr == "null" || bodyStr == "[]" || bodyStr == "{}") // Not empty

	return hasData, nil
}