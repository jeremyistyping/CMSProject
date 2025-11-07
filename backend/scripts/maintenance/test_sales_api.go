package main

import (
	"encoding/json"
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	// Initialize database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("=== Testing Sales API Response for GET /sales/21 ===")
	
	// Initialize repositories and services like in the actual API
	salesRepo := repositories.NewSalesRepository(db)
	productRepo := repositories.NewProductRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	pdfService := services.NewPDFService()
	salesService := services.NewSalesService(salesRepo, productRepo, contactRepo, accountRepo, nil, pdfService)
	
	// Test the exact same method that API controller calls
	sale, err := salesService.GetSaleByID(21)
	if err != nil {
		fmt.Printf("❌ Error getting sale: %v\n", err)
		return
	}
	
	if sale == nil {
		fmt.Println("❌ Sale not found")
		return
	}
	
	// Convert to JSON like the API would do
	jsonData, err := json.MarshalIndent(sale, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling to JSON: %v\n", err)
		return
	}
	
	fmt.Println("✅ Sale found! API Response structure:")
	fmt.Println("=====================================")
	fmt.Println(string(jsonData))
	
	// Analyze the response structure for frontend integration
	fmt.Println("\n=== Response Analysis for Frontend Integration ===")
	
	// Check core fields
	fmt.Printf("Sale ID: %d\n", sale.ID)
	fmt.Printf("Code: %s\n", sale.Code)
	fmt.Printf("Type: %s\n", sale.Type)
	fmt.Printf("Status: %s\n", sale.Status)
	fmt.Printf("Date: %s\n", sale.Date.Format("2006-01-02"))
	fmt.Printf("Due Date: %s\n", sale.DueDate.Format("2006-01-02"))
	
	// Check customer info
	if sale.Customer.ID > 0 {
		fmt.Printf("Customer: %s (ID: %d)\n", sale.Customer.Name, sale.Customer.ID)
	} else {
		fmt.Println("⚠️  Customer data not loaded properly")
	}
	
	// Check sales person
	if sale.SalesPersonID != nil && sale.SalesPerson != nil {
		fmt.Printf("Sales Person: %s (ID: %d)\n", sale.SalesPerson.Name, *sale.SalesPersonID)
	} else if sale.SalesPersonID != nil {
		fmt.Printf("⚠️  Sales Person ID: %d (but data not loaded)\n", *sale.SalesPersonID)
	} else {
		fmt.Println("Sales Person: None")
	}
	
	// Check amounts and calculate consistency
	fmt.Printf("Subtotal: %.2f\n", sale.Subtotal)
	fmt.Printf("SubTotal alias: %.2f\n", sale.SubTotal) // Frontend uses this field
	fmt.Printf("Discount Amount: %.2f\n", sale.DiscountAmount)
	fmt.Printf("Taxable Amount: %.2f\n", sale.TaxableAmount)
	fmt.Printf("PPN: %.2f\n", sale.PPN)
	fmt.Printf("PPh: %.2f\n", sale.PPh)
	fmt.Printf("Total Tax: %.2f\n", sale.TotalTax)
	fmt.Printf("Shipping Cost: %.2f\n", sale.ShippingCost)
	fmt.Printf("Total Amount: %.2f\n", sale.TotalAmount)
	fmt.Printf("Paid Amount: %.2f\n", sale.PaidAmount)
	fmt.Printf("Outstanding Amount: %.2f\n", sale.OutstandingAmount)
	
	// Check items
	fmt.Printf("Sale Items: %d\n", len(sale.SaleItems))
	for i, item := range sale.SaleItems {
		fmt.Printf("  Item %d: %s x%d @ %.2f = %.2f\n", 
			i+1, item.Description, item.Quantity, item.UnitPrice, item.LineTotal)
		if item.Product.ID > 0 {
			fmt.Printf("    Product: %s (ID: %d)\n", item.Product.Name, item.Product.ID)
		} else {
			fmt.Printf("    ⚠️  Product data not loaded for item %d\n", i+1)
		}
	}
	
	// Check payment history
	fmt.Printf("Payment Records: %d\n", len(sale.SalePayments))
	for i, payment := range sale.SalePayments {
		fmt.Printf("  Payment %d: %.2f on %s via %s\n", 
			i+1, payment.Amount, payment.Date.Format("2006-01-02"), payment.Method)
	}
	
	// Check returns
	fmt.Printf("Return Records: %d\n", len(sale.SaleReturns))
	
	// Frontend integration checklist
	fmt.Println("\n=== Frontend Integration Checklist ===")
	checkList := []struct {
		name   string
		check  bool
		issue  string
	}{
		{"Sale ID exists", sale.ID > 0, ""},
		{"Code exists", sale.Code != "", ""},
		{"Customer loaded", sale.Customer.ID > 0, "Customer relation not preloaded"},
		{"Sales person loaded", sale.SalesPersonID == nil || sale.SalesPerson != nil, "SalesPerson relation may not be preloaded"},
		{"Sale items loaded", len(sale.SaleItems) > 0, "No sale items found"},
		{"Product data in items", len(sale.SaleItems) == 0 || sale.SaleItems[0].Product.ID > 0, "Product relation in sale items not preloaded"},
		{"Amount calculations consistent", sale.TotalAmount > 0, "Total amount calculation may be wrong"},
		{"SubTotal alias working", sale.SubTotal == sale.Subtotal, "SubTotal alias field not working"},
	}
	
	allGood := true
	for _, check := range checkList {
		if check.check {
			fmt.Printf("✅ %s\n", check.name)
		} else {
			fmt.Printf("❌ %s - %s\n", check.name, check.issue)
			allGood = false
		}
	}
	
	if allGood {
		fmt.Println("\n🎉 All checks passed! API response is ready for frontend integration.")
	} else {
		fmt.Println("\n⚠️  Some issues found. Check the items marked with ❌ above.")
	}
	
	// Check if SubTotal alias is working properly
	if sale.SubTotal != sale.Subtotal {
		fmt.Printf("\n🔧 Fix needed: SubTotal alias (%.2f) != Subtotal (%.2f)\n", sale.SubTotal, sale.Subtotal)
	}
}
