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

	fmt.Println("🧪 Testing Director Escalation Fix...")
	fmt.Println(strings.Repeat("=", 60))

	// Test 1: Verify workflows have director steps
	testWorkflowConfiguration(db)

	// Test 2: Simulate escalation scenario
	testEscalationScenario(db)

	// Test 3: Check for orphaned director steps
	checkOrphanedDirectorSteps(db)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✅ Testing completed!")
}

func testWorkflowConfiguration(db *gorm.DB) {
	fmt.Println("\n📝 Test 1: Verifying workflow configurations...")

	var workflows []models.ApprovalWorkflow
	err := db.Preload("Steps").
		Where("module = ? AND is_active = ?", models.ApprovalModulePurchase, true).
		Order("min_amount ASC").
		Find(&workflows).Error

	if err != nil {
		log.Printf("Failed to query workflows: %v", err)
		return
	}

	fmt.Printf("Found %d active purchase workflows\n", len(workflows))

	for _, workflow := range workflows {
		fmt.Printf("\n  Workflow: %s (%.0f - %.0f)\n", workflow.Name, workflow.MinAmount, workflow.MaxAmount)
		
		hasDirectorStep := false
		directorStepOptional := false
		directorStepOrder := 0
		
		for _, step := range workflow.Steps {
			fmt.Printf("    - Step %d: %s (Role: %s, Optional: %v)\n", 
				step.StepOrder, step.StepName, step.ApproverRole, step.IsOptional)
			
			if strings.ToLower(step.ApproverRole) == "director" {
				hasDirectorStep = true
				directorStepOptional = step.IsOptional
				directorStepOrder = step.StepOrder
			}
		}

		// Validate director step presence
		if workflow.RequireDirector || workflow.MinAmount >= 25000000 {
			if !hasDirectorStep {
				fmt.Printf("    ❌ ERROR: Workflow should have director step but doesn't!\n")
			} else {
				fmt.Printf("    ✓ Has director step (Order: %d, Optional: %v)\n", 
					directorStepOrder, directorStepOptional)
				
				// Check if director step should be mandatory for large amounts
				if workflow.MinAmount >= 25000000 && directorStepOptional {
					fmt.Printf("    ⚠️ WARNING: Director step is optional for large purchase (>= 25M)\n")
				}
			}
		} else {
			if hasDirectorStep {
				fmt.Printf("    ✓ Has optional director step for escalation\n")
			} else {
				fmt.Printf("    ℹ️ No director step (OK for small purchases)\n")
			}
		}
	}
}

func testEscalationScenario(db *gorm.DB) {
	fmt.Println("\n🔬 Test 2: Testing escalation scenario...")

	// Find a pending approval request that can be used for testing
	var testRequest models.ApprovalRequest
	err := db.Preload("ApprovalSteps.Step").
		Where("status = ?", models.ApprovalStatusPending).
		First(&testRequest).Error

	if err == gorm.ErrRecordNotFound {
		fmt.Println("  No pending requests found for testing")
		fmt.Println("  Creating a test scenario...")
		
		// Get or create a test purchase for simulation
		testRequest = createTestApprovalRequest(db)
		if testRequest.ID == 0 {
			fmt.Println("  ❌ Failed to create test approval request")
			return
		}
	}

	fmt.Printf("  Testing with request: %s (ID: %d)\n", testRequest.RequestCode, testRequest.ID)

	// Check current active step
	var currentActiveStep *models.ApprovalAction
	for i := range testRequest.ApprovalSteps {
		step := &testRequest.ApprovalSteps[i]
		if step.IsActive && step.Status == models.ApprovalStatusPending {
			currentActiveStep = step
			fmt.Printf("  Current active step: %s (Role: %s)\n", 
				step.Step.StepName, step.Step.ApproverRole)
			break
		}
	}

	// Check if director step exists
	var directorStep *models.ApprovalAction
	hasDirectorStep := false
	for i := range testRequest.ApprovalSteps {
		step := &testRequest.ApprovalSteps[i]
		if strings.ToLower(step.Step.ApproverRole) == "director" {
			hasDirectorStep = true
			directorStep = step
			break
		}
	}

	if hasDirectorStep {
		fmt.Printf("  ✓ Director step exists (Active: %v, Status: %s)\n", 
			directorStep.IsActive, directorStep.Status)
		
		// Simulate what happens when finance approves with escalation
		if currentActiveStep != nil && strings.ToLower(currentActiveStep.Step.ApproverRole) == "finance" {
			fmt.Println("\n  Simulating finance approval with escalation...")
			
			// Check if director step would be activated properly
			if directorStep.Status == models.ApprovalStatusPending && !directorStep.IsActive {
				fmt.Println("  ✓ Director step is pending and would be activated on escalation")
			} else if directorStep.IsActive {
				fmt.Println("  ⚠️ Director step is already active")
			} else {
				fmt.Println("  ❌ Director step status issue: ", directorStep.Status)
			}
		}
	} else {
		fmt.Println("  ⚠️ No director step in workflow - escalation would fail!")
		fmt.Println("     Solution: Add director step to workflow or create dynamically during escalation")
	}
}

