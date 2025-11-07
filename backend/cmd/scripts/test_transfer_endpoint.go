package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TransferRequest represents the request format for transfer
type TransferRequest struct {
	FromAccountID uint      `json:"from_account_id"`
	ToAccountID   uint      `json:"to_account_id"`
	Amount        float64   `json:"amount"`
	Date          time.Time `json:"date"`
	Reference     string    `json:"reference"`
	Notes         string    `json:"notes"`
	ExchangeRate  float64   `json:"exchange_rate"`
}

func main() {
	fmt.Println("=== TESTING TRANSFER ENDPOINT ===")
	
	// Test data
	transferReq := TransferRequest{
		FromAccountID: 1, // BCA
		ToAccountID:   4, // UOB
		Amount:        1000000,
		Date:          time.Now(),
		Reference:     "TEST-TRANSFER-001",
		Notes:         "Test transfer endpoint verification",
		ExchangeRate:  1.0,
	}
	
	// Convert to JSON
	jsonData, err := json.Marshal(transferReq)
	if err != nil {
		fmt.Printf("❌ Failed to marshal JSON: %v\n", err)
		return
	}
	
	fmt.Printf("📋 Testing with data: %s\n", string(jsonData))
	
	// Test endpoints
	endpoints := []string{
		"http://localhost:8080/api/v1/cashbank/transfer",
		"http://localhost:8080/api/v1/cash-bank/transactions/transfer",
	}
	
	for i, endpoint := range endpoints {
		fmt.Printf("\n🔄 Testing endpoint %d: %s\n", i+1, endpoint)
		
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("❌ Failed to create request: %v\n", err)
			continue
		}
		
		req.Header.Set("Content-Type", "application/json")
		// Note: In real testing, you would need to add Authorization header
		// req.Header.Set("Authorization", "Bearer YOUR_TOKEN")
		
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Failed to send request: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ Failed to read response: %v\n", err)
			continue
		}
		
		fmt.Printf("📊 Status Code: %d\n", resp.StatusCode)
		fmt.Printf("📄 Response: %s\n", string(body))
		
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			fmt.Printf("✅ Endpoint %s is working!\n", endpoint)
		} else if resp.StatusCode == 401 {
			fmt.Printf("🔐 Endpoint %s requires authentication (normal)\n", endpoint)
		} else {
			fmt.Printf("⚠️  Endpoint %s returned status %d\n", endpoint, resp.StatusCode)
		}
	}
	
	fmt.Println("\n=== ENDPOINT TEST COMPLETE ===")
	fmt.Println("📝 Notes:")
	fmt.Println("- If you get 401 (Unauthorized), that's normal - add Bearer token for real testing")
	fmt.Println("- If you get 404 (Not Found), the endpoint is not registered")
	fmt.Println("- If you get 400 (Bad Request), the endpoint exists but data format might be wrong")
	fmt.Println("- If you get connection error, make sure the server is running on port 8080")
}