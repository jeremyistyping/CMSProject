package utils

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger with additional functionality
type Logger struct {
	*logrus.Logger
}

// LogLevel represents log levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

// Fields represents structured log fields
type Fields map[string]interface{}

var defaultLogger *Logger

// init initializes the default logger
func init() {
	defaultLogger = NewLogger()
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	logger := logrus.New()
	
	// Set formatter
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	
	// Set output
	logger.SetOutput(os.Stdout)
	
	// Set log level from environment or default to info
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)
	
	return &Logger{Logger: logger}
}

// GetLogger returns the default logger
func GetLogger() *Logger {
	return defaultLogger
}

// WithFields adds fields to the log entry
func (l *Logger) WithFields(fields Fields) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

// WithContext adds context to the log entry
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	return l.Logger.WithContext(ctx)
}

// WithError adds error to the log entry
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// Debug logs a debug message
func (l *Logger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.Logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Logger.Fatalf(format, args...)
}

// Convenience functions using the default logger
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

func WithFields(fields Fields) *logrus.Entry {
	return defaultLogger.WithFields(fields)
}

func WithContext(ctx context.Context) *logrus.Entry {
	return defaultLogger.WithContext(ctx)
}

func WithError(err error) *logrus.Entry {
	return defaultLogger.WithError(err)
}

// ReportLogger provides specialized logging for reports
type ReportLogger struct {
	*Logger
}

// NewReportLogger creates a new report logger
func NewReportLogger() *ReportLogger {
	return &ReportLogger{
		Logger: NewLogger(),
	}
}

// LogSalesQuery logs sales query details
func (rl *ReportLogger) LogSalesQuery(operation string, startDate, endDate time.Time, groupBy string, recordsFound int, totalAmount float64) {
	rl.WithFields(Fields{
		"operation":     operation,
		"start_date":    startDate.Format("2006-01-02 15:04:05"),
		"end_date":      endDate.Format("2006-01-02 15:04:05"),
		"group_by":      groupBy,
		"records_found": recordsFound,
		"total_amount":  totalAmount,
	}).Info("Sales Query Executed")
}

// LogQueryPerformance logs query execution performance
func (rl *ReportLogger) LogQueryPerformance(operation string, duration time.Duration, recordCount int) {
	avgTime := float64(0)
	if recordCount > 0 {
		avgTime = float64(duration.Milliseconds()) / float64(recordCount)
	}
	
	rl.WithFields(Fields{
		"operation":           operation,
		"duration_ms":         duration.Milliseconds(),
		"records":             recordCount,
		"avg_ms_per_record":   avgTime,
	}).Info("Query Performance")
}

// LogReportError logs detailed error information
func (rl *ReportLogger) LogReportError(operation string, err error, context Fields) {
	logEntry := rl.WithError(err).WithFields(logrus.Fields(context))
	logEntry.Errorf("Report operation failed: %s", operation)
}

// LogDataQuality logs data quality issues
func (rl *ReportLogger) LogDataQuality(reportType string, issues []string, totalRecords int, validRecords int) {
	qualityScore := float64(0)
	if totalRecords > 0 {
		qualityScore = (float64(validRecords) / float64(totalRecords)) * 100
	}
	
	rl.WithFields(Fields{
		"report_type":    reportType,
		"total_records":  totalRecords,
		"valid_records":  validRecords,
		"quality_score":  fmt.Sprintf("%.2f%%", qualityScore),
		"issues":         issues,
	}).Warn("Data Quality Issues Detected")
}

// LogReportGeneration logs report generation details
func (rl *ReportLogger) LogReportGeneration(reportType string, params Fields, processingTime time.Duration, success bool) {
	logFields := Fields{
		"report_type":     reportType,
		"processing_time": processingTime.String(),
		"success":         success,
	}
	
	// Merge with params
	for k, v := range params {
		logFields[k] = v
	}
	
	if success {
		rl.WithFields(logFields).Info("Report generated successfully")
	} else {
		rl.WithFields(logFields).Error("Report generation failed")
	}
}

// Global report logger instance
var ReportLog = NewReportLogger()

// LogRequest logs HTTP request information
func LogRequest(method, path, userAgent, ip string, statusCode int, duration time.Duration) {
	WithFields(Fields{
		"method":      method,
		"path":        path,
		"user_agent":  userAgent,
		"ip":          ip,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
	}).Info("HTTP Request")
}

// LogDatabaseOperation logs database operations
func LogDatabaseOperation(operation, table string, duration time.Duration, err error) {
	fields := Fields{
		"operation":   operation,
		"table":       table,
		"duration_ms": duration.Milliseconds(),
	}
	
	if err != nil {
		WithFields(fields).WithError(err).Error("Database operation failed")
	} else {
		WithFields(fields).Debug("Database operation completed")
	}
}

// LogBusinessEvent logs business logic events
func LogBusinessEvent(event string, userID uint, details Fields) {
	logFields := Fields{
		"event":   event,
		"user_id": userID,
	}
	
	// Merge details into log fields
	for k, v := range details {
		logFields[k] = v
	}
	
	WithFields(logFields).Info("Business event")
}

// LogSecurityEvent logs security-related events
func LogSecurityEvent(event, userID, ip, userAgent string, success bool, details Fields) {
	logFields := Fields{
		"event":      event,
		"user_id":    userID,
		"ip":         ip,
		"user_agent": userAgent,
		"success":    success,
	}
	
	// Merge details into log fields
	for k, v := range details {
		logFields[k] = v
	}
	
	if success {
		WithFields(logFields).Info("Security event")
	} else {
		WithFields(logFields).Warn("Security event failed")
	}
}

// LogError logs application errors with context
func LogError(err error, context string, fields Fields) {
	logFields := Fields{
		"context": context,
	}
	
	// Merge fields
	for k, v := range fields {
		logFields[k] = v
	}
	
	WithFields(logFields).WithError(err).Error("Application error")
}

// Audit logs audit trail events
func Audit(userID uint, action, resource string, resourceID interface{}, changes Fields) {
	WithFields(Fields{
		"user_id":     userID,
		"action":      action,
		"resource":    resource,
		"resource_id": resourceID,
		"changes":     changes,
	}).Info("Audit event")
}

// Performance logs performance metrics
func Performance(operation string, duration time.Duration, fields Fields) {
	logFields := Fields{
		"operation":   operation,
		"duration_ms": duration.Milliseconds(),
	}
	
	// Merge fields
	for k, v := range fields {
		logFields[k] = v
	}
	
	if duration > 1000*time.Millisecond {
		WithFields(logFields).Warn("Slow operation detected")
	} else {
		WithFields(logFields).Debug("Performance metric")
	}
}

// JWT Token utility functions
// GetUserIDFromToken extracts user ID from JWT token in gin context
func GetUserIDFromToken(c *gin.Context) (uint, error) {
	// Try to get user_id from context (set by JWT middleware)
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id, nil
		}
	}
	
	// If not found in context, try to parse from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, fmt.Errorf("no authorization header")
	}
	
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return 0, fmt.Errorf("invalid authorization header format")
	}
	
	// Parse token (simplified - in production use proper JWT parsing)
	// This is a fallback, normally the middleware should set the context
	return 0, fmt.Errorf("user ID not found in token")
}

// GetUserRoleFromToken extracts user role from JWT token in gin context
func GetUserRoleFromToken(c *gin.Context) (string, error) {
	// JWT middleware sets the role under key "role"
	if userRole, exists := c.Get("role"); exists {
		if role, ok := userRole.(string); ok {
			return role, nil
		}
	}
	return "", fmt.Errorf("user role not found in token")
}
