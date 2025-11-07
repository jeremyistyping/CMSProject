package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("🧪 Testing Journal Entry Analysis API Directly")
	log.Println("===============================================")

	// Test the API endpoint directly
	testAPIEndpoint()
}

func testAPIEndpoint() {
	baseURL := "http://localhost:8080"
	
	// Test different parameter combinations
	testCases := []struct {
		name        string
		params      string
		expectData  bool
		description string
	}{
		{
			name:        "All entries",
			params:      "start_date=2025-09-01&end_date=2025-09-19&status=ALL&reference_type=ALL&format=json",
			expectData:  true,
			description: "Should return 2 entries (1 DRAFT + 1 POSTED)",
		},
		{
			name:        "Only POSTED",
			params:      "start_date=2025-09-01&end_date=2025-09-19&status=POSTED&reference_type=ALL&format=json",
			expectData:  true,
			description: "Should return 1 POSTED entry",
		},
		{
			name:        "Only DRAFT", 
			params:      "start_date=2025-09-01&end_date=2025-09-19&status=DRAFT&reference_type=ALL&format=json",
			expectData:  true,
			description: "Should return 1 DRAFT entry",
		},
		{
			name:        "Frontend simulation (current month)",
			params:      getCurrentMonthParams(),
			expectData:  true,
			description: "Simulates what frontend sends currently",
		},
	}

	for i, testCase := range testCases {
		log.Printf("\n🔬 Test %d: %s", i+1, testCase.name)
		log.Printf("   Description: %s", testCase.description)
		log.Printf("   Parameters: %s", testCase.params)
		
		url := fmt.Sprintf("%s/api/v1/reports/journal-entry-analysis?%s", baseURL, testCase.params)
		log.Printf("   URL: %s", url)
		
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("   ❌ Request failed: %v", err)
			continue
		}
		defer resp.Body.Close()
		
		log.Printf("   📡 Status Code: %d", resp.StatusCode)
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("   ❌ Failed to read response: %v", err)
			continue
		}
		
		if resp.StatusCode != 200 {
			log.Printf("   ❌ Error Response: %s", string(body))
			continue
		}
		
		// Parse JSON response
		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err != nil {
			log.Printf("   ❌ Failed to parse JSON: %v", err)
			log.Printf("   Raw response: %s", string(body))
			continue
		}
		
		// Check response structure
		if status, ok := response["status"].(string); ok && status == "success" {
			if data, ok := response["data"].(map[string]interface{}); ok {
				if totalEntries, ok := data["total_entries"].(float64); ok {
					log.Printf("   ✅ Success: Found %.0f entries", totalEntries)
					
					// Show some entry details
					if entriesByType, ok := data["entries_by_type"].([]interface{}); ok {
						log.Printf("   📊 Entries by type:")
						for _, entry := range entriesByType {
							if entryMap, ok := entry.(map[string]interface{}); ok {
								refType := entryMap["reference_type"]
								count := entryMap["count"]
								amount := entryMap["total_amount"]
								log.Printf("      %s: %.0f entries (Rp %.2f)", refType, count, amount)
							}
						}
					}
					
					// Show compliance check
					if compliance, ok := data["compliance_check"].(map[string]interface{}); ok {
						balanced := compliance["balanced_entries"]
						unbalanced := compliance["unbalanced_entries"]
						rate := compliance["compliance_rate"]
						log.Printf("   🔍 Compliance: %.0f balanced, %.0f unbalanced (%.1f%%)", balanced, unbalanced, rate)
					}
					
					// Show recent entries
					if recentEntries, ok := data["recent_entries"].([]interface{}); ok {
						log.Printf("   📄 Recent entries (%d):", len(recentEntries))
						for i, entry := range recentEntries {
							if i < 3 { // Show first 3
								if entryMap, ok := entry.(map[string]interface{}); ok {
									desc := entryMap["description"]
									amount := entryMap["debit_amount"]
									date := entryMap["date"]
									log.Printf("      %d. %s: %s (Rp %.2f)", i+1, date, desc, amount)
								}
							}
						}
					}
				} else {
					log.Printf("   ⚠️  No total_entries field found")
				}
			} else {
				log.Printf("   ⚠️  No data field found in successful response")
			}
		} else {
			log.Printf("   ❌ Response indicates failure")
		}
		
		// Show raw response for debugging (first 500 chars)
		responseStr := string(body)
		if len(responseStr) > 500 {
			responseStr = responseStr[:500] + "..."
		}
		log.Printf("   📄 Response preview: %s", responseStr)
	}
	
	log.Println("\n🏁 API testing completed!")
}

func getCurrentMonthParams() string {
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	
	return fmt.Sprintf("start_date=%s&end_date=%s&status=ALL&reference_type=ALL&format=json",
		firstDay.Format("2006-01-02"),
		now.Format("2006-01-02"))
}