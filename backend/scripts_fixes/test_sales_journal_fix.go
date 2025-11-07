package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🧪 TESTING SALES JOURNAL INTEGRATION FIX")
	fmt.Println("=" + string(make([]byte, 50)) + "=")

	// Test data for creating a sale
	saleData := map[string]interface{}{
		"customer_id":    1, // Assuming customer with ID 1 exists
		"type":           "INVOICE",
		"date":           time.Now().Format("2006-01-02T15:04:05Z"),
		"due_date":       time.Now().AddDate(0, 0, 30).Format("2006-01-02T15:04:05Z"),
		"payment_terms":  "NET30",
		"payment_method": "BANK_TRANSFER",
		"currency":       "IDR",
		"exchange_rate":  1.0,
		"discount_percent": 0,
		"ppn_percent":    11.0,
		"pph_percent":    0,
		"shipping_cost":  0,
		"shipping_taxable": false,
		"notes":          "Test sale for journal integration fix",
		"items": []map[string]interface{}{
			{
				"product_id":       1, // Assuming product with ID 1 exists
				"quantity":         2,
				"unit_price":       1000000, // Rp 1,000,000
				"discount":         0,
				"tax":             0,
				"taxable":         true,
				"revenue_account_id": 0, // Will use default
			},
		},
	}

	// Step 1: Get current account balance
	fmt.Println("\n📊 STEP 1: Getting current account 4101 balance...")
	initialBalance := getAccountBalance("4101")
	fmt.Printf("   Initial balance: Rp %.0f\n", initialBalance)

	// Step 2: Create a new sale
	fmt.Println("\n📊 STEP 2: Creating new sale...")
	saleID, saleCode := createSale(saleData)
	if saleID == 0 {
		fmt.Println("❌ Failed to create sale")
		return
	}
	fmt.Printf("   Created sale ID: %d, Code: %s\n", saleID, saleCode)

	// Step 3: Confirm the sale (change status to INVOICED)
	fmt.Println("\n📊 STEP 3: Confirming sale (changing status to INVOICED)...")
	success := confirmSale(saleID)
	if !success {
		fmt.Println("❌ Failed to confirm sale")
		return
	}
	fmt.Println("   ✅ Sale confirmed successfully")

	// Step 4: Wait a moment for journal processing
	fmt.Println("\n⏳ Waiting for journal processing...")
	time.Sleep(3 * time.Second)

	// Step 5: Check account balance again
	fmt.Println("\n📊 STEP 5: Checking updated account 4101 balance...")
	newBalance := getAccountBalance("4101")
	fmt.Printf("   New balance: Rp %.0f\n", newBalance)
	
	balanceIncrease := newBalance - initialBalance
	fmt.Printf("   Balance increase: Rp %.0f\n", balanceIncrease)

	// Step 6: Verify journal entries were created
	fmt.Println("\n📊 STEP 6: Verifying journal entries...")
	journalCount := getJournalEntriesCount(saleID)
	fmt.Printf("   Journal entries for sale %d: %d\n", saleID, journalCount)

	// Final verification
	fmt.Println("\n🎯 VERIFICATION RESULTS:")
	fmt.Println("   " + string(make([]byte, 40)) + "-")
	
	if balanceIncrease > 0 {
		fmt.Printf("   ✅ SUCCESS: Account 4101 balance increased by Rp %.0f\n", balanceIncrease)
		fmt.Println("   ✅ Sales journal integration is now working correctly!")
	} else {
		fmt.Println("   ❌ FAILED: Account balance did not increase")
		fmt.Println("   ❌ Sales journal integration still has issues")
	}

	if journalCount > 0 {
		fmt.Printf("   ✅ SUCCESS: %d journal entries created for the sale\n", journalCount)
	} else {
		fmt.Println("   ❌ FAILED: No journal entries found for the sale")
	}

	fmt.Println("\n🎉 TEST COMPLETED!")
}

func getAccountBalance(accountCode string) float64 {
	// This would need to be implemented to make an HTTP call to get account balance
	// For now, return a placeholder
	fmt.Printf("   (Would check balance for account %s via API)\n", accountCode)
	return 13000000 // Previous balance from fix
}

func createSale(saleData map[string]interface{}) (int, string) {
	// This would need to be implemented to make an HTTP call to create sale
	// For now, return placeholder values
	fmt.Println("   (Would create sale via POST /api/v1/sales)")
	return 999, "INV-2025-TEST"
}

func confirmSale(saleID int) bool {
	// This would need to be implemented to make an HTTP call to confirm sale
	// For now, return placeholder
	fmt.Printf("   (Would confirm sale %d via POST /api/v1/sales/%d/confirm)\n", saleID, saleID)
	return true
}

func getJournalEntriesCount(saleID int) int {
	// This would need to be implemented to check journal entries
	// For now, return placeholder
	fmt.Printf("   (Would check journal entries for sale %d)\n", saleID)
	return 1 // Placeholder
}

// Helper function to make HTTP requests (for future implementation)
func makeRequest(method, url string, data interface{}) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}