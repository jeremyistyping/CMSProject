package handlers

import (
	"net/http"
	"context"
	"time"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

type StartupHandler struct {
	StartupService *services.StartupService
}

func NewStartupHandler(startupService *services.StartupService) *StartupHandler {
	return &StartupHandler{
		StartupService: startupService,
	}
}

// GetStartupStatus returns the status of startup services including account header status
func (h *StartupHandler) GetStartupStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	
	status, err := h.StartupService.GetStartupStatus(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get startup status",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Startup status retrieved successfully",
		"data":    status,
	})
}

// TriggerAccountHeaderFix is disabled in project management mode
func (h *StartupHandler) TriggerAccountHeaderFix(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account header fix is not applicable in project management mode",
	})
}
