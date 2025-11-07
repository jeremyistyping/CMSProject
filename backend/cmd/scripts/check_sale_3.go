package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	var sale models.Sale
	if err := db.First(&sale, 3).Error; err != nil {
		log.Fatalf("Failed to find Sale #3: %v", err)
	}

	fmt.Println("\n📊 Sale #3 Data:")
	fmt.Printf("  Subtotal:             %12.2f\n", sale.Subtotal)
	fmt.Printf("  PPN:                  %12.2f\n", sale.PPN)
	fmt.Printf("  OtherTaxAdditions:    %12.2f\n", sale.OtherTaxAdditions)
	fmt.Printf("  OtherTaxDeductions:   %12.2f\n", sale.OtherTaxDeductions)
	fmt.Printf("  ShippingCost:         %12.2f\n", sale.ShippingCost)
	fmt.Printf("  ─────────────────────────────────\n")
	fmt.Printf("  TotalAmount (DB):     %12.2f\n", sale.TotalAmount)
	fmt.Printf("  PaidAmount:           %12.2f\n", sale.PaidAmount)
	fmt.Printf("  OutstandingAmount:    %12.2f\n", sale.OutstandingAmount)
	
	// Calculate what TotalAmount should be
	expected := sale.Subtotal + sale.PPN + sale.OtherTaxAdditions + sale.ShippingCost - sale.OtherTaxDeductions
	fmt.Printf("\n💡 Expected TotalAmount: %12.2f\n", expected)
	fmt.Printf("   Difference:           %12.2f\n", expected - sale.TotalAmount)
	
	if expected != sale.TotalAmount {
		fmt.Println("\n⚠️  DATA CORRUPTION DETECTED!")
		fmt.Println("   TotalAmount in database is incorrect.")
	}
}
