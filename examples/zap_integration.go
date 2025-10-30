//go:build ignore
// +build ignore

package main

// This example shows how to integrate Reforge log level management with zap.
// Copy this code into your project and adapt as needed.
//
// Compatible with zap v1.10.0+ and uber-go/zap
//
// To use: go get go.uber.org/zap

import (
	"os"
	"time"

	reforge "github.com/ReforgeHQ/sdk-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ReforgeZapLevel is a dynamic zap level that queries Reforge for the
// appropriate log level. It implements zapcore.LevelEnabler.
type ReforgeZapLevel struct {
	client     reforge.ClientInterface
	loggerName string
}

// NewReforgeZapLevel creates a new dynamic zap level that integrates with Reforge.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	dynamicLevel := NewReforgeZapLevel(client, "com.example.myapp")
//	logger, _ := zap.NewProduction(zap.IncreaseLevel(dynamicLevel))
func NewReforgeZapLevel(client reforge.ClientInterface, loggerName string) *ReforgeZapLevel {
	return &ReforgeZapLevel{
		client:     client,
		loggerName: loggerName,
	}
}

// Enabled implements zapcore.LevelEnabler
func (l *ReforgeZapLevel) Enabled(level zapcore.Level) bool {
	reforgeLevel := l.client.GetLogLevel(l.loggerName)
	zapLevel := l.reforgeToZapLevel(reforgeLevel)
	return level >= zapLevel
}

// reforgeToZapLevel converts a Reforge LogLevel to zapcore.Level
func (l *ReforgeZapLevel) reforgeToZapLevel(level reforge.LogLevel) zapcore.Level {
	switch level {
	case reforge.Trace:
		return zapcore.DebugLevel - 1 // Lower than debug
	case reforge.Debug:
		return zapcore.DebugLevel
	case reforge.Info:
		return zapcore.InfoLevel
	case reforge.Warn:
		return zapcore.WarnLevel
	case reforge.Error:
		return zapcore.ErrorLevel
	case reforge.Fatal:
		return zapcore.FatalLevel
	default:
		return zapcore.DebugLevel
	}
}

// ReforgeAtomicLevel wraps zap.AtomicLevel and provides automatic updates
// from Reforge configuration.
type ReforgeAtomicLevel struct {
	client     reforge.ClientInterface
	loggerName string
	atomic     zap.AtomicLevel
	stopChan   chan struct{}
}

// NewReforgeAtomicLevel creates a new atomic level that automatically updates
// from Reforge configuration at the specified interval.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	atomicLevel := NewReforgeAtomicLevel(client, "com.example.myapp", 30*time.Second)
//	defer atomicLevel.Stop()
//
//	config := zap.NewProductionConfig()
//	config.Level = atomicLevel.AtomicLevel()
//	logger, _ := config.Build()
func NewReforgeAtomicLevel(client reforge.ClientInterface, loggerName string, updateInterval time.Duration) *ReforgeAtomicLevel {
	atomic := zap.NewAtomicLevel()
	ral := &ReforgeAtomicLevel{
		client:     client,
		loggerName: loggerName,
		atomic:     atomic,
		stopChan:   make(chan struct{}),
	}

	// Set initial level
	ral.updateLevel()

	// Start background updater
	go ral.backgroundUpdater(updateInterval)

	return ral
}

// AtomicLevel returns the underlying zap.AtomicLevel
func (r *ReforgeAtomicLevel) AtomicLevel() zap.AtomicLevel {
	return r.atomic
}

// Stop stops the background level updater
func (r *ReforgeAtomicLevel) Stop() {
	close(r.stopChan)
}

// updateLevel fetches the current level from Reforge and updates the atomic level
func (r *ReforgeAtomicLevel) updateLevel() {
	reforgeLevel := r.client.GetLogLevel(r.loggerName)
	zapLevel := r.reforgeToZapLevel(reforgeLevel)
	r.atomic.SetLevel(zapLevel)
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

// reforgeToZapLevel converts a Reforge LogLevel to zapcore.Level
func (r *ReforgeAtomicLevel) reforgeToZapLevel(level reforge.LogLevel) zapcore.Level {
	switch level {
	case reforge.Trace:
		return zapcore.DebugLevel - 1
	case reforge.Debug:
		return zapcore.DebugLevel
	case reforge.Info:
		return zapcore.InfoLevel
	case reforge.Warn:
		return zapcore.WarnLevel
	case reforge.Error:
		return zapcore.ErrorLevel
	case reforge.Fatal:
		return zapcore.FatalLevel
	default:
		return zapcore.DebugLevel
	}
}

// ReforgeZapCore is a custom zapcore.Core that wraps another core and
// provides dynamic level filtering based on Reforge configuration.
type ReforgeZapCore struct {
	zapcore.Core
	client     reforge.ClientInterface
	loggerName string
}

// NewReforgeZapCore creates a new core that wraps another core with Reforge
// dynamic level filtering.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
//	core := NewReforgeZapCore(
//	    zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
//	    client,
//	    "com.example.myapp",
//	)
//	logger := zap.New(core)
func NewReforgeZapCore(core zapcore.Core, client reforge.ClientInterface, loggerName string) *ReforgeZapCore {
	return &ReforgeZapCore{
		Core:       core,
		client:     client,
		loggerName: loggerName,
	}
}

// Enabled returns true if the given level is at or above the configured level.
func (c *ReforgeZapCore) Enabled(level zapcore.Level) bool {
	reforgeLevel := c.client.GetLogLevel(c.loggerName)
	zapLevel := c.reforgeToZapLevel(reforgeLevel)
	return level >= zapLevel && c.Core.Enabled(level)
}

// Check determines whether the supplied Entry should be logged. If so, it
// adds the Entry to the Core. Returns nil if the entry should not be logged.
func (c *ReforgeZapCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(entry.Level) {
		return checked
	}
	return c.Core.Check(entry, checked)
}

// With adds structured context to the Core.
func (c *ReforgeZapCore) With(fields []zapcore.Field) zapcore.Core {
	return &ReforgeZapCore{
		Core:       c.Core.With(fields),
		client:     c.client,
		loggerName: c.loggerName,
	}
}

// reforgeToZapLevel converts a Reforge LogLevel to zapcore.Level
func (c *ReforgeZapCore) reforgeToZapLevel(level reforge.LogLevel) zapcore.Level {
	switch level {
	case reforge.Trace:
		return zapcore.DebugLevel - 1
	case reforge.Debug:
		return zapcore.DebugLevel
	case reforge.Info:
		return zapcore.InfoLevel
	case reforge.Warn:
		return zapcore.WarnLevel
	case reforge.Error:
		return zapcore.ErrorLevel
	case reforge.Fatal:
		return zapcore.FatalLevel
	default:
		return zapcore.DebugLevel
	}
}

func main() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Approach 1: Using ReforgeZapLevel with IncreaseLevel
	// This checks Reforge on every log call (most dynamic, slight performance cost)
	dynamicLevel := NewReforgeZapLevel(client, "com.example.myapp")
	logger1, _ := zap.NewProduction(zap.IncreaseLevel(dynamicLevel))
	defer logger1.Sync()

	logger1.Debug("Debug message - controlled by Reforge")
	logger1.Info("Info message - controlled by Reforge")
	logger1.Error("Error message - controlled by Reforge")

	// Approach 2: Using ReforgeAtomicLevel with automatic updates
	// This updates the level periodically (good balance of performance and dynamism)
	atomicLevel := NewReforgeAtomicLevel(client, "com.example.myapp", 30*time.Second)
	defer atomicLevel.Stop()

	config := zap.NewProductionConfig()
	config.Level = atomicLevel.AtomicLevel()
	logger2, _ := config.Build()
	defer logger2.Sync()

	logger2.Debug("Debug message")
	logger2.Info("Info message")

	// Approach 3: Using ReforgeZapCore for fine-grained control
	// This wraps the core directly
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	baseCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
	reforgeCore := NewReforgeZapCore(baseCore, client, "com.example.myapp")
	logger3 := zap.New(reforgeCore)
	defer logger3.Sync()

	logger3.Debug("Debug message from custom core")
	logger3.Info("Info message from custom core")

	// Approach 4: Multiple loggers for different components
	dbLevel := NewReforgeZapLevel(client, "com.example.database")
	dbLogger, _ := zap.NewProduction(
		zap.IncreaseLevel(dbLevel),
		zap.Fields(zap.String("component", "database")),
	)
	defer dbLogger.Sync()

	apiLevel := NewReforgeZapLevel(client, "com.example.api")
	apiLogger, _ := zap.NewProduction(
		zap.IncreaseLevel(apiLevel),
		zap.Fields(zap.String("component", "api")),
	)
	defer apiLogger.Sync()

	dbLogger.Debug("Database query executed", zap.Duration("duration", 42*time.Millisecond))
	apiLogger.Info("API request received", zap.String("method", "GET"))

	// You can also manually trigger level updates:
	// atomicLevel.updateLevel()
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

Change these levels in Reforge to dynamically control your application's logging!

Performance Notes:
- ReforgeZapLevel: Checks on every log call (most dynamic, slight overhead)
- ReforgeAtomicLevel: Updates periodically (good balance)
- ReforgeZapCore: Checks on every log call, more control over core behavior
*/
