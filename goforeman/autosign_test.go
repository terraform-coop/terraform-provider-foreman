package goforeman

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateForemanAutosign(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/smart_proxies/5/autosign", r.URL.Path)
		var body map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "*.example.com", body["id"])
		w.WriteHeader(http.StatusCreated)
		require.NoError(t, json.NewEncoder(w).Encode(ForemanAutosign{ID: "*.example.com"}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	result, err := client.CreateForemanAutosign(context.Background(), 5, "*.example.com")
	require.NoError(t, err)
	assert.Equal(t, "*.example.com", result.ID)
}

func TestDeleteForemanAutosign(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/smart_proxies/5/autosign/%2A.example.com", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	err := client.DeleteForemanAutosign(context.Background(), 5, "*.example.com")
	require.NoError(t, err)
}

func TestReadForemanAutosign_Found(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/smart_proxies/5/autosign", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(QueryResponse{
			Results: []json.RawMessage{
				[]byte(`{"id":"other.example.com"}`),
				[]byte(`{"id":"*.example.com"}`),
			},
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	result, err := client.ReadForemanAutosign(context.Background(), 5, "*.example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "*.example.com", result.ID)
}

func TestReadForemanAutosign_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(QueryResponse{Results: []json.RawMessage{}}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	result, err := client.ReadForemanAutosign(context.Background(), 5, "*.example.com")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestReadForemanAutosign_Error(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	_, err := client.ReadForemanAutosign(context.Background(), 5, "*.example.com")
	assert.Error(t, err)
}
