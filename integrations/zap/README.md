# Zap Integration

Integration for [uber-go/zap](https://github.com/uber-go/zap) with dynamic log level control from Reforge.

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

// Approach 1: Dynamic level checking on each log
dynamicLevel := reforgezap.NewReforgeZapLevel(client, "com.example.myapp")
logger, _ := zap.NewProduction(zap.IncreaseLevel(dynamicLevel))
logger.Info("Dynamic logging!")

// Approach 2: Automatic background updates
atomicLevel := reforgezap.NewReforgeAtomicLevel(client, "com.example.myapp", 30*time.Second)
defer atomicLevel.Stop()
config := zap.NewProductionConfig()
config.Level = atomicLevel.AtomicLevel()
logger, _ := config.Build()
```

## API

### ReforgeZapLevel

Implements `zapcore.LevelEnabler` and checks Reforge on every log call:

```go
dynamicLevel := reforgezap.NewReforgeZapLevel(client, "com.example.myapp")
logger, _ := zap.NewProduction(zap.IncreaseLevel(dynamicLevel))
```

Most dynamic approach - instant updates, slight performance cost.

### ReforgeAtomicLevel

Wraps `zap.AtomicLevel` and updates periodically in the background:

```go
atomicLevel := reforgezap.NewReforgeAtomicLevel(client, "com.example.myapp", 30*time.Second)
defer atomicLevel.Stop()
config := zap.NewProductionConfig()
config.Level = atomicLevel.AtomicLevel()
logger, _ := config.Build()
```

Good balance of performance and dynamism.

### ReforgeZapCore

Wraps a `zapcore.Core` for fine-grained control:

```go
encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
baseCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
core := reforgezap.NewReforgeZapCore(baseCore, client, "com.example.myapp")
logger := zap.New(core)
```

## Examples

See [example_test.go](./example_test.go) for complete examples.

## Configuration

Configure log levels in Reforge using LOG_LEVEL_V2. See the [parent README](../README.md) for configuration format.

## Performance Notes

- **ReforgeZapLevel**: Checks on every log call (most dynamic, slight overhead)
- **ReforgeAtomicLevel**: Updates periodically (good balance)
- **ReforgeZapCore**: Checks on every log call, more control over core behavior
