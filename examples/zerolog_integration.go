//go:build ignore
// +build ignore

package main

// This example shows how to integrate Reforge log level management with zerolog.
// Copy this code into your project and adapt as needed.
//
// Compatible with zerolog v1.15.0+
//
// To use: go get github.com/rs/zerolog

import (
	"io"
	"os"

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
//	hook := NewReforgeZerologHook(client, "com.example.myapp")
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
//	levelWriter := NewReforgeZerologLevelWriter(client, "com.example.myapp", os.Stdout)
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

func main() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Approach 1: Using Hook (checks level on every log event)
	hook := NewReforgeZerologHook(client, "com.example.myapp")
	logger1 := zerolog.New(os.Stdout).Hook(hook).With().Timestamp().Logger()

	logger1.Debug().Msg("Debug message - controlled by Reforge")
	logger1.Info().Msg("Info message - controlled by Reforge")
	logger1.Error().Msg("Error message - controlled by Reforge")

	// Approach 2: Using LevelWriter with periodic level updates
	// This is more efficient as it sets the level once rather than checking on each event
	levelWriter := NewReforgeZerologLevelWriter(client, "com.example.myapp", os.Stdout)
	logger2 := zerolog.New(levelWriter).
		Level(levelWriter.GetDynamicLevel()).
		With().
		Timestamp().
		Logger()

	logger2.Debug().Msg("Debug message")
	logger2.Info().Msg("Info message")

	// You can update the level periodically:
	// ticker := time.NewTicker(30 * time.Second)
	// go func() {
	//     for range ticker.C {
	//         logger2 = logger2.Level(levelWriter.GetDynamicLevel())
	//     }
	// }()

	// Approach 3: Multiple loggers for different components
	dbLogger := zerolog.New(os.Stdout).
		Hook(NewReforgeZerologHook(client, "com.example.database")).
		With().
		Str("component", "database").
		Timestamp().
		Logger()

	apiLogger := zerolog.New(os.Stdout).
		Hook(NewReforgeZerologHook(client, "com.example.api")).
		With().
		Str("component", "api").
		Timestamp().
		Logger()

	dbLogger.Debug().Msg("Database query executed")
	apiLogger.Info().Msg("API request received")
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
        }
      ],
      "value": { "log_level": "WARN" }
    }
  ]
}

Change these levels in Reforge to dynamically control your application's logging!
*/
