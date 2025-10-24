package reforge

import (
	prefabProto "github.com/ReforgeHQ/sdk-go/proto"
)

// LogLevel represents the severity level for logging
type LogLevel int

const (
	// Trace is the most verbose log level, for very detailed debugging
	Trace LogLevel = iota + 1
	// Debug is for debugging information
	Debug
	// Info is for informational messages
	Info
	// Warn is for warning messages
	Warn
	// Error is for error messages
	Error
	// Fatal is for fatal error messages that may cause the application to exit
	Fatal
)

// String returns the string representation of the LogLevel
func (l LogLevel) String() string {
	switch l {
	case Trace:
		return "TRACE"
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// protoLogLevelToLogLevel converts a proto LogLevel to our SDK LogLevel
func protoLogLevelToLogLevel(protoLevel prefabProto.LogLevel) LogLevel {
	switch protoLevel {
	case prefabProto.LogLevel_TRACE:
		return Trace
	case prefabProto.LogLevel_DEBUG:
		return Debug
	case prefabProto.LogLevel_INFO:
		return Info
	case prefabProto.LogLevel_WARN:
		return Warn
	case prefabProto.LogLevel_ERROR:
		return Error
	case prefabProto.LogLevel_FATAL:
		return Fatal
	default:
		return Debug // Default to Debug for unknown/unset levels
	}
}
