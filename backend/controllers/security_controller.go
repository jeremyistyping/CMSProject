package controllers

import (
	"net/http"
	"strconv"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SecurityController struct {
	db              *gorm.DB
	securityService *services.SecurityService
}

func NewSecurityController(db *gorm.DB) *SecurityController {
	return &SecurityController{
		db:              db,
		securityService: services.NewSecurityService(db),
	}
}

// GetSecurityIncidents godoc
// @Summary Get security incidents
// @Description Retrieve security incidents with pagination and filters
// @Tags Security
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param severity query string false "Filter by severity (low, medium, high, critical)"
// @Param incident_type query string false "Filter by incident type"
// @Param resolved query bool false "Filter by resolution status"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/security/incidents [get]
// @Security BearerAuth
func (sc *SecurityController) GetSecurityIncidents(c *gin.Context) {
	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Filter parameters
	severity := c.Query("severity")
	incidentType := c.Query("incident_type")
	resolvedStr := c.Query("resolved")

	// Build query
	query := sc.db.Model(&models.SecurityIncident{}).Preload("User").Preload("ResolvedByUser")

	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if incidentType != "" {
		query = query.Where("incident_type = ?", incidentType)
	}
	if resolvedStr != "" {
		if resolved, err := strconv.ParseBool(resolvedStr); err == nil {
			query = query.Where("resolved = ?", resolved)
		}
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get incidents with pagination
	var incidents []models.SecurityIncident
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&incidents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch security incidents",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"incidents": incidents,
		"pagination": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetSecurityIncident godoc
// @Summary Get security incident details
// @Description Retrieve detailed information about a specific security incident
// @Tags Security
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Success 200 {object} models.SecurityIncident
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/admin/security/incidents/{id} [get]
// @Security BearerAuth
func (sc *SecurityController) GetSecurityIncident(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident ID",
		})
		return
	}

	var incident models.SecurityIncident
	if err := sc.db.Preload("User").Preload("ResolvedByUser").First(&incident, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Security incident not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch security incident",
		})
		return
	}

	c.JSON(http.StatusOK, incident)
}

// ResolveSecurityIncident godoc
// @Summary Resolve security incident
// @Description Mark a security incident as resolved with notes
// @Tags Security
// @Accept json
// @Produce json
// @Param id path int true "Incident ID"
// @Param request body map[string]string true "Resolution notes"
// @Success 200 {object} models.SecurityIncident
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/admin/security/incidents/{id}/resolve [put]
// @Security BearerAuth
func (sc *SecurityController) ResolveSecurityIncident(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid incident ID",
		})
		return
	}

	var request struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get current user
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}
	currentUser := user.(models.User)

	// Find and update incident
	var incident models.SecurityIncident
	if err := sc.db.First(&incident, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Security incident not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch security incident",
		})
		return
	}

	// Update incident
	now := time.Now()
	incident.Resolved = true
	incident.ResolvedAt = &now
	incident.ResolvedBy = &currentUser.ID
	incident.Notes = request.Notes

	if err := sc.db.Save(&incident).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to resolve incident",
		})
		return
	}

	// Reload with relations
	sc.db.Preload("User").Preload("ResolvedByUser").First(&incident, uint(id))

	c.JSON(http.StatusOK, incident)
}

// GetSystemAlerts godoc
// @Summary Get system alerts
// @Description Retrieve system security alerts with pagination and filters
// @Tags Security
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param level query string false "Filter by alert level"
// @Param acknowledged query bool false "Filter by acknowledgment status"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/security/alerts [get]
// @Security BearerAuth
func (sc *SecurityController) GetSystemAlerts(c *gin.Context) {
	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Filter parameters
	level := c.Query("level")
	acknowledgedStr := c.Query("acknowledged")

	// Build query
	query := sc.db.Model(&models.SystemAlert{}).Preload("AcknowledgedByUser")

	if level != "" {
		query = query.Where("level = ?", level)
	}
	if acknowledgedStr != "" {
		if acknowledged, err := strconv.ParseBool(acknowledgedStr); err == nil {
			query = query.Where("acknowledged = ?", acknowledged)
		}
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get alerts with pagination
	var alerts []models.SystemAlert
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch system alerts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"pagination": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// AcknowledgeAlert godoc
// @Summary Acknowledge system alert
// @Description Mark a system alert as acknowledged
// @Tags Security
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} models.SystemAlert
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/admin/security/alerts/{id}/acknowledge [put]
// @Security BearerAuth
func (sc *SecurityController) AcknowledgeAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid alert ID",
		})
		return
	}

	// Get current user
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}
	currentUser := user.(models.User)

	// Find and update alert
	var alert models.SystemAlert
	if err := sc.db.First(&alert, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "System alert not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch system alert",
		})
		return
	}

	// Update alert
	now := time.Now()
	alert.Acknowledged = true
	alert.AcknowledgedAt = &now
	alert.AcknowledgedBy = &currentUser.ID

	if err := sc.db.Save(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to acknowledge alert",
		})
		return
	}

	// Reload with relations
	sc.db.Preload("AcknowledgedByUser").First(&alert, uint(id))

	c.JSON(http.StatusOK, alert)
}

