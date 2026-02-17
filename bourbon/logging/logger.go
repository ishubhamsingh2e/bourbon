package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogRotation defines how logs should be rotated
type LogRotation string

const (
	RotationHourly LogRotation = "hourly"
	RotationDaily  LogRotation = "daily"
	RotationWeekly LogRotation = "weekly"
	RotationNone   LogRotation = "none"
)

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	FileLogging bool
	StoragePath string
	Rotation    LogRotation
	MaxSize     int
	MaxAge      int
	MaxBackups  int
	Compress    bool
	Level       string
	Development bool
}

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
	config *LoggerConfig
	sugar  *zap.SugaredLogger
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Set defaults
	if config.StoragePath == "" {
		config.StoragePath = "storage/logs"
	}
	if config.MaxSize == 0 {
		config.MaxSize = 100 // 100MB
	}
	if config.MaxAge == 0 {
		config.MaxAge = 30 // 30 days
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = 10
	}

	// Ensure storage directory exists
	if config.FileLogging {
		if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Parse log level
	level := zapcore.InfoLevel
	if config.Level != "" {
		if err := level.UnmarshalText([]byte(config.Level)); err != nil {
			return nil, fmt.Errorf("invalid log level: %w", err)
		}
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	if config.Development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create encoder
	var encoder zapcore.Encoder
	if config.Development {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create cores
	var cores []zapcore.Core

	// Console output
	consoleCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)
	cores = append(cores, consoleCore)

	// File output
	if config.FileLogging {
		fileWriter := getLogWriter(config)
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(fileWriter),
			level,
		)
		cores = append(cores, fileCore)
	}

	// Create logger
	core := zapcore.NewTee(cores...)
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		Logger: zapLogger,
		config: config,
		sugar:  zapLogger.Sugar(),
	}, nil
}

// getLogWriter creates a writer based on rotation strategy
func getLogWriter(config *LoggerConfig) *lumberjack.Logger {
	filename := getLogFilename(config.StoragePath, config.Rotation)

	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		Compress:   config.Compress,
		LocalTime:  true,
	}
}

// getLogFilename generates filename based on rotation strategy
func getLogFilename(basePath string, rotation LogRotation) string {
	now := time.Now()

	switch rotation {
	case RotationHourly:
		return filepath.Join(basePath, fmt.Sprintf("app-%s.log", now.Format("2006-01-02-15")))
	case RotationDaily:
		return filepath.Join(basePath, fmt.Sprintf("app-%s.log", now.Format("2006-01-02")))
	case RotationWeekly:
		year, week := now.ISOWeek()
		return filepath.Join(basePath, fmt.Sprintf("app-%d-W%02d.log", year, week))
	default:
		return filepath.Join(basePath, "app.log")
	}
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *LoggerConfig {
	return &LoggerConfig{
		FileLogging: false,
		StoragePath: "storage/logs",
		Rotation:    RotationDaily,
		MaxSize:     100,
		MaxAge:      30,
		MaxBackups:  10,
		Compress:    true,
		Level:       "info",
		Development: false,
	}
}

// Sugar returns the sugared logger for easier logging
func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// WithContext returns a logger with additional context fields
func (l *Logger) WithContext(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		config: l.config,
		sugar:  l.Logger.With(fields...).Sugar(),
	}
}

// Helper methods for common logging patterns

// HTTP logs an HTTP request with standard fields
func (l *Logger) HTTP(method, path string, status int, duration time.Duration, fields ...zap.Field) {
	baseFields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", status),
		zap.Duration("duration", duration),
	}
	baseFields = append(baseFields, fields...)

	if status >= 500 {
		l.Error("HTTP request failed", baseFields...)
	} else if status >= 400 {
		l.Warn("HTTP client error", baseFields...)
	} else {
		l.Info("HTTP request", baseFields...)
	}
}

// Request logs HTTP request details
func (l *Logger) Request(method, path, ip string, fields ...zap.Field) {
	baseFields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("ip", ip),
	}
	baseFields = append(baseFields, fields...)
	l.Info("Incoming request", baseFields...)
}

// Database logs database operations
func (l *Logger) Database(operation string, duration time.Duration, fields ...zap.Field) {
	baseFields := []zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
	}
	baseFields = append(baseFields, fields...)
	l.Debug("Database operation", baseFields...)
}

// Security logs security-related events
func (l *Logger) Security(event string, fields ...zap.Field) {
	baseFields := []zap.Field{
		zap.String("event", event),
		zap.Time("timestamp", time.Now()),
	}
	baseFields = append(baseFields, fields...)
	l.Warn("Security event", baseFields...)
}

// Business logs business logic events
func (l *Logger) Business(event string, fields ...zap.Field) {
	baseFields := []zap.Field{
		zap.String("event", event),
	}
	baseFields = append(baseFields, fields...)
	l.Info("Business event", baseFields...)
}
