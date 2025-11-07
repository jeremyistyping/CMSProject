package main

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"fmt"
	"log"
	"os"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Get database connection string from environment
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}
	
	// Initialize database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// Initialize repositories
	purchaseRepo := repositories.NewPurchaseRepository(db)
	productRepo := repositories.NewProductRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	
	// Initialize services
	approvalService := services.NewApprovalService(db)
	var journalService services.JournalServiceInterface = nil // Use nil for testing
	pdfService := services.NewPDFService()
	
	purchaseService := services.NewPurchaseService(
		db,
		purchaseRepo,
		productRepo,
		contactRepo,
		accountRepo,
		approvalService,
		journalService,
		pdfService,
	)
	
	// Test 1: Check current state
	fmt.Println("=== Testing Purchase Code Generation ===")
	
	// Get last purchase number for current month
	year := time.Now().Year()
	month := int(time.Now().Month())
	lastNumber, err := purchaseRepo.GetLastPurchaseNumberByMonth(year, month)
	if err != nil {
		log.Printf("Error getting last number: %v", err)
	} else {
		fmt.Printf("Last purchase number for %04d/%02d: %d\n", year, month, lastNumber)
	}
	
	// Check if PO/2025/09/0001 exists
	exists, err := purchaseRepo.CodeExists("PO/2025/09/0001")
	if err != nil {
		log.Printf("Error checking code: %v", err)
	} else {
		fmt.Printf("Code PO/2025/09/0001 exists: %v\n", exists)
	}
	
	// Test 2: Create a test purchase to verify unique code generation
	fmt.Println("\n=== Creating Test Purchase ===")
	
	testRequest := models.PurchaseCreateRequest{
		VendorID: 35,
		Date:     time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
		DueDate:  time.Date(2025, 10, 12, 0, 0, 0, 0, time.UTC),
		Discount: 10,
		Tax:      11,
		Notes:    "Test unique code generation",
		Items: []models.PurchaseItemRequest{
			{
				ProductID:        20,
				Quantity:         100,
				UnitPrice:        1000,
				Discount:         0,
				Tax:              0,
				ExpenseAccountID: 24,
			},
		},
	}
	
	purchase, err := purchaseService.CreatePurchase(testRequest, 11)
	if err != nil {
		log.Printf("Error creating purchase: %v", err)
	} else {
		fmt.Printf("Successfully created purchase with code: %s\n", purchase.Code)
		fmt.Printf("Purchase ID: %d\n", purchase.ID)
		fmt.Printf("Total Amount: %.2f\n", purchase.TotalAmount)
	}
	
	// Test 3: Try creating another purchase to verify unique code
	fmt.Println("\n=== Creating Second Test Purchase ===")
	
	testRequest2 := models.PurchaseCreateRequest{
		VendorID: 35,
		Date:     time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
		DueDate:  time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
		Discount: 5,
		Tax:      11,
		Notes:    "Second test for unique code",
		Items: []models.PurchaseItemRequest{
			{
				ProductID:        20,
				Quantity:         50,
				UnitPrice:        1500,
				Discount:         0,
				Tax:              0,
				ExpenseAccountID: 24,
			},
		},
	}
	
	purchase2, err := purchaseService.CreatePurchase(testRequest2, 11)
	if err != nil {
		log.Printf("Error creating second purchase: %v", err)
	} else {
		fmt.Printf("Successfully created second purchase with code: %s\n", purchase2.Code)
		fmt.Printf("Purchase ID: %d\n", purchase2.ID)
		fmt.Printf("Total Amount: %.2f\n", purchase2.TotalAmount)
	}
	
	// List all purchases for September 2025
	fmt.Println("\n=== Listing Purchases for 2025/09 ===")
	var purchases []models.Purchase
	err = db.Model(&models.Purchase{}).
		Where("code LIKE ?", "PO/2025/09/%").
		Order("code ASC").
		Find(&purchases).Error
	
	if err != nil {
		log.Printf("Error listing purchases: %v", err)
	} else {
		for _, p := range purchases {
			fmt.Printf("- %s (ID: %d, Created: %s)\n", p.Code, p.ID, p.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}
	
	fmt.Println("\n=== Test Complete ===")
}