func checkOrphanedDirectorSteps(db *gorm.DB) {
	fmt.Println("\n🔍 Test 3: Checking for orphaned director steps...")

	// Find director steps that are pending but not active
	var orphanedSteps []models.ApprovalAction
	err := db.Preload("Step").Preload("Request").
		Joins("JOIN approval_steps ON approval_actions.step_id = approval_steps.id").
		Where("LOWER(approval_steps.approver_role) = ? AND approval_actions.status = ? AND approval_actions.is_active = ?", 
			"director", models.ApprovalStatusPending, false).
		Find(&orphanedSteps).Error

	if err != nil {
		log.Printf("Failed to query orphaned steps: %v", err)
		return
	}

	if len(orphanedSteps) == 0 {
		fmt.Println("  ✓ No orphaned director steps found")
	} else {
		fmt.Printf("  ⚠️ Found %d orphaned director steps:\n", len(orphanedSteps))
		
		for _, step := range orphanedSteps {
			fmt.Printf("    - Request %s: Director step pending but not active\n", 
				step.Request.RequestCode)
			
			// Check if all previous steps are completed
			allPreviousCompleted := true
			var previousSteps []string
			
			for _, otherStep := range step.Request.ApprovalSteps {
				if otherStep.Step.StepOrder < step.Step.StepOrder {
					if otherStep.Status == models.ApprovalStatusPending {
						allPreviousCompleted = false
						previousSteps = append(previousSteps, 
							fmt.Sprintf("%s (Order: %d)", otherStep.Step.StepName, otherStep.Step.StepOrder))
					}
				}
			}
			
			if allPreviousCompleted {
				fmt.Printf("      → All previous steps completed. Director step should be active!\n")
				fmt.Printf("      → FIX: Activate this director step\n")
			} else {
				fmt.Printf("      → Previous steps still pending: %s\n", strings.Join(previousSteps, ", "))
				fmt.Printf("      → This is normal workflow progression\n")
			}
		}
	}
}

func createTestApprovalRequest(db *gorm.DB) models.ApprovalRequest {
	// This is a helper function to create a test approval request
	// In real scenario, this would be created through the normal purchase flow
	
	fmt.Println("    Creating test approval request...")
	
	// Find a workflow with director step
	var workflow models.ApprovalWorkflow
	err := db.Preload("Steps").
		Where("module = ? AND is_active = ? AND require_director = ?", 
			models.ApprovalModulePurchase, true, true).
		First(&workflow).Error
	
	if err != nil {
		fmt.Printf("    Failed to find suitable workflow: %v\n", err)
		return models.ApprovalRequest{}
	}
	
	fmt.Printf("    Using workflow: %s\n", workflow.Name)
	
	// Create a minimal approval request for testing
	request := models.ApprovalRequest{
		RequestCode:    fmt.Sprintf("TEST-%s", time.Now().Format("20060102150405")),
		WorkflowID:     workflow.ID,
		RequesterID:    1, // Assuming user ID 1 exists
		EntityType:     models.EntityTypePurchase,
		EntityID:       999999, // Dummy ID for testing
		Amount:         50000000, // 50M - should require director
		Status:         models.ApprovalStatusPending,
		Priority:       models.ApprovalPriorityNormal,
		RequestTitle:   "Test Purchase for Director Escalation",
		RequestMessage: "This is a test request to verify director escalation workflow",
	}
	
	if err := db.Create(&request).Error; err != nil {
		fmt.Printf("    Failed to create test request: %v\n", err)
		return models.ApprovalRequest{}
	}
	
	// Create approval actions for each workflow step
	for _, step := range workflow.Steps {
		action := models.ApprovalAction{
			RequestID: request.ID,
			StepID:    step.ID,
			Status:    models.ApprovalStatusPending,
			IsActive:  step.StepOrder == 1, // Only first step is active initially
		}
		
		if err := db.Create(&action).Error; err != nil {
			fmt.Printf("    Failed to create action for step %s: %v\n", step.StepName, err)
		}
	}
	
	// Reload with all relations
	db.Preload("ApprovalSteps.Step").First(&request, request.ID)
	
	fmt.Printf("    ✓ Created test request: %s\n", request.RequestCode)
	return request
}
