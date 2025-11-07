package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"app-sistem-akuntansi/config"
)

type TestResult struct {
	TestName    string
	Method      string
	URL         string
	StatusCode  int
	Expected    int
	Passed      bool
	Response    string
	Duration    time.Duration
	Error       string
}

type TestSuite struct {
	Name    string
	Results []TestResult
	Passed  int
	Failed  int
	Total   int
}

type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
	User    struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Role string `json:"role"`
	} `json:"user"`
}

type JournalEntry struct {
	ID           int                `json:"id"`
	EntryNumber  string            `json:"entry_number"`
	SourceType   string            `json:"source_type"`
	EntryDate    string            `json:"entry_date"`
	Description  string            `json:"description"`
	Reference    string            `json:"reference"`
	Notes        string            `json:"notes"`
	TotalDebit   float64           `json:"total_debit"`
	TotalCredit  float64           `json:"total_credit"`
	Status       string            `json:"status"`
	IsBalanced   bool              `json:"is_balanced"`
	JournalLines []JournalLine     `json:"journal_lines"`
}

type JournalLine struct {
	AccountID     int     `json:"account_id"`
	Description   string  `json:"description"`
	DebitAmount   float64 `json:"debit_amount"`
	CreditAmount  float64 `json:"credit_amount"`
}

type CreateJournalRequest struct {
	EntryDate   string        `json:"entry_date"`
	Description string        `json:"description"`
	Reference   string        `json:"reference"`
	Notes       string        `json:"notes"`
	Lines       []JournalLine `json:"lines"`
}

var (
	baseURL = "http://localhost:8080"
	authToken string
	testJournalID int
)

func main() {
	fmt.Println("🧪 Comprehensive SSOT API Testing Suite")
	fmt.Println("=======================================")
	fmt.Println("Testing all SSOT journal endpoints with real scenarios\n")

	// Load config
	_ = config.LoadConfig()

	// Start comprehensive testing
	suites := []TestSuite{
		runAuthenticationTests(),
		runJournalCRUDTests(),
		runJournalOperationTests(),
		runReportingTests(),
		runPerformanceTests(),
	}

	// Generate test report
	generateTestReport(suites)
	
	fmt.Println("\n🎯 Testing Complete!")
	printSummary(suites)
}

func runAuthenticationTests() TestSuite {
	fmt.Println("🔐 Running Authentication Tests...")
	suite := TestSuite{Name: "Authentication"}
	
	// Test login
	loginData := map[string]interface{}{
		"email":    "admin@example.com",
		"password": "admin123",
	}
	
	result := makeRequest("POST", "/api/v1/auth/login", loginData, "", 200)
	suite.Results = append(suite.Results, result)
	
	if result.Passed {
		// Extract token
		var loginResp LoginResponse
		if err := json.Unmarshal([]byte(result.Response), &loginResp); err == nil {
			authToken = loginResp.Token
			fmt.Printf("   ✅ Authentication successful, token acquired\n")
		}
	}
	
	// Test token validation
	result = makeRequest("GET", "/api/v1/auth/validate-token", nil, authToken, 200)
	suite.Results = append(suite.Results, result)
	
	suite.Total = len(suite.Results)
	for _, r := range suite.Results {
		if r.Passed {
			suite.Passed++
		} else {
			suite.Failed++
		}
	}
	
	return suite
}

