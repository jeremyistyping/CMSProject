package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"app-sistem-akuntansi/services"
)

func main() {
	fmt.Println("=== Testing Cash Flow Export Service ===")
	
	// Create export service
	exportService := services.NewCashFlowExportService()
	
	// Create sample Cash Flow data for testing
	sampleData := createSampleCashFlowData()
	
	fmt.Println("✓ Created sample cash flow data")
	fmt.Printf("  Company: %s\n", sampleData.Company.Name)
	fmt.Printf("  Period: %s to %s\n", sampleData.StartDate.Format("2006-01-02"), sampleData.EndDate.Format("2006-01-02"))
	fmt.Printf("  Net Cash Flow: %.2f\n", sampleData.NetCashFlow)
	
	// Test CSV Export
	fmt.Println("\n--- Testing CSV Export ---")
	csvData, err := exportService.ExportToCSV(sampleData)
	if err != nil {
		log.Printf("❌ CSV Export failed: %v", err)
	} else {
		fmt.Printf("✓ CSV Export successful, size: %d bytes\n", len(csvData))
		
		// Save to file for inspection
		csvFilename := exportService.GetCSVFilename(sampleData)
		err = os.WriteFile(csvFilename, csvData, 0644)
		if err != nil {
			log.Printf("❌ Failed to write CSV file: %v", err)
		} else {
			fmt.Printf("✓ CSV saved as: %s\n", csvFilename)
		}
	}
	
	// Test PDF Export
	fmt.Println("\n--- Testing PDF Export ---")
	pdfData, err := exportService.ExportToPDF(sampleData)
	if err != nil {
		log.Printf("❌ PDF Export failed: %v", err)
	} else {
		fmt.Printf("✓ PDF Export successful, size: %d bytes\n", len(pdfData))
		
		// Save to file for inspection
		pdfFilename := exportService.GetPDFFilename(sampleData)
		err = os.WriteFile(pdfFilename, pdfData, 0644)
		if err != nil {
			log.Printf("❌ Failed to write PDF file: %v", err)
		} else {
			fmt.Printf("✓ PDF saved as: %s\n", pdfFilename)
		}
	}
	
	fmt.Println("\n=== Test Complete ===")
	fmt.Println("Check the generated files to verify the export formats are correct.")
}

func createSampleCashFlowData() *services.SSOTCashFlowData {
	now := time.Now()
	startDate := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
	
	data := &services.SSOTCashFlowData{
		Company: services.CompanyInfo{
			Name:    "PT. Sistem Akuntansi Test",
			Address: "Jl. Testing No. 123",
			City:    "Jakarta",
			Phone:   "021-1234567",
			Email:   "test@example.com",
		},
		StartDate:   startDate,
		EndDate:     endDate,
		Currency:    "IDR",
		GeneratedAt: now,
		Enhanced:    true,
	}
	
	// Operating Activities
	data.OperatingActivities.NetIncome = -5000000
	data.OperatingActivities.Adjustments.Depreciation = 2000000
	data.OperatingActivities.Adjustments.Amortization = 500000
	data.OperatingActivities.Adjustments.BadDebtExpense = 100000
	data.OperatingActivities.Adjustments.TotalAdjustments = 2600000
	data.OperatingActivities.Adjustments.Items = []services.CFSectionItem{
		{
			AccountCode: "6201",
			AccountName: "Depreciation Expense",
			Amount:      2000000,
			Type:        "adjustment",
		},
		{
			AccountCode: "6202",
			AccountName: "Amortization Expense", 
			Amount:      500000,
			Type:        "adjustment",
		},
		{
			AccountCode: "6203",
			AccountName: "Bad Debt Expense",
			Amount:      100000,
			Type:        "adjustment",
		},
	}
	
	data.OperatingActivities.WorkingCapitalChanges.AccountsReceivableChange = -1000000
	data.OperatingActivities.WorkingCapitalChanges.InventoryChange = 500000
	data.OperatingActivities.WorkingCapitalChanges.AccountsPayableChange = 800000
	data.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges = 300000
	data.OperatingActivities.WorkingCapitalChanges.Items = []services.CFSectionItem{
		{
			AccountCode: "1201",
			AccountName: "Accounts Receivable",
			Amount:      -1000000,
			Type:        "increase",
		},
		{
			AccountCode: "1301",
			AccountName: "Inventory",
			Amount:      500000,
			Type:        "decrease",
		},
		{
			AccountCode: "2101",
			AccountName: "Accounts Payable",
			Amount:      800000,
			Type:        "increase",
		},
	}
	
	data.OperatingActivities.TotalOperatingCashFlow = data.OperatingActivities.NetIncome + 
		data.OperatingActivities.Adjustments.TotalAdjustments + 
		data.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges
	
	// Investing Activities
	data.InvestingActivities.PurchaseOfFixedAssets = -3000000
	data.InvestingActivities.SaleOfFixedAssets = 500000
	data.InvestingActivities.TotalInvestingCashFlow = -2500000
	data.InvestingActivities.Items = []services.CFSectionItem{
		{
			AccountCode: "1601",
			AccountName: "Purchase of Equipment",
			Amount:      -3000000,
			Type:        "outflow",
		},
		{
			AccountCode: "1602",
			AccountName: "Sale of Old Equipment",
			Amount:      500000,
			Type:        "inflow",
		},
	}
	
	// Financing Activities
	data.FinancingActivities.ShareCapitalIncrease = 5000000
	data.FinancingActivities.DividendsPaid = -1000000
	data.FinancingActivities.TotalFinancingCashFlow = 4000000
	data.FinancingActivities.Items = []services.CFSectionItem{
		{
			AccountCode: "3101",
			AccountName: "Share Capital Increase",
			Amount:      5000000,
			Type:        "inflow",
		},
		{
			AccountCode: "3201",
			AccountName: "Dividends Paid",
			Amount:      -1000000,
			Type:        "outflow",
		},
	}
	
	// Summary calculations
	data.NetCashFlow = data.OperatingActivities.TotalOperatingCashFlow + 
		data.InvestingActivities.TotalInvestingCashFlow + 
		data.FinancingActivities.TotalFinancingCashFlow
	data.CashAtBeginning = 2000000
	data.CashAtEnd = data.CashAtBeginning + data.NetCashFlow
	
	// Cash Flow Ratios
	data.CashFlowRatios.OperatingCashFlowRatio = 0.85
	data.CashFlowRatios.CashFlowToDebtRatio = 0.45
	data.CashFlowRatios.FreeCashFlow = data.OperatingActivities.TotalOperatingCashFlow + data.InvestingActivities.PurchaseOfFixedAssets
	
	// Account Details for drilldown
	data.AccountDetails = []services.SSOTAccountBalance{
		{
			AccountID:     1,
			AccountCode:   "1101",
			AccountName:   "Cash",
			AccountType:   "ASSET",
			DebitTotal:    float64(data.CashAtEnd),
			CreditTotal:   0,
			NetBalance:    float64(data.CashAtEnd),
		},
	}
	
	return data
}