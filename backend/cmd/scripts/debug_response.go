package main

import (
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"gorm.io/gorm"
)

func main() {
	log.Println("🧪 Debugging Journal Entry Analysis Response")
	log.Println("==============================================")

	// Initialize database connection
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	debugJournalAnalysisResponse(db)
}

func debugJournalAnalysisResponse(db *gorm.DB) {
	log.Println("🔍 Debugging Journal Entry Analysis Response Structure")

	// Initialize Enhanced Report Service
	accountRepo := repositories.NewAccountRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	productRepo := repositories.NewProductRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	purchaseRepo := repositories.NewPurchaseRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)

	enhancedReportService := services.NewEnhancedReportService(db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo, nil)

	// Test the exact same parameters that frontend sends (without status)
	startDate := time.Date(2025, 8, 31, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 9, 19, 0, 0, 0, 0, time.UTC)

	log.Printf("📅 Testing with exact frontend parameters:")
	log.Printf("   Start Date: %s", startDate.Format("2006-01-02"))
	log.Printf("   End Date: %s", endDate.Format("2006-01-02"))
	log.Printf("   Status: ALL (default)")

	// Call with default parameters (status = "ALL")
	analysisData, err := enhancedReportService.GenerateJournalEntryAnalysisWithFilters(startDate, endDate, "ALL", "ALL")
	if err != nil {
		log.Printf("❌ Error generating analysis: %v", err)
		return
	}

	// Show the complete response structure that backend returns
	log.Printf("✅ Analysis generated successfully!")
	log.Printf("📊 Response Analysis:")
	log.Printf("   Company Name: %s", analysisData.Company.Name)
	log.Printf("   Currency: %s", analysisData.Currency)
	log.Printf("   Start Date: %s", analysisData.StartDate.Format("2006-01-02"))
	log.Printf("   End Date: %s", analysisData.EndDate.Format("2006-01-02"))
	log.Printf("   Total Entries: %d", analysisData.TotalEntries)
	log.Printf("   Total Debit Amount: %.2f", analysisData.TotalDebitAmount)
	log.Printf("   Total Credit Amount: %.2f", analysisData.TotalCreditAmount)

	// This is the key check - if TotalEntries > 0, then data exists
	if analysisData.TotalEntries == 0 {
		log.Println("❌ PROBLEM: TotalEntries is 0 - this is why frontend shows 'No Data Available'")
		
		// Debug raw journal entries
		var journalEntries []models.JournalEntry
		err = db.Where("entry_date BETWEEN ? AND ?", startDate, endDate).Find(&journalEntries).Error
		if err != nil {
			log.Printf("❌ Error fetching raw entries: %v", err)
		} else {
			log.Printf("🔬 Raw journal entries in date range: %d", len(journalEntries))
			for i, entry := range journalEntries {
				log.Printf("   %d. ID: %d, Date: %s, Status: %s, Type: %s", 
					i+1, entry.ID, entry.EntryDate.Format("2006-01-02"), entry.Status, entry.ReferenceType)
			}
		}
	} else {
		log.Printf("✅ GOOD: TotalEntries = %d, frontend should show data", analysisData.TotalEntries)

		// Show entries by type
		log.Printf("📋 Entries by Type (%d types):", len(analysisData.EntriesByType))
		for i, typeData := range analysisData.EntriesByType {
			log.Printf("   %d. %s: %d entries (%.1f%%) - Rp %.2f", 
				i+1, typeData.ReferenceType, typeData.Count, typeData.Percentage, typeData.TotalAmount)
		}

		// Show entries by status
		log.Printf("📋 Entries by Status (%d statuses):", len(analysisData.EntriesByStatus))
		for i, statusData := range analysisData.EntriesByStatus {
			log.Printf("   %d. %s: %d entries (%.1f%%) - Rp %.2f", 
				i+1, statusData.Status, statusData.Count, statusData.Percentage, statusData.TotalAmount)
		}

		// Show recent entries
		log.Printf("📄 Recent Entries (%d entries):", len(analysisData.RecentEntries))
		for i, entry := range analysisData.RecentEntries {
			if i < 3 { // Show first 3
				log.Printf("   %d. %s: %s (Debit: %.2f, Credit: %.2f, Status: %s)", 
					i+1, entry.Date.Format("2006-01-02"), entry.Description, 
					entry.DebitAmount, entry.CreditAmount, entry.Status)
			}
		}

		// Show compliance check
		compliance := analysisData.ComplianceCheck
		log.Printf("🔍 Compliance Check:")
		log.Printf("   Balanced: %d", compliance.BalancedEntries)
		log.Printf("   Unbalanced: %d", compliance.UnbalancedEntries)
		log.Printf("   Rate: %.1f%%", compliance.ComplianceRate)
		log.Printf("   Missing Refs: %d", compliance.MissingReferences)
		log.Printf("   Issues: %v", compliance.ComplianceIssues)
	}

	// Conclusion
	log.Println("\n📋 CONCLUSION:")
	if analysisData.TotalEntries > 0 {
		log.Println("   ✅ Backend is returning valid data")
		log.Println("   ✅ The backend fix is working correctly")
		log.Println("   ⚠️  If frontend still shows 'No Data Available', the issue is in:")
		log.Println("      - Frontend not restarted with updated code")
		log.Println("      - Frontend parsing issue")
		log.Println("      - Browser cache issue")
		log.Println("   💡 SOLUTION: Restart frontend and clear browser cache")
	} else {
		log.Println("   ❌ Backend is not returning data - need to investigate further")
	}

	log.Println("\n🏁 Debug completed!")
}