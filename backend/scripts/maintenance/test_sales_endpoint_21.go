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

// Test structures
type SaleResponse struct {
	ID                 uint             `json:"id"`
	Code               string           `json:"code"`
	CustomerID         uint             `json:"customer_id"`
	Status             string           `json:"status"`
	Date               string           `json:"date"`
	DueDate            string           `json:"due_date"`
	Subtotal           float64          `json:"subtotal"`
	SubTotal           float64          `json:"sub_total"`
	DiscountAmount     float64          `json:"discount_amount"`
	Discount           float64          `json:"discount"`
	TotalTax           float64          `json:"total_tax"`
	Tax                float64          `json:"tax"`
	TotalAmount        float64          `json:"total_amount"`
	OutstandingAmount  float64          `json:"outstanding_amount"`
	PaidAmount         float64          `json:"paid_amount"`
	Customer           CustomerInfo     `json:"customer"`
	SalesPerson        *ContactInfo     `json:"sales_person"`
	SaleItems          []SaleItemInfo   `json:"sale_items"`
	Items              []SaleItemInfo   `json:"items"`
}

type CustomerInfo struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type ContactInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type SaleItemInfo struct {
	ID            uint        `json:"id"`
	ProductID     uint        `json:"product_id"`
	Quantity      int         `json:"quantity"`
	UnitPrice     float64     `json:"unit_price"`
	LineTotal     float64     `json:"line_total"`
	TotalPrice    float64     `json:"total_price"`
	Description   string      `json:"description"`
	Product       ProductInfo `json:"product"`
}

type ProductInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

const (
	baseURL = "http://localhost:8080/api/v1"
	token   = "" // Add your JWT token here if needed
)

func main() {
	log.Println("Testing Sales Endpoint /sales/21...")

	// Test 1: Get specific sale by ID
	testGetSale(21)

	// Test 2: Test confirm sale
	// testConfirmSale(21)

	// Test 3: Test create invoice from sale
	// testCreateInvoice(21)

	// Test 4: Test cancel sale with reason
	// testCancelSale(21, "Testing cancellation")

	log.Println("Sales endpoint testing completed!")
}

func testGetSale(saleID int) {
	log.Printf("Testing GET /sales/%d", saleID)

	url := fmt.Sprintf("%s/sales/%d", baseURL, saleID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	log.Printf("Response Status: %d", resp.StatusCode)
	log.Printf("Response Headers: %v", resp.Header)
	
	if resp.StatusCode == 200 {
		var sale SaleResponse
		if err := json.Unmarshal(body, &sale); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			log.Printf("Raw response: %s", string(body))
			return
		}

		log.Printf("✅ Sale found successfully!")
		log.Printf("   ID: %d", sale.ID)
		log.Printf("   Code: %s", sale.Code)
		log.Printf("   Status: %s", sale.Status)
		log.Printf("   Customer: %s (ID: %d)", sale.Customer.Name, sale.Customer.ID)
		
		if sale.SalesPerson != nil {
			log.Printf("   Sales Person: %s (ID: %d)", sale.SalesPerson.Name, sale.SalesPerson.ID)
		} else {
			log.Printf("   Sales Person: Not assigned")
		}

		log.Printf("   Financial Summary:")
		log.Printf("     Subtotal: %.2f (sub_total: %.2f)", sale.Subtotal, sale.SubTotal)
		log.Printf("     Discount: %.2f (discount_amount: %.2f)", sale.Discount, sale.DiscountAmount)
		log.Printf("     Tax: %.2f (total_tax: %.2f)", sale.Tax, sale.TotalTax)
		log.Printf("     Total Amount: %.2f", sale.TotalAmount)
		log.Printf("     Outstanding: %.2f", sale.OutstandingAmount)

		// Test field compatibility
		log.Printf("   Field Compatibility Check:")
		log.Printf("     Subtotal == SubTotal: %t (%.2f == %.2f)", 
			sale.Subtotal == sale.SubTotal, sale.Subtotal, sale.SubTotal)
		log.Printf("     Tax == TotalTax: %t (%.2f == %.2f)", 
			sale.Tax == sale.TotalTax, sale.Tax, sale.TotalTax)

		// Check items
		items := sale.SaleItems
		if len(items) == 0 && len(sale.Items) > 0 {
			items = sale.Items
		}

		log.Printf("   Items (%d total):", len(items))
		for i, item := range items {
			log.Printf("     Item %d: %s (Qty: %d × %.2f = %.2f/%.2f)", 
				i+1, item.Product.Name, item.Quantity, item.UnitPrice, 
				item.LineTotal, item.TotalPrice)
		}

	} else {
		log.Printf("❌ Error response: %s", string(body))
	}
}

func testConfirmSale(saleID int) {
	log.Printf("Testing POST /sales/%d/confirm", saleID)

	url := fmt.Sprintf("%s/sales/%d/confirm", baseURL, saleID)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	log.Printf("Response Status: %d", resp.StatusCode)
	if resp.StatusCode == 200 {
		log.Printf("✅ Sale confirmed successfully!")
	} else {
		log.Printf("❌ Error confirming sale: %s", string(body))
	}
}

func testCreateInvoice(saleID int) {
	log.Printf("Testing POST /sales/%d/invoice", saleID)

	url := fmt.Sprintf("%s/sales/%d/invoice", baseURL, saleID)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	log.Printf("Response Status: %d", resp.StatusCode)
	if resp.StatusCode == 200 {
		var sale SaleResponse
		if err := json.Unmarshal(body, &sale); err == nil {
			log.Printf("✅ Invoice created successfully!")
			log.Printf("   Invoice Number: %s", sale.Code) // Assuming code contains invoice info
			log.Printf("   Status: %s", sale.Status)
		} else {
			log.Printf("✅ Invoice created (couldn't parse response)")
		}
	} else {
		log.Printf("❌ Error creating invoice: %s", string(body))
	}
}

func testCancelSale(saleID int, reason string) {
	log.Printf("Testing POST /sales/%d/cancel", saleID)

	url := fmt.Sprintf("%s/sales/%d/cancel", baseURL, saleID)
	
	requestBody := map[string]string{
		"reason": reason,
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	log.Printf("Response Status: %d", resp.StatusCode)
	if resp.StatusCode == 200 {
		log.Printf("✅ Sale cancelled successfully!")
		log.Printf("   Reason: %s", reason)
	} else {
		log.Printf("❌ Error cancelling sale: %s", string(body))
	}
}
