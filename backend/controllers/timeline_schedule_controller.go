package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"log"
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
)

type TimelineScheduleController struct {
	service services.TimelineScheduleService
}

// NewTimelineScheduleController creates a new timeline schedule controller
func NewTimelineScheduleController(service services.TimelineScheduleService) *TimelineScheduleController {
	return &TimelineScheduleController{service: service}
}

// GetSchedules godoc
// @Summary Get all timeline schedules for a project
// @Description Get list of all timeline schedules for a specific project
// @Tags timeline-schedules
// @Produce json
// @Param projectId path int true "Project ID"
// @Success 200 {array} models.TimelineSchedule
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{projectId}/timeline-schedules [get]
func (tc *TimelineScheduleController) GetSchedules(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid project ID",
			"project_id": projectIDStr,
			"details":    err.Error(),
		})
		return
	}
	
	schedules, err := tc.service.GetSchedulesByProject(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to response format with computed fields
	responses := make([]*models.TimelineScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = schedule.ToResponse()
	}
	
	c.JSON(http.StatusOK, responses)
}

// GetSchedule godoc
// @Summary Get timeline schedule by ID
// @Description Get a single timeline schedule by ID
// @Tags timeline-schedules
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Schedule ID"
// @Success 200 {object} models.TimelineScheduleResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/timeline-schedules/{id} [get]
func (tc *TimelineScheduleController) GetSchedule(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	scheduleID, err := strconv.ParseUint(c.Param("scheduleId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}
	
	schedule, err := tc.service.GetScheduleByID(uint(projectID), uint(scheduleID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, schedule.ToResponse())
}

// CreateSchedule godoc
// @Summary Create a new timeline schedule
// @Description Create a new timeline schedule for a project
// @Tags timeline-schedules
// @Accept json
// @Produce json
// @Param projectId path int true "Project ID"
// @Param schedule body models.TimelineSchedule true "Timeline Schedule object"
// @Success 201 {object} models.TimelineScheduleResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{projectId}/timeline-schedules [post]
func (tc *TimelineScheduleController) CreateSchedule(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid project ID",
			"project_id": projectIDStr,
			"details":    err.Error(),
		})
		return
	}
	
	var schedule models.TimelineSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Set the project ID from URL param
	schedule.ProjectID = uint(projectID)
	
	// Validate required fields
	if schedule.WorkArea == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Work area is required",
		})
		return
	}
	
	if schedule.StartDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Start date is required",
		})
		return
	}
	
	if schedule.EndDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "End date is required",
		})
		return
	}
	
	if err := tc.service.CreateSchedule(&schedule); err != nil {
		// Log the error for debugging
		log.Printf("[ERROR] Failed to create timeline schedule: %v, Schedule: %+v", err, schedule)
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to create timeline schedule",
			"details":    err.Error(),
			"project_id": projectID,
			"schedule":   schedule,
		})
		return
	}
	
	c.JSON(http.StatusCreated, schedule.ToResponse())
}

// UpdateSchedule godoc
// @Summary Update a timeline schedule
// @Description Update an existing timeline schedule
// @Tags timeline-schedules
// @Accept json
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Schedule ID"
// @Param schedule body models.TimelineSchedule true "Timeline Schedule object"
// @Success 200 {object} models.TimelineScheduleResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/timeline-schedules/{id} [put]
func (tc *TimelineScheduleController) UpdateSchedule(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	scheduleID, err := strconv.ParseUint(c.Param("scheduleId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}
	
	var schedule models.TimelineSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	schedule.ID = uint(scheduleID)
	schedule.ProjectID = uint(projectID)
	
	if err := tc.service.UpdateSchedule(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, schedule.ToResponse())
}

// DeleteSchedule godoc
// @Summary Delete a timeline schedule
// @Description Delete a timeline schedule by ID
// @Tags timeline-schedules
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Schedule ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/timeline-schedules/{id} [delete]
func (tc *TimelineScheduleController) DeleteSchedule(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	scheduleID, err := strconv.ParseUint(c.Param("scheduleId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}
	
	if err := tc.service.DeleteSchedule(uint(projectID), uint(scheduleID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Timeline schedule deleted successfully",
	})
}

// UpdateScheduleStatus godoc
// @Summary Update timeline schedule status
// @Description Update the status of a timeline schedule
// @Tags timeline-schedules
// @Accept json
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Schedule ID"
// @Param status body map[string]string true "Status object"
// @Success 200 {object} models.TimelineScheduleResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/timeline-schedules/{id}/status [patch]
func (tc *TimelineScheduleController) UpdateScheduleStatus(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	scheduleID, err := strconv.ParseUint(c.Param("scheduleId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}
	
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := tc.service.UpdateScheduleStatus(uint(projectID), uint(scheduleID), req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get updated schedule
	schedule, err := tc.service.GetScheduleByID(uint(projectID), uint(scheduleID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, schedule.ToResponse())
}

