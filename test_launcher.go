//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"time"

	reforge "github.com/ReforgeHQ/sdk-go"
	"github.com/charmbracelet/log"
)

// ReforgeCharmLogger wraps a charmbracelet log.Logger and provides dynamic
// log level filtering based on Reforge configuration - calls GetLogLevel on every log call
type ReforgeCharmLogger struct {
	logger     *log.Logger
	client     reforge.ClientInterface
	loggerName string
}

// NewReforgeCharmLogger creates a new ReforgeCharmLogger
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

// Info logs an info message if enabled
func (l *ReforgeCharmLogger) Info(msg interface{}, keyvals ...interface{}) {
	if l.isEnabled(reforge.Info) {
		l.logger.Info(msg, keyvals...)
	}
}

// Warn logs a warning message if enabled
func (l *ReforgeCharmLogger) Warn(msg interface{}, keyvals ...interface{}) {
	if l.isEnabled(reforge.Warn) {
		l.logger.Warn(msg, keyvals...)
	}
}

// Error logs an error message if enabled
func (l *ReforgeCharmLogger) Error(msg interface{}, keyvals ...interface{}) {
	if l.isEnabled(reforge.Error) {
		l.logger.Error(msg, keyvals...)
	}
}

func main() {
	fmt.Println("=== Reforge Dynamic Log Level Test Launcher ===")
	fmt.Println("SDK Key: 1302-Staging-...")
	fmt.Println("Logger name: test.launcher")
	fmt.Println("Checks Reforge for log level on EVERY log call")
	fmt.Println("Logging every 10 seconds at DEBUG, INFO, WARN, and ERROR levels")
	fmt.Println("")

	// Initialize Reforge SDK with staging key
	client, err := reforge.NewSdk(
		reforge.WithSdkKey("1302-Staging-69151316-e520-4116-9c71-802d77c3f7eb-backend-d680ab42-892b-4e68-a9bd-604546307e02"),
	)
	if err != nil {
		panic(err)
	}

	// Create charmbracelet logger - set to DebugLevel so nothing is filtered at logger level
	baseLogger := log.NewWithOptions(os.Stdout, log.Options{
		ReportTimestamp: true,
		ReportCaller:    false,
		Level:           log.DebugLevel,
	})

	// Wrap with Reforge integration that checks on every call
	logger := NewReforgeCharmLogger(client, baseLogger, "test.launcher")

	fmt.Println("Logger initialized. Starting log loop...")
	fmt.Println("")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	counter := 0
	for range ticker.C {
		counter++

		// Get current log level from Reforge (for display purposes)
		currentLevel := client.GetLogLevel("test.launcher")
		fmt.Printf("\n[Loop %d] Current configured level: %s\n", counter, currentLevel.String())

		logger.Debug("🔍 DEBUG message", "counter", counter, "level", "DEBUG")
		logger.Info("ℹ️  INFO message", "counter", counter, "level", "INFO")
		logger.Warn("⚠️  WARN message", "counter", counter, "level", "WARN")
		logger.Error("❌ ERROR message", "counter", counter, "level", "ERROR")

		fmt.Println("---")
	}
}
