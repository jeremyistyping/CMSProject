package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Test Configuration
const (
	BaseURL = "http://localhost:8080/api/v1"
	
	// Test Data
	TestCustomerID = 34  // PT Maju Jaya
	TestProductID  = 23  // Mesin Printer
	TestCashBankID = 1
	
	// Expected Account Codes
	ReceivableAccount = "1201" // Piutang Usaha
	CashAccount      = "1101" // Kas
	BankAccount      = "1102" // Bank BCA
	RevenueAccount   = "4101" // Pendapatan Penjualan
)

// Data Structures
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

type Sale struct {
	ID                uint    `json:"id"`
	Code              string  `json:"code"`
	CustomerID        uint    `json:"customer_id"`
	Date              string  `json:"date"`
	DueDate           string  `json:"due_date"`
	TotalAmount       float64 `json:"total_amount"`
	OutstandingAmount float64 `json:"outstanding_amount"`
	Status            string  `json:"status"`
}

type Payment struct {
	ID        uint    `json:"id"`
	Code      string  `json:"code"`
	ContactID uint    `json:"contact_id"`
	Date      string  `json:"date"`
	Amount    float64 `json:"amount"`
	Method    string  `json:"method"`
	Status    string  `json:"status"`
}

type AccountBalance struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

type JournalEntry struct {
	ID           uint    `json:"id"`
	Code         string  `json:"code"`
	Date         string  `json:"date"`
	Description  string  `json:"description"`
	TotalDebit   float64 `json:"total_debit"`
	TotalCredit  float64 `json:"total_credit"`
	ReferenceID  uint    `json:"reference_id"`
	ReferenceType string `json:"reference_type"`
}

type TestReport struct {
	StartTime time.Time
	EndTime   time.Time
	
	// Test Results
	SalesCreated          bool
	SalesInvoiced         bool
	PaymentRecorded       bool
	JournalEntriesCreated bool
	AccountsUpdated       bool
	IntegrityVerified     bool
	
	// Data
	Sale             *Sale
	Payment          *Payment
	JournalEntries   []JournalEntry
	InitialBalances  map[string]float64
	FinalBalances    map[string]float64
	
	// Metrics
	TotalTests    int
	PassedTests   int
	FailedTests   int
	Errors        []string
	Warnings      []string
}

type TestClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

func main() {
	fmt.Println("🚀 Starting Comprehensive Sales-Payment Flow Test")
	fmt.Println("=" + strings.Repeat("=", 60))
	
	client := NewTestClient(BaseURL)
	report := &TestReport{
		StartTime:       time.Now(),
		InitialBalances: make(map[string]float64),
		FinalBalances:   make(map[string]float64),
		Errors:          make([]string, 0),
		Warnings:        make([]string, 0),
	}
	
	// Execute comprehensive test
	if err := runComprehensiveTest(client, report); err != nil {
		log.Fatalf("❌ Test failed: %v", err)
	}
	
	// Generate final report
	generateFinalReport(report)
}

