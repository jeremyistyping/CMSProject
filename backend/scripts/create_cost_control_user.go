package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"log"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	log.Println("🔄 Creating Cost Control user...")

	// Check if cost_control user already exists
	var existingUser models.User
	if err := db.Where("username = ?", "cost_control").First(&existingUser).Error; err == nil {
		log.Println("⏭️  User cost_control already exists, skipping creation")
		log.Printf("   User ID: %d, Email: %s", existingUser.ID, existingUser.Email)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create cost_control user
	costControlUser := models.User{
		Username:  "cost_control",
		Email:     "patrick@company.com",
		Password:  string(hashedPassword),
		Role:      "cost_control",
		FirstName: "Patrick",
		LastName:  "Cost Control",
		IsActive:  true,
	}

	if err := db.Create(&costControlUser).Error; err != nil {
		log.Fatalf("❌ Failed to create cost_control user: %v", err)
	}

	log.Printf("✅ Cost Control user created successfully! (ID: %d)", costControlUser.ID)
	log.Println("   Username: cost_control")
	log.Println("   Email: patrick@company.com")
	log.Println("   Password: password123")
	log.Println("   Role: cost_control")

	// Create permissions for cost_control user
	log.Println("🔄 Setting up permissions for cost_control user...")
	
	defaultPerms := models.GetDefaultPermissions("cost_control")
	modules := []string{"accounts", "products", "contacts", "assets", "sales", "purchases", "payments", "cash_bank", "cost_control", "reports", "settings"}
	
	successCount := 0
	for _, module := range modules {
		perm := defaultPerms[module]
		if perm != nil {
			permission := models.ModulePermissionRecord{
				UserID:     costControlUser.ID,
				Module:     module,
				CanView:    perm.CanView,
				CanCreate:  perm.CanCreate,
				CanEdit:    perm.CanEdit,
				CanDelete:  perm.CanDelete,
				CanApprove: perm.CanApprove,
				CanExport:  perm.CanExport,
				CanMenu:    perm.CanMenu,
			}

			if err := db.Create(&permission).Error; err != nil {
				log.Printf("⚠️  Warning: Failed to create permission for module %s: %v", module, err)
			} else {
				if perm.CanMenu {
					log.Printf("   ✓ %s: CanApprove=%v, CanMenu=%v", module, perm.CanApprove, perm.CanMenu)
				}
				successCount++
			}
		}
	}

	log.Printf("📊 Permissions setup completed: %d/%d modules configured", successCount, len(modules))
	log.Println("✅ Cost Control user and permissions created successfully!")
	log.Println("")
	log.Println("You can now login with:")
	log.Println("   Email: patrick@company.com")
	log.Println("   Password: password123")
}
