package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database config from environment
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	
	// Construct DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database successfully!")

	// Check purchase ID 6
	var purchase models.Purchase
	err = db.Preload("Vendor").Preload("PurchaseItems.Product").First(&purchase, 6).Error
	if err != nil {
		log.Fatalf("Failed to fetch purchase: %v", err)
	}

	fmt.Printf("\n=== Purchase ID 6 Details ===\n")
	fmt.Printf("Code: %s\n", purchase.Code)
	fmt.Printf("Vendor: %s (ID: %d)\n", purchase.Vendor.Name, purchase.VendorID)
	fmt.Printf("Status: %s\n", purchase.Status)
	fmt.Printf("Approval Status: %s\n", purchase.ApprovalStatus)
	fmt.Printf("Total Amount: %.2f\n", purchase.TotalAmount)
	fmt.Printf("Approval Base Amount: %.2f\n", purchase.ApprovalBaseAmount)
	fmt.Printf("Approval Amount Basis: %s\n", purchase.ApprovalAmountBasis)
	fmt.Printf("Requires Approval: %v\n", purchase.RequiresApproval)
	
	if purchase.ApprovalRequestID != nil {
		fmt.Printf("Approval Request ID: %d\n", *purchase.ApprovalRequestID)
		
		// Check approval request details
		var approvalRequest models.ApprovalRequest
		err = db.Preload("Workflow").Preload("ApprovalSteps.Step").First(&approvalRequest, *purchase.ApprovalRequestID).Error
		if err != nil {
			fmt.Printf("Error fetching approval request: %v\n", err)
		} else {
			fmt.Printf("\n--- Approval Request Details ---\n")
			fmt.Printf("Request Code: %s\n", approvalRequest.RequestCode)
			fmt.Printf("Status: %s\n", approvalRequest.Status)
			fmt.Printf("Workflow: %s (ID: %d)\n", approvalRequest.Workflow.Name, approvalRequest.WorkflowID)
			fmt.Printf("Created At: %v\n", approvalRequest.CreatedAt)
			
			// Show approval steps
			fmt.Printf("\n--- Approval Steps ---\n")
			for _, action := range approvalRequest.ApprovalSteps {
				fmt.Printf("Step %d: Role=%s, Status=%s, Active=%v\n", 
					action.Step.StepOrder, 
					action.Step.ApproverRole,
					action.Status,
					action.IsActive)
			}
		}
	} else {
		fmt.Println("No Approval Request ID")
	}
	
	fmt.Printf("\n=== Purchase Items ===\n")
	for i, item := range purchase.PurchaseItems {
		fmt.Printf("Item %d: Product=%s, Qty=%d, Price=%.2f, Total=%.2f\n", 
			i+1, item.Product.Name, item.Quantity, item.UnitPrice, item.TotalPrice)
	}
	
	// Pretty print the purchase as JSON
	purchaseJSON, _ := json.MarshalIndent(purchase, "", "  ")
	fmt.Printf("\n=== Full Purchase JSON ===\n%s\n", string(purchaseJSON))
}
