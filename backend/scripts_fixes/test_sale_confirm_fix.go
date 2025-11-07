package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	baseURL = "http://localhost:8080"
	authToken = "your_jwt_token_here"
)

// Sale creation request
type SaleCreateRequest struct {
	CustomerID        uint                      `json:"customer_id"`
	Type              string                    `json:"type"`
	Date              time.Time                 `json:"date"`
	PaymentMethodType string                    `json:"payment_method_type"`
	CashBankID        *uint                     `json:"cash_bank_id"`
	Items             []SaleItemCreateRequest   `json:"items"`
}

type SaleItemCreateRequest struct {
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Taxable   bool    `json:"taxable"`
}

// Response structures
type SaleResponse struct {
	ID                uint     `json:"id"`
	Code             string   `json:"code"`
	Status           string   `json:"status"`
	TotalAmount      float64  `json:"total_amount"`
	PaidAmount       float64  `json:"paid_amount"`
	OutstandingAmount float64 `json:"outstanding_amount"`
	PaymentMethodType string  `json:"payment_method_type"`
}

type Account struct {
	ID      uint    `json:"id"`
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
	Type    string  `json:"type"`
}

type SalePayment struct {
	ID            uint      `json:"id"`
	SaleID        uint      `json:"sale_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
	PaymentDate   time.Time `json:"payment_date"`
	Reference     string    `json:"reference"`
}

func makeRequest(method, url string, body interface{}, useAuth bool) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if useAuth {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func getAccountBalance(accountID uint) (float64, error) {
	url := fmt.Sprintf("%s/accounts/%d", baseURL, accountID)
	respBody, err := makeRequest("GET", url, nil, true)
	if err != nil {
		return 0, err
	}

	var account Account
	if err := json.Unmarshal(respBody, &account); err != nil {
		return 0, fmt.Errorf("failed to parse account response: %v", err)
	}

	return account.Balance, nil
}

func getSalePayments(saleID uint) ([]SalePayment, error) {
	url := fmt.Sprintf("%s/sales/%d/payments", baseURL, saleID)
	respBody, err := makeRequest("GET", url, nil, true)
	if err != nil {
		return nil, err
	}

	var payments []SalePayment
	if err := json.Unmarshal(respBody, &payments); err != nil {
		return nil, fmt.Errorf("failed to parse payments response: %v", err)
	}

	return payments, nil
}

func main() {
	log.Printf("🧪 Testing Sale Confirmation Fix for Cash/Bank Payments")
	log.Printf("================================================")

	// Test configuration
	customerID := uint(1) // Adjust as needed
	cashBankID := uint(1) // Adjust as needed - should be a valid cash/bank account
	productID := uint(1)  // Adjust as needed - should be a product with stock

	// Step 1: Get initial account balances
	log.Printf("\n📊 Step 1: Getting initial account balances...")
	initialCashBalance, err := getAccountBalance(cashBankID)
	if err != nil {
		log.Printf("⚠️  Warning: Could not get initial cash balance: %v", err)
		initialCashBalance = 0
	} else {
		log.Printf("💰 Initial Cash/Bank balance: %.2f", initialCashBalance)
	}

	// Step 2: Create a cash sale (DRAFT status)
	log.Printf("\n📝 Step 2: Creating cash sale...")
	saleRequest := SaleCreateRequest{
		CustomerID:        customerID,
		Type:              "INVOICE",
		Date:              time.Now(),
		PaymentMethodType: "CASH",
		CashBankID:        &cashBankID,
		Items: []SaleItemCreateRequest{
			{
				ProductID: productID,
				Quantity:  2,
				UnitPrice: 50000,
				Taxable:   true,
			},
		},
	}

	respBody, err := makeRequest("POST", baseURL+"/sales", saleRequest, true)
	if err != nil {
		log.Fatalf("❌ Failed to create sale: %v", err)
	}

	var sale SaleResponse
	if err := json.Unmarshal(respBody, &sale); err != nil {
		log.Fatalf("❌ Failed to parse sale response: %v", err)
	}

	log.Printf("✅ Sale created: ID=%d, Code=%s, Status=%s", sale.ID, sale.Code, sale.Status)
	log.Printf("   💰 Total: %.2f, Paid: %.2f, Outstanding: %.2f", 
		sale.TotalAmount, sale.PaidAmount, sale.OutstandingAmount)

	if sale.Status != "DRAFT" {
		log.Printf("⚠️  Expected status DRAFT, got %s", sale.Status)
	}

	// Step 3: Confirm the sale
	log.Printf("\n🔄 Step 3: Confirming sale...")
	confirmURL := fmt.Sprintf("%s/sales/%d/confirm", baseURL, sale.ID)
	_, err = makeRequest("POST", confirmURL, nil, true)
	if err != nil {
		log.Fatalf("❌ Failed to confirm sale: %v", err)
	}

	// Step 4: Get updated sale details
	log.Printf("\n📋 Step 4: Getting updated sale details...")
	saleURL := fmt.Sprintf("%s/sales/%d", baseURL, sale.ID)
	respBody, err = makeRequest("GET", saleURL, nil, true)
	if err != nil {
		log.Fatalf("❌ Failed to get updated sale: %v", err)
	}

	var updatedSale SaleResponse
	if err := json.Unmarshal(respBody, &updatedSale); err != nil {
		log.Fatalf("❌ Failed to parse updated sale response: %v", err)
	}

	log.Printf("✅ Updated sale: ID=%d, Status=%s", updatedSale.ID, updatedSale.Status)
	log.Printf("   💰 Total: %.2f, Paid: %.2f, Outstanding: %.2f", 
		updatedSale.TotalAmount, updatedSale.PaidAmount, updatedSale.OutstandingAmount)

	// Step 5: Verify status is PAID for cash sale
	log.Printf("\n🔍 Step 5: Verifying sale status and amounts...")
	if updatedSale.Status == "PAID" {
		log.Printf("✅ ✅ Sale status is correctly PAID")
	} else {
		log.Printf("❌ ❌ Sale status should be PAID but is %s", updatedSale.Status)
	}

	if updatedSale.PaidAmount == updatedSale.TotalAmount {
		log.Printf("✅ ✅ PaidAmount matches TotalAmount (%.2f)", updatedSale.PaidAmount)
	} else {
		log.Printf("❌ ❌ PaidAmount (%.2f) should equal TotalAmount (%.2f)", 
			updatedSale.PaidAmount, updatedSale.TotalAmount)
	}

	if updatedSale.OutstandingAmount == 0 {
		log.Printf("✅ ✅ OutstandingAmount is correctly 0")
	} else {
		log.Printf("❌ ❌ OutstandingAmount should be 0 but is %.2f", updatedSale.OutstandingAmount)
	}

	// Step 6: Check sale payments
	log.Printf("\n💳 Step 6: Checking sale payments...")
	payments, err := getSalePayments(sale.ID)
	if err != nil {
		log.Printf("⚠️  Warning: Could not get sale payments: %v", err)
	} else {
		log.Printf("💰 Found %d payment(s) for sale %d", len(payments), sale.ID)
		for _, payment := range payments {
			log.Printf("   - Payment ID=%d, Amount=%.2f, Method=%s, Status=%s, Ref=%s", 
				payment.ID, payment.Amount, payment.PaymentMethod, payment.Status, payment.Reference)
		}
	}

	// Step 7: Check updated account balances
	log.Printf("\n📊 Step 7: Checking updated account balances...")
	finalCashBalance, err := getAccountBalance(cashBankID)
	if err != nil {
		log.Printf("⚠️  Warning: Could not get final cash balance: %v", err)
	} else {
		log.Printf("💰 Final Cash/Bank balance: %.2f", finalCashBalance)
		balanceIncrease := finalCashBalance - initialCashBalance
		log.Printf("📈 Balance increase: %.2f", balanceIncrease)
		
		if balanceIncrease == updatedSale.TotalAmount {
			log.Printf("✅ ✅ Cash balance increased by sale amount")
		} else {
			log.Printf("⚠️  Cash balance increase (%.2f) doesn't match sale amount (%.2f)", 
				balanceIncrease, updatedSale.TotalAmount)
		}
	}

	// Summary
	log.Printf("\n🎯 SUMMARY")
	log.Printf("=========")
	log.Printf("Sale ID: %d", sale.ID)
	log.Printf("Sale Code: %s", updatedSale.Code)
	log.Printf("Final Status: %s", updatedSale.Status)
	log.Printf("Payment Method: %s", updatedSale.PaymentMethodType)
	log.Printf("Total Amount: %.2f", updatedSale.TotalAmount)
	log.Printf("Paid Amount: %.2f", updatedSale.PaidAmount)
	log.Printf("Outstanding: %.2f", updatedSale.OutstandingAmount)
	log.Printf("Number of Payments: %d", len(payments))

	// Test results
	allPassed := true
	if updatedSale.Status != "PAID" {
		allPassed = false
	}
	if updatedSale.PaidAmount != updatedSale.TotalAmount {
		allPassed = false
	}
	if updatedSale.OutstandingAmount != 0 {
		allPassed = false
	}
	if len(payments) == 0 {
		allPassed = false
	}

	if allPassed {
		log.Printf("\n🎉 🎉 ALL TESTS PASSED! Sale confirmation fix is working correctly.")
	} else {
		log.Printf("\n❌ ❌ SOME TESTS FAILED. Please check the issues above.")
	}
}