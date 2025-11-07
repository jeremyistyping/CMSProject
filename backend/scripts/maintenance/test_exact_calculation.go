package main

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/models"
)

func testExactCalculation() {
	fmt.Println("=== Testing Exact Purchase Calculation (Latest Screenshot) ===")
	
	// Data exact dari screenshot:
	// Product: 22 - Aqua
	// Quantity: 100 pcs  
	// Unit Price: 10,000
	// Discount: 50,000
	// Subtotal: 950,000
	// Order Discount: 10%
	// After Discount: 855,500
	
	purchaseRequest := models.PurchaseCreateRequest{
		VendorID: 1,
		Date:     time.Now(),
		DueDate:  time.Now().Add(30 * 24 * time.Hour),
		Discount: 10, // 10% order-level discount visible in screenshot
		
		// Tax rates dari screenshot
		PPNRate:              11,  // 11% PPN
		OtherTaxAdditions:    5,   // 5% other additions (to match screenshot)
		PPh21Rate:            2,   // 2% PPh 21  
		PPh23Rate:            2,   // 2% PPh 23
		OtherTaxDeductions:   2,   // 2% other deductions
		
		Items: []models.PurchaseItemRequest{
			{
				ProductID:        22,     // 22 - Aqua
				Quantity:         100,    // 100 pcs
				UnitPrice:        10000,  // Rp 10,000
				Discount:         50000,  // Rp 50,000 item discount
				Tax:              0,      // No item-level tax
				ExpenseAccountID: 1,      
			},
		},
	}
	
	// Create mock purchase
	purchase := &models.Purchase{
		VendorID:         purchaseRequest.VendorID,
		Date:             purchaseRequest.Date,
		DueDate:          purchaseRequest.DueDate,
		Discount:         purchaseRequest.Discount,
		
		PPNRate:              purchaseRequest.PPNRate,
		OtherTaxAdditions:    purchaseRequest.OtherTaxAdditions,
		PPh21Rate:            purchaseRequest.PPh21Rate,
		PPh23Rate:            purchaseRequest.PPh23Rate,
		OtherTaxDeductions:   purchaseRequest.OtherTaxDeductions,
	}
	
	// Calculate based on backend logic
	subtotalBeforeDiscount := 0.0
	itemDiscountAmount := 0.0
	
	for _, itemReq := range purchaseRequest.Items {
		lineSubtotal := float64(itemReq.Quantity) * itemReq.UnitPrice
		subtotalBeforeDiscount += lineSubtotal
		itemDiscountAmount += itemReq.Discount
		
		fmt.Printf("Item: %d x Rp %.2f = Rp %.2f\n", 
			itemReq.Quantity, itemReq.UnitPrice, lineSubtotal)
		fmt.Printf("Item Discount: Rp %.2f\n", itemReq.Discount)
	}
	
	// Calculate order-level discount
	orderDiscountAmount := 0.0
	if purchase.Discount > 0 {
		orderDiscountAmount = (subtotalBeforeDiscount - itemDiscountAmount) * purchase.Discount / 100
	}
	
	// Set basic calculated fields
	purchase.SubtotalBeforeDiscount = subtotalBeforeDiscount
	purchase.ItemDiscountAmount = itemDiscountAmount
	purchase.OrderDiscountAmount = orderDiscountAmount
	purchase.NetBeforeTax = subtotalBeforeDiscount - itemDiscountAmount - orderDiscountAmount
	
	fmt.Printf("\n=== Calculation Breakdown ===\n")
	fmt.Printf("Subtotal Before Discount: Rp %.2f\n", purchase.SubtotalBeforeDiscount)
	fmt.Printf("Item Discount Amount: Rp %.2f\n", purchase.ItemDiscountAmount)
	fmt.Printf("Order Discount (%.0f%%): Rp %.2f\n", purchase.Discount, purchase.OrderDiscountAmount)
	fmt.Printf("Net Before Tax: Rp %.2f\n", purchase.NetBeforeTax)
	
	// Calculate tax additions
	if purchase.PPNRate > 0 {
		purchase.PPNAmount = purchase.NetBeforeTax * purchase.PPNRate / 100
	}
	
	// Calculate other tax additions amount
	otherTaxAdditionsAmount := 0.0
	if purchase.OtherTaxAdditions > 0 {
		otherTaxAdditionsAmount = purchase.NetBeforeTax * purchase.OtherTaxAdditions / 100
	}
	
	purchase.TotalTaxAdditions = purchase.PPNAmount + otherTaxAdditionsAmount
	
	fmt.Printf("\n=== Tax Additions (Penambahan) ===\n")
	fmt.Printf("PPN (%.1f%%): Rp %.2f\n", purchase.PPNRate, purchase.PPNAmount)
	fmt.Printf("Other Additions (%.1f%%): Rp %.2f\n", purchase.OtherTaxAdditions, otherTaxAdditionsAmount)
	fmt.Printf("Total Tax Additions: Rp %.2f\n", purchase.TotalTaxAdditions)
	
	// Calculate tax deductions
	if purchase.PPh21Rate > 0 {
		purchase.PPh21Amount = purchase.NetBeforeTax * purchase.PPh21Rate / 100
	}
	
	if purchase.PPh23Rate > 0 {
		purchase.PPh23Amount = purchase.NetBeforeTax * purchase.PPh23Rate / 100
	}
	
	// Calculate other tax deductions amount
	otherTaxDeductionsAmount := 0.0
	if purchase.OtherTaxDeductions > 0 {
		otherTaxDeductionsAmount = purchase.NetBeforeTax * purchase.OtherTaxDeductions / 100
	}
	
	purchase.TotalTaxDeductions = purchase.PPh21Amount + purchase.PPh23Amount + otherTaxDeductionsAmount
	
	fmt.Printf("\n=== Tax Deductions (Pemotongan) ===\n")
	fmt.Printf("PPh 21 (%.1f%%): Rp %.2f\n", purchase.PPh21Rate, purchase.PPh21Amount)
	fmt.Printf("PPh 23 (%.1f%%): Rp %.2f\n", purchase.PPh23Rate, purchase.PPh23Amount)
	fmt.Printf("Other Deductions (%.1f%%): Rp %.2f\n", purchase.OtherTaxDeductions, otherTaxDeductionsAmount)
	fmt.Printf("Total Tax Deductions: Rp %.2f\n", purchase.TotalTaxDeductions)
	
	// Calculate final total
	purchase.TotalAmount = purchase.NetBeforeTax + purchase.TotalTaxAdditions - purchase.TotalTaxDeductions
	purchase.TaxAmount = purchase.PPNAmount
	
	fmt.Printf("\n=== Final Calculation ===\n")
	fmt.Printf("Net Before Tax: Rp %.2f\n", purchase.NetBeforeTax)
	fmt.Printf("+ Tax Additions: Rp %.2f\n", purchase.TotalTaxAdditions)
	fmt.Printf("- Tax Deductions: Rp %.2f\n", purchase.TotalTaxDeductions)
	fmt.Printf("= FINAL TOTAL: Rp %.2f\n", purchase.TotalAmount)
	
	// Compare with expected values from screenshot
	expectedFinalTotal := 883560.0  // From form
	storedTotal := 914853.0         // From database
	
	fmt.Printf("\n=== Comparison ===\n")
	fmt.Printf("Expected Final Total (form): Rp %.2f\n", expectedFinalTotal)
	fmt.Printf("Stored Total (database): Rp %.2f\n", storedTotal)
	fmt.Printf("Calculated Total (backend): Rp %.2f\n", purchase.TotalAmount)
	
	fmt.Printf("\nDifference from form: Rp %.2f\n", purchase.TotalAmount - expectedFinalTotal)
	fmt.Printf("Difference from stored: Rp %.2f\n", purchase.TotalAmount - storedTotal)
	
	if purchase.TotalAmount == expectedFinalTotal {
		fmt.Println("\n✅ Backend calculation matches form!")
	} else if purchase.TotalAmount == storedTotal {
		fmt.Println("\n⚠️ Backend calculation matches stored value but not form")
	} else {
		fmt.Println("\n❌ Backend calculation doesn't match either value")
	}
}

func main() {
	testExactCalculation()
}