// GetSecurityMetrics godoc
// @Summary Get security metrics
// @Description Retrieve security metrics for a specific date range
// @Tags Security
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/security/metrics [get]
// @Security BearerAuth
func (sc *SecurityController) GetSecurityMetrics(c *gin.Context) {
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start date format",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end date format",
		})
		return
	}

	// Get metrics for date range
	var metrics []models.SecurityMetrics
	if err := sc.db.Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date DESC").Find(&metrics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch security metrics",
		})
		return
	}

	// Calculate summary statistics
	var totalRequests, totalSuspiciousRequests, totalIncidents int64
	var avgResponseTime float64

	for _, metric := range metrics {
		totalRequests += metric.TotalRequests
		totalSuspiciousRequests += metric.SuspiciousRequestCount
		totalIncidents += metric.SecurityIncidentCount
		avgResponseTime += metric.AvgResponseTime
	}

	if len(metrics) > 0 {
		avgResponseTime = avgResponseTime / float64(len(metrics))
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"summary": gin.H{
			"total_requests":           totalRequests,
			"total_suspicious_requests": totalSuspiciousRequests,
			"total_incidents":          totalIncidents,
			"avg_response_time":        avgResponseTime,
			"date_range": gin.H{
				"start": startDateStr,
				"end":   endDateStr,
			},
		},
	})
}

// GetIPWhitelist godoc
// @Summary Get IP whitelist
// @Description Retrieve IP whitelist entries with pagination
// @Tags Security
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param environment query string false "Filter by environment"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/security/ip-whitelist [get]
// @Security BearerAuth
func (sc *SecurityController) GetIPWhitelist(c *gin.Context) {
	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Filter parameters
	environment := c.Query("environment")

	// Build query
	query := sc.db.Model(&models.IpWhitelist{}).Preload("AddedByUser")

	if environment != "" {
		query = query.Where("environment = ? OR environment = 'all'", environment)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get whitelist entries with pagination
	var whitelist []models.IpWhitelist
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&whitelist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch IP whitelist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"whitelist": whitelist,
		"pagination": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// AddIPToWhitelist godoc
// @Summary Add IP to whitelist
// @Description Add a new IP address or range to the whitelist
// @Tags Security
// @Accept json
// @Produce json
// @Param request body models.IpWhitelist true "IP whitelist entry"
// @Success 201 {object} models.IpWhitelist
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/admin/security/ip-whitelist [post]
// @Security BearerAuth
func (sc *SecurityController) AddIPToWhitelist(c *gin.Context) {
	var request models.IpWhitelist
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get current user
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}
	currentUser := user.(models.User)

	// Set the added by user
	request.AddedBy = currentUser.ID

	if err := sc.db.Create(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add IP to whitelist",
		})
		return
	}

	// Reload with relations
	sc.db.Preload("AddedByUser").First(&request, request.ID)

	c.JSON(http.StatusCreated, request)
}

// GetSecurityConfig godoc
// @Summary Get security configuration
// @Description Retrieve security configuration settings
// @Tags Security
// @Accept json
// @Produce json
// @Param environment query string false "Filter by environment"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/security/config [get]
// @Security BearerAuth
func (sc *SecurityController) GetSecurityConfig(c *gin.Context) {
	environment := c.DefaultQuery("environment", "all")

	// Build query
	query := sc.db.Model(&models.SecurityConfig{}).Preload("LastModifiedByUser")

	if environment != "all" {
		query = query.Where("environment = ? OR environment = 'all'", environment)
	}

	var configs []models.SecurityConfig
	if err := query.Order("key ASC").Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch security configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
	})
}

// CleanupSecurityLogs godoc
// @Summary Cleanup old security logs
// @Description Trigger cleanup of old security logs and incidents
// @Tags Security
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/security/cleanup [post]
// @Security BearerAuth
func (sc *SecurityController) CleanupSecurityLogs(c *gin.Context) {
	if err := sc.securityService.CleanupOldLogs(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cleanup security logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Security logs cleanup completed successfully",
	})
}
