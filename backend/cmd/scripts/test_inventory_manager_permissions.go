package main

import (
	"fmt"
	"log"
	"os"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := loadEnv(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Initialize database connection
	db, err := database.Initialize()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🧪 Testing Inventory Manager Permissions")
	fmt.Println("=========================================")

	// Test default permissions for inventory_manager role
	testDefaultPermissions()

	// Test database permissions for inventory_manager users
	testDatabasePermissions(db)

	fmt.Println("\n✅ Permission testing completed!")
}

func testDefaultPermissions() {
	fmt.Println("\n📋 Testing Default Permissions for inventory_manager role...")
	
	permissions := models.GetDefaultPermissions("inventory_manager")
	
	expectedModules := map[string]map[string]bool{
		// Core inventory modules - should have full access
		"products": {
			"CanView": true, "CanCreate": true, "CanEdit": true, 
			"CanDelete": false, "CanApprove": false, "CanExport": true,
		},
		"purchases": {
			"CanView": true, "CanCreate": true, "CanEdit": true,
			"CanDelete": false, "CanApprove": false, "CanExport": true,
		},
		"sales": {
			"CanView": true, "CanCreate": true, "CanEdit": true,
			"CanDelete": false, "CanApprove": false, "CanExport": true,
		},
		// Supporting modules - should have good access
		"contacts": {
			"CanView": true, "CanCreate": true, "CanEdit": true,
			"CanDelete": false, "CanApprove": false, "CanExport": true,
		},
		"assets": {
			"CanView": true, "CanCreate": true, "CanEdit": true,
			"CanDelete": false, "CanApprove": false, "CanExport": true,
		},
		"reports": {
			"CanView": true, "CanCreate": true, "CanEdit": true,
			"CanDelete": false, "CanApprove": false, "CanExport": true,
		},
		// Financial support modules - limited but functional
		"accounts": {
			"CanView": true, "CanCreate": true, "CanEdit": false,
			"CanDelete": false, "CanApprove": false, "CanExport": true,
		},
		"payments": {
			"CanView": true, "CanCreate": true, "CanEdit": false,
			"CanDelete": false, "CanApprove": false, "CanExport": true,
		},
		"cash_bank": {
			"CanView": true, "CanCreate": true, "CanEdit": false,
			"CanDelete": false, "CanApprove": false, "CanExport": true,
		},
	}

	allTestsPassed := true

	for module, expectedPerms := range expectedModules {
		if perm, exists := permissions[module]; exists {
			fmt.Printf("  📦 Module: %s\n", module)
			
			// Test each permission
			if perm.CanView != expectedPerms["CanView"] {
				fmt.Printf("    ❌ CanView: expected %v, got %v\n", expectedPerms["CanView"], perm.CanView)
				allTestsPassed = false
			} else {
				fmt.Printf("    ✅ CanView: %v\n", perm.CanView)
			}

			if perm.CanCreate != expectedPerms["CanCreate"] {
				fmt.Printf("    ❌ CanCreate: expected %v, got %v\n", expectedPerms["CanCreate"], perm.CanCreate)
				allTestsPassed = false
			} else {
				fmt.Printf("    ✅ CanCreate: %v\n", perm.CanCreate)
			}

			if perm.CanEdit != expectedPerms["CanEdit"] {
				fmt.Printf("    ❌ CanEdit: expected %v, got %v\n", expectedPerms["CanEdit"], perm.CanEdit)
				allTestsPassed = false
			} else {
				fmt.Printf("    ✅ CanEdit: %v\n", perm.CanEdit)
			}

			if perm.CanDelete != expectedPerms["CanDelete"] {
				fmt.Printf("    ❌ CanDelete: expected %v, got %v\n", expectedPerms["CanDelete"], perm.CanDelete)
				allTestsPassed = false
			} else {
				fmt.Printf("    ✅ CanDelete: %v\n", perm.CanDelete)
			}

			if perm.CanApprove != expectedPerms["CanApprove"] {
				fmt.Printf("    ❌ CanApprove: expected %v, got %v\n", expectedPerms["CanApprove"], perm.CanApprove)
				allTestsPassed = false
			} else {
				fmt.Printf("    ✅ CanApprove: %v\n", perm.CanApprove)
			}

			if perm.CanExport != expectedPerms["CanExport"] {
				fmt.Printf("    ❌ CanExport: expected %v, got %v\n", expectedPerms["CanExport"], perm.CanExport)
				allTestsPassed = false
			} else {
				fmt.Printf("    ✅ CanExport: %v\n", perm.CanExport)
			}

			fmt.Println()
		} else {
			fmt.Printf("  ❌ Module %s not found in permissions!\n", module)
			allTestsPassed = false
		}
	}

	if allTestsPassed {
		fmt.Println("🎉 All default permission tests PASSED!")
	} else {
		fmt.Println("❌ Some default permission tests FAILED!")
	}
}

func testDatabasePermissions(db *gorm.DB) {
	fmt.Println("\n🗃️  Testing Database Permissions for inventory_manager users...")

	// Find inventory_manager users
	var users []models.User
	if err := db.Where("role = ?", "inventory_manager").Find(&users).Error; err != nil {
		log.Printf("Error finding inventory_manager users: %v", err)
		return
	}

	if len(users) == 0 {
		fmt.Println("  ⚠️  No inventory_manager users found in database")
		return
	}

	for _, user := range users {
		fmt.Printf("  👤 Testing user: %s (ID: %d)\n", user.Username, user.ID)

		// Get user's database permissions
		var permissions []models.ModulePermissionRecord
		if err := db.Where("user_id = ?", user.ID).Find(&permissions).Error; err != nil {
			log.Printf("    Error fetching permissions: %v", err)
			continue
		}

		if len(permissions) == 0 {
			fmt.Printf("    📋 No custom permissions found - will use defaults\n")
		} else {
			fmt.Printf("    📋 Found %d custom permission records:\n", len(permissions))
			for _, perm := range permissions {
				fmt.Printf("      - %s: View=%v Create=%v Edit=%v Delete=%v Approve=%v Export=%v\n",
					perm.Module, perm.CanView, perm.CanCreate, perm.CanEdit, 
					perm.CanDelete, perm.CanApprove, perm.CanExport)
			}
		}

		// Test critical modules access
		criticalModules := []string{"products", "purchases", "sales", "contacts", "accounts"}
		fmt.Printf("    🧪 Testing critical modules access:\n")
		
		for _, module := range criticalModules {
			hasAccess := hasModuleAccess(db, user.ID, user.Role, module, "view")
			if hasAccess {
				fmt.Printf("      ✅ %s: Can access\n", module)
			} else {
				fmt.Printf("      ❌ %s: Access denied\n", module)
			}
		}
		fmt.Println()
	}
}

// Helper function to check if user has access to a module/action
func hasModuleAccess(db *gorm.DB, userID uint, userRole, module, action string) bool {
	// First check if user has custom permission
	var permission models.ModulePermissionRecord
	err := db.Where("user_id = ? AND module = ?", userID, module).First(&permission).Error
	
	if err == nil {
		// Custom permission found
		switch action {
		case "view":
			return permission.CanView
		case "create":
			return permission.CanCreate
		case "edit":
			return permission.CanEdit
		case "delete":
			return permission.CanDelete
		case "approve":
			return permission.CanApprove
		case "export":
			return permission.CanExport
		}
	} else if err == gorm.ErrRecordNotFound {
		// No custom permission, use default based on role
		defaultPerms := models.GetDefaultPermissions(userRole)
		if modPerm, ok := defaultPerms[module]; ok {
			switch action {
			case "view":
				return modPerm.CanView
			case "create":
				return modPerm.CanCreate
			case "edit":
				return modPerm.CanEdit
			case "delete":
				return modPerm.CanDelete
			case "approve":
				return modPerm.CanApprove
			case "export":
				return modPerm.CanExport
			}
		}
	}
	
	return false
}

func loadEnv() error {
	// Simple .env loader - you can use godotenv package for production
	return nil
}