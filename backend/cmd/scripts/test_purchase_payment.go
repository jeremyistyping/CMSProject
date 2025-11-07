package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Test configuration
const (
	BaseURL = "http://localhost:8080/api/v1"
	TestEmail = "admin@example.com" 
	TestPassword = "password123"
)

// Login response structure
type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
}

// Payment request structure
type PaymentRequest struct {
	Amount     float64   `json:"amount"`
	Date       time.Time `json:"date"`
	Method     string    `json:"method"`
	CashBankID uint      `json:"cash_bank_id"`
	Reference  string    `json:"reference"`
	Notes      string    `json:"notes"`
}

func main() {
	fmt.Println("🧪 Testing Purchase Payment Management Integration...")
	
	// Step 1: Login to get JWT token
	token, err := login()
	if err != nil {
		log.Fatalf("❌ Failed to login: %v", err)
	}
	fmt.Println("✅ Successfully logged in")
	
	// Step 2: Test getting purchase payments
	testGetPurchasePayments(token)
	
	// Step 3: Test creating a payment for an existing purchase (if any)
	testCreatePurchasePayment(token)
}

func login() (string, error) {
	loginData := map[string]string{
		"email":    TestEmail,
		"password": TestPassword,
	}
	
	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(BaseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}
	
	return loginResp.Token, nil
}

func testGetPurchasePayments(token string) {
	fmt.Println("\n📋 Testing GET /purchases/1/payments...")
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", BaseURL+"/purchases/1/payments", nil)
	if err != nil {
		log.Printf("❌ Failed to create request: %v", err)
		return
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Request failed: %v", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ Successfully retrieved purchase payments (Status: %d)\n", resp.StatusCode)
		fmt.Printf("📄 Response: %s\n", string(body))
	} else {
		fmt.Printf("⚠️  Get payments returned status %d: %s\n", resp.StatusCode, string(body))
	}
}

func testCreatePurchasePayment(token string) {
	fmt.Println("\n💰 Testing POST /purchases/1/payments...")
	
	// Create a test payment request
	paymentReq := PaymentRequest{
		Amount:     100000.0, // 100k
		Date:       time.Now(),
		Method:     "BANK_TRANSFER",
		CashBankID: 1, // Assuming cash bank ID 1 exists
		Reference:  "TEST-PAY-001",
		Notes:      "Test payment via Payment Management integration",
	}
	
	jsonData, _ := json.Marshal(paymentReq)
	
	client := &http.Client{}
	req, err := http.NewRequest("POST", BaseURL+"/purchases/1/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to create request: %v", err)
		return
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Request failed: %v", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Successfully created purchase payment (Status: %d)\n", resp.StatusCode)
		fmt.Printf("💳 Response: %s\n", string(body))
	} else {
		fmt.Printf("⚠️  Create payment returned status %d: %s\n", resp.StatusCode, string(body))
	}
}