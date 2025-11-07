package main

import (
	"errors"
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

	fmt.Println("🧪 Direct test of improved auto-migration...")
	fmt.Println("✅ Connected to database")

	// Check current state
	var existingWorkflow ApprovalWorkflow
	var stepCount int64
	result := db.Where("name = ? AND module = ?", "Standard Purchase Approval", "PURCHASE").First(&existingWorkflow)
	
	if result.Error == nil {
		fmt.Printf("✅ Standard Purchase Approval workflow found (ID: %d)\n", existingWorkflow.ID)
		
		// Check if workflow has steps
		db.Model(&ApprovalStep{}).Where("workflow_id = ?", existingWorkflow.ID).Count(&stepCount)
		fmt.Printf("📊 Current step count: %d\n", stepCount)
		
		if stepCount == 0 {
			fmt.Println("⚠️  Workflow exists but has no steps - creating steps...")
			
			// Create steps for existing workflow
			steps := []ApprovalStep{
				{
					WorkflowID:   existingWorkflow.ID,
					StepOrder:    1,
					StepName:     "Employee Submission",
					ApproverRole: "employee",
					IsOptional:   false,
					TimeLimit:    24,
				},
				{
					WorkflowID:   existingWorkflow.ID,
					StepOrder:    2,
					StepName:     "Finance Approval",
					ApproverRole: "finance",
					IsOptional:   false,
					TimeLimit:    48,
				},
				{
					WorkflowID:   existingWorkflow.ID,
					StepOrder:    3,
					StepName:     "Director Approval",
					ApproverRole: "director",
					IsOptional:   true,
					TimeLimit:    72,
				},
			}
			
			fmt.Printf("📝 Creating %d steps for workflow ID %d...\n", len(steps), existingWorkflow.ID)
			
			for i, step := range steps {
				fmt.Printf("   Creating step %d: %s - %s\n", i+1, step.StepName, step.ApproverRole)
				if err := db.Create(&step).Error; err != nil {
					fmt.Printf("❌ Failed to create step '%s': %v\n", step.StepName, err)
					return
				}
			}
			
			fmt.Printf("✅ Created %d workflow steps for Standard Purchase Approval\n", len(steps))
			fmt.Println("🎯 Standard Purchase Approval workflow setup completed!")
		} else {
			fmt.Printf("✅ Workflow already has %d steps - no action needed\n", stepCount)
		}
	} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("❌ Standard Purchase Approval workflow NOT FOUND!")
		fmt.Println("   This should not happen if auto-migration was working")
	} else {
		fmt.Printf("❌ Error checking workflow: %v\n", result.Error)
	}

	// Verify final state
	fmt.Println("\n🔍 Final verification:")
	db.Where("workflow_id = ?", existingWorkflow.ID).Model(&ApprovalStep{}).Count(&stepCount)
	fmt.Printf("   Workflow ID %d now has %d steps\n", existingWorkflow.ID, stepCount)
	
	fmt.Println("\n🎯 Direct migration test completed!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}