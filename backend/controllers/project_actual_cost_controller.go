package controllers

import (
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ProjectActualCostController handles HTTP endpoints for project_actual_costs view
// exposing derived actual costs per project.
type ProjectActualCostController struct {
	service services.ProjectActualCostService
}

// NewProjectActualCostController creates a new controller instance
func NewProjectActualCostController(service services.ProjectActualCostService) *ProjectActualCostController {
	return &ProjectActualCostController{service: service}
}

// GetProjectActualCosts godoc
// @Summary Get actual costs for a project
// @Description Get list of actual cost rows for a project, optionally filtered by date range
// @Tags projects
// @Produce json
// @Param id path int true "Project ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} models.ProjectActualCost
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{id}/actual-costs [get]
func (pc *ProjectActualCostController) GetProjectActualCosts(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startStr != "" {
		parsed, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, expected YYYY-MM-DD"})
			return
		}
		startDate = &parsed
	}
	if endStr != "" {
		parsed, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, expected YYYY-MM-DD"})
			return
		}
		endDate = &parsed
	}

	rows, err := pc.service.GetActualCosts(uint(projectID), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rows)
}
