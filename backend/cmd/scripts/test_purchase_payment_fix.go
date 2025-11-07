package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Test script to verify purchase payment fix
func main() {
	baseURL := "http://localhost:8080/api"
	
	// Get auth token first
	token, err := login()
	if err != nil {
		log.Fatal("Failed to login:", err)
	}
	
	fmt.Println("✅ Successfully logged in")
	
	// Test the purchase payment flow
	err = testPurchasePaymentFlow(baseURL, token)
	if err != nil {
		log.Fatal("Purchase payment test failed:", err)
	}
	
	fmt.Println("🎉 All tests passed! Purchase payment fix is working correctly.")
}

func login() (string, error) {
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	
	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if token, ok := result["token"].(string); ok {
		return token, nil
	}
	
	return "", fmt.Errorf("login failed: %v", result)
}

func testPurchasePaymentFlow(baseURL, token string) error {
	// Step 1: Create a credit purchase
	fmt.Println("🔄 Step 1: Creating credit purchase...")
	purchase, err := createTestCreditPurchase(baseURL, token)
	if err != nil {
		return fmt.Errorf("failed to create purchase: %v", err)
	}
	
	purchaseID := int(purchase["id"].(float64))
	fmt.Printf("✅ Created purchase ID: %d, Code: %s\n", purchaseID, purchase["code"])
	
	// Step 2: Approve the purchase
	fmt.Println("🔄 Step 2: Approving purchase...")
	err = approvePurchase(baseURL, token, purchaseID)
	if err != nil {
		return fmt.Errorf("failed to approve purchase: %v", err)
	}
	fmt.Println("✅ Purchase approved")
	
	// Step 3: Check initial outstanding amount
	fmt.Println("🔄 Step 3: Checking initial outstanding amount...")
	updatedPurchase, err := getPurchase(baseURL, token, purchaseID)
	if err != nil {
		return fmt.Errorf("failed to get purchase: %v", err)
	}
	
	totalAmount := updatedPurchase["total_amount"].(float64)
	outstandingAmount := updatedPurchase["outstanding_amount"].(float64)
	paidAmount := updatedPurchase["paid_amount"].(float64)
	
	fmt.Printf("💰 Purchase amounts - Total: %.2f, Outstanding: %.2f, Paid: %.2f\n", 
		totalAmount, outstandingAmount, paidAmount)
	
	if outstandingAmount != totalAmount {
		return fmt.Errorf("❌ Outstanding amount should equal total amount after approval. Expected: %.2f, Got: %.2f", 
			totalAmount, outstandingAmount)
	}
	
	if paidAmount != 0 {
		return fmt.Errorf("❌ Paid amount should be 0 initially. Got: %.2f", paidAmount)
	}
	
	fmt.Println("✅ Initial amounts are correct")
	
	// Step 4: Make a payment
	fmt.Println("🔄 Step 4: Making payment...")
	paymentAmount := totalAmount / 2 // Pay half
	err = makePayment(baseURL, token, purchaseID, paymentAmount)
	if err != nil {
		return fmt.Errorf("failed to make payment: %v", err)
	}
	fmt.Printf("✅ Payment of %.2f made successfully\n", paymentAmount)
	
	// Step 5: Check updated amounts after payment
	fmt.Println("🔄 Step 5: Checking amounts after payment...")
	time.Sleep(1 * time.Second) // Wait for database update
	
	updatedPurchase, err = getPurchase(baseURL, token, purchaseID)
	if err != nil {
		return fmt.Errorf("failed to get updated purchase: %v", err)
	}
	
	newOutstanding := updatedPurchase["outstanding_amount"].(float64)
	newPaid := updatedPurchase["paid_amount"].(float64)
	status := updatedPurchase["status"].(string)
	
	fmt.Printf("💰 After payment - Outstanding: %.2f, Paid: %.2f, Status: %s\n", 
		newOutstanding, newPaid, status)
	
	expectedOutstanding := totalAmount - paymentAmount
	if newOutstanding != expectedOutstanding {
		return fmt.Errorf("❌ Outstanding amount incorrect. Expected: %.2f, Got: %.2f", 
			expectedOutstanding, newOutstanding)
	}
	
	if newPaid != paymentAmount {
		return fmt.Errorf("❌ Paid amount incorrect. Expected: %.2f, Got: %.2f", 
			paymentAmount, newPaid)
	}
	
	if status != "APPROVED" {
		return fmt.Errorf("❌ Status should still be APPROVED for partial payment. Got: %s", status)
	}
	
	fmt.Println("✅ Payment amounts updated correctly")
	
	// Step 6: Make final payment
	fmt.Println("🔄 Step 6: Making final payment...")
	finalPayment := newOutstanding
	err = makePayment(baseURL, token, purchaseID, finalPayment)
	if err != nil {
		return fmt.Errorf("failed to make final payment: %v", err)
	}
	fmt.Printf("✅ Final payment of %.2f made successfully\n", finalPayment)
	
	// Step 7: Check final status
	fmt.Println("🔄 Step 7: Checking final status...")
	time.Sleep(1 * time.Second) // Wait for database update
	
	finalPurchase, err := getPurchase(baseURL, token, purchaseID)
	if err != nil {
		return fmt.Errorf("failed to get final purchase: %v", err)
	}
	
	finalOutstanding := finalPurchase["outstanding_amount"].(float64)
	finalPaid := finalPurchase["paid_amount"].(float64)
	finalStatus := finalPurchase["status"].(string)
	
	fmt.Printf("💰 Final amounts - Outstanding: %.2f, Paid: %.2f, Status: %s\n", 
		finalOutstanding, finalPaid, finalStatus)
	
	if finalOutstanding != 0 {
		return fmt.Errorf("❌ Outstanding amount should be 0 after full payment. Got: %.2f", finalOutstanding)
	}
	
	if finalPaid != totalAmount {
		return fmt.Errorf("❌ Paid amount should equal total amount. Expected: %.2f, Got: %.2f", 
			totalAmount, finalPaid)
	}
	
	if finalStatus != "PAID" {
		return fmt.Errorf("❌ Status should be PAID after full payment. Got: %s", finalStatus)
	}
	
	fmt.Println("✅ Final status is correct - purchase is fully paid")
	
	return nil
}

