package controllers

import (
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MaterialTrackingController struct {
	service *services.MaterialTrackingService
}

func NewMaterialTrackingController(service *services.MaterialTrackingService) *MaterialTrackingController {
	return &MaterialTrackingController{service: service}
}

// GetSummary - GET /api/v1/material-tracking/:projectId/summary
func (c *MaterialTrackingController) GetSummary(ctx *gin.Context) {
	projectIDStr := ctx.Param("projectId")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	stats, err := c.service.GetMaterialSummary(uint(projectID))
	if err != nil {
		println("❌ GetSummary ERROR for project", projectID, ":", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetItems - GET /api/v1/material-tracking/:projectId/items
func (c *MaterialTrackingController) GetItems(ctx *gin.Context) {
	projectIDStr := ctx.Param("projectId")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	items, err := c.service.GetMaterialItems(uint(projectID))
	if err != nil {
		println("❌ GetItems ERROR for project", projectID, ":", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   items,
	})
}

// GetMovements - GET /api/v1/material-tracking/:projectId/movements
func (c *MaterialTrackingController) GetMovements(ctx *gin.Context) {
	projectIDStr := ctx.Param("projectId")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	movements, err := c.service.GetMaterialMovements(uint(projectID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   movements,
	})
}
