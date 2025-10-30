//go:build ignore
// +build ignore

package main

// This example shows how to integrate Reforge log level management with charmbracelet/log.
// Copy this code into your project and adapt as needed.
//
// Compatible with charmbracelet/log v0.2.0+
//
// To use: go get github.com/charmbracelet/log

import (
	"context"
	"io"
	"os"
	"time"

	reforge "github.com/ReforgeHQ/sdk-go"
	"github.com/charmbracelet/log"
)

// ReforgeLevelFunc provides dynamic log level control for charmbracelet/log
// based on Reforge configuration.
type ReforgeLevelFunc struct {
	client     reforge.ClientInterface
	loggerName string
}

// NewReforgeLevelFunc creates a new ReforgeLevelFunc that queries Reforge for log levels.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	levelFunc := NewReforgeLevelFunc(client, "com.example.myapp")
//	logger := log.NewWithOptions(os.Stdout, log.Options{
//	    Level: levelFunc.GetLevel(),
//	})
func NewReforgeLevelFunc(client reforge.ClientInterface, loggerName string) *ReforgeLevelFunc {
	return &ReforgeLevelFunc{
		client:     client,
		loggerName: loggerName,
	}
}

// GetLevel returns the current log level from Reforge configuration.
func (l *ReforgeLevelFunc) GetLevel() log.Level {
	reforgeLevel := l.client.GetLogLevel(l.loggerName)
	return l.reforgeToCharmLogLevel(reforgeLevel)
}

// reforgeToCharmLogLevel converts a Reforge LogLevel to charmbracelet log.Level
func (l *ReforgeLevelFunc) reforgeToCharmLogLevel(level reforge.LogLevel) log.Level {
	switch level {
	case reforge.Trace:
		return log.DebugLevel - 1 // Trace is more verbose than Debug
	case reforge.Debug:
		return log.DebugLevel
	case reforge.Info:
		return log.InfoLevel
	case reforge.Warn:
		return log.WarnLevel
	case reforge.Error:
		return log.ErrorLevel
	case reforge.Fatal:
		return log.FatalLevel
	default:
		return log.DebugLevel
	}
}

// ReforgeCharmLogger wraps a charmbracelet log.Logger and provides dynamic
// log level filtering based on Reforge configuration.
type ReforgeCharmLogger struct {
	logger     *log.Logger
	client     reforge.ClientInterface
	loggerName string
}

// NewReforgeCharmLogger creates a new ReforgeCharmLogger that wraps a charmbracelet logger.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	baseLogger := log.New(os.Stdout)
//	reforgeLogger := NewReforgeCharmLogger(client, baseLogger, "com.example.myapp")
//	reforgeLogger.Info("This is controlled by Reforge")
func NewReforgeCharmLogger(client reforge.ClientInterface, logger *log.Logger, loggerName string) *ReforgeCharmLogger {
	return &ReforgeCharmLogger{
		logger:     logger,
		client:     client,
		loggerName: loggerName,
	}
}

// isEnabled checks if a log level is enabled in Reforge config
func (l *ReforgeCharmLogger) isEnabled(level reforge.LogLevel) bool {
	configuredLevel := l.client.GetLogLevel(l.loggerName)
	return level >= configuredLevel
}

// Debug logs a debug message if enabled
func (l *ReforgeCharmLogger) Debug(msg interface{}, keyvals ...interface{}) {
	if l.isEnabled(reforge.Debug) {
		l.logger.Debug(msg, keyvals...)
	}
}

// Debugf logs a formatted debug message if enabled
func (l *ReforgeCharmLogger) Debugf(format string, args ...interface{}) {
	if l.isEnabled(reforge.Debug) {
		l.logger.Debugf(format, args...)
	}
}

// Info logs an info message if enabled
func (l *ReforgeCharmLogger) Info(msg interface{}, keyvals ...interface{}) {
	if l.isEnabled(reforge.Info) {
		l.logger.Info(msg, keyvals...)
	}
}

// Infof logs a formatted info message if enabled
func (l *ReforgeCharmLogger) Infof(format string, args ...interface{}) {
	if l.isEnabled(reforge.Info) {
		l.logger.Infof(format, args...)
	}
}

// Warn logs a warning message if enabled
func (l *ReforgeCharmLogger) Warn(msg interface{}, keyvals ...interface{}) {
	if l.isEnabled(reforge.Warn) {
		l.logger.Warn(msg, keyvals...)
	}
}

