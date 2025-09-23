package reforge_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	reforge "github.com/ReforgeHQ/sdk-go/pkg"
	"github.com/ReforgeHQ/sdk-go/pkg/internal/options"
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

	str, ok, err := client.GetStringValue("string.key", reforge.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "value", str)

	i, ok, err := client.GetIntValue("int.key", reforge.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, int64(42), i)

	b, ok, err := client.GetBoolValue("bool.key", reforge.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, b)

	f, ok, err := client.GetFloatValue("float.key", reforge.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.InDelta(t, 3.14, f, 0.0001)

	slice, ok, err := client.GetStringSliceValue("slice.key", reforge.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b", "c"}, slice)

	json, ok, err := client.GetJSONValue("json.key", reforge.ContextSet{})
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
		reforge.WithOfflineSources([]string{"dump://testdata/example.dump"}),
		reforge.WithContextTelemetryMode(options.ContextTelemetryModes.None),
		reforge.WithProjectEnvID(8),
	)

	require.Error(t, err)
	assert.Equal(t, "cannot use WithConfigs with other sources", err.Error())
}

func TestWithAJSONConfigDump(t *testing.T) {
	t.Setenv("PREFAB_DATAFILE", "testdata/download.json")

	client, err := reforge.NewSdk(reforge.WithContextTelemetryMode(options.ContextTelemetryModes.None))
	require.NoError(t, err)

	str, ok, err := client.GetStringValue("my.test.string", reforge.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "hello world", str)

	boolean, ok, err := client.GetBoolValue("flag.list.environments", reforge.ContextSet{})
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
	boolean, ok, err = client.WithContext(contextSet).GetBoolValue("flag.list.environments", reforge.ContextSet{})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, boolean)

	// This one is deleted
	_, ok, err = client.GetLogLevelStringValue("log-level", reforge.ContextSet{})
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
}
