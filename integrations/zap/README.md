# Zap Integration

Integration for [uber-go/zap](https://github.com/uber-go/zap) with real-time dynamic log level control from Reforge.

## Installation

```bash
go get github.com/ReforgeHQ/sdk-go/integrations/zap
```

## Quick Start

```go
import (
    reforge "github.com/ReforgeHQ/sdk-go"
    reforgezap "github.com/ReforgeHQ/sdk-go/integrations/zap"
    "go.uber.org/zap"
)

client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))

// Approach 1: Using ReforgeZapLevel with IncreaseLevel
dynamicLevel := reforgezap.NewReforgeZapLevel(client, "com.example.myapp")
logger, _ := zap.NewProduction(zap.IncreaseLevel(dynamicLevel))
logger.Info("Dynamic logging!")

// Approach 2: Using ReforgeZapCore for fine-grained control
encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
baseCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
core := reforgezap.NewReforgeZapCore(baseCore, client, "com.example.myapp")
logger := zap.New(core)
```

## How It Works

Both approaches check Reforge configuration **on every log call** for real-time log level updates. When you change the log level in Reforge, it takes effect immediately via SSE without any polling or manual updates.

## API

### ReforgeZapLevel

Implements `zapcore.LevelEnabler` and checks Reforge on every log call:

```go
dynamicLevel := reforgezap.NewReforgeZapLevel(client, "com.example.myapp")
logger, _ := zap.NewProduction(zap.IncreaseLevel(dynamicLevel))
```

### ReforgeZapCore

Wraps a `zapcore.Core` for fine-grained control:

```go
encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
baseCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
core := reforgezap.NewReforgeZapCore(baseCore, client, "com.example.myapp")
logger := zap.New(core)
```

### Multiple Loggers

Different components can have different log levels:

```go
dbLevel := reforgezap.NewReforgeZapLevel(client, "com.example.database")
apiLevel := reforgezap.NewReforgeZapLevel(client, "com.example.api")
```

## Examples

See [example_test.go](./example_test.go) for complete examples.

## Configuration

Configure log levels in Reforge using LOG_LEVEL_V2. See the [parent README](../README.md) for configuration format.

Changes to log levels in Reforge are propagated to your application in real-time via SSE, with no polling or restart required.
