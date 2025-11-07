package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

// This script backfills missing role_permissions for the inventory_manager role
// so it aligns with the module-based permission system already used by the app.
// Safe to run multiple times (idempotent): it only inserts missing rows.
func main() {
	log.Println("Fixing inventory_manager role permissions…")

	db, err := database.Initialize()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Required permission names for inventory_manager
	required := []string{
		"products:read", "products:create", "products:update",
		"sales:read", "sales:create", "sales:update",
		"purchases:read", "purchases:create", "purchases:update",
		"contacts:read", "contacts:create", "contacts:update",
		"assets:read", "assets:create", "assets:update",
		"accounts:read",
		"reports:read",
	}

	// Build permission name -> ID map
	var perms []models.Permission
	if err := db.Find(&perms).Error; err != nil {
		log.Fatalf("failed to list permissions: %v", err)
	}
	nameToID := map[string]uint{}
	for _, p := range perms {
		nameToID[p.Name] = p.ID
	}

	role := models.RoleInventoryManager
	inserted := 0
	for _, name := range required {
		permID, ok := nameToID[name]
		if !ok {
			log.Printf("WARNING: permission %q not found; skipping", name)
			continue
		}

		var count int64
		if err := db.Model(&models.RolePermission{}).
			Where("role = ? AND permission_id = ?", role, permID).
			Count(&count).Error; err != nil {
			log.Fatalf("failed checking existing role_permission for %s: %v", name, err)
		}
		if count > 0 {
			continue
		}

		if err := db.Create(&models.RolePermission{Role: role, PermissionID: permID}).Error; err != nil {
			log.Fatalf("failed inserting role_permission %s: %v", name, err)
		}
		inserted++
		log.Printf("added %s to role %s", name, role)
	}

	fmt.Printf("Done. Inserted %d new permissions for role %q.\n", inserted, role)
}
