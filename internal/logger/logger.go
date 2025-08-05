package logger

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.SugaredLogger
	once         sync.Once
	mu           sync.RWMutex
)

// Init initializes the global logger instance with proper error handling
func Init(logLevel string) error {
	var err error
	once.Do(func() {
		config := zap.NewProductionConfig()

		// Parse log level
		level, err := zapcore.ParseLevel(logLevel)
		if err != nil {
			err = fmt.Errorf("invalid log level %s: %w", logLevel, err)
			return
		}
		config.Level = zap.NewAtomicLevelAt(level)

		// Configure structured logging
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.LevelKey = "level"
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.CallerKey = "caller"
		config.EncoderConfig.StacktraceKey = "stacktrace"

		// Create logger
		zapLogger, loggerErr := config.Build()
		if loggerErr != nil {
			err = fmt.Errorf("failed to create logger: %w", loggerErr)
			return
		}

		mu.Lock()
		globalLogger = zapLogger.Sugar()
		mu.Unlock()
	})

	return err
}

// Get returns the global logger instance with thread safety
func Get() *zap.SugaredLogger {
	mu.RLock()
	defer mu.RUnlock()

	if globalLogger == nil {
		// Fallback to production logger if not initialized
		zapLogger, err := zap.NewProduction()
		if err != nil {
			// Last resort - create a basic logger
			zapLogger, _ = zap.NewDevelopment()
		}
		globalLogger = zapLogger.Sugar()
	}
	return globalLogger
}

// Sync flushes any buffered log entries
func Sync() error {
	mu.RLock()
	defer mu.RUnlock()

	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// SetLogger allows setting a custom logger (mainly for testing)
func SetLogger(logger *zap.SugaredLogger) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger = logger
}

// RequestLogger provides structured logging for HTTP/gRPC requests
type RequestLogger struct {
	logger *zap.SugaredLogger
	start  time.Time
	fields map[string]interface{}
}

// NewRequestLogger creates a new request logger with tracing
func NewRequestLogger(method, path string) *RequestLogger {
	return &RequestLogger{
		logger: Get(),
		start:  time.Now(),
		fields: map[string]interface{}{
			"method":   method,
			"path":     path,
			"trace_id": generateTraceID(),
		},
	}
}

// AddField adds a field to the request log
func (rl *RequestLogger) AddField(key string, value interface{}) {
	rl.fields[key] = value
}

// LogRequest logs the start of a request
func (rl *RequestLogger) LogRequest() {
	rl.logger.Infow("Request started", rl.getFields()...)
}

// LogResponse logs the completion of a request with metrics
func (rl *RequestLogger) LogResponse(statusCode int, err error) {
	duration := time.Since(rl.start)

	fields := rl.getFields()
	fields = append(fields,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
		"duration_ns", duration.Nanoseconds(),
	)

	if err != nil {
		fields = append(fields, "error", err.Error())
		rl.logger.Errorw("Request failed", fields...)
	} else {
		rl.logger.Infow("Request completed", fields...)
	}
}

// getFields converts the fields map to key-value pairs
func (rl *RequestLogger) getFields() []interface{} {
	fields := make([]interface{}, 0, len(rl.fields)*2)
	for k, v := range rl.fields {
		fields = append(fields, k, v)
	}
	return fields
}

// generateTraceID creates a simple trace ID for request tracing
func generateTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}

// MetricsLogger provides basic metrics logging
type MetricsLogger struct {
	logger *zap.SugaredLogger
}

// NewMetricsLogger creates a new metrics logger
func NewMetricsLogger() *MetricsLogger {
	return &MetricsLogger{
		logger: Get(),
	}
}

// LogMetric logs a metric with structured data
func (ml *MetricsLogger) LogMetric(metricName string, value float64, tags map[string]string) {
	fields := []interface{}{
		"metric_name", metricName,
		"value", value,
		"timestamp", time.Now().Unix(),
	}

	for k, v := range tags {
		fields = append(fields, k, v)
	}

	ml.logger.Infow("Metric recorded", fields...)
}

// LogCounter logs a counter metric
func (ml *MetricsLogger) LogCounter(counterName string, increment int64, tags map[string]string) {
	ml.LogMetric(counterName, float64(increment), tags)
}

// LogHistogram logs a histogram metric
func (ml *MetricsLogger) LogHistogram(histogramName string, value float64, tags map[string]string) {
	ml.LogMetric(histogramName, value, tags)
}

// LogGauge logs a gauge metric
func (ml *MetricsLogger) LogGauge(gaugeName string, value float64, tags map[string]string) {
	ml.LogMetric(gaugeName, value, tags)
}
