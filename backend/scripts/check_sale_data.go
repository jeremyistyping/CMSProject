package main

import (
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Println("🔍 Checking sale data...")
	
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}
	
	// Check sale #1
	var sale models.Sale
	if err := db.Preload("Customer").Preload("SaleItems").First(&sale, 1).Error; err != nil {
		log.Fatalf("Error loading sale: %v", err)
	}
	
	log.Printf("\n📝 Sale Details:")
	log.Printf("  ID: %d", sale.ID)
	log.Printf("  Code: %s", sale.Code)
	log.Printf("  Invoice Number: %s", sale.InvoiceNumber)
	log.Printf("  Customer: %s", sale.Customer.Name)
	log.Printf("  Status: %s", sale.Status)
	log.Printf("  Payment Method: %s", sale.PaymentMethodType)
	log.Printf("  Date: %s", sale.Date.Format("2006-01-02"))
	log.Printf("\n💰 Financial Details:")
	log.Printf("  Subtotal: %.2f", sale.Subtotal)
	log.Printf("  PPN (%%.2f): %.2f", sale.PPNPercent, sale.PPN)
	log.Printf("  PPh: %.2f", sale.PPh)
	log.Printf("  Total Tax: %.2f", sale.TotalTax)
	log.Printf("  TOTAL AMOUNT: %.2f  ⬅️ THIS IS THE ISSUE!", sale.TotalAmount)
	log.Printf("  Paid Amount: %.2f", sale.PaidAmount)
	log.Printf("  Outstanding: %.2f", sale.OutstandingAmount)
	
	log.Printf("\n📦 Sale Items:")
	for i, item := range sale.SaleItems {
		log.Printf("  Item #%d:", i+1)
		log.Printf("    Product ID: %d", item.ProductID)
		log.Printf("    Quantity: %.2f", item.Quantity)
		log.Printf("    Unit Price: %.2f", item.UnitPrice)
		log.Printf("    Line Total: %.2f", item.LineTotal)
		log.Printf("    PPN Amount: %.2f", item.PPNAmount)
		log.Printf("    Final Amount: %.2f", item.FinalAmount)
	}
	
	// Calculate expected totals
	expectedSubtotal := 0.0
	for _, item := range sale.SaleItems {
		expectedSubtotal += item.LineTotal
	}
	expectedPPN := expectedSubtotal * (sale.PPNPercent / 100)
	expectedTotal := expectedSubtotal + expectedPPN
	
	log.Printf("\n🧮 Expected Calculations:")
	log.Printf("  Expected Subtotal: %.2f", expectedSubtotal)
	log.Printf("  Expected PPN (11%%): %.2f", expectedPPN)
	log.Printf("  Expected Total: %.2f", expectedTotal)
	
	log.Printf("\n❓ Comparison:")
	if sale.TotalAmount != expectedTotal {
		log.Printf("  ❌ MISMATCH! Stored total (%.2f) != Expected total (%.2f)", sale.TotalAmount, expectedTotal)
		log.Printf("  ❌ Difference: %.2f", sale.TotalAmount-expectedTotal)
	} else {
		log.Printf("  ✅ Total matches expected!")
	}
}

