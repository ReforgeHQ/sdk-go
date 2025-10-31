# Charmbracelet Log Integration

Integration for [charmbracelet/log](https://github.com/charmbracelet/log) with real-time dynamic log level control from Reforge.

## Installation

```bash
go get github.com/ReforgeHQ/sdk-go/integrations/charmbracelet
```

## Quick Start

```go
import (
    reforge "github.com/ReforgeHQ/sdk-go"
    charmbracelet "github.com/ReforgeHQ/sdk-go/integrations/charmbracelet"
    "github.com/charmbracelet/log"
)

client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))

// Create a Reforge-wrapped logger
baseLogger := log.New(os.Stdout)
logger := charmbracelet.NewReforgeCharmLogger(client, baseLogger, "com.example.myapp")

// Log messages are filtered in real-time based on Reforge configuration
logger.Debug("Debug message")
logger.Info("Info message", "key", "value")
logger.Error("Error message")
```

## How It Works

The `ReforgeCharmLogger` wraps a charmbracelet logger and checks the Reforge configuration **on every log call** for real-time log level updates. When you change the log level in Reforge, it takes effect immediately via SSE without any polling or manual updates.

## API

### NewReforgeCharmLogger

Creates a logger that queries Reforge for the log level on each log call:

```go
logger := charmbracelet.NewReforgeCharmLogger(client, baseLogger, "com.example.myapp")
logger.Info("Checked on every call")
```

### Structured Logging

Supports all charmbracelet/log features:

```go
// Add context
requestLogger := logger.With("request_id", "abc-123", "user_id", "user-456")
requestLogger.Info("Processing request", "endpoint", "/api/data")

// Add prefix
serviceLogger := logger.WithPrefix("payment-service")
serviceLogger.Info("Payment processed", "amount", 99.99)
```

### Multiple Loggers

Different components can have different log levels:

```go
dbLogger := charmbracelet.NewReforgeCharmLogger(client, baseLogger, "com.example.database")
apiLogger := charmbracelet.NewReforgeCharmLogger(client, baseLogger, "com.example.api")

dbLogger.Debug("Database query") // Filtered based on com.example.database
apiLogger.Info("API request")     // Filtered based on com.example.api
```

## Examples

See [example_test.go](./example_test.go) for complete examples.

## Configuration

Configure log levels in Reforge using LOG_LEVEL_V2. See the [parent README](../README.md) for configuration format.

Changes to log levels in Reforge are propagated to your application in real-time via SSE, with no polling or restart required.
