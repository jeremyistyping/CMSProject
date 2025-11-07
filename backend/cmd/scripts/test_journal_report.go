package main

import (
	"encoding/json"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"gorm.io/gorm"
)

func main() {
	log.Println("🧪 Testing Journal Entry Analysis Report")
	log.Println("=========================================")

	// Initialize database connection
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	testJournalEntryReport(db)
}

func testJournalEntryReport(db *gorm.DB) {
	log.Println("🔍 Testing Journal Entry Analysis Implementation")
	
	// 1. First check if we have any journal entries
	var journalCount int64
	db.Model(&models.JournalEntry{}).Count(&journalCount)
	log.Printf("📊 Total journal entries in database: %d", journalCount)
	
	if journalCount == 0 {
		log.Println("⚠️  No journal entries found in database!")
		return
	}
	
	// 2. Check recent journal entries
	var recentEntries []models.JournalEntry
	err := db.Preload("Creator").
		Preload("JournalLines").
		Order("created_at DESC").
		Limit(5).
		Find(&recentEntries).Error
	
	if err != nil {
		log.Printf("❌ Error fetching recent journal entries: %v", err)
		return
	}
	
	log.Printf("📄 Recent journal entries (%d):", len(recentEntries))
	for i, entry := range recentEntries {
		log.Printf("   %d. ID: %d, Date: %s, Description: %s, Debit: %.2f, Credit: %.2f", 
			i+1, entry.ID, entry.EntryDate.Format("2006-01-02"), entry.Description, 
			entry.TotalDebit, entry.TotalCredit)
	}
	
	// 3. Initialize EnhancedReportService
	log.Println("🚀 Initializing Enhanced Report Service...")
	
	// Initialize necessary repositories
	accountRepo := repositories.NewAccountRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	productRepo := repositories.NewProductRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	purchaseRepo := repositories.NewPurchaseRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	
	enhancedReportService := services.NewEnhancedReportService(db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo, nil)
	
	// 4. Test GenerateJournalEntryAnalysis function
	startDate := time.Now().AddDate(0, 0, -30) // Last 30 days
	endDate := time.Now()
	
	log.Printf("📅 Testing journal entry analysis from %s to %s", 
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	analysisData, err := enhancedReportService.GenerateJournalEntryAnalysis(startDate, endDate)
	if err != nil {
		log.Printf("❌ Error generating journal entry analysis: %v", err)
		return
	}
	
	// 5. Display results
	log.Printf("✅ Journal Entry Analysis Generated Successfully!")
	log.Printf("📊 Analysis Results:")
	log.Printf("   Company: %s", analysisData.Company.Name)
	log.Printf("   Period: %s to %s", 
		analysisData.StartDate.Format("2006-01-02"), 
		analysisData.EndDate.Format("2006-01-02"))
	log.Printf("   Total Entries: %d", analysisData.TotalEntries)
	log.Printf("   Total Debit Amount: Rp %.2f", analysisData.TotalDebitAmount)
	log.Printf("   Total Credit Amount: Rp %.2f", analysisData.TotalCreditAmount)
	
	// Compliance check
	log.Printf("🔍 Compliance Check:")
	log.Printf("   Balanced Entries: %d", analysisData.ComplianceCheck.BalancedEntries)
	log.Printf("   Unbalanced Entries: %d", analysisData.ComplianceCheck.UnbalancedEntries)
	log.Printf("   Compliance Rate: %.2f%%", analysisData.ComplianceCheck.ComplianceRate)
	log.Printf("   Missing References: %d", analysisData.ComplianceCheck.MissingReferences)
	
	// Entries by type
	log.Printf("📋 Entries by Type:")
	for i, typeData := range analysisData.EntriesByType {
		if i < 5 { // Show first 5
			log.Printf("   %s: %d entries (%.2f%%) - Rp %.2f", 
				typeData.ReferenceType, typeData.Count, typeData.Percentage, typeData.TotalAmount)
		}
	}
	
	// Recent entries
	log.Printf("📄 Recent Entries (%d):", len(analysisData.RecentEntries))
	for i, entry := range analysisData.RecentEntries {
		if i < 3 { // Show first 3
			log.Printf("   %d. %s: %s (Rp %.2f)", 
				i+1, entry.Date.Format("2006-01-02"), entry.Description, entry.DebitAmount)
		}
	}
	
	// 6. Check if we have data to return
	if analysisData.TotalEntries == 0 {
		log.Println("❌ ERROR: Analysis shows 0 total entries - this is the bug!")
		log.Println("   The query might not be filtering correctly by date range")
		
		// Debug the query
		debugJournalEntryQuery(db, startDate, endDate)
	} else {
		log.Printf("✅ SUCCESS: Analysis returned %d entries", analysisData.TotalEntries)
		
		// Pretty print the JSON response for debugging
		jsonData, _ := json.MarshalIndent(analysisData, "", "  ")
		log.Printf("📄 Full JSON Response Preview (first 1000 chars):")
		responseStr := string(jsonData)
		if len(responseStr) > 1000 {
			responseStr = responseStr[:1000] + "..."
		}
		log.Printf("%s", responseStr)
	}
	
	log.Println("🏁 Journal Entry Analysis Test Completed!")
}

func debugJournalEntryQuery(db *gorm.DB, startDate, endDate time.Time) {
	log.Println("🐛 Debugging journal entry query...")
	
	// Test raw query without date filter first
	var allEntries []models.JournalEntry
	err := db.Find(&allEntries).Error
	if err != nil {
		log.Printf("❌ Error fetching all entries: %v", err)
		return
	}
	log.Printf("📊 Total entries (no date filter): %d", len(allEntries))
	
	if len(allEntries) > 0 {
		log.Printf("📅 Sample entry dates:")
		for i, entry := range allEntries {
			if i < 5 { // Show first 5
				log.Printf("   Entry %d: ID=%d, EntryDate=%s, CreatedAt=%s", 
					i+1, entry.ID, entry.EntryDate.Format("2006-01-02 15:04:05"), 
					entry.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		}
		
		// Now test with date filter
		var filteredEntries []models.JournalEntry
		err = db.Where("entry_date BETWEEN ? AND ?", startDate, endDate).Find(&filteredEntries).Error
		if err != nil {
			log.Printf("❌ Error with date filter: %v", err)
			return
		}
		log.Printf("📊 Entries with date filter (%s to %s): %d", 
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), len(filteredEntries))
		
		if len(filteredEntries) == 0 {
			log.Printf("⚠️  Date filter is excluding all entries!")
			log.Printf("   Start date: %s", startDate.Format("2006-01-02 15:04:05"))
			log.Printf("   End date: %s", endDate.Format("2006-01-02 15:04:05"))
			
			// Try with broader date range
			broaderStart := time.Now().AddDate(-1, 0, 0) // 1 year ago
			broaderEnd := time.Now().AddDate(0, 0, 1)    // Tomorrow
			
			var broaderEntries []models.JournalEntry
			err = db.Where("entry_date BETWEEN ? AND ?", broaderStart, broaderEnd).Find(&broaderEntries).Error
			if err == nil {
				log.Printf("📊 Entries with broader date range (1 year): %d", len(broaderEntries))
			}
		}
	}
}