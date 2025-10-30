# Running the Test Launcher

This test launcher will help you verify dynamic log level changes in staging with debug output.

## Setup

1. Make sure you're on the debug logging branch:
```bash
git checkout add-debug-logging-for-sse-and-loglevel
```

2. Install charmbracelet/log:
```bash
go get github.com/charmbracelet/log
```

3. Run the test launcher:
```bash
go run test_launcher.go
```

## What it does

- Connects to Reforge staging with the hardcoded SDK key
- Creates a logger named `test.launcher`
- Every 10 seconds, logs messages at DEBUG, INFO, and WARN levels
- Each message includes the currently configured log level from Reforge

## Expected output

You should see:
- `[SSE]` messages when config updates arrive
- `[GetLogLevel]` debug output showing how log levels are resolved
- Log messages filtered based on your Reforge configuration

## Testing dynamic updates

1. Start the launcher
2. In Reforge staging, change the log level for `test.launcher`
3. Watch the SSE update messages appear
4. See which log messages now appear/disappear based on the new level

## Configuration in Reforge

Configure a LOG_LEVEL_V2 config (probably `log-levels.default`) with:
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
                  "values": ["test.launcher"]
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

Change the log_level value to test different levels: TRACE, DEBUG, INFO, WARN, ERROR, FATAL
