# Charmbracelet Log Integration

Integration for [charmbracelet/log](https://github.com/charmbracelet/log) with dynamic log level control from Reforge.

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

// Approach 1: Wrapped logger with per-call level checking
baseLogger := log.New(os.Stdout)
logger := charmbracelet.NewReforgeCharmLogger(client, baseLogger, "com.example.myapp")
logger.Info("Dynamic logging!", "controlled_by", "reforge")

// Approach 2: Automatic background updates
logger := log.New(os.Stdout)
atomicLevel := charmbracelet.NewReforgeAtomicLevel(client, logger, "com.example.myapp", 30*time.Second)
defer atomicLevel.Stop()
```

## API

### ReforgeLevelFunc

Sets the initial level and allows manual updates:

```go
levelFunc := charmbracelet.NewReforgeLevelFunc(client, "com.example.myapp")
logger := log.NewWithOptions(os.Stdout, log.Options{
    Level: levelFunc.GetLevel(),
})
// Update manually: logger.SetLevel(levelFunc.GetLevel())
```

### ReforgeCharmLogger

Wraps a charmbracelet logger and checks Reforge on every log call (most dynamic):

```go
logger := charmbracelet.NewReforgeCharmLogger(client, baseLogger, "com.example.myapp")
logger.Info("Checked on every call")
logger.With("key", "value").Debug("Supports structured logging")
```

### ReforgeAtomicLevel

Automatically updates the log level in the background at intervals:

```go
atomicLevel := charmbracelet.NewReforgeAtomicLevel(client, logger, "com.example.myapp", 30*time.Second)
defer atomicLevel.Stop()
```

## Examples

See [example_test.go](./example_test.go) for complete examples.

## Configuration

Configure log levels in Reforge using LOG_LEVEL_V2. See the [parent README](../README.md) for configuration format.
