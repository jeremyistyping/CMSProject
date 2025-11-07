package main

import (
	"log"
	"math"
	"os"
	"strconv"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Get sale ID from command line argument
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/scripts/fix_sale_calculation.go <sale_id>")
	}

	saleID, err := strconv.Atoi(os.Args[1])
	if err != nil || saleID <= 0 {
		log.Fatalf("Invalid sale ID: %s", os.Args[1])
	}

	// Load config from environment
	cfg := config.LoadConfig()

	// Connect to database using config
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Connected to database successfully")

	log.Printf("=== Fixing Sale #%d Calculation ===\n", saleID)

	// Load Sale with items
	var sale models.Sale
	if err := db.Preload("SaleItems.Product").First(&sale, saleID).Error; err != nil {
		log.Fatalf("Failed to load Sale #%d: %v", saleID, err)
	}

	log.Printf("📊 Current Sale #%d Data:", sale.ID)
	log.Printf("  Subtotal: %.2f", sale.Subtotal)
	log.Printf("  DiscountPercent: %.2f%%", sale.DiscountPercent)
	log.Printf("  DiscountAmount: %.2f", sale.DiscountAmount)
	log.Printf("  TaxableAmount: %.2f", sale.TaxableAmount)
	log.Printf("  PPN: %.2f", sale.PPN)
	log.Printf("  ShippingCost: %.2f", sale.ShippingCost)
	log.Printf("  OtherTaxAdditions: %.2f", sale.OtherTaxAdditions)
	log.Printf("  OtherTaxDeductions: %.2f", sale.OtherTaxDeductions)
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

	// Subtotal is correct (sum of LineTotal after item discounts)
	subtotalAfterItemDiscount := calculatedSubtotal

	// Apply sale-level discount if exists
	var saleLevelDiscountAmount float64
	if sale.DiscountPercent > 0 {
		saleLevelDiscountAmount = subtotalAfterItemDiscount * sale.DiscountPercent / 100
		log.Printf("  Sale-level discount (%.2f%%): %.2f", sale.DiscountPercent, saleLevelDiscountAmount)
	}

	// Net after all discounts (before tax)
	netBeforeTax := subtotalAfterItemDiscount - saleLevelDiscountAmount
	log.Printf("  Net before tax: %.2f", netBeforeTax)

	// Calculate PPN based on net amount after discount
	var calculatedPPN float64
	ppnRate := sale.PPNPercent
	if ppnRate == 0 {
		ppnRate = sale.PPNRate
	}
	if ppnRate > 0 {
		calculatedPPN = netBeforeTax * ppnRate / 100
		log.Printf("  Calculated PPN (%.2f%%): %.2f", ppnRate, calculatedPPN)
	}

	// Gross amount = Net + PPN + OtherTaxAdditions + Shipping
	grossAmount := netBeforeTax + calculatedPPN + sale.OtherTaxAdditions + sale.ShippingCost
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

	if math.Abs(saleLevelDiscountAmount-sale.DiscountAmount) > tolerance {
		log.Printf("⚠️ DiscountAmount needs fix: %.2f → %.2f", sale.DiscountAmount, saleLevelDiscountAmount)
		needsFix = true
	}

	if math.Abs(netBeforeTax-sale.TaxableAmount) > tolerance {
		log.Printf("⚠️ TaxableAmount needs fix: %.2f → %.2f", sale.TaxableAmount, netBeforeTax)
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

	if !needsFix {
		log.Println("\n✅ No fixes needed - data is consistent!")
		return
	}

	log.Println("\n🔧 Applying fixes...")

	// Update sale with corrected values
	updates := map[string]interface{}{
		"subtotal":         calculatedSubtotal,
		"discount_amount":  saleLevelDiscountAmount,
		"taxable_amount":   netBeforeTax,
		"net_before_tax":   netBeforeTax,
		"ppn":              calculatedPPN,
		"ppn_amount":       calculatedPPN, // Enhanced field
		"total_amount":     calculatedTotal,
		"outstanding_amount": calculatedTotal - sale.PaidAmount,
	}

	if err := db.Model(&sale).Updates(updates).Error; err != nil {
		log.Fatalf("❌ Failed to update Sale #%d: %v", saleID, err)
	}

	log.Println("✅ Sale has been updated with corrected values")
	log.Println()

	// Reload and display
	if err := db.Preload("SaleItems.Product").First(&sale, saleID).Error; err != nil {
		log.Fatalf("Failed to reload Sale #%d: %v", saleID, err)
	}

	log.Printf("📊 Updated Sale #%d Data:", sale.ID)
	log.Printf("  Subtotal: %.2f", sale.Subtotal)
	log.Printf("  DiscountAmount: %.2f (%.2f%%)", sale.DiscountAmount, sale.DiscountPercent)
	log.Printf("  TaxableAmount: %.2f", sale.TaxableAmount)
	log.Printf("  NetBeforeTax: %.2f", sale.NetBeforeTax)
	log.Printf("  PPN: %.2f", sale.PPN)
	log.Printf("  OtherTaxAdditions: %.2f", sale.OtherTaxAdditions)
	log.Printf("  OtherTaxDeductions: %.2f", sale.OtherTaxDeductions)
	log.Printf("  ShippingCost: %.2f", sale.ShippingCost)
	log.Printf("  TotalAmount: %.2f", sale.TotalAmount)
	log.Printf("  OutstandingAmount: %.2f", sale.OutstandingAmount)
	log.Println()

	log.Println("✅ Fix Complete! You can now try creating the invoice.")
	log.Println("\n💡 TIP: Run debug script again to verify:")
	log.Printf("   go run cmd/scripts/debug_sale_3_calc.go %d", saleID)
}
