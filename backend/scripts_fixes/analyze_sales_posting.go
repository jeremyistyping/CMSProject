package main

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔍 Analyzing Sales Posting, Journal Entries, and COA Balance Updates")
	log.Printf("==================================================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Get the latest sales transaction
	var latestSale models.Sale
	err = db.Preload("SaleItems.RevenueAccount").
		Preload("Customer").
		Order("created_at DESC").
		First(&latestSale).Error
	
	if err != nil {
		log.Printf("❌ No sales found: %v", err)
		return
	}

	log.Printf("📄 Analyzing Latest Sale:")
	log.Printf("   Sale ID: %d", latestSale.ID)
	log.Printf("   Code: %s", latestSale.Code)
	log.Printf("   Status: %s", latestSale.Status)
	log.Printf("   Customer: %s", latestSale.Customer.Name)
	log.Printf("   Total Amount: %.2f", latestSale.TotalAmount)
	log.Printf("   PPN Amount: %.2f", latestSale.PPNAmount)
	log.Printf("   Created: %v", latestSale.CreatedAt.Format("2006-01-02 15:04:05"))

	// Analyze sale items and their revenue accounts
	log.Printf("\n📋 Sale Items Analysis:")
	for i, item := range latestSale.SaleItems {
		log.Printf("   Item %d:", i+1)
		log.Printf("     Product: %s", item.Description)
		log.Printf("     Quantity: %.2f", item.Quantity)
		log.Printf("     Unit Price: %.2f", item.UnitPrice)
		log.Printf("     Amount: %.2f", item.LineTotal)
		log.Printf("     PPN Amount: %.2f", item.PPNAmount)
		
		if item.RevenueAccountID > 0 {
			log.Printf("     Revenue Account: %s (%s)", item.RevenueAccount.Name, item.RevenueAccount.Code)
			log.Printf("     Revenue Account Type: %s", item.RevenueAccount.Type)
		} else {
			log.Printf("     ⚠️ No Revenue Account assigned!")
		}
	}

	// Get journal entries for this sale
	log.Printf("\n📖 Journal Entries for this Sale:")
	var journalEntries []models.JournalEntry
	err = db.Preload("JournalLines.Account").
		Where("reference_type = ? AND reference_id = ?", "SALE", latestSale.ID).
		Order("id ASC").
		Find(&journalEntries).Error
	
	if err != nil {
		log.Printf("❌ Failed to get journal entries: %v", err)
	} else if len(journalEntries) == 0 {
		log.Printf("⚠️ No journal entries found for this sale!")
	} else {
		totalDebit := 0.0
		totalCredit := 0.0
		
		for i, entry := range journalEntries {
			log.Printf("   Entry %d:", i+1)
			log.Printf("     Code: %s", entry.Code)
			log.Printf("     Description: %s", entry.Description)
			log.Printf("     Status: %s", entry.Status)
			log.Printf("     Total Debit: %.2f", entry.TotalDebit)
			log.Printf("     Total Credit: %.2f", entry.TotalCredit)
			log.Printf("     Entry Date: %v", entry.EntryDate.Format("2006-01-02"))
			
			totalDebit += entry.TotalDebit
			totalCredit += entry.TotalCredit
			
			// Show journal lines
			for j, line := range entry.JournalLines {
				log.Printf("     Line %d: %s (%s) - Debit: %.2f, Credit: %.2f", 
					j+1, line.Account.Name, line.Account.Code, 
					line.DebitAmount, line.CreditAmount)
			}
		}
		
		log.Printf("   💰 Journal Summary:")
		log.Printf("     Total Debit: %.2f", totalDebit)
		log.Printf("     Total Credit: %.2f", totalCredit)
		log.Printf("     Balance: %.2f", totalDebit-totalCredit)
		
		if totalDebit != totalCredit {
			log.Printf("     ⚠️ JOURNAL NOT BALANCED!")
		} else {
			log.Printf("     ✅ Journal is balanced")
		}
	}

	// Check specific account balances mentioned in the issue
	log.Printf("\n💳 Account Balance Analysis:")
	
	// 1. Cash/Bank Account Balance
	if latestSale.CashBankID != nil {
		var cashBank models.CashBank
		db.Preload("Account").Where("id = ?", *latestSale.CashBankID).First(&cashBank)
		log.Printf("   💰 Cash/Bank Account: %s (%s)", 
			cashBank.Account.Name, cashBank.Account.Code)
		log.Printf("     Current Balance: %.2f", cashBank.Account.Balance)
	}

	// 2. Revenue Accounts Balance
	log.Printf("   📈 Revenue Accounts:")
	for i, item := range latestSale.SaleItems {
		if item.RevenueAccountID > 0 {
			// Get fresh account data
			var revenueAccount models.Account
			db.Where("id = ?", item.RevenueAccountID).First(&revenueAccount)
			
			log.Printf("     Item %d Revenue (%s): %.2f", i+1, 
				revenueAccount.Code, revenueAccount.Balance)
		}
	}

	// 3. Tax/PPN Account Balance
	log.Printf("   🧾 Tax/PPN Accounts:")
	var ppnAccounts []models.Account
	db.Where("name ILIKE ? OR code ILIKE ?", "%PPN%", "%PPN%").Find(&ppnAccounts)
	
	if len(ppnAccounts) == 0 {
		log.Printf("     ⚠️ No PPN accounts found in COA")
	} else {
		for _, ppnAcc := range ppnAccounts {
			log.Printf("     %s (%s): %.2f", ppnAcc.Name, ppnAcc.Code, ppnAcc.Balance)
		}
	}

	// Check for any transactions on these accounts today
	log.Printf("\n📊 Recent Transactions (Today):")
	today := time.Now().Truncate(24 * time.Hour)
	var recentTransactions []models.Transaction
	
	err = db.Preload("Account").
		Where("created_at >= ?", today).
		Order("created_at DESC").
		Limit(20).
		Find(&recentTransactions).Error
	
	if err != nil {
		log.Printf("❌ Failed to get recent transactions: %v", err)
	} else if len(recentTransactions) == 0 {
		log.Printf("   No transactions found for today")
	} else {
		for _, txn := range recentTransactions {
			log.Printf("   %s: %s (%s) - %s %.2f", 
				txn.CreatedAt.Format("15:04:05"),
				txn.Account.Name,
				txn.Account.Code,
				txn.Type,
				txn.Amount)
		}
	}

	// Check for posting conflicts or issues
	log.Printf("\n🔍 Checking for Posting Issues:")
	
	// Check if there are pending journal entries
	var pendingEntries []models.JournalEntry
	db.Where("is_posted = ?", false).Find(&pendingEntries)
	log.Printf("   Pending journal entries: %d", len(pendingEntries))
	
	// Check for duplicate entries
	var duplicateEntries []models.JournalEntry
	db.Raw(`
		SELECT * FROM journal_entries 
		WHERE reference_type = ? AND reference_id = ?
		AND id NOT IN (
			SELECT MIN(id) FROM journal_entries 
			WHERE reference_type = ? AND reference_id = ?
			GROUP BY account_id, type, amount
		)`, "SALE", latestSale.ID, "SALE", latestSale.ID).Find(&duplicateEntries)
	
	if len(duplicateEntries) > 0 {
		log.Printf("   ⚠️ Found %d potential duplicate journal entries", len(duplicateEntries))
		for _, dup := range duplicateEntries {
			log.Printf("     Duplicate: Code %s, Debit %.2f, Credit %.2f", 
				dup.Code, dup.TotalDebit, dup.TotalCredit)
		}
	} else {
		log.Printf("   ✅ No duplicate journal entries found")
	}

	// Summary and recommendations
	log.Printf("\n📋 SUMMARY & RECOMMENDATIONS:")
	
	if len(journalEntries) == 0 {
		log.Printf("❌ CRITICAL: No journal entries found for the sale")
		log.Printf("   → Journal entries may not be created during sale confirmation")
		log.Printf("   → Check sales service posting logic")
	}
	
	if len(ppnAccounts) == 0 {
		log.Printf("⚠️ WARNING: No PPN accounts found")
		log.Printf("   → PPN account may not exist in COA")
		log.Printf("   → Check if tax account is properly configured")
	}
	
	expectedRevenue := latestSale.TotalAmount - latestSale.PPNAmount
	log.Printf("💡 Expected Revenue Amount: %.2f", expectedRevenue)
	log.Printf("💡 Expected PPN Amount: %.2f", latestSale.PPNAmount)
	
	log.Printf("\n🎯 Next Steps:")
	log.Printf("1. Verify journal entry creation in sales service")
	log.Printf("2. Check if accounts are being updated correctly")
	log.Printf("3. Ensure PPN account exists and is configured")
	log.Printf("4. Test revenue account balance calculation")
}