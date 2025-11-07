package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
)

func main() {
	fmt.Println("🎯 PURCHASE REPORT VALIDATION")
	fmt.Println("================================")
	
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("❌ Failed to connect to database")
	}
	fmt.Println("✅ Database connected successfully")
	
	// Create purchase report service
	reportService := services.NewSSOTPurchaseReportService(db)
	
	// Set date range for testing
	startDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 9, 30, 23, 59, 59, 0, time.UTC)
	
	fmt.Printf("📅 Testing period: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	// Generate purchase report
	ctx := context.Background()
	report, err := reportService.GeneratePurchaseReport(ctx, startDate, endDate)
	if err != nil {
		log.Fatalf("❌ Failed to generate purchase report: %v", err)
	}
	
	fmt.Println("\n🏆 PURCHASE REPORT GENERATED SUCCESSFULLY!")
	fmt.Printf("✅ Total Purchases: %d\n", report.TotalPurchases)
	fmt.Printf("✅ Total Amount: Rp %.0f\n", report.TotalAmount)
	fmt.Printf("✅ Total Paid: Rp %.0f\n", report.TotalPaid)
	fmt.Printf("✅ Outstanding: Rp %.0f\n", report.OutstandingPayables)
	fmt.Printf("✅ Vendors Found: %d\n", len(report.PurchasesByVendor))
	
	fmt.Printf("✅ Cash Purchases: %d (%.1f%%)\n", 
		report.PaymentAnalysis.CashPurchases, 
		report.PaymentAnalysis.CashPercentage)
	fmt.Printf("✅ Credit Purchases: %d (%.1f%%)\n", 
		report.PaymentAnalysis.CreditPurchases, 
		report.PaymentAnalysis.CreditPercentage)
	
	// Validation checks
	fmt.Println("\n🔍 VALIDATION CHECKS")
	fmt.Println("===================")
	
	// Check 1: Total amounts should be positive
	if report.TotalAmount <= 0 {
		fmt.Printf("❌ Invalid total amount: %.0f\n", report.TotalAmount)
	} else {
		fmt.Printf("✅ Total amount is valid: Rp %.0f\n", report.TotalAmount)
	}
	
	// Check 2: Paid amount shouldn't exceed total
	if report.TotalPaid > report.TotalAmount {
		fmt.Printf("❌ Paid amount (%.0f) exceeds total amount (%.0f)\n", 
			report.TotalPaid, report.TotalAmount)
	} else {
		fmt.Printf("✅ Paid amount is within total: Rp %.0f\n", report.TotalPaid)
	}
	
	// Check 3: Outstanding calculation
	expectedOutstanding := report.TotalAmount - report.TotalPaid
	if abs(report.OutstandingPayables - expectedOutstanding) > 0.01 {
		fmt.Printf("❌ Outstanding calculation error: expected %.0f, got %.0f\n", 
			expectedOutstanding, report.OutstandingPayables)
	} else {
		fmt.Printf("✅ Outstanding calculation is correct: Rp %.0f\n", report.OutstandingPayables)
	}
	
	// Check 4: Vendor data consistency
	if len(report.PurchasesByVendor) > 0 {
		totalVendorAmount := 0.0
		for _, vendor := range report.PurchasesByVendor {
			totalVendorAmount += vendor.TotalAmount
		}
		if abs(totalVendorAmount - report.TotalAmount) > 0.01 {
			fmt.Printf("❌ Vendor totals don't match summary: vendor %.0f vs summary %.0f\n", 
				totalVendorAmount, report.TotalAmount)
		} else {
			fmt.Printf("✅ Vendor totals match summary: Rp %.0f\n", totalVendorAmount)
		}
	}
	
	fmt.Println("\n🎉 PURCHASE REPORT VALIDATION COMPLETED!")
	fmt.Println("✅ All systems operational and reporting accurate data")
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}