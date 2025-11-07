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

	// Begin transaction
	tx := db.Begin()

	// Update existing workflows to include Employee as first step
	var existingWorkflows []models.ApprovalWorkflow
	if err := tx.Preload("Steps").Where("module = ?", models.ApprovalModulePurchase).Find(&existingWorkflows).Error; err != nil {
		tx.Rollback()
		log.Fatal("Failed to fetch existing workflows:", err)
	}

	fmt.Printf("Found %d existing workflows to update\n", len(existingWorkflows))

	for _, workflow := range existingWorkflows {
		fmt.Printf("\nUpdating workflow: %s\n", workflow.Name)
		
		// Check if Employee step already exists
		hasEmployeeStep := false
		for _, step := range workflow.Steps {
			if step.ApproverRole == "employee" {
				hasEmployeeStep = true
				break
			}
		}

		if hasEmployeeStep {
			fmt.Printf("  - Employee step already exists, skipping\n")
			continue
		}

		// Increment existing step orders by 1
		for _, step := range workflow.Steps {
			if err := tx.Model(&step).Update("step_order", step.StepOrder+1).Error; err != nil {
				tx.Rollback()
				log.Fatal("Failed to update step order:", err)
			}
			fmt.Printf("  - Updated step order: %s (Order: %d -> %d)\n", step.StepName, step.StepOrder, step.StepOrder+1)
		}

		// Add Employee step as first step
		employeeStep := models.ApprovalStep{
			WorkflowID:   workflow.ID,
			StepOrder:    1,
			StepName:     "Employee Submission",
			ApproverRole: "employee",
			IsOptional:   false,
			TimeLimit:    24,
		}

		if err := tx.Create(&employeeStep).Error; err != nil {
			tx.Rollback()
			log.Fatal("Failed to create employee step:", err)
		}
		fmt.Printf("  - Created Employee step: %s (Order: 1, Role: employee)\n", employeeStep.StepName)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}

	fmt.Println("\n✅ Approval workflows updated successfully!")
	
	// Verify the updated workflows
	var updatedWorkflows []models.ApprovalWorkflow
	db.Preload("Steps").Where("module = ?", models.ApprovalModulePurchase).Find(&updatedWorkflows)
	
	fmt.Printf("\nUpdated %d workflows for Purchase module:\n", len(updatedWorkflows))
	for _, w := range updatedWorkflows {
		fmt.Printf("\n- %s (%.0f - %.0f)\n", w.Name, w.MinAmount, w.MaxAmount)
		
		// Sort steps by order for display
		steps := w.Steps
		for i := 0; i < len(steps)-1; i++ {
			for j := i + 1; j < len(steps); j++ {
				if steps[i].StepOrder > steps[j].StepOrder {
					steps[i], steps[j] = steps[j], steps[i]
				}
			}
		}
		
		for _, s := range steps {
			fmt.Printf("  Step %d: %s (Role: %s)\n", s.StepOrder, s.StepName, s.ApproverRole)
		}
	}
}