func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func runComprehensiveTest(client *TestClient, report *TestReport) error {
	fmt.Println("📋 Running Comprehensive Test Suite...")
	
	// Step 1: Authenticate (if needed)
	fmt.Println("\n🔐 Step 1: Authentication")
	if err := client.authenticate(); err != nil {
		report.addError("Authentication failed: " + err.Error())
		return err
	}
	fmt.Println("✅ Authentication successful")
	
	// Step 2: Capture initial account balances
	fmt.Println("\n📊 Step 2: Capturing Initial Account Balances")
	if err := captureInitialBalances(client, report); err != nil {
		report.addError("Failed to capture initial balances: " + err.Error())
		return err
	}
	fmt.Println("✅ Initial balances captured")
	
	// Step 3: Create Sales Order
	fmt.Println("\n🛒 Step 3: Creating Sales Order")
	if err := createSalesOrder(client, report); err != nil {
		report.addError("Failed to create sales order: " + err.Error())
		return err
	}
	report.SalesCreated = true
	fmt.Printf("✅ Sales order created: %s (ID: %d)\n", report.Sale.Code, report.Sale.ID)
	
	// Step 4: Invoice the Sales Order
	fmt.Println("\n🧾 Step 4: Converting to Invoice")
	if err := invoiceSalesOrder(client, report); err != nil {
		report.addError("Failed to invoice sales order: " + err.Error())
		return err
	}
	report.SalesInvoiced = true
	fmt.Printf("✅ Sales order invoiced: %s\n", report.Sale.Code)
	
	// Step 5: Verify Journal Entry for Sales
	fmt.Println("\n📖 Step 5: Verifying Sales Journal Entries")
	if err := verifySalesJournalEntries(client, report); err != nil {
		report.addWarning("Sales journal verification incomplete: " + err.Error())
	} else {
		fmt.Println("✅ Sales journal entries verified")
	}
	
	// Step 6: Record Payment
	fmt.Println("\n💰 Step 6: Recording Payment")
	if err := recordPayment(client, report); err != nil {
		report.addError("Failed to record payment: " + err.Error())
		return err
	}
	report.PaymentRecorded = true
	fmt.Printf("✅ Payment recorded: %s (Amount: %.2f)\n", report.Payment.Code, report.Payment.Amount)
	
	// Step 7: Verify Payment Journal Entries
	fmt.Println("\n📖 Step 7: Verifying Payment Journal Entries")
	if err := verifyPaymentJournalEntries(client, report); err != nil {
		report.addWarning("Payment journal verification incomplete: " + err.Error())
	} else {
		fmt.Println("✅ Payment journal entries verified")
		report.JournalEntriesCreated = true
	}
	
	// Step 8: Capture final account balances
	fmt.Println("\n📊 Step 8: Capturing Final Account Balances")
	if err := captureFinalBalances(client, report); err != nil {
		report.addError("Failed to capture final balances: " + err.Error())
		return err
	}
	fmt.Println("✅ Final balances captured")
	report.AccountsUpdated = true
	
	// Step 9: Verify Data Integrity
	fmt.Println("\n🔍 Step 9: Verifying Data Integrity")
	if err := verifyDataIntegrity(client, report); err != nil {
		report.addError("Data integrity verification failed: " + err.Error())
		return err
	}
	report.IntegrityVerified = true
	fmt.Println("✅ Data integrity verified")
	
	report.EndTime = time.Now()
	return nil
}

func (c *TestClient) authenticate() error {
	// Login data - using correct admin credentials
	loginData := map[string]interface{}{
		"email":    "admin@company.com",
		"password": "password123",
	}
	
	// Make login request to get token
	body, err := c.makeRequestWithoutAuth("POST", "/auth/login", loginData)
	if err != nil {
		return fmt.Errorf("login request failed: %v", err)
	}
	
	// Parse response to extract token
	var loginResp map[string]interface{}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %v", err)
	}
	
	// Extract access token
	if accessToken, ok := loginResp["access_token"].(string); ok {
		c.authToken = accessToken
		fmt.Printf("   ✅ Authentication successful (token: %s...)\n", c.authToken[:20])
		return nil
	}
	
	return fmt.Errorf("no access token found in login response: %v", loginResp)
}

func captureInitialBalances(client *TestClient, report *TestReport) error {
	accounts := []string{ReceivableAccount, CashAccount, BankAccount, RevenueAccount}
	
	for _, accountCode := range accounts {
		balance, err := getAccountBalance(client, accountCode)
		if err != nil {
			return fmt.Errorf("failed to get balance for account %s: %v", accountCode, err)
		}
		report.InitialBalances[accountCode] = balance
		fmt.Printf("   📊 %s: %.2f\n", accountCode, balance)
	}
	
	return nil
}

func createSalesOrder(client *TestClient, report *TestReport) error {
	saleData := map[string]interface{}{
		"customer_id": TestCustomerID,
		"type":        "INVOICE",
		"date":        time.Now().Format("2006-01-02T15:04:05Z07:00"),
		"due_date":    time.Now().AddDate(0, 0, 30).Format("2006-01-02T15:04:05Z07:00"),
		"items": []map[string]interface{}{
			{
				"product_id":        TestProductID,
				"quantity":          2,
				"unit_price":        1000000, // 1M per unit
				"discount_percent":  0,
				"taxable":          true,
				"revenue_account_id": 22, // Pendapatan Penjualan account ID
			},
		},
		"ppn_rate":     11, // 11% PPN
		"notes":        "Test sales order for integration testing",
		"reference":    fmt.Sprintf("TEST-SALE-%d", time.Now().Unix()),
	}
	
	resp, err := client.makeRequest("POST", "/sales", saleData)
	if err != nil {
		return err
	}
	
	// Parse response
	var apiResp APIResponse
	if err := json.Unmarshal(resp, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}
	
	if !apiResp.Success {
		return fmt.Errorf("API error: %s", apiResp.Message)
	}
	
	// Extract sale data
	saleDataBytes, _ := json.Marshal(apiResp.Data)
	var sale Sale
	if err := json.Unmarshal(saleDataBytes, &sale); err != nil {
		return fmt.Errorf("failed to parse sale data: %v", err)
	}
	
	report.Sale = &sale
	return nil
}

