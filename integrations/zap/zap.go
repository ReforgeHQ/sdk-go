package zap

import (
	reforge "github.com/ReforgeHQ/sdk-go"
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
