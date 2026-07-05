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

func TestCreateComputeProfile(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/compute_profiles", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(ComputeProfile{
			Base: Base{ID: 1, Name: "small"},
			Name: "small",
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.CreateComputeProfile(context.Background(), &ComputeProfileRequest{Name: "small"})
	require.NoError(t, err)
	assert.Equal(t, "small", result.Name)
}

func TestReadComputeProfile(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/compute_profiles/1", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(ComputeProfile{
			Base: Base{ID: 1, Name: "small"},
			Name: "small",
			ComputeAttributes: []*ComputeAttribute{
				{Base: Base{ID: 9}, ComputeResourceID: 3, VMAttrs: json.RawMessage(`{"cpus":1}`)},
			},
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.ReadComputeProfile(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, result.ComputeAttributes, 1)
	assert.Equal(t, 3, result.ComputeAttributes[0].ComputeResourceID)
}

func TestUpdateComputeProfile(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/compute_profiles/1", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)
		require.NoError(t, json.NewEncoder(w).Encode(ComputeProfile{
			Base: Base{ID: 1, Name: "renamed"},
			Name: "renamed",
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.UpdateComputeProfile(context.Background(), 1, &ComputeProfileRequest{Name: "renamed"})
	require.NoError(t, err)
	assert.Equal(t, "renamed", result.Name)
}

func TestDeleteComputeProfile(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/compute_profiles/1", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	err := client.DeleteComputeProfile(context.Background(), 1)
	require.NoError(t, err)
}

func TestFindComputeProfileByName(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(QueryResponse{
			Results: []json.RawMessage{[]byte(`{"id":1,"name":"small"}`)},
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.FindComputeProfileByName(context.Background(), "small")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "small", result.Name)
}

func TestFindComputeProfileByName_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(QueryResponse{Results: []json.RawMessage{}}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.FindComputeProfileByName(context.Background(), "missing")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestCreateComputeAttribute(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/compute_profiles/1/compute_resources/3/compute_attributes", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(ComputeAttribute{
			Base:              Base{ID: 9},
			ComputeResourceID: 3,
			VMAttrs:           json.RawMessage(`{"cpus":1}`),
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.CreateComputeAttribute(context.Background(), 1, 3, json.RawMessage(`{"cpus":1}`))
	require.NoError(t, err)
	assert.Equal(t, 3, result.ComputeResourceID)
}

func TestUpdateComputeAttribute(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/compute_profiles/1/compute_resources/3/compute_attributes/9", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)
		require.NoError(t, json.NewEncoder(w).Encode(ComputeAttribute{
			Base:              Base{ID: 9},
			ComputeResourceID: 3,
			VMAttrs:           json.RawMessage(`{"cpus":2}`),
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.UpdateComputeAttribute(context.Background(), 1, 3, 9, json.RawMessage(`{"cpus":2}`))
	require.NoError(t, err)
	assert.Equal(t, json.RawMessage(`{"cpus":2}`), result.VMAttrs)
}

func TestDeleteComputeAttribute(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/compute_profiles/1/compute_resources/3/compute_attributes/9", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	err := client.DeleteComputeAttribute(context.Background(), 1, 3, 9)
	require.NoError(t, err)
}
