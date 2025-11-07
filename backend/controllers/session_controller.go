package controllers

import (
	"net/http"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
)

type SessionController struct {
	sessionCleanupService *services.SessionCleanupService
}

func NewSessionController(sessionCleanupService *services.SessionCleanupService) *SessionController {
	return &SessionController{
		sessionCleanupService: sessionCleanupService,
	}
}

// GetSessionStats returns session statistics
// @Summary Get session statistics
// @Description Get statistics about active, expired, and total sessions
// @Tags Session Management
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Session statistics"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/sessions/stats [get]
func (sc *SessionController) GetSessionStats(c *gin.Context) {
	stats := sc.sessionCleanupService.GetSessionStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ForceCleanup manually triggers session cleanup
// @Summary Force session cleanup
// @Description Manually trigger cleanup of expired sessions and tokens
// @Tags Session Management
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Cleanup completed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/sessions/cleanup [post]
func (sc *SessionController) ForceCleanup(c *gin.Context) {
	if err := sc.sessionCleanupService.ForceCleanup(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to cleanup sessions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Session cleanup completed successfully",
	})
}

// GetActiveSessions returns active sessions for the current user
// @Summary Get active sessions
// @Description Get all active sessions for the current user
// @Tags Session Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Active sessions"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/monitoring/sessions/active [get]
func (sc *SessionController) GetActiveSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found in context",
		})
		return
	}

	// Get active sessions for the current user
	var sessions []models.UserSession
	if err := sc.sessionCleanupService.GetDB().Where("user_id = ? AND is_active = ?", userID, true).
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch active sessions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"sessions": sessions,
			"count":    len(sessions),
		},
	})
}