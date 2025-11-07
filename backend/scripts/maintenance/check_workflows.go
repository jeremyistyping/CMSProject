package main

import (
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

	// Check approval workflows
	var workflows []models.ApprovalWorkflow
	result := db.Find(&workflows)
	if result.Error != nil {
		log.Fatalf("Failed to fetch workflows: %v", result.Error)
	}

	fmt.Printf("\n=== Approval Workflows ===\n")
	fmt.Printf("Total workflows: %d\n\n", len(workflows))
	
	for _, w := range workflows {
		fmt.Printf("ID: %d\n", w.ID)
		fmt.Printf("Name: %s\n", w.Name)
		fmt.Printf("Module: %s\n", w.Module)
		fmt.Printf("Active: %v\n", w.IsActive)
		fmt.Printf("Min Amount: %.2f\n", w.MinAmount)
		fmt.Printf("Max Amount: %.2f\n", w.MaxAmount)
		fmt.Printf("---\n")
	}

	// Check approval steps
	var steps []models.ApprovalStep
	result = db.Find(&steps)
	if result.Error != nil {
		log.Fatalf("Failed to fetch approval steps: %v", result.Error)
	}

	fmt.Printf("\n=== Approval Steps ===\n")
	fmt.Printf("Total steps: %d\n\n", len(steps))
	
	for _, s := range steps {
		fmt.Printf("ID: %d\n", s.ID)
		fmt.Printf("Workflow ID: %d\n", s.WorkflowID)
		fmt.Printf("Step Order: %d\n", s.StepOrder)
		fmt.Printf("Approver Role: %s\n", s.ApproverRole)
		fmt.Printf("Parallel: %v\n", s.IsParallel)
		fmt.Printf("---\n")
	}

	// Check if there's any active PURCHASE workflow
	var purchaseWorkflow models.ApprovalWorkflow
	err = db.Where("module = ? AND is_active = ?", "PURCHASE", true).First(&purchaseWorkflow).Error
	if err != nil {
		fmt.Printf("\n⚠️  WARNING: No active PURCHASE workflow found! Error: %v\n", err)
		fmt.Println("This will cause 500 errors when trying to approve purchases.")
		fmt.Println("\nCreating default approval workflows...")
		
		// Create default workflows
		createDefaultWorkflows(db)
	} else {
		fmt.Printf("\n✓ Found active PURCHASE workflow: %s (ID: %d)\n", purchaseWorkflow.Name, purchaseWorkflow.ID)
	}
}

func createDefaultWorkflows(db *gorm.DB) {
	// Create standard purchase workflow
	workflow1 := models.ApprovalWorkflow{
		Name:        "Standard Purchase Approval",
		Module:      "PURCHASE",
		MinAmount:   0,
		MaxAmount:   25000000,
		IsActive:    true,
	}
	
	if err := db.Create(&workflow1).Error; err != nil {
		fmt.Printf("Failed to create standard workflow: %v\n", err)
	} else {
		fmt.Printf("✓ Created standard purchase workflow (ID: %d)\n", workflow1.ID)
		
		// Create steps for standard workflow
		steps := []models.ApprovalStep{
			{
				WorkflowID:   workflow1.ID,
				StepOrder:    1,
				ApproverRole: "employee",
				IsParallel:   false,
			},
			{
				WorkflowID:   workflow1.ID,
				StepOrder:    2,
				ApproverRole: "finance",
				IsParallel:   false,
			},
		}
		
		for _, step := range steps {
			if err := db.Create(&step).Error; err != nil {
				fmt.Printf("Failed to create step: %v\n", err)
			}
		}
	}
	
	// Create high-value purchase workflow
	workflow2 := models.ApprovalWorkflow{
		Name:        "High-Value Purchase Approval",
		Module:      "PURCHASE",
		MinAmount:   25000001,
		MaxAmount:   999999999999,
		IsActive:    true,
	}
	
	if err := db.Create(&workflow2).Error; err != nil {
		fmt.Printf("Failed to create high-value workflow: %v\n", err)
	} else {
		fmt.Printf("✓ Created high-value purchase workflow (ID: %d)\n", workflow2.ID)
		
		// Create steps for high-value workflow
		steps := []models.ApprovalStep{
			{
				WorkflowID:   workflow2.ID,
				StepOrder:    1,
				ApproverRole: "employee",
				IsParallel:   false,
			},
			{
				WorkflowID:   workflow2.ID,
				StepOrder:    2,
				ApproverRole: "finance",
				IsParallel:   false,
			},
			{
				WorkflowID:   workflow2.ID,
				StepOrder:    3,
				ApproverRole: "director",
				IsParallel:   false,
			},
		}
		
		for _, step := range steps {
			if err := db.Create(&step).Error; err != nil {
				fmt.Printf("Failed to create step: %v\n", err)
			}
		}
	}

	// Create sales workflow
	workflow3 := models.ApprovalWorkflow{
		Name:        "Standard Sales Approval",
		Module:      "SALES",
		MinAmount:   0,
		MaxAmount:   999999999999,
		IsActive:    true,
	}
	
	if err := db.Create(&workflow3).Error; err != nil {
		fmt.Printf("Failed to create sales workflow: %v\n", err)
	} else {
		fmt.Printf("✓ Created sales workflow (ID: %d)\n", workflow3.ID)
		
		// Create steps for sales workflow
		steps := []models.ApprovalStep{
			{
				WorkflowID:   workflow3.ID,
				StepOrder:    1,
				ApproverRole: "employee",
				IsParallel:   false,
			},
			{
				WorkflowID:   workflow3.ID,
				StepOrder:    2,
				ApproverRole: "finance",
				IsParallel:   false,
			},
		}
		
		for _, step := range steps {
			if err := db.Create(&step).Error; err != nil {
				fmt.Printf("Failed to create step: %v\n", err)
			}
		}
	}
	
	fmt.Println("\n✓ Default workflows created successfully!")
}
