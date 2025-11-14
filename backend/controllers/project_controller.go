package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
)

type ProjectController struct {
	service services.ProjectService
}

// NewProjectController creates a new project controller
func NewProjectController(service services.ProjectService) *ProjectController {
	return &ProjectController{service: service}
}

// GetAllProjects godoc
// @Summary Get all projects
// @Description Get list of all projects
// @Tags projects
// @Produce json
// @Success 200 {array} models.Project
// @Failure 500 {object} map[string]interface{}
// @Router /projects [get]
func (pc *ProjectController) GetAllProjects(c *gin.Context) {
	projects, err := pc.service.GetAllProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, projects)
}

// GetProjectByID godoc
// @Summary Get project by ID
// @Description Get a single project by ID
// @Tags projects
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} models.Project
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{id} [get]
func (pc *ProjectController) GetProjectByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	project, err := pc.service.GetProjectByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, project)
}

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new construction project
// @Tags projects
// @Accept json
// @Produce json
// @Param project body models.Project true "Project object"
// @Success 201 {object} models.Project
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /projects [post]
func (pc *ProjectController) CreateProject(c *gin.Context) {
	var project models.Project
	
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := pc.service.CreateProject(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Ensure the created project is returned with all fields including ID
	// Fetch the created project from database to get the complete data
	createdProject, err := pc.service.GetProjectByID(project.ID)
	if err != nil {
		// If we can't fetch it, return what we have (with ID from Create)
		c.JSON(http.StatusCreated, project)
		return
	}
	
	c.JSON(http.StatusCreated, createdProject)
}

// UpdateProject godoc
// @Summary Update a project
// @Description Update an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param project body models.Project true "Project object"
// @Success 200 {object} models.Project
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{id} [put]
func (pc *ProjectController) UpdateProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	project.ID = uint(id)
	
	if err := pc.service.UpdateProject(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, project)
}

// DeleteProject godoc
// @Summary Delete a project
// @Description Delete a project by ID
// @Tags projects
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{id} [delete]
func (pc *ProjectController) DeleteProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	if err := pc.service.DeleteProject(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// ArchiveProject godoc
// @Summary Archive a project
// @Description Archive a project by changing its status
// @Tags projects
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{id}/archive [post]
func (pc *ProjectController) ArchiveProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	if err := pc.service.ArchiveProject(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Project archived successfully"})
}

// UpdateProgress godoc
// @Summary Update project progress
// @Description Update progress values for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param progress body map[string]float64 true "Progress data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{id}/progress [patch]
func (pc *ProjectController) UpdateProgress(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	var progressData map[string]float64
	if err := c.ShouldBindJSON(&progressData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := pc.service.UpdateProgress(uint(id), progressData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Progress updated successfully"})
}

// GetProjectsByStatus godoc
// @Summary Get projects by status
// @Description Get list of projects filtered by status
// @Tags projects
// @Produce json
// @Param status query string true "Project status"
// @Success 200 {array} models.Project
// @Failure 500 {object} map[string]interface{}
// @Router /projects/status [get]
func (pc *ProjectController) GetProjectsByStatus(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status parameter is required"})
		return
	}
	
	projects, err := pc.service.GetProjectsByStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, projects)
}

// GetProjectCostSummary godoc
// @Summary Get project cost summary (Budget vs Actual + Progress)
// @Description High-level Budget vs Actual per project including progress percentage
// @Tags projects
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} models.ProjectCostSummary
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /projects/{id}/cost-summary [get]
func (pc *ProjectController) GetProjectCostSummary(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	summary, err := pc.service.GetProjectCostSummary(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetActiveProjects godoc
// @Summary Get active projects
// @Description Get list of all active projects (not archived)
// @Tags projects
// @Produce json
// @Success 200 {array} models.Project
// @Failure 500 {object} map[string]interface{}
// @Router /projects/active [get]
func (pc *ProjectController) GetActiveProjects(c *gin.Context) {
	projects, err := pc.service.GetActiveProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, projects)
}

