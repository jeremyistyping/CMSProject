package main

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
)

func main() {
	log.Printf("🧪 Testing Sales Creation After Stock Fix")
	log.Printf("========================================")

	// Connect to database
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Initialize repositories
	salesRepo := repositories.NewSalesRepository(db)
	productRepo := repositories.NewProductRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	accountRepo := repositories.NewAccountRepository(db)

	// Initialize sales service
	salesService := services.NewSalesService(db, salesRepo, productRepo, contactRepo, accountRepo, nil, nil)

	// Get a product with stock
	var product models.Product
	err = db.Where("stock > ?", 0).First(&product).Error
	if err != nil {
		log.Fatalf("❌ No product with stock found: %v", err)
	}

	log.Printf("✅ Found product with stock: ID=%d, Name=%s, Stock=%d", 
		product.ID, product.Name, product.Stock)

	// Get first customer
	var customer models.Contact
	err = db.Where("type = ?", "CUSTOMER").First(&customer).Error
	if err != nil {
		log.Fatalf("❌ No customer found: %v", err)
	}

	// Get cash bank account
	var cashBank models.CashBank
	err = db.Where("is_active = ?", true).First(&cashBank).Error
	if err != nil {
		log.Fatalf("❌ No active cash bank found: %v", err)
	}

	// Create test sale request
	saleRequest := models.SaleCreateRequest{
		CustomerID:        customer.ID,
		Type:              models.SaleTypeInvoice,
		Date:              time.Now(),
		DueDate:           time.Now().AddDate(0, 0, 30),
		PaymentMethodType: "CASH",
		CashBankID:        &cashBank.ID,
		Items: []models.SaleItemRequest{
			{
				ProductID: product.ID,
				Quantity:  1,
				UnitPrice: product.SalePrice,
				Taxable:   true,
			},
		},
	}

	log.Printf("🔄 Creating test sale...")
	log.Printf("   Customer: %s", customer.Name)
	log.Printf("   Product: %s (Qty: 1, Price: %.2f)", product.Name, product.SalePrice)
	log.Printf("   Payment: CASH via %s", cashBank.Name)

	// Create sale
	createdSale, err := salesService.CreateSale(saleRequest, 1)
	if err != nil {
		log.Printf("❌ Failed to create sale: %v", err)
		return
	}

	log.Printf("✅ Sale created successfully!")
	log.Printf("   Sale ID: %d", createdSale.ID)
	log.Printf("   Sale Code: %s", createdSale.Code)
	log.Printf("   Status: %s", createdSale.Status)
	log.Printf("   Total Amount: %.2f", createdSale.TotalAmount)

	// Test sale confirmation
	log.Printf("\n🔄 Confirming sale...")
	err = salesService.ConfirmSale(createdSale.ID, 1)
	if err != nil {
		log.Printf("❌ Failed to confirm sale: %v", err)
		return
	}

	// Get updated sale details
	updatedSale, err := salesService.GetSaleByID(createdSale.ID)
	if err != nil {
		log.Printf("❌ Failed to get updated sale: %v", err)
		return
	}

	log.Printf("✅ Sale confirmed successfully!")
	log.Printf("   Status: %s", updatedSale.Status)
	log.Printf("   Paid Amount: %.2f", updatedSale.PaidAmount)
	log.Printf("   Outstanding: %.2f", updatedSale.OutstandingAmount)

	// Verify expected results
	if updatedSale.Status == models.SaleStatusPaid && updatedSale.OutstandingAmount == 0 {
		log.Printf("\n🎉 🎉 SUCCESS! Sales creation and confirmation working correctly!")
		log.Printf("✅ Stock validation passed")
		log.Printf("✅ Sale created and confirmed")
		log.Printf("✅ Status updated to PAID for cash sale")
	} else {
		log.Printf("\n⚠️  Partial success - sale created but status issue:")
		log.Printf("   Expected: PAID with 0 outstanding")
		log.Printf("   Actual: %s with %.2f outstanding", updatedSale.Status, updatedSale.OutstandingAmount)
	}

	// Cleanup - delete the test sale
	log.Printf("\n🧹 Cleaning up test data...")
	db.Delete(&models.SalePayment{}, "sale_id = ?", createdSale.ID)
	db.Delete(&models.SaleItem{}, "sale_id = ?", createdSale.ID)
	db.Delete(&models.Sale{}, "id = ?", createdSale.ID)
	log.Printf("✅ Cleanup completed")
}