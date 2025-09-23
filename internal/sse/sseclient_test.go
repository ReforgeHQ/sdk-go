package sse_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ReforgeHQ/sdk-go/internal"
	"github.com/ReforgeHQ/sdk-go/internal/options"
	sse "github.com/ReforgeHQ/sdk-go/internal/sse"
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
