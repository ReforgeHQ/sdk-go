//go:build ignore
// +build ignore

package main

import (
	"log/slog"
	"os"

	reforge "github.com/ReforgeHQ/sdk-go"
)

func main() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(
		reforge.WithSdkKey("your-sdk-key"),
		// Optional: customize the logger config key (default is "log-levels.default")
		reforge.WithLoggerKey("log-levels.default"),
	)
	if err != nil {
		panic(err)
	}

	// Example 1: Using ReforgeHandler
	// This wraps any slog.Handler and dynamically controls log levels
	baseHandler := slog.NewJSONHandler(os.Stdout, nil)
	reforgeHandler := reforge.NewReforgeHandler(client, baseHandler, "com.example.myapp")
	logger1 := slog.New(reforgeHandler)

	logger1.Debug("This will be logged if Reforge config sets DEBUG for com.example.myapp")
	logger1.Info("This will be logged if Reforge config sets INFO or lower")
	logger1.Error("This will be logged if Reforge config sets ERROR or lower")

	// Example 2: Using ReforgeLeveler with HandlerOptions
	// This provides dynamic level control via slog's standard mechanism
	leveler := reforge.NewReforgeLeveler(client, "com.example.myapp")
	handlerWithLeveler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: leveler,
	})
	logger2 := slog.New(handlerWithLeveler)

	logger2.Debug("Controlled by Reforge config")
	logger2.Info("Controlled by Reforge config")

	// Example 3: Direct log level check
	// You can also directly query the log level if needed
	level := client.GetLogLevel("com.example.myapp")
	slog.Info("Current log level", "level", level.String())

	// Example 4: Using with structured attributes
	logger3 := logger1.With("service", "api", "version", "1.0")
	logger3.Info("Request processed", "method", "GET", "path", "/api/users")

	// Example 5: Creating grouped loggers
	logger4 := logger1.WithGroup("database")
	logger4.Info("Query executed", "duration_ms", 42, "rows", 100)
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

This config will:
- Set DEBUG level for "com.example.myapp"
- Set INFO level for "com.example.database"
- Default to WARN for all other loggers

You can dynamically change these levels in Reforge without redeploying your application!
*/
