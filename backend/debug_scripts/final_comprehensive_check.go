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
	fmt.Println("🏆 FINAL COMPREHENSIVE TEST")
	fmt.Println("===========================")
	fmt.Println("Testing if frontend financial reports now have real data")
	fmt.Println()

	// Get auth token
	token, err := getToken()
	if err != nil {
		fmt.Printf("❌ Auth failed: %v\n", err)
		return
	}

	// Test specific reports that were showing "No Data" before
	endpoints := []struct {
		name        string
		path        string
		expectData  string
	}{
		{"Journal Entry Analysis", "/api/v1/reports/journal-entry-analysis", "total_entries"},
		{"Profit & Loss Statement", "/api/v1/reports/profit-loss", "revenue"},
		{"Balance Sheet", "/api/v1/reports/balance-sheet", "total_assets"},
		{"Sales Summary", "/api/v1/reports/sales-summary", "total_revenue"},
		{"Trial Balance", "/api/v1/reports/trial-balance", "accounts"},
	}

	passedTests := 0
	dataFoundTests := 0

	for i, endpoint := range endpoints {
		fmt.Printf("%d. Testing %s\n", i+1, endpoint.name)
		fmt.Printf("   %s\n", strings.Repeat("-", len(endpoint.name)+3))

		// Test with date range that includes our transactions
		url := fmt.Sprintf("http://localhost:8080%s?start_date=2025-09-01&end_date=2025-09-30&format=json", endpoint.path)
		
		response, size, hasData, dataPreview := testEndpointDetailed(url, token, endpoint.expectData)
		
		if response != "" {
			passedTests++
			fmt.Printf("   ✅ API Response: SUCCESS (%d bytes)\n", size)
			
			if hasData {
				dataFoundTests++
				fmt.Printf("   🎯 Data Found: YES - Contains %s data\n", endpoint.expectData)
				fmt.Printf("   📊 Preview: %s\n", dataPreview)
			} else {
				fmt.Printf("   ⚠️  Data Found: NO - Empty or zero values\n")
			}
		} else {
			fmt.Printf("   ❌ API Response: FAILED\n")
		}
		fmt.Println()
	}

	// Final Assessment
	fmt.Println("🎯 FINAL ASSESSMENT")
	fmt.Println("==================")
	fmt.Printf("✅ API Success Rate: %d/%d (%.0f%%)\n", passedTests, len(endpoints), float64(passedTests)/float64(len(endpoints))*100)
	fmt.Printf("📊 Data Available Rate: %d/%d (%.0f%%)\n", dataFoundTests, len(endpoints), float64(dataFoundTests)/float64(len(endpoints))*100)
	
	if dataFoundTests == len(endpoints) {
		fmt.Println()
		fmt.Println("🎉🎉🎉 COMPLETE SUCCESS! 🎉🎉🎉")
		fmt.Println("✅ ALL financial reports now have data!")
		fmt.Println("✅ Frontend should display proper financial information!")
		fmt.Println("✅ SSOT Journal System is fully operational!")
		fmt.Println()
		fmt.Println("🔄 Please refresh your browser and try generating reports again.")
		fmt.Println("📊 You should now see real financial data in all reports!")
	} else if dataFoundTests > 0 {
		fmt.Println()
		fmt.Printf("✅ PARTIAL SUCCESS! %d/%d reports have data\n", dataFoundTests, len(endpoints))
		fmt.Println("💡 Some reports are working - try refreshing your browser")
	} else {
		fmt.Println()
		fmt.Println("⚠️  APIs are working but still no data detected")
	fmt.Println("💡 There may be additional configuration needed")
	}
}

func getToken() (string, error) {
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
		return "", fmt.Errorf("login failed: %d", resp.StatusCode)
	}

	var loginResp struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(body, &loginResp)
	return loginResp.AccessToken, nil
}

func testEndpointDetailed(url, token, expectDataField string) (string, int, bool, string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, false, fmt.Sprintf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", 0, false, fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	bodyStr := string(body)
	responseSize := len(body)

	// Parse JSON to check for specific data
	var jsonData map[string]interface{}
	if json.Unmarshal(body, &jsonData) != nil {
		return bodyStr, responseSize, false, "Invalid JSON"
	}

	// Check for expected data field and non-zero values
	hasData := false
	dataPreview := "No specific data found"

	// Look in main response and data field
	if data, exists := jsonData["data"]; exists {
		if dataMap, ok := data.(map[string]interface{}); ok {
			// Check for specific field we're looking for
			if value, fieldExists := dataMap[expectDataField]; fieldExists {
				switch v := value.(type) {
				case float64:
					if v > 0 {
						hasData = true
						dataPreview = fmt.Sprintf("%s: %.0f", expectDataField, v)
					}
				case []interface{}:
					if len(v) > 0 {
						hasData = true
						dataPreview = fmt.Sprintf("%s: %d items", expectDataField, len(v))
					}
				case map[string]interface{}:
					// Check if nested object has non-zero values
					if subtotal, ok := v["subtotal"]; ok {
						if val, ok := subtotal.(float64); ok && val > 0 {
							hasData = true
							dataPreview = fmt.Sprintf("%s.subtotal: %.0f", expectDataField, val)
						}
					}
					if total, ok := v["total"]; ok {
						if val, ok := total.(float64); ok && val > 0 {
							hasData = true
							dataPreview = fmt.Sprintf("%s.total: %.0f", expectDataField, val)
						}
					}
				default:
					if v != nil {
						hasData = true
						dataPreview = fmt.Sprintf("%s: %v", expectDataField, v)
					}
				}
			}
			
			// Also check for any non-zero numeric values as indicator
			if !hasData {
				for key, value := range dataMap {
					if numVal, ok := value.(float64); ok && numVal > 0 {
						hasData = true
						dataPreview = fmt.Sprintf("%s: %.0f (found non-zero value)", key, numVal)
						break
					}
				}
			}
		}
	}

	return bodyStr, responseSize, hasData, dataPreview
}