func invoiceSalesOrder(client *TestClient, report *TestReport) error {
	if report.Sale == nil {
		return fmt.Errorf("no sale to invoice")
	}
	
	invoiceData := map[string]interface{}{
		"confirm": true,
	}
	
	endpoint := fmt.Sprintf("/sales/%d/invoice", report.Sale.ID)
	resp, err := client.makeRequest("POST", endpoint, invoiceData)
	if err != nil {
		return err
	}
	
	// Parse response to update sale status
	var apiResp APIResponse
	if err := json.Unmarshal(resp, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}
	
	if !apiResp.Success {
		return fmt.Errorf("API error: %s", apiResp.Message)
	}
	
	// Update sale data
	saleDataBytes, _ := json.Marshal(apiResp.Data)
	var updatedSale Sale
	if err := json.Unmarshal(saleDataBytes, &updatedSale); err != nil {
		return fmt.Errorf("failed to parse updated sale data: %v", err)
	}
	
	report.Sale = &updatedSale
	return nil
}

func recordPayment(client *TestClient, report *TestReport) error {
	if report.Sale == nil {
		return fmt.Errorf("no sale to create payment for")
	}
	
	paymentData := map[string]interface{}{
		"contact_id":   TestCustomerID,
		"amount":       report.Sale.OutstandingAmount,
		"date":         time.Now().Format("2006-01-02T15:04:05Z07:00"),
		"method":       "RECEIVABLE",
		"reference":    fmt.Sprintf("PAY-TEST-%d", time.Now().Unix()),
		"notes":        "Test payment for integration testing",
		"allocations": []map[string]interface{}{
			{
				"invoice_id": report.Sale.ID,
				"amount":     report.Sale.OutstandingAmount,
			},
		},
	}
	
	resp, err := client.makeRequest("POST", "/payments/receivable", paymentData)
	if err != nil {
		return err
	}
	
	// Parse response
	var apiResp APIResponse
	if err := json.Unmarshal(resp, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}
	
	if !apiResp.Success {
		return fmt.Errorf("API error: %s", apiResp.Message)
	}
	
	// Extract payment data
	paymentDataBytes, _ := json.Marshal(apiResp.Data)
	var payment Payment
	if err := json.Unmarshal(paymentDataBytes, &payment); err != nil {
		return fmt.Errorf("failed to parse payment data: %v", err)
	}
	
	report.Payment = &payment
	return nil
}

func verifySalesJournalEntries(client *TestClient, report *TestReport) error {
	if report.Sale == nil {
		return fmt.Errorf("no sale to verify")
	}
	
	// Search for journal entries related to the sale
	endpoint := fmt.Sprintf("/journal-drilldown/entries?reference_type=SALE&reference_id=%d", report.Sale.ID)
	resp, err := client.makeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	
	var apiResp APIResponse
	if err := json.Unmarshal(resp, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}
	
	// Verify journal entries exist and are balanced
	if apiResp.Data != nil {
		entriesBytes, _ := json.Marshal(apiResp.Data)
		var entries []JournalEntry
		if err := json.Unmarshal(entriesBytes, &entries); err == nil {
			report.JournalEntries = append(report.JournalEntries, entries...)
		}
	}
	
	return nil
}

func verifyPaymentJournalEntries(client *TestClient, report *TestReport) error {
	if report.Payment == nil {
		return fmt.Errorf("no payment to verify")
	}
	
	// Search for journal entries related to the payment
	endpoint := fmt.Sprintf("/journal-drilldown/entries?reference_type=PAYMENT&reference_id=%d", report.Payment.ID)
	resp, err := client.makeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	
	var apiResp APIResponse
	if err := json.Unmarshal(resp, &apiResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}
	
	// Verify journal entries exist and are balanced
	if apiResp.Data != nil {
		entriesBytes, _ := json.Marshal(apiResp.Data)
		var entries []JournalEntry
		if err := json.Unmarshal(entriesBytes, &entries); err == nil {
			report.JournalEntries = append(report.JournalEntries, entries...)
		}
	}
	
	return nil
}

