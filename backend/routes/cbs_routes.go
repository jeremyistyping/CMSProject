package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"

	"github.com/gin-gonic/gin"
)

func SetupCBSRoutes(router *gin.RouterGroup, cbsController *controllers.CBSController, permMiddleware *middleware.PermissionMiddleware) {
	// CBS Routes
	cbs := router.Group("/cbs-nodes")
	{
		cbs.POST("", permMiddleware.CanCreate("cbs"), cbsController.CreateCBSNode)
		cbs.PUT("/:id", permMiddleware.CanEdit("cbs"), cbsController.UpdateCBSNode)
		cbs.DELETE("/:id", permMiddleware.CanDelete("cbs"), cbsController.DeleteCBSNode)
		cbs.GET("/:id/summary", permMiddleware.CanView("cbs"), cbsController.GetNodeSummary)
	}

	// Project CBS Routes
	projects := router.Group("/projects")
	{
		projects.GET("/:id/cbs", permMiddleware.CanView("cbs"), cbsController.GetProjectCBSTree)
		projects.GET("/:id/cbs/summary", permMiddleware.CanView("cbs"), cbsController.GetProjectBudgetSummary)
	}

	// Purchase Request CBS Mappings Routes
	purchaseRequests := router.Group("/purchase-requests")
	{
		purchaseRequests.GET("/:id/cbs-mappings", permMiddleware.CanView("cbs"), cbsController.GetPRCBSMappings)
		purchaseRequests.POST("/:id/verify", permMiddleware.CanCreate("cbs"), cbsController.VerifyPurchaseRequest)
	}
}
