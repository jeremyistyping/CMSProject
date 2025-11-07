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

	fmt.Println("🔍 Detailed Check of Purchase Approval Workflows...")
	fmt.Println("✅ Connected to database")

	// Check workflows
	var workflows []ApprovalWorkflow
	result := db.Where("module = ?", "PURCHASE").Find(&workflows)
	if result.Error != nil {
		log.Fatal("❌ Error fetching workflows:", result.Error)
	}

	fmt.Printf("\n📋 Found %d PURCHASE workflow(s):\n", len(workflows))
	for _, workflow := range workflows {
		fmt.Printf("\n🔧 Workflow: %s (ID: %d)\n", workflow.Name, workflow.ID)
		fmt.Printf("   - Amount Range: %.0f - %.0f\n", workflow.MinAmount, workflow.MaxAmount)
		fmt.Printf("   - Active: %t\n", workflow.IsActive)
		fmt.Printf("   - Require Director: %t\n", workflow.RequireDirector)
		fmt.Printf("   - Require Finance: %t\n", workflow.RequireFinance)
		
		// Check steps for this workflow
		var steps []ApprovalStep
		db.Where("workflow_id = ?", workflow.ID).Order("step_order").Find(&steps)
		fmt.Printf("   - Steps: %d\n", len(steps))
		
		if len(steps) == 0 {
			fmt.Printf("   ⚠️  WARNING: No steps found for this workflow!\n")
		} else {
			for _, step := range steps {
				optional := ""
				if step.IsOptional {
					optional = " (Optional)"
				}
				fmt.Printf("       %d. %s - %s%s (%dh)\n", 
					step.StepOrder, step.StepName, step.ApproverRole, optional, step.TimeLimit)
			}
		}
	}

	// Check for the specific "Standard Purchase Approval" workflow
	fmt.Println("\n🎯 Checking 'Standard Purchase Approval' specifically...")
	var standardWorkflow ApprovalWorkflow
	result = db.Where("name = ? AND module = ?", "Standard Purchase Approval", "PURCHASE").First(&standardWorkflow)
	
	if result.Error != nil {
		fmt.Println("❌ Standard Purchase Approval workflow NOT FOUND!")
		fmt.Printf("   Error: %v\n", result.Error)
		
		// Check if any workflow has similar name
		var similarWorkflows []ApprovalWorkflow
		db.Where("name ILIKE ? AND module = ?", "%Standard%", "PURCHASE").Find(&similarWorkflows)
		if len(similarWorkflows) > 0 {
			fmt.Println("\n🔍 Found similar workflows:")
			for _, wf := range similarWorkflows {
				fmt.Printf("   - %s (ID: %d)\n", wf.Name, wf.ID)
			}
		}
	} else {
		fmt.Printf("✅ Standard Purchase Approval found (ID: %d)\n", standardWorkflow.ID)
		
		// Check steps
		var steps []ApprovalStep
		db.Where("workflow_id = ?", standardWorkflow.ID).Order("step_order").Find(&steps)
		
		if len(steps) == 0 {
			fmt.Printf("🚨 CRITICAL: Standard Purchase Approval has NO STEPS!\n")
			fmt.Printf("   This explains why employee submissions are failing\n")
		} else {
			fmt.Printf("✅ Standard Purchase Approval has %d steps:\n", len(steps))
			for _, step := range steps {
				optional := ""
				if step.IsOptional {
					optional = " (Optional)"
				}
				fmt.Printf("   %d. %s - %s%s (%dh)\n", 
					step.StepOrder, step.StepName, step.ApproverRole, optional, step.TimeLimit)
			}
		}
	}

	fmt.Println("\n📊 Summary:")
	fmt.Printf("   - Total PURCHASE workflows: %d\n", len(workflows))
	
	var totalSteps int
	for _, workflow := range workflows {
		var stepCount int64
		db.Model(&ApprovalStep{}).Where("workflow_id = ?", workflow.ID).Count(&stepCount)
		totalSteps += int(stepCount)
	}
	fmt.Printf("   - Total approval steps: %d\n", totalSteps)
	
	if totalSteps == 0 {
		fmt.Printf("🚨 ISSUE IDENTIFIED: No workflow steps exist!\n")
		fmt.Printf("   This is why employees cannot submit purchases for approval\n")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}