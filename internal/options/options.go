package options

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ReforgeHQ/sdk-go/internal/contexts"
)

type OnInitializationFailure int

const (
	ReturnError    OnInitializationFailure = iota // ReturnError = 0
	ReturnNilMatch                                // ReturnNilMatch = 1
)

const (
	// SdkKeyEnvVar #nosec G101 -- This is just the env var name
	SdkKeyEnvVar       = "REFORGE_BACKEND_SDK_KEY"
	LegacyApiKeyEnvVar = "PREFAB_API_KEY"
	APIURLVar          = "REFORGE_API_URL"
)

func GetDefaultAPIURLs() []string {
	return []string{
		"https://primary.reforge.com",
		"https://secondary.reforge.com",
	}
}

type ContextTelemetryMode string

var ContextTelemetryModes = struct {
	PeriodicExample ContextTelemetryMode
	Shapes          ContextTelemetryMode
	None            ContextTelemetryMode
}{
	PeriodicExample: "periodic_example",
	Shapes:          "shapes",
	None:            "",
}

type Options struct {
	GlobalContext                *contexts.ContextSet
	Configs                      map[string]interface{}
	SdkKey                       string
	APIURLs                      []string
	Sources                      []ConfigSource
	CustomStores                 []interface{} // ConfigStoreGetter implementations
	CustomEnvLookup              interface{}   // EnvLookup implementation
	EnvironmentNames             []string
	ProjectEnvID                 int64
	InitializationTimeoutSeconds float64
	OnInitializationFailure      OnInitializationFailure
	ContextTelemetryMode         ContextTelemetryMode
	CollectEvaluationSummaries   bool
	TelemetrySyncInterval        time.Duration
	TelemetryHost                string
	InstanceHash                 string
}

const timeoutDefault = 10.0

func GetDefaultOptions() Options {
	var apiURLs []string

	if os.Getenv("REFORGE_API_URL_OVERRIDE") != "" {
		apiURLs = []string{os.Getenv("REFORGE_API_URL_OVERRIDE")}
	}

	sources := GetDefaultConfigSources()

	if os.Getenv("REFORGE_DATAFILE") != "" {
		sources = []ConfigSource{
			{
				Path:    os.Getenv("REFORGE_DATAFILE"),
				Raw:     os.Getenv("REFORGE_DATAFILE"),
				Store:   DataFile,
				Default: false,
			},
		}
	}

	return Options{
		SdkKey:                       "",
		APIURLs:                      apiURLs,
		InitializationTimeoutSeconds: timeoutDefault,
		OnInitializationFailure:      ReturnError,
		GlobalContext:                contexts.NewContextSet(),
		Sources:                      sources,
		ContextTelemetryMode:         ContextTelemetryModes.PeriodicExample,
		TelemetrySyncInterval:        1 * time.Minute,
		TelemetryHost:                "https://telemetry.reforge.com",
		CollectEvaluationSummaries:   true,
		InstanceHash:                 uuid.New().String(),
	}
}

func (o *Options) TelemetryEnabled() bool {
	return o.CollectEvaluationSummaries || o.ContextTelemetryMode != ContextTelemetryModes.None
}

func (o *Options) SdkKeySettingOrEnvVar() (string, error) {
	if o.SdkKey == "" {
		// Attempt to read from environment variables if SdkKey is not directly set
		// Try new env var first, then fall back to legacy env var for backward compatibility
		envSdkKey := os.Getenv(SdkKeyEnvVar)
		if envSdkKey == "" {
			envSdkKey = os.Getenv(LegacyApiKeyEnvVar)
		}
		if envSdkKey == "" {
			return "", fmt.Errorf("SDK key is not set and not found in environment variables %s or %s", SdkKeyEnvVar, LegacyApiKeyEnvVar)
		}

		o.SdkKey = envSdkKey
	}

	return o.SdkKey, nil
}

func (o *Options) PrefabAPIURLEnvVarOrSetting() ([]string, error) {
	apiURLs := []string{}

	if os.Getenv(APIURLVar) != "" {
		for _, url := range strings.Split(os.Getenv(APIURLVar), ",") {
			if url != "" {
				apiURLs = append(apiURLs, url)
			}
		}

		if len(apiURLs) == 0 {
			return nil, fmt.Errorf("environment variable %s is blank", APIURLVar)
		}

		return apiURLs, nil
	}

	if os.Getenv("REFORGE_API_URL_OVERRIDE") != "" {
		apiURLs = []string{os.Getenv("REFORGE_API_URL_OVERRIDE")}
	} else {
		for _, url := range o.APIURLs {
			if url != "" {
				apiURLs = append(apiURLs, url)
			}
		}
	}

	if o.APIURLs == nil {
		apiURLs = GetDefaultAPIURLs()
	}

	return apiURLs, nil
}
