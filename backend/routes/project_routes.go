package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupProjectRoutes registers all project-related routes
func SetupProjectRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Initialize layers
	projectRepo := repositories.NewProjectRepository(db)
	projectService := services.NewProjectService(projectRepo)
	projectController := controllers.NewProjectController(projectService)
	
	// Project routes
	projects := router.Group("/projects")
	{
		// Get routes
		projects.GET("", projectController.GetAllProjects)           // GET /api/v1/projects
		projects.GET("/active", projectController.GetActiveProjects) // GET /api/v1/projects/active
		projects.GET("/status", projectController.GetProjectsByStatus) // GET /api/v1/projects/status?status=active
		projects.GET("/:id", projectController.GetProjectByID)       // GET /api/v1/projects/:id
		
		// Post routes
		projects.POST("", projectController.CreateProject)           // POST /api/v1/projects
		projects.POST("/:id/archive", projectController.ArchiveProject) // POST /api/v1/projects/:id/archive
		
		// Put/Patch routes
		projects.PUT("/:id", projectController.UpdateProject)        // PUT /api/v1/projects/:id
		projects.PATCH("/:id/progress", projectController.UpdateProgress) // PATCH /api/v1/projects/:id/progress
		
		// Delete routes
		projects.DELETE("/:id", projectController.DeleteProject)     // DELETE /api/v1/projects/:id
	}
}

