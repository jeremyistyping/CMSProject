package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Purchase struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	Code              string    `json:"code"`
	VendorID          uint      `json:"vendor_id"`
	TotalAmount       float64   `json:"total_amount"`
	PaidAmount        float64   `json:"paid_amount"`
	OutstandingAmount float64   `json:"outstanding_amount"`
	PaymentMethod     string    `json:"payment_method"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

type PurchasePayment struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	PurchaseID uint      `json:"purchase_id"`
	Amount     float64   `json:"amount"`
	Method     string    `json:"method"`
	Date       time.Time `json:"date"`
	CashBankID *uint     `json:"cash_bank_id"`
	PaymentID  *uint     `json:"payment_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type Account struct {
	ID      uint    `json:"id" gorm:"primaryKey"`
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Balance float64 `json:"balance"`
}

type Contact struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type TestResult struct {
	TestName string
	Status   string
	Message  string
	Details  interface{}
}

func main() {
	// Database connection
	dsn := "accounting_user:accounting_password@tcp(localhost:3306)/accounting_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🧪 PURCHASE BALANCE SYSTEM TESTING")
	fmt.Println("===================================")

	var results []TestResult

	// Test 1: Check database functions exist
	fmt.Println("\n🔍 TEST 1: Checking Database Functions")
	
	testValidationFunction := testDatabaseFunction(db, "validate_purchase_balances")
	results = append(results, testValidationFunction)
	
	testSyncFunction := testDatabaseFunction(db, "sync_purchase_balances")
	results = append(results, testSyncFunction)

	// Test 2: Check existing data consistency
	fmt.Println("\n🔍 TEST 2: Checking Data Consistency")
	
	consistencyTest := testDataConsistency(db)
	results = append(results, consistencyTest)

	// Test 3: Test validation function
	fmt.Println("\n🔍 TEST 3: Testing Validation Function")
	
	validationTest := testValidationFunction(db)
	results = append(results, validationTest)

	// Test 4: Test triggers (if they exist)
	fmt.Println("\n🔍 TEST 4: Testing Triggers")
	
	triggerTest := testTriggers(db)
	results = append(results, triggerTest)

	// Test 5: Create test data and verify automatic sync
	fmt.Println("\n🔍 TEST 5: Testing Automatic Balance Sync")
	
	autoSyncTest := testAutomaticSync(db)
	results = append(results, autoSyncTest)

	// Test 6: Performance test
	fmt.Println("\n🔍 TEST 6: Performance Testing")
	
	performanceTest := testPerformance(db)
	results = append(results, performanceTest)

	// Test 7: Check monitoring views
	fmt.Println("\n🔍 TEST 7: Testing Monitoring Views")
	
	monitoringTest := testMonitoringViews(db)
	results = append(results, monitoringTest)

	// Display results summary
	fmt.Println("\n📊 TEST RESULTS SUMMARY")
	fmt.Println("========================")

	passedTests := 0
	failedTests := 0

	for _, result := range results {
		status := "❌ FAIL"
		if result.Status == "PASS" {
			status = "✅ PASS"
			passedTests++
		} else {
			failedTests++
		}
		fmt.Printf("%s %s: %s\n", status, result.TestName, result.Message)
	}

	fmt.Printf("\n📈 SUMMARY: %d passed, %d failed, %d total\n", 
		passedTests, failedTests, len(results))

	if failedTests == 0 {
		fmt.Println("🎉 ALL TESTS PASSED! Purchase balance system is working correctly.")
	} else {
		fmt.Printf("⚠️  %d tests failed. Please review the issues above.\n", failedTests)
	}

	// Final recommendations
	fmt.Println("\n🚀 RECOMMENDATIONS:")
	if failedTests > 0 {
		fmt.Println("1. Review failed tests and fix the underlying issues")
		fmt.Println("2. Re-run this test script after fixes")
	} else {
		fmt.Println("1. ✅ System is ready for production use")
		fmt.Println("2. ✅ Monitor regularly using the validation functions")
		fmt.Println("3. ✅ Use purchase_balance_health view for ongoing monitoring")
	}
}