// Warnf logs a formatted warning message if enabled
func (l *ReforgeCharmLogger) Warnf(format string, args ...interface{}) {
	if l.isEnabled(reforge.Warn) {
		l.logger.Warnf(format, args...)
	}
}

// Error logs an error message if enabled
func (l *ReforgeCharmLogger) Error(msg interface{}, keyvals ...interface{}) {
	if l.isEnabled(reforge.Error) {
		l.logger.Error(msg, keyvals...)
	}
}

// Errorf logs a formatted error message if enabled
func (l *ReforgeCharmLogger) Errorf(format string, args ...interface{}) {
	if l.isEnabled(reforge.Error) {
		l.logger.Errorf(format, args...)
	}
}

// Fatal logs a fatal message (always logged)
func (l *ReforgeCharmLogger) Fatal(msg interface{}, keyvals ...interface{}) {
	l.logger.Fatal(msg, keyvals...)
}

// Fatalf logs a formatted fatal message (always logged)
func (l *ReforgeCharmLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// With returns a new ReforgeCharmLogger with additional context
func (l *ReforgeCharmLogger) With(keyvals ...interface{}) *ReforgeCharmLogger {
	return &ReforgeCharmLogger{
		logger:     l.logger.With(keyvals...),
		client:     l.client,
		loggerName: l.loggerName,
	}
}

// WithPrefix returns a new ReforgeCharmLogger with a prefix
func (l *ReforgeCharmLogger) WithPrefix(prefix string) *ReforgeCharmLogger {
	return &ReforgeCharmLogger{
		logger:     l.logger.WithPrefix(prefix),
		client:     l.client,
		loggerName: l.loggerName,
	}
}

// SetOutput sets the output destination
func (l *ReforgeCharmLogger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

// GetLogger returns the underlying logger
func (l *ReforgeCharmLogger) GetLogger() *log.Logger {
	return l.logger
}

// WithContext returns a new logger with the given context
func (l *ReforgeCharmLogger) WithContext(ctx context.Context) *ReforgeCharmLogger {
	return &ReforgeCharmLogger{
		logger:     log.FromContext(ctx),
		client:     l.client,
		loggerName: l.loggerName,
	}
}

// ReforgeAtomicLevel wraps a charmbracelet logger and provides automatic updates
// from Reforge configuration at regular intervals.
type ReforgeAtomicLevel struct {
	logger       *log.Logger
	client       reforge.ClientInterface
	loggerName   string
	stopChan     chan struct{}
	levelFunc    *ReforgeLevelFunc
}

// NewReforgeAtomicLevel creates a new atomic level that automatically updates
// from Reforge configuration at the specified interval.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	logger := log.New(os.Stdout)
//	atomicLevel := NewReforgeAtomicLevel(client, logger, "com.example.myapp", 30*time.Second)
//	defer atomicLevel.Stop()
func NewReforgeAtomicLevel(client reforge.ClientInterface, logger *log.Logger, loggerName string, updateInterval time.Duration) *ReforgeAtomicLevel {
	levelFunc := NewReforgeLevelFunc(client, loggerName)
	ral := &ReforgeAtomicLevel{
		logger:     logger,
		client:     client,
		loggerName: loggerName,
		stopChan:   make(chan struct{}),
		levelFunc:  levelFunc,
	}

	// Set initial level
	ral.updateLevel()

	// Start background updater
	go ral.backgroundUpdater(updateInterval)

	return ral
}

// Logger returns the underlying logger
func (r *ReforgeAtomicLevel) Logger() *log.Logger {
	return r.logger
}

// Stop stops the background level updater
func (r *ReforgeAtomicLevel) Stop() {
	close(r.stopChan)
}

// updateLevel fetches the current level from Reforge and updates the logger
func (r *ReforgeAtomicLevel) updateLevel() {
	level := r.levelFunc.GetLevel()
	r.logger.SetLevel(level)
}

// backgroundUpdater periodically updates the log level from Reforge
func (r *ReforgeAtomicLevel) backgroundUpdater(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.updateLevel()
		case <-r.stopChan:
			return
		}
	}
}

