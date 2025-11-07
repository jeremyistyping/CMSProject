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

	fmt.Println("🗑️  Removing ALL purchase workflows for auto-migration testing...")
	fmt.Println("✅ Connected to database")

	// Find all PURCHASE workflows
	var workflows []ApprovalWorkflow
	result := db.Where("module = ?", "PURCHASE").Find(&workflows)
	if result.Error != nil {
		log.Fatal("❌ Error finding workflows:", result.Error)
	}

	fmt.Printf("📋 Found %d PURCHASE workflow(s)\n", len(workflows))
	for _, workflow := range workflows {
		fmt.Printf("   - %s (%.0f-%.0f) [Active: %t] [ID: %d]\n", 
			workflow.Name, workflow.MinAmount, workflow.MaxAmount, workflow.IsActive, workflow.ID)
	}

	if len(workflows) == 0 {
		fmt.Println("✅ No workflows to remove - ready for auto-migration test")
		return
	}

	// Get all workflow IDs for deletion
	var workflowIDs []uint
	for _, workflow := range workflows {
		workflowIDs = append(workflowIDs, workflow.ID)
	}

	// Update approval_requests to remove workflow references (set to NULL)
	fmt.Println("🔄 Removing workflow references from approval_requests...")
	result = db.Exec("UPDATE approval_requests SET workflow_id = NULL WHERE workflow_id IN (?)", workflowIDs)
	if result.Error != nil {
		// If column doesn't allow NULL, we might need to delete the records instead
		fmt.Println("⚠️  Cannot set to NULL, deleting approval_requests...")
		result = db.Exec("DELETE FROM approval_requests WHERE workflow_id IN (?)", workflowIDs)
		if result.Error != nil {
			log.Printf("⚠️  Warning: Error handling approval_requests: %v", result.Error)
		} else {
			fmt.Printf("   ✅ Deleted %d approval_request(s)\n", result.RowsAffected)
		}
	} else {
		fmt.Printf("   ✅ Updated %d approval_request(s)\n", result.RowsAffected)
	}

	// Delete workflow steps
	fmt.Println("🔄 Deleting workflow steps...")
	for _, workflowID := range workflowIDs {
		result := db.Where("workflow_id = ?", workflowID).Delete(&ApprovalStep{})
		if result.Error != nil {
			log.Printf("⚠️  Warning: Error deleting steps for workflow ID %d: %v", workflowID, result.Error)
		} else {
			fmt.Printf("   ✅ Deleted %d steps for workflow ID %d\n", result.RowsAffected, workflowID)
		}
	}

	// Delete workflows
	fmt.Println("🔄 Deleting workflows...")
	result = db.Where("id IN ?", workflowIDs).Delete(&ApprovalWorkflow{})
	if result.Error != nil {
		log.Fatal("❌ Error deleting workflows:", result.Error)
	}
	fmt.Printf("✅ Deleted %d workflow(s)\n", result.RowsAffected)

	// Verify deletion
	var count int64
	db.Where("module = ?", "PURCHASE").Model(&ApprovalWorkflow{}).Count(&count)
	fmt.Printf("📊 Remaining PURCHASE workflows: %d\n", count)

	fmt.Println("🎯 All PURCHASE workflows removed successfully!")
	fmt.Println("💡 Now restart the backend to test auto-migration")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}