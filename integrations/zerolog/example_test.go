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

	// Using Hook (checks level on every log event)
	hook := reforgezerolog.NewReforgeZerologHook(client, "com.example.myapp")
	logger := zerolog.New(os.Stdout).Hook(hook).With().Timestamp().Logger()

	logger.Debug().Msg("Debug message - controlled by Reforge")
	logger.Info().Msg("Info message - controlled by Reforge")
	logger.Error().Msg("Error message - controlled by Reforge")
}

func Example_reforgeZerologLevelWriter() {
	// Initialize Reforge SDK
	client, err := reforge.NewSdk(reforge.WithSdkKey("your-sdk-key"))
	if err != nil {
		panic(err)
	}

	// Using LevelWriter with periodic level updates
	// This is more efficient as it sets the level once rather than checking on each event
	levelWriter := reforgezerolog.NewReforgeZerologLevelWriter(client, "com.example.myapp", os.Stdout)
	logger := zerolog.New(levelWriter).
		Level(levelWriter.GetDynamicLevel()).
		With().
		Timestamp().
		Logger()

	logger.Debug().Msg("Debug message")
	logger.Info().Msg("Info message")

	// You can update the level periodically:
	// ticker := time.NewTicker(30 * time.Second)
	// go func() {
	//     for range ticker.C {
	//         logger = logger.Level(levelWriter.GetDynamicLevel())
	//     }
	// }()
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
