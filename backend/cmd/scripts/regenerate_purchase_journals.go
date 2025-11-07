package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("REGENERATE PURCHASE JOURNALS")
	fmt.Println("For approved purchases that don't have journal entries")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// 1. Find all approved purchases
	var purchases []models.Purchase
	err = db.Preload("Vendor").
		Preload("PurchaseItems.Product").
		Preload("PurchaseItems.ExpenseAccount").
		Where("status = ? OR status = ?", "APPROVED", "COMPLETED").
		Where("approval_status = ?", "APPROVED").
		Find(&purchases).Error
	
	if err != nil {
		log.Fatalf("Failed to fetch approved purchases: %v", err)
	}

	fmt.Printf("Found %d approved purchases\n\n", len(purchases))

	if len(purchases) == 0 {
		fmt.Println("No approved purchases found. Exiting.")
		return
	}

	// 2. Check which purchases don't have journal entries
	needsJournal := []models.Purchase{}
	for _, p := range purchases {
		var count int64
		
		// Check in simple_ssot_journals table
		err := db.Model(&models.SimpleSSOTJournal{}).
			Where("transaction_type = ? AND transaction_id = ?", "PURCHASE", p.ID).
			Count(&count).Error
		
		if err != nil {
			fmt.Printf("⚠️  Error checking journal for Purchase %s: %v\n", p.Code, err)
			continue
		}
		
		if count == 0 {
			needsJournal = append(needsJournal, p)
			fmt.Printf("❌ Purchase %s (%s) - NO JOURNAL ENTRY\n", p.Code, p.Vendor.Name)
		} else {
			fmt.Printf("✅ Purchase %s (%s) - Already has journal\n", p.Code, p.Vendor.Name)
		}
	}

	fmt.Printf("\n%d purchases need journal entries\n\n", len(needsJournal))

	if len(needsJournal) == 0 {
		fmt.Println("All approved purchases already have journal entries. Exiting.")
		return
	}

	// 3. Create journal entries for purchases without them
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("CREATING JOURNAL ENTRIES")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println()

	successCount := 0
	errorCount := 0

	for _, purchase := range needsJournal {
		fmt.Printf("\n📝 Processing Purchase %s...\n", purchase.Code)
		
		err := createPurchaseJournal(db, &purchase)
		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			errorCount++
		} else {
			fmt.Printf("✅ Successfully created journal for %s\n", purchase.Code)
			successCount++
		}
	}

	// 4. Summary
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("✅ Success: %d\n", successCount)
	fmt.Printf("❌ Errors:  %d\n", errorCount)
	fmt.Printf("📊 Total:   %d\n", len(needsJournal))
	fmt.Println(strings.Repeat("=", 80))
}

