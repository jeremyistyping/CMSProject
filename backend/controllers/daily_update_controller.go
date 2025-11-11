package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"
	"log"
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
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
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	updateID, err := strconv.ParseUint(c.Param("updateId"), 10, 32)
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
	
	log.Printf("📝 CreateDailyUpdate - Project ID: %d, Content-Type: %s", projectID, c.GetHeader("Content-Type"))
	
	var dailyUpdate models.DailyUpdate
	
	// Check Content-Type to determine how to parse
	contentType := c.GetHeader("Content-Type")
	
	// Check if it's multipart/form-data (contains "multipart/form-data" prefix)
	isMultipart := contentType != "" && (contentType == "multipart/form-data" || 
		len(contentType) > 19 && contentType[:19] == "multipart/form-data")
	
	if !isMultipart && contentType == "application/json" {
		// JSON request without files
		if err := c.ShouldBindJSON(&dailyUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}
	} else {
		// Multipart form data with files
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid multipart form",
				"details": err.Error(),
			})
			return
		}
		
		// Parse form fields
		dateStr := c.PostForm("date")
		if dateStr != "" {
			// Try parsing ISO 8601 format first
			parsedDate, err := time.Parse(time.RFC3339, dateStr)
			if err != nil {
				// Try parsing YYYY-MM-DD format
				parsedDate, err = time.Parse("2006-01-02", dateStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid date format",
						"details": "Expected ISO 8601 or YYYY-MM-DD format",
					})
					return
				}
			}
			dailyUpdate.Date = parsedDate
		}
		
		dailyUpdate.Weather = c.PostForm("weather")
		dailyUpdate.WorkDescription = c.PostForm("work_description")
		dailyUpdate.MaterialsUsed = c.PostForm("materials_used")
		dailyUpdate.Issues = c.PostForm("issues")
		dailyUpdate.CreatedBy = c.PostForm("created_by")
		
		log.Printf("📋 Parsed Data - Date: %v, Weather: %s, Workers: %d, Description: %s", dailyUpdate.Date, dailyUpdate.Weather, dailyUpdate.WorkersPresent, dailyUpdate.WorkDescription)
		
		// Parse workers_present
		workersStr := c.PostForm("workers_present")
		if workersStr != "" {
			workers, err := strconv.Atoi(workersStr)
			if err == nil {
				dailyUpdate.WorkersPresent = workers
			}
		}
		
		// Handle photo uploads
		files := form.File["photos"]
		if len(files) > 0 {
			config := utils.DefaultUploadConfig()
			filePaths, err := utils.SaveMultipleFiles(files, config)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Failed to upload photos",
					"details": err.Error(),
				})
				return
			}
			// Convert file paths to public URLs
			var photoURLs []string
			for _, path := range filePaths {
				photoURLs = append(photoURLs, utils.GetPublicURL(path))
			}
			dailyUpdate.Photos = pq.StringArray(photoURLs)
		}
	}
	
	// Set the project ID from URL param
	dailyUpdate.ProjectID = uint(projectID)
	
	// Validate required fields
	if dailyUpdate.Date.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Date is required",
		})
		return
	}
	
	if dailyUpdate.WorkDescription == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Work description is required",
		})
		return
	}
	
	log.Printf("💾 Calling service.CreateDailyUpdate...")
	if err := dc.service.CreateDailyUpdate(&dailyUpdate); err != nil {
		log.Printf("❌ Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to create daily update",
			"details":    err.Error(),
			"project_id": projectID,
		})
		return
	}
	
	log.Printf("✅ Daily update created successfully - ID: %d", dailyUpdate.ID)
	
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
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	updateID, err := strconv.ParseUint(c.Param("updateId"), 10, 32)
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
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	updateID, err := strconv.ParseUint(c.Param("updateId"), 10, 32)
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

