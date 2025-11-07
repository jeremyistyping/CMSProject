package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	fmt.Println("📋 Verifying Receipt Creation Permissions for All Roles...")
	
	// Initialize database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Get all unique roles from users table
	var roles []string
	if err := db.Model(&models.User{}).Distinct("role").Pluck("role", &roles).Error; err != nil {
		log.Fatal("Failed to get roles:", err)
	}

	fmt.Printf("\n🔍 Checking permissions for %d role(s)...\n", len(roles))
	fmt.Println("==========================================")

	for _, role := range roles {
		if role == "" {
			continue
		}

		fmt.Printf("\n🏷️  ROLE: %s\n", role)
		
		// Get default permissions for this role
		defaultPerms := models.GetDefaultPermissions(role)
		
		// Check purchases module permissions
		if purchasePerm, exists := defaultPerms["purchases"]; exists {
			fmt.Printf("   📦 PURCHASES MODULE:\n")
			fmt.Printf("      👁️  Can View:    %v\n", purchasePerm.CanView)
			fmt.Printf("      ➕ Can Create:  %v\n", purchasePerm.CanCreate)
			fmt.Printf("      ✏️  Can Edit:    %v ← (Required for receipts)\n", purchasePerm.CanEdit)
			fmt.Printf("      🗑️  Can Delete:  %v\n", purchasePerm.CanDelete)
			fmt.Printf("      ✅ Can Approve: %v\n", purchasePerm.CanApprove)
			fmt.Printf("      📄 Can Export:  %v\n", purchasePerm.CanExport)
			
			// Determine if this role can create receipts
			canCreateReceipts := purchasePerm.CanEdit
			if canCreateReceipts {
				fmt.Printf("      🎯 RECEIPT CREATION: ✅ ALLOWED\n")
			} else {
				fmt.Printf("      🎯 RECEIPT CREATION: ❌ DENIED (needs CanEdit permission)\n")
			}
		} else {
			fmt.Printf("   📦 PURCHASES MODULE: ❌ NO PERMISSIONS FOUND\n")
			fmt.Printf("      🎯 RECEIPT CREATION: ❌ DENIED (no module access)\n")
		}
		
		// Count users with this role
		var userCount int64
		db.Model(&models.User{}).Where("role = ?", role).Count(&userCount)
		fmt.Printf("      👥 Users with this role: %d\n", userCount)
	}

	fmt.Println("\n==========================================")
	fmt.Println("✅ Permission verification completed!")
	fmt.Println("\n📌 Summary:")
	fmt.Println("   - To create receipts, users need 'CanEdit' permission on 'purchases' module")
	fmt.Println("   - The API endpoint POST /purchases/receipts requires permMiddleware.CanEdit(\"purchases\")")
	fmt.Println("   - Users without this permission will see '403 Forbidden' when trying to create receipts")
}