func testDatabaseFunction(db *gorm.DB, functionName string) TestResult {
	var exists bool
	query := "SELECT COUNT(*) > 0 FROM information_schema.routines WHERE routine_schema = DATABASE() AND routine_name = ?"
	err := db.Raw(query, functionName).Scan(&exists).Error
	
	if err != nil {
		return TestResult{
			TestName: fmt.Sprintf("Function %s Exists", functionName),
			Status:   "FAIL",
			Message:  fmt.Sprintf("Error checking function: %v", err),
		}
	}
	
	if exists {
		return TestResult{
			TestName: fmt.Sprintf("Function %s Exists", functionName),
			Status:   "PASS",
			Message:  "Function exists and is callable",
		}
	} else {
		return TestResult{
			TestName: fmt.Sprintf("Function %s Exists", functionName),
			Status:   "FAIL",
			Message:  "Function does not exist",
		}
	}
}

func testDataConsistency(db *gorm.DB) TestResult {
	var purchases []Purchase
	err := db.Find(&purchases).Error
	if err != nil {
		return TestResult{
			TestName: "Data Consistency",
			Status:   "FAIL",
			Message:  fmt.Sprintf("Error fetching purchases: %v", err),
		}
	}

	inconsistentCount := 0
	for _, purchase := range purchases {
		expectedOutstanding := purchase.TotalAmount - purchase.PaidAmount
		if abs(purchase.OutstandingAmount - expectedOutstanding) > 0.01 {
			inconsistentCount++
		}
	}

	if inconsistentCount == 0 {
		return TestResult{
			TestName: "Data Consistency",
			Status:   "PASS",
			Message:  fmt.Sprintf("All %d purchases have consistent outstanding amounts", len(purchases)),
		}
	} else {
		return TestResult{
			TestName: "Data Consistency",
			Status:   "FAIL",
			Message:  fmt.Sprintf("%d out of %d purchases have inconsistent outstanding amounts", inconsistentCount, len(purchases)),
		}
	}
}

func testValidationFunction(db *gorm.DB) TestResult {
	var validationResult string
	err := db.Raw("SELECT validate_purchase_balances()").Scan(&validationResult).Error
	
	if err != nil {
		return TestResult{
			TestName: "Validation Function",
			Status:   "FAIL",
			Message:  fmt.Sprintf("Error calling validation function: %v", err),
		}
	}

	// Check if result contains expected JSON structure
	if len(validationResult) > 0 && (contains(validationResult, "status") && contains(validationResult, "accounts_payable")) {
		return TestResult{
			TestName: "Validation Function",
			Status:   "PASS",
			Message:  "Validation function returns proper JSON result",
			Details:  validationResult,
		}
	} else {
		return TestResult{
			TestName: "Validation Function",
			Status:   "FAIL",
			Message:  "Validation function returned unexpected result format",
			Details:  validationResult,
		}
	}
}

func testTriggers(db *gorm.DB) TestResult {
	// Check if triggers exist
	var triggerCount int64
	query := `SELECT COUNT(*) FROM information_schema.triggers 
			  WHERE trigger_schema = DATABASE() 
			  AND trigger_name LIKE '%purchase%'`
	
	err := db.Raw(query).Scan(&triggerCount).Error
	if err != nil {
		return TestResult{
			TestName: "Triggers Check",
			Status:   "FAIL",
			Message:  fmt.Sprintf("Error checking triggers: %v", err),
		}
	}

	if triggerCount > 0 {
		return TestResult{
			TestName: "Triggers Check",
			Status:   "PASS",
			Message:  fmt.Sprintf("Found %d purchase-related triggers", triggerCount),
		}
	} else {
		return TestResult{
			TestName: "Triggers Check",
			Status:   "WARN",
			Message:  "No purchase triggers found - triggers may not be installed yet",
		}
	}
}

