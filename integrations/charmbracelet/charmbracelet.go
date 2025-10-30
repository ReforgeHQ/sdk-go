package charmbracelet

import (
	"context"
	"io"
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
//	levelFunc := charmbracelet.NewReforgeLevelFunc(client, "com.example.myapp")
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
//	reforgeLogger := charmbracelet.NewReforgeCharmLogger(client, baseLogger, "com.example.myapp")
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
	logger    *log.Logger
	client    reforge.ClientInterface
	loggerName string
	stopChan  chan struct{}
	levelFunc *ReforgeLevelFunc
}

// NewReforgeAtomicLevel creates a new atomic level that automatically updates
// from Reforge configuration at the specified interval.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	logger := log.New(os.Stdout)
//	atomicLevel := charmbracelet.NewReforgeAtomicLevel(client, logger, "com.example.myapp", 30*time.Second)
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
