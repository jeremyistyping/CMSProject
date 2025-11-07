package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/sistem_akuntansi?sslmode=disable"
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔧 Adding Director Steps to Purchase Workflows...")
	fmt.Println(strings.Repeat("=", 60))

	// Get all purchase workflows that don't have director steps
	err = addDirectorStepsToWorkflows(db)
	if err != nil {
		log.Printf("Error adding director steps: %v", err)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✅ Director Steps Addition Completed!")
}

func addDirectorStepsToWorkflows(db *gorm.DB) error {
	fmt.Println("\n📋 Adding director steps to workflows that need them...")

	// Get all purchase workflows
	var workflows []models.ApprovalWorkflow
	err := db.Preload("Steps").
		Where("module = ? AND is_active = ?", models.ApprovalModulePurchase, true).
		Order("min_amount ASC").
		Find(&workflows).Error
	if err != nil {
		return fmt.Errorf("failed to query workflows: %v", err)
	}

	for _, workflow := range workflows {
		fmt.Printf("\nWorkflow: %s (%.0f - %.0f)\n", workflow.Name, workflow.MinAmount, workflow.MaxAmount)
		
		// Check if workflow has director step
		hasDirectorStep := false
		for _, step := range workflow.Steps {
			if strings.ToLower(step.ApproverRole) == "director" {
				hasDirectorStep = true
				break
			}
		}

		if !hasDirectorStep {
			fmt.Printf("  ⚠️ Missing director step. Adding...\n")
			
			// Find the highest step order
			maxOrder := 0
			for _, step := range workflow.Steps {
				if step.StepOrder > maxOrder {
					maxOrder = step.StepOrder
				}
			}

			// Create director step
			newStep := models.ApprovalStep{
				WorkflowID:   workflow.ID,
				StepOrder:    maxOrder + 1,
				StepName:     "Director Approval (Escalation)",
				ApproverRole: "director",
				IsOptional:   true, // Optional by default, can be made mandatory during escalation
				TimeLimit:    48,
			}

			if err := db.Create(&newStep).Error; err != nil {
				log.Printf("Failed to create director step for workflow %s: %v", workflow.Name, err)
			} else {
				fmt.Printf("  ✓ Added director step (order: %d)\n", newStep.StepOrder)
			}
		} else {
			fmt.Printf("  ✓ Already has director step\n")
		}
	}

	// Update all workflows to support escalation
	fmt.Println("\n🔧 Updating workflows to support director escalation...")
	
	err = db.Model(&models.ApprovalWorkflow{}).
		Where("module = ? AND is_active = ?", models.ApprovalModulePurchase, true).
		Update("require_director", false).Error // Set to false so they can use escalation
	if err != nil {
		log.Printf("Failed to update workflows: %v", err)
	} else {
		fmt.Println("✓ Updated all workflows to support escalation")
	}

	return nil
}
