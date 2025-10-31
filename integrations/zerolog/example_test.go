package zerolog_test

import (
	"os"

	reforge "github.com/ReforgeHQ/sdk-go"
	reforgezerolog "github.com/ReforgeHQ/sdk-go/integrations/zerolog"
	"github.com/rs/zerolog"
)

func Example_reforgeZerologHook() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Using Hook (checks level on every log event for real-time dynamic control)
	hook := reforgezerolog.NewReforgeZerologHook(client, "com.example.myapp")
	logger := zerolog.New(os.Stdout).Hook(hook).With().Timestamp().Logger()

	logger.Debug().Msg("Debug message - controlled by Reforge")
	logger.Info().Msg("Info message - controlled by Reforge")
	logger.Error().Msg("Error message - controlled by Reforge")
}

func Example_multipleLoggers() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Multiple loggers for different components
	dbLogger := zerolog.New(os.Stdout).
		Hook(reforgezerolog.NewReforgeZerologHook(client, "com.example.database")).
		With().
		Str("component", "database").
		Timestamp().
		Logger()

	apiLogger := zerolog.New(os.Stdout).
		Hook(reforgezerolog.NewReforgeZerologHook(client, "com.example.api")).
		With().
		Str("component", "api").
		Timestamp().
		Logger()

	dbLogger.Debug().Msg("Database query executed")
	apiLogger.Info().Msg("API request received")
}
