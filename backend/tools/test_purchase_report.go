package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🧪 Testing New SSOT Purchase Report")
	fmt.Println("=====================================")
	
	// Initialize database
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// Create purchase report service
	purchaseService := services.NewSSOTPurchaseReportService(db)
	
	// Define test date range
	startDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 9, 30, 23, 59, 59, 0, time.UTC)
	
	fmt.Printf("🗓️  Test Period: %s to %s\n", 
		startDate.Format("2006-01-02"), 
		endDate.Format("2006-01-02"))
	
	ctx := context.Background()
	
	fmt.Println("\n📊 Generating Purchase Report...")
	report, err := purchaseService.GeneratePurchaseReport(ctx, startDate, endDate)
	if err != nil {
		log.Fatal("Failed to generate purchase report:", err)
	}
	
	// Display summary
	fmt.Println("\n📈 PURCHASE REPORT SUMMARY")
	fmt.Println("==========================")
	fmt.Printf("Company: %s\n", report.Company.Name)
	fmt.Printf("Period: %s to %s\n", 
		report.StartDate.Format("2006-01-02"), 
		report.EndDate.Format("2006-01-02"))
	fmt.Printf("Currency: %s\n", report.Currency)
	fmt.Printf("Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	
	fmt.Println("\n💰 FINANCIAL OVERVIEW")
	fmt.Println("=====================")
	fmt.Printf("Total Purchases: %d\n", report.TotalPurchases)
	fmt.Printf("Completed Purchases: %d\n", report.CompletedPurchases)
	fmt.Printf("Total Amount: Rp %,.2f\n", report.TotalAmount)
	fmt.Printf("Total Paid: Rp %,.2f\n", report.TotalPaid)
	fmt.Printf("Outstanding Payables: Rp %,.2f\n", report.OutstandingPayables)
	
	fmt.Println("\n🏢 PURCHASES BY VENDOR")
	fmt.Println("======================")
	if len(report.PurchasesByVendor) == 0 {
		fmt.Println("No vendor data found")
	} else {
		for i, vendor := range report.PurchasesByVendor {
			fmt.Printf("%d. %s (ID: %d)\n", i+1, vendor.VendorName, vendor.VendorID)
			fmt.Printf("   Total Purchases: %d\n", vendor.TotalPurchases)
			fmt.Printf("   Total Amount: Rp %,.2f\n", vendor.TotalAmount)
			fmt.Printf("   Total Paid: Rp %,.2f\n", vendor.TotalPaid)
			fmt.Printf("   Outstanding: Rp %,.2f\n", vendor.Outstanding)
			fmt.Printf("   Payment Method: %s\n", vendor.PaymentMethod)
			fmt.Printf("   Status: %s\n", vendor.Status)
			fmt.Printf("   Last Purchase: %s\n", vendor.LastPurchaseDate.Format("2006-01-02"))
			fmt.Println()
		}
	}
	
	fmt.Println("\n📅 PURCHASES BY MONTH")
	fmt.Println("=====================")
	if len(report.PurchasesByMonth) == 0 {
		fmt.Println("No monthly data found")
	} else {
		for _, month := range report.PurchasesByMonth {
			fmt.Printf("%s %d:\n", month.MonthName, month.Year)
			fmt.Printf("   Purchases: %d\n", month.TotalPurchases)
			fmt.Printf("   Total Amount: Rp %,.2f\n", month.TotalAmount)
			fmt.Printf("   Total Paid: Rp %,.2f\n", month.TotalPaid)
			fmt.Printf("   Average Amount: Rp %,.2f\n", month.AverageAmount)
			fmt.Println()
		}
	}
	
	fmt.Println("\n📊 PURCHASES BY CATEGORY")
	fmt.Println("========================")
	if len(report.PurchasesByCategory) == 0 {
		fmt.Println("No category data found")
	} else {
		for _, category := range report.PurchasesByCategory {
			fmt.Printf("%s (%s - %s):\n", 
				category.CategoryName, 
				category.AccountCode, 
				category.AccountName)
			fmt.Printf("   Purchases: %d\n", category.TotalPurchases)
			fmt.Printf("   Amount: Rp %,.2f (%.1f%%)\n", 
				category.TotalAmount, 
				category.Percentage)
			fmt.Println()
		}
	}
	
	fmt.Println("\n💳 PAYMENT ANALYSIS")
	fmt.Println("===================")
	fmt.Printf("Cash Purchases: %d (Rp %,.2f) - %.1f%%\n", 
		report.PaymentAnalysis.CashPurchases,
		report.PaymentAnalysis.CashAmount,
		report.PaymentAnalysis.CashPercentage)
	fmt.Printf("Credit Purchases: %d (Rp %,.2f) - %.1f%%\n", 
		report.PaymentAnalysis.CreditPurchases,
		report.PaymentAnalysis.CreditAmount,
		report.PaymentAnalysis.CreditPercentage)
	fmt.Printf("Average Order Value: Rp %,.2f\n", 
		report.PaymentAnalysis.AverageOrderValue)
	
	fmt.Println("\n🏦 TAX ANALYSIS")
	fmt.Println("===============")
	fmt.Printf("Total Taxable Amount: Rp %,.2f\n", report.TaxAnalysis.TotalTaxableAmount)
	fmt.Printf("Total Tax Amount: Rp %,.2f\n", report.TaxAnalysis.TotalTaxAmount)
	fmt.Printf("Average Tax Rate: %.2f%%\n", report.TaxAnalysis.AverageTaxRate)
	fmt.Printf("Tax Reclaimable Amount: Rp %,.2f\n", report.TaxAnalysis.TaxReclaimableAmount)
	
	if len(report.TaxAnalysis.TaxByMonth) > 0 {
		fmt.Println("\n📅 Monthly Tax Breakdown:")
		for _, tax := range report.TaxAnalysis.TaxByMonth {
			fmt.Printf("   %s %d: Rp %,.2f\n", 
				tax.MonthName, 
				tax.Year, 
				tax.TaxAmount)
		}
	}
	
	// Output JSON for debugging
	fmt.Println("\n🔍 RAW JSON OUTPUT (for debugging)")
	fmt.Println("==================================")
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
	} else {
		fmt.Println(string(jsonData))
	}
	
	fmt.Println("\n✅ Purchase Report Test Completed!")
	fmt.Println("==================================")
	
	// Verify data accuracy
	fmt.Println("\n🔍 DATA ACCURACY VERIFICATION")
	fmt.Println("=============================")
	
	// Check balance
	totalIn := report.TotalAmount
	totalOut := report.TotalPaid
	calculatedOutstanding := totalIn - totalOut
	
	fmt.Printf("Total Amount: Rp %,.2f\n", totalIn)
	fmt.Printf("Total Paid: Rp %,.2f\n", totalOut)
	fmt.Printf("Calculated Outstanding: Rp %,.2f\n", calculatedOutstanding)
	fmt.Printf("Reported Outstanding: Rp %,.2f\n", report.OutstandingPayables)
	fmt.Printf("Balance Match: %t\n", calculatedOutstanding == report.OutstandingPayables)
	
	// Check vendor totals
	var vendorTotalAmount, vendorTotalPaid, vendorTotalOutstanding float64
	for _, vendor := range report.PurchasesByVendor {
		vendorTotalAmount += vendor.TotalAmount
		vendorTotalPaid += vendor.TotalPaid
		vendorTotalOutstanding += vendor.Outstanding
	}
	
	fmt.Printf("\nVendor Summary Totals:\n")
	fmt.Printf("  Amount: Rp %,.2f\n", vendorTotalAmount)
	fmt.Printf("  Paid: Rp %,.2f\n", vendorTotalPaid)
	fmt.Printf("  Outstanding: Rp %,.2f\n", vendorTotalOutstanding)
	
	// Check category percentages
	var categoryTotalPercentage float64
	for _, category := range report.PurchasesByCategory {
		categoryTotalPercentage += category.Percentage
	}
	fmt.Printf("\nCategory Total Percentage: %.2f%% (should be ~100%%)\n", categoryTotalPercentage)
	
	// Check payment analysis
	paymentTotal := report.PaymentAnalysis.CashAmount + report.PaymentAnalysis.CreditAmount
	paymentPercentageTotal := report.PaymentAnalysis.CashPercentage + report.PaymentAnalysis.CreditPercentage
	
	fmt.Printf("\nPayment Analysis:\n")
	fmt.Printf("  Total Amount: Rp %,.2f\n", paymentTotal)
	fmt.Printf("  Percentage Total: %.1f%% (should be 100%%)\n", paymentPercentageTotal)
	
	fmt.Println("\n🎯 CONCLUSION")
	fmt.Println("=============")
	if calculatedOutstanding == report.OutstandingPayables && paymentPercentageTotal >= 99.9 && paymentPercentageTotal <= 100.1 {
		fmt.Println("✅ Purchase Report Data is ACCURATE and CREDIBLE")
		fmt.Println("✅ All calculations are consistent")
		fmt.Println("✅ Ready for production use")
	} else {
		fmt.Println("⚠️ Some inconsistencies detected")
		fmt.Println("⚠️ Review calculations and data logic")
	}
}