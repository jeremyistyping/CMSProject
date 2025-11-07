package utils

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Context keys
type contextKey string

const (
	UserIDKey     contextKey = "user_id"
	UserRoleKey   contextKey = "user_role"
	UserNameKey   contextKey = "user_name"
	CompanyIDKey  contextKey = "company_id"
	RequestIDKey  contextKey = "request_id"
)

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (uint, error) {
	if userID, ok := ctx.Value(UserIDKey).(uint); ok {
		return userID, nil
	}
	
	// Try to get from Gin context if available
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if userIDStr, exists := ginCtx.Get("user_id"); exists {
			if userID, ok := userIDStr.(uint); ok {
				return userID, nil
			}
			// Try to parse from string
			if userIDStr, ok := userIDStr.(string); ok {
				if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
					return uint(userID), nil
				}
			}
		}
	}
	
	return 0, errors.New("user ID not found in context")
}

// GetUserRoleFromContext extracts user role from context
func GetUserRoleFromContext(ctx context.Context) (string, error) {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role, nil
	}
	
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if role, exists := ginCtx.Get("user_role"); exists {
			if roleStr, ok := role.(string); ok {
				return roleStr, nil
			}
		}
	}
	
	return "", errors.New("user role not found in context")
}

// GetUserNameFromContext extracts username from context
func GetUserNameFromContext(ctx context.Context) (string, error) {
	if username, ok := ctx.Value(UserNameKey).(string); ok {
		return username, nil
	}
	
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if username, exists := ginCtx.Get("username"); exists {
			if usernameStr, ok := username.(string); ok {
				return usernameStr, nil
			}
		}
	}
	
	return "", errors.New("username not found in context")
}

// SetUserContext sets user information in context
func SetUserContext(ctx context.Context, userID uint, role, username string) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, UserRoleKey, role)
	ctx = context.WithValue(ctx, UserNameKey, username)
	return ctx
}

// SetUserContextInGin sets user information in Gin context
func SetUserContextInGin(c *gin.Context, userID uint, role, username string) {
	c.Set("user_id", userID)
	c.Set("user_role", role)
	c.Set("username", username)
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if requestID, exists := ginCtx.Get("request_id"); exists {
			if requestIDStr, ok := requestID.(string); ok {
				return requestIDStr
			}
		}
	}
	
	return ""
}

// GetClientIP gets client IP from context
func GetClientIPFromContext(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.ClientIP()
	}
	return "unknown"
}

// GetUserAgent gets user agent from context
func GetUserAgentFromContext(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.GetHeader("User-Agent")
	}
	return "unknown"
}