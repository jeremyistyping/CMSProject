package main

import (
	"fmt"
	"log"
	"net/http"
	"bytes"
	"encoding/json"
	"time"
	"strings"
)

func main() {
	fmt.Println("🧪 Testing Sales Confirm API with SSOT Journal Integration")
	fmt.Println("=" + strings.Repeat("=", 60))
	
	// Test confirm sale endpoint
	saleID := 7 // Sale ID yang gagal sebelumnya
	
	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Prepare confirm sale request
	url := fmt.Sprintf("http://localhost:8080/api/v1/sales/%d/confirm", saleID)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	
	// Add authentication header (you may need to adjust this)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer your-token-here") // Replace with actual token
	
	fmt.Printf("🚀 Sending POST request to: %s\n", url)
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}
	
	fmt.Printf("📊 Response Status: %s\n", resp.Status)
	fmt.Printf("📝 Response Body: %+v\n", response)
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ Sales confirm test PASSED - SSOT journal integration working!")
	} else {
		fmt.Printf("❌ Sales confirm test FAILED - Status: %d\n", resp.StatusCode)
		if errorMsg, exists := response["error"]; exists {
			fmt.Printf("Error details: %s\n", errorMsg)
		}
	}
}