package controllers

import (
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ProjectProgressController handles HTTP endpoints for project progress history
// backed by project_progress table.
type ProjectProgressController struct {
	service services.ProjectProgressService
}

// NewProjectProgressController creates a new controller instance
func NewProjectProgressController(service services.ProjectProgressService) *ProjectProgressController {
	return &ProjectProgressController{service: service}
}

// GetProjectProgressHistory godoc
// @Summary Get project progress history
// @Description Get list of progress snapshots for a project (optionally filtered by date range)
// @Tags projects
// @Produce json
// @Param id path int true "Project ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} models.ProjectProgress
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{id}/progress-history [get]
func (pc *ProjectProgressController) GetProjectProgressHistory(c *gin.Context) {
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

	records, err := pc.service.GetProgressHistory(uint(projectID), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

// projectProgressInput represents payload for creating/updating a progress snapshot.
type projectProgressInput struct {
	Date                    string   `json:"date" binding:"required"`
	PhysicalProgressPercent float64  `json:"physical_progress_percent" binding:"required"`
	VolumeAchieved          *float64 `json:"volume_achieved"`
	Remarks                 string   `json:"remarks"`
}

// UpsertProjectProgress godoc
// @Summary Create or update project progress snapshot
// @Description Create/update a progress snapshot for a given date (per project)
// @Tags projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param body body projectProgressInput true "Progress snapshot"
// @Success 200 {object} models.ProjectProgress
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{id}/progress-history [post]
func (pc *ProjectProgressController) UpsertProjectProgress(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var input projectProgressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		// fallback to RFC3339 if full timestamp is sent
		parsedDate, err = time.Parse(time.RFC3339, input.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, expected YYYY-MM-DD or RFC3339"})
			return
		}
	}

	progress, err := pc.service.UpsertProgressSnapshot(uint(projectID), parsedDate, input.PhysicalProgressPercent, input.VolumeAchieved, input.Remarks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}
