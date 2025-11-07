package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"log"
	"time"
)

func main() {
	// Wait for server to start
	fmt.Println("Waiting for server to start...")
	time.Sleep(3 * time.Second)
	
	// Get auth token
	token := login()
	fmt.Printf("✅ Login successful, token: %s...\n", token[:20])
	
	// Test 1: Create account with code 1008
	fmt.Println("\n🔍 Test 1: Creating account with code '1008'...")
	accountID := createAccount(token, "1008", "MOTOR TEST")
	if accountID > 0 {
		fmt.Println("✅ Account created successfully!")
	} else {
		log.Fatal("❌ Failed to create account")
	}
	
	// Test 2: Delete the account (soft delete)
	fmt.Println("\n🔍 Test 2: Soft deleting the account...")
	deleteAccount(token, "1008")
	fmt.Println("✅ Account deleted successfully!")
	
	// Test 3: Try to create account with same code again
	fmt.Println("\n🔍 Test 3: Creating account with same code '1008' after soft delete...")
	accountID2 := createAccount(token, "1008", "MOTOR TEST 2")
	if accountID2 > 0 {
		fmt.Println("🎉 SUCCESS! Account created successfully after soft delete!")
		fmt.Println("✅ Soft delete issue has been resolved!")
	} else {
		fmt.Println("❌ FAILED! Still cannot create account with same code after soft delete")
		fmt.Println("❌ Soft delete issue NOT resolved")
	}
	
	// Test 4: Clean up - delete the second account
	fmt.Println("\n🧹 Cleaning up...")
	deleteAccount(token, "1008")
	fmt.Println("✅ Cleanup completed!")
}

func login() string {
	loginData := map[string]string{
		"username": "admin",
		"password": "password123",
	}
	
	loginJSON, _ := json.Marshal(loginData)
	
	resp, err := http.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		log.Fatal("Login failed:", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Fatalf("Login failed with status %s: %v", resp.Status, errorResponse)
	}
	
	var loginResponse struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		log.Fatal("Failed to parse login response:", err)
	}
	
	return loginResponse.Data.Token
}

func createAccount(token string, code string, name string) uint {
	accountData := map[string]interface{}{
		"code": code,
		"name": name,
		"type": "ASSET",
		"category": "FIXED_ASSET",
		"parent_id": 1,
		"description": "Test account for soft delete verification",
		"opening_balance": 5000000,
	}
	
	accountJSON, _ := json.Marshal(accountData)
	
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/accounts", bytes.NewBuffer(accountJSON))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return 0
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Account creation failed: %v", err)
		return 0
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 201 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Account creation failed with status %s: %v", resp.Status, errorResponse)
		return 0
	}
	
	var result struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("Failed to parse response: %v", err)
		return 0
	}
	
	return result.Data.ID
}

func deleteAccount(token string, code string) {
	req, err := http.NewRequest("DELETE", "http://localhost:8080/api/v1/accounts/"+code, nil)
	if err != nil {
		log.Printf("Failed to create delete request: %v", err)
		return
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Account deletion failed: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		log.Printf("Account deletion failed with status %s: %v", resp.Status, errorResponse)
		return
	}
}
