package main

import (
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
)

type CashBankInfo struct {
	ID           uint
	Code         string
	Name         string
	Balance      float64
	AccountID    uint
	AccountCode  string
	AccountName  string
	AccountBalance float64
	Difference   float64
}

type TransactionDetail struct {
	ID              uint
	Date            time.Time
	ReferenceType   string
	ReferenceID     uint
	Amount          float64
	BalanceAfter    float64
	Notes           string
}

type JournalDetail struct {
	ID            uint
	Date          time.Time
	ReferenceType string
	ReferenceID   *uint
	Reference     string
	Description   string
	Status        string
	TotalDebit    float64
	TotalCredit   float64
}

func main() {
	// Load configuration
	_ = config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()

	fmt.Println("🔍 DIAGNOSA MISMATCH KAS/BANK vs COA")
	fmt.Println("=" + string(make([]byte, 60)))

	// Step 1: Get all cash_banks with their linked COA accounts
	var cashBanks []CashBankInfo
	
	query := `
		SELECT 
			cb.id,
			cb.code,
			cb.name,
			cb.balance as balance,
			cb.account_id,
			a.code as account_code,
			a.name as account_name,
			a.balance as account_balance,
			(cb.balance - a.balance) as difference
		FROM cash_banks cb
		LEFT JOIN accounts a ON cb.account_id = a.id
		WHERE cb.deleted_at IS NULL
		ORDER BY cb.id
	`
	
	if err := db.Raw(query).Scan(&cashBanks).Error; err != nil {
		log.Fatalf("❌ Failed to get cash banks: %v", err)
	}

	fmt.Printf("\n📊 SUMMARY KAS/BANK vs COA:\n")
	fmt.Println("-" + string(make([]byte, 100)))
	
	hasMismatch := false
	var mismatchedBanks []CashBankInfo
	
	for _, cb := range cashBanks {
		status := "✅"
		if cb.Difference != 0 {
			status = "❌"
			hasMismatch = true
			mismatchedBanks = append(mismatchedBanks, cb)
		}
		
		fmt.Printf("%s [%s] %s\n", status, cb.Code, cb.Name)
		fmt.Printf("   Kas/Bank Balance: Rp %.2f\n", cb.Balance)
		fmt.Printf("   COA Balance:      Rp %.2f (Account: %s - %s)\n", cb.AccountBalance, cb.AccountCode, cb.AccountName)
		
		if cb.Difference != 0 {
			fmt.Printf("   ⚠️  SELISIH:        Rp %.2f\n", cb.Difference)
		}
		fmt.Println()
	}

	if !hasMismatch {
		fmt.Println("✅ SEMUA KAS/BANK SUDAH SYNC DENGAN COA!")
		return
	}

	// Step 2: Detail analysis for mismatched banks
	fmt.Println("\n🔬 DETAIL ANALYSIS UNTUK ACCOUNT YANG TIDAK SYNC:")
	fmt.Println("=" + string(make([]byte, 60)))

	for _, cb := range mismatchedBanks {
		fmt.Printf("\n\n📋 ANALYZING: [%s] %s\n", cb.Code, cb.Name)
		fmt.Println("-" + string(make([]byte, 80)))

		// Get all transactions for this cash bank
		var transactions []TransactionDetail
		txQuery := `
			SELECT 
				id,
				transaction_date as date,
				reference_type,
				reference_id,
				amount,
				balance_after,
				notes
			FROM cash_bank_transactions
			WHERE cash_bank_id = ?
			  AND deleted_at IS NULL
			ORDER BY transaction_date, id
		`
		
		if err := db.Raw(txQuery, cb.ID).Scan(&transactions).Error; err != nil {
			log.Printf("⚠️ Failed to get transactions for %s: %v\n", cb.Name, err)
			continue
		}

		// Calculate sum of transactions
		var txSum float64
		fmt.Printf("\n💰 CASH BANK TRANSACTIONS (Total: %d):\n", len(transactions))
		for i, tx := range transactions {
			txSum += tx.Amount
			fmt.Printf("  %d. [%s] %s | Ref: %s-%d | Amount: %+.2f | Balance After: %.2f\n",
				i+1, tx.Date.Format("2006-01-02"), tx.Notes, tx.ReferenceType, tx.ReferenceID, 
				tx.Amount, tx.BalanceAfter)
		}
		fmt.Printf("\n  📊 SUM of All Transactions: Rp %.2f\n", txSum)
		fmt.Printf("  📊 Current Cash/Bank Balance: Rp %.2f\n", cb.Balance)
		
		if txSum != cb.Balance {
			fmt.Printf("  ⚠️  MISMATCH: Transaction sum (%.2f) != Balance (%.2f)\n", txSum, cb.Balance)
		}

		// Get journal entries related to this account
		var journals []JournalDetail
		jrnQuery := `
			SELECT 
				je.id,
				je.entry_date as date,
				je.reference_type,
				je.reference_id,
				je.reference,
				je.description,
				je.status,
				je.total_debit,
				je.total_credit
			FROM journal_entries je
			WHERE EXISTS (
				SELECT 1 FROM journal_lines jl
				WHERE jl.journal_entry_id = je.id
				  AND jl.account_id = ?
			)
			ORDER BY je.entry_date, je.id
		`
		
		if err := db.Raw(jrnQuery, cb.AccountID).Scan(&journals).Error; err != nil {
			log.Printf("⚠️ Failed to get journals for account %s: %v\n", cb.AccountCode, err)
			continue
		}

		// Calculate net effect on COA from journals
		fmt.Printf("\n📗 JOURNAL ENTRIES FOR COA ACCOUNT %s (Total: %d):\n", cb.AccountCode, len(journals))
		
		var totalDebit, totalCredit float64
		for i, jrn := range journals {
			// Get journal lines for this specific account
			var lineDebit, lineCredit float64
			db.Raw(`
				SELECT 
					COALESCE(SUM(debit_amount), 0) as debit,
					COALESCE(SUM(credit_amount), 0) as credit
				FROM journal_lines
				WHERE journal_entry_id = ? AND account_id = ?
			`, jrn.ID, cb.AccountID).Row().Scan(&lineDebit, &lineCredit)
			
			totalDebit += lineDebit
			totalCredit += lineCredit
			
			fmt.Printf("  %d. [%s] %s | Ref: %s | D: %.2f | C: %.2f\n",
				i+1, jrn.Date.Format("2006-01-02"), jrn.Description, jrn.Reference,
				lineDebit, lineCredit)
		}
		
		netEffect := totalDebit - totalCredit // For Asset account (debit normal balance)
		fmt.Printf("\n  📊 Total Debit:  Rp %.2f\n", totalDebit)
		fmt.Printf("  📊 Total Credit: Rp %.2f\n", totalCredit)
		fmt.Printf("  📊 Net Effect (Debit - Credit): Rp %.2f\n", netEffect)
		fmt.Printf("  📊 Current COA Balance: Rp %.2f\n", cb.AccountBalance)

		// Compare with cash bank balance
		fmt.Printf("\n🔍 COMPARISON:\n")
		fmt.Printf("  Cash/Bank Balance:     Rp %.2f (from cash_bank_transactions)\n", cb.Balance)
		fmt.Printf("  Journal Net Effect:    Rp %.2f (from journal_lines)\n", netEffect)
		fmt.Printf("  Current COA Balance:   Rp %.2f (from accounts table)\n", cb.AccountBalance)
		
		expectedCOABalance := netEffect
		fmt.Printf("\n  ✅ Expected COA Balance: Rp %.2f\n", expectedCOABalance)
		fmt.Printf("  ❌ Actual COA Balance:   Rp %.2f\n", cb.AccountBalance)
		fmt.Printf("  ⚠️  DISCREPANCY:         Rp %.2f\n", expectedCOABalance - cb.AccountBalance)
	}

	// Step 3: Offer to fix
	fmt.Println("\n\n🔧 PERBAIKAN:")
	fmt.Println("=" + string(make([]byte, 60)))
	
	fmt.Println("\n❓ Apakah Anda ingin memperbaiki COA balance agar sesuai dengan Cash/Bank balance?")
	fmt.Println("   Option 1: Sync COA balance = Cash/Bank balance (Simple fix)")
	fmt.Println("   Option 2: Sync COA balance = Journal Net Effect (Accounting-correct fix)")
	fmt.Println("\nPilihan yang direkomendasikan: Option 1 (karena Cash/Bank balance sudah benar)")
	
	fmt.Print("\nMasukkan pilihan (1/2) atau 'n' untuk skip: ")
	var choice string
	fmt.Scanln(&choice)

	if choice != "1" && choice != "2" {
		fmt.Println("❌ Perbaikan dibatalkan")
		return
	}

	// Apply fix
	fmt.Println("\n⚙️ Menerapkan perbaikan...")
	
	tx := db.Begin()
	
	for _, cb := range mismatchedBanks {
		var newBalance float64
		
		if choice == "1" {
			// Option 1: Use cash_bank balance as source of truth
			newBalance = cb.Balance
			fmt.Printf("  [%s] %s: Setting COA balance to %.2f (from Cash/Bank)\n", cb.Code, cb.Name, newBalance)
		} else {
			// Option 2: Calculate from journal entries
			var totalDebit, totalCredit float64
			tx.Raw(`
				SELECT 
					COALESCE(SUM(jl.debit_amount), 0) as total_debit,
					COALESCE(SUM(jl.credit_amount), 0) as total_credit
				FROM journal_lines jl
				JOIN journal_entries je ON jl.journal_entry_id = je.id
				WHERE jl.account_id = ?
				  AND je.status = 'POSTED'
			`, cb.AccountID).Row().Scan(&totalDebit, &totalCredit)
			
			newBalance = totalDebit - totalCredit
			fmt.Printf("  [%s] %s: Setting COA balance to %.2f (from Journals)\n", cb.Code, cb.Name, newBalance)
		}

		// Update COA balance
		if err := tx.Exec(`
			UPDATE accounts 
			SET balance = ?, updated_at = CURRENT_TIMESTAMP 
			WHERE id = ?
		`, newBalance, cb.AccountID).Error; err != nil {
			tx.Rollback()
			log.Fatalf("❌ Failed to update account %s: %v", cb.AccountCode, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Fatalf("❌ Failed to commit changes: %v", err)
	}

	fmt.Println("\n✅ PERBAIKAN BERHASIL!")
	
	// Verify fix
	fmt.Println("\n🔍 VERIFIKASI SETELAH PERBAIKAN:")
	fmt.Println("-" + string(make([]byte, 80)))
	
	var verifyBanks []CashBankInfo
	if err := db.Raw(query).Scan(&verifyBanks).Error; err != nil {
		log.Fatalf("❌ Failed to verify: %v", err)
	}

	allFixed := true
	for _, cb := range verifyBanks {
		status := "✅"
		if cb.Difference != 0 {
			status = "❌"
			allFixed = false
		}
		
		fmt.Printf("%s [%s] %s: Kas/Bank %.2f | COA %.2f | Diff: %.2f\n",
			status, cb.Code, cb.Name, cb.Balance, cb.AccountBalance, cb.Difference)
	}

	if allFixed {
		fmt.Println("\n🎉 SEMUA ACCOUNT SUDAH SYNC!")
	} else {
		fmt.Println("\n⚠️ Masih ada account yang belum sync. Mungkin perlu investigasi lebih lanjut.")
	}
}
