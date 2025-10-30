package reforge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLogLevel(t *testing.T) {
	client, err := NewSdk(WithOfflineSources([]string{
		"datafile://testdata/loglevel_test.json",
	}))
	require.NoError(t, err)

	tests := []struct {
		name           string
		loggerName     string
		expectedLevel  LogLevel
		expectedString string
	}{
		{
			name:           "debug logger gets DEBUG level",
			loggerName:     "com.example.debug",
			expectedLevel:  Debug,
			expectedString: "DEBUG",
		},
		{
			name:           "error logger gets ERROR level",
			loggerName:     "com.example.error",
			expectedLevel:  Error,
			expectedString: "ERROR",
		},
		{
			name:           "unknown logger gets default INFO level",
			loggerName:     "com.example.unknown",
			expectedLevel:  Info,
			expectedString: "INFO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := client.GetLogLevel(tt.loggerName)
			assert.Equal(t, tt.expectedLevel, level)
			assert.Equal(t, tt.expectedString, level.String())
		})
	}
}

func TestGetLogLevel_NoConfig(t *testing.T) {
	client, err := NewSdk(WithOfflineSources([]string{
		"datafile://testdata/loglevel_noconfig.json",
	}))
	require.NoError(t, err)

	// Should return Debug (default) when config doesn't exist
	level := client.GetLogLevel("any.logger")
	assert.Equal(t, Debug, level)
}

func TestGetLogLevel_CustomLoggerKey(t *testing.T) {
	client, err := NewSdk(
		WithOfflineSources([]string{
			"datafile://testdata/loglevel_custom_key.json",
		}),
		WithLoggerKey("custom.log.key"),
	)
	require.NoError(t, err)

	level := client.GetLogLevel("any.logger")
	assert.Equal(t, Warn, level)
	assert.Equal(t, "WARN", level.String())
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{Trace, "TRACE"},
		{Debug, "DEBUG"},
		{Info, "INFO"},
		{Warn, "WARN"},
		{Error, "ERROR"},
		{Fatal, "FATAL"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}
