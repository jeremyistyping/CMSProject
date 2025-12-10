package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupExpenseTransactionRoutes sets up all expense transaction routes
func SetupExpenseTransactionRoutes(router *gin.RouterGroup, db *gorm.DB, permMiddleware *middleware.PermissionMiddleware) {
	// Initialize repositories
	expenseRepo := repositories.NewExpenseTransactionRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	coaRepo := repositories.NewCOARepository(db)

	// Initialize services
	expenseService := services.NewExpenseTransactionService(expenseRepo, projectRepo, coaRepo)

	// Initialize controllers
	expenseController := controllers.NewExpenseTransactionController(expenseService)

	// Expense Transaction routes
	expenses := router.Group("/expense-transactions")
	{
		expenses.GET("", permMiddleware.CanView("cost_control"), expenseController.GetAll)
		expenses.GET("/:id", permMiddleware.CanView("cost_control"), expenseController.GetByID)
		expenses.PUT("/:id", permMiddleware.CanEdit("cost_control"), expenseController.Update)
		expenses.DELETE("/:id", permMiddleware.CanDelete("cost_control"), expenseController.Delete)
	}

	// Note: Project-specific routes are handled in project_routes.go to avoid wildcard conflicts
	// The routes are:
	// GET    /api/v1/projects/:id/expenses
	// POST   /api/v1/projects/:id/expenses
	// POST   /api/v1/projects/:id/expenses/batch
	// GET    /api/v1/projects/:id/reports/budget-vs-actual
}
