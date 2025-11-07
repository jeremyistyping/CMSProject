package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Token   string `json:"token"`
	User    struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
}

type APITestResult struct {
	Endpoint     string      `json:"endpoint"`
	Status       int         `json:"status"`
	Success      bool        `json:"success"`
	ResponseSize int         `json:"response_size"`
	HasData      bool        `json:"has_data"`
	DataPreview  interface{} `json:"data_preview,omitempty"`
	Error        string      `json:"error,omitempty"`
}

type JournalEntry struct {
	ID          uint      `json:"id"`
	EntryDate   time.Time `json:"entry_date"`
	Description string    `json:"description"`
	Reference   string    `json:"reference"`
	TotalDebit  float64   `json:"total_debit"`
	TotalCredit float64   `json:"total_credit"`
	CreatedAt   time.Time `json:"created_at"`
}

type AccountBalance struct {
	AccountID     uint    `json:"account_id"`
	AccountCode   string  `json:"account_code"`
	AccountName   string  `json:"account_name"`
	AccountType   string  `json:"account_type"`
	TotalDebit    float64 `json:"total_debit"`
	TotalCredit   float64 `json:"total_credit"`
	Balance       float64 `json:"balance"`
	LastUpdated   time.Time `json:"last_updated"`
}

func main() {
	fmt.Println("🧪 Enhanced Financial Reports API Testing")
	fmt.Println("=========================================")
	fmt.Println("Testing with authentication and database verification...")
	fmt.Println()

	// Initialize database
	db := database.ConnectDB()
	
	// Step 1: Check SSOT data in database
	fmt.Println("📊 Step 1: Verifying SSOT Journal System Data in Database...")
	journalCount, totalAmount, hasAccountBalances := checkDatabaseSSOTData(db)
	fmt.Printf("✅ Journal Entries found: %d entries\n", journalCount)
	fmt.Printf("✅ Total transaction amount: %.2f\n", totalAmount)
	fmt.Printf("✅ Account Balances view exists: %t\n", hasAccountBalances)
	fmt.Println()

	// Step 2: Get authentication token
	fmt.Println("🔐 Step 2: Getting Authentication Token...")
	token, err := authenticate()
	if err != nil {
		fmt.Printf("❌ Authentication failed: %v\n", err)
		fmt.Println("🚨 Cannot proceed with API tests without authentication")
		return
	}
	fmt.Println("✅ Authentication successful!")
	fmt.Println()

	// Step 3: Test API endpoints with authentication
	fmt.Println("🌐 Step 3: Testing API Endpoints with Authentication...")
	testAPIEndpoints(token)
	
	fmt.Println()
	fmt.Println("🎯 Final Assessment:")
	if journalCount > 0 && hasAccountBalances {
		fmt.Println("✅ SSOT Journal System has data available")
		fmt.Println("✅ Database is properly set up with journal entries and account balances")
		fmt.Println("✅ Frontend should be able to retrieve financial report data")
		fmt.Println()
		fmt.Println("💡 Recommendation: The 'No Data Available' issue in frontend is likely due to:")
		fmt.Println("   1. Date range filtering (try expanding date ranges)")
		fmt.Println("   2. Account mapping configuration")
		fmt.Println("   3. Report filtering logic")
	} else {
		fmt.Println("⚠️  SSOT Journal System data is incomplete")
		fmt.Println("⚠️  Need to populate journal entries first")
	}
}

