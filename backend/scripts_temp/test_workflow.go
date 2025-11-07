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
	fmt.Println("🔍 Testing Purchase Approval Workflow")

	// Connect to database
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "sistem_akuntansi")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Connected to database")

	// Test workflow query
	var workflows []ApprovalWorkflow
	if err := db.Where("module = ?", "PURCHASE").Find(&workflows).Error; err != nil {
		log.Fatal("Failed to query workflows:", err)
	}

	fmt.Printf("📋 Found %d PURCHASE workflow(s):\n", len(workflows))
	for _, w := range workflows {
		fmt.Printf("   - %s (%.0f-%.0f) [Active: %v]\n", w.Name, w.MinAmount, w.MaxAmount, w.IsActive)
	}

	// Test steps query for each workflow
	for _, workflow := range workflows {
		var steps []ApprovalStep
		if err := db.Where("workflow_id = ?", workflow.ID).Order("step_order").Find(&steps).Error; err != nil {
			log.Printf("Warning: Failed to query steps for workflow %d", workflow.ID)
			continue
		}

		fmt.Printf("\n🔄 Steps for '%s':\n", workflow.Name)
		for _, step := range steps {
			optional := ""
			if step.IsOptional {
				optional = " (Optional)"
			}
			fmt.Printf("   Step %d: %s - %s%s (%dh)\n", 
				step.StepOrder, step.StepName, step.ApproverRole, optional, step.TimeLimit)
		}
	}

	// Test workflow selection logic
	testAmounts := []float64{1000000, 10000000, 50000000, 200000000}
	
	fmt.Println("\n💰 Testing workflow selection by amount:")
	for _, amount := range testAmounts {
		var selectedWorkflow ApprovalWorkflow
		err := db.Where("module = ? AND min_amount <= ? AND (max_amount >= ? OR max_amount = 0) AND is_active = true", 
			"PURCHASE", amount, amount).
			Order("min_amount DESC").
			First(&selectedWorkflow).Error
		
		if err != nil {
			fmt.Printf("   Rp %.0f: ❌ No workflow found\n", amount)
		} else {
			fmt.Printf("   Rp %.0f: ✅ %s\n", amount, selectedWorkflow.Name)
		}
	}

	fmt.Println("\n🎯 Workflow is ready for employee purchase submission!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}