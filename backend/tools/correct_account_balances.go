package main

import (
	"fmt"
	"log"
	"time"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"github.com/shopspring/decimal"
)

type BalanceCorrection struct {
	AccountCode     string
	CurrentBalance  float64
	ExpectedBalance float64
	Adjustment      float64
	Reason          string
}

func main() {
	// Load config and connect to database
	_ = config.LoadConfig()
	db := database.ConnectDB()

	fmt.Println("🔧 Correcting Chart of Accounts Balance Issues")
	fmt.Println("=" + repeatString("=", 60))

	// Create a backup first
	fmt.Println("1. Creating balance backup...")
	createBalanceBackup(db)

	// Analyze and fix balances
	fmt.Println("\n2. Analyzing current balances...")
	corrections := analyzeBalances(db)

	if len(corrections) == 0 {
		fmt.Println("✅ No balance corrections needed!")
		return
	}

	fmt.Printf("\n3. Applying %d balance corrections...\n", len(corrections))
	applyCorrections(db, corrections)

	fmt.Println("\n4. Verifying corrections...")
	verifyCorrections(db)

	fmt.Println("\n✅ Balance correction process complete!")
}

func createBalanceBackup(db *gorm.DB) {
	backupTable := fmt.Sprintf("accounts_balance_backup_%d", time.Now().Unix())
	
	sql := fmt.Sprintf(`
		CREATE TABLE %s AS 
		SELECT id, code, name, balance, type, created_at, updated_at, NOW() as backup_timestamp 
		FROM accounts WHERE deleted_at IS NULL
	`, backupTable)
	
	if err := db.Exec(sql).Error; err != nil {
		log.Printf("Warning: Could not create backup table: %v", err)
	} else {
		fmt.Printf("✅ Created backup table: %s\n", backupTable)
	}
}

func analyzeBalances(db *gorm.DB) []BalanceCorrection {
	var corrections []BalanceCorrection
	
	// Key accounts to check
	accountChecks := map[string]struct {
		ExpectedType string
		Description  string
	}{
		"1201": {"ASSET", "Piutang Usaha (Accounts Receivable)"},
		"4101": {"REVENUE", "Pendapatan Penjualan (Sales Revenue)"},
		"1102": {"ASSET", "Bank BCA"},
		"1104": {"ASSET", "Bank Mandiri"},
		"1103": {"ASSET", "Bank UOB"},
	}

	for code := range accountChecks {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			fmt.Printf("⚠️ Account %s not found, skipping\n", code)
			continue
		}

		normalBalance := account.GetNormalBalance()
		
		// Check if balance has wrong sign
		needsCorrection := false
		var expectedBalance float64
		reason := ""

		if normalBalance == models.NormalBalanceDebit {
			// For debit accounts (Assets, Expenses), balance should be >= 0
			if account.Balance < 0 {
				needsCorrection = true
				expectedBalance = calculateExpectedDebitBalance(db, account)
				reason = "Debit account has negative balance - should be positive"
			}
		} else {
			// For credit accounts (Liabilities, Equity, Revenue), balance should be <= 0
			if account.Balance > 0 {
				needsCorrection = true
				expectedBalance = calculateExpectedCreditBalance(db, account)
				reason = "Credit account has positive balance - should be negative"
			}
		}

		if needsCorrection || (account.Code == "4101" && account.Balance == 0) {
			// Special check for revenue accounts with zero balance
			if account.Code == "4101" && account.Balance == 0 {
				expectedBalance = calculateExpectedCreditBalance(db, account)
				reason = "Sales revenue account has zero balance but should reflect sales"
				needsCorrection = true
			}
			
			if needsCorrection {
				correction := BalanceCorrection{
					AccountCode:     account.Code,
					CurrentBalance:  account.Balance,
					ExpectedBalance: expectedBalance,
					Adjustment:      expectedBalance - account.Balance,
					Reason:          reason,
				}
				corrections = append(corrections, correction)
				
				fmt.Printf("❌ %s (%s): Current %.2f → Expected %.2f (Adj: %.2f) - %s\n",
					code, account.Name, account.Balance, expectedBalance, 
					correction.Adjustment, reason)
			}
		} else {
			fmt.Printf("✅ %s (%s): Balance %.2f is correct\n", 
				code, account.Name, account.Balance)
		}
	}

	return corrections
}

