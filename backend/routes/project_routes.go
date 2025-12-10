package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupProjectRoutes registers all project-related routes
func SetupProjectRoutes(router *gin.RouterGroup, db *gorm.DB, permMiddleware *middleware.PermissionMiddleware) {
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

	// Initialize Expense Transaction service and controller
	expenseRepo := repositories.NewExpenseTransactionRepository(db)
	coaRepo := repositories.NewCOARepository(db)
	expenseService := services.NewExpenseTransactionService(expenseRepo, projectRepo, coaRepo)
	expenseController := controllers.NewExpenseTransactionController(expenseService)

	// Initialize CBS service and controller
	cbsRepo := repositories.NewCBSRepository(db)
	projectBudgetRepo := repositories.NewProjectBudgetRepository(db)
	cbsService := services.NewCBSService(cbsRepo, projectBudgetRepo)
	cbsController := controllers.NewCBSController(cbsService)

	// Project routes
	projects := router.Group("/projects")
	{
		// Get routes - View Project Permission
		// Note: Static routes must come before dynamic routes with :id
		projects.GET("", permMiddleware.CanView("projects"), projectController.GetAllProjects)
		projects.GET("/active", permMiddleware.CanView("projects"), projectController.GetActiveProjects)
		projects.GET("/status", permMiddleware.CanView("projects"), projectController.GetProjectsByStatus)
		projects.GET("/daily-updates/pending", permMiddleware.CanApprove("daily_updates"), dailyUpdateController.GetPendingDailyUpdates)

		// Project budgets (nested under projects) - Edit Project Permission (Budgets are sensitive)
		projects.GET("/:id/budgets", permMiddleware.CanView("projects"), projectBudgetController.GetProjectBudgets)
		projects.POST("/:id/budgets", permMiddleware.CanEdit("projects"), projectBudgetController.UpsertProjectBudgets)
		projects.DELETE("/:id/budgets/:budgetId", permMiddleware.CanEdit("projects"), projectBudgetController.DeleteProjectBudget)

		// Post routes - Create Project Permission
		projects.POST("", permMiddleware.CanCreate("projects"), projectController.CreateProject)
		projects.POST("/:id/archive", permMiddleware.CanDelete("projects"), projectController.ArchiveProject) // Archive treated as delete/admin action
		projects.POST("/:id/progress-history", permMiddleware.CanEdit("projects"), projectProgressController.UpsertProjectProgress)

		// Put/Patch routes - Edit Project Permission
		projects.PUT("/:id", permMiddleware.CanEdit("projects"), projectController.UpdateProject)
		projects.PATCH("/:id/progress", permMiddleware.CanEdit("projects"), projectController.UpdateProgress)

		// Delete routes - Delete Project Permission
		projects.DELETE("/:id", permMiddleware.CanDelete("projects"), projectController.DeleteProject)

		// Daily Updates routes (nested under projects) - Daily Updates Permission
		// Note: Field officers need Create/Edit on daily_updates but only View on projects
		projects.GET("/:id/daily-updates", permMiddleware.CanView("daily_updates"), dailyUpdateController.GetDailyUpdates)
		projects.GET("/:id/daily-updates/:updateId", permMiddleware.CanView("daily_updates"), dailyUpdateController.GetDailyUpdate)
		projects.POST("/:id/daily-updates", permMiddleware.CanCreate("daily_updates"), dailyUpdateController.CreateDailyUpdate)
		projects.PUT("/:id/daily-updates/:updateId", permMiddleware.CanEdit("daily_updates"), dailyUpdateController.UpdateDailyUpdate)
		projects.DELETE("/:id/daily-updates/:updateId", permMiddleware.CanDelete("daily_updates"), dailyUpdateController.DeleteDailyUpdate)
		projects.POST("/:id/daily-updates/:updateId/approve", permMiddleware.CanApprove("daily_updates"), dailyUpdateController.ApproveDailyUpdate)
		projects.POST("/:id/daily-updates/:updateId/reject", permMiddleware.CanApprove("daily_updates"), dailyUpdateController.RejectDailyUpdate)

		// Milestones routes (nested under projects) - Project Edit Permission
		projects.GET("/:id/milestones", permMiddleware.CanView("projects"), milestoneController.GetMilestones)
		projects.GET("/:id/milestones/:milestoneId", permMiddleware.CanView("projects"), milestoneController.GetMilestone)
		projects.POST("/:id/milestones", permMiddleware.CanEdit("projects"), milestoneController.CreateMilestone)
		projects.PUT("/:id/milestones/:milestoneId", permMiddleware.CanEdit("projects"), milestoneController.UpdateMilestone)
		projects.DELETE("/:id/milestones/:milestoneId", permMiddleware.CanEdit("projects"), milestoneController.DeleteMilestone)
		projects.POST("/:id/milestones/:milestoneId/complete", permMiddleware.CanEdit("projects"), milestoneController.CompleteMilestone)

		// Weekly Reports routes (nested under projects) - Project View/Edit Permission
		projects.GET("/:id/weekly-reports", permMiddleware.CanView("projects"), weeklyReportController.GetWeeklyReports)
		projects.GET("/:id/weekly-reports/export-all", permMiddleware.CanExport("projects"), weeklyReportController.ExportAllPDF)
		projects.GET("/:id/weekly-reports/:reportId", permMiddleware.CanView("projects"), weeklyReportController.GetWeeklyReport)
		projects.GET("/:id/weekly-reports/:reportId/pdf", permMiddleware.CanExport("projects"), weeklyReportController.GeneratePDF)
		projects.POST("/:id/weekly-reports", permMiddleware.CanEdit("projects"), weeklyReportController.CreateWeeklyReport)
		projects.PUT("/:id/weekly-reports/:reportId", permMiddleware.CanEdit("projects"), weeklyReportController.UpdateWeeklyReport)
		projects.DELETE("/:id/weekly-reports/:reportId", permMiddleware.CanEdit("projects"), weeklyReportController.DeleteWeeklyReport)

		// Timeline Schedule routes (nested under projects) - Project Edit Permission
		projects.GET("/:id/timeline-schedules", permMiddleware.CanView("projects"), timelineScheduleController.GetSchedules)
		projects.GET("/:id/timeline-schedules/:scheduleId", permMiddleware.CanView("projects"), timelineScheduleController.GetSchedule)
		projects.POST("/:id/timeline-schedules", permMiddleware.CanEdit("projects"), timelineScheduleController.CreateSchedule)
		projects.PUT("/:id/timeline-schedules/:scheduleId", permMiddleware.CanEdit("projects"), timelineScheduleController.UpdateSchedule)
		projects.DELETE("/:id/timeline-schedules/:scheduleId", permMiddleware.CanEdit("projects"), timelineScheduleController.DeleteSchedule)
		projects.PATCH("/:id/timeline-schedules/:scheduleId/status", permMiddleware.CanEdit("projects"), timelineScheduleController.UpdateScheduleStatus)

		// Expense Transaction routes (nested under projects) - Cost Control Permission
		projects.GET("/:id/expenses", permMiddleware.CanView("cost_control"), expenseController.GetByProject)
		projects.POST("/:id/expenses", permMiddleware.CanCreate("cost_control"), expenseController.Create)
		projects.POST("/:id/expenses/batch", permMiddleware.CanCreate("cost_control"), expenseController.BatchCreate)
		projects.GET("/:id/reports/budget-vs-actual", permMiddleware.CanView("cost_control"), expenseController.GetBudgetReport)
		projects.GET("/:id/reports/budget-vs-actual/pdf", permMiddleware.CanView("cost_control"), expenseController.ExportBudgetReportPDF)

		// CBS routes (nested under projects) - Cost Control Permission
		projects.GET("/:id/cbs", permMiddleware.CanView("cost_control"), cbsController.GetProjectCBSTree)
		projects.GET("/:id/cbs/tree", permMiddleware.CanView("cost_control"), cbsController.GetProjectCBSTree)
		projects.GET("/:id/cbs/summary", permMiddleware.CanView("cost_control"), cbsController.GetProjectBudgetSummary)

		// Single project routes - Must be at the end to avoid matching sub-routes
		projects.GET("/:id", permMiddleware.CanView("projects"), projectController.GetProjectByID)
		projects.GET("/:id/cost-summary", permMiddleware.CanView("projects"), projectController.GetProjectCostSummary)
		projects.GET("/:id/progress-history", permMiddleware.CanView("projects"), projectProgressController.GetProjectProgressHistory)
		projects.GET("/:id/actual-costs", permMiddleware.CanView("projects"), projectActualCostController.GetProjectActualCosts)
	}
}
