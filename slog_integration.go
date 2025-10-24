package reforge

import (
	"context"
	"log/slog"
)

// ToSlogLevel converts a Reforge LogLevel to a slog.Level
func (l LogLevel) ToSlogLevel() slog.Level {
	switch l {
	case Trace:
		return slog.LevelDebug - 4 // Trace is more verbose than Debug
	case Debug:
		return slog.LevelDebug
	case Info:
		return slog.LevelInfo
	case Warn:
		return slog.LevelWarn
	case Error:
		return slog.LevelError
	case Fatal:
		return slog.LevelError + 4 // Fatal is more severe than Error
	default:
		return slog.LevelDebug
	}
}

// ReforgeHandler is a slog.Handler that dynamically determines log levels
// based on Reforge configuration. It wraps another handler and filters
// log records based on the log level configured in Reforge.
type ReforgeHandler struct {
	client       ClientInterface
	wrappedHandler slog.Handler
	loggerName   string
}

// NewReforgeHandler creates a new ReforgeHandler that wraps the provided handler.
// The loggerName is used to look up the log level in Reforge configuration.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	handler := reforge.NewReforgeHandler(client, slog.NewJSONHandler(os.Stdout, nil), "myapp")
//	logger := slog.New(handler)
func NewReforgeHandler(client ClientInterface, wrappedHandler slog.Handler, loggerName string) *ReforgeHandler {
	return &ReforgeHandler{
		client:         client,
		wrappedHandler: wrappedHandler,
		loggerName:     loggerName,
	}
}

// Enabled reports whether the handler handles records at the given level.
// It checks the Reforge configuration for the appropriate log level.
func (h *ReforgeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	reforgeLevel := h.client.GetLogLevel(h.loggerName)
	return level >= reforgeLevel.ToSlogLevel()
}

// Handle handles the Record.
func (h *ReforgeHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.wrappedHandler.Handle(ctx, r)
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
func (h *ReforgeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ReforgeHandler{
		client:         h.client,
		wrappedHandler: h.wrappedHandler.WithAttrs(attrs),
		loggerName:     h.loggerName,
	}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
func (h *ReforgeHandler) WithGroup(name string) slog.Handler {
	return &ReforgeHandler{
		client:         h.client,
		wrappedHandler: h.wrappedHandler.WithGroup(name),
		loggerName:     h.loggerName,
	}
}

// ReforgeLeveler is a slog.Leveler that dynamically determines the log level
// based on Reforge configuration.
type ReforgeLeveler struct {
	client     ClientInterface
	loggerName string
}

// NewReforgeLeveler creates a new ReforgeLeveler that queries Reforge for log levels.
//
// Example:
//
//	client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
//	leveler := reforge.NewReforgeLeveler(client, "myapp")
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: leveler}))
func NewReforgeLeveler(client ClientInterface, loggerName string) *ReforgeLeveler {
	return &ReforgeLeveler{
		client:     client,
		loggerName: loggerName,
	}
}

// Level returns the current log level from Reforge configuration.
func (l *ReforgeLeveler) Level() slog.Level {
	reforgeLevel := l.client.GetLogLevel(l.loggerName)
	return reforgeLevel.ToSlogLevel()
}
