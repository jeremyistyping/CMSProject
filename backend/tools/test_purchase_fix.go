package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 Testing Purchase Report Fixes")
	fmt.Println("================================")
	
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
	
	ctx := context.Background()
	
	fmt.Println("📊 Testing Purchase Report with Fixes...")
	report, err := purchaseService.GeneratePurchaseReport(ctx, startDate, endDate)
	if err != nil {
		log.Fatal("Failed to generate purchase report:", err)
	}
	
	// Display key fixes
	fmt.Println("\n🏆 PURCHASE REPORT FIXES VALIDATION")
	fmt.Println("===================================")
	
	fmt.Printf("Period: %s to %s\n", 
		report.StartDate.Format("2006-01-02"), 
		report.EndDate.Format("2006-01-02"))
	
	fmt.Println("\n💰 FINANCIAL SUMMARY (FIXED)")
	fmt.Println("============================")
	fmt.Printf("Total Purchases: %d\n", report.TotalPurchases)
	fmt.Printf("Total Amount: Rp %.0f\n", report.TotalAmount)
	fmt.Printf("Total Paid: Rp %.0f ← FIXED (should show cash payment)\n", report.TotalPaid)
	fmt.Printf("Outstanding Payables: Rp %.0f\n", report.OutstandingPayables)
	
	fmt.Println("\n🏢 VENDOR ANALYSIS (FIXED)")
	fmt.Println("=========================")
	if len(report.PurchasesByVendor) == 0 {
		fmt.Println("❌ No vendor data found (GROUP BY issue still exists)")
	} else {
		fmt.Println("✅ Vendor data found:")
		for i, vendor := range report.PurchasesByVendor {
			fmt.Printf("%d. Vendor: %s\n", i+1, vendor.VendorName)
			fmt.Printf("   Total Amount: Rp %.0f\n", vendor.TotalAmount)
			fmt.Printf("   Total Paid: Rp %.0f\n", vendor.TotalPaid)
			fmt.Printf("   Outstanding: Rp %.0f\n", vendor.Outstanding)
			fmt.Printf("   Payment Method: %s\n", vendor.PaymentMethod)
			fmt.Println()
		}
	}
	
	fmt.Println("\n💳 PAYMENT ANALYSIS (FIXED)")
	fmt.Println("===========================")
	fmt.Printf("Cash Purchases: %d (Amount: Rp %.0f)\n", 
		report.PaymentAnalysis.CashPurchases,
		report.PaymentAnalysis.CashAmount)
	fmt.Printf("Credit Purchases: %d (Amount: Rp %.0f)\n", 
		report.PaymentAnalysis.CreditPurchases,
		report.PaymentAnalysis.CreditAmount)
	fmt.Printf("Cash Percentage: %.1f%%\n", report.PaymentAnalysis.CashPercentage)
	fmt.Printf("Credit Percentage: %.1f%%\n", report.PaymentAnalysis.CreditPercentage)
	
	fmt.Println("\n🎯 VALIDATION RESULTS")
	fmt.Println("=====================")
	
	// Test 1: Cash purchase should be fully paid
	expectedCashPaid := report.PaymentAnalysis.CashAmount
	actualTotalPaid := report.TotalPaid
	
	fmt.Printf("✅ Test 1: Cash Purchase Payment Logic\n")
	fmt.Printf("   Expected Total Paid (Cash Amount): Rp %.0f\n", expectedCashPaid)
	fmt.Printf("   Actual Total Paid: Rp %.0f\n", actualTotalPaid)
	
	if actualTotalPaid == expectedCashPaid {
		fmt.Printf("   Result: ✅ PASS - Cash purchases are correctly marked as paid\n")
	} else {
		fmt.Printf("   Result: ❌ FAIL - Cash payment logic still needs fixing\n")
	}
	
	// Test 2: Outstanding should be Credit amount only
	expectedOutstanding := report.PaymentAnalysis.CreditAmount
	actualOutstanding := report.OutstandingPayables
	
	fmt.Printf("\n✅ Test 2: Outstanding Calculation\n")
	fmt.Printf("   Expected Outstanding (Credit Amount): Rp %.0f\n", expectedOutstanding)
	fmt.Printf("   Actual Outstanding: Rp %.0f\n", actualOutstanding)
	
	if actualOutstanding == expectedOutstanding {
		fmt.Printf("   Result: ✅ PASS - Outstanding correctly shows only credit purchases\n")
	} else {
		fmt.Printf("   Result: ❌ FAIL - Outstanding calculation still needs fixing\n")
	}
	
	// Test 3: Payment percentages should total 100%
	totalPercentage := report.PaymentAnalysis.CashPercentage + report.PaymentAnalysis.CreditPercentage
	
	fmt.Printf("\n✅ Test 3: Payment Percentage Totals\n")
	fmt.Printf("   Cash + Credit Percentage: %.1f%%\n", totalPercentage)
	
	if totalPercentage >= 99.9 && totalPercentage <= 100.1 {
		fmt.Printf("   Result: ✅ PASS - Percentages total 100%%\n")
	} else {
		fmt.Printf("   Result: ❌ FAIL - Percentage calculation error\n")
	}
	
	fmt.Println("\n🏆 OVERALL ASSESSMENT")
	fmt.Println("====================")
	
	passCount := 0
	if actualTotalPaid == expectedCashPaid { passCount++ }
	if actualOutstanding == expectedOutstanding { passCount++ }
	if totalPercentage >= 99.9 && totalPercentage <= 100.1 { passCount++ }
	
	fmt.Printf("Tests Passed: %d/3\n", passCount)
	
	if passCount == 3 {
		fmt.Println("🎉 ALL FIXES WORKING CORRECTLY!")
		fmt.Println("✅ Purchase Report is now CREDIBLE and ACCURATE")
	} else {
		fmt.Printf("⚠️ Still need to fix %d issue(s)\n", 3-passCount)
		if len(report.PurchasesByVendor) == 0 {
			fmt.Println("🔧 Next: Fix GROUP BY clause in vendor query")
		}
	}
}