func main() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Approach 1: Using ReforgeLevelFunc with initial level setting
	// Update level periodically by calling SetLevel
	levelFunc := NewReforgeLevelFunc(client, "com.example.myapp")
	logger1 := log.NewWithOptions(os.Stdout, log.Options{
		Level:           levelFunc.GetLevel(),
		ReportTimestamp: true,
	})

	logger1.Debug("Debug message - controlled by Reforge")
	logger1.Info("Info message - controlled by Reforge")
	logger1.Error("Error message - controlled by Reforge")

	// You can update the level manually:
	// logger1.SetLevel(levelFunc.GetLevel())

	// Approach 2: Using ReforgeCharmLogger for automatic level checking
	// This checks the level on every log call (most dynamic)
	baseLogger := log.NewWithOptions(os.Stdout, log.Options{
		ReportTimestamp: true,
	})
	reforgeLogger := NewReforgeCharmLogger(client, baseLogger, "com.example.myapp")

	reforgeLogger.Debug("Debug message - checked dynamically")
	reforgeLogger.Info("Info message - checked dynamically")
	reforgeLogger.Error("Error message - checked dynamically")

	// Approach 3: Using ReforgeAtomicLevel with automatic periodic updates
	// This updates the level in the background (good balance of performance and dynamism)
	logger3 := log.NewWithOptions(os.Stdout, log.Options{
		ReportTimestamp: true,
	})
	atomicLevel := NewReforgeAtomicLevel(client, logger3, "com.example.myapp", 30*time.Second)
	defer atomicLevel.Stop()

	atomicLogger := atomicLevel.Logger()
	atomicLogger.Debug("Debug message - updates every 30 seconds")
	atomicLogger.Info("Info message - updates every 30 seconds")

	// Approach 4: Multiple loggers for different components
	dbLogger := log.NewWithOptions(os.Stdout, log.Options{
		Prefix:          "database",
		ReportTimestamp: true,
	})
	reforgeDbLogger := NewReforgeCharmLogger(client, dbLogger, "com.example.database")

	apiLogger := log.NewWithOptions(os.Stdout, log.Options{
		Prefix:          "api",
		ReportTimestamp: true,
	})
	reforgeApiLogger := NewReforgeCharmLogger(client, apiLogger, "com.example.api")

	reforgeDbLogger.Debug("Database query executed", "duration_ms", 42, "rows", 100)
	reforgeApiLogger.Info("API request received", "method", "GET", "path", "/api/users")

	// Approach 5: Using With for structured logging
	requestLogger := reforgeLogger.With("request_id", "abc-123", "user_id", "user-456")
	requestLogger.Info("Processing request", "endpoint", "/api/data")

	// Approach 6: Using WithPrefix for logger hierarchies
	serviceLogger := reforgeLogger.WithPrefix("payment-service")
	serviceLogger.Info("Payment processed", "amount", 99.99, "currency", "USD")
}

/* Configuration in Reforge:

Create a LOG_LEVEL_V2 config with key "log-levels.default":

{
  "rows": [
    {
      "values": [
        {
          "criteria": [
            {
              "operator": "PROP_IS_ONE_OF",
              "property_name": "reforge-sdk-logging.logger-path",
              "value_to_match": {
                "string_list": {
                  "values": ["com.example.myapp"]
                }
              }
            }
          ],
          "value": { "log_level": "DEBUG" }
        },
        {
          "criteria": [
            {
              "operator": "PROP_IS_ONE_OF",
              "property_name": "reforge-sdk-logging.logger-path",
              "value_to_match": {
                "string_list": {
                  "values": ["com.example.database"]
                }
              }
            }
          ],
          "value": { "log_level": "INFO" }
        },
        {
          "criteria": [
            {
              "operator": "PROP_IS_ONE_OF",
              "property_name": "reforge-sdk-logging.logger-path",
              "value_to_match": {
                "string_list": {
                  "values": ["com.example.api"]
                }
              }
            }
          ],
          "value": { "log_level": "WARN" }
        }
      ],
      "value": { "log_level": "INFO" }
    }
  ]
}

This config will:
- Set DEBUG level for "com.example.myapp"
- Set INFO level for "com.example.database"
- Set WARN level for "com.example.api"
- Default to INFO for all other loggers

You can dynamically change these levels in Reforge without redeploying your application!

Performance Notes:
- ReforgeLevelFunc: Set level once, update manually or periodically (best performance)
- ReforgeCharmLogger: Checks on every log call (most dynamic, slight overhead)
- ReforgeAtomicLevel: Updates periodically in background (good balance)

Choose the approach that best fits your needs:
- Use ReforgeLevelFunc if you want manual control or infrequent updates
- Use ReforgeCharmLogger if you need instant level changes (e.g., debugging production)
- Use ReforgeAtomicLevel for automatic updates with minimal overhead
*/