package charmbracelet_test

import (
	"os"
	"time"

	reforge "github.com/ReforgeHQ/sdk-go"
	charmbracelet "github.com/ReforgeHQ/sdk-go/integrations/charmbracelet"
	"github.com/charmbracelet/log"
)

func Example_reforgeLevelFunc() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Using ReforgeLevelFunc with initial level setting
	// Update level periodically by calling SetLevel
	levelFunc := charmbracelet.NewReforgeLevelFunc(client, "com.example.myapp")
	logger := log.NewWithOptions(os.Stdout, log.Options{
		Level:           levelFunc.GetLevel(),
		ReportTimestamp: true,
	})

	logger.Debug("Debug message - controlled by Reforge")
	logger.Info("Info message - controlled by Reforge")
	logger.Error("Error message - controlled by Reforge")

	// You can update the level manually:
	// logger.SetLevel(levelFunc.GetLevel())
}

func Example_reforgeCharmLogger() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Using ReforgeCharmLogger for automatic level checking
	// This checks the level on every log call (most dynamic)
	baseLogger := log.NewWithOptions(os.Stdout, log.Options{
		ReportTimestamp: true,
	})
	reforgeLogger := charmbracelet.NewReforgeCharmLogger(client, baseLogger, "com.example.myapp")

	reforgeLogger.Debug("Debug message - checked dynamically")
	reforgeLogger.Info("Info message - checked dynamically")
	reforgeLogger.Error("Error message - checked dynamically")

	// Using With for structured logging
	requestLogger := reforgeLogger.With("request_id", "abc-123", "user_id", "user-456")
	requestLogger.Info("Processing request", "endpoint", "/api/data")

	// Using WithPrefix for logger hierarchies
	serviceLogger := reforgeLogger.WithPrefix("payment-service")
	serviceLogger.Info("Payment processed", "amount", 99.99, "currency", "USD")
}

func Example_reforgeAtomicLevel() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Using ReforgeAtomicLevel with automatic periodic updates
	// This updates the level in the background (good balance of performance and dynamism)
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportTimestamp: true,
	})
	atomicLevel := charmbracelet.NewReforgeAtomicLevel(client, logger, "com.example.myapp", 30*time.Second)
	defer atomicLevel.Stop()

	atomicLogger := atomicLevel.Logger()
	atomicLogger.Debug("Debug message - updates every 30 seconds")
	atomicLogger.Info("Info message - updates every 30 seconds")
}

func Example_multipleLoggers() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Multiple loggers for different components
	dbLogger := log.NewWithOptions(os.Stdout, log.Options{
		Prefix:          "database",
		ReportTimestamp: true,
	})
	reforgeDbLogger := charmbracelet.NewReforgeCharmLogger(client, dbLogger, "com.example.database")

	apiLogger := log.NewWithOptions(os.Stdout, log.Options{
		Prefix:          "api",
		ReportTimestamp: true,
	})
	reforgeApiLogger := charmbracelet.NewReforgeCharmLogger(client, apiLogger, "com.example.api")

	reforgeDbLogger.Debug("Database query executed", "duration_ms", 42, "rows", 100)
	reforgeApiLogger.Info("API request received", "method", "GET", "path", "/api/users")
}
