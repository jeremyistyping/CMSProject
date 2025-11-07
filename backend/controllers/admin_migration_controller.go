package controllers

import (
	"net/http"
	"app-sistem-akuntansi/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminMigrationController struct {
	db *gorm.DB
}

func NewAdminMigrationController(db *gorm.DB) *AdminMigrationController {
	return &AdminMigrationController{db: db}
}

// ForceRunDateConstraintFix godoc
// @Summary Force run date constraint fix migration
// @Description Manually trigger the date constraint fix for period closing
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/migrations/fix-date-constraint [post]
func (amc *AdminMigrationController) ForceRunDateConstraintFix(c *gin.Context) {
	// Run the fix
	database.ForceRunDateConstraintFix(amc.db)
	
	// Verify constraint
	var constraintDef string
	err := amc.db.Raw(`
		SELECT pg_get_constraintdef(oid) as constraint_definition
		FROM pg_constraint 
		WHERE conname = 'chk_journal_entries_date_valid'
	`).Scan(&constraintDef).Error
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to verify constraint",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Date constraint fix completed successfully",
		"constraint_definition": constraintDef,
		"info": "Constraint now allows dates from 2000-01-01 to 2099-12-31",
	})
}

// GetConstraintInfo godoc
// @Summary Get current date constraint info
// @Description Check the current date validation constraint on journal_entries
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/migrations/constraint-info [get]
func (amc *AdminMigrationController) GetConstraintInfo(c *gin.Context) {
	var constraintDef string
	err := amc.db.Raw(`
		SELECT pg_get_constraintdef(oid) as constraint_definition
		FROM pg_constraint 
		WHERE conname = 'chk_journal_entries_date_valid'
	`).Scan(&constraintDef).Error
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get constraint info",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"constraint_name": "chk_journal_entries_date_valid",
		"constraint_definition": constraintDef,
	})
}