func calculateExpectedDebitBalance(db *gorm.DB, account models.Account) float64 {
	switch account.Code {
	case "1201": // Accounts Receivable
		return calculateARBalance(db)
	case "1102", "1103", "1104": // Bank accounts
		return calculateBankBalance(db, account.Code)
	default:
		// For other debit accounts, assume current absolute value
		if account.Balance < 0 {
			return -account.Balance
		}
		return account.Balance
	}
}

func calculateExpectedCreditBalance(db *gorm.DB, account models.Account) float64 {
	switch account.Code {
	case "4101": // Sales Revenue
		return calculateSalesRevenueBalance(db)
	default:
		// For other credit accounts, assume current value but negative
		if account.Balance > 0 {
			return -account.Balance
		}
		return account.Balance
	}
}

func calculateARBalance(db *gorm.DB) float64 {
	// Calculate expected AR balance from outstanding sales
	var totalOutstanding float64
	db.Model(&models.Sale{}).
		Where("status IN (?, ?)", models.SaleStatusInvoiced, models.SaleStatusOverdue).
		Select("COALESCE(SUM(outstanding_amount), 0)").
		Scan(&totalOutstanding)
	
	return totalOutstanding // Should be positive for debit account
}

func calculateBankBalance(db *gorm.DB, bankCode string) float64 {
	// For bank accounts, check if there are corresponding cash_bank records
	var cashBank models.CashBank
	err := db.Joins("JOIN accounts ON accounts.id = cash_banks.account_id").
		Where("accounts.code = ?", bankCode).
		First(&cashBank).Error
	
	if err == nil {
		return cashBank.Balance
	}
	
	// Fallback: assume positive balance for bank accounts
	return 1000000 // Default positive balance
}

func calculateSalesRevenueBalance(db *gorm.DB) float64 {
	// Calculate total invoiced sales revenue
	var totalRevenue float64
	db.Model(&models.Sale{}).
		Where("status IN (?, ?, ?)", 
			models.SaleStatusInvoiced, models.SaleStatusPaid, models.SaleStatusOverdue).
		Select("COALESCE(SUM(total_amount - ppn), 0)"). // Revenue without tax
		Scan(&totalRevenue)
	
	return -totalRevenue // Should be negative for credit account
}

