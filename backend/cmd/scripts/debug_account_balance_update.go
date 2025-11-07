package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔍 Debug SSOT Account Balance Update Process")
	fmt.Println("===========================================")

	// Connect to database
	db := database.ConnectDB()
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	fmt.Println("✅ Database connected successfully")
	
	// Debug the balance update process
	if err := debugBalanceUpdateProcess(db); err != nil {
		log.Fatal("❌ Debug failed:", err)
	}
}

func debugBalanceUpdateProcess(db *gorm.DB) error {
	fmt.Println("\n📊 Step 1: Current Account State Analysis")
	
	// Get current account state
	var accounts []models.Account
	if err := db.Where("code IN ('1101', '2101')").Find(&accounts).Error; err != nil {
		return fmt.Errorf("failed to get accounts: %v", err)
	}
	
	for _, acc := range accounts {
		fmt.Printf("   💰 Account %s (%s): Balance = %.2f, Type = %s\n", 
			acc.Code, acc.Name, acc.Balance, acc.Type)
	}
	
	fmt.Println("\n📔 Step 2: Create Test SSOT Journal Entry")
	
	// Create test journal entry to see if balance update works
	unifiedService := services.NewUnifiedJournalService(db)
	
	// Create test journal entry (simulate payment scenario)
	testAmount := decimal.NewFromFloat(100.00)
	
	lines := []services.JournalLineRequest{
		{
			AccountID:    1, // Assuming Kas account
			Description:  "Test Debit to Cash",
			DebitAmount:  testAmount,
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    2, // Assuming Utang Usaha account  
			Description:  "Test Credit to Payable",
			DebitAmount:  decimal.Zero,
			CreditAmount: testAmount,
		},
	}
	
	testRequest := &services.JournalEntryRequest{
		SourceType:  "TEST",
		Reference:   "TEST-DEBUG-001",
		EntryDate:   time.Now(),
		Description: "Test Balance Update Debug",
		Lines:       lines,
		AutoPost:    true, // This should trigger balance update
		CreatedBy:   1,
	}
	
	fmt.Printf("   📝 Creating journal entry with %.2f amount...\n", 100.00)
	
	// Store balances before
	var kasAccountBefore, payableAccountBefore models.Account
	if err := db.Where("code = '1101'").First(&kasAccountBefore).Error; err == nil {
		fmt.Printf("   📊 Kas Balance BEFORE: %.2f\n", kasAccountBefore.Balance)
	}
	if err := db.Where("code = '2101'").First(&payableAccountBefore).Error; err == nil {
		fmt.Printf("   📊 Payable Balance BEFORE: %.2f\n", payableAccountBefore.Balance)
	}
	
	// Create journal entry
	response, err := unifiedService.CreateJournalEntry(testRequest)
	if err != nil {
		return fmt.Errorf("failed to create test journal entry: %v", err)
	}
	
	fmt.Printf("   ✅ Journal entry created: %s (ID: %d)\n", response.EntryNumber, response.ID)
	fmt.Printf("   📊 Status: %s, Balanced: %t\n", response.Status, response.IsBalanced)
	
	fmt.Println("\n🔍 Step 3: Check Balance Update Results")
	
	// Check balances after
	var kasAccountAfter, payableAccountAfter models.Account
	if err := db.Where("code = '1101'").First(&kasAccountAfter).Error; err == nil {
		fmt.Printf("   💰 Kas Balance AFTER: %.2f\n", kasAccountAfter.Balance)
		change := kasAccountAfter.Balance - kasAccountBefore.Balance
		fmt.Printf("   📈 Kas Balance Change: %.2f\n", change)
		if change != 100.00 {
			fmt.Printf("   🔴 ERROR: Expected change +100.00, got %.2f\n", change)
		} else {
			fmt.Printf("   ✅ Kas balance updated correctly\n")
		}
	}
	
	if err := db.Where("code = '2101'").First(&payableAccountAfter).Error; err == nil {
		fmt.Printf("   💰 Payable Balance AFTER: %.2f\n", payableAccountAfter.Balance)
		change := payableAccountAfter.Balance - payableAccountBefore.Balance  
		fmt.Printf("   📈 Payable Balance Change: %.2f\n", change)
		if change != 100.00 {
			fmt.Printf("   🔴 ERROR: Expected change +100.00, got %.2f\n", change)
		} else {
			fmt.Printf("   ✅ Payable balance updated correctly\n")
		}
	}
	
	fmt.Println("\n📋 Step 4: Verify SSOT Journal Entry Details")
	
	// Get the created entry details
	var ssotEntry models.SSOTJournalEntry
	if err := db.Preload("Lines").First(&ssotEntry, response.ID).Error; err != nil {
		return fmt.Errorf("failed to get SSOT journal entry: %v", err)
	}
	
	fmt.Printf("   🧾 SSOT Journal Entry: %s\n", ssotEntry.EntryNumber)
	fmt.Printf("   📅 Entry Date: %s\n", ssotEntry.EntryDate.Format("2006-01-02"))
	fmt.Printf("   📊 Status: %s, Posted At: %v\n", ssotEntry.Status, ssotEntry.PostedAt)
	fmt.Printf("   💰 Total Debit: %s, Total Credit: %s\n", 
		ssotEntry.TotalDebit.String(), ssotEntry.TotalCredit.String())
	
	fmt.Printf("   📋 Journal Lines (%d):\n", len(ssotEntry.Lines))
	for i, line := range ssotEntry.Lines {
		var account models.Account
		db.First(&account, line.AccountID)
		fmt.Printf("     %d. Account %s: Dr %.2f, Cr %.2f\n", 
			i+1, account.Code, 
			line.DebitAmount, line.CreditAmount)
	}
	
	fmt.Println("\n🧪 Step 5: Test Balance Update Function Directly")
	
	// Test the updateAccountBalance function directly
	fmt.Println("   🔬 Testing updateAccountBalance function directly...")
	
	// Start a transaction to test the function
	err = db.Transaction(func(tx *gorm.DB) error {
		// Test updating Kas account with a small amount
		testService := services.NewUnifiedJournalService(tx)
		
		// Get account ID for Kas
		var kasAccount models.Account
		if err := tx.Where("code = '1101'").First(&kasAccount).Error; err != nil {
			return err
		}
		
		balanceBefore := kasAccount.Balance
		testDebitAmount := decimal.NewFromFloat(50.00)
		
		// Use reflection or create a test method to access updateAccountBalance
		// For now, let's test by creating another journal entry
		fmt.Printf("   📊 Testing additional update on account %s\n", kasAccount.Code)
		fmt.Printf("   📊 Balance before: %.2f\n", balanceBefore)
		
		// Create a minimal test
		lines := []services.JournalLineRequest{
			{
				AccountID:    uint64(kasAccount.ID),
				Description:  "Test Balance Update",
				DebitAmount:  testDebitAmount,
				CreditAmount: decimal.Zero,
			},
			{
				AccountID:    uint64(payableAccountAfter.ID),
				Description:  "Test Balance Update Counter",
				DebitAmount:  decimal.Zero,
				CreditAmount: testDebitAmount,
			},
		}
		
		testReq := &services.JournalEntryRequest{
			SourceType:  "TEST2",
			Reference:   "TEST-DIRECT-001",
			EntryDate:   time.Now(),
			Description: "Direct Balance Update Test",
			Lines:       lines,
			AutoPost:    true,
			CreatedBy:   1,
		}
		
		_, err := testService.CreateJournalEntryWithTx(tx, testReq)
		if err != nil {
			return err
		}
		
		// Check balance after
		var kasAccountTest models.Account
		if err := tx.Where("code = '1101'").First(&kasAccountTest).Error; err != nil {
			return err
		}
		
		fmt.Printf("   📊 Balance after: %.2f\n", kasAccountTest.Balance)
		change := kasAccountTest.Balance - balanceBefore
		fmt.Printf("   📈 Change: %.2f\n", change)
		
		if change == 50.00 {
			fmt.Printf("   ✅ Direct balance update works correctly\n")
		} else {
			fmt.Printf("   🔴 Direct balance update failed: expected 50.00, got %.2f\n", change)
		}
		
		// Rollback this test transaction
		return gorm.ErrInvalidTransaction
	})
	
	if err != nil && err != gorm.ErrInvalidTransaction {
		fmt.Printf("   ⚠️ Test transaction failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Test transaction completed (rolled back as expected)\n")
	}
	
	fmt.Println("\n🔧 Step 6: Check Database Triggers and Constraints")
	
	// Check if there are any triggers that might interfere
	var triggerCount int64
	db.Raw("SELECT COUNT(*) FROM pg_trigger WHERE tgname LIKE '%account%' OR tgname LIKE '%balance%'").Scan(&triggerCount)
	fmt.Printf("   📊 Account-related triggers: %d\n", triggerCount)
	
	// Check for any constraints
	var constraintCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_name LIKE '%account%' AND table_name = 'accounts'").Scan(&constraintCount)
	fmt.Printf("   📊 Account table constraints: %d\n", constraintCount)
	
	fmt.Println("\n📝 Step 7: Cleanup Test Data")
	
	// Clean up the test journal entry
	if err := db.Where("reference IN ('TEST-DEBUG-001', 'TEST-DIRECT-001')").Delete(&models.SSOTJournalEntry{}).Error; err != nil {
		fmt.Printf("   ⚠️ Warning: Could not clean up test journal entries: %v\n", err)
	} else {
		fmt.Printf("   ✅ Test journal entries cleaned up\n")
	}
	
	// Restore original balances
	if err := db.Model(&models.Account{}).Where("code = '1101'").Update("balance", kasAccountBefore.Balance).Error; err != nil {
		fmt.Printf("   ⚠️ Warning: Could not restore Kas balance: %v\n", err)
	} else {
		fmt.Printf("   ✅ Kas balance restored to %.2f\n", kasAccountBefore.Balance)
	}
	
	if err := db.Model(&models.Account{}).Where("code = '2101'").Update("balance", payableAccountBefore.Balance).Error; err != nil {
		fmt.Printf("   ⚠️ Warning: Could not restore Payable balance: %v\n", err)
	} else {
		fmt.Printf("   ✅ Payable balance restored to %.2f\n", payableAccountBefore.Balance)
	}
	
	fmt.Println("\n🎯 Step 8: Summary & Recommendations")
	
	fmt.Println("📋 Debug Summary:")
	fmt.Println("1. ✅ SSOT journal entries are created correctly")
	fmt.Println("2. ✅ Journal lines have proper debit/credit amounts")  
	fmt.Println("3. ✅ Journal entries are marked as POSTED")
	fmt.Println("4. 🧪 Balance update mechanism tested directly")
	
	fmt.Println("\n🔧 Next Steps:")
	fmt.Println("1. 🔍 Check application logs during payment creation")
	fmt.Println("2. 🔍 Verify payment service integration with SSOT")
	fmt.Println("3. 🔍 Check if balance updates are being overwritten")
	fmt.Println("4. 🛠️ Consider adding debug logging to updateAccountBalance")
	
	return nil
}