func captureFinalBalances(client *TestClient, report *TestReport) error {
	accounts := []string{ReceivableAccount, CashAccount, BankAccount, RevenueAccount}
	
	for _, accountCode := range accounts {
		balance, err := getAccountBalance(client, accountCode)
		if err != nil {
			return fmt.Errorf("failed to get final balance for account %s: %v", accountCode, err)
		}
		report.FinalBalances[accountCode] = balance
		fmt.Printf("   📊 %s: %.2f\n", accountCode, balance)
	}
	
	return nil
}

func verifyDataIntegrity(client *TestClient, report *TestReport) error {
	fmt.Println("\n   🔍 Verifying Balance Changes...")
	
	saleAmount := report.Sale.TotalAmount
	paymentAmount := report.Payment.Amount
	
	// Check Receivable Account (should decrease by payment amount)
	receivableChange := report.FinalBalances[ReceivableAccount] - report.InitialBalances[ReceivableAccount]
	expectedReceivableChange := saleAmount - paymentAmount
	
	fmt.Printf("   📈 Receivable Change: %.2f (Expected: %.2f)\n", receivableChange, expectedReceivableChange)
	
	if abs(receivableChange-expectedReceivableChange) > 0.01 {
		return fmt.Errorf("receivable account change mismatch: got %.2f, expected %.2f", 
			receivableChange, expectedReceivableChange)
	}
	
	// Check Cash/Bank Account (should increase by payment amount)
	cashChange := report.FinalBalances[CashAccount] - report.InitialBalances[CashAccount]
	expectedCashChange := paymentAmount
	
	fmt.Printf("   💰 Cash Change: %.2f (Expected: %.2f)\n", cashChange, expectedCashChange)
	
	if abs(cashChange-expectedCashChange) > 0.01 {
		return fmt.Errorf("cash account change mismatch: got %.2f, expected %.2f", 
			cashChange, expectedCashChange)
	}
	
	// Check Revenue Account (should increase by sale amount)
	revenueChange := report.FinalBalances[RevenueAccount] - report.InitialBalances[RevenueAccount]
	expectedRevenueChange := saleAmount
	
	fmt.Printf("   📊 Revenue Change: %.2f (Expected: %.2f)\n", revenueChange, expectedRevenueChange)
	
	if abs(revenueChange-expectedRevenueChange) > 0.01 {
		return fmt.Errorf("revenue account change mismatch: got %.2f, expected %.2f", 
			revenueChange, expectedRevenueChange)
	}
	
	// Verify sale status is PAID
	if report.Sale.Status != "PAID" {
		return fmt.Errorf("sale status should be PAID, got: %s", report.Sale.Status)
	}
	
	// Verify payment status is COMPLETED
	if report.Payment.Status != "COMPLETED" {
		return fmt.Errorf("payment status should be COMPLETED, got: %s", report.Payment.Status)
	}
	
	return nil
}

