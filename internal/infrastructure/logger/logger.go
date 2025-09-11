package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"go-web-starter/internal/config"
)

// Logger wraps logrus.Logger with additional functionality
type Logger struct {
	*logrus.Logger
}

// New creates a new logger instance based on configuration
func New(cfg *config.LoggerConfig) (*Logger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	logger.SetLevel(level)

	// Set log format
	switch cfg.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		return nil, fmt.Errorf("unsupported log format: %s", cfg.Format)
	}

	// Set output
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "file":
		// Ensure log directory exists
		logDir := filepath.Dir(cfg.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Configure log rotation
		output = &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
	case "both":
		// Log to both stdout and file
		logDir := filepath.Dir(cfg.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		fileOutput := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		output = io.MultiWriter(os.Stdout, fileOutput)
	default:
		return nil, fmt.Errorf("unsupported log output: %s", cfg.Output)
	}

	logger.SetOutput(output)

	return &Logger{Logger: logger}, nil
}

// WithRequestID adds request ID to log context
func (l *Logger) WithRequestID(requestID string) *logrus.Entry {
	return l.WithField("request_id", requestID)
}

// WithUserID adds user ID to log context
func (l *Logger) WithUserID(userID string) *logrus.Entry {
	return l.WithField("user_id", userID)
}

// WithDuration adds duration to log context
func (l *Logger) WithDuration(duration interface{}) *logrus.Entry {
	return l.WithField("duration", duration)
}

// WithError adds error to log context
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// WithFields adds multiple fields to log context
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// LogHTTPRequest logs HTTP request information
func (l *Logger) LogHTTPRequest(method, path, userAgent, clientIP string, statusCode int, duration interface{}) {
	l.WithFields(logrus.Fields{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"user_agent":  userAgent,
		"client_ip":   clientIP,
		"duration":    duration,
		"type":        "http_request",
	}).Info("HTTP request processed")
}

// LogDatabaseQuery logs database query information
func (l *Logger) LogDatabaseQuery(query string, duration interface{}, rowsAffected int64) {
	l.WithFields(logrus.Fields{
		"query":         query,
		"duration":      duration,
		"rows_affected": rowsAffected,
		"type":          "database_query",
	}).Debug("Database query executed")
}

// LogCacheOperation logs cache operation information
func (l *Logger) LogCacheOperation(operation, key string, hit bool, duration interface{}) {
	l.WithFields(logrus.Fields{
		"operation": operation,
		"key":       key,
		"hit":       hit,
		"duration":  duration,
		"type":      "cache_operation",
	}).Debug("Cache operation performed")
}

// LogKafkaMessage logs Kafka message information
func (l *Logger) LogKafkaMessage(topic, operation string, messageSize int, duration interface{}) {
	l.WithFields(logrus.Fields{
		"topic":        topic,
		"operation":    operation,
		"message_size": messageSize,
		"duration":     duration,
		"type":         "kafka_message",
	}).Info("Kafka message processed")
}

// LogBusinessEvent logs business-related events
func (l *Logger) LogBusinessEvent(event, entity string, entityID interface{}, details map[string]interface{}) {
	fields := logrus.Fields{
		"event":     event,
		"entity":    entity,
		"entity_id": entityID,
		"type":      "business_event",
	}
	
	// Add details to fields
	for k, v := range details {
		fields[k] = v
	}
	
	l.WithFields(fields).Info("Business event occurred")
}

// LogSystemEvent logs system-related events
func (l *Logger) LogSystemEvent(event string, details map[string]interface{}) {
	fields := logrus.Fields{
		"event": event,
		"type":  "system_event",
	}
	
	// Add details to fields
	for k, v := range details {
		fields[k] = v
	}
	
	l.WithFields(fields).Info("System event occurred")
}