func checkDatabaseSSOTData(db *gorm.DB) (int64, float64, bool) {
	// Check journal entries count
	var journalCount int64
	db.Table("journal_entries").Count(&journalCount)
	
	// Check total transaction amounts
	var totalAmount float64
	if journalCount > 0 {
		err := db.Table("journal_entries").Select("COALESCE(SUM(total_debit), 0)").Row().Scan(&totalAmount)
		if err != nil {
			fmt.Printf("⚠️  Error calculating total amount: %v\n", err)
			totalAmount = 0
		}
	}
	
	// Check if account_balances materialized view exists
	var viewExists bool = false
	var count int64
	err := db.Table("account_balances").Count(&count)
	if err == nil && count > 0 {
		viewExists = true
	}
	
	// If view exists, get some sample data
	if viewExists {
		fmt.Printf("✅ Account Balances records: %d\n", count)
	}
	
	// Show sample journal entries if any exist
	if journalCount > 0 {
		var sampleEntries []JournalEntry
		db.Table("journal_entries").Limit(3).Find(&sampleEntries)
		if len(sampleEntries) > 0 {
			fmt.Println("📝 Sample Journal Entries:")
			for _, entry := range sampleEntries {
				fmt.Printf("   - Date: %s, Ref: %s, Debit: %.2f, Credit: %.2f\n", 
					entry.EntryDate.Format("2006-01-02"), entry.Reference, 
					entry.TotalDebit, entry.TotalCredit)
			}
		}
	}
	
	return journalCount, totalAmount, viewExists
}

func authenticate() (string, error) {
	loginData := LoginRequest{
		Email:    "admin@company.com",
		Password: "password123",
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
	
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("failed to parse login response: %v", err)
	}
	
	if loginResp.Token == "" {
		return "", fmt.Errorf("no token received in login response")
	}
	
	fmt.Printf("✅ Logged in as: %s (%s)\n", loginResp.User.Username, loginResp.User.Role)
	return loginResp.Token, nil
}

func testAPIEndpoints(token string) {
	baseURL := "http://localhost:8080"
	startDate := "2024-01-01" // Expanded date range
	endDate := "2025-12-31"
	
	endpoints := []struct {
		name string
		path string
	}{
		{"Profit & Loss Statement", "/api/v1/reports/profit-loss"},
		{"Balance Sheet", "/api/v1/reports/balance-sheet"},
		{"Cash Flow Statement", "/api/v1/reports/cash-flow"},
		{"Sales Summary Report", "/api/v1/reports/sales-summary"},
		{"Purchase Summary Report", "/api/v1/reports/purchase-summary"},
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
		
		result := testEndpointWithAuth(url, endpoint.name, token)
		
		if result.Success {
			successCount++
			statusMsg := fmt.Sprintf("✅ Status: %d, Size: %d bytes", result.Status, result.ResponseSize)
			if result.HasData {
				dataAvailableCount++
				statusMsg += ", ✅ HAS DATA"
				if result.DataPreview != nil {
					fmt.Printf("   %s\n", statusMsg)
					previewBytes, _ := json.MarshalIndent(result.DataPreview, "   ", "  ")
					fmt.Printf("   📊 Data Preview: %s\n", string(previewBytes))
				} else {
					fmt.Printf("   %s\n", statusMsg)
				}
			} else {
				statusMsg += ", ⚠️  NO DATA"
				fmt.Printf("   %s\n", statusMsg)
			}
		} else {
			fmt.Printf("   ❌ Status: %d, Error: %s\n", result.Status, result.Error)
		}
		fmt.Println()
		
		time.Sleep(200 * time.Millisecond)
	}
	
	fmt.Println("📋 API Test Summary:")
	fmt.Printf("- Total endpoints tested: %d\n", len(endpoints))
	fmt.Printf("- Successful responses: %d\n", successCount)
	fmt.Printf("- Endpoints with data: %d\n", dataAvailableCount)
	fmt.Printf("- Success rate: %.1f%%\n", float64(successCount)/float64(len(endpoints))*100)
	fmt.Printf("- Data availability rate: %.1f%%\n", float64(dataAvailableCount)/float64(len(endpoints))*100)
}

