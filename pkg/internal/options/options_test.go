package options_test

import (
	"testing"

	reforge "github.com/ReforgeHQ/sdk-go/pkg"

	"github.com/stretchr/testify/assert"

	"github.com/ReforgeHQ/sdk-go/pkg/internal/options"
)

func TestGetDefaultOptions(t *testing.T) {
	// When ENV var PREFAB_API_URL_OVERRIDE is not set we use the default API URL.
	t.Setenv("PREFAB_API_URL_OVERRIDE", "")

	o := options.GetDefaultOptions()

	assert.Empty(t, o.SdkKey)
	assert.Nil(t, o.APIURLs)
	assert.Equal(t, 10.0, o.InitializationTimeoutSeconds)
	assert.Equal(t, options.ReturnError, o.OnInitializationFailure)
	assert.NotNil(t, o.GlobalContext)
	assert.Len(t, o.Sources, 1)
	assert.Equal(t, options.ConfigSource{
		Store:   options.APIStore,
		Raw:     "api:prefab",
		Default: true,
	}, o.Sources[0])

	// When ENV var PREFAB_API_URL_OVERRIDE is set, that should be used instead
	// of the default API URL.
	desiredAPIURL := "https://api.staging-reforge.com"

	t.Setenv("PREFAB_API_URL_OVERRIDE", desiredAPIURL)

	o = options.GetDefaultOptions()
	assert.Equal(t, []string{desiredAPIURL}, o.APIURLs)

	t.Setenv("PREFAB_API_URL_OVERRIDE", "")
	t.Setenv("PREFAB_DATAFILE", "testdata/download.json")

	o = options.GetDefaultOptions()
	assert.Equal(t, []options.ConfigSource([]options.ConfigSource{{Store: "DataFile", Raw: "testdata/download.json", Path: "testdata/download.json", Default: false}}), o.Sources)
}

func TestOptions_TelemetryEnabledOverride(t *testing.T) {
	defaultOptions := options.GetDefaultOptions()
	assert.True(t, defaultOptions.TelemetryEnabled())
	_ = reforge.WithAllTelemetryDisabled()(&defaultOptions)
	assert.False(t, defaultOptions.TelemetryEnabled())
}

func TestOptions_TelemetryEnabledCalculatesBasedOnOptions(t *testing.T) {
	defaultOptions := options.GetDefaultOptions()
	assert.True(t, defaultOptions.TelemetryEnabled())

	_ = reforge.WithCollectEvaluationSummaries(false)(&defaultOptions)
	assert.True(t, defaultOptions.TelemetryEnabled())

	_ = reforge.WithContextTelemetryMode(options.ContextTelemetryModes.None)(&defaultOptions)
	assert.False(t, defaultOptions.TelemetryEnabled())
}
