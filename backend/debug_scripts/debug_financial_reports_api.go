package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	fmt.Println("🔍 Financial Reports API Detailed Debug")
	fmt.Println("======================================")
	fmt.Println()

	// Get authentication token
	fmt.Println("🔐 Getting authentication token...")
	token, err := getAuthToken()
	if err != nil {
		fmt.Printf("❌ Authentication failed: %v\n", err)
		return
	}
	fmt.Println("✅ Authentication successful")
	fmt.Println()

	// Test each report endpoint with detailed analysis
	endpoints := []struct {
		name string
		path string
	}{
		{"Journal Entry Analysis", "/api/v1/reports/journal-entry-analysis"},
		{"Profit & Loss Statement", "/api/v1/reports/profit-loss"},
		{"Balance Sheet", "/api/v1/reports/balance-sheet"},
		{"Sales Summary Report", "/api/v1/reports/sales-summary"},
		{"Trial Balance", "/api/v1/reports/trial-balance"},
	}

	// Test different date ranges that should contain data
	dateRanges := []struct {
		name      string
		startDate string
		endDate   string
	}{
		{"Current Period", "2025-09-01", "2025-09-30"},
		{"Extended Range", "2025-01-01", "2025-12-31"},
		{"Wide Range", "2024-01-01", "2026-12-31"},
	}

	for _, endpoint := range endpoints {
		fmt.Printf("📊 Testing %s\n", endpoint.name)
		fmt.Printf("=" + strings.Repeat("=", len(endpoint.name)+10) + "\n")

		for _, dateRange := range dateRanges {
			fmt.Printf("\n📅 Date Range: %s (%s to %s)\n", 
				dateRange.name, dateRange.startDate, dateRange.endDate)
			
			response, err := callAPI(endpoint.path, dateRange.startDate, dateRange.endDate, token)
			if err != nil {
				fmt.Printf("   ❌ Error: %v\n", err)
				continue
			}

			analyzeResponse(response, endpoint.name, dateRange.name)
		}
		fmt.Println()
	}

	// Also check specific Journal Entry Analysis issue
	fmt.Println("🎯 Specific Journal Entry Analysis Debug")
	fmt.Println("=======================================")
	debugJournalEntryAnalysis(token)
}

func getAuthToken() (string, error) {
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
		return "", fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	var loginResp struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(body, &loginResp)
	return loginResp.AccessToken, nil
}

func callAPI(path, startDate, endDate, token string) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://localhost:8080%s?start_date=%s&end_date=%s&format=json", 
		path, startDate, endDate)
	
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("JSON parse error: %v", err)
	}

	return response, nil
}

