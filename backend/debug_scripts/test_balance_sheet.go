package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	fmt.Println("🔍 Balance Sheet Debug Test")
	fmt.Println("===========================")

	// Get token
	token, _ := getToken()
	
	// Test Balance Sheet API
	url := "http://localhost:8080/api/v1/reports/balance-sheet?start_date=2025-09-01&end_date=2025-09-30&format=json"
	
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	
	// Pretty print the response
	var jsonData map[string]interface{}
	json.Unmarshal(body, &jsonData)
	
	prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")
	fmt.Println("📄 Balance Sheet Response:")
	fmt.Println(string(prettyJSON))
	
	// Check account_balances directly
	fmt.Println("\n🔍 Let's check account_balances view directly...")
	
	// This would require database access but let's see the API response structure
	if data, exists := jsonData["data"]; exists {
		if dataMap, ok := data.(map[string]interface{}); ok {
			fmt.Println("\n📊 Data structure analysis:")
			for key, value := range dataMap {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
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
	var loginResp struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(body, &loginResp)
	return loginResp.AccessToken, nil
}