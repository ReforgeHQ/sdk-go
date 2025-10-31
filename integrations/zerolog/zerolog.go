package zerolog

import (
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
