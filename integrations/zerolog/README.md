# Zerolog Integration

Integration for [rs/zerolog](https://github.com/rs/zerolog) with dynamic log level control from Reforge.

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

// Approach 1: Hook-based (checks on every log event)
hook := reforgezerolog.NewReforgeZerologHook(client, "com.example.myapp")
logger := zerolog.New(os.Stdout).Hook(hook).With().Timestamp().Logger()
logger.Info().Msg("Dynamic logging!")

// Approach 2: LevelWriter (more efficient, set level periodically)
levelWriter := reforgezerolog.NewReforgeZerologLevelWriter(client, "com.example.myapp", os.Stdout)
logger := zerolog.New(levelWriter).
    Level(levelWriter.GetDynamicLevel()).
    With().
    Timestamp().
    Logger()
```

## API

### ReforgeZerologHook

Implements `zerolog.Hook` and checks Reforge on every log event:

```go
hook := reforgezerolog.NewReforgeZerologHook(client, "com.example.myapp")
logger := zerolog.New(os.Stdout).Hook(hook).With().Timestamp().Logger()
```

Most dynamic approach - filters events based on current Reforge configuration.

### ReforgeZerologLevelWriter

Wraps an `io.Writer` and provides level querying:

```go
levelWriter := reforgezerolog.NewReforgeZerologLevelWriter(client, "com.example.myapp", os.Stdout)
logger := zerolog.New(levelWriter).
    Level(levelWriter.GetDynamicLevel()).
    With().
    Timestamp().
    Logger()

// Update level periodically:
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        logger = logger.Level(levelWriter.GetDynamicLevel())
    }
}()
```

More efficient - set the level once rather than checking on each event.

## Examples

See [example_test.go](./example_test.go) for complete examples.

## Configuration

Configure log levels in Reforge using LOG_LEVEL_V2. See the [parent README](../README.md) for configuration format.

## Performance Notes

- **ReforgeZerologHook**: Checks on every log event (most dynamic)
- **ReforgeZerologLevelWriter**: Set level once, update manually or periodically (better performance)
