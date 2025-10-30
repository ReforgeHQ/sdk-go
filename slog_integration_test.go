package reforge

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToSlogLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         LogLevel
		expectedLevel slog.Level
	}{
		{"Trace converts to Debug-4", Trace, slog.LevelDebug - 4},
		{"Debug converts to Debug", Debug, slog.LevelDebug},
		{"Info converts to Info", Info, slog.LevelInfo},
		{"Warn converts to Warn", Warn, slog.LevelWarn},
		{"Error converts to Error", Error, slog.LevelError},
		{"Fatal converts to Error+4", Fatal, slog.LevelError + 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.ToSlogLevel()
			assert.Equal(t, tt.expectedLevel, result)
		})
	}
}

func TestReforgeHandler_Enabled(t *testing.T) {
	client, err := NewSdk(WithOfflineSources([]string{
		"datafile://testdata/loglevel_test.json",
	}))
	require.NoError(t, err)

	tests := []struct {
		name         string
		loggerName   string
		slogLevel    slog.Level
		shouldEnable bool
	}{
		{
			name:         "Debug logger allows debug messages",
			loggerName:   "com.example.debug",
			slogLevel:    slog.LevelDebug,
			shouldEnable: true,
		},
		{
			name:         "Debug logger allows info messages",
			loggerName:   "com.example.debug",
			slogLevel:    slog.LevelInfo,
			shouldEnable: true,
		},
		{
			name:         "Error logger blocks info messages",
			loggerName:   "com.example.error",
			slogLevel:    slog.LevelInfo,
			shouldEnable: false,
		},
		{
			name:         "Error logger allows error messages",
			loggerName:   "com.example.error",
			slogLevel:    slog.LevelError,
			shouldEnable: true,
		},
		{
			name:         "Info logger blocks debug messages",
			loggerName:   "com.example.unknown",
			slogLevel:    slog.LevelDebug,
			shouldEnable: false,
		},
		{
			name:         "Info logger allows warn messages",
			loggerName:   "com.example.unknown",
			slogLevel:    slog.LevelWarn,
			shouldEnable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewReforgeHandler(client, slog.NewTextHandler(&bytes.Buffer{}, nil), tt.loggerName)
			enabled := handler.Enabled(context.Background(), tt.slogLevel)
			assert.Equal(t, tt.shouldEnable, enabled)
		})
	}
}

func TestReforgeHandler_ActualLogging(t *testing.T) {
	client, err := NewSdk(WithOfflineSources([]string{
		"datafile://testdata/loglevel_test.json",
	}))
	require.NoError(t, err)

	tests := []struct {
		name            string
		loggerName      string
		logFunc         func(*slog.Logger)
		shouldHaveOutput bool
	}{
		{
			name:       "Debug logger logs debug message",
			loggerName: "com.example.debug",
			logFunc: func(l *slog.Logger) {
				l.Debug("test debug message")
			},
			shouldHaveOutput: true,
		},
		{
			name:       "Error logger does not log debug message",
			loggerName: "com.example.error",
			logFunc: func(l *slog.Logger) {
				l.Debug("test debug message")
			},
			shouldHaveOutput: false,
		},
		{
			name:       "Error logger logs error message",
			loggerName: "com.example.error",
			logFunc: func(l *slog.Logger) {
				l.Error("test error message")
			},
			shouldHaveOutput: true,
		},
		{
			name:       "Info logger does not log debug",
			loggerName: "com.example.unknown",
			logFunc: func(l *slog.Logger) {
				l.Debug("test debug message")
			},
			shouldHaveOutput: false,
		},
		{
			name:       "Info logger logs info",
			loggerName: "com.example.unknown",
			logFunc: func(l *slog.Logger) {
				l.Info("test info message")
			},
			shouldHaveOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := NewReforgeHandler(client, slog.NewTextHandler(&buf, nil), tt.loggerName)
			logger := slog.New(handler)

			tt.logFunc(logger)

			output := buf.String()
			if tt.shouldHaveOutput {
				assert.NotEmpty(t, output, "Expected log output but got none")
				assert.Contains(t, output, "test")
			} else {
				assert.Empty(t, output, "Expected no log output but got: %s", output)
			}
		})
	}
}

func TestReforgeLeveler(t *testing.T) {
	client, err := NewSdk(WithOfflineSources([]string{
		"datafile://testdata/loglevel_test.json",
	}))
	require.NoError(t, err)

	tests := []struct {
		name          string
		loggerName    string
		expectedLevel slog.Level
	}{
		{
			name:          "Debug logger returns Debug level",
			loggerName:    "com.example.debug",
			expectedLevel: slog.LevelDebug,
		},
		{
			name:          "Error logger returns Error level",
			loggerName:    "com.example.error",
			expectedLevel: slog.LevelError,
		},
		{
			name:          "Unknown logger returns Info level (default)",
			loggerName:    "com.example.unknown",
			expectedLevel: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leveler := NewReforgeLeveler(client, tt.loggerName)
			level := leveler.Level()
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

func TestReforgeLeveler_WithHandlerOptions(t *testing.T) {
	client, err := NewSdk(WithOfflineSources([]string{
		"datafile://testdata/loglevel_test.json",
	}))
	require.NoError(t, err)

	// Test with error logger - should only log error and above
	var buf bytes.Buffer
	leveler := NewReforgeLeveler(client, "com.example.error")
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: leveler})
	logger := slog.New(handler)

	// These should not log
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")

	// This should log
	logger.Error("error message")

	output := buf.String()
	assert.NotContains(t, output, "debug message")
	assert.NotContains(t, output, "info message")
	assert.NotContains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestReforgeHandler_WithAttrs(t *testing.T) {
	client, err := NewSdk(WithOfflineSources([]string{
		"datafile://testdata/loglevel_test.json",
	}))
	require.NoError(t, err)

	var buf bytes.Buffer
	baseHandler := slog.NewJSONHandler(&buf, nil)
	handler := NewReforgeHandler(client, baseHandler, "com.example.debug")

	// Add attributes
	handlerWithAttrs := handler.WithAttrs([]slog.Attr{
		slog.String("service", "test-service"),
		slog.Int("version", 1),
	})

	logger := slog.New(handlerWithAttrs)
	logger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "service")
	assert.Contains(t, output, "test-service")
	assert.Contains(t, output, "version")
}

func TestReforgeHandler_WithGroup(t *testing.T) {
	client, err := NewSdk(WithOfflineSources([]string{
		"datafile://testdata/loglevel_test.json",
	}))
	require.NoError(t, err)

	var buf bytes.Buffer
	baseHandler := slog.NewJSONHandler(&buf, nil)
	handler := NewReforgeHandler(client, baseHandler, "com.example.debug")

	// Add group
	handlerWithGroup := handler.WithGroup("request")

	logger := slog.New(handlerWithGroup)
	logger.Info("test message", "method", "GET", "path", "/api/test")

	output := buf.String()
	assert.Contains(t, output, "request")
	// In grouped output, the attributes should be nested under "request"
	lines := strings.Split(output, "\n")
	assert.Greater(t, len(lines), 0, "Expected at least one line of output")
}
