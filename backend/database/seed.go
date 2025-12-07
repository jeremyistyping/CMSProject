package database

import (
	"log"

	"app-sistem-akuntansi/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedData seeds initial data for the project management system
func SeedData(db *gorm.DB) {
	log.Println("Starting database seeding...")

	// Seed Users
	seedUsers(db)

	// Ensure module permissions table and defaults are prepared
	if err := MigratePermissions(db); err != nil {
		log.Printf("Error migrating/initializing module permissions: %v", err)
	}

	// Seed Permissions
	seedPermissions(db)

	// Seed Role Permissions
	seedRolePermissions(db)

	// Seed default approval workflows for PURCHASE_REQUEST module
	seedApprovalWorkflows(db)

	log.Println("Database seeding completed successfully")
}

func seedUsers(db *gorm.DB) {
	log.Println("🔄 Starting user seeding...")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	allUsers := []models.User{
		// Admin
		{
			Username:  "admin",
			Email:     "admin@company.com",
			Password:  string(hashedPassword),
			Role:      "admin",
			FirstName: "Admin",
			LastName:  "User",
			IsActive:  true,
		},
		// Finance
		{
			Username:  "finance",
			Email:     "finance@company.com",
			Password:  string(hashedPassword),
			Role:      "finance",
			FirstName: "Finance",
			LastName:  "User",
			IsActive:  true,
		},
		// Director
		{
			Username:  "director",
			Email:     "director@company.com",
			Password:  string(hashedPassword),
			Role:      "director",
			FirstName: "Director",
			LastName:  "User",
			IsActive:  true,
		},
		// Employee
		{
			Username:  "employee",
			Email:     "employee@company.com",
			Password:  string(hashedPassword),
			Role:      "employee",
			FirstName: "Employee",
			LastName:  "User",
			IsActive:  true,
		},
		// Purchasing (Andi)
		{
			Username:  "purchasing",
			Email:     "purchasing@company.com",
			Password:  string(hashedPassword),
			Role:      "purchasing",
			FirstName: "Andi",
			LastName:  "Purchasing",
			IsActive:  true,
		},
		// Cost Control (Patrick)
		{
			Username:  "cost_control",
			Email:     "patrick@company.com",
			Password:  string(hashedPassword),
			Role:      "cost_control",
			FirstName: "Patrick",
			LastName:  "Cost Control",
			IsActive:  true,
		},
		// GM (Marlin)
		{
			Username:  "gm",
			Email:     "gm@company.com",
			Password:  string(hashedPassword),
			Role:      "gm",
			FirstName: "Marlin",
			LastName:  "GM",
			IsActive:  true,
		},
		// Project Director (Christopher)
		{
			Username:  "project_director",
			Email:     "christopher@company.com",
			Password:  string(hashedPassword),
			Role:      "project_director",
			FirstName: "Christopher",
			LastName:  "Project Director",
			IsActive:  true,
		},
		// Managing Director (Jason)
		{
			Username:  "managing_director",
			Email:     "jason@company.com",
			Password:  string(hashedPassword),
			Role:      "managing_director",
			FirstName: "Jason",
			LastName:  "Managing Director",
			IsActive:  true,
		},
	}

	successCount := 0
	for _, user := range allUsers {
		// Check if user already exists
		var existingUser models.User
		if err := db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
			continue // User exists, skip
		}

		// Create user
		if err := db.Create(&user).Error; err != nil {
			log.Printf("⚠️  Warning: Failed to create user %s: %v", user.Username, err)
		} else {
			log.Printf("✅ User %s created successfully", user.Username)
			successCount++
		}
	}

	log.Printf("📊 User seeding completed: %d users created", successCount)
}

