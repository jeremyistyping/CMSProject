package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== PAYMENT TEST WITH AUTHENTICATION ===")
	
	// Connect to database
	db := database.ConnectDB()
	
	// Step 1: Get admin user for auth
	fmt.Println("\n🔐 Step 1: Getting admin user for authentication...")
	var adminUser models.User
	if err := db.Where("role = ? AND is_active = ?", "admin", true).First(&adminUser).Error; err != nil {
		log.Printf("No admin user found, trying any active user: %v", err)
		if err := db.Where("is_active = ?", true).First(&adminUser).Error; err != nil {
			log.Fatal("No active user found for authentication:", err)
		}
	}
	
	fmt.Printf("Using user: %s (ID: %d, Role: %s)\n", adminUser.Username, adminUser.ID, adminUser.Role)
	
	// Step 2: Try to get or create a valid JWT token
	fmt.Println("\n🎫 Step 2: Creating authentication token...")
	
	// For this test, we'll try to login first to get a valid token
	loginData := map[string]string{
		"username": adminUser.Username,
		"password": "admin123", // Default password - you might need to adjust this
	}
	
	jsonLoginData, err := json.Marshal(loginData)
	if err != nil {
		log.Fatal("Failed to marshal login JSON:", err)
	}
	
	// Try login
	client := &http.Client{Timeout: 30 * time.Second}
	loginReq, err := http.NewRequest("POST", "http://localhost:8080/api/auth/login", bytes.NewBuffer(jsonLoginData))
	if err != nil {
		log.Fatal("Failed to create login request:", err)
	}
	
	loginReq.Header.Set("Content-Type", "application/json")
	
	fmt.Println("Attempting login...")
	loginResp, err := client.Do(loginReq)
	if err != nil {
		log.Printf("Login failed: %v", err)
		fmt.Println("\n⚠️ Will try payment without authentication...")
		testPaymentWithoutAuth(client)
		return
	}
	
	defer loginResp.Body.Close()
	loginBody, err := io.ReadAll(loginResp.Body)
	if err != nil {
		log.Printf("Failed to read login response: %v", err)
		return
	}
	
	fmt.Printf("Login response status: %s\n", loginResp.Status)
	fmt.Printf("Login response: %s\n", string(loginBody))
	
	// Parse login response to get token
	var loginResult map[string]interface{}
	if err := json.Unmarshal(loginBody, &loginResult); err != nil {
		log.Printf("Failed to parse login response: %v", err)
		fmt.Println("Will try payment without authentication...")
		testPaymentWithoutAuth(client)
		return
	}
	
	// Get token from response
	var token string
	if tokenInterface, ok := loginResult["token"]; ok {
		if tokenStr, ok := tokenInterface.(string); ok {
			token = tokenStr
			fmt.Printf("✅ Got auth token: %s...\n", token[:20])
		}
	} else if accessTokenInterface, ok := loginResult["access_token"]; ok {
		if tokenStr, ok := accessTokenInterface.(string); ok {
			token = tokenStr
			fmt.Printf("✅ Got access token: %s...\n", token[:20])
		}
	}
	
	if token == "" {
		fmt.Println("❌ No token found in login response")
		fmt.Println("Will try payment without authentication...")
		testPaymentWithoutAuth(client)
		return
	}
	
	// Step 3: Test payment with authentication
	fmt.Println("\n🚀 Step 3: Testing payment with authentication...")
	testPaymentWithAuth(client, token)
	
	// Step 4: Verify results
	fmt.Println("\n🔍 Step 4: Verifying payment results...")
	verifyPaymentResults(db)
}

