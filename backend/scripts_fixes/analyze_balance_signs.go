package main

import (
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

func main() {
	log.Printf("🔍 Analyzing why Revenue and PPN balances show minus instead of plus...")

	// Initialize database connection
	db := database.ConnectDB()

	log.Printf("\n📊 Current Account Balances (Raw):")
	checkRawBalances(db)

	log.Printf("\n🧮 Accounting Balance Logic:")
	explainAccountingRules()

	log.Printf("\n📋 Journal Lines Analysis:")
	analyzeJournalLines(db)

	log.Printf("\n🔧 How to Fix Display:")
	explainFixForDisplay()
}

func checkRawBalances(db *gorm.DB) {
	targetCodes := []string{"1101", "1201", "4101", "2103"}
	
	var accounts []models.Account
	db.Where("code IN ?", targetCodes).Order("code").Find(&accounts)
	
	for _, acc := range accounts {
		log.Printf("   Account %s (%s): Raw Balance %.2f [Type: %s]", 
			acc.Code, acc.Name, acc.Balance, acc.Type)
	}
}

func explainAccountingRules() {
	log.Printf("📚 Double-Entry Accounting Rules:")
	log.Printf("   ASSET accounts (1xxx):")
	log.Printf("     - DEBIT increases balance (positive)")
	log.Printf("     - CREDIT decreases balance (negative)")
	log.Printf("     - Normal balance: POSITIVE")
	log.Printf("")
	log.Printf("   REVENUE accounts (4xxx):")
	log.Printf("     - DEBIT decreases balance (negative)")
	log.Printf("     - CREDIT increases balance (positive)")
	log.Printf("     - Normal balance: POSITIVE (but stored as NEGATIVE)")
	log.Printf("")
	log.Printf("   LIABILITY accounts (2xxx):")
	log.Printf("     - DEBIT decreases balance (negative)")
	log.Printf("     - CREDIT increases balance (positive)")
	log.Printf("     - Normal balance: POSITIVE (but stored as NEGATIVE)")
	log.Printf("")
	log.Printf("💡 INSIGHT: Revenue dan PPN Keluaran di-CREDIT, makanya balance jadi MINUS")
	log.Printf("   Tapi untuk DISPLAY di laporan, harus dibalik jadi PLUS!")
}

func analyzeJournalLines(db *gorm.DB) {
	// Check recent journal lines for revenue and PPN accounts
	query := `
		SELECT 
			a.code,
			a.name,
			a.type,
			jl.debit_amount::float as debit_amount,
			jl.credit_amount::float as credit_amount,
			je.description
		FROM accounts a
		JOIN ssot_journal_lines jl ON a.id = jl.account_id
		JOIN ssot_journal_entries je ON jl.journal_entry_id = je.id
		WHERE a.code IN ('4101', '2103')
		ORDER BY je.created_at DESC
		LIMIT 10
	`
	
	var lines []struct {
		Code         string
		Name         string 
		Type         string
		DebitAmount  float64
		CreditAmount float64
		Description  string
	}
	
	err := db.Raw(query).Scan(&lines).Error
	if err != nil {
		log.Printf("⚠️  Could not analyze journal lines (table might not exist yet): %v", err)
		return
	}
	
	if len(lines) == 0 {
		log.Printf("ℹ️  No journal lines found for Revenue/PPN accounts")
		return
	}
	
	log.Printf("📊 Recent journal activity:")
	for _, line := range lines {
		log.Printf("   %s (%s) - %s:", line.Code, line.Name, line.Description)
		log.Printf("     Debit: %.2f, Credit: %.2f", line.DebitAmount, line.CreditAmount)
		
		if line.CreditAmount > 0 {
			log.Printf("     → CREDITED (increases %s balance, but makes it more NEGATIVE)", line.Type)
		}
		if line.DebitAmount > 0 {
			log.Printf("     → DEBITED (decreases %s balance, makes it more POSITIVE)", line.Type)
		}
	}
}

func explainFixForDisplay() {
	log.Printf("🔧 SOLUTION: Frontend harus flip sign untuk display!")
	log.Printf("")
	log.Printf("📱 Frontend Logic yang Benar:")
	log.Printf("   ASSET accounts (1xxx): Display as-is")
	log.Printf("     Example: Balance = 1000000 → Display: 1,000,000")
	log.Printf("")
	log.Printf("   REVENUE accounts (4xxx): Flip sign (multiply by -1)")
	log.Printf("     Example: Balance = -1000000 → Display: 1,000,000")
	log.Printf("     Reason: Revenue SHOULD show as positive income")
	log.Printf("")
	log.Printf("   LIABILITY accounts (2xxx): Flip sign (multiply by -1)")
	log.Printf("     Example: Balance = -110000 → Display: 110,000")  
	log.Printf("     Reason: PPN owed SHOULD show as positive amount owed")
	log.Printf("")
	log.Printf("💻 Code Example:")
	log.Printf("   function displayBalance(account) {")
	log.Printf("     if (account.type === 'REVENUE' || account.type === 'LIABILITY') {")
	log.Printf("       return account.balance * -1; // Flip sign")
	log.Printf("     }")
	log.Printf("     return account.balance; // Keep as-is for ASSET/EXPENSE")
	log.Printf("   }")
	log.Printf("")
	log.Printf("✅ HASIL: Revenue dan PPN akan tampil POSITIVE di COA seperti yang diharapkan!")
}