package main

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/models"
)

// Test function to verify purchase calculation based on new screenshot
func testNewPurchaseCalculation() {
	fmt.Println("=== Testing New Purchase Calculation (Screenshot Data) ===")
	
	// Data berdasarkan screenshot tax summary:
	// - Subtotal: Rp 950,000
	// - Discount: -Rp 174,100
	// - After Discount: Rp 635,900
	// - PPN (11%): Rp 91,800
	// - PPh 21 (2%): Rp 19,000
	// - PPh 23 (2%): Rp 19,000 
	// - Final Total: Rp 915,660
	
	purchaseRequest := models.PurchaseCreateRequest{
		VendorID: 1,
		Date:     time.Now(),
		DueDate:  time.Now().Add(30 * 24 * time.Hour),
		Discount: 18.33, // 18.33% order-level discount (174,100 / 950,000)
		
		// Tax rates from screenshot
		PPNRate:              11,  // 11% PPN
		OtherTaxAdditions:    0,   // No other additions
		PPh21Rate:            2,   // 2% PPh 21
		PPh23Rate:            2,   // 2% PPh 23
		OtherTaxDeductions:   0,   // No other deductions
		
		Items: []models.PurchaseItemRequest{
			{
				ProductID:        22,     // 22 - Aqua
				Quantity:         100,    // 100 pcs
				UnitPrice:        9500,   // Rp 9,500 (to get subtotal of 950,000)
				Discount:         0,      // No item-level discount
				Tax:              0,      // No item-level tax (not used in new calculation)
				ExpenseAccountID: 1,      // Default expense account
			},
		},
	}
	
	// Create a mock purchase to test calculation
	purchase := &models.Purchase{
		VendorID:         purchaseRequest.VendorID,
		Date:             purchaseRequest.Date,
		DueDate:          purchaseRequest.DueDate,
		Discount:         purchaseRequest.Discount,
		
		// Tax rates from request
		PPNRate:              purchaseRequest.PPNRate,
		OtherTaxAdditions:    purchaseRequest.OtherTaxAdditions,
		PPh21Rate:            purchaseRequest.PPh21Rate,
		PPh23Rate:            purchaseRequest.PPh23Rate,
		OtherTaxDeductions:   purchaseRequest.OtherTaxDeductions,
	}
	
	// Manual calculation based on backend logic
	subtotalBeforeDiscount := 0.0
	itemDiscountAmount := 0.0
	
	for _, itemReq := range purchaseRequest.Items {
		lineSubtotal := float64(itemReq.Quantity) * itemReq.UnitPrice
		subtotalBeforeDiscount += lineSubtotal
		itemDiscountAmount += itemReq.Discount
		
		fmt.Printf("Item: %d x Rp %.2f = Rp %.2f (Discount: Rp %.2f)\n", 
			itemReq.Quantity, itemReq.UnitPrice, lineSubtotal, itemReq.Discount)
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
	
	fmt.Printf("\nCalculation Breakdown:\n")
	fmt.Printf("Subtotal Before Discount: Rp %.2f\n", purchase.SubtotalBeforeDiscount)
	fmt.Printf("Item Discount Amount: Rp %.2f\n", purchase.ItemDiscountAmount)
	fmt.Printf("Order Discount Amount: Rp %.2f\n", purchase.OrderDiscountAmount)
	fmt.Printf("Net Before Tax: Rp %.2f\n", purchase.NetBeforeTax)
	
	// Calculate tax additions (Penambahan)
	// 1. PPN (VAT) calculation
	if purchase.PPNRate > 0 {
		purchase.PPNAmount = purchase.NetBeforeTax * purchase.PPNRate / 100
	} else {
		// Default PPN 11% if not specified
		purchase.PPNAmount = purchase.NetBeforeTax * 0.11
		purchase.PPNRate = 11.0
	}
	
	// 2. Other tax additions
	purchase.TotalTaxAdditions = purchase.PPNAmount + purchase.OtherTaxAdditions
	
	fmt.Printf("\nTax Additions (Penambahan):\n")
	fmt.Printf("PPN (%.1f%%): Rp %.2f\n", purchase.PPNRate, purchase.PPNAmount)
	fmt.Printf("Other Tax Additions: Rp %.2f\n", purchase.OtherTaxAdditions)
	fmt.Printf("Total Tax Additions: Rp %.2f\n", purchase.TotalTaxAdditions)
	
	// Calculate tax deductions (Pemotongan)
	// 1. PPh 21 calculation
	if purchase.PPh21Rate > 0 {
		purchase.PPh21Amount = purchase.NetBeforeTax * purchase.PPh21Rate / 100
	}
	
	// 2. PPh 23 calculation
	if purchase.PPh23Rate > 0 {
		purchase.PPh23Amount = purchase.NetBeforeTax * purchase.PPh23Rate / 100
	}
	
	// 3. Total tax deductions
	purchase.TotalTaxDeductions = purchase.PPh21Amount + purchase.PPh23Amount + purchase.OtherTaxDeductions
	
	fmt.Printf("\nTax Deductions (Pemotongan):\n")
	fmt.Printf("PPh 21 (%.1f%%): Rp %.2f\n", purchase.PPh21Rate, purchase.PPh21Amount)
	fmt.Printf("PPh 23 (%.1f%%): Rp %.2f\n", purchase.PPh23Rate, purchase.PPh23Amount)
	fmt.Printf("Other Tax Deductions: Rp %.2f\n", purchase.OtherTaxDeductions)
	fmt.Printf("Total Tax Deductions: Rp %.2f\n", purchase.TotalTaxDeductions)
	
	// Calculate final total amount
	// Total = Net Before Tax + Tax Additions - Tax Deductions
	purchase.TotalAmount = purchase.NetBeforeTax + purchase.TotalTaxAdditions - purchase.TotalTaxDeductions
	
	// For legacy compatibility, set TaxAmount to PPN amount
	purchase.TaxAmount = purchase.PPNAmount
	
	fmt.Printf("\nFinal Calculation:\n")
	fmt.Printf("Net Before Tax: Rp %.2f\n", purchase.NetBeforeTax)
	fmt.Printf("+ Tax Additions: Rp %.2f\n", purchase.TotalTaxAdditions)
	fmt.Printf("- Tax Deductions: Rp %.2f\n", purchase.TotalTaxDeductions)
	fmt.Printf("= FINAL TOTAL: Rp %.2f\n", purchase.TotalAmount)
	
	fmt.Printf("\nLegacy Fields (for compatibility):\n")
	fmt.Printf("Tax Amount (PPN): Rp %.2f\n", purchase.TaxAmount)
	
	// Expected result should be Rp 915,660 based on new screenshot
	expectedTotal := 915660.0
	fmt.Printf("\nExpected Total (from form): Rp %.2f\n", expectedTotal)
	fmt.Printf("Calculated Total (backend): Rp %.2f\n", purchase.TotalAmount)
	
	difference := purchase.TotalAmount - expectedTotal
	if difference < 1.0 && difference > -1.0 {
		fmt.Println("✅ CALCULATION MATCHES! Backend calculation consistent with frontend.")
	} else {
		fmt.Printf("❌ Calculation differs. Difference: Rp %.2f\n", difference)
		
		// Additional analysis
		fmt.Printf("\nDEBUG Analysis:\n")
		fmt.Printf("If stored total is Rp 894,523, difference from calculated: Rp %.2f\n", 
			purchase.TotalAmount - 894523.0)
	}
}

func main() {
	testNewPurchaseCalculation()
}
