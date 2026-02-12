package reforge

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ReforgeHQ/sdk-go/internal/options"
)

// newNeverInitClient creates a client with valid config data but whose
// initializationComplete channel is left open, simulating an async source
// that never finishes loading.
func newNeverInitClient(t *testing.T, timeoutSec float64, failureMode options.OnInitializationFailure) *Client {
	t.Helper()

	configs := map[string]interface{}{"test.key": "value"}

	client, err := NewSdk(
		WithConfigs(configs),
		WithInitializationTimeoutSeconds(timeoutSec),
		WithOnInitializationFailure(failureMode),
	)
	require.NoError(t, err)

	// Replace the closed channel (memory store is synchronous) with a fresh
	// open one to simulate an async source that never calls finishedLoading.
	client.initializationComplete = make(chan struct{})
	client.closeInitializationCompleteOnce = sync.Once{}

	return client
}

// channelIsClosed returns true if the channel is closed (non-blocking check).
func channelIsClosed(ch chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func TestReturnError_ClosesChannelOnTimeout(t *testing.T) {
	client := newNeverInitClient(t, 1, options.ReturnError)

	// Precondition: channel is open
	assert.False(t, channelIsClosed(client.initializationComplete),
		"channel should be open before any calls")

	// First call triggers the timeout
	_, _, err := client.GetStringValue("test.key", *NewContextSet())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "initialization timeout")

	// Core assertion: channel is now closed so subsequent calls won't re-block
	assert.True(t, channelIsClosed(client.initializationComplete),
		"channel must be closed after first timeout in ReturnError mode")
}

func TestReturnNilMatch_ClosesChannelOnTimeout(t *testing.T) {
	client := newNeverInitClient(t, 1, options.ReturnNilMatch)

	assert.False(t, channelIsClosed(client.initializationComplete))

	// ReturnNilMatch falls through and resolves from the config store
	val, ok, err := client.GetStringValue("test.key", *NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "value", val)

	assert.True(t, channelIsClosed(client.initializationComplete),
		"channel must be closed after first timeout in ReturnNilMatch mode")
}

func TestReturnError_KeysClosesChannelOnTimeout(t *testing.T) {
	client := newNeverInitClient(t, 1, options.ReturnError)

	assert.False(t, channelIsClosed(client.initializationComplete))

	_, err := client.Keys()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "initialization timeout")

	assert.True(t, channelIsClosed(client.initializationComplete),
		"channel must be closed after first Keys() timeout")
}

func TestReturnError_SubsequentCallsFast(t *testing.T) {
	client := newNeverInitClient(t, 1, options.ReturnError)

	// First call: blocks for the full timeout
	_, _, err := client.GetStringValue("test.key", *NewContextSet())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "initialization timeout")

	// Second call via GetStringValue: should be nearly instant
	start := time.Now()
	val, ok, err := client.GetStringValue("test.key", *NewContextSet())
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "value", val)
	assert.Less(t, time.Since(start), 50*time.Millisecond,
		"second GetStringValue must not re-block")

	// Call via Keys(): also fast
	start = time.Now()
	keys, err := client.Keys()
	require.NoError(t, err)
	assert.NotNil(t, keys)
	assert.Less(t, time.Since(start), 50*time.Millisecond,
		"Keys() after timeout must not re-block")
}
