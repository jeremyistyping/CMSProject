package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatalf("Failed to connect to database")
	}

	fmt.Println("🔍 Checking Report Data in Database")
	fmt.Println("=====================================")

	// Check journal entries
	checkJournalEntries(db)
	
	// Check journal lines  
	checkJournalLines(db)
	
	// Check accounts with balances
	checkAccountsWithBalances(db)
	
	// Check sales data
	checkSalesData(db)
	
	// Check purchases data
	checkPurchasesData(db)

	fmt.Println("\n✅ Database check completed!")
}

func checkJournalEntries(db *gorm.DB) {
	fmt.Println("\n📊 Journal Entries:")
	
	var count int64
	db.Model(&models.JournalEntry{}).Count(&count)
	fmt.Printf("   Total entries: %d\n", count)
	
	var postedCount int64
	db.Model(&models.JournalEntry{}).Where("status = ?", models.JournalStatusPosted).Count(&postedCount)
	fmt.Printf("   Posted entries: %d\n", postedCount)
	
	var draftCount int64
	db.Model(&models.JournalEntry{}).Where("status = ?", models.JournalStatusDraft).Count(&draftCount)
	fmt.Printf("   Draft entries: %d\n", draftCount)
	
	// Check recent entries
	var recentEntries []models.JournalEntry
	db.Where("status = ?", models.JournalStatusPosted).
		Order("entry_date DESC").
		Limit(5).
		Find(&recentEntries)
	
	fmt.Println("   Recent Posted Entries:")
	for _, entry := range recentEntries {
		fmt.Printf("     - %s: %s (%s) - Amount: %.2f\n", 
			entry.EntryDate.Format("2006-01-02"), 
			entry.Code, 
			entry.Description,
			entry.TotalDebit)
	}
}

func checkJournalLines(db *gorm.DB) {
	fmt.Println("\n📝 Journal Lines:")
	
	var count int64
	db.Model(&models.JournalLine{}).Count(&count)
	fmt.Printf("   Total lines: %d\n", count)
	
	// Check lines with journal entries that are posted
	var postedLinesCount int64
	db.Table("journal_lines").
		Joins("JOIN journal_entries ON journal_lines.journal_entry_id = journal_entries.id").
		Where("journal_entries.status = ?", models.JournalStatusPosted).
		Count(&postedLinesCount)
	fmt.Printf("   Lines in posted entries: %d\n", postedLinesCount)
	
	// Check total amounts
	var totalDebit, totalCredit float64
	db.Table("journal_lines").
		Joins("JOIN journal_entries ON journal_lines.journal_entry_id = journal_entries.id").
		Where("journal_entries.status = ?", models.JournalStatusPosted).
		Select("COALESCE(SUM(debit_amount), 0) as total_debit, COALESCE(SUM(credit_amount), 0) as total_credit").
		Row().Scan(&totalDebit, &totalCredit)
	
	fmt.Printf("   Total Debit: %.2f\n", totalDebit)
	fmt.Printf("   Total Credit: %.2f\n", totalCredit)
	fmt.Printf("   Balance: %.2f\n", totalDebit - totalCredit)
}

func checkAccountsWithBalances(db *gorm.DB) {
	fmt.Println("\n💰 Account Balances:")
	
	var accounts []models.Account
	db.Where("is_active = ?", true).Order("code").Find(&accounts)
	
	fmt.Printf("   Total active accounts: %d\n", len(accounts))
	
	// Calculate balances for each account type
	assetCount := 0
	liabilityCount := 0
	equityCount := 0
	revenueCount := 0
	expenseCount := 0
	
	for _, account := range accounts {
		// Calculate balance for this account
		var totalDebit, totalCredit float64
		db.Table("journal_lines").
			Joins("JOIN journal_entries ON journal_lines.journal_entry_id = journal_entries.id").
			Where("journal_lines.account_id = ? AND journal_entries.status = ?",
				account.ID, models.JournalStatusPosted).
			Select("COALESCE(SUM(journal_lines.debit_amount), 0) as total_debit, COALESCE(SUM(journal_lines.credit_amount), 0) as total_credit").
			Row().Scan(&totalDebit, &totalCredit)
		
		balance := 0.0
		if account.GetNormalBalance() == models.NormalBalanceDebit {
			balance = totalDebit - totalCredit
		} else {
			balance = totalCredit - totalDebit
		}
		
		// Only show accounts with non-zero balance
		if balance != 0 {
			switch account.Type {
			case models.AccountTypeAsset:
				assetCount++
			case models.AccountTypeLiability:
				liabilityCount++
			case models.AccountTypeEquity:
				equityCount++
			case models.AccountTypeRevenue:
				revenueCount++
			case models.AccountTypeExpense:
				expenseCount++
			}
			
			fmt.Printf("     %s - %s: %.2f (Type: %s)\n", 
				account.Code, account.Name, balance, account.Type)
		}
	}
	
	fmt.Printf("   Accounts with balances by type:\n")
	fmt.Printf("     Assets: %d\n", assetCount)
	fmt.Printf("     Liabilities: %d\n", liabilityCount)
	fmt.Printf("     Equity: %d\n", equityCount)
	fmt.Printf("     Revenue: %d\n", revenueCount)
	fmt.Printf("     Expenses: %d\n", expenseCount)
}

func checkSalesData(db *gorm.DB) {
	fmt.Println("\n💼 Sales Data:")
	
	var count int64
	db.Model(&models.Sale{}).Count(&count)
	fmt.Printf("   Total sales: %d\n", count)
	
	var confirmedCount int64
	db.Model(&models.Sale{}).Where("status = ?", "CONFIRMED").Count(&confirmedCount)
	fmt.Printf("   Confirmed sales: %d\n", confirmedCount)
	
	var totalAmount float64
	db.Model(&models.Sale{}).Where("status = ?", "CONFIRMED").Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&totalAmount)
	fmt.Printf("   Total confirmed sales amount: %.2f\n", totalAmount)
}

func checkPurchasesData(db *gorm.DB) {
	fmt.Println("\n🛒 Purchase Data:")
	
	var count int64
	db.Model(&models.Purchase{}).Count(&count)
	fmt.Printf("   Total purchases: %d\n", count)
	
	var approvedCount int64
	db.Model(&models.Purchase{}).Where("status = ?", "APPROVED").Count(&approvedCount)
	fmt.Printf("   Approved purchases: %d\n", approvedCount)
	
	var totalAmount float64
	db.Model(&models.Purchase{}).Where("status = ?", "APPROVED").Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&totalAmount)
	fmt.Printf("   Total approved purchase amount: %.2f\n", totalAmount)
}