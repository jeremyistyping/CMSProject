package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	// Add services here if needed
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// DEPRECATED: This handler is replaced by DashboardController.GetAnalytics
// Use DashboardController for real data from database instead of dummy data
func (h *DashboardHandler) GetDashboardAnalytics(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "This endpoint is deprecated. Use DashboardController.GetAnalytics for real data.",
	})
}

// DEPRECATED: This handler is replaced by DashboardController.GetStockAlertsBanner
func (h *DashboardHandler) GetStockAlerts(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "This endpoint is deprecated. Use DashboardController.GetStockAlertsBanner for real data.",
	})
}

