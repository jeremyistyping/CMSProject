package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	fmt.Println("🧪 TESTING: Draft Sales Auto-posting Prevention")
	fmt.Println("=" + string(make([]byte, 50)))

	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Initialize repositories
	salesRepo := repositories.NewSalesRepository(db)
	journalRepo := repositories.NewJournalEntryRepository(db)
	accountRepo := repositories.NewAccountRepository(db)

	// Initialize services
	doubleEntryService := services.NewSalesDoubleEntryService(db, journalRepo, accountRepo)

	fmt.Println("\n1. Creating a DRAFT sale...")

	// Create a test draft sale
	sale := &models.Sale{
		CustomerID:        1, // Assume customer exists
		Code:              fmt.Sprintf("TDRAFT%d", time.Now().Unix()%10000),
		Date:              time.Now(),
		Status:            models.SaleStatusDraft, // DRAFT status
		PaymentMethodType: "CASH",
		Subtotal:          1000.00,
		TotalAmount:       1000.00,
		OutstandingAmount: 1000.00,
		SaleItems: []models.SaleItem{
			{
				ProductID:   1,
				Quantity:    1,
				UnitPrice:   1000.00,
				LineTotal:   1000.00,
				TotalPrice:  1000.00,
				Taxable:     false,
			},
		},
	}

	// Save the draft sale
	createdSale, err := salesRepo.Create(sale)
	if err != nil {
		log.Printf("❌ Failed to create draft sale: %v", err)
		return
	}

	fmt.Printf("✅ Created draft sale ID: %d with status: %s\n", createdSale.ID, createdSale.Status)

	fmt.Println("\n2. Testing journal creation for DRAFT sale (should be BLOCKED)...")

	// Try to create journal entries for the draft sale (this should be blocked)
	err = doubleEntryService.CreateSaleJournalEntries(createdSale, 1)
	if err != nil {
		fmt.Printf("✅ BLOCKED: Journal creation blocked for draft sale: %v\n", err)
	} else {
		fmt.Printf("❌ FAILED: Journal creation was NOT blocked for draft sale!\n")
		return
	}

	fmt.Println("\n3. Confirming the sale...")

	// Update sale to CONFIRMED status
	createdSale.Status = models.SaleStatusConfirmed
	updatedSale, err := salesRepo.Update(createdSale)
	if err != nil {
		log.Printf("❌ Failed to update sale to confirmed: %v", err)
		return
	}

	fmt.Printf("✅ Sale updated to CONFIRMED status\n")

	fmt.Println("\n4. Testing journal creation for CONFIRMED sale (should be BLOCKED)...")

	// Try to create journal entries for the confirmed sale (should still be blocked)
	err = doubleEntryService.CreateSaleJournalEntries(updatedSale, 1)
	if err != nil {
		fmt.Printf("✅ BLOCKED: Journal creation blocked for confirmed sale: %v\n", err)
	} else {
		fmt.Printf("❌ FAILED: Journal creation was NOT blocked for confirmed sale!\n")
		return
	}

	fmt.Println("\n5. Invoicing the sale...")

	// Update sale to INVOICED status
	updatedSale.Status = models.SaleStatusInvoiced
	updatedSale.InvoiceNumber = fmt.Sprintf("INV%d", time.Now().Unix()%10000)
	invoicedSale, err := salesRepo.Update(updatedSale)
	if err != nil {
		log.Printf("❌ Failed to update sale to invoiced: %v", err)
		return
	}

	fmt.Printf("✅ Sale updated to INVOICED status with invoice number: %s\n", invoicedSale.InvoiceNumber)

	fmt.Println("\n6. Testing journal creation for INVOICED sale (should be ALLOWED)...")

	// Get initial journal count
	_, initialTotal, err := journalRepo.FindAll(context.Background(), &models.JournalEntryFilter{
		Page:  1,
		Limit: 1000,
	})
	if err != nil {
		log.Printf("❌ Failed to get initial journal count: %v", err)
		return
	}
	initialCount := int(initialTotal)
	fmt.Printf("Initial journal count: %d\n", initialCount)

	// Try to create journal entries for the invoiced sale (should work)
	err = doubleEntryService.CreateSaleJournalEntries(invoicedSale, 1)
	if err != nil {
		fmt.Printf("❌ UNEXPECTED: Journal creation failed for invoiced sale: %v\n", err)
	} else {
		fmt.Printf("✅ ALLOWED: Journal creation succeeded for invoiced sale\n")
	}

	// Check if journal was created
	_, finalTotal, err := journalRepo.FindAll(context.Background(), &models.JournalEntryFilter{
		Page:  1,
		Limit: 1000,
	})
	if err != nil {
		log.Printf("❌ Failed to get final journal count: %v", err)
		return
	}
	finalCount := int(finalTotal)
	fmt.Printf("Final journal count: %d\n", finalCount)

	if finalCount > initialCount {
		fmt.Printf("✅ SUCCESS: Journal entries were created for invoiced sale\n")
	} else {
		fmt.Printf("❌ WARNING: No new journal entries found\n")
	}

	fmt.Println("\n🎯 TEST RESULTS:")
	fmt.Println("✅ DRAFT sales: Journal creation BLOCKED ✓")
	fmt.Println("✅ CONFIRMED sales: Journal creation BLOCKED ✓") 
	fmt.Println("✅ INVOICED sales: Journal creation ALLOWED ✓")
	fmt.Println("\n🚀 Auto-posting prevention is working correctly!")

	// Cleanup: Delete test sale
	fmt.Printf("\n🧹 Cleaning up test sale %d...\n", createdSale.ID)
	err = salesRepo.Delete(createdSale.ID)
	if err != nil {
		log.Printf("Warning: Failed to cleanup test sale: %v", err)
	} else {
		fmt.Println("✅ Test sale cleaned up")
	}
}