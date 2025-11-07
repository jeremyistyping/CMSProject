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
	log.Println("🧪 Testing Journal Entry Analysis Fix")
	log.Println("=====================================")

	// Initialize database connection
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	testJournalEntryFilters(db)
}

func testJournalEntryFilters(db *gorm.DB) {
	log.Println("🔍 Testing Journal Entry Analysis with Filters")
	
	// Check available journal entries and their statuses
	var journalEntries []models.JournalEntry
	err := db.Find(&journalEntries).Error
	if err != nil {
		log.Printf("❌ Error fetching journal entries: %v", err)
		return
	}
	
	log.Printf("📊 Available journal entries and their statuses:")
	statusCount := make(map[string]int)
	typeCount := make(map[string]int)
	for i, entry := range journalEntries {
		log.Printf("   %d. ID: %d, Status: %s, Type: %s, Date: %s", 
			i+1, entry.ID, entry.Status, entry.ReferenceType, entry.EntryDate.Format("2006-01-02"))
		statusCount[entry.Status]++
		typeCount[entry.ReferenceType]++
	}
	
	log.Printf("📈 Status Summary:")
	for status, count := range statusCount {
		log.Printf("   %s: %d entries", status, count)
	}
	
	log.Printf("📈 Type Summary:")
	for refType, count := range typeCount {
		log.Printf("   %s: %d entries", refType, count)
	}
	
	// Initialize Enhanced Report Service
	accountRepo := repositories.NewAccountRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	productRepo := repositories.NewProductRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	purchaseRepo := repositories.NewPurchaseRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	
	enhancedReportService := services.NewEnhancedReportService(db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo, nil)
	
	// Test with different filter scenarios
	startDate := time.Now().AddDate(0, 0, -30) // Last 30 days
	endDate := time.Now()
	
	log.Printf("📅 Testing period: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	// Test 1: All entries (no filter)
	log.Println("\n🔬 Test 1: All entries (no filter)")
	analysisAll, err := enhancedReportService.GenerateJournalEntryAnalysisWithFilters(startDate, endDate, "ALL", "ALL")
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		log.Printf("✅ Found %d entries", analysisAll.TotalEntries)
	}
	
	// Test 2: Only POSTED entries (this simulates frontend behavior)
	log.Println("\n🔬 Test 2: Only POSTED entries")
	analysisPosted, err := enhancedReportService.GenerateJournalEntryAnalysisWithFilters(startDate, endDate, "POSTED", "ALL")
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		log.Printf("✅ Found %d POSTED entries", analysisPosted.TotalEntries)
		if analysisPosted.TotalEntries == 0 {
			log.Println("   ⚠️  No POSTED entries found - this explains why frontend shows 'No Data Available'")
		}
	}
	
	// Test 3: Only DRAFT entries
	log.Println("\n🔬 Test 3: Only DRAFT entries")
	analysisDraft, err := enhancedReportService.GenerateJournalEntryAnalysisWithFilters(startDate, endDate, "DRAFT", "ALL")
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		log.Printf("✅ Found %d DRAFT entries", analysisDraft.TotalEntries)
	}
	
	// Test 4: Only PAYMENT type
	log.Println("\n🔬 Test 4: Only PAYMENT type entries")
	analysisPayment, err := enhancedReportService.GenerateJournalEntryAnalysisWithFilters(startDate, endDate, "ALL", "PAYMENT")
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		log.Printf("✅ Found %d PAYMENT entries", analysisPayment.TotalEntries)
	}
	
	// Test 5: POSTED PAYMENT entries
	log.Println("\n🔬 Test 5: POSTED PAYMENT entries")
	analysisPostedPayment, err := enhancedReportService.GenerateJournalEntryAnalysisWithFilters(startDate, endDate, "POSTED", "PAYMENT")
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		log.Printf("✅ Found %d POSTED PAYMENT entries", analysisPostedPayment.TotalEntries)
	}
	
	// Conclusion and recommendation
	log.Println("\n📋 CONCLUSION:")
	log.Printf("   - Total available entries: %d", len(journalEntries))
	log.Printf("   - Entries with ALL filter: %d", analysisAll.TotalEntries)
	log.Printf("   - Entries with POSTED filter: %d", analysisPosted.TotalEntries)
	log.Printf("   - Entries with DRAFT filter: %d", analysisDraft.TotalEntries)
	
	if analysisPosted.TotalEntries == 0 && analysisAll.TotalEntries > 0 {
		log.Println("\n🎯 RECOMMENDATION:")
		log.Println("   The frontend should change from status='POSTED' to status='ALL'")
		log.Println("   or remove the status filter entirely to show all journal entries.")
	}
	
	log.Println("\n🏁 Test completed!")
}