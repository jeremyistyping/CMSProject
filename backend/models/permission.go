package models

import (
	"time"

	"gorm.io/gorm"
)

// ModulePermissionRecord represents a specific permission for modules
type ModulePermissionRecord struct {
	ID         uint           `json:"id" gorm:"primaryKey;table:module_permissions"`
	UserID     uint           `json:"user_id" gorm:"not null;index"`
	Module     string         `json:"module" gorm:"not null;size:50;index"` // accounts, products, contacts, assets, sales, purchases, payments, cash_bank, settings
	CanView    bool           `json:"can_view" gorm:"default:false"`
	CanCreate  bool           `json:"can_create" gorm:"default:false"`
	CanEdit    bool           `json:"can_edit" gorm:"default:false"`
	CanDelete  bool           `json:"can_delete" gorm:"default:false"`
	CanApprove bool           `json:"can_approve" gorm:"default:false"`
	CanExport  bool           `json:"can_export" gorm:"default:false"`
	CanMenu    bool           `json:"can_menu" gorm:"default:false"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserPermission is a simplified structure for API responses
type UserPermission struct {
	UserID      uint                         `json:"user_id"`
	Username    string                       `json:"username"`
	Email       string                       `json:"email"`
	Role        string                       `json:"role"`
	Permissions map[string]*ModulePermission `json:"permissions"`
}

// ModulePermission represents permissions for a specific module
type ModulePermission struct {
	CanView    bool `json:"can_view"`
	CanCreate  bool `json:"can_create"`
	CanEdit    bool `json:"can_edit"`
	CanDelete  bool `json:"can_delete"`
	CanApprove bool `json:"can_approve"`
	CanExport  bool `json:"can_export"`
	CanMenu    bool `json:"can_menu"`
}

// GetDefaultPermissions returns default permissions based on role
// Modules: projects, cost_control, material_tracking, cbs, purchase_requests, daily_reports, settings
// Role Responsibilities:
// - admin: Full system access
// - managing_director: Final approval authority, view all
// - director: High-level oversight, approve purchase requests
// - project_director: Project management, approve daily reports & purchase requests
// - gm: Approve daily reports, view cost control
// - finance: Cost control & purchase request management
// - cost_control: CBS, material tracking, budget analysis
// - purchasing: Create & manage purchase requests
// - inventory_manager: Material tracking & inventory
// - employee: Daily reports input only
func GetDefaultPermissions(role string) map[string]*ModulePermission {
	permissions := make(map[string]*ModulePermission)
	// Cost Control focused modules
	modules := []string{"projects", "cost_control", "material_tracking", "cbs", "purchase_requests", "daily_reports", "settings"}

	switch role {
	case "admin":
		// Admin: Full access to all modules
		for _, module := range modules {
			permissions[module] = &ModulePermission{
				CanView:    true,
				CanCreate:  true,
				CanEdit:    true,
				CanDelete:  true,
				CanApprove: true,
				CanExport:  true,
				CanMenu:    true,
			}
		}

	case "managing_director":
		// Managing Director (Direktur Utama): Final approval authority, view all, approve all
		for _, module := range modules {
			if module == "settings" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    false,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: true,
					CanExport:  true,
					CanMenu:    true,
				}
			}
		}

	case "director":
		// Director: High-level oversight, approve purchase requests, view all cost control
		for _, module := range modules {
			if module == "purchase_requests" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: true,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "projects" || module == "cost_control" || module == "material_tracking" || module == "cbs" || module == "daily_reports" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    false,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    false,
				}
			}
		}

	case "project_director":
		// Project Director (Direktur Proyek): Manage projects, approve daily reports & purchase requests
		for _, module := range modules {
			if module == "projects" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  true,
					CanApprove: true,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "daily_reports" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: true,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "purchase_requests" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: true,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "cost_control" || module == "material_tracking" || module == "cbs" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    false,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    false,
				}
			}
		}

	case "gm":
		// GM (General Manager): Approve daily reports, view all cost control modules
		for _, module := range modules {
			if module == "daily_reports" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: true,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "projects" || module == "cost_control" || module == "material_tracking" || module == "cbs" || module == "purchase_requests" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    false,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    false,
				}
			}
		}

	case "finance", "finance_manager":
		// Finance: Manage cost control, approve purchase requests, view all reports
		for _, module := range modules {
			if module == "cost_control" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "purchase_requests" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: true,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "projects" || module == "material_tracking" || module == "cbs" || module == "daily_reports" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    false,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    false,
				}
			}
		}

	case "cost_control":
		// Cost Control: Full access to CBS, material tracking, budget analysis
		for _, module := range modules {
			if module == "cbs" || module == "material_tracking" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  true,
					CanApprove: true,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "cost_control" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "purchase_requests" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "projects" || module == "daily_reports" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    false,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    false,
				}
			}
		}

	case "purchasing":
		// Purchasing: Create & manage purchase requests
		for _, module := range modules {
			if module == "purchase_requests" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  true,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "projects" || module == "cost_control" || module == "material_tracking" || module == "cbs" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    true,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    false,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    false,
				}
			}
		}

	case "inventory_manager":
		// Inventory Manager: Material tracking & inventory management
		for _, module := range modules {
			if module == "material_tracking" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  true,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "purchase_requests" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true,
					CanMenu:    true,
				}
			} else if module == "projects" || module == "cost_control" || module == "cbs" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    true,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    false,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    false,
				}
			}
		}

	case "field_officer", "site_manager":
		// Field Team: Input daily reports, view projects, record material usage
		for _, module := range modules {
			if module == "daily_reports" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    true,
				}
			} else if module == "projects" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    true,
				}
			} else if module == "material_tracking" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    true,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    false,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    false,
				}
			}
		}

	case "employee":
		// Employee: View projects, create daily reports
		for _, module := range modules {
			if module == "daily_reports" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    true,
				}
			} else if module == "projects" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    true,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    false,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
					CanMenu:    false,
				}
			}
		}

	default:
		// Default: No permissions
		for _, module := range modules {
			permissions[module] = &ModulePermission{
				CanView:    false,
				CanCreate:  false,
				CanEdit:    false,
				CanDelete:  false,
				CanApprove: false,
				CanExport:  false,
				CanMenu:    false,
			}
		}
	}

	return permissions
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
