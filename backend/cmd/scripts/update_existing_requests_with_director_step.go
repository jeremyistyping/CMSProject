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

	fmt.Println("🔧 Updating Existing Approval Requests with Director Steps...")
	fmt.Println(strings.Repeat("=", 60))

	// Update existing approval requests
	err = updateExistingRequestsWithDirectorSteps(db)
	if err != nil {
		log.Printf("Error updating requests: %v", err)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✅ Existing Requests Update Completed!")
}

func updateExistingRequestsWithDirectorSteps(db *gorm.DB) error {
	fmt.Println("\n📋 Updating existing approval requests...")

	// Get all pending approval requests that don't have director steps
	var requests []models.ApprovalRequest
	err := db.Preload("Workflow.Steps").Preload("ApprovalSteps.Step").
		Where("status = ?", models.ApprovalStatusPending).
		Find(&requests).Error
	if err != nil {
		return fmt.Errorf("failed to query approval requests: %v", err)
	}

	fmt.Printf("Found %d pending approval requests\n", len(requests))

	for _, request := range requests {
		fmt.Printf("\nRequest: %s (Workflow: %s)\n", request.RequestCode, request.Workflow.Name)
		
		// Check if request has director step
		hasDirectorStep := false
		for _, action := range request.ApprovalSteps {
			if strings.ToLower(action.Step.ApproverRole) == "director" {
				hasDirectorStep = true
				break
			}
		}

		if !hasDirectorStep {
			// Find director step in workflow
			var directorStep *models.ApprovalStep
			for _, step := range request.Workflow.Steps {
				if strings.ToLower(step.ApproverRole) == "director" {
					directorStep = &step
					break
				}
			}

			if directorStep != nil {
				fmt.Printf("  ⚠️ Request missing director step. Adding...\n")
				
				// Create approval action for director step
				directorAction := models.ApprovalAction{
					RequestID: request.ID,
					StepID:    directorStep.ID,
					Status:    models.ApprovalStatusPending,
					IsActive:  false, // Not active until escalated
				}

				if err := db.Create(&directorAction).Error; err != nil {
					log.Printf("Failed to create director action for request %s: %v", request.RequestCode, err)
				} else {
					fmt.Printf("  ✓ Added director approval action\n")
				}
			} else {
				fmt.Printf("  ⚠️ Workflow doesn't have director step - skipping\n")
			}
		} else {
			fmt.Printf("  ✓ Already has director step\n")
		}
	}

	return nil
}
