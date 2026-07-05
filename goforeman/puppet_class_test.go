package goforeman

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadForemanPuppetClass(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/puppetclasses/2", r.URL.Path)
		_, _ = w.Write([]byte(`{"id": 2, "name": "testing"}`))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	result, err := client.ReadForemanPuppetClass(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, "testing", result.Name)
}

// TestQueryForemanPuppetClass_GroupedByEnvironment validates the real,
// environment-grouped response shape (confirmed against a real Foreman
// 3.1.2 server) - not the flat {"results": [...]} array every other
// resource's index endpoint uses.
func TestQueryForemanPuppetClass_GroupedByEnvironment(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/puppetclasses", r.URL.Path)
		_, _ = w.Write([]byte(`{"results": {"production": [{"id": 2, "name": "testing"}]}}`))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	result, err := client.QueryForemanPuppetClass(context.Background(), "testing")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.ID)
}

func TestQueryForemanPuppetClass_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results": {"production": [{"id": 2, "name": "other"}]}}`))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	result, err := client.QueryForemanPuppetClass(context.Background(), "testing")
	require.NoError(t, err)
	assert.Nil(t, result)
}
