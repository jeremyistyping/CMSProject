package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PaymentRequest struct {
	Amount     float64 `json:"amount"`
	Date       string  `json:"date"`
	Method     string  `json:"method"`
	CashBankID uint    `json:"cash_bank_id"`
	Reference  string  `json:"reference"`
	Notes      string  `json:"notes"`
}

func main() {
	fmt.Println("🧪 Make Test Payment for Purchase Integration")
	fmt.Println("============================================")

	// Test payment data
	paymentReq := PaymentRequest{
		Amount:     1000000.0, // Rp 1.000.000 test payment
		Date:       time.Now().Format("2006-01-02"),
		Method:     "Bank Transfer",
		CashBankID: 6, // Assuming this is a valid cash bank ID (Bank Mandiri)
		Reference:  "TEST-PAYMENT",
		Notes:      "Test payment to verify SSOT integration",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(paymentReq)
	if err != nil {
		fmt.Printf("❌ Failed to marshal payment request: %v\n", err)
		return
	}

	// Test endpoint (assuming purchase ID 1 based on previous analysis)
	url := "http://localhost:8080/api/v1/purchases/1/payments"
	
	fmt.Printf("📋 Test Payment Request:\n")
	fmt.Printf("  URL: %s\n", url)
	fmt.Printf("  Amount: Rp %.2f\n", paymentReq.Amount)
	fmt.Printf("  Method: %s\n", paymentReq.Method)
	fmt.Printf("  Bank ID: %d\n", paymentReq.CashBankID)
	
	fmt.Printf("\n💡 To test manually:\n")
	fmt.Printf("1. Make sure backend server is running on port 8080\n")
	fmt.Printf("2. You need valid JWT token for authentication\n")
	fmt.Printf("3. Use the frontend to make payment, or\n")
	fmt.Printf("4. Use curl/Postman with proper authentication\n")
	
	fmt.Printf("\n🔍 Expected behavior:\n")
	fmt.Printf("1. Payment record created: PAY/2025/09/XXXXX\n")
	fmt.Printf("2. SSOT journal entry created with source_type = 'PAYMENT'\n")
	fmt.Printf("3. Journal lines:\n")
	fmt.Printf("   - Dr. Accounts Payable (2101): Rp 1.000.000\n")
	fmt.Printf("   - Cr. Bank Mandiri (1103): Rp 1.000.000\n")
	fmt.Printf("4. Account balances updated:\n")
	fmt.Printf("   - Utang Usaha reduces by Rp 1.000.000\n")
	fmt.Printf("   - Bank Mandiri reduces by Rp 1.000.000\n")

	fmt.Printf("\n🧪 Manual Test Instructions:\n")
	fmt.Printf("1. Go to http://localhost:3000/purchases\n")
	fmt.Printf("2. Find purchase PO/2025/09/0005 with outstanding Rp 5.550.000\n")
	fmt.Printf("3. Click 'Record Payment' button\n")
	fmt.Printf("4. Enter amount: 1000000 (Rp 1.000.000)\n")
	fmt.Printf("5. Select Bank Mandiri\n")
	fmt.Printf("6. Click Submit\n")
	fmt.Printf("7. Check if success message appears\n")
	fmt.Printf("8. Run verification script to check SSOT integration\n")

	// Prepare curl command for manual testing
	fmt.Printf("\n🔧 Curl Command (replace JWT_TOKEN):\n")
	curlCmd := fmt.Sprintf(`curl -X POST "%s" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer JWT_TOKEN" \
  -d '%s'`, url, string(jsonData))
	
	fmt.Printf("%s\n", curlCmd)

	fmt.Printf("\n✅ After making payment, run this to verify:\n")
	fmt.Printf("go run cmd/scripts/verify_payment_accounting.go\n")

	// Try to make actual request (will likely fail due to auth)
	fmt.Printf("\n🔍 Attempting test request (will likely fail without auth)...\n")
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Request failed (expected): %v\n", err)
		fmt.Printf("   This is normal - authentication required\n")
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📊 Response Status: %s\n", resp.Status)
	
	if resp.StatusCode == 401 {
		fmt.Printf("✅ Expected: Authentication required (401)\n")
		fmt.Printf("   Use frontend or provide valid JWT token\n")
	} else {
		fmt.Printf("Response status: %d\n", resp.StatusCode)
	}

	fmt.Printf("\n🎯 Next Steps:\n")
	fmt.Printf("1. Use frontend to make the payment\n")
	fmt.Printf("2. Check backend logs for SSOT journal creation\n")
	fmt.Printf("3. Run verification script to confirm integration\n")
}