func createPurchaseJournal(db *gorm.DB, purchase *models.Purchase) error {
	// Create Simple SSOT Journal Entry
	ssotEntry := &models.SimpleSSOTJournal{
		EntryNumber:       fmt.Sprintf("PURCHASE-%d", purchase.ID),
		TransactionType:   "PURCHASE",
		TransactionID:     purchase.ID,
		TransactionNumber: purchase.Code,
		Date:              purchase.Date,
		Description:       fmt.Sprintf("Purchase Order #%s - %s", purchase.Code, purchase.Vendor.Name),
		TotalAmount:       purchase.TotalAmount,
		Status:            "POSTED",
		CreatedBy:         purchase.UserID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Create SSOT Journal Entry
	if err := db.Create(ssotEntry).Error; err != nil {
		return fmt.Errorf("failed to create SSOT journal: %v", err)
	}

	fmt.Printf("   Created journal entry ID: %d\n", ssotEntry.ID)

	// Helper to resolve account by code
	resolveByCode := func(code string) (*models.Account, error) {
		var acc models.Account
		if err := db.Where("code = ?", code).First(&acc).Error; err != nil {
			return nil, fmt.Errorf("account code %s not found: %v", code, err)
		}
		return &acc, nil
	}

	// Get default expense account (6001 or from settings)
	defaultExpenseAccount, err := resolveByCode("6001")
	if err != nil {
		// Try 5001 (COGS) as fallback
		defaultExpenseAccount, err = resolveByCode("5001")
		if err != nil {
			return fmt.Errorf("failed to find default expense account (6001 or 5001): %v", err)
		}
	}

	// Prepare journal items
	var journalItems []models.SimpleSSOTJournalItem

	// DEBIT SIDE - Expense/Inventory Accounts
	for _, item := range purchase.PurchaseItems {
		var account models.Account
		
		// Use expense account from item if available, otherwise use default
		if item.ExpenseAccountID != 0 {
			if err := db.First(&account, item.ExpenseAccountID).Error; err != nil {
				fmt.Printf("   ⚠️  Warning: ExpenseAccountID %d not found for item %s, using default\n", 
					item.ExpenseAccountID, item.Product.Name)
				account = *defaultExpenseAccount
			}
		} else {
			account = *defaultExpenseAccount
		}

		journalItems = append(journalItems, models.SimpleSSOTJournalItem{
			JournalID:   ssotEntry.ID,
			AccountID:   account.ID,
			AccountCode: account.Code,
			AccountName: account.Name,
			Debit:       item.TotalPrice,
			Credit:      0,
			Description: fmt.Sprintf("Purchase - %s", item.Product.Name),
		})
		
		fmt.Printf("   Dr. %s (%s) = Rp %.2f\n", account.Name, account.Code, item.TotalPrice)
	}

	// DEBIT: PPN Masukan (if applicable)
	if purchase.PPNAmount > 0 {
		ppnAccount, err := resolveByCode("1240")
		if err != nil {
			return fmt.Errorf("failed to find PPN Masukan account (1240): %v", err)
		}

		journalItems = append(journalItems, models.SimpleSSOTJournalItem{
			JournalID:   ssotEntry.ID,
			AccountID:   ppnAccount.ID,
			AccountCode: ppnAccount.Code,
			AccountName: ppnAccount.Name,
			Debit:       purchase.PPNAmount,
			Credit:      0,
			Description: "PPN Masukan",
		})
		
		fmt.Printf("   Dr. %s (%s) = Rp %.2f\n", ppnAccount.Name, ppnAccount.Code, purchase.PPNAmount)
	}

	// CREDIT SIDE - Based on payment method
	var creditAccount *models.Account

	switch purchase.PaymentMethod {
	case "CASH", "TRANSFER":
		// Credit Cash/Bank Account
		if purchase.BankAccountID != nil {
			// Get the cash/bank account
			var cashBank models.CashBank
			if err := db.First(&cashBank, *purchase.BankAccountID).Error; err == nil {
				if cashBank.AccountID != 0 {
					if err := db.First(&creditAccount, cashBank.AccountID).Error; err != nil {
						return fmt.Errorf("failed to load cash/bank account: %v", err)
					}
				}
			}
		}
		
		// Fallback to default cash account
		if creditAccount == nil {
			creditAccount, err = resolveByCode("1110")
			if err != nil {
				return fmt.Errorf("failed to find default cash account (1110): %v", err)
			}
		}

	case "CREDIT":
		// Credit Accounts Payable (Hutang Usaha)
		creditAccount, err = resolveByCode("2100")
		if err != nil {
			return fmt.Errorf("failed to find accounts payable account (2100): %v", err)
		}

	default:
		// Default to Accounts Payable
		creditAccount, err = resolveByCode("2100")
		if err != nil {
			return fmt.Errorf("failed to find accounts payable account (2100): %v", err)
		}
	}

	journalItems = append(journalItems, models.SimpleSSOTJournalItem{
		JournalID:   ssotEntry.ID,
		AccountID:   creditAccount.ID,
		AccountCode: creditAccount.Code,
		AccountName: creditAccount.Name,
		Debit:       0,
		Credit:      purchase.TotalAmount,
		Description: fmt.Sprintf("Purchase payment - %s", purchase.PaymentMethod),
	})
	
	fmt.Printf("   Cr. %s (%s) = Rp %.2f\n", creditAccount.Name, creditAccount.Code, purchase.TotalAmount)

	// Create all journal items
	if err := db.Create(&journalItems).Error; err != nil {
		return fmt.Errorf("failed to create journal items: %v", err)
	}

	fmt.Printf("   Created %d journal line items\n", len(journalItems))

	return nil
}

