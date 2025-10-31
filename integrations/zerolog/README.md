# Zerolog Integration

Integration for [rs/zerolog](https://github.com/rs/zerolog) with real-time dynamic log level control from Reforge.

## Installation

```bash
go get github.com/ReforgeHQ/sdk-go/integrations/zerolog
```

## Quick Start

```go
import (
    reforge "github.com/ReforgeHQ/sdk-go"
    reforgezerolog "github.com/ReforgeHQ/sdk-go/integrations/zerolog"
    "github.com/rs/zerolog"
)

client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))

// Create a Reforge-controlled logger
hook := reforgezerolog.NewReforgeZerologHook(client, "com.example.myapp")
logger := zerolog.New(os.Stdout).Hook(hook).With().Timestamp().Logger()

// Log messages are filtered in real-time based on Reforge configuration
logger.Debug().Msg("Debug message")
logger.Info().Msg("Info message")
logger.Error().Msg("Error message")
```

## How It Works

The `ReforgeZerologHook` checks the Reforge configuration **on every log event** for real-time log level updates. When you change the log level in Reforge, it takes effect immediately via SSE without any polling or manual updates.

## API

### NewReforgeZerologHook

Creates a hook that queries Reforge for the log level on each log event:

```go
hook := reforgezerolog.NewReforgeZerologHook(client, "com.example.myapp")
logger := zerolog.New(os.Stdout).Hook(hook).With().Timestamp().Logger()
logger.Info().Msg("Checked on every log event")
```

### Structured Logging

Supports all zerolog features:

```go
// Add context
logger.Info().
    Str("request_id", "abc-123").
    Str("user_id", "user-456").
    Msg("Processing request")

// Sub-loggers
subLogger := logger.With().Str("component", "database").Logger()
subLogger.Debug().Msg("Database query")
```

### Multiple Loggers

Different components can have different log levels:

```go
dbLogger := zerolog.New(os.Stdout).
    Hook(reforgezerolog.NewReforgeZerologHook(client, "com.example.database")).
    With().Timestamp().Logger()

apiLogger := zerolog.New(os.Stdout).
    Hook(reforgezerolog.NewReforgeZerologHook(client, "com.example.api")).
    With().Timestamp().Logger()

dbLogger.Debug().Msg("Database query") // Filtered based on com.example.database
apiLogger.Info().Msg("API request")    // Filtered based on com.example.api
```

## Examples

See [example_test.go](./example_test.go) for complete examples.

## Configuration

Configure log levels in Reforge using LOG_LEVEL_V2. See the [parent README](../README.md) for configuration format.

Changes to log levels in Reforge are propagated to your application in real-time via SSE, with no polling or restart required.
