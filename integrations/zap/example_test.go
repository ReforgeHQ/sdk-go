package zap_test

import (
	"os"

	reforge "github.com/ReforgeHQ/sdk-go"
	reforgezap "github.com/ReforgeHQ/sdk-go/integrations/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Example_reforgeZapLevel() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Using ReforgeZapLevel with IncreaseLevel
	// This checks Reforge on every log call for real-time dynamic control
	dynamicLevel := reforgezap.NewReforgeZapLevel(client, "com.example.myapp")
	logger, _ := zap.NewProduction(zap.IncreaseLevel(dynamicLevel))
	defer logger.Sync()

	logger.Debug("Debug message - controlled by Reforge")
	logger.Info("Info message - controlled by Reforge")
	logger.Error("Error message - controlled by Reforge")
}

func Example_reforgeZapCore() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Using ReforgeZapCore for fine-grained control
	// This wraps the core directly
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	baseCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
	reforgeCore := reforgezap.NewReforgeZapCore(baseCore, client, "com.example.myapp")
	logger := zap.New(reforgeCore)
	defer logger.Sync()

	logger.Debug("Debug message from custom core")
	logger.Info("Info message from custom core")
}

func Example_multipleLoggers() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Multiple loggers for different components
	dbLevel := reforgezap.NewReforgeZapLevel(client, "com.example.database")
	dbLogger, _ := zap.NewProduction(
		zap.IncreaseLevel(dbLevel),
		zap.Fields(zap.String("component", "database")),
	)
	defer dbLogger.Sync()

	apiLevel := reforgezap.NewReforgeZapLevel(client, "com.example.api")
	apiLogger, _ := zap.NewProduction(
		zap.IncreaseLevel(apiLevel),
		zap.Fields(zap.String("component", "api")),
	)
	defer apiLogger.Sync()

	dbLogger.Debug("Database query executed", zap.Int("duration_ms", 42))
	apiLogger.Info("API request received", zap.String("method", "GET"))
}