func seedPermissions(db *gorm.DB) {
	// Check if permissions already exist
	var count int64
	db.Model(&models.Permission{}).Count(&count)
	if count > 0 {
		return
	}

	permissions := []models.Permission{
		// User permissions
		{Name: "users:read", Resource: "users", Action: "read", Description: "View users"},
		{Name: "users:create", Resource: "users", Action: "create", Description: "Create users"},
		{Name: "users:update", Resource: "users", Action: "update", Description: "Update users"},
		{Name: "users:delete", Resource: "users", Action: "delete", Description: "Delete users"},
		{Name: "users:manage", Resource: "users", Action: "manage", Description: "Full user management"},

		// Project permissions
		{Name: "projects:read", Resource: "projects", Action: "read", Description: "View projects"},
		{Name: "projects:create", Resource: "projects", Action: "create", Description: "Create projects"},
		{Name: "projects:update", Resource: "projects", Action: "update", Description: "Update projects"},
		{Name: "projects:delete", Resource: "projects", Action: "delete", Description: "Delete projects"},

		// Purchase Request permissions
		{Name: "purchases:read", Resource: "purchases", Action: "read", Description: "View purchase requests"},
		{Name: "purchases:create", Resource: "purchases", Action: "create", Description: "Create purchase requests"},
		{Name: "purchases:update", Resource: "purchases", Action: "update", Description: "Update purchase requests"},
		{Name: "purchases:delete", Resource: "purchases", Action: "delete", Description: "Delete purchase requests"},
		{Name: "purchases:approve", Resource: "purchases", Action: "approve", Description: "Approve purchase requests"},

		// Report permissions
		{Name: "reports:read", Resource: "reports", Action: "read", Description: "View reports"},
		{Name: "reports:create", Resource: "reports", Action: "create", Description: "Create reports"},

		// Budget permissions
		{Name: "budgets:read", Resource: "budgets", Action: "read", Description: "View budgets"},
		{Name: "budgets:create", Resource: "budgets", Action: "create", Description: "Create budgets"},
		{Name: "budgets:update", Resource: "budgets", Action: "update", Description: "Update budgets"},
		{Name: "budgets:delete", Resource: "budgets", Action: "delete", Description: "Delete budgets"},
	}

	for _, permission := range permissions {
		db.Create(&permission)
	}

	log.Printf("✅ Seeded %d permissions", len(permissions))
}

func seedRolePermissions(db *gorm.DB) {
	// Check if role permissions already exist
	var count int64
	db.Model(&models.RolePermission{}).Count(&count)
	if count > 0 {
		return
	}

	// Get all permissions
	var permissions []models.Permission
	db.Find(&permissions)

	permissionMap := make(map[string]uint)
	for _, perm := range permissions {
		permissionMap[perm.Name] = perm.ID
	}

	// Define role permissions
	rolePermissions := map[string][]string{
		"admin": {
			"users:read", "users:create", "users:update", "users:delete", "users:manage",
			"projects:read", "projects:create", "projects:update", "projects:delete",
			"purchases:read", "purchases:create", "purchases:update", "purchases:delete", "purchases:approve",
			"reports:read", "reports:create",
			"budgets:read", "budgets:create", "budgets:update", "budgets:delete",
		},
		"finance": {
			"projects:read",
			"purchases:read", "purchases:approve",
			"reports:read", "reports:create",
			"budgets:read", "budgets:create", "budgets:update",
		},
		"director": {
			"users:read",
			"projects:read",
			"purchases:read", "purchases:approve",
			"reports:read", "reports:create",
			"budgets:read", "budgets:create", "budgets:update",
		},
		"employee": {
			"projects:read",
			"purchases:read", "purchases:create", "purchases:update",
			"reports:read",
		},
	}

	// Create role permissions
	for role, perms := range rolePermissions {
		for _, permName := range perms {
			if permID, exists := permissionMap[permName]; exists {
				rolePermission := models.RolePermission{
					Role:         role,
					PermissionID: permID,
				}
				db.Create(&rolePermission)
			}
		}
	}

	log.Println("✅ Role permissions seeded")
}

// seedApprovalWorkflows creates default PURCHASE_REQUEST workflows
func seedApprovalWorkflows(db *gorm.DB) {
	var count int64
	db.Model(&models.ApprovalWorkflow{}).Where("module = ?", models.ApprovalModulePurchaseRequest).Count(&count)
	if count > 0 {
		return
	}

	// Standard Purchase Approval workflow
	workflow := models.ApprovalWorkflow{
		Name:            "Standard Purchase Approval",
		Module:          models.ApprovalModulePurchaseRequest,
		MinAmount:       0,
		MaxAmount:       0, // No upper limit
		IsActive:        true,
		RequireFinance:  true,
		RequireDirector: true,
	}
	db.Create(&workflow)

	// Create approval steps
	steps := []models.ApprovalStep{
		{WorkflowID: workflow.ID, StepOrder: 1, StepName: "Purchasing Review", ApproverRole: "purchasing"},
		{WorkflowID: workflow.ID, StepOrder: 2, StepName: "Cost Control Review", ApproverRole: "cost_control"},
		{WorkflowID: workflow.ID, StepOrder: 3, StepName: "GM Approval", ApproverRole: "gm"},
		{WorkflowID: workflow.ID, StepOrder: 4, StepName: "Project Director Approval", ApproverRole: "project_director"},
		{WorkflowID: workflow.ID, StepOrder: 5, StepName: "Managing Director Approval", ApproverRole: "managing_director"},
	}

	for _, step := range steps {
		db.Create(&step)
	}

	log.Println("✅ Created Standard Purchase Approval workflow with 5 steps")
}
