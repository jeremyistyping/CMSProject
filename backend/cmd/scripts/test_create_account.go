package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"log"
)

func main() {
	// First, login to get a token
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
		log.Fatal("Login failed with status:", resp.Status)
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
	
	token := loginResponse.Data.Token
	fmt.Printf("Login successful, token: %s...\n", token[:20])
	
	// Now create an account with code 1008
	accountData := map[string]interface{}{
		"code": "1008",
		"name": "MOTOR",
		"type": "ASSET",
		"category": "FIXED_ASSET",
		"parent_id": 1,
		"description": "Kendaraan bermotor",
		"opening_balance": 5000000,
	}
	
	accountJSON, _ := json.Marshal(accountData)
	
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/accounts", bytes.NewBuffer(accountJSON))
	if err != nil {
		log.Fatal("Failed to create request:", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal("Account creation failed:", err)
	}
	defer resp.Body.Close()
	
	fmt.Printf("Account creation response status: %s\n", resp.Status)
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal("Failed to parse response:", err)
	}
	
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Response body:\n%s\n", resultJSON)
	
	if resp.StatusCode == 201 {
		fmt.Println("✅ Account created successfully!")
	} else {
		fmt.Println("❌ Account creation failed!")
	}
}
