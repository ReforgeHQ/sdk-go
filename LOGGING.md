# Dynamic Log Level Management

Reforge SDK provides dynamic log level management that allows you to control your application's logging behavior in real-time without redeployment. This is useful for:

- Debugging production issues by temporarily increasing log verbosity
- Reducing log volume during normal operations
- Different log levels for different components/packages
- A/B testing logging configurations

## Core Functionality

### GetLogLevel Method

The SDK provides a `GetLogLevel(loggerName string)` method that evaluates log level configuration from Reforge:

```go
client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
level := client.GetLogLevel("com.example.myapp")
// Returns: reforge.LogLevel (Trace, Debug, Info, Warn, Error, Fatal)
```

### Configuration

Log levels are configured in Reforge using `LOG_LEVEL_V2` config type with the key `log-levels.default` (configurable via `WithLoggerKey` option).

The SDK creates a context with:
```go
{
  "reforge-sdk-logging": {
    "lang": "go",
    "logger-path": "<loggerName>"
  }
}
```

This allows you to target different loggers with different log levels using Reforge's rule engine.

### LogLevel Type

The SDK exposes its own `LogLevel` type (not the proto enum):

```go
type LogLevel int

const (
    Trace LogLevel = iota + 1
    Debug
    Info
    Warn
    Error
    Fatal
)
```

## Integration with Logging Frameworks

### slog (Standard Library)

Reforge provides **built-in** slog integration with no additional dependencies.

#### Option 1: ReforgeHandler

Wraps any `slog.Handler` and dynamically filters log records:

```go
import (
    "log/slog"
    "os"
    reforge "github.com/ReforgeHQ/sdk-go"
)

client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
handler := reforge.NewReforgeHandler(
    client,
    slog.NewJSONHandler(os.Stdout, nil),
    "com.example.myapp",
)
logger := slog.New(handler)

logger.Debug("This respects Reforge config")
logger.Info("Dynamic log level control")
```

#### Option 2: ReforgeLeveler

Implements `slog.Leveler` for use with `HandlerOptions`:

```go
client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
leveler := reforge.NewReforgeLeveler(client, "com.example.myapp")
handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: leveler,
})
logger := slog.New(handler)
```

**See:** `examples/slog_example.go` for complete examples

### zerolog

Reforge provides **example integration code** for zerolog that you can copy into your project.

#### Option 1: Hook-based (Per-event checking)

```go
import (
    "github.com/rs/zerolog"
    reforge "github.com/ReforgeHQ/sdk-go"
)

client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
hook := NewReforgeZerologHook(client, "com.example.myapp")
logger := zerolog.New(os.Stdout).Hook(hook)
```

#### Option 2: LevelWriter (Periodic updates)

```go
levelWriter := NewReforgeZerologLevelWriter(client, "com.example.myapp", os.Stdout)
logger := zerolog.New(levelWriter).Level(levelWriter.GetDynamicLevel())

// Update periodically:
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        logger = logger.Level(levelWriter.GetDynamicLevel())
    }
}()
```

**See:** `examples/zerolog_integration.go` for complete implementation

**Compatible with:** zerolog v1.15.0+

### zap (uber-go/zap)

Reforge provides **example integration code** for zap that you can copy into your project.

#### Option 1: ReforgeZapLevel (Most Dynamic)

Implements `zapcore.LevelEnabler` for per-log checking:

```go
import (
    "go.uber.org/zap"
    reforge "github.com/ReforgeHQ/sdk-go"
)

client, _ := reforge.NewSdk(reforge.WithSdkKey("your-key"))
dynamicLevel := NewReforgeZapLevel(client, "com.example.myapp")
logger, _ := zap.NewProduction(zap.IncreaseLevel(dynamicLevel))
```

#### Option 2: ReforgeAtomicLevel (Balanced)

Wraps `zap.AtomicLevel` with automatic periodic updates:

```go
atomicLevel := NewReforgeAtomicLevel(client, "com.example.myapp", 30*time.Second)
defer atomicLevel.Stop()

config := zap.NewProductionConfig()
config.Level = atomicLevel.AtomicLevel()
logger, _ := config.Build()
```

#### Option 3: ReforgeZapCore (Fine-grained)

Custom `zapcore.Core` implementation:

```go
encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
baseCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
reforgeCore := NewReforgeZapCore(baseCore, client, "com.example.myapp")
logger := zap.New(reforgeCore)
```

**See:** `examples/zap_integration.go` for complete implementation

**Compatible with:** zap v1.10.0+ (uber-go/zap)

## Configuration Examples

### Basic Configuration

```json
{
  "configType": "LOG_LEVEL_V2",
  "valueType": "LOG_LEVEL",
  "rows": [
    {
      "values": [
        {
          "value": { "logLevel": "INFO" }
        }
      ]
    }
  ]
}
```

This sets INFO level for all loggers.

### Per-Logger Configuration