func analyzeResponse(response map[string]interface{}, endpointName, dateRangeName string) {
	// Pretty print the response for analysis
	responseBytes, _ := json.MarshalIndent(response, "   ", "  ")
	responseStr := string(responseBytes)
	
	fmt.Printf("   📄 Response size: %d bytes\n", len(responseStr))
	
	// Check for common "no data" indicators
	noDataIndicators := []string{
		"No journal entries found",
		"No P&L relevant transactions found",
		"No data available",
		"Data received: 0 entries",
		`"total_entries":0`,
		`"total_amount":0`,
		`"items":[]`,
		`"items":null`,
	}

	hasNoDataIndicator := false
	for _, indicator := range noDataIndicators {
		if strings.Contains(responseStr, indicator) {
			fmt.Printf("   ⚠️  Found 'no data' indicator: %s\n", indicator)
			hasNoDataIndicator = true
		}
	}

	// Check for positive data indicators
	dataIndicators := []string{
		`"total":`,
		`"amount":`,
		`"balance":`,
		`"revenue":`,
		`"expenses":`,
		`"entries":`,
		`"items":[`,
		`"data":[`,
	}

	hasDataIndicator := false
	for _, indicator := range dataIndicators {
		if strings.Contains(responseStr, indicator) && 
		   !strings.Contains(responseStr, indicator+"0") &&
		   !strings.Contains(responseStr, indicator+"null") &&
		   !strings.Contains(responseStr, indicator+"[]") {
			fmt.Printf("   ✅ Found data indicator: %s\n", indicator)
			hasDataIndicator = true
		}
	}

	// Show key fields from response
	if data, ok := response["data"]; ok && data != nil {
		fmt.Printf("   📊 Response has 'data' field\n")
		if dataMap, ok := data.(map[string]interface{}); ok {
			for key, value := range dataMap {
				fmt.Printf("      - %s: %v\n", key, value)
			}
		}
	}

	// Check for specific fields that frontend might be looking for
	importantFields := []string{"total_entries", "total_amount", "items", "entries", "data", "message", "status"}
	for _, field := range importantFields {
		if value, exists := response[field]; exists {
			fmt.Printf("   📋 %s: %v\n", field, value)
		}
	}

	// Overall assessment
	if hasNoDataIndicator {
		fmt.Printf("   🔍 Assessment: API correctly reports NO DATA for this date range\n")
	} else if hasDataIndicator {
		fmt.Printf("   🔍 Assessment: API contains DATA but frontend may not be parsing correctly\n")
	} else {
		fmt.Printf("   🔍 Assessment: Unclear - need to check response format\n")
	}

	// Show first 500 characters of response for debugging
	if len(responseStr) > 0 {
		preview := responseStr
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		fmt.Printf("   📝 Response preview:\n%s\n", preview)
	}
}

func debugJournalEntryAnalysis(token string) {
	fmt.Println("\n🔎 Detailed Journal Entry Analysis Debug...")
	
	// Try different date formats
	dateFormats := []struct {
		name   string
		start  string
		end    string
	}{
		{"DD/MM/YYYY format", "19/09/2025", "20/09/2025"},
		{"YYYY-MM-DD format", "2025-09-19", "2025-09-20"},
		{"Extended range", "2025-09-01", "2025-09-30"},
		{"Wide range", "2025-01-01", "2025-12-31"},
	}

	for _, dateFormat := range dateFormats {
		fmt.Printf("\n📅 Testing with %s: %s to %s\n", 
			dateFormat.name, dateFormat.start, dateFormat.end)
			
		url := fmt.Sprintf("http://localhost:8080/api/v1/reports/journal-entry-analysis?start_date=%s&end_date=%s&format=json", 
			dateFormat.start, dateFormat.end)
		
		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("   ❌ Request failed: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		
		fmt.Printf("   📊 Status: %d, Size: %d bytes\n", resp.StatusCode, len(body))
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var response map[string]interface{}
			if json.Unmarshal(body, &response) == nil {
				// Look for the specific message shown in frontend
				if message, exists := response["message"]; exists {
					fmt.Printf("   💬 Message: %v\n", message)
				}
				
				// Check if there's actual data
				if data, exists := response["data"]; exists && data != nil {
					fmt.Printf("   📈 Has data field: %v\n", data)
				}

				if entries, exists := response["entries"]; exists {
					fmt.Printf("   📝 Entries field: %v\n", entries)
				}
			}
		} else {
			fmt.Printf("   ❌ Error response: %s\n", string(body))
		}
	}

	// Also check what's actually in the database for journal entries in that date range
	fmt.Println("\n🗄️  Direct database check would show:")
	fmt.Println("   - Journal entries from 2025-09-18 to 2025-09-19")
	fmt.Println("   - 5 total entries with descriptions like:")
	fmt.Println("     * Sale Invoice #INV/2025/09/0002")
	fmt.Println("     * Purchase Order #PO/2025/09/0001") 
	fmt.Println("     * Payment #test")
	fmt.Println()
	fmt.Println("💡 Recommendation:")
	fmt.Println("   1. Check if date format parsing is correct in backend")
	fmt.Println("   2. Verify the journal-entry-analysis endpoint logic")
	fmt.Println("   3. Check if frontend is using correct date format")
	fmt.Println("   4. Ensure timezone handling is consistent")
}