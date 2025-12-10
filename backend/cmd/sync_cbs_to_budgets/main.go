package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"os"

	"gorm.io/gorm"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	fmt.Println("🔄 Starting CBS to Project Budgets sync...")
	fmt.Println("============================================================")

	// Get all CBS nodes with COA mapping
	var cbsNodes []models.CBSNode
	err := db.Where("coa_account_id IS NOT NULL AND budget_amount > 0 AND deleted_at IS NULL").
		Find(&cbsNodes).Error
	if err != nil {
		log.Fatalf("Failed to fetch CBS nodes: %v", err)
	}

	fmt.Printf("Found %d CBS nodes with COA mapping\n\n", len(cbsNodes))

	synced := 0
	skipped := 0
	errors := 0

	for _, node := range cbsNodes {
		fmt.Printf("Processing: %s - %s (Budget: %d)\n", node.Code, node.Name, node.BudgetAmount)

		// Check if project budget already exists
		var existingBudget models.ProjectBudget
		err := db.Where("project_id = ? AND account_id = ? AND deleted_at IS NULL", 
			node.ProjectID, *node.COAAccountID).
			First(&existingBudget).Error

		if err == nil {
			// Budget exists, update it
			existingBudget.EstimatedAmount = float64(node.BudgetAmount)
			if err := db.Save(&existingBudget).Error; err != nil {
				fmt.Printf("  ❌ Error updating: %v\n", err)
				errors++
			} else {
				fmt.Printf("  ✅ Updated existing budget (ID: %d)\n", existingBudget.ID)
				synced++
			}
		} else if err == gorm.ErrRecordNotFound {
			// Budget doesn't exist, create it
			newBudget := models.ProjectBudget{
				ProjectID:       node.ProjectID,
				AccountID:       *node.COAAccountID,
				EstimatedAmount: float64(node.BudgetAmount),
			}
			if err := db.Create(&newBudget).Error; err != nil {
				fmt.Printf("  ❌ Error creating: %v\n", err)
				errors++
			} else {
				fmt.Printf("  ✅ Created new budget (ID: %d)\n", newBudget.ID)
				synced++
			}
		} else {
			fmt.Printf("  ⚠️  Database error: %v\n", err)
			errors++
		}
	}

	fmt.Println("\n============================================================")
	fmt.Printf("✅ Sync completed!\n")
	fmt.Printf("   Synced: %d\n", synced)
	fmt.Printf("   Skipped: %d\n", skipped)
	fmt.Printf("   Errors: %d\n", errors)

	if errors > 0 {
		os.Exit(1)
	}
}
