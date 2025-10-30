package zerolog

import (
	"io"

	reforge "github.com/ReforgeHQ/sdk-go"
	"github.com/rs/zerolog"
)

// ReforgeZerologHook is a zerolog Hook that filters log events based on
// Reforge-configured log levels. It queries Reforge for the log level
// dynamically on each log event.
type ReforgeZerologHook struct {
	client     reforge.ClientInterface
	loggerName string
}

// NewReforgeZerologHook creates a new hook that integrates Reforge log level
// management with zerolog.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	hook := zerolog.NewReforgeZerologHook(client, "com.example.myapp")
//	logger := zerolog.New(os.Stdout).Hook(hook)
func NewReforgeZerologHook(client reforge.ClientInterface, loggerName string) *ReforgeZerologHook {
	return &ReforgeZerologHook{
		client:     client,
		loggerName: loggerName,
	}
}

// Run implements zerolog.Hook interface
func (h *ReforgeZerologHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	reforgeLevel := h.client.GetLogLevel(h.loggerName)
	zerologLevel := h.reforgeToZerologLevel(reforgeLevel)

	// If the event level is less severe than configured level, disable it
	if level < zerologLevel {
		e.Discard()
	}
}

// reforgeToZerologLevel converts a Reforge LogLevel to zerolog.Level
func (h *ReforgeZerologHook) reforgeToZerologLevel(level reforge.LogLevel) zerolog.Level {
	switch level {
	case reforge.Trace:
		return zerolog.TraceLevel
	case reforge.Debug:
		return zerolog.DebugLevel
	case reforge.Info:
		return zerolog.InfoLevel
	case reforge.Warn:
		return zerolog.WarnLevel
	case reforge.Error:
		return zerolog.ErrorLevel
	case reforge.Fatal:
		return zerolog.FatalLevel
	default:
		return zerolog.DebugLevel
	}
}

// ReforgeZerologLevelWriter wraps a zerolog.Logger and provides dynamic
// level filtering at the logger level (more efficient than hook-based filtering).
type ReforgeZerologLevelWriter struct {
	client     reforge.ClientInterface
	loggerName string
	writer     io.Writer
}

// NewReforgeZerologLevelWriter creates a writer that can be used with zerolog
// to provide dynamic log level control.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	levelWriter := zerolog.NewReforgeZerologLevelWriter(client, "com.example.myapp", os.Stdout)
//	logger := zerolog.New(levelWriter)
func NewReforgeZerologLevelWriter(client reforge.ClientInterface, loggerName string, writer io.Writer) *ReforgeZerologLevelWriter {
	return &ReforgeZerologLevelWriter{
		client:     client,
		loggerName: loggerName,
		writer:     writer,
	}
}

// Write implements io.Writer
func (w *ReforgeZerologLevelWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

// GetDynamicLevel returns the current log level from Reforge configuration.
// Call this to set the logger's level dynamically.
func (w *ReforgeZerologLevelWriter) GetDynamicLevel() zerolog.Level {
	reforgeLevel := w.client.GetLogLevel(w.loggerName)
	return w.reforgeToZerologLevel(reforgeLevel)
}

// reforgeToZerologLevel converts a Reforge LogLevel to zerolog.Level
func (w *ReforgeZerologLevelWriter) reforgeToZerologLevel(level reforge.LogLevel) zerolog.Level {
	switch level {
	case reforge.Trace:
		return zerolog.TraceLevel
	case reforge.Debug:
		return zerolog.DebugLevel
	case reforge.Info:
		return zerolog.InfoLevel
	case reforge.Warn:
		return zerolog.WarnLevel
	case reforge.Error:
		return zerolog.ErrorLevel
	case reforge.Fatal:
		return zerolog.FatalLevel
	default:
		return zerolog.DebugLevel
	}
}
