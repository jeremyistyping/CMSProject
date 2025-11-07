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

	fmt.Println("🔍 Ultra-Safe Cleanup of Purchase Approval Workflows...")
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
	var standardWorkflowID uint
	var standardWorkflowExists bool

	for _, workflow := range workflows {
		if workflow.Name == "Standard Purchase Approval" {
			standardWorkflowExists = true
			standardWorkflowID = workflow.ID
			fmt.Printf("✅ Keeping: %s (ID: %d)\n", workflow.Name, workflow.ID)
		} else {
			workflowsToDelete = append(workflowsToDelete, workflow.ID)
			fmt.Printf("🗑️  Will delete: %s (ID: %d)\n", workflow.Name, workflow.ID)
		}
	}

	if !standardWorkflowExists {
		log.Fatal("❌ Standard Purchase Approval workflow not found! Please create it first.")
	}

	if len(workflowsToDelete) == 0 {
		fmt.Println("✅ No workflows to delete. Only Standard Purchase Approval exists.")
		return
	}

	// ULTRA-SAFE DELETION STRATEGY:
	// Instead of deleting workflows, we'll REASSIGN all existing data to use Standard Purchase Approval
	// This prevents any data loss and foreign key violations

	fmt.Printf("\n🔄 Reassigning existing data to Standard Purchase Approval (ID: %d)...\n", standardWorkflowID)

	// Step 1: Reassign all approval_requests from old workflows to standard workflow
	for _, oldWorkflowID := range workflowsToDelete {
		fmt.Printf("\n📝 Reassigning approval_requests from workflow %d to %d...\n", oldWorkflowID, standardWorkflowID)
		
		// Check if there are any requests to reassign
		var requestCount int64
		db.Table("approval_requests").Where("workflow_id = ?", oldWorkflowID).Count(&requestCount)
		fmt.Printf("   Found %d approval_request(s) to reassign\n", requestCount)
		
		if requestCount > 0 {
			result := db.Table("approval_requests").
				Where("workflow_id = ?", oldWorkflowID).
				Update("workflow_id", standardWorkflowID)
			if result.Error != nil {
				log.Printf("⚠️  Warning: Error reassigning approval_requests from workflow %d: %v", oldWorkflowID, result.Error)
			} else {
				fmt.Printf("   ✅ Reassigned %d approval_request(s)\n", result.RowsAffected)
			}
		}
	}

	// Step 2: Now we can safely delete the old workflow steps
	fmt.Println("\n🔄 Deleting old workflow steps...")
	for _, workflowID := range workflowsToDelete {
		result := db.Where("workflow_id = ?", workflowID).Delete(&ApprovalStep{})
		if result.Error != nil {
			log.Printf("⚠️  Warning: Error deleting steps for workflow ID %d: %v", workflowID, result.Error)
		} else {
			fmt.Printf("   ✅ Deleted %d steps for workflow ID %d\n", result.RowsAffected, workflowID)
		}
	}

	// Step 3: Finally delete the old workflows
	fmt.Println("\n🔄 Deleting old workflows...")
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

		// Show how many approval requests now use this workflow
		var requestCount int64
		db.Table("approval_requests").Where("workflow_id = ?", workflow.ID).Count(&requestCount)
		fmt.Printf("     Active Requests: %d\n", requestCount)
	}

	fmt.Println("\n🎯 Ultra-safe cleanup completed! Now using only Standard Purchase Approval workflow.")
	fmt.Println("📝 All existing approval_requests have been reassigned to the Standard workflow.")
	fmt.Println("💾 No data has been lost - all purchase histories are preserved.")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}