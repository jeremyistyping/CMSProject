package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ApprovalWorkflow struct {
	ID              uint    `gorm:"primaryKey"`
	Name            string  `gorm:"not null;size:100"`
	Module          string  `gorm:"not null;size:50"`
	MinAmount       float64 `gorm:"type:decimal(15,2);default:0"`
	MaxAmount       float64 `gorm:"type:decimal(15,2)"`
	IsActive        bool    `gorm:"default:true"`
	RequireDirector bool    `gorm:"default:false"`
	RequireFinance  bool    `gorm:"default:false"`
}

type ApprovalStep struct {
	ID           uint   `gorm:"primaryKey"`
	WorkflowID   uint   `gorm:"not null;index"`
	StepOrder    int    `gorm:"not null"`
	StepName     string `gorm:"not null;size:100"`
	ApproverRole string `gorm:"not null;size:50"`
	IsOptional   bool   `gorm:"default:false"`
	TimeLimit    int    `gorm:"default:24"`
}

type ApprovalRequest struct {
	ID         uint `gorm:"primaryKey"`
	WorkflowID uint `gorm:"not null;index"`
	// Add other fields as needed
}

type ApprovalAction struct {
	ID     uint `gorm:"primaryKey"`
	StepID uint `gorm:"not null;index"`
	// Add other fields as needed
}

func main() {
	// Database connection using environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "sistem_akuntansi")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	fmt.Println("🔍 Safe Cleanup of Purchase Approval Workflows...")
	fmt.Println("✅ Connected to database")

	// First, let's see what workflows we have
	var workflows []ApprovalWorkflow
	result := db.Where("module = ?", "PURCHASE").Find(&workflows)
	if result.Error != nil {
		log.Fatal("❌ Error fetching workflows:", result.Error)
	}

	fmt.Printf("📋 Found %d PURCHASE workflow(s):\n", len(workflows))
	for _, workflow := range workflows {
		fmt.Printf("   - %s (%.0f-%.0f) [Active: %t] [ID: %d]\n", 
			workflow.Name, workflow.MinAmount, workflow.MaxAmount, workflow.IsActive, workflow.ID)
	}

	// Identify workflows to delete (amount-based ones)
	var workflowsToDelete []uint
	var stepsToDelete []uint
	var standardWorkflowExists bool

	for _, workflow := range workflows {
		if workflow.Name == "Standard Purchase Approval" {
			standardWorkflowExists = true
			fmt.Printf("✅ Keeping: %s (ID: %d)\n", workflow.Name, workflow.ID)
		} else {
			workflowsToDelete = append(workflowsToDelete, workflow.ID)
			fmt.Printf("🗑️  Will delete: %s (ID: %d)\n", workflow.Name, workflow.ID)
			
			// Get steps for this workflow
			var steps []ApprovalStep
			db.Where("workflow_id = ?", workflow.ID).Find(&steps)
			for _, step := range steps {
				stepsToDelete = append(stepsToDelete, step.ID)
			}
		}
	}

	if !standardWorkflowExists {
		log.Fatal("❌ Standard Purchase Approval workflow not found! Please create it first.")
	}

	if len(workflowsToDelete) == 0 {
		fmt.Println("✅ No workflows to delete. Only Standard Purchase Approval exists.")
		return
	}

	fmt.Printf("📋 Will delete %d workflows and %d steps\n", len(workflowsToDelete), len(stepsToDelete))

	// SAFE DELETION ORDER (to handle foreign key constraints):
	// 1. Delete approval_actions that reference steps
	// 2. Delete approval_requests that reference workflows  
	// 3. Delete approval_steps
	// 4. Delete approval_workflows

	// Step 1: Delete approval_actions
	if len(stepsToDelete) > 0 {
		fmt.Println("\n🔄 Step 1: Deleting approval_actions...")
		result := db.Where("step_id IN ?", stepsToDelete).Delete(&ApprovalAction{})
		if result.Error != nil {
			log.Printf("⚠️  Warning: Error deleting approval_actions: %v", result.Error)
		} else {
			fmt.Printf("   ✅ Deleted %d approval_action(s)\n", result.RowsAffected)
		}
	}

	// Step 2: Delete approval_requests
	if len(workflowsToDelete) > 0 {
		fmt.Println("\n🔄 Step 2: Deleting approval_requests...")
		result := db.Where("workflow_id IN ?", workflowsToDelete).Delete(&ApprovalRequest{})
		if result.Error != nil {
			log.Printf("⚠️  Warning: Error deleting approval_requests: %v", result.Error)
		} else {
			fmt.Printf("   ✅ Deleted %d approval_request(s)\n", result.RowsAffected)
		}
	}

	// Step 3: Delete workflow steps
	fmt.Println("\n🔄 Step 3: Deleting workflow steps...")
	for _, workflowID := range workflowsToDelete {
		result := db.Where("workflow_id = ?", workflowID).Delete(&ApprovalStep{})
		if result.Error != nil {
			log.Printf("⚠️  Warning: Error deleting steps for workflow ID %d: %v", workflowID, result.Error)
		} else {
			fmt.Printf("   ✅ Deleted %d steps for workflow ID %d\n", result.RowsAffected, workflowID)
		}
	}

	// Step 4: Delete workflows
	fmt.Println("\n🔄 Step 4: Deleting workflows...")
	result = db.Where("id IN ?", workflowsToDelete).Delete(&ApprovalWorkflow{})
	if result.Error != nil {
		log.Fatal("❌ Error deleting workflows:", result.Error)
	}
	fmt.Printf("✅ Deleted %d workflow(s)\n", result.RowsAffected)

	// Verify final state
	fmt.Println("\n🔍 Final workflow state:")
	var finalWorkflows []ApprovalWorkflow
	db.Where("module = ?", "PURCHASE").Find(&finalWorkflows)
	
	for _, workflow := range finalWorkflows {
		fmt.Printf("   - %s (%.0f-%.0f) [Active: %t]\n", 
			workflow.Name, workflow.MinAmount, workflow.MaxAmount, workflow.IsActive)
		
		// Show steps
		var steps []ApprovalStep
		db.Where("workflow_id = ?", workflow.ID).Order("step_order").Find(&steps)
		fmt.Printf("     Steps: %d\n", len(steps))
		for _, step := range steps {
			optional := ""
			if step.IsOptional {
				optional = " (Optional)"
			}
			fmt.Printf("       Step %d: %s - %s%s (%dh)\n", 
				step.StepOrder, step.StepName, step.ApproverRole, optional, step.TimeLimit)
		}
	}

	fmt.Println("\n🎯 Safe cleanup completed! Now using only Standard Purchase Approval workflow.")
	fmt.Println("📝 All related approval_actions and approval_requests have been cleaned up.")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}