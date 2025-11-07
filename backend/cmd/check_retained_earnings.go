package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Println("Checking Retained Earnings and Closing Journal...")
	
	// Connect to database
	db := database.ConnectDB()
	
	// 1. Check Retained Earnings Account
	var retainedEarnings models.Account
	err := db.Where("code = ? OR name LIKE ?", "3201", "%Laba Ditahan%").First(&retainedEarnings).Error
	if err != nil {
		log.Printf("Error finding Retained Earnings account: %v", err)
	} else {
		fmt.Printf("\n=== RETAINED EARNINGS ACCOUNT ===\n")
		fmt.Printf("ID: %d\n", retainedEarnings.ID)
		fmt.Printf("Code: %s\n", retainedEarnings.Code)
		fmt.Printf("Name: %s\n", retainedEarnings.Name)
		fmt.Printf("Type: %s\n", retainedEarnings.Type)
		fmt.Printf("Balance: %.2f\n", retainedEarnings.Balance)
	}
	
	// 2. Check if there are any closing journals
	var closingJournals []models.JournalEntry
	db.Where("reference_type = ? OR code LIKE ?", "CLOSING", "CLO-%").
		Order("entry_date DESC").
		Limit(5).
		Find(&closingJournals)
	
	fmt.Printf("\n=== CLOSING JOURNALS (Last 5) ===\n")
	if len(closingJournals) == 0 {
		fmt.Println("No closing journals found!")
	} else {
		for i, journal := range closingJournals {
			fmt.Printf("\n%d. Journal ID: %d\n", i+1, journal.ID)
			fmt.Printf("   Code: %s\n", journal.Code)
			fmt.Printf("   Date: %s\n", journal.EntryDate.Format("2006-01-02"))
			fmt.Printf("   Description: %s\n", journal.Description)
			fmt.Printf("   Status: %s\n", journal.Status)
			fmt.Printf("   Total Debit: %.2f\n", journal.TotalDebit)
			fmt.Printf("   Total Credit: %.2f\n", journal.TotalCredit)
			fmt.Printf("   Is Balanced: %v\n", journal.IsBalanced)
			
			// Get journal lines for this closing entry
			var lines []models.JournalLine
			db.Where("journal_entry_id = ?", journal.ID).Find(&lines)
			
			fmt.Printf("   Journal Lines (%d):\n", len(lines))
			for _, line := range lines {
				var account models.Account
				db.First(&account, line.AccountID)
				fmt.Printf("     - %s (%s): Debit=%.2f, Credit=%.2f\n", 
					account.Code, account.Name, line.DebitAmount, line.CreditAmount)
			}
		}
	}
	
	// 3. Check Revenue and Expense account balances
	fmt.Printf("\n=== REVENUE ACCOUNTS ===\n")
	var revenueAccounts []models.Account
	db.Where("type = ?", "REVENUE").Find(&revenueAccounts)
	totalRevenue := 0.0
	for _, acc := range revenueAccounts {
		if acc.Balance != 0 {
			fmt.Printf("- %s (%s): %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalRevenue += acc.Balance
		}
	}
	fmt.Printf("Total Revenue Balance: %.2f\n", totalRevenue)
	
	fmt.Printf("\n=== EXPENSE ACCOUNTS ===\n")
	var expenseAccounts []models.Account
	db.Where("type = ?", "EXPENSE").Find(&expenseAccounts)
	totalExpense := 0.0
	for _, acc := range expenseAccounts {
		if acc.Balance != 0 {
			fmt.Printf("- %s (%s): %.2f\n", acc.Code, acc.Name, acc.Balance)
			totalExpense += acc.Balance
		}
	}
	fmt.Printf("Total Expense Balance: %.2f\n", totalExpense)
	
	// 4. Check accounting periods
	fmt.Printf("\n=== ACCOUNTING PERIODS ===\n")
	var periods []models.AccountingPeriod
	db.Order("end_date DESC").Limit(3).Find(&periods)
	if len(periods) == 0 {
		fmt.Println("No accounting periods found!")
	} else {
		for i, period := range periods {
			fmt.Printf("\n%d. Period ID: %d\n", i+1, period.ID)
			fmt.Printf("   Start: %s, End: %s\n", 
				period.StartDate.Format("2006-01-02"), 
				period.EndDate.Format("2006-01-02"))
			fmt.Printf("   Is Closed: %v, Is Locked: %v\n", period.IsClosed, period.IsLocked)
			fmt.Printf("   Total Revenue: %.2f\n", period.TotalRevenue)
			fmt.Printf("   Total Expense: %.2f\n", period.TotalExpense)
			fmt.Printf("   Net Income: %.2f\n", period.NetIncome)
			if period.ClosingJournalID != nil {
				fmt.Printf("   Closing Journal ID: %d\n", *period.ClosingJournalID)
			}
		}
	}
	
	fmt.Println("\n=== ANALYSIS ===")
	if totalRevenue == 0 && totalExpense == 0 {
		fmt.Println("✅ Revenue and Expense accounts are zeroed (period has been closed)")
	} else {
		fmt.Println("⚠️  Revenue/Expense accounts still have balances (period NOT closed)")
	}
	
	expectedRetainedEarnings := 11.500000 + 18.500000 // From previous data
	fmt.Printf("\nExpected Retained Earnings: %.2f\n", expectedRetainedEarnings)
	fmt.Printf("Actual Retained Earnings: %.2f\n", retainedEarnings.Balance)
	fmt.Printf("Difference: %.2f\n", retainedEarnings.Balance - expectedRetainedEarnings)
}
