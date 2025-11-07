package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"log"
	"time"
)

func main() {
	// Test data - sesuaikan dengan product yang ada
	productID := "23" // ID dari hasil script sebelumnya
	baseURL := "http://localhost:8080"
	
	fmt.Println("=== TESTING PRODUCT API ENDPOINTS ===")
	
	// Step 1: Get current product data
	fmt.Println("\n1. Getting current product data...")
	getCurrentProduct(baseURL, productID)
	
	// Step 2: Test update product with stock change
	fmt.Println("\n2. Testing update product with stock change...")
	testUpdateProduct(baseURL, productID)
	
	// Step 3: Verify the update
	fmt.Println("\n3. Verifying the update...")
	getCurrentProduct(baseURL, productID)
}

func getCurrentProduct(baseURL, productID string) {
	url := fmt.Sprintf("%s/api/v1/products/%s", baseURL, productID)
	
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("❌ Error getting product: %v", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Error reading response: %v", err)
		return
	}
	
	fmt.Printf("Status: %d\n", resp.StatusCode)
	
	if resp.StatusCode == 200 {
		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err == nil {
			if data, ok := response["data"].(map[string]interface{}); ok {
				fmt.Printf("✅ Current Product:\n")
				fmt.Printf("   ID: %.0f\n", data["id"])
				fmt.Printf("   Code: %s\n", data["code"])
				fmt.Printf("   Name: %s\n", data["name"])
				fmt.Printf("   Stock: %.0f\n", data["stock"])
				fmt.Printf("   Purchase Price: %.2f\n", data["purchase_price"])
				fmt.Printf("   Sale Price: %.2f\n", data["sale_price"])
			}
		}
	} else {
		fmt.Printf("❌ Response: %s\n", string(body))
	}
}

func testUpdateProduct(baseURL, productID string) {
	url := fmt.Sprintf("%s/api/v1/products/%s", baseURL, productID)
	
	// Prepare update data - mengubah stock dari yang sekarang
	updateData := map[string]interface{}{
		"code":            "PRN-CAN-2025-001",
		"name":            "Mesin Printer",
		"stock":           50, // Update stock ke 50
		"purchase_price":  3521739.13,
		"sale_price":      3500000.0,
		"category_id":     1,
		"warehouse_location_id": 1,
		"unit":            "pcs",
		"is_active":       true,
		"is_service":      false,
		"taxable":         true,
	}
	
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		log.Printf("❌ Error marshaling JSON: %v", err)
		return
	}
	
	fmt.Printf("Sending update request with data:\n%s\n", string(jsonData))
	
	// Create PUT request
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Error creating request: %v", err)
		return
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	// Note: Untuk production perlu authentication token
	// req.Header.Set("Authorization", "Bearer YOUR_TOKEN")
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Error reading response: %v", err)
		return
	}
	
	fmt.Printf("Update Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Update Response Body: %s\n", string(body))
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ Product update successful!")
	} else {
		fmt.Printf("❌ Product update failed with status: %d\n", resp.StatusCode)
	}
}