func createTestCreditPurchase(baseURL, token string) (map[string]interface{}, error) {
	purchaseData := map[string]interface{}{
		"vendor_id":      1,
		"date":           time.Now().Format("2006-01-02"),
		"due_date":       time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
		"payment_method": "CREDIT",
		"items": []map[string]interface{}{
			{
				"product_id":  1,
				"quantity":    2,
				"unit_price":  500000,
			},
		},
		"notes": "Test purchase for payment fix verification",
	}
	
	jsonData, _ := json.Marshal(purchaseData)
	req, _ := http.NewRequest("POST", baseURL+"/purchases", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("failed to create purchase: %v", result)
	}
	
	return result, nil
}

func approvePurchase(baseURL, token string, purchaseID int) error {
	approvalData := map[string]interface{}{
		"approved": true,
		"comments": "Test approval for payment fix verification",
	}
	
	jsonData, _ := json.Marshal(approvalData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/purchases/%d/approve-reject", baseURL, purchaseID), bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("approval failed: %v", result)
	}
	
	return nil
}

func getPurchase(baseURL, token string, purchaseID int) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/purchases/%d", baseURL, purchaseID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get purchase: %v", result)
	}
	
	return result, nil
}

func makePayment(baseURL, token string, purchaseID int, amount float64) error {
	paymentData := map[string]interface{}{
		"amount":       amount,
		"date":         time.Now().Format("2006-01-02"),
		"method":       "Bank Transfer",
		"cash_bank_id": 1,
		"reference":    fmt.Sprintf("TEST-PAY-%d", purchaseID),
		"notes":        "Test payment for fix verification",
	}
	
	jsonData, _ := json.Marshal(paymentData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/purchases/%d/payments", baseURL, purchaseID), bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second} // Payment operations might take longer
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if resp.StatusCode != 201 {
		return fmt.Errorf("payment failed: %v", result)
	}
	
	success, _ := result["success"].(bool)
	if !success {
		return fmt.Errorf("payment not successful: %v", result)
	}
	
	return nil
}