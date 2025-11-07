package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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

	fmt.Println("🔧 Starting Director Escalation Bug Fix...")
	fmt.Println(strings.Repeat("=", 60))

	// 1. Check and fix workflow configurations
	err = fixWorkflowConfigurations(db)
	if err != nil {
		log.Printf("Error fixing workflow configurations: %v", err)
	}

	// 2. Fix purchases that were incorrectly approved without director review
	err = fixIncorrectlyApprovedPurchases(db) 
	if err != nil {
		log.Printf("Error fixing incorrectly approved purchases: %v", err)
	}

	// 3. Ensure director steps are properly configured in workflows
	err = ensureDirectorStepsInWorkflows(db)
	if err != nil {
		log.Printf("Error ensuring director steps: %v", err)
	}

	// 4. Validate escalation logic
	err = validateEscalationLogic(db)
	if err != nil {
		log.Printf("Error validating escalation logic: %v", err)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✅ Director Escalation Bug Fix Completed!")
}

// fixWorkflowConfigurations ensures workflows have proper director steps
func fixWorkflowConfigurations(db *gorm.DB) error {
	fmt.Println("\n📋 Checking workflow configurations...")

	// Get all purchase workflows
	var workflows []models.ApprovalWorkflow
	err := db.Preload("Steps").
		Where("module = ? AND is_active = ?", models.ApprovalModulePurchase, true).
		Find(&workflows).Error
	if err != nil {
		return fmt.Errorf("failed to query workflows: %v", err)
	}

	for _, workflow := range workflows {
		fmt.Printf("\nWorkflow: %s (%.0f - %.0f)\n", workflow.Name, workflow.MinAmount, workflow.MaxAmount)
		
		// Check if workflow has director step
		hasDirectorStep := false
		var directorStep *models.ApprovalStep
		
		for i := range workflow.Steps {
			step := &workflow.Steps[i]
			if strings.ToLower(step.ApproverRole) == "director" {
				hasDirectorStep = true
				directorStep = step
				break
			}
		}

		if workflow.RequireDirector && !hasDirectorStep {
			// Workflow requires director but doesn't have director step
			fmt.Printf("  ⚠️ Workflow requires director but missing director step. Adding...\n")
			
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
				StepName:     "Director Approval",
				ApproverRole: "director",
				IsOptional:   false,
				TimeLimit:    48,
			}

			if err := db.Create(&newStep).Error; err != nil {
				log.Printf("Failed to create director step for workflow %s: %v", workflow.Name, err)
			} else {
				fmt.Printf("  ✓ Added director step (order: %d)\n", newStep.StepOrder)
			}
		} else if hasDirectorStep {
			// Check if director step is properly positioned
			fmt.Printf("  ✓ Has director step: %s (order: %d, optional: %v)\n", 
				directorStep.StepName, directorStep.StepOrder, directorStep.IsOptional)
			
			// For escalation scenarios, director step should ALWAYS be mandatory when triggered
			if directorStep.IsOptional && workflow.MinAmount >= 25000000 {
				// Large and Very Large purchases should have mandatory director approval
				fmt.Printf("    ⚠️ Director step is optional but should be mandatory for large purchases. Fixing...\n")
				err := db.Model(directorStep).Update("is_optional", false).Error
				if err != nil {
					log.Printf("Failed to update director step: %v", err)
				} else {
					fmt.Printf("    ✓ Updated director step to mandatory\n")
				}
			}
		}
	}

	// Ensure all workflows have proper escalation capability
	// Create an escalation workflow if it doesn't exist
	var escalationWorkflow models.ApprovalWorkflow
	err = db.Where("name = ?", "Escalated Purchase Approval").First(&escalationWorkflow).Error
	if err == gorm.ErrRecordNotFound {
		fmt.Println("\n📝 Creating escalation workflow...")
		
		escalationWorkflow = models.ApprovalWorkflow{
			Name:            "Escalated Purchase Approval",
			Module:          models.ApprovalModulePurchase,
			MinAmount:       0,
			MaxAmount:       0, // No limit - used for all escalations
			RequireDirector: true,
			RequireFinance:  true,
			IsActive:        true,
		}

		if err := db.Create(&escalationWorkflow).Error; err != nil {
			return fmt.Errorf("failed to create escalation workflow: %v", err)
		}

		// Create steps for escalation workflow
		steps := []models.ApprovalStep{
			{
				WorkflowID:   escalationWorkflow.ID,
				StepOrder:    1,
				StepName:     "Finance Review",
				ApproverRole: "finance",
				IsOptional:   false,
				TimeLimit:    24,
			},
			{
				WorkflowID:   escalationWorkflow.ID,
				StepOrder:    2,
				StepName:     "Director Final Approval",
				ApproverRole: "director",
				IsOptional:   false,
				TimeLimit:    48,
			},
		}

		for _, step := range steps {
			if err := db.Create(&step).Error; err != nil {
				log.Printf("Failed to create step: %v", err)
			}
		}

		fmt.Println("✓ Created escalation workflow")
	}

	return nil
}

