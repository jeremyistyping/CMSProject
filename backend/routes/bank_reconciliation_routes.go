package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"

	"github.com/gin-gonic/gin"
)

func SetupBankReconciliationRoutes(router *gin.Engine, controller *controllers.BankReconciliationController, jwtManager *middleware.JWTManager) {
	// Public group dengan authentication
	api := router.Group("/api/bank-reconciliation")
	api.Use(jwtManager.AuthRequired())

	// ========== SNAPSHOT ROUTES ==========
	snapshots := api.Group("/snapshots")
	{
		snapshots.POST("", controller.GenerateSnapshot)           // Generate new snapshot
		snapshots.GET("", controller.GetSnapshots)                // Get all snapshots for account
		snapshots.GET("/:id", controller.GetSnapshotByID)         // Get snapshot detail
		snapshots.POST("/:id/lock", controller.LockSnapshot)      // Lock snapshot (period close)
	}

	// ========== RECONCILIATION ROUTES ==========
	reconciliations := api.Group("/reconciliations")
	{
		reconciliations.GET("", controller.GetReconciliations)              // Get all reconciliations for account
		reconciliations.GET("/:id", controller.GetReconciliationByID)       // Get reconciliation detail
		reconciliations.POST("/:id/approve", controller.ApproveReconciliation) // Approve reconciliation (manager)
		reconciliations.POST("/:id/reject", controller.RejectReconciliation)   // Reject reconciliation (manager)
	}

	// Perform reconciliation (separate endpoint)
	api.POST("/reconcile", controller.PerformReconciliation)

	// ========== AUDIT TRAIL ROUTES ==========
	auditTrail := api.Group("/audit-trail")
	{
		auditTrail.GET("", controller.GetAuditTrail)      // Get audit trail for account
		auditTrail.POST("", controller.LogAuditTrail)     // Log audit trail entry
	}
}