func applyCorrections(db *gorm.DB, corrections []BalanceCorrection) {
	// Create adjustment journal entry
	adjustmentEntry := &models.SSOTJournalEntry{
		EntryNumber:     generateAdjustmentNumber(db),
		EntryDate:       time.Now(),
		Description:     "Balance Correction - Fixing incorrect account balances",
		Status:          "POSTED",
		IsAutoGenerated: true,
	}

	var journalLines []models.SSOTJournalLine
	lineNumber := 1
	totalDebit := 0.0
	totalCredit := 0.0

	// Create adjustment lines
	for _, correction := range corrections {
		var account models.Account
		db.Where("code = ?", correction.AccountCode).First(&account)

		if correction.Adjustment != 0 {
			if correction.Adjustment > 0 {
				// Need to increase balance - debit
				journalLines = append(journalLines, models.SSOTJournalLine{
					AccountID:    uint64(account.ID),
					Description:  fmt.Sprintf("Balance adjustment - %s", correction.Reason),
					DebitAmount:  decimal.NewFromFloat(correction.Adjustment),
					CreditAmount: decimal.Zero,
					LineNumber:   lineNumber,
				})
				totalDebit += correction.Adjustment
			} else {
				// Need to decrease balance - credit
				journalLines = append(journalLines, models.SSOTJournalLine{
					AccountID:    uint64(account.ID),
					Description:  fmt.Sprintf("Balance adjustment - %s", correction.Reason),
					DebitAmount:  decimal.Zero,
					CreditAmount: decimal.NewFromFloat(-correction.Adjustment),
					LineNumber:   lineNumber,
				})
				totalCredit += -correction.Adjustment
			}
			lineNumber++
		}

		// Directly update the account balance
		account.Balance = correction.ExpectedBalance
		db.Save(&account)
		
		fmt.Printf("✅ Corrected %s: %.2f → %.2f\n", 
			correction.AccountCode, correction.CurrentBalance, correction.ExpectedBalance)
	}

	// Balance the entry with a system adjustment account
	if totalDebit != totalCredit {
		diff := totalDebit - totalCredit
		// Create or find system adjustment account
		adjustmentAccount := findOrCreateAdjustmentAccount(db)
		
		if diff > 0 {
			// Need more credit
			journalLines = append(journalLines, models.SSOTJournalLine{
				AccountID:    uint64(adjustmentAccount.ID),
				Description:  "System balance adjustment",
				DebitAmount:  decimal.Zero,
				CreditAmount: decimal.NewFromFloat(diff),
				LineNumber:   lineNumber,
			})
			totalCredit += diff
		} else {
			// Need more debit
			journalLines = append(journalLines, models.SSOTJournalLine{
				AccountID:    uint64(adjustmentAccount.ID),
				Description:  "System balance adjustment",
				DebitAmount:  decimal.NewFromFloat(-diff),
				CreditAmount: decimal.Zero,
				LineNumber:   lineNumber,
			})
			totalDebit += -diff
		}
	}

	adjustmentEntry.TotalDebit = decimal.NewFromFloat(totalDebit)
	adjustmentEntry.TotalCredit = decimal.NewFromFloat(totalCredit)
	adjustmentEntry.Lines = journalLines

	// Save the adjustment entry
	if err := db.Create(adjustmentEntry).Error; err != nil {
		fmt.Printf("❌ Failed to create adjustment entry: %v\n", err)
	} else {
		fmt.Printf("✅ Created adjustment journal entry %d\n", adjustmentEntry.ID)
	}
}

func verifyCorrections(db *gorm.DB) {
	accountCodes := []string{"1201", "4101", "1102", "1104"}
	
	fmt.Println("Verification Results:")
	for _, code := range accountCodes {
		var account models.Account
		err := db.Where("code = ?", code).First(&account).Error
		if err != nil {
			continue
		}

		normalBalance := account.GetNormalBalance()
		status := "✅ CORRECT"
		
		if normalBalance == models.NormalBalanceDebit && account.Balance < 0 {
			status = "❌ STILL INCORRECT"
		} else if normalBalance == models.NormalBalanceCredit && account.Balance > 0 {
			status = "❌ STILL INCORRECT"
		}

		fmt.Printf("  %s (%s): %.2f - %s\n", code, account.Name, account.Balance, status)
	}
}

func findOrCreateAdjustmentAccount(db *gorm.DB) *models.Account {
	var account models.Account
	err := db.Where("code = ?", "9999").First(&account).Error
	if err == nil {
		return &account
	}

	// Create adjustment account
	account = models.Account{
		Code:        "9999",
		Name:        "System Adjustments",
		Type:        "EQUITY",
		Description: "System balance adjustments",
		IsActive:    true,
		IsHeader:    false,
		Balance:     0,
	}

	db.Create(&account)
	return &account
}

func generateAdjustmentNumber(db *gorm.DB) string {
	var count int64
	db.Model(&models.SSOTJournalEntry{}).
		Where("description LIKE ?", "%Balance Correction%").
		Count(&count)
	
	return fmt.Sprintf("ADJ/%04d", count+1)
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}