// fixIncorrectlyApprovedPurchases fixes purchases that were approved without proper director review
func fixIncorrectlyApprovedPurchases(db *gorm.DB) error {
	fmt.Println("\n🔍 Checking for incorrectly approved purchases...")

	// Find purchases that were approved after escalation but without director approval
	var histories []models.ApprovalHistory
	err := db.Preload("Request.ApprovalSteps.Step").
		Where("action = ? AND created_at > ?", "ESCALATED_TO_DIRECTOR", time.Now().Add(-30*24*time.Hour)).
		Find(&histories).Error
	if err != nil {
		return fmt.Errorf("failed to query escalation history: %v", err)
	}

	fmt.Printf("Found %d escalation events in the last 30 days\n", len(histories))

	for _, history := range histories {
		request := history.Request
		
		// Check if request was approved without director step being completed
		directorApproved := false
		for _, step := range request.ApprovalSteps {
			if strings.ToLower(step.Step.ApproverRole) == "director" && 
			   step.Status == models.ApprovalStatusApproved {
				directorApproved = true
				break
			}
		}

		if request.Status == models.ApprovalStatusApproved && !directorApproved {
			fmt.Printf("\n⚠️ Request %s was approved without director approval after escalation\n", request.RequestCode)
			
			// Get the purchase
			var purchase models.Purchase
			err := db.Where("approval_request_id = ?", request.ID).First(&purchase).Error
			if err != nil {
				log.Printf("Failed to find purchase for request %d: %v", request.ID, err)
				continue
			}

			fmt.Printf("  Purchase: %s (Amount: %.2f)\n", purchase.Code, purchase.TotalAmount)
			
			// Option 1: Revert to pending for director approval
			// Option 2: Log for manual review
			fmt.Printf("  📌 Marked for manual review - Purchase may need director approval\n")
			
			// Create a notification for admin/director
			notification := models.Notification{
				UserID:   1, // Admin user
				Type:     models.NotificationTypeApprovalPending,
				Title:    fmt.Sprintf("Review Required: Purchase %s", purchase.Code),
				Message:  fmt.Sprintf("Purchase %s (%.2f) was approved after escalation but may not have proper director approval", 
					purchase.Code, purchase.TotalAmount),
				Priority: models.ApprovalPriorityUrgent,
				Data:     fmt.Sprintf(`{"purchase_id":%d,"request_id":%d}`, purchase.ID, request.ID),
			}
			db.Create(&notification)
		}
	}

	return nil
}

