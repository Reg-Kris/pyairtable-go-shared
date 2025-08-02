package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/Reg-Kris/pyairtable-go-shared/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// RequestLogging returns a middleware that logs HTTP requests
func RequestLogging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Capture request body if needed (for POST/PUT requests)
		var requestBody []byte
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if c.Request.Body != nil {
				requestBody, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}
		}
		
		// Wrap response writer to capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:          bytes.NewBufferString(""),
		}
		c.Writer = writer
		
		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Header("X-Request-ID", requestID)
		}
		
		// Add request ID to context
		c.Request = c.Request.WithContext(
			AddRequestIDToContext(c.Request.Context(), requestID),
		)
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(start)
		
		// Get response size
		responseSize := writer.body.Len()
		
		// Log request details
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("status_code", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.Int64("request_size", c.Request.ContentLength),
			zap.Int("response_size", responseSize),
		}
		
		// Add user context if available
		if userID := GetUserIDFromContext(c); userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}
		
		if tenantID := GetTenantIDFromContext(c); tenantID != "" {
			fields = append(fields, zap.String("tenant_id", tenantID))
		}
		
		// Add request body for debug level (truncated)
		if log.Core().Enabled(zap.DebugLevel) && len(requestBody) > 0 {
			body := string(requestBody)
			if len(body) > 1000 {
				body = body[:1000] + "... [truncated]"
			}
			fields = append(fields, zap.String("request_body", body))
		}
		
		// Log based on status code
		statusCode := c.Writer.Status()
		switch {
		case statusCode >= 500:
			log.Error("HTTP request", fields...)
		case statusCode >= 400:
			log.Warn("HTTP request", fields...)
		default:
			log.Info("HTTP request", fields...)
		}
	}
}

// ErrorLogging returns middleware that logs errors
func ErrorLogging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Log any errors that occurred during request processing
		for _, err := range c.Errors {
			log.Error("Request error",
				zap.String("request_id", GetRequestIDFromContext(c.Request.Context())),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Error(err.Err),
				zap.String("type", err.Type.String()),
			)
		}
	}
}

// SecurityLogging returns middleware that logs security events
func SecurityLogging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		statusCode := c.Writer.Status()
		
		// Log security-related events
		switch statusCode {
		case 401:
			log.LogSecurityEvent(
				"unauthorized_access_attempt",
				GetUserIDFromContext(c),
				c.ClientIP(),
				"medium",
				map[string]interface{}{
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"user_agent": c.Request.UserAgent(),
				},
			)
		case 403:
			log.LogSecurityEvent(
				"forbidden_access_attempt",
				GetUserIDFromContext(c),
				c.ClientIP(),
				"medium",
				map[string]interface{}{
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"user_agent": c.Request.UserAgent(),
				},
			)
		case 429:
			log.LogSecurityEvent(
				"rate_limit_exceeded",
				GetUserIDFromContext(c),
				c.ClientIP(),
				"low",
				map[string]interface{}{
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"user_agent": c.Request.UserAgent(),
				},
			)
		}
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// In a real implementation, you'd use a proper UUID library
	// For now, using a simple timestamp-based ID
	return "req_" + string(rune(time.Now().UnixNano()))
}