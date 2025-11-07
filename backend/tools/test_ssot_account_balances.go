package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type APIResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data"`
	AsOf    string      `json:"as_of"`
	Source  string      `json:"source"`
	Message string      `json:"message"`
	Error   string      `json:"error"`
}

func main() {
	baseURL := "http://localhost:8080"
	
	// Step 1: Login to get JWT token
	fmt.Println("🔐 Logging in to get JWT token...")
	
	loginReq := LoginRequest{
		Email:    "admin@company.com",
		Password: "password123", // Default admin password from seed
	}
	
	loginData, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginData))
	if err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Login failed with status %d: %s\n", resp.StatusCode, string(body))
		fmt.Println("\n💡 Make sure backend server is running and admin user exists")
		os.Exit(1)
	}
	
	var loginResp LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginResp)
	
	if loginResp.Token == "" {
		fmt.Println("❌ No token received from login")
		os.Exit(1)
	}
	
	fmt.Printf("✅ Login successful, token: %s...\n", loginResp.Token[:20])
	
	// Step 2: Test SSOT Account Balances endpoint
	fmt.Println("\n📊 Testing SSOT Account Balances endpoint...")
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"/api/v1/ssot-reports/account-balances", nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		os.Exit(1)
	}
	
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ API request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("📡 Response Status: %d\n", resp.StatusCode)
	
	if resp.StatusCode != 200 {
		fmt.Printf("❌ API request failed with status %d\n", resp.StatusCode)
		fmt.Printf("Response: %s\n", string(body))
		os.Exit(1)
	}
	
	// Parse response
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(body))
		os.Exit(1)
	}
	
	// Display results
	fmt.Printf("✅ API call successful!\n")
	fmt.Printf("Status: %s\n", apiResp.Status)
	fmt.Printf("Source: %s\n", apiResp.Source)
	fmt.Printf("As Of Date: %s\n", apiResp.AsOf)
	if apiResp.Message != "" {
		fmt.Printf("Message: %s\n", apiResp.Message)
	}
	
	// Count data
	if dataArray, ok := apiResp.Data.([]interface{}); ok {
		fmt.Printf("📊 Found %d accounts with balance data\n", len(dataArray))
		
		// Show first few entries as sample
		fmt.Println("\n📋 Sample Account Balances:")
		for i, item := range dataArray {
			if i >= 5 { // Show only first 5
				break
			}
			if account, ok := item.(map[string]interface{}); ok {
				fmt.Printf("  %s (%s): %s - Balance: %.2f\n", 
					account["account_code"], 
					account["account_type"],
					account["account_name"],
					account["net_balance"])
			}
		}
		
		if len(dataArray) > 5 {
			fmt.Printf("  ... and %d more accounts\n", len(dataArray)-5)
		}
	} else {
		fmt.Printf("Data: %+v\n", apiResp.Data)
	}
	
	fmt.Println("\n✅ SSOT Account Balances API test completed successfully!")
	fmt.Println("🎯 The frontend should now be able to fetch account balances correctly.")
}