// ensureDirectorStepsInWorkflows ensures all workflows have proper director steps configured
func ensureDirectorStepsInWorkflows(db *gorm.DB) error {
	fmt.Println("\n🔧 Ensuring director steps in workflows...")

	// Get all active purchase workflows
	var workflows []models.ApprovalWorkflow
	err := db.Preload("Steps").
		Where("module = ? AND is_active = ?", models.ApprovalModulePurchase, true).
		Order("min_amount ASC").
		Find(&workflows).Error
	if err != nil {
		return fmt.Errorf("failed to query workflows: %v", err)
	}

	for _, workflow := range workflows {
		// Check if this workflow should have a director step based on amount
		shouldHaveDirector := workflow.RequireDirector || workflow.MinAmount >= 25000000

		if shouldHaveDirector {
			// Ensure director step exists and is properly configured
			hasProperDirectorStep := false
			
			for _, step := range workflow.Steps {
				if strings.ToLower(step.ApproverRole) == "director" {
					// Check if it's the last or second-to-last step (after finance)
					isProperlyOrdered := step.StepOrder >= len(workflow.Steps)-1
					
					if !isProperlyOrdered {
						fmt.Printf("⚠️ Workflow '%s': Director step has incorrect order (%d). Should be near the end.\n", 
							workflow.Name, step.StepOrder)
					} else {
						hasProperDirectorStep = true
					}
					break
				}
			}

			if !hasProperDirectorStep {
				fmt.Printf("⚠️ Workflow '%s' needs a properly ordered director step\n", workflow.Name)
			}
		}
	}

	return nil
}

// validateEscalationLogic validates that escalation logic works correctly
func validateEscalationLogic(db *gorm.DB) error {
	fmt.Println("\n✔️ Validating escalation logic...")

	// Test scenario: When finance escalates, director step should become active
	fmt.Println("Testing escalation scenarios:")
	
	// Get a sample pending approval request
	var testRequest models.ApprovalRequest
	err := db.Preload("ApprovalSteps.Step").
		Where("status = ?", models.ApprovalStatusPending).
		First(&testRequest).Error
	
	if err == nil {
		fmt.Printf("  Sample request: %s\n", testRequest.RequestCode)
		
		// Check if director step exists
		var directorStep *models.ApprovalAction
		for i := range testRequest.ApprovalSteps {
			step := &testRequest.ApprovalSteps[i]
			if strings.ToLower(step.Step.ApproverRole) == "director" {
				directorStep = step
				fmt.Printf("  ✓ Director step found (Active: %v, Status: %s)\n", 
					step.IsActive, step.Status)
				break
			}
		}

		if directorStep == nil {
			fmt.Println("  ⚠️ No director step in workflow - escalation would fail")
		}
	} else {
		fmt.Println("  No pending requests to test")
	}

	// Check for orphaned director steps (director steps that are pending but not active)
	var orphanedSteps []models.ApprovalAction
	err = db.Preload("Step").Preload("Request").
		Joins("JOIN approval_steps ON approval_actions.step_id = approval_steps.id").
		Where("LOWER(approval_steps.approver_role) = ? AND approval_actions.status = ? AND approval_actions.is_active = ?", 
			"director", models.ApprovalStatusPending, false).
		Find(&orphanedSteps).Error
	
	if err == nil && len(orphanedSteps) > 0 {
		fmt.Printf("\n⚠️ Found %d orphaned director steps (pending but not active)\n", len(orphanedSteps))
		for _, step := range orphanedSteps {
			fmt.Printf("  - Request %s: Director step pending but not active\n", step.Request.RequestCode)
			
			// Check if this should be activated
			allPreviousCompleted := true
			for _, otherStep := range step.Request.ApprovalSteps {
				if otherStep.Step.StepOrder < step.Step.StepOrder && 
				   otherStep.Status == models.ApprovalStatusPending {
					allPreviousCompleted = false
					break
				}
			}

			if allPreviousCompleted {
				fmt.Printf("    → Activating director step...\n")
				err := db.Model(&step).Update("is_active", true).Error
				if err != nil {
					log.Printf("Failed to activate director step: %v", err)
				} else {
					fmt.Printf("    ✓ Activated\n")
				}
			}
		}
	}

	return nil
}

