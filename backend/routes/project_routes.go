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
	projectService := services.NewProjectService(projectRepo, db)
	projectController := controllers.NewProjectController(projectService)

	// Initialize Project Budget service and controller
	projectBudgetService := services.NewProjectBudgetService(db)
	projectBudgetController := controllers.NewProjectBudgetController(projectBudgetService)
	
	// Initialize Project Progress service and controller
	projectProgressService := services.NewProjectProgressService(db)
	projectProgressController := controllers.NewProjectProgressController(projectProgressService)
	
	// Initialize Project Actual Cost service and controller
	projectActualCostService := services.NewProjectActualCostService(db)
	projectActualCostController := controllers.NewProjectActualCostController(projectActualCostService)
	
	// Initialize Daily Update service and controller
	dailyUpdateService := services.NewDailyUpdateService(db)
	dailyUpdateController := controllers.NewDailyUpdateController(dailyUpdateService)
	
	// Initialize Milestone service and controller
	milestoneService := services.NewMilestoneService(db)
	milestoneController := controllers.NewMilestoneController(milestoneService)
	
	// Initialize Weekly Report service and controller
	weeklyReportService := services.NewWeeklyReportService(db)
	weeklyReportController := controllers.NewWeeklyReportController(weeklyReportService)
	
	// Initialize Timeline Schedule service and controller
	timelineScheduleService := services.NewTimelineScheduleService(db)
	timelineScheduleController := controllers.NewTimelineScheduleController(timelineScheduleService)
	
	// Project routes
	projects := router.Group("/projects")
	{
		// Get routes
		projects.GET("", projectController.GetAllProjects)           // GET /api/v1/projects
		projects.GET("/active", projectController.GetActiveProjects) // GET /api/v1/projects/active
		projects.GET("/status", projectController.GetProjectsByStatus) // GET /api/v1/projects/status?status=active
		projects.GET("/:id", projectController.GetProjectByID)       // GET /api/v1/projects/:id
		projects.GET("/:id/cost-summary", projectController.GetProjectCostSummary) // GET /api/v1/projects/:id/cost-summary
		projects.GET("/:id/progress-history", projectProgressController.GetProjectProgressHistory) // GET /api/v1/projects/:id/progress-history
		projects.GET("/:id/actual-costs", projectActualCostController.GetProjectActualCosts)       // GET /api/v1/projects/:id/actual-costs
		
		// Project budgets (nested under projects)
		projects.GET("/:id/budgets", projectBudgetController.GetProjectBudgets)            // GET /api/v1/projects/:id/budgets
		projects.POST("/:id/budgets", projectBudgetController.UpsertProjectBudgets)        // POST /api/v1/projects/:id/budgets
		projects.DELETE("/:id/budgets/:budgetId", projectBudgetController.DeleteProjectBudget) // DELETE /api/v1/projects/:id/budgets/:budgetId
		
		// Post routes
		projects.POST("", projectController.CreateProject)           // POST /api/v1/projects
		projects.POST("/:id/archive", projectController.ArchiveProject) // POST /api/v1/projects/:id/archive
		projects.POST("/:id/progress-history", projectProgressController.UpsertProjectProgress) // POST /api/v1/projects/:id/progress-history
		
		// Put/Patch routes
		projects.PUT("/:id", projectController.UpdateProject)        // PUT /api/v1/projects/:id
		projects.PATCH("/:id/progress", projectController.UpdateProgress) // PATCH /api/v1/projects/:id/progress
		
		// Delete routes
		projects.DELETE("/:id", projectController.DeleteProject)     // DELETE /api/v1/projects/:id
		
		// Daily Updates routes (nested under projects)
		projects.GET("/:id/daily-updates", dailyUpdateController.GetDailyUpdates)       // GET /api/v1/projects/:id/daily-updates
		projects.GET("/:id/daily-updates/:updateId", dailyUpdateController.GetDailyUpdate)   // GET /api/v1/projects/:id/daily-updates/:updateId
		projects.POST("/:id/daily-updates", dailyUpdateController.CreateDailyUpdate)   // POST /api/v1/projects/:id/daily-updates
		projects.PUT("/:id/daily-updates/:updateId", dailyUpdateController.UpdateDailyUpdate) // PUT /api/v1/projects/:id/daily-updates/:updateId
		projects.DELETE("/:id/daily-updates/:updateId", dailyUpdateController.DeleteDailyUpdate) // DELETE /api/v1/projects/:id/daily-updates/:updateId
		
		// Milestones routes (nested under projects)
		projects.GET("/:id/milestones", milestoneController.GetMilestones)                    // GET /api/v1/projects/:id/milestones
		projects.GET("/:id/milestones/:milestoneId", milestoneController.GetMilestone)        // GET /api/v1/projects/:id/milestones/:milestoneId
		projects.POST("/:id/milestones", milestoneController.CreateMilestone)                 // POST /api/v1/projects/:id/milestones
		projects.PUT("/:id/milestones/:milestoneId", milestoneController.UpdateMilestone)     // PUT /api/v1/projects/:id/milestones/:milestoneId
		projects.DELETE("/:id/milestones/:milestoneId", milestoneController.DeleteMilestone)  // DELETE /api/v1/projects/:id/milestones/:milestoneId
		projects.POST("/:id/milestones/:milestoneId/complete", milestoneController.CompleteMilestone) // POST /api/v1/projects/:id/milestones/:milestoneId/complete
		
		// Weekly Reports routes (nested under projects)
		projects.GET("/:id/weekly-reports", weeklyReportController.GetWeeklyReports)                    // GET /api/v1/projects/:id/weekly-reports
		projects.GET("/:id/weekly-reports/export-all", weeklyReportController.ExportAllPDF)             // GET /api/v1/projects/:id/weekly-reports/export-all
		projects.GET("/:id/weekly-reports/:reportId", weeklyReportController.GetWeeklyReport)            // GET /api/v1/projects/:id/weekly-reports/:reportId
		projects.GET("/:id/weekly-reports/:reportId/pdf", weeklyReportController.GeneratePDF)            // GET /api/v1/projects/:id/weekly-reports/:reportId/pdf
		projects.POST("/:id/weekly-reports", weeklyReportController.CreateWeeklyReport)                  // POST /api/v1/projects/:id/weekly-reports
		projects.PUT("/:id/weekly-reports/:reportId", weeklyReportController.UpdateWeeklyReport)         // PUT /api/v1/projects/:id/weekly-reports/:reportId
		projects.DELETE("/:id/weekly-reports/:reportId", weeklyReportController.DeleteWeeklyReport)      // DELETE /api/v1/projects/:id/weekly-reports/:reportId
		
		// Timeline Schedule routes (nested under projects)
		projects.GET("/:id/timeline-schedules", timelineScheduleController.GetSchedules)                       // GET /api/v1/projects/:id/timeline-schedules
		projects.GET("/:id/timeline-schedules/:scheduleId", timelineScheduleController.GetSchedule)            // GET /api/v1/projects/:id/timeline-schedules/:scheduleId
		projects.POST("/:id/timeline-schedules", timelineScheduleController.CreateSchedule)                    // POST /api/v1/projects/:id/timeline-schedules
		projects.PUT("/:id/timeline-schedules/:scheduleId", timelineScheduleController.UpdateSchedule)         // PUT /api/v1/projects/:id/timeline-schedules/:scheduleId
		projects.DELETE("/:id/timeline-schedules/:scheduleId", timelineScheduleController.DeleteSchedule)      // DELETE /api/v1/projects/:id/timeline-schedules/:scheduleId
		projects.PATCH("/:id/timeline-schedules/:scheduleId/status", timelineScheduleController.UpdateScheduleStatus) // PATCH /api/v1/projects/:id/timeline-schedules/:scheduleId/status
	}
}