func runJournalCRUDTests() TestSuite {
	fmt.Println("📝 Running Journal CRUD Tests...")
	suite := TestSuite{Name: "Journal CRUD Operations"}
	
	if authToken == "" {
		suite.Results = append(suite.Results, TestResult{
			TestName: "CRUD Tests Skipped",
			Error:    "No auth token available",
			Passed:   false,
		})
		suite.Failed = 1
		suite.Total = 1
		return suite
	}
	
	// Test 1: Create Journal Entry
	createData := CreateJournalRequest{
		EntryDate:   time.Now().Format("2006-01-02"),
		Description: "API Test Journal Entry",
		Reference:   "API-TEST-001",
		Notes:       "Created via API test suite",
		Lines: []JournalLine{
			{
				AccountID:     1,
				Description:   "Test Debit Entry",
				DebitAmount:   100000,
				CreditAmount:  0,
			},
			{
				AccountID:     2,
				Description:   "Test Credit Entry",
				DebitAmount:   0,
				CreditAmount:  100000,
			},
		},
	}
	
	result := makeRequest("POST", "/api/v1/journals", createData, authToken, 201)
	suite.Results = append(suite.Results, result)
	
	if result.Passed {
		// Extract journal ID for further tests
		var response map[string]interface{}
		if err := json.Unmarshal([]byte(result.Response), &response); err == nil {
			if data, ok := response["data"].(map[string]interface{}); ok {
				if id, ok := data["id"].(float64); ok {
					testJournalID = int(id)
					fmt.Printf("   ✅ Journal created with ID: %d\n", testJournalID)
				}
			}
		}
	}
	
	// Test 2: Get All Journals
	result = makeRequest("GET", "/api/v1/journals", nil, authToken, 200)
	suite.Results = append(suite.Results, result)
	
	// Test 3: Get Specific Journal
	if testJournalID > 0 {
		url := fmt.Sprintf("/api/v1/journals/%d", testJournalID)
		result = makeRequest("GET", url, nil, authToken, 200)
		suite.Results = append(suite.Results, result)
	}
	
	// Test 4: Update Journal Entry
	if testJournalID > 0 {
		updateData := CreateJournalRequest{
			EntryDate:   time.Now().Format("2006-01-02"),
			Description: "Updated API Test Journal Entry",
			Reference:   "API-TEST-001-UPDATED",
			Notes:       "Updated via API test suite",
			Lines: []JournalLine{
				{
					AccountID:     1,
					Description:   "Updated Test Debit Entry",
					DebitAmount:   150000,
					CreditAmount:  0,
				},
				{
					AccountID:     2,
					Description:   "Updated Test Credit Entry",
					DebitAmount:   0,
					CreditAmount:  150000,
				},
			},
		}
		
		url := fmt.Sprintf("/api/v1/journals/%d", testJournalID)
		result = makeRequest("PUT", url, updateData, authToken, 200)
		suite.Results = append(suite.Results, result)
	}
	
	suite.Total = len(suite.Results)
	for _, r := range suite.Results {
		if r.Passed {
			suite.Passed++
		} else {
			suite.Failed++
		}
	}
	
	return suite
}

func runJournalOperationTests() TestSuite {
	fmt.Println("⚡ Running Journal Operation Tests...")
	suite := TestSuite{Name: "Journal Operations"}
	
	if authToken == "" || testJournalID == 0 {
		suite.Results = append(suite.Results, TestResult{
			TestName: "Operation Tests Skipped",
			Error:    "No auth token or journal ID available",
			Passed:   false,
		})
		suite.Failed = 1
		suite.Total = 1
		return suite
	}
	
	// Test 1: Post Journal Entry
	url := fmt.Sprintf("/api/v1/journals/%d/post", testJournalID)
	result := makeRequest("POST", url, nil, authToken, 200)
	suite.Results = append(suite.Results, result)
	
	// Test 2: Get Journal Summary
	result = makeRequest("GET", "/api/v1/journals/summary", nil, authToken, 200)
	suite.Results = append(suite.Results, result)
	
	// Test 3: Get Account Balances
	result = makeRequest("GET", "/api/v1/journals/account-balances", nil, authToken, 200)
	suite.Results = append(suite.Results, result)
	
	// Test 4: Refresh Account Balances
	result = makeRequest("POST", "/api/v1/journals/account-balances/refresh", nil, authToken, 200)
	suite.Results = append(suite.Results, result)
	
	// Test 5: Reverse Journal Entry (if posted successfully)
	if suite.Results[0].Passed {
		url := fmt.Sprintf("/api/v1/journals/%d/reverse", testJournalID)
		reverseData := map[string]string{
			"reason": "API Test Reversal",
		}
		result = makeRequest("POST", url, reverseData, authToken, 200)
		suite.Results = append(suite.Results, result)
	}
	
	suite.Total = len(suite.Results)
	for _, r := range suite.Results {
		if r.Passed {
			suite.Passed++
		} else {
			suite.Failed++
		}
	}
	
	return suite
}

