package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("=== COMPREHENSIVE PAYMENT TEST & JOURNAL ANALYSIS ===")
	
	// Load configuration
	_ = config.LoadConfig()
	
	// Connect to database
	db := database.ConnectDB()
	
	// Step 1: Check current Sale status
	fmt.Println("\n🔍 Step 1: Checking Sale ID 1 status...")
	var sale models.Sale
	if err := db.Preload("Customer").Preload("SalePayments").First(&sale, 1).Error; err != nil {
		log.Fatal("Failed to get sale:", err)
	}
	
	fmt.Printf("📊 Sale Status:\n")
	fmt.Printf("  Sale ID: %d\n", sale.ID)
	fmt.Printf("  Code: %s\n", sale.Code)
	fmt.Printf("  Invoice Number: %s\n", sale.InvoiceNumber)
	fmt.Printf("  Customer: %s (ID: %d)\n", sale.Customer.Name, sale.CustomerID)
	fmt.Printf("  Status: %s\n", sale.Status)
	fmt.Printf("  Total Amount: %.2f\n", sale.TotalAmount)
	fmt.Printf("  Paid Amount: %.2f\n", sale.PaidAmount)
	fmt.Printf("  Outstanding Amount: %.2f\n", sale.OutstandingAmount)
	
	// Step 2: Check available Cash/Bank accounts
	fmt.Println("\n🏦 Step 2: Checking available Cash/Bank accounts...")
	var cashBanks []models.CashBank
	if err := db.Find(&cashBanks).Error; err != nil {
		log.Printf("Error getting cash banks: %v", err)
	} else {
		fmt.Printf("Available Cash/Bank accounts:\n")
		for _, cb := range cashBanks {
			fmt.Printf("  ID: %d, Name: %s, Balance: %.2f, Active: %v\n", 
				cb.ID, cb.Name, cb.Balance, cb.IsActive)
		}
	}
	
	// Step 3: Test API call directly
	fmt.Println("\n🚀 Step 3: Testing API call directly...")
	
	// Prepare payment data
	paymentData := map[string]interface{}{
		"amount":       3330000.0,
		"date":         time.Now().Format("2006-01-02T15:04:05Z07:00"),
		"method":       "BANK_TRANSFER",
		"cash_bank_id": 1, // Use first cash bank
		"reference":    "test-payment-comprehensive",
		"notes":        "Test payment from comprehensive script",
	}
	
	jsonData, err := json.Marshal(paymentData)
	if err != nil {
		log.Fatal("Failed to marshal JSON:", err)
	}
	
	// Make API call
	client := &http.Client{Timeout: 30 * time.Second}
	url := "http://localhost:8080/api/sales/1/integrated-payment"
	
	// Declare variables early to avoid goto issues
	var body []byte
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("Failed to create request:", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	// Add auth header if needed - you might need to adjust this
	req.Header.Set("Authorization", "Bearer your-token-here") // Replace with actual token
	
	fmt.Printf("Making POST request to: %s\n", url)
	fmt.Printf("Payload: %s\n", string(jsonData))
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ API call failed: %v", err)
		fmt.Println("\n🔄 Trying without auth header...")
		
		// Try without auth header
		req.Header.Del("Authorization")
		resp, err = client.Do(req)
		if err != nil {
			log.Printf("❌ API call failed again: %v", err)
			fmt.Println("\n⚠️ API endpoint seems to be down. Continuing with database analysis...")
			goto DatabaseAnalysis
		}
	}
	
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
	}
	
	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(body))
	
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("❌ Payment creation failed with status %d\n", resp.StatusCode)
		
		// Parse error response
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if details, ok := errorResp["details"]; ok {
				fmt.Printf("Error details: %v\n", details)
			}
			if expectedFields, ok := errorResp["expected_fields"]; ok {
				fmt.Printf("Expected fields: %v\n", expectedFields)
			}
		}
	} else {
		fmt.Printf("✅ Payment creation successful!\n")
	}

