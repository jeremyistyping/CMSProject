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

	fmt.Println("🧹 Cleaning up old renamed workflow...")
	fmt.Println("✅ Connected to database")

	// Find old renamed workflow
	var oldWorkflow ApprovalWorkflow
	result := db.Where("name = ?", "OLD - Standard Purchase Approval (Renamed for Test)").First(&oldWorkflow)
	
	if result.Error != nil {
		fmt.Println("❌ Old workflow not found - nothing to clean up")
		return
	}

	fmt.Printf("📋 Found old workflow (ID: %d)\n", oldWorkflow.ID)

	// Just deactivate it instead of deleting (safer)
	result = db.Model(&oldWorkflow).Update("is_active", false)
	if result.Error != nil {
		log.Fatal("❌ Error deactivating old workflow:", result.Error)
	}

	fmt.Printf("✅ Deactivated old workflow\n")

	// Verify final state
	var activeWorkflows []ApprovalWorkflow
	db.Where("module = ? AND is_active = true", "PURCHASE").Find(&activeWorkflows)
	
	fmt.Printf("📊 Active PURCHASE workflows: %d\n", len(activeWorkflows))
	for _, workflow := range activeWorkflows {
		fmt.Printf("   - %s (%.0f-%.0f)\n", workflow.Name, workflow.MinAmount, workflow.MaxAmount)
	}

	fmt.Println("🎯 Cleanup completed!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}