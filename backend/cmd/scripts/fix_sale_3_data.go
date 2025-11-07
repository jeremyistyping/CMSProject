package main

import (
	"log"
	"math"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load config from environment
	cfg := config.LoadConfig()

	// Connect to database using config
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Connected to database successfully")

	log.Println("=== Fixing Sale #3 Data Corruption ===\n")

	// Load Sale #3 with items
	var sale models.Sale
	if err := db.Preload("SaleItems.Product").First(&sale, 3).Error; err != nil {
		log.Fatalf("Failed to load Sale #3: %v", err)
	}

	log.Printf("📊 Current Sale #3 Data:")
	log.Printf("  Subtotal: %.2f", sale.Subtotal)
	log.Printf("  DiscountAmount: %.2f (%.2f%%)", sale.DiscountAmount, sale.DiscountPercent)
	log.Printf("  PPN: %.2f", sale.PPN)
	log.Printf("  OtherTaxAdditions: %.2f", sale.OtherTaxAdditions)
	log.Printf("  OtherTaxDeductions: %.2f", sale.OtherTaxDeductions)
	log.Printf("  ShippingCost: %.2f", sale.ShippingCost)
	log.Printf("  TotalAmount: %.2f", sale.TotalAmount)
	log.Println()

	// Calculate correct values from sale items
	var calculatedSubtotal float64
	for _, item := range sale.SaleItems {
		// LineTotal should already be correct (Qty * UnitPrice - Discount)
		calculatedSubtotal += item.LineTotal
		log.Printf("  Item: %s - LineTotal: %.2f", item.Product.Name, item.LineTotal)
	}
	log.Printf("\n  Calculated Subtotal from items: %.2f", calculatedSubtotal)

	// Apply sale-level discount if exists
	var discountAmount float64
	if sale.DiscountPercent > 0 {
		discountAmount = calculatedSubtotal * sale.DiscountPercent / 100
		log.Printf("  Sale-level discount (%.2f%%): %.2f", sale.DiscountPercent, discountAmount)
	} else {
		discountAmount = sale.DiscountAmount
	}

	// Net after discount (before tax)
	netBeforeTax := calculatedSubtotal - discountAmount
	log.Printf("  Net before tax: %.2f", netBeforeTax)

	// Calculate PPN (11% of net amount)
	var calculatedPPN float64
	if sale.PPNPercent > 0 || sale.PPNRate > 0 {
		ppnRate := sale.PPNPercent
		if ppnRate == 0 {
			ppnRate = sale.PPNRate
		}
		calculatedPPN = netBeforeTax * ppnRate / 100
		log.Printf("  Calculated PPN (%.2f%%): %.2f", ppnRate, calculatedPPN)
	}

	// Add shipping cost (if not already in subtotal)
	var calculatedShipping float64
	if sale.ShippingCost > 0 {
		calculatedShipping = sale.ShippingCost
		log.Printf("  Shipping cost: %.2f", calculatedShipping)
	}

	// Gross amount = Net + PPN + Shipping + OtherTaxAdditions
	grossAmount := netBeforeTax + calculatedPPN + calculatedShipping + sale.OtherTaxAdditions
	log.Printf("  Gross amount: %.2f", grossAmount)

	// Net amount = Gross - Tax Deductions
	totalTaxDeductions := sale.PPh + sale.PPh21Amount + sale.PPh23Amount + sale.OtherTaxDeductions
	log.Printf("  Total tax deductions: %.2f", totalTaxDeductions)

	calculatedTotal := grossAmount - totalTaxDeductions
	log.Printf("  Calculated Total: %.2f", calculatedTotal)
	log.Printf("  Current TotalAmount: %.2f", sale.TotalAmount)
	log.Printf("  Difference: %.2f", math.Abs(calculatedTotal-sale.TotalAmount))
	log.Println()

	// Check if we need to fix
	needsFix := false
	tolerance := 1.0 // 1 Rupiah tolerance

	if math.Abs(calculatedSubtotal-sale.Subtotal) > tolerance {
		log.Printf("⚠️ Subtotal needs fix: %.2f → %.2f", sale.Subtotal, calculatedSubtotal)
		needsFix = true
	}

	if math.Abs(calculatedPPN-sale.PPN) > tolerance {
		log.Printf("⚠️ PPN needs fix: %.2f → %.2f", sale.PPN, calculatedPPN)
		needsFix = true
	}

	if math.Abs(calculatedTotal-sale.TotalAmount) > tolerance {
		log.Printf("⚠️ TotalAmount needs fix: %.2f → %.2f", sale.TotalAmount, calculatedTotal)
		needsFix = true
	}

	// Check for suspicious values
	if sale.OtherTaxAdditions > 0 && sale.OtherTaxAdditions < 200 {
		log.Printf("⚠️ Suspicious OtherTaxAdditions: %.2f (might be incorrect)", sale.OtherTaxAdditions)
		log.Printf("   This value will be KEPT but flagged for review")
	}

	if sale.OtherTaxDeductions > 0 && sale.OtherTaxDeductions < 200 {
		log.Printf("⚠️ Suspicious OtherTaxDeductions: %.2f (might be incorrect)", sale.OtherTaxDeductions)
		log.Printf("   This value will be KEPT but flagged for review")
	}

	if !needsFix {
		log.Println("\n✅ No fixes needed - data is consistent!")
		return
	}

	log.Println("\n🔧 Applying fixes...")

	// Update sale with corrected values
	updates := map[string]interface{}{
		"subtotal":        calculatedSubtotal,
		"discount_amount": discountAmount,
		"net_before_tax":  netBeforeTax,
		"ppn":             calculatedPPN,
		"ppn_amount":      calculatedPPN, // Enhanced field
		"total_amount":    calculatedTotal,
	}

	// Calculate outstanding amount (if not paid)
	if sale.PaidAmount > 0 {
		updates["outstanding_amount"] = calculatedTotal - sale.PaidAmount
	} else {
		updates["outstanding_amount"] = calculatedTotal
	}

	if err := db.Model(&sale).Updates(updates).Error; err != nil {
		log.Fatalf("❌ Failed to update Sale #3: %v", err)
	}

	log.Println("✅ Sale #3 has been updated with corrected values")
	log.Println()

	// Reload and display
	if err := db.Preload("SaleItems.Product").First(&sale, 3).Error; err != nil {
		log.Fatalf("Failed to reload Sale #3: %v", err)
	}

	log.Println("📊 Updated Sale #3 Data:")
	log.Printf("  Subtotal: %.2f", sale.Subtotal)
	log.Printf("  DiscountAmount: %.2f", sale.DiscountAmount)
	log.Printf("  NetBeforeTax: %.2f", sale.NetBeforeTax)
	log.Printf("  PPN: %.2f", sale.PPN)
	log.Printf("  OtherTaxAdditions: %.2f", sale.OtherTaxAdditions)
	log.Printf("  OtherTaxDeductions: %.2f", sale.OtherTaxDeductions)
	log.Printf("  ShippingCost: %.2f", sale.ShippingCost)
	log.Printf("  TotalAmount: %.2f", sale.TotalAmount)
	log.Printf("  OutstandingAmount: %.2f", sale.OutstandingAmount)
	log.Println()

	log.Println("✅ Fix Complete! You can now try creating the invoice again.")
	log.Println()
	log.Println("⚠️ NOTE: If OtherTaxAdditions/Deductions are still suspicious,")
	log.Println("   please manually verify and update them in the frontend.")
}