DatabaseAnalysis:
	// Step 4: Database Analysis
	fmt.Println("\n📊 Step 4: Current Database Analysis...")
	
	// Check payments for customer
	var payments []models.Payment
	if err := db.Where("contact_id = ?", sale.CustomerID).Order("created_at DESC").Limit(10).Find(&payments).Error; err != nil {
		log.Printf("Error getting payments: %v", err)
	} else {
		fmt.Printf("\n💰 Payments for customer %d:\n", sale.CustomerID)
		for i, payment := range payments {
			fmt.Printf("  %d. ID=%d, Code=%s, Amount=%.2f, Status=%s, Date=%s\n", 
				i+1, payment.ID, payment.Code, payment.Amount, payment.Status, 
				payment.Date.Format("2006-01-02"))
		}
	}
	
	// Check payment allocations
	var allocations []models.PaymentAllocation
	if err := db.Preload("Payment").Order("created_at DESC").Limit(10).Find(&allocations).Error; err != nil {
		log.Printf("Error getting allocations: %v", err)
	} else {
		fmt.Printf("\n🔗 Recent Payment Allocations:\n")
		for i, alloc := range allocations {
			invoiceID := "NULL"
			if alloc.InvoiceID != nil {
				invoiceID = fmt.Sprintf("%d", *alloc.InvoiceID)
			}
			fmt.Printf("  %d. PaymentID=%d, InvoiceID=%s, Amount=%.2f, CreatedAt=%s\n", 
				i+1, alloc.PaymentID, invoiceID, alloc.AllocatedAmount,
				alloc.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}
	
	// Check sale payments
	var salePayments []models.SalePayment
	if err := db.Where("sale_id = ?", 1).Order("created_at DESC").Find(&salePayments).Error; err != nil {
		log.Printf("Error getting sale payments: %v", err)
	} else {
		fmt.Printf("\n💳 Sale Payments for Sale ID 1:\n")
		for i, sp := range salePayments {
			paymentID := "NULL"
			if sp.PaymentID != nil {
				paymentID = fmt.Sprintf("%d", *sp.PaymentID)
			}
			fmt.Printf("  %d. ID=%d, Amount=%.2f, Method=%s, PaymentID=%s, Date=%s\n", 
				i+1, sp.ID, sp.Amount, sp.Method, paymentID, sp.Date.Format("2006-01-02"))
		}
	}
	
	// Step 5: Journal Analysis
	fmt.Println("\n📋 Step 5: JOURNAL ANALYSIS...")
	
	// Check journal entries related to Sale
	var saleJournals []models.JournalEntry
	if err := db.Preload("JournalLines.Account").Where("reference_type = ? AND reference_id = ?", models.JournalRefSale, 1).Find(&saleJournals).Error; err != nil {
		log.Printf("Error getting sale journals: %v", err)
	} else {
		fmt.Printf("\n📊 Journal Entries for Sale ID 1:\n")
		for i, journal := range saleJournals {
			fmt.Printf("  %d. ID=%d, Code=%s, Description=%s\n", i+1, journal.ID, journal.Code, journal.Description)
			fmt.Printf("     Status=%s, Date=%s, TotalDebit=%.2f, TotalCredit=%.2f\n", 
				journal.Status, journal.EntryDate.Format("2006-01-02"), journal.TotalDebit, journal.TotalCredit)
			
			// Show journal lines
			for j, line := range journal.JournalLines {
				fmt.Printf("       Line %d: %s - Debit=%.2f, Credit=%.2f, Account=%s\n", 
					j+1, line.Description, line.DebitAmount, line.CreditAmount, line.Account.Name)
			}
		}
	}
	
	// Check journal entries for payments
	var paymentJournals []models.JournalEntry
	if err := db.Preload("JournalLines.Account").Where("reference_type = ?", models.JournalRefPayment).Order("created_at DESC").Limit(5).Find(&paymentJournals).Error; err != nil {
		log.Printf("Error getting payment journals: %v", err)
	} else {
		fmt.Printf("\n💰 Recent Payment Journal Entries:\n")
		for i, journal := range paymentJournals {
			fmt.Printf("  %d. ID=%d, Code=%s, Description=%s\n", i+1, journal.ID, journal.Code, journal.Description)
			fmt.Printf("     Status=%s, Date=%s, TotalDebit=%.2f, TotalCredit=%.2f\n", 
				journal.Status, journal.EntryDate.Format("2006-01-02"), journal.TotalDebit, journal.TotalCredit)
			
			// Show journal lines
			for j, line := range journal.JournalLines {
				fmt.Printf("       Line %d: %s - Debit=%.2f, Credit=%.2f, Account=%s\n", 
					j+1, line.Description, line.DebitAmount, line.CreditAmount, line.Account.Name)
			}
		}
	}
	
	// Check accounts used
	fmt.Println("\n🏦 Step 6: Account Analysis...")
	var accounts []models.Account
	if err := db.Where("code IN (?)", []string{"1101", "1201", "4101", "1104"}).Find(&accounts).Error; err != nil {
		log.Printf("Error getting accounts: %v", err)
	} else {
		fmt.Printf("\nKey Accounts Status:\n")
		for _, acc := range accounts {
			fmt.Printf("  Code=%s, Name=%s, Type=%s, Balance=%.2f\n", 
				acc.Code, acc.Name, acc.Type, acc.Balance)
		}
	}
	
	// Final Sale Status Check
	fmt.Println("\n🔄 Step 7: Final Sale Status Check...")
	if err := db.Preload("Customer").First(&sale, 1).Error; err != nil {
		log.Printf("Error refreshing sale: %v", err)
	} else {
		fmt.Printf("📊 Updated Sale Status:\n")
		fmt.Printf("  Status: %s\n", sale.Status)
		fmt.Printf("  Total Amount: %.2f\n", sale.TotalAmount)
		fmt.Printf("  Paid Amount: %.2f\n", sale.PaidAmount)
		fmt.Printf("  Outstanding Amount: %.2f\n", sale.OutstandingAmount)
		fmt.Printf("  Updated At: %s\n", sale.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	
	// Recommendations
	fmt.Println("\n💡 ANALYSIS & RECOMMENDATIONS:")
	fmt.Println("1. Check if backend server is running on port 8080")
	fmt.Println("2. Verify authentication token is correct")
	fmt.Println("3. Check if Payment Management integration is working")
	fmt.Println("4. Ensure Journal Entry creation is enabled")
	fmt.Println("5. Verify Account mapping is correct (1201=AR, 1101/1104=Cash/Bank)")
	
	if sale.PaidAmount == 0 && len(salePayments) == 0 {
		fmt.Println("\n⚠️  ISSUE FOUND: No payments recorded for Sale ID 1")
		fmt.Println("   - Try making payment with smaller amount first")
		fmt.Println("   - Check backend logs for detailed errors")
		fmt.Println("   - Verify database constraints are not blocking payment creation")
	}
	
	if len(saleJournals) == 0 {
		fmt.Println("\n⚠️  ISSUE FOUND: No journal entries found for Sale ID 1")
		fmt.Println("   - Journal creation might be disabled")
		fmt.Println("   - Check if accounts exist (1201, 4101, etc.)")
		fmt.Println("   - Verify journal service integration")
	}
	
	fmt.Println("\n=== COMPREHENSIVE TEST COMPLETE ===")
}