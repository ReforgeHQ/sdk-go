# Reforge Logger Integrations

This directory contains optional integrations for popular Go logging libraries. Each integration is a separate Go module, so you only pull in the logging dependencies you actually use.

## Available Integrations

- **[charmbracelet](./charmbracelet)** - Integration for [charmbracelet/log](https://github.com/charmbracelet/log)
- **[zap](./zap)** - Integration for [uber-go/zap](https://github.com/uber-go/zap)
- **[zerolog](./zerolog)** - Integration for [rs/zerolog](https://github.com/rs/zerolog)

For **slog** (standard library), use the built-in integration in the main `github.com/ReforgeHQ/sdk-go` module - no extra dependencies needed!

## Usage

Import only the integration you need:

```go
import (
    reforge "github.com/ReforgeHQ/sdk-go"
    charmbracelet "github.com/ReforgeHQ/sdk-go/integrations/charmbracelet"
)
```

Each integration provides dynamic log level control from Reforge configuration, allowing you to adjust log verbosity in production without redeploying.

## Installation

```bash
# For charmbracelet/log
go get github.com/ReforgeHQ/sdk-go/integrations/charmbracelet

# For zap
go get github.com/ReforgeHQ/sdk-go/integrations/zap

# For zerolog
go get github.com/ReforgeHQ/sdk-go/integrations/zerolog
```

## Benefits of Separate Modules

- **Zero bloat** - Main SDK has no logging dependencies
- **Pick what you need** - Only import integrations you use
- **Independent versioning** - Each integration can be updated independently
- **Clean dependency trees** - Your project only includes the loggers you need

## Configuration

All integrations use the same Reforge LOG_LEVEL_V2 configuration format:

```json
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
        }
      ],
      "value": { "log_level": "INFO" }
    }
  ]
}
```

See each integration's directory for specific usage examples and API documentation.
