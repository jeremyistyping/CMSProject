package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	log.Printf("🔍 Debugging SSOT Balance Calculation Query")

	db := database.ConnectDB()

	targetAccountID := uint64(4)
	
	log.Printf("🎯 Testing different query approaches for Account ID: %d", targetAccountID)

	// 1. Test direct query to unified_journal_lines
	var totalDebits1, totalCredits1 float64
	var lineCount1 int64
	
	err1 := db.Table("unified_journal_lines").
		Select("SUM(debit_amount) as total_debits, SUM(credit_amount) as total_credits, COUNT(*) as line_count").
		Where("account_id = ?", targetAccountID).
		Scan(&map[string]interface{}{
			"total_debits": &totalDebits1,
			"total_credits": &totalCredits1, 
			"line_count": &lineCount1,
		}).Error
	
	if err1 != nil {
		log.Printf("❌ Direct query failed: %v", err1)
	} else {
		log.Printf("📊 Direct Query Results:")
		log.Printf("   Total Debits: %.2f", totalDebits1)
		log.Printf("   Total Credits: %.2f", totalCredits1)
		log.Printf("   Line Count: %d", lineCount1)
	}

	// 2. Test query with JOIN to ledger (original problematic query)
	var totalDebits2, totalCredits2 float64
	var lineCount2 int64
	
	err2 := db.Table("unified_journal_lines").
		Select("SUM(debit_amount) as total_debits, SUM(credit_amount) as total_credits, COUNT(*) as line_count").
		Where("account_id = ?", targetAccountID).
		Joins("JOIN unified_journal_ledger ON unified_journal_lines.journal_id = unified_journal_ledger.id").
		Where("unified_journal_ledger.status = ?", "POSTED").
		Scan(&map[string]interface{}{
			"total_debits": &totalDebits2,
			"total_credits": &totalCredits2, 
			"line_count": &lineCount2,
		}).Error
	
	if err2 != nil {
		log.Printf("❌ JOIN query failed: %v", err2)
	} else {
		log.Printf("📊 JOIN Query Results (with POSTED filter):")
		log.Printf("   Total Debits: %.2f", totalDebits2)
		log.Printf("   Total Credits: %.2f", totalCredits2)
		log.Printf("   Line Count: %d", lineCount2)
	}

	// 3. Test manual calculation using GORM models
	var journalLines []models.SSOTJournalLine
	if err := db.Where("account_id = ?", targetAccountID).Find(&journalLines).Error; err == nil {
		var manualDebits, manualCredits float64
		postedLines := 0
		
		for _, line := range journalLines {
			// Check if the parent journal is POSTED
			var parentJournal models.SSOTJournalEntry
			if err := db.First(&parentJournal, line.JournalID).Error; err == nil {
				if parentJournal.Status == "POSTED" {
					manualDebits += line.DebitAmount.InexactFloat64()
					manualCredits += line.CreditAmount.InexactFloat64()
					postedLines++
				}
			}
		}
		
		log.Printf("📊 Manual Calculation Results:")
		log.Printf("   Total Lines Found: %d", len(journalLines))
		log.Printf("   Posted Lines: %d", postedLines)
		log.Printf("   Manual Total Debits: %.2f", manualDebits)
		log.Printf("   Manual Total Credits: %.2f", manualCredits)
		log.Printf("   Expected Balance: %.2f", manualDebits - manualCredits)
	}

	// 4. Check table structure
	log.Printf("\n🔧 Checking table structure:")
	
	// Check if unified_journal_ledger table exists and has correct structure
	var tableExists bool
	db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_ledger')").Scan(&tableExists)
	log.Printf("   unified_journal_ledger table exists: %v", tableExists)
	
	// Check if unified_journal_lines table exists and has correct structure  
	var linesTableExists bool
	db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_lines')").Scan(&linesTableExists)
	log.Printf("   unified_journal_lines table exists: %v", linesTableExists)

	// 5. Raw SQL debug
	log.Printf("\n🔍 Raw SQL Investigation:")
	
	// Check actual data with raw SQL
	var rawResult struct {
		TotalDebits  float64 `json:"total_debits"`
		TotalCredits float64 `json:"total_credits"`
		LineCount    int64   `json:"line_count"`
	}
	
	err3 := db.Raw(`
		SELECT 
			SUM(ujl.debit_amount) as total_debits,
			SUM(ujl.credit_amount) as total_credits,
			COUNT(*) as line_count
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uj ON ujl.journal_id = uj.id
		WHERE ujl.account_id = ? AND uj.status = 'POSTED'
	`, targetAccountID).Scan(&rawResult).Error
	
	if err3 != nil {
		log.Printf("❌ Raw SQL query failed: %v", err3)
	} else {
		log.Printf("📊 Raw SQL Results:")
		log.Printf("   Total Debits: %.2f", rawResult.TotalDebits)
		log.Printf("   Total Credits: %.2f", rawResult.TotalCredits)
		log.Printf("   Line Count: %d", rawResult.LineCount)
	}

	log.Printf("\n✅ SSOT query debugging completed")
}