func runReportingTests() TestSuite {
	fmt.Println("📊 Running Reporting Tests...")
	suite := TestSuite{Name: "Reporting & Analytics"}
	
	if authToken == "" {
		suite.Results = append(suite.Results, TestResult{
			TestName: "Reporting Tests Skipped",
			Error:    "No auth token available",
			Passed:   false,
		})
		suite.Failed = 1
		suite.Total = 1
		return suite
	}
	
	// Test 1: Journal Summary with Date Range
	startDate := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")
	
	params := fmt.Sprintf("?start_date=%s&end_date=%s", startDate, endDate)
	result := makeRequest("GET", "/api/v1/journals/summary"+params, nil, authToken, 200)
	suite.Results = append(suite.Results, result)
	
	// Test 2: Filtered Journal Entries
	params = fmt.Sprintf("?start_date=%s&end_date=%s&status=POSTED&limit=10", startDate, endDate)
	result = makeRequest("GET", "/api/v1/journals"+params, nil, authToken, 200)
	suite.Results = append(suite.Results, result)
	
	// Test 3: Search Journal Entries
	result = makeRequest("GET", "/api/v1/journals?search=API+Test&limit=10", nil, authToken, 200)
	suite.Results = append(suite.Results, result)
	
	suite.Total = len(suite.Results)
	for _, r := range suite.Results {
		if r.Passed {
			suite.Passed++
		} else {
			suite.Failed++
		}
	}
	
	return suite
}

func runPerformanceTests() TestSuite {
	fmt.Println("🚀 Running Performance Tests...")
	suite := TestSuite{Name: "Performance"}
	
	if authToken == "" {
		suite.Results = append(suite.Results, TestResult{
			TestName: "Performance Tests Skipped",
			Error:    "No auth token available",
			Passed:   false,
		})
		suite.Failed = 1
		suite.Total = 1
		return suite
	}
	
	// Test 1: Large Data Retrieval (stress test)
	start := time.Now()
	result := makeRequest("GET", "/api/v1/journals?limit=1000", nil, authToken, 200)
	result.TestName = "Large Data Retrieval (1000 records)"
	result.Duration = time.Since(start)
	
	// Performance threshold check (should complete within 5 seconds)
	if result.Duration > 5*time.Second {
		result.Error = fmt.Sprintf("Performance issue: took %v (>5s)", result.Duration)
		result.Passed = false
	}
	
	suite.Results = append(suite.Results, result)
	
	// Test 2: Concurrent Requests Simulation
	start = time.Now()
	concurrentResults := make(chan TestResult, 5)
	
	for i := 0; i < 5; i++ {
		go func(id int) {
			result := makeRequest("GET", "/api/v1/journals/summary", nil, authToken, 200)
			result.TestName = fmt.Sprintf("Concurrent Request %d", id+1)
			concurrentResults <- result
		}(i)
	}
	
	// Collect results
	for i := 0; i < 5; i++ {
		result := <-concurrentResults
		suite.Results = append(suite.Results, result)
	}
	
	totalConcurrentTime := time.Since(start)
	fmt.Printf("   ⏱️  5 concurrent requests completed in %v\n", totalConcurrentTime)
	
	suite.Total = len(suite.Results)
	for _, r := range suite.Results {
		if r.Passed {
			suite.Passed++
		} else {
			suite.Failed++
		}
	}
	
	return suite
}

func makeRequest(method, endpoint string, data interface{}, token string, expectedStatus int) TestResult {
	startTime := time.Now()
	
	result := TestResult{
		TestName:   fmt.Sprintf("%s %s", method, endpoint),
		Method:     method,
		URL:        baseURL + endpoint,
		Expected:   expectedStatus,
		Duration:   0,
	}
	
	var reqBody []byte
	var err error
	
	if data != nil {
		reqBody, err = json.Marshal(data)
		if err != nil {
			result.Error = fmt.Sprintf("JSON marshal error: %v", err)
			result.Passed = false
			return result
		}
	}
	
	req, err := http.NewRequest(method, result.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		result.Error = fmt.Sprintf("Request creation error: %v", err)
		result.Passed = false
		return result
	}
	
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Request error: %v", err)
		result.Passed = false
		return result
	}
	defer resp.Body.Close()
	
	result.Duration = time.Since(startTime)
	result.StatusCode = resp.StatusCode
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("Response read error: %v", err)
		result.Passed = false
		return result
	}
	
	result.Response = string(body)
	result.Passed = resp.StatusCode == expectedStatus
	
	if !result.Passed {
		result.Error = fmt.Sprintf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
	
	// Log result
	status := "✅ PASS"
	if !result.Passed {
		status = "❌ FAIL"
	}
	fmt.Printf("   %s %s (%v)\n", status, result.TestName, result.Duration)
	
	return result
}

