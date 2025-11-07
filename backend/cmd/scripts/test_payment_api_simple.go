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

func main() {
	log.Println("Testing Payment API...")

	// Test data for payment creation
	testPaymentAPI()
}

func testPaymentAPI() {
	// 1. Test with sale ID from the screenshot (INV-2025-3637)
	saleID := "21" // Based on the screenshot showing this sale
	
	// 2. Create payment request (based on screenshot data)
	paymentData := map[string]interface{}{
		"amount":      4440000.0,  // Full payment amount
		"date":        "2025-09-15T00:00:00Z",
		"method":      "Cash",
		"cash_bank_id": 1, // Assuming first cash bank account
		"reference":   "test",
		"notes":       "test",
	}

	jsonData, err := json.Marshal(paymentData)
	if err != nil {
		log.Fatalf("Error marshaling payment data: %v", err)
	}

	// 3. Test the integrated payment endpoint
	url := fmt.Sprintf("http://localhost:8080/api/v1/sales/%s/integrated-payment", saleID)
	
	log.Printf("Testing Payment API:")
	log.Printf("URL: %s", url)
	log.Printf("Payload: %s", string(jsonData))

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Note: In real scenario, you'd need to add Authorization header with valid JWT token
	// req.Header.Set("Authorization", "Bearer YOUR_TOKEN_HERE")

	// Make the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Error reading response: %v", err)
		return
	}

	log.Printf("Response Status: %d", resp.StatusCode)
	log.Printf("Response Body: %s", string(body))

	// Check response status
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		log.Println("✅ Payment API responded successfully")
	} else if resp.StatusCode == 401 {
		log.Println("🔐 Authentication required - this is expected without token")
	} else {
		log.Printf("❌ Payment API failed with status %d", resp.StatusCode)
		
		// Try to parse error response
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if details, ok := errorResp["details"]; ok {
				log.Printf("Error details: %v", details)
			}
			if errorMsg, ok := errorResp["error"]; ok {
				log.Printf("Error message: %v", errorMsg)
			}
		}
	}
}