```json
{
  "configType": "LOG_LEVEL_V2",
  "valueType": "LOG_LEVEL",
  "rows": [
    {
      "values": [
        {
          "criteria": [
            {
              "operator": "PROP_IS_ONE_OF",
              "propertyName": "reforge-sdk-logging.logger-path",
              "valueToMatch": {
                "stringList": {
                  "values": ["com.example.database"]
                }
              }
            }
          ],
          "value": { "logLevel": "DEBUG" }
        },
        {
          "criteria": [
            {
              "operator": "PROP_IS_ONE_OF",
              "propertyName": "reforge-sdk-logging.logger-path",
              "valueToMatch": {
                "stringList": {
                  "values": ["com.example.api"]
                }
              }
            }
          ],
          "value": { "logLevel": "INFO" }
        },
        {
          "value": { "logLevel": "WARN" }
        }
      ]
    }
  ]
}
```

This configuration:
- Sets DEBUG for `com.example.database`
- Sets INFO for `com.example.api`
- Sets WARN as default for all other loggers

### Pattern Matching

You can use Reforge's rule operators for more sophisticated matching:

```json
{
  "criteria": [
    {
      "operator": "PROP_STARTS_WITH",
      "propertyName": "reforge-sdk-logging.logger-path",
      "valueToMatch": {
        "string": "com.example.services"
      }
    }
  ],
  "value": { "logLevel": "DEBUG" }
}
```

This sets DEBUG for all loggers starting with `com.example.services`.

## Customization

### Custom Logger Config Key

By default, the SDK looks for log level config at key `log-levels.default`. You can customize this:

```go
client, _ := reforge.NewSdk(
    reforge.WithSdkKey("your-key"),
    reforge.WithLoggerKey("my.custom.log.config"),
)
```

### Logger Naming Conventions

It's recommended to use hierarchical logger names that mirror your package structure:

```go
// Good
dbLogger := logger.GetLogLevel("com.mycompany.myapp.database")
apiLogger := logger.GetLogLevel("com.mycompany.myapp.api")
authLogger := logger.GetLogLevel("com.mycompany.myapp.api.auth")

// This allows you to configure:
// - All of myapp: com.mycompany.myapp
// - All of api: com.mycompany.myapp.api
// - Just auth: com.mycompany.myapp.api.auth
```

## Performance Considerations

### slog
- **ReforgeHandler**: Checks Reforge on every log event. Minimal overhead due to Go's efficient context switching.
- **ReforgeLeveler**: Same performance characteristics as ReforgeHandler.

### zerolog
- **Hook**: Checks Reforge on every log event. Slight overhead but very fast.
- **LevelWriter**: Only updates level periodically. Best performance but less dynamic.

### zap
- **ReforgeZapLevel**: Checks Reforge on every log event. Slight overhead.
- **ReforgeAtomicLevel**: Updates periodically (configurable interval). Good balance of performance and dynamism.
- **ReforgeZapCore**: Similar to ReforgeZapLevel but with more control.

**Recommendation:** For most applications, the per-event checking overhead is negligible. If you have extremely high-throughput logging (>100k logs/sec), consider periodic update approaches.

## Version Compatibility

| Framework | Minimum Version | Notes |
|-----------|----------------|-------|
| slog | Go 1.21+ | Built into Go standard library |
| zerolog | v1.15.0+ | Example code provided |
| zap | v1.10.0+ | Example code provided (uber-go/zap) |

The zerolog and zap integrations are provided as example code that you can copy and adapt to your specific version and needs. This ensures maximum compatibility across versions.

## Testing

The SDK includes comprehensive tests for log level functionality:

```bash
go test -v -run TestGetLogLevel          # Core functionality
go test -v -run TestReforgeHandler       # slog integration
go test -v -run TestReforgeLeveler       # slog leveler
```

## FAQ

**Q: How often does the SDK check Reforge for log level changes?**

A: It depends on the integration approach:
- Hook/Handler based: Checks on every log event (very efficient, ~microseconds)
- Atomic/Periodic: Updates at your specified interval (e.g., every 30 seconds)

**Q: What happens if Reforge is unavailable?**

A: The SDK returns `Debug` as the default log level. Your application continues to run with full debug logging.

**Q: Can I use different log levels for different environments?**

A: Yes! Use Reforge's environment targeting in your rules to set different levels for dev, staging, and production.

**Q: Do I need to add zerolog/zap as dependencies?**

A: No. The zerolog and zap integrations are provided as example code in the `examples/` directory. Copy the relevant code into your project and it will work with your existing zerolog/zap installation.

## Examples

All examples are in the `examples/` directory:
- `examples/slog_example.go` - Complete slog integration examples
- `examples/zerolog_integration.go` - Complete zerolog integration code
- `examples/zap_integration.go` - Complete zap integration code

## Support

For issues or questions, please open an issue on GitHub or contact support@reforge.com.