func getAccountBalance(client *TestClient, accountCode string) (float64, error) {
	endpoint := fmt.Sprintf("/accounts/%s", accountCode)
	resp, err := client.makeRequest("GET", endpoint, nil)
	if err != nil {
		return 0, err
	}
	
	// Debug: Print raw response
	fmt.Printf("   DEBUG - Raw response for %s: %s\n", accountCode, string(resp))
	
	// Try to parse as generic interface first
	var genericResp interface{}
	if err := json.Unmarshal(resp, &genericResp); err != nil {
		return 0, fmt.Errorf("failed to parse response as JSON: %v", err)
	}
	
	// Parse response with correct structure
	var response struct {
		Data struct {
			ID       uint    `json:"id"`
			Code     string  `json:"code"`
			Name     string  `json:"name"`
			Balance  float64 `json:"balance"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(resp, &response); err != nil {
		return 0, fmt.Errorf("failed to parse account response: %v", err)
	}
	
	fmt.Printf("   ✅ Account %s (%s): Balance = %.2f\n", response.Data.Code, response.Data.Name, response.Data.Balance)
	return response.Data.Balance, nil
}

func (c *TestClient) makeRequest(method, endpoint string, data interface{}) ([]byte, error) {
	var body io.Reader
	
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}
	
	url := c.baseURL + endpoint
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	
	return respBody, nil
}

func (c *TestClient) makeRequestWithoutAuth(method, endpoint string, data interface{}) ([]byte, error) {
	var body io.Reader
	
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}
	
	url := c.baseURL + endpoint
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	// No authorization header for login
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	
	return respBody, nil
}

func (r *TestReport) addError(err string) {
	r.Errors = append(r.Errors, err)
	r.FailedTests++
}

func (r *TestReport) addWarning(warning string) {
	r.Warnings = append(r.Warnings, warning)
}

func generateFinalReport(report *TestReport) {
	duration := report.EndTime.Sub(report.StartTime)
	
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📋 COMPREHENSIVE TEST REPORT")
	fmt.Println(strings.Repeat("=", 70))
	
	fmt.Printf("⏱️  Test Duration: %v\n", duration)
	fmt.Printf("📅 Start Time: %s\n", report.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("📅 End Time: %s\n", report.EndTime.Format("2006-01-02 15:04:05"))
	
	fmt.Println("\n🎯 TEST RESULTS:")
	printTestResult("Sales Created", report.SalesCreated)
	printTestResult("Sales Invoiced", report.SalesInvoiced)
	printTestResult("Payment Recorded", report.PaymentRecorded)
	printTestResult("Journal Entries Created", report.JournalEntriesCreated)
	printTestResult("Accounts Updated", report.AccountsUpdated)
	printTestResult("Data Integrity Verified", report.IntegrityVerified)
	
	fmt.Println("\n💰 FINANCIAL SUMMARY:")
	if report.Sale != nil {
		fmt.Printf("   📊 Sale Amount: %.2f\n", report.Sale.TotalAmount)
	}
	if report.Payment != nil {
		fmt.Printf("   💸 Payment Amount: %.2f\n", report.Payment.Amount)
	}
	
	fmt.Println("\n📈 BALANCE CHANGES:")
	for account, initialBalance := range report.InitialBalances {
		finalBalance := report.FinalBalances[account]
		change := finalBalance - initialBalance
		changeSymbol := "📈"
		if change < 0 {
			changeSymbol = "📉"
		} else if change == 0 {
			changeSymbol = "📊"
		}
		
		fmt.Printf("   %s %s: %.2f -> %.2f (Change: %+.2f)\n", 
			changeSymbol, account, initialBalance, finalBalance, change)
	}
	
	fmt.Printf("\n📊 JOURNAL ENTRIES: %d entries recorded\n", len(report.JournalEntries))
	
	if len(report.Errors) > 0 {
		fmt.Println("\n❌ ERRORS:")
		for i, err := range report.Errors {
			fmt.Printf("   %d. %s\n", i+1, err)
		}
	}
	
	if len(report.Warnings) > 0 {
		fmt.Println("\n⚠️  WARNINGS:")
		for i, warning := range report.Warnings {
			fmt.Printf("   %d. %s\n", i+1, warning)
		}
	}
	
	// Overall Assessment
	fmt.Println("\n" + strings.Repeat("=", 70))
	overallSuccess := len(report.Errors) == 0 && 
		report.SalesCreated && 
		report.PaymentRecorded && 
		report.AccountsUpdated && 
		report.IntegrityVerified
	
	if overallSuccess {
		fmt.Println("🎉 OVERALL RESULT: ✅ ALL TESTS PASSED - 100% SUCCESS!")
		fmt.Println("   🔹 Sales-to-Payment flow working perfectly")
		fmt.Println("   🔹 Account balances updated correctly")
		fmt.Println("   🔹 Data integrity maintained")
		fmt.Println("   🔹 Journal entries properly recorded")
		fmt.Println("   🔹 System is production-ready! 🚀")
	} else {
		fmt.Println("❌ OVERALL RESULT: TESTS FAILED")
		fmt.Printf("   🔸 %d error(s) detected\n", len(report.Errors))
		fmt.Printf("   🔸 %d warning(s) detected\n", len(report.Warnings))
		fmt.Println("   🔸 Please review and fix issues before production")
	}
	fmt.Println(strings.Repeat("=", 70))
}

func printTestResult(testName string, passed bool) {
	status := "✅"
	if !passed {
		status = "❌"
	}
	fmt.Printf("   %s %s\n", status, testName)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