func generateTestReport(suites []TestSuite) {
	report := fmt.Sprintf(`# SSOT API Test Report
Generated: %s

## Executive Summary
`, time.Now().Format("2006-01-02 15:04:05"))

	totalPassed := 0
	totalFailed := 0
	totalTests := 0

	for _, suite := range suites {
		totalPassed += suite.Passed
		totalFailed += suite.Failed
		totalTests += suite.Total
	}

	report += fmt.Sprintf(`
- **Total Tests:** %d
- **Passed:** %d (%.1f%%)  
- **Failed:** %d (%.1f%%)
- **Success Rate:** %.1f%%

`, totalTests, totalPassed, 
		float64(totalPassed)/float64(totalTests)*100,
		totalFailed,
		float64(totalFailed)/float64(totalTests)*100,
		float64(totalPassed)/float64(totalTests)*100)

	report += "## Test Suite Results\n\n"

	for _, suite := range suites {
		report += fmt.Sprintf("### %s\n", suite.Name)
		report += fmt.Sprintf("- **Total:** %d tests\n", suite.Total)
		report += fmt.Sprintf("- **Passed:** %d\n", suite.Passed)
		report += fmt.Sprintf("- **Failed:** %d\n", suite.Failed)
		report += fmt.Sprintf("- **Success Rate:** %.1f%%\n\n", float64(suite.Passed)/float64(suite.Total)*100)

		if suite.Failed > 0 {
			report += "**Failed Tests:**\n"
			for _, result := range suite.Results {
				if !result.Passed {
					report += fmt.Sprintf("- %s: %s\n", result.TestName, result.Error)
				}
			}
			report += "\n"
		}
	}

	report += "## Performance Metrics\n\n"
	
	for _, suite := range suites {
		if suite.Name == "Performance" {
			for _, result := range suite.Results {
				if result.Duration > 0 {
					report += fmt.Sprintf("- **%s:** %v\n", result.TestName, result.Duration)
				}
			}
		}
	}

	// Write report to file
	filename := "SSOT_API_TEST_REPORT.md"
	if err := ioutil.WriteFile(filename, []byte(report), 0644); err != nil {
		log.Printf("Failed to write test report: %v", err)
	} else {
		fmt.Printf("📄 Test report saved: %s\n", filename)
	}
}

func printSummary(suites []TestSuite) {
	fmt.Println("============================================")
	fmt.Println("📊 Final Test Summary")
	fmt.Println("============================================")

	totalPassed := 0
	totalFailed := 0
	totalTests := 0

	for _, suite := range suites {
		status := "✅"
		if suite.Failed > 0 {
			status = "❌"
		}
		
		fmt.Printf("%s %-25s %d/%d passed (%.1f%%)\n", 
			status, suite.Name+":", suite.Passed, suite.Total, 
			float64(suite.Passed)/float64(suite.Total)*100)

		totalPassed += suite.Passed
		totalFailed += suite.Failed
		totalTests += suite.Total
	}

	fmt.Println("============================================")
	fmt.Printf("🎯 **OVERALL: %d/%d tests passed (%.1f%%)**\n", 
		totalPassed, totalTests, float64(totalPassed)/float64(totalTests)*100)

	if totalFailed == 0 {
		fmt.Println("🎉 All tests passed! SSOT API is ready for production.")
	} else {
		fmt.Printf("⚠️  %d tests failed. Please review the issues above.\n", totalFailed)
	}

	// Clean up test data
	if testJournalID > 0 && authToken != "" {
		fmt.Printf("\n🧹 Cleaning up test data (Journal ID: %d)...\n", testJournalID)
		url := fmt.Sprintf("/api/v1/journals/%d", testJournalID)
		result := makeRequest("DELETE", url, nil, authToken, 200)
		if result.Passed {
			fmt.Println("✅ Test data cleaned up successfully")
		}
	}
}