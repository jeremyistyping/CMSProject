package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"log"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

type DepositRequest struct {
	AccountID uint   `json:"account_id"`
	Amount    float64 `json:"amount"`
	Date      string `json:"date"`
	Reference string `json:"reference"`
	Notes     string `json:"notes"`
}

type DepositResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func main() {
	baseURL := "http://localhost:8080/api/v1"
	
	// Test login first
	log.Println("🔐 Step 1: Logging in as admin...")
	token, err := login(baseURL)
	if err != nil {
		log.Fatalf("❌ Login failed: %v", err)
	}
	log.Printf("✅ Login successful, got token: %s...", token[:10])
	
	// Test deposit
	log.Println("💰 Step 2: Testing deposit functionality...")
	depositReq := DepositRequest{
		AccountID: 1,  // Assuming cash account ID is 1
		Amount:    1000.00,
		Date:      time.Now().Format("2006-01-02"),
		Reference: "TEST-DEPOSIT-001",
		Notes:     "Test deposit for instant processing verification",
	}
	
	start := time.Now()
	success := testDeposit(baseURL, token, depositReq)
	duration := time.Since(start)
	
	if success {
		log.Printf("🎉 Deposit test PASSED in %v", duration)
		log.Println("✅ Deposit processed instantly without timeout!")
	} else {
		log.Printf("❌ Deposit test FAILED after %v", duration)
		os.Exit(1)
	}
}

func login(baseURL string) (string, error) {
	loginReq := LoginRequest{
		Email:    "admin@admin.com",
		Password: "admin123",
	}
	
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login request: %v", err)
	}
	
	resp, err := http.Post(baseURL+"/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read login response: %v", err)
	}
	
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal login response: %v", err)
	}
	
	if !loginResp.Success {
		return "", fmt.Errorf("login failed: %s", loginResp.Message)
	}
	
	return loginResp.Token, nil
}

func testDeposit(baseURL, token string, depositReq DepositRequest) bool {
	jsonData, err := json.Marshal(depositReq)
	if err != nil {
		log.Printf("❌ Failed to marshal deposit request: %v", err)
		return false
	}
	
	log.Printf("📤 Sending deposit request: AccountID=%d, Amount=%.2f", 
		depositReq.AccountID, depositReq.Amount)
	
	client := &http.Client{
		Timeout: 30 * time.Second,  // 30 second timeout
	}
	
	req, err := http.NewRequest("POST", baseURL+"/cash-bank/deposit", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to create deposit request: %v", err)
		return false
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		log.Printf("❌ Deposit request failed after %v: %v", duration, err)
		return false
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read deposit response: %v", err)
		return false
	}
	
	log.Printf("📥 Got response in %v (Status: %d)", duration, resp.StatusCode)
	
	var depositResp DepositResponse
	if err := json.Unmarshal(body, &depositResp); err != nil {
		log.Printf("❌ Failed to unmarshal deposit response: %v", err)
		log.Printf("Response body: %s", string(body))
		return false
	}
	
	if !depositResp.Success {
		log.Printf("❌ Deposit failed: %s", depositResp.Message)
		return false
	}
	
	log.Printf("✅ Deposit successful: %s", depositResp.Message)
	return true
}