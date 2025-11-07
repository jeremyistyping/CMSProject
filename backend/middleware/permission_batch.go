package middleware

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// BatchCheckPermissions checks permissions for multiple modules at once
// Returns map[module]map[action]bool
func (pm *PermissionMiddleware) BatchCheckPermissions(userID uint, modules []string) (map[string]map[string]bool, error) {
	result := make(map[string]map[string]bool)
	
	// Initialize result map
	for _, module := range modules {
		result[module] = map[string]bool{
			"view":    false,
			"create":  false,
			"edit":    false,
			"delete":  false,
			"approve": false,
			"export":  false,
		}
	}
	
	// Batch query - single DB call for all modules
	var permissions []models.ModulePermissionRecord
	err := pm.db.Where("user_id = ? AND module IN ?", userID, modules).Find(&permissions).Error
	
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	
	// Process found permissions
	for _, perm := range permissions {
		if _, exists := result[perm.Module]; exists {
			result[perm.Module] = map[string]bool{
				"view":    perm.CanView,
				"create":  perm.CanCreate,
				"edit":    perm.CanEdit,
				"delete":  perm.CanDelete,
				"approve": perm.CanApprove,
				"export":  perm.CanExport,
			}
		}
	}
	
	// Store all results in cache
	for module, perms := range result {
		for action, hasPermission := range perms {
			pm.cache.Set(userID, module, action, hasPermission)
		}
	}
	
	return result, nil
}

// PreloadPermissions can be called at login/token refresh to preload all permissions
func (pm *PermissionMiddleware) PreloadPermissions(userID uint) error {
	// Get all modules for this user
	var permissions []models.ModulePermissionRecord
	err := pm.db.Where("user_id = ?", userID).Find(&permissions).Error
	
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	
	// Cache all permissions
	for _, perm := range permissions {
		pm.cache.Set(userID, perm.Module, "view", perm.CanView)
		pm.cache.Set(userID, perm.Module, "create", perm.CanCreate)
		pm.cache.Set(userID, perm.Module, "edit", perm.CanEdit)
		pm.cache.Set(userID, perm.Module, "delete", perm.CanDelete)
		pm.cache.Set(userID, perm.Module, "approve", perm.CanApprove)
		pm.cache.Set(userID, perm.Module, "export", perm.CanExport)
	}
	
	return nil
}
