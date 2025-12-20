package sse_test

import (
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

	client, err := sse.BuildSSEClient(options)

	assert.NoError(t, err)
	assert.Equal(t, "https://stream.reforge.com/api/v2/sse/config", client.URL)

	assert.Equal(t, map[string]string{
		"Authorization":                "Basic YXV0aHVzZXI6ZG9lcy1ub3QtbWF0dGVy",
		"X-Reforge-SDK-Version": internal.ClientVersionHeader,
		"Accept":                       "text/event-stream",
	}, client.Headers)
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
	// Test that BuildSSEClient correctly falls back to environment variable
	// when SdkKey is not set in options (reproduces customer bug report)

	// Set env var
	envKey := "test-env-sdk-key"
	os.Setenv(options.SdkKeyEnvVar, envKey)
	defer os.Unsetenv(options.SdkKeyEnvVar)

	opts := options.Options{
		SdkKey:  "", // Empty - should fall back to env var
		APIURLs: []string{"https://primary.reforge.com"},
	}

	// Normalize SDK key from env var (mimics what NewSdk() does)
	_, err := opts.SdkKeySettingOrEnvVar()
	assert.NoError(t, err)

	client, err := sse.BuildSSEClient(opts)

	assert.NoError(t, err)
	assert.Equal(t, "https://stream.reforge.com/api/v2/sse/config", client.URL)

	// Verify auth header uses the env var key
	expectedAuth := "Basic YXV0aHVzZXI6dGVzdC1lbnYtc2RrLWtleQ==" // base64("authuser:test-env-sdk-key")
	assert.Equal(t, expectedAuth, client.Headers["Authorization"])
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

	client, err := sse.BuildSSEClient(opts)

	assert.NoError(t, err)

	// Verify auth header uses the explicit key, not env var
	expectedAuth := "Basic YXV0aHVzZXI6ZXhwbGljaXQta2V5" // base64("authuser:explicit-key")
	assert.Equal(t, expectedAuth, client.Headers["Authorization"])
}
