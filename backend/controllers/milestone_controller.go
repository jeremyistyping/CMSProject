package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"log"
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
)

type MilestoneController struct {
	service services.MilestoneService
}

// NewMilestoneController creates a new milestone controller
func NewMilestoneController(service services.MilestoneService) *MilestoneController {
	return &MilestoneController{service: service}
}

// GetMilestones godoc
// @Summary Get all milestones for a project
// @Description Get list of all milestones for a specific project
// @Tags milestones
// @Produce json
// @Param projectId path int true "Project ID"
// @Success 200 {array} models.Milestone
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{projectId}/milestones [get]
func (mc *MilestoneController) GetMilestones(c *gin.Context) {
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
	
	milestones, err := mc.service.GetMilestonesByProject(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, milestones)
}

// GetMilestone godoc
// @Summary Get milestone by ID
// @Description Get a single milestone by ID
// @Tags milestones
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Milestone ID"
// @Success 200 {object} models.Milestone
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/milestones/{id} [get]
func (mc *MilestoneController) GetMilestone(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	milestoneID, err := strconv.ParseUint(c.Param("milestoneId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone ID"})
		return
	}
	
	milestone, err := mc.service.GetMilestoneByID(uint(projectID), uint(milestoneID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, milestone)
}

// CreateMilestone godoc
// @Summary Create a new milestone
// @Description Create a new milestone for a project
// @Tags milestones
// @Accept json
// @Produce json
// @Param projectId path int true "Project ID"
// @Param milestone body models.Milestone true "Milestone object"
// @Success 201 {object} models.Milestone
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects/{projectId}/milestones [post]
func (mc *MilestoneController) CreateMilestone(c *gin.Context) {
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
	
	var milestone models.Milestone
	if err := c.ShouldBindJSON(&milestone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Set the project ID from URL param
	milestone.ProjectID = uint(projectID)
	
	// Validate required fields
	if milestone.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Title is required",
		})
		return
	}
	
	if milestone.TargetDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Target date is required",
		})
		return
	}
	
	if err := mc.service.CreateMilestone(&milestone); err != nil {
		// Log the error for debugging
		log.Printf("[ERROR] Failed to create milestone: %v, Milestone: %+v", err, milestone)
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to create milestone",
			"details":    err.Error(),
			"project_id": projectID,
			"milestone":  milestone,
		})
		return
	}
	
	c.JSON(http.StatusCreated, milestone)
}

// UpdateMilestone godoc
// @Summary Update a milestone
// @Description Update an existing milestone
// @Tags milestones
// @Accept json
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Milestone ID"
// @Param milestone body models.Milestone true "Milestone object"
// @Success 200 {object} models.Milestone
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/milestones/{id} [put]
func (mc *MilestoneController) UpdateMilestone(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	milestoneID, err := strconv.ParseUint(c.Param("milestoneId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone ID"})
		return
	}
	
	var milestone models.Milestone
	if err := c.ShouldBindJSON(&milestone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	milestone.ID = uint(milestoneID)
	milestone.ProjectID = uint(projectID)
	
	if err := mc.service.UpdateMilestone(&milestone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, milestone)
}

// DeleteMilestone godoc
// @Summary Delete a milestone
// @Description Delete a milestone by ID
// @Tags milestones
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Milestone ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/milestones/{id} [delete]
func (mc *MilestoneController) DeleteMilestone(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	milestoneID, err := strconv.ParseUint(c.Param("milestoneId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone ID"})
		return
	}
	
	if err := mc.service.DeleteMilestone(uint(projectID), uint(milestoneID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Milestone deleted successfully"})
}

// CompleteMilestone godoc
// @Summary Mark milestone as completed
// @Description Mark a milestone as completed with current date
// @Tags milestones
// @Produce json
// @Param projectId path int true "Project ID"
// @Param id path int true "Milestone ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{projectId}/milestones/{id}/complete [post]
func (mc *MilestoneController) CompleteMilestone(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	milestoneID, err := strconv.ParseUint(c.Param("milestoneId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid milestone ID"})
		return
	}
	
	if err := mc.service.CompleteMilestone(uint(projectID), uint(milestoneID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":      "Milestone marked as completed",
		"completed_at": time.Now(),
	})
}