func testPaymentWithAuth(client *http.Client, token string) {
	// Prepare payment data
	paymentData := map[string]interface{}{
		"amount":       3330000.0,
		"date":         time.Now().Format("2006-01-02T15:04:05Z07:00"),
		"method":       "BANK_TRANSFER",
		"cash_bank_id": 1,
		"reference":    "test-payment-with-auth",
		"notes":        "Test payment with proper authentication",
	}
	
	jsonData, err := json.Marshal(paymentData)
	if err != nil {
		log.Printf("Failed to marshal payment JSON: %v", err)
		return
	}
	
	// Make authenticated request
	url := "http://localhost:8080/api/sales/1/integrated-payment"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create payment request: %v", err)
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	fmt.Printf("Making authenticated POST request to: %s\n", url)
	fmt.Printf("Payload: %s\n", string(jsonData))
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Payment request failed: %v", err)
		return
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read payment response: %v", err)
		return
	}
	
	fmt.Printf("Payment response status: %s\n", resp.Status)
	fmt.Printf("Payment response: %s\n", string(body))
	
	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Payment creation successful!\n")
		
		// Parse success response
		var paymentResp map[string]interface{}
		if err := json.Unmarshal(body, &paymentResp); err == nil {
			if payment, ok := paymentResp["payment"]; ok {
				if paymentMap, ok := payment.(map[string]interface{}); ok {
					if paymentID, ok := paymentMap["id"]; ok {
						fmt.Printf("   Payment ID: %.0f\n", paymentID)
					}
					if code, ok := paymentMap["code"]; ok {
						fmt.Printf("   Payment Code: %s\n", code)
					}
				}
			}
			if updatedSale, ok := paymentResp["updated_sale"]; ok {
				if saleMap, ok := updatedSale.(map[string]interface{}); ok {
					if paidAmount, ok := saleMap["paid_amount"]; ok {
						fmt.Printf("   Updated Paid Amount: %.2f\n", paidAmount)
					}
					if outstanding, ok := saleMap["outstanding_amount"]; ok {
						fmt.Printf("   Updated Outstanding: %.2f\n", outstanding)
					}
				}
			}
		}
	} else {
		fmt.Printf("❌ Payment creation failed!\n")
		
		// Parse error response
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if details, ok := errorResp["details"]; ok {
				fmt.Printf("Error details: %v\n", details)
			}
			if expectedFields, ok := errorResp["expected_fields"]; ok {
				fmt.Printf("Expected fields: %v\n", expectedFields)
			}
		}
	}
}

func testPaymentWithoutAuth(client *http.Client) {
	fmt.Println("\n🚀 Testing payment without authentication...")
	
	paymentData := map[string]interface{}{
		"amount":       2220000.0,
		"date":         time.Now().Format("2006-01-02T15:04:05Z07:00"),
		"method":       "BANK_TRANSFER",
		"cash_bank_id": 1,
		"reference":    "test-payment-no-auth",
		"notes":        "Test payment without authentication",
	}
	
	jsonData, err := json.Marshal(paymentData)
	if err != nil {
		log.Printf("Failed to marshal payment JSON: %v", err)
		return
	}
	
	url := "http://localhost:8080/api/sales/1/integrated-payment"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create payment request: %v", err)
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Payment request failed: %v", err)
		return
	}
	
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		return
	}
	
	fmt.Printf("Response status: %s\n", resp.Status)
	fmt.Printf("Response: %s\n", string(body))
}

func verifyPaymentResults(db *gorm.DB) {
	// Check if payment was created
	var payments []models.Payment
	if err := db.Where("contact_id = ?", 37).Order("created_at DESC").Limit(5).Find(&payments).Error; err != nil {
		log.Printf("Error getting payments: %v", err)
		return
	}
	
	fmt.Printf("Recent payments for customer 37:\n")
	for i, payment := range payments {
		fmt.Printf("  %d. ID=%d, Code=%s, Amount=%.2f, Status=%s, Date=%s\n", 
			i+1, payment.ID, payment.Code, payment.Amount, payment.Status, 
			payment.Date.Format("2006-01-02 15:04:05"))
	}
	
	// Check sale status
	var sale models.Sale
	if err := db.First(&sale, 1).Error; err != nil {
		log.Printf("Error getting sale: %v", err)
		return
	}
	
	fmt.Printf("\nSale ID 1 status:\n")
	fmt.Printf("  Paid Amount: %.2f\n", sale.PaidAmount)
	fmt.Printf("  Outstanding: %.2f\n", sale.OutstandingAmount)
	fmt.Printf("  Status: %s\n", sale.Status)
	
	// Check for journal entries
	var journalEntries []models.JournalEntry
	if err := db.Where("reference_type = ?", models.JournalRefPayment).Order("created_at DESC").Limit(3).Find(&journalEntries).Error; err != nil {
		log.Printf("Error getting journal entries: %v", err)
		return
	}
	
	fmt.Printf("\nRecent payment journal entries:\n")
	for i, journal := range journalEntries {
		fmt.Printf("  %d. ID=%d, Code=%s, Amount=%.2f, Status=%s\n", 
			i+1, journal.ID, journal.Code, journal.TotalDebit, journal.Status)
	}
}