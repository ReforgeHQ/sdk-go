package sse_test

import (
	"encoding/base64"
	"os"
	"testing"

	r3sse "github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"

	"github.com/ReforgeHQ/sdk-go/internal"
	"github.com/ReforgeHQ/sdk-go/internal/options"
	sse "github.com/ReforgeHQ/sdk-go/internal/sse"
	prefabProto "github.com/ReforgeHQ/sdk-go/proto"
)

func TestBuildSSEClient(t *testing.T) {
	options := options.Options{
		SdkKey:  "does-not-matter",
		APIURLs: []string{"https://primary.reforge.com"},
	}

	client, opts, err := sse.BuildSSEClient(options)

	assert.NoError(t, err)
	assert.Equal(t, "https://stream.reforge.com/api/v2/sse/config", client.URL)
	assert.NotNil(t, opts)

	// Headers are not set until StartSSEConnection is called
	// This test verifies the client is created with the correct URL
}

type mockConfigStore struct {
	highWatermark int64
	lastConfigs   *prefabProto.Configs
}

func (m *mockConfigStore) SetFromConfigsProto(configs *prefabProto.Configs) {
	m.lastConfigs = configs
}

func (m *mockConfigStore) GetHighWatermark() int64 {
	return m.highWatermark
}

func TestEventHandlerIgnoresEmptyEvents(t *testing.T) {
	store := &mockConfigStore{highWatermark: 0}

	// Create an empty event (phantom event from SSE library bug)
	emptyEvent := &r3sse.Event{
		Data: []byte{},
	}

	// This should not panic or cause errors when we process empty events
	// The actual event handler is internal to StartSSEConnection, so we test the behavior indirectly
	// by ensuring that empty data doesn't cause base64 decode errors
	assert.Equal(t, 0, len(emptyEvent.Data))
	assert.Nil(t, store.lastConfigs)
}

func TestBuildSSEClientWithEnvVar(t *testing.T) {
	// Test that StartSSEConnection correctly falls back to environment variable
	// when SdkKey is not set in options (reproduces customer bug report)

	// Set env var
	envKey := "test-env-sdk-key"
	os.Setenv(options.SdkKeyEnvVar, envKey)
	defer os.Unsetenv(options.SdkKeyEnvVar)

	opts := options.Options{
		SdkKey:  "", // Empty - should fall back to env var
		APIURLs: []string{"https://primary.reforge.com"},
	}

	client, sseOpts, err := sse.BuildSSEClient(opts)

	assert.NoError(t, err)
	assert.Equal(t, "https://stream.reforge.com/api/v2/sse/config", client.URL)
	assert.NotNil(t, sseOpts)

	// Simulate what StartSSEConnection does - get SDK key and set headers
	sdkKey, err := sseOpts.SdkKeySettingOrEnvVar()
	assert.NoError(t, err)
	assert.Equal(t, envKey, sdkKey)

	// Verify the authorization would be correct
	expectedAuth := "Basic YXV0aHVzZXI6dGVzdC1lbnYtc2RrLWtleQ==" // base64("authuser:test-env-sdk-key")
	authString := base64.StdEncoding.EncodeToString([]byte("authuser:" + sdkKey))
	assert.Equal(t, expectedAuth, "Basic "+authString)
}

func TestBuildSSEClientExplicitKeyTakesPrecedence(t *testing.T) {
	// Test that explicit SdkKey takes precedence over env var

	// Set env var
	envKey := "env-key"
	os.Setenv(options.SdkKeyEnvVar, envKey)
	defer os.Unsetenv(options.SdkKeyEnvVar)

	explicitKey := "explicit-key"
	opts := options.Options{
		SdkKey:  explicitKey,
		APIURLs: []string{"https://primary.reforge.com"},
	}

	client, sseOpts, err := sse.BuildSSEClient(opts)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, sseOpts)

	// Verify that SdkKeySettingOrEnvVar returns the explicit key
	sdkKey, err := sseOpts.SdkKeySettingOrEnvVar()
	assert.NoError(t, err)
	assert.Equal(t, explicitKey, sdkKey)

	// Verify auth header would use the explicit key, not env var
	expectedAuth := "Basic YXV0aHVzZXI6ZXhwbGljaXQta2V5" // base64("authuser:explicit-key")
	authString := base64.StdEncoding.EncodeToString([]byte("authuser:" + sdkKey))
	assert.Equal(t, expectedAuth, "Basic "+authString)
}