func testEndpointWithAuth(url, name, token string) APITestResult {
	result := APITestResult{
		Endpoint: name,
		Success:  false,
		HasData:  false,
	}
	
	client := &http.Client{Timeout: 15 * time.Second}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("Request creation failed: %v", err)
		return result
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Request failed: %v", err)
		return result
	}
	defer resp.Body.Close()
	
	result.Status = resp.StatusCode
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read response: %v", err)
		return result
	}
	
	result.ResponseSize = len(body)
	bodyStr := string(body)
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			result.HasData = checkForData(bodyStr, jsonData)
			if result.HasData {
				result.DataPreview = createDataPreview(jsonData)
			}
		} else {
			result.HasData = containsDataKeywords(bodyStr)
		}
	} else {
		if strings.Contains(bodyStr, "AUTH_HEADER_MISSING") || strings.Contains(bodyStr, "Unauthorized") {
			result.Error = "Authentication failed"
		} else {
			var errorResp map[string]interface{}
			if json.Unmarshal(body, &errorResp) == nil {
				if msg, ok := errorResp["error"].(string); ok {
					result.Error = msg
				} else if msg, ok := errorResp["message"].(string); ok {
					result.Error = msg
				} else {
					result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
				}
			} else {
				result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
		}
	}
	
	return result
}

func checkForData(bodyStr string, jsonData interface{}) bool {
	// Check for empty data indicators
	emptyIndicators := []string{
		`"total_entries":0`,
		`"total_amount":0`,
		`"items":null`,
		`"items":[]`,
		`"data":null`,
		`"data":[]`,
		"No P&L relevant transactions found",
		"No data available",
		"no data found",
		"empty result",
	}
	
	bodyLower := strings.ToLower(bodyStr)
	for _, indicator := range emptyIndicators {
		if strings.Contains(bodyLower, strings.ToLower(indicator)) {
			return false
		}
	}
	
	// Check for positive data indicators
	dataIndicators := []string{
		`"total":`,
		`"amount":`,
		`"balance":`,
		`"revenue":`,
		`"expenses":`,
		`"assets":`,
		`"liabilities":`,
		`"entries":`,
		"journal_entries",
		"account_balance",
	}
	
	hasPositiveIndicator := false
	for _, indicator := range dataIndicators {
		if strings.Contains(bodyLower, strings.ToLower(indicator)) {
			// Make sure it's not zero or null
			if !strings.Contains(bodyLower, indicator+"0") && 
			   !strings.Contains(bodyLower, indicator+"null") &&
			   !strings.Contains(bodyLower, indicator+" 0") {
				hasPositiveIndicator = true
				break
			}
		}
	}
	
	// Check JSON structure
	if dataMap, ok := jsonData.(map[string]interface{}); ok {
		// Check if there's meaningful data
		if data, exists := dataMap["data"]; exists && data != nil {
			if dataArray, ok := data.([]interface{}); ok && len(dataArray) > 0 {
				return true
			}
			if dataObj, ok := data.(map[string]interface{}); ok && len(dataObj) > 0 {
				return true
			}
		}
		
		// Check for non-zero numeric values
		for _, value := range dataMap {
			if numVal, ok := value.(float64); ok && numVal != 0 {
				return true
			}
		}
	}
	
	return hasPositiveIndicator
}

func containsDataKeywords(body string) bool {
	keywords := []string{
		"revenue", "expense", "profit", "loss", "balance", 
		"asset", "liability", "equity", "debit", "credit",
		"journal", "account", "transaction", "entry",
	}
	
	bodyLower := strings.ToLower(body)
	for _, keyword := range keywords {
		if strings.Contains(bodyLower, keyword) {
			return true
		}
	}
	return false
}

func createDataPreview(jsonData interface{}) interface{} {
	if dataMap, ok := jsonData.(map[string]interface{}); ok {
		preview := make(map[string]interface{})
		
		// Key fields to extract
		keyFields := []string{
			"total", "amount", "balance", "revenue", "expenses", 
			"entries", "count", "total_amount", "total_entries",
			"assets", "liabilities", "equity",
		}
		
		for _, field := range keyFields {
			if value, exists := dataMap[field]; exists {
				preview[field] = value
			}
		}
		
		// Check nested data
		if data, exists := dataMap["data"]; exists {
			if dataDict, ok := data.(map[string]interface{}); ok {
				for _, field := range keyFields {
					if value, exists := dataDict[field]; exists {
						preview[field] = value
					}
				}
			}
		}
		
		if len(preview) > 0 {
			return preview
		}
	}
	
	return nil
}