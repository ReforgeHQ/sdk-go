package zap

import (
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
//	dynamicLevel := zap.NewReforgeZapLevel(client, "com.example.myapp")
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
//	atomicLevel := zap.NewReforgeAtomicLevel(client, "com.example.myapp", 30*time.Second)
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
//	core := zap.NewReforgeZapCore(
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
