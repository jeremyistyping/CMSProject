package main

import (
	"log"
	"time"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
)

func main() {
	log.Printf("🔬 Testing Unified Payment Service")
	
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// Initialize repositories
	salesRepo := repositories.NewSalesRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	
	// Initialize UNIFIED service
	unifiedService := services.NewUnifiedSalesPaymentService(db, salesRepo, accountRepo)
	
	// Test validation
	testRequest := models.SalePaymentRequest{
		Amount:        500000.0,
		PaymentDate:   time.Now(),
		PaymentMethod: "BANK_TRANSFER",
		Reference:     "TEST-UNIFIED",
		Notes:         "Test unified payment service",
	}
	
	if err := unifiedService.ValidatePaymentRequest(testRequest); err != nil {
		log.Printf("❌ Validation failed: %v", err)
	} else {
		log.Printf("✅ Validation passed")
	}
	
	log.Printf("🎉 Unified Payment Service is ready!")
}