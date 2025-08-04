package logger

import (
	"go.uber.org/zap"
)

var globalLogger *zap.SugaredLogger

// Init initializes the global logger instance
func Init() {
	zapLogger, _ := zap.NewProduction()
	globalLogger = zapLogger.Sugar()
}

// Get returns the global logger instance
func Get() *zap.SugaredLogger {
	if globalLogger == nil {
		zapLogger, _ := zap.NewProduction()
		globalLogger = zapLogger.Sugar()
	}
	return globalLogger
}
