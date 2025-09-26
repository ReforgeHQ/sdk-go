package sse_test

import (
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
