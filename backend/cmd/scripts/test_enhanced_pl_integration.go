package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Test the enhanced profit loss API endpoints
func main() {
	fmt.Println("🧪 Testing Enhanced Profit & Loss API Integration")
	fmt.Println(strings.Repeat("=", 60))
	
	// Test endpoints to check
	endpoints := []struct {
		name string
		url  string
		method string
	}{
		{
			name: "Enhanced P&L (GET method)",
			url: "http://localhost:8080/api/reports/enhanced/profit-loss?start_date=2024-01-01&end_date=2024-12-31&format=json",
			method: "GET",
		},
		{
			name: "Enhanced P&L (POST method)",
			url: "http://localhost:8080/reports/enhanced/profit-loss",
			method: "POST",
		},
		{
			name: "Comprehensive P&L",
			url: "http://localhost:8080/api/reports/comprehensive/profit-loss?start_date=2024-01-01&end_date=2024-12-31&format=json",
			method: "GET",
		},
		{
			name: "Frontend P&L (v1 API)",
			url: "http://localhost:8080/api/v1/reports/profit-loss?start_date=2025-08-31&end_date=2025-09-17&format=json",
			method: "GET",
		},
	}
	
	for i, endpoint := range endpoints {
		fmt.Printf("\n%d. Testing %s\n", i+1, endpoint.name)
		fmt.Printf("   Method: %s\n", endpoint.method)
		fmt.Printf("   URL: %s\n", endpoint.url)
		
		err := testEndpoint(endpoint.url, endpoint.method)
		if err != nil {
			fmt.Printf("   ❌ FAILED: %v\n", err)
		} else {
			fmt.Printf("   ✅ SUCCESS\n")
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 Integration test completed!")
	fmt.Println("💡 If any endpoints failed, check:")
	fmt.Println("   - Backend server is running on localhost:8080")
	fmt.Println("   - Required routes are properly registered") 
	fmt.Println("   - Journal entries exist in database")
	fmt.Println("   - Account categorization is correct")
}

func testEndpoint(url, method string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	
	var req *http.Request
	var err error
	
	if method == "POST" {
		// For POST requests, include request body
		reqBody := `{
			"start_date": "2024-01-01",
			"end_date": "2024-12-31", 
			"format": "json"
		}`
		req, err = http.NewRequest(method, url, strings.NewReader(reqBody))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}
	}
	
	// Add common headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Enhanced-PL-Integration-Test/1.0")
	
	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}
	
	// Log response details
	fmt.Printf("   Status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("   Response Size: %d bytes\n", len(body))
	
	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("   Response Body: %s\n", string(body))
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}
	
	// Try to parse as JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("   Warning: Response is not valid JSON: %v\n", err)
		fmt.Printf("   Raw Response: %s\n", string(body)[:min(500, len(body))])
	} else {
		// Check for expected fields in profit loss response
		checkProfitLossStructure(result)
	}
	
	return nil
}

func checkProfitLossStructure(data map[string]interface{}) {
	fmt.Printf("   🔍 Checking data structure...\n")
	
	// Check for enhanced structure fields
	enhancedFields := []string{"revenue", "cost_of_goods_sold", "gross_profit", "operating_expenses", "net_income"}
	enhancedFieldsFound := 0
	
	for _, field := range enhancedFields {
		if _, exists := data[field]; exists {
			enhancedFieldsFound++
		}
	}
	
	if enhancedFieldsFound >= 3 {
		fmt.Printf("   📊 Enhanced P&L structure detected (%d/5 fields found)\n", enhancedFieldsFound)
		
		// Check revenue structure
		if revenue, ok := data["revenue"].(map[string]interface{}); ok {
			if totalRev, ok := revenue["total_revenue"].(float64); ok && totalRev > 0 {
				fmt.Printf("   💰 Revenue data found: %.2f\n", totalRev)
			}
		}
		
		// Check COGS structure  
		if cogs, ok := data["cost_of_goods_sold"].(map[string]interface{}); ok {
			if totalCogs, ok := cogs["total_cogs"].(float64); ok && totalCogs > 0 {
				fmt.Printf("   📦 COGS data found: %.2f\n", totalCogs)
			}
		}
		
		// Check net income
		if netIncome, ok := data["net_income"].(float64); ok {
			fmt.Printf("   🎯 Net Income: %.2f\n", netIncome)
		}
		
	} else {
		fmt.Printf("   📋 Legacy P&L structure detected\n")
		
		// Check for basic fields
		basicFields := []string{"total_revenue", "total_expenses", "net_income"}
		for _, field := range basicFields {
			if value, exists := data[field]; exists {
				fmt.Printf("   %s: %v\n", field, value)
			}
		}
	}
	
	// Check for success indicators
	if success, ok := data["success"].(bool); ok && success {
		fmt.Printf("   ✅ Success flag: true\n")
	}
	
	if message, ok := data["message"].(string); ok && message != "" {
		fmt.Printf("   💬 Message: %s\n", message)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}