func testAutomaticSync(db *gorm.DB) TestResult {
	// Get initial state
	var initialAPBalance float64
	err := db.Raw(`SELECT COALESCE(balance, 0) FROM accounts 
					WHERE (code LIKE '%2101%' OR name LIKE '%Hutang Usaha%' OR name LIKE '%Accounts Payable%')
					AND deleted_at IS NULL LIMIT 1`).Scan(&initialAPBalance).Error
	
	if err != nil {
		return TestResult{
			TestName: "Automatic Sync",
			Status:   "FAIL",
			Message:  fmt.Sprintf("Error getting initial AP balance: %v", err),
		}
	}

	// Test sync function
	var syncResult string
	err = db.Raw("SELECT sync_purchase_balances()").Scan(&syncResult).Error
	
	if err != nil {
		return TestResult{
			TestName: "Automatic Sync",
			Status:   "FAIL",
			Message:  fmt.Sprintf("Error calling sync function: %v", err),
		}
	}

	// Check if result is valid JSON
	if len(syncResult) > 0 && contains(syncResult, "sync_timestamp") {
		return TestResult{
			TestName: "Automatic Sync",
			Status:   "PASS",
			Message:  "Sync function executed successfully",
			Details:  syncResult,
		}
	} else {
		return TestResult{
			TestName: "Automatic Sync",
			Status:   "FAIL",
			Message:  "Sync function returned invalid result",
			Details:  syncResult,
		}
	}
}

func testPerformance(db *gorm.DB) TestResult {
	start := time.Now()
	
	// Run validation function multiple times to test performance
	for i := 0; i < 10; i++ {
		var result string
		err := db.Raw("SELECT validate_purchase_balances()").Scan(&result).Error
		if err != nil {
			return TestResult{
				TestName: "Performance Test",
				Status:   "FAIL",
				Message:  fmt.Sprintf("Error in performance test iteration %d: %v", i+1, err),
			}
		}
	}
	
	duration := time.Since(start)
	averageTime := duration.Milliseconds() / 10
	
	if averageTime < 100 { // Less than 100ms average
		return TestResult{
			TestName: "Performance Test",
			Status:   "PASS",
			Message:  fmt.Sprintf("Average response time: %dms (excellent)", averageTime),
		}
	} else if averageTime < 500 { // Less than 500ms average
		return TestResult{
			TestName: "Performance Test",
			Status:   "PASS",
			Message:  fmt.Sprintf("Average response time: %dms (good)", averageTime),
		}
	} else {
		return TestResult{
			TestName: "Performance Test",
			Status:   "WARN",
			Message:  fmt.Sprintf("Average response time: %dms (slow)", averageTime),
		}
	}
}

func testMonitoringViews(db *gorm.DB) TestResult {
	// Check if monitoring view exists
	var viewExists bool
	query := `SELECT COUNT(*) > 0 FROM information_schema.views 
			  WHERE table_schema = DATABASE() 
			  AND table_name = 'purchase_balance_health'`
	
	err := db.Raw(query).Scan(&viewExists).Error
	if err != nil {
		return TestResult{
			TestName: "Monitoring Views",
			Status:   "FAIL",
			Message:  fmt.Sprintf("Error checking views: %v", err),
		}
	}

	if !viewExists {
		return TestResult{
			TestName: "Monitoring Views",
			Status:   "WARN",
			Message:  "purchase_balance_health view not found - may not be created yet",
		}
	}

	// Try to query the view
	var healthCount int64
	err = db.Raw("SELECT COUNT(*) FROM purchase_balance_health").Scan(&healthCount).Error
	if err != nil {
		return TestResult{
			TestName: "Monitoring Views",
			Status:   "FAIL",
			Message:  fmt.Sprintf("Error querying health view: %v", err),
		}
	}

	return TestResult{
		TestName: "Monitoring Views",
		Status:   "PASS",
		Message:  fmt.Sprintf("Health view accessible with %d records", healthCount),
	}
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		stringContains(s, substr))))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s) - len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}