package reforge_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	reforge "github.com/ReforgeHQ/sdk-go"
	"github.com/ReforgeHQ/sdk-go/internal/options"
)

func TestWithConfig(t *testing.T) {
	configs := map[string]interface{}{
		"string.key": "value",
		"int.key":    int64(42),
		"bool.key":   true,
		"float.key":  3.14,
		"slice.key":  []string{"a", "b", "c"},
		"json.key": map[string]interface{}{
			"nested": "value",
		},
	}

	client, err := reforge.NewSdk(
		reforge.WithConfigs(configs),
		reforge.WithInitializationTimeoutSeconds(5.0),
		reforge.WithContextTelemetryMode(options.ContextTelemetryModes.None))

	require.NoError(t, err)

	str, ok, err := client.GetStringValue("string.key", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "value", str)

	i, ok, err := client.GetIntValue("int.key", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, int64(42), i)

	b, ok, err := client.GetBoolValue("bool.key", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, b)

	f, ok, err := client.GetFloatValue("float.key", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 3.14, f, 0.0001)

	slice, ok, err := client.GetStringSliceValue("slice.key", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b", "c"}, slice)

	json, ok, err := client.GetJSONValue("json.key", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"nested": "value"}, json)
}

func TestCannotUseWithConfigAndOtherSources(t *testing.T) {
	configs := map[string]interface{}{
		"string.key": "value",
	}

	_, err := reforge.NewSdk(
		reforge.WithConfigs(configs),
		reforge.WithContextTelemetryMode(options.ContextTelemetryModes.None),
		reforge.WithSources([]string{}, false)) // Explicitly try to use online sources

	require.Error(t, err)
	assert.Equal(t, "cannot use WithConfigs with other sources", err.Error())

	_, err = reforge.NewSdk(
		reforge.WithConfigs(configs),
		reforge.WithOfflineSources([]string{"datafile://testdata/download.json"}),
		reforge.WithContextTelemetryMode(options.ContextTelemetryModes.None),
		reforge.WithProjectEnvID(8),
	)

	require.Error(t, err)
	assert.Equal(t, "cannot use WithConfigs with other sources", err.Error())
}

func TestWithAJSONConfigDump(t *testing.T) {
	t.Setenv("REFORGE_DATAFILE", "testdata/download.json")

	client, err := reforge.NewSdk(reforge.WithContextTelemetryMode(options.ContextTelemetryModes.None))
	require.NoError(t, err)

	str, ok, err := client.GetStringValue("my.test.string", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "hello world", str)

	boolean, ok, err := client.GetBoolValue("flag.list.environments", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.False(t, boolean)

	contextSet := reforge.NewContextSet().
		WithNamedContextValues("user", map[string]interface{}{
			"key": "5905ecd1-9bbf-4711-a663-4f713628a78c",
		})

	boolean, ok, err = client.GetBoolValue("flag.list.environments", *contextSet)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, boolean)

	// Same thing as above, but with client.WithContext
	boolean, ok, err = client.WithContext(contextSet).GetBoolValue("flag.list.environments", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, boolean)

	// This one is deleted
	_, ok, err = client.GetLogLevelStringValue("log-level", *reforge.NewContextSet())
	require.ErrorContains(t, err, "config did not produce a result and no default is specified")
	assert.False(t, ok)
}

func TestGetConfigMatchWithAJSONConfigDumpAndGlobalContext(t *testing.T) {
	ctx := reforge.NewContextSet().WithNamedContextValues("prefab-api-key", map[string]any{"user-id": 1039})
	client, err := reforge.NewSdk(reforge.WithOfflineSources([]string{
		fmt.Sprintf("datafile://%s", "testdata/download.json"),
	}), reforge.WithGlobalContext(ctx))
	require.NoError(t, err)

	str, ok, err := client.GetStringValue("test.with.rule", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "targeted", str)

	// this tests client.GetConfigMatch which is used internally
	configMatch, err := client.GetConfigMatch("test.with.rule", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.Equal(t, "targeted", configMatch.OriginalMatch.GetString_())

	// Verify EnvId is populated correctly
	require.NotNil(t, configMatch.EnvId)
	assert.Equal(t, int64(308), *configMatch.EnvId)

	valueAny, ok, err := reforge.ExtractValue(configMatch.OriginalMatch)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "targeted", valueAny)

	// now show that shadowing the default context produces a different result
	contextSet := reforge.NewContextSet().
		WithNamedContextValues("prefab-api-key", map[string]interface{}{
			"user-id": 0,
		})

	str, ok, err = client.GetStringValue("test.with.rule", *contextSet)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "default", str)

	// Verify that when matching the default row (no projectEnvId), EnvId is nil
	configMatch, err = client.GetConfigMatch("test.with.rule", *contextSet)
	require.NoError(t, err)
	assert.Equal(t, "default", configMatch.OriginalMatch.GetString_())
	assert.Nil(t, configMatch.EnvId)
}

func TestSdkKeyNormalizationFromEnvVar(t *testing.T) {
	// Test that SDK key is properly normalized from env var during NewSdk()
	// This reproduces the customer bug where NewSdk() without WithSdkKey() didn't get live updates

	// Set env var
	testKey := "test-sdk-key-from-env"
	os.Setenv(options.SdkKeyEnvVar, testKey)
	defer os.Unsetenv(options.SdkKeyEnvVar)

	// Create client without explicitly passing SDK key
	// Use WithConfigs so we don't need a real API connection
	configs := map[string]interface{}{
		"test.key": "test-value",
	}

	client, err := reforge.NewSdk(
		reforge.WithConfigs(configs),
		reforge.WithInitializationTimeoutSeconds(1.0),
	)

	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify the SDK works
	value, ok, err := client.GetStringValue("test.key", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestSdkKeyExplicitTakesPrecedenceOverEnvVar(t *testing.T) {
	// Verify explicit SDK key takes precedence over env var

	// Set env var
	envKey := "env-sdk-key"
	os.Setenv(options.SdkKeyEnvVar, envKey)
	defer os.Unsetenv(options.SdkKeyEnvVar)

	// Create client with explicit SDK key
	explicitKey := "explicit-sdk-key"
	configs := map[string]interface{}{
		"test.key": "test-value",
	}

	client, err := reforge.NewSdk(
		reforge.WithSdkKey(explicitKey),
		reforge.WithConfigs(configs),
		reforge.WithInitializationTimeoutSeconds(1.0),
	)

	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify the SDK works
	value, ok, err := client.GetStringValue("test.key", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "test-value", value)
}

func TestSdkKeyMissingReturnsError(t *testing.T) {
	// Verify that missing SDK key returns an error when no env var is set

	// Make sure env var is not set
	os.Unsetenv(options.SdkKeyEnvVar)
	os.Unsetenv(options.LegacyApiKeyEnvVar)

	// Try to create client without SDK key - this should fail immediately during NewSdk()
	// because we normalize the SDK key before building stores
	_, err := reforge.NewSdk(
		reforge.WithInitializationTimeoutSeconds(1.0),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SDK key is not set")
}
