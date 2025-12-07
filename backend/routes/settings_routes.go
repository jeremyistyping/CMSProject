package routes

import (
	"github.com/gin-gonic/gin"
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/services"
	"gorm.io/gorm"
)

// SetupSettingsRoutes registers all settings-related routes
func SetupSettingsRoutes(protected *gin.RouterGroup, db *gorm.DB) {
	// Initialize settings service and controller
	settingsService := services.NewSettingsService(db)
	settingsController := controllers.NewSettingsController(settingsService)
	
	// Settings routes (admin only)
	settings := protected.Group("/settings")
	settings.Use(middleware.RoleRequired("admin"))
	{
		// Main settings endpoints
		settings.GET("", settingsController.GetSettings)
		settings.PUT("", settingsController.UpdateSettings)
		
		// Specific settings endpoints
		settings.PUT("/company", settingsController.UpdateCompanyInfo)
		settings.PUT("/system", settingsController.UpdateSystemConfig)
		settings.POST("/company/logo", settingsController.UploadCompanyLogo)
		
		// Additional endpoints
		settings.POST("/reset", settingsController.ResetToDefaults)
		settings.GET("/validation-rules", settingsController.GetValidationRules)
		settings.GET("/history", settingsController.GetSettingsHistory)
	}
}
