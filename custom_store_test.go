package reforge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	reforge "github.com/ReforgeHQ/sdk-go"
	"github.com/ReforgeHQ/sdk-go/internal/options"
	"github.com/ReforgeHQ/sdk-go/internal/utils"
	prefabProto "github.com/ReforgeHQ/sdk-go/proto"
)

// TestCustomStore implements ConfigStoreGetter for testing
type TestCustomStore struct {
	configs map[string]*prefabProto.Config
}

func NewTestCustomStore() *TestCustomStore {
	return &TestCustomStore{
		configs: make(map[string]*prefabProto.Config),
	}
}

func (s *TestCustomStore) GetConfig(key string) (*prefabProto.Config, bool) {
	config, exists := s.configs[key]
	return config, exists
}

func (s *TestCustomStore) Keys() []string {
	keys := make([]string, 0, len(s.configs))
	for k := range s.configs {
		keys = append(keys, k)
	}
	return keys
}

func (s *TestCustomStore) GetContextValue(propertyName string) (interface{}, bool) {
	return nil, false
}

func (s *TestCustomStore) GetProjectEnvID() int64 {
	return 123
}

func (s *TestCustomStore) AddConfig(key string, value string) {
	configValue, _ := utils.Create(value)

	s.configs[key] = &prefabProto.Config{
		Id:  1,
		Key: key,
		Rows: []*prefabProto.ConfigRow{
			{
				Values: []*prefabProto.ConditionalValue{
					{
						Value: configValue,
					},
				},
			},
		},
	}
}

func TestWithCustomStore(t *testing.T) {
	customStore := NewTestCustomStore()
	customStore.AddConfig("custom.key", "custom value")

	client, err := reforge.NewSdk(
		reforge.WithCustomStore(customStore),
		reforge.WithOfflineSources([]string{}), // No default sources
		reforge.WithContextTelemetryMode(options.ContextTelemetryModes.None))

	require.NoError(t, err)

	str, ok, err := client.GetStringValue("custom.key", *reforge.NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "custom value", str)
}

func TestWithCustomStoreAndConfigs(t *testing.T) {
	customStore := NewTestCustomStore()
	customStore.AddConfig("custom.key", "custom value")

	configs := map[string]interface{}{
		"memory.key": "memory value",
	}

	_, err := reforge.NewSdk(
		reforge.WithCustomStore(customStore),
		reforge.WithConfigs(configs))

	require.Error(t, err)
	assert.Equal(t, "cannot use WithConfigs with custom stores", err.Error())
}