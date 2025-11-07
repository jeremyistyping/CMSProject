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
	// Load .env file
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database connection string
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPass, dbName, dbPort)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Create approval workflows
	workflows := []models.ApprovalWorkflow{
		{
			Name:            "Small Purchase Approval",
			Module:          models.ApprovalModulePurchase,
			MinAmount:       0,
			MaxAmount:       5000000, // 5 million
			RequireDirector: false,
			RequireFinance:  true,
			IsActive:        true,
		},
		{
			Name:            "Medium Purchase Approval",
			Module:          models.ApprovalModulePurchase,
			MinAmount:       5000000,
			MaxAmount:       25000000, // 25 million
			RequireDirector: false,
			RequireFinance:  true,
			IsActive:        true,
		},
		{
			Name:            "Large Purchase Approval",
			Module:          models.ApprovalModulePurchase,
			MinAmount:       25000000,
			MaxAmount:       100000000, // 100 million
			RequireDirector: true,
			RequireFinance:  true,
			IsActive:        true,
		},
		{
			Name:            "Very Large Purchase Approval",
			Module:          models.ApprovalModulePurchase,
			MinAmount:       100000000,
			MaxAmount:       0, // No upper limit
			RequireDirector: true,
			RequireFinance:  true,
			IsActive:        true,
		},
	}

	// Workflow steps configuration with Employee as first step
	workflowSteps := map[string][]models.ApprovalStep{
		"Small Purchase Approval": {
			{
				StepOrder:    1,
				StepName:     "Employee Submission",
				ApproverRole: "employee",
				IsOptional:   false,
				TimeLimit:    24, // 24 hours
			},
			{
				StepOrder:    2,
				StepName:     "Finance Approval",
				ApproverRole: "finance",
				IsOptional:   false,
				TimeLimit:    24,
			},
			{
				StepOrder:    3,
				StepName:     "Director Approval (Optional)",
				ApproverRole: "director",
				IsOptional:   true,
				TimeLimit:    48,
			},
		},
		"Medium Purchase Approval": {
			{
				StepOrder:    1,
				StepName:     "Employee Submission",
				ApproverRole: "employee",
				IsOptional:   false,
				TimeLimit:    24,
			},
			{
				StepOrder:    2,
				StepName:     "Finance Approval",
				ApproverRole: "finance",
				IsOptional:   false,
				TimeLimit:    24,
			},
			{
				StepOrder:    3,
				StepName:     "Director Approval (Optional)",
				ApproverRole: "director",
				IsOptional:   true,
				TimeLimit:    48,
			},
		},
		"Large Purchase Approval": {
			{
				StepOrder:    1,
				StepName:     "Employee Submission",
				ApproverRole: "employee",
				IsOptional:   false,
				TimeLimit:    24,
			},
			{
				StepOrder:    2,
				StepName:     "Finance Approval",
				ApproverRole: "finance",
				IsOptional:   false,
				TimeLimit:    24,
			},
			{
				StepOrder:    3,
				StepName:     "Director Approval",
				ApproverRole: "director",
				IsOptional:   false,
				TimeLimit:    48,
			},
		},
		"Very Large Purchase Approval": {
			{
				StepOrder:    1,
				StepName:     "Employee Submission",
				ApproverRole: "employee",
				IsOptional:   false,
				TimeLimit:    24,
			},
			{
				StepOrder:    2,
				StepName:     "Finance Approval",
				ApproverRole: "finance",
				IsOptional:   false,
				TimeLimit:    24,
			},
			{
				StepOrder:    3,
				StepName:     "Director Approval",
				ApproverRole: "director",
				IsOptional:   false,
				TimeLimit:    48,
			},
			{
				StepOrder:    4,
				StepName:     "Admin Final Approval",
				ApproverRole: "admin",
				IsOptional:   false,
				TimeLimit:    48,
			},
		},
	}

	// Begin transaction
	tx := db.Begin()

	// Check if workflows already exist
	var count int64
	tx.Model(&models.ApprovalWorkflow{}).Where("module = ?", models.ApprovalModulePurchase).Count(&count)
	
	if count > 0 {
		fmt.Printf("Found %d existing purchase approval workflows. Skipping creation.\n", count)
		tx.Rollback()
		return
	}

	// Create workflows and steps
	for _, workflow := range workflows {
		fmt.Printf("Creating workflow: %s\n", workflow.Name)
		
		if err := tx.Create(&workflow).Error; err != nil {
			tx.Rollback()
			log.Fatal("Failed to create workflow:", err)
		}

		// Create steps for this workflow
		if steps, ok := workflowSteps[workflow.Name]; ok {
			for _, step := range steps {
				step.WorkflowID = workflow.ID
				fmt.Printf("  - Creating step: %s (Order: %d, Role: %s)\n", 
					step.StepName, step.StepOrder, step.ApproverRole)
				
				if err := tx.Create(&step).Error; err != nil {
					tx.Rollback()
					log.Fatal("Failed to create step:", err)
				}
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("\n✅ Approval workflows created successfully!")
	
	// Verify the created workflows
	var createdWorkflows []models.ApprovalWorkflow
	db.Preload("Steps").Where("module = ?", models.ApprovalModulePurchase).Find(&createdWorkflows)
	
	fmt.Printf("\nCreated %d workflows for Purchase module:\n", len(createdWorkflows))
	for _, w := range createdWorkflows {
		fmt.Printf("\n- %s (%.0f - %.0f)\n", w.Name, w.MinAmount, w.MaxAmount)
		for _, s := range w.Steps {
			fmt.Printf("  Step %d: %s (Role: %s)\n", s.StepOrder, s.StepName, s.ApproverRole)
		}
	}
}
