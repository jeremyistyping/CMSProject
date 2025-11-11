package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
)

type DailyUpdateController struct {
	service services.DailyUpdateService
}

// NewDailyUpdateController creates a new daily update controller
func NewDailyUpdateController(service services.DailyUpdateService) *DailyUpdateController {
	return &DailyUpdateController{service: service}
}

// GetDailyUpdates godoc
// @Summary Get all daily updates for a project
// @Description Get list of all daily updates for a specific project
// @Tags daily-updates
// @Produce json
// @Param projectId path int true "Project ID"
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Success 200 {array} models.DailyUpdate
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{projectId}/daily-updates [get]
func (dc *DailyUpdateController) GetDailyUpdates(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	// Check for date range filters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	
	var updates []models.DailyUpdate
	
	if startDate != "" || endDate != "" {
		updates, err = dc.service.GetDailyUpdatesByDateRange(uint(projectID), startDate, endDate)
	} else {
		updates, err = dc.service.GetDailyUpdatesByProject(uint(projectID))
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, updates)
}

// GetDailyUpdate godoc
// @Summary Get daily update by ID
// @Description Get a single daily update by ID
// @Tags daily-updates
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Daily Update ID"
// @Success 200 {object} models.DailyUpdate
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/daily-updates/{id} [get]
func (dc *DailyUpdateController) GetDailyUpdate(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	updateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid daily update ID"})
		return
	}
	
	update, err := dc.service.GetDailyUpdateByID(uint(projectID), uint(updateID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, update)
}

// CreateDailyUpdate godoc
// @Summary Create a new daily update
// @Description Create a new daily update for a project
// @Tags daily-updates
// @Accept json
// @Produce json
// @Param projectId path int true "Project ID"
// @Param dailyUpdate body models.DailyUpdate true "Daily Update object"
// @Success 201 {object} models.DailyUpdate
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{projectId}/daily-updates [post]
func (dc *DailyUpdateController) CreateDailyUpdate(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	var dailyUpdate models.DailyUpdate
	
	if err := c.ShouldBindJSON(&dailyUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Set the project ID from URL param
	dailyUpdate.ProjectID = uint(projectID)
	
	if err := dc.service.CreateDailyUpdate(&dailyUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, dailyUpdate)
}

// UpdateDailyUpdate godoc
// @Summary Update a daily update
// @Description Update an existing daily update
// @Tags daily-updates
// @Accept json
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Daily Update ID"
// @Param dailyUpdate body models.DailyUpdate true "Daily Update object"
// @Success 200 {object} models.DailyUpdate
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/daily-updates/{id} [put]
func (dc *DailyUpdateController) UpdateDailyUpdate(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	updateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid daily update ID"})
		return
	}
	
	var dailyUpdate models.DailyUpdate
	if err := c.ShouldBindJSON(&dailyUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	dailyUpdate.ID = uint(updateID)
	dailyUpdate.ProjectID = uint(projectID)
	
	if err := dc.service.UpdateDailyUpdate(&dailyUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, dailyUpdate)
}

// DeleteDailyUpdate godoc
// @Summary Delete a daily update
// @Description Delete a daily update by ID
// @Tags daily-updates
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Daily Update ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/daily-updates/{id} [delete]
func (dc *DailyUpdateController) DeleteDailyUpdate(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	updateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid daily update ID"})
		return
	}
	
	if err := dc.service.DeleteDailyUpdate(uint(projectID), uint(updateID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Daily update deleted successfully"})
}

