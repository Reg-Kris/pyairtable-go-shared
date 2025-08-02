// Package logger provides structured logging utilities using Zap
package logger

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Reg-Kris/pyairtable-go-shared/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
	config *config.LoggerConfig
}

// New creates a new logger instance
func New(cfg *config.LoggerConfig) (*Logger, error) {
	zapConfig := zap.NewProductionConfig()
	
	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %s: %w", cfg.Level, err)
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	
	// Set encoding format
	switch strings.ToLower(cfg.Format) {
	case "json":
		zapConfig.Encoding = "json"
	case "console":
		zapConfig.Encoding = "console"
	default:
		zapConfig.Encoding = "json"
	}
	
	// Set output paths
	if cfg.OutputPath != "stdout" && cfg.OutputPath != "" {
		zapConfig.OutputPaths = []string{cfg.OutputPath}
		zapConfig.ErrorOutputPaths = []string{cfg.OutputPath}
	}
	
	// Configure encoder
	zapConfig.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	
	logger, err := zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}
	
	return &Logger{
		Logger: logger,
		config: cfg,
	}, nil
}

// NewDevelopment creates a development logger
func NewDevelopment() (*Logger, error) {
	cfg := &config.LoggerConfig{
		Level:      "debug",
		Format:     "console",
		OutputPath: "stdout",
	}
	return New(cfg)
}

// NewProduction creates a production logger
func NewProduction() (*Logger, error) {
	cfg := &config.LoggerConfig{
		Level:      "info",
		Format:     "json",
		OutputPath: "stdout",
	}
	return New(cfg)
}

// WithContext returns a logger with context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.Logger
	
	// Add request ID if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		logger = logger.With(zap.String("request_id", requestID.(string)))
	}
	
	// Add user ID if available
	if userID := ctx.Value("user_id"); userID != nil {
		logger = logger.With(zap.String("user_id", fmt.Sprintf("%v", userID)))
	}
	
	// Add tenant ID if available
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		logger = logger.With(zap.String("tenant_id", fmt.Sprintf("%v", tenantID)))
	}
	
	return &Logger{
		Logger: logger,
		config: l.config,
	}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	
	return &Logger{
		Logger: l.Logger.With(zapFields...),
		config: l.config,
	}
}

// WithError returns a logger with error field
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.Error(err)),
		config: l.config,
	}
}

// LogHTTPRequest logs HTTP request details
func (l *Logger) LogHTTPRequest(method, path, userAgent, clientIP string, statusCode int, duration int64) {
	l.Info("HTTP request",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("user_agent", userAgent),
		zap.String("client_ip", clientIP),
		zap.Int("status_code", statusCode),
		zap.Int64("duration_ms", duration),
	)
}

// LogDatabaseQuery logs database query details
func (l *Logger) LogDatabaseQuery(query string, duration int64, affected int64) {
	l.Debug("Database query",
		zap.String("query", query),
		zap.Int64("duration_ms", duration),
		zap.Int64("affected_rows", affected),
	)
}

// LogCacheOperation logs cache operation details
func (l *Logger) LogCacheOperation(operation, key string, hit bool, duration int64) {
	l.Debug("Cache operation",
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Bool("hit", hit),
		zap.Int64("duration_ms", duration),
	)
}

// LogMetric logs metric information
func (l *Logger) LogMetric(name string, value float64, tags map[string]string) {
	fields := []zap.Field{
		zap.String("metric_name", name),
		zap.Float64("value", value),
	}
	
	for key, val := range tags {
		fields = append(fields, zap.String("tag_"+key, val))
	}
	
	l.Info("Metric", fields...)
}

// LogUserAction logs user action for audit purposes
func (l *Logger) LogUserAction(userID, action, resource string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
	}
	
	for key, value := range metadata {
		fields = append(fields, zap.Any("meta_"+key, value))
	}
	
	l.Info("User action", fields...)
}

// LogSecurityEvent logs security-related events
func (l *Logger) LogSecurityEvent(event, userID, clientIP string, severity string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "security"),
		zap.String("event", event),
		zap.String("severity", severity),
	}
	
	if userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	
	if clientIP != "" {
		fields = append(fields, zap.String("client_ip", clientIP))
	}
	
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}
	
	l.Warn("Security event", fields...)
}

// LogPerformanceMetric logs performance metrics
func (l *Logger) LogPerformanceMetric(operation string, duration int64, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Int64("duration_ms", duration),
	}
	
	for key, value := range metadata {
		fields = append(fields, zap.Any(key, value))
	}
	
	l.Info("Performance metric", fields...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Close closes the logger
func (l *Logger) Close() error {
	return l.Sync()
}

// Global logger instance
var defaultLogger *Logger

// init initializes the default logger
func init() {
	var err error
	if os.Getenv("ENVIRONMENT") == "development" {
		defaultLogger, err = NewDevelopment()
	} else {
		defaultLogger, err = NewProduction()
	}
	
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize default logger: %v", err))
	}
}

// Global logging functions

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	defaultLogger.Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	defaultLogger.Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	defaultLogger.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	defaultLogger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	defaultLogger.Fatal(msg, fields...)
}

// With returns a logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	return defaultLogger.With(fields...)
}

// SetDefault sets the default logger
func SetDefault(logger *Logger) {
	defaultLogger = logger
}