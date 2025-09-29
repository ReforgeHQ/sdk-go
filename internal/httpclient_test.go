package internal_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ReforgeHQ/sdk-go/internal"
	"github.com/ReforgeHQ/sdk-go/internal/options"
)

func TestLoadFromURIHandlesEmptyResponse(t *testing.T) {
	// Create a test server that returns an empty body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 200 OK with empty body
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create an HTTP client with test options
	opts := options.Options{
		SdkKey:  "test-key",
		APIURLs: []string{server.URL},
	}

	client, err := internal.BuildHTTPClient(opts)
	assert.NoError(t, err)

	// Try to load from the URI that returns empty body
	configs, err := client.LoadFromURI(server.URL, "test-key", 0)

	// Should return an error for empty response body
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty response body")
	assert.Nil(t, configs)
}

func TestLoadFromURIWithValidProtobuf(t *testing.T) {
	// This test would require a valid protobuf response
	// For now, we're just testing that empty responses are handled correctly
	t.Skip("Test requires valid protobuf test data")
}