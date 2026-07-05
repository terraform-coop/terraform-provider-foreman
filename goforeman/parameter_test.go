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

func TestCreateParameter(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/hosts/5/parameters", r.URL.Path)
		var body map[string]map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "ntp_server", body["parameter"]["name"])
		w.WriteHeader(http.StatusCreated)
		require.NoError(t, json.NewEncoder(w).Encode(Parameter{
			Base:          Base{ID: 42, Name: "ntp_server"},
			Value:         json.RawMessage(`"pool.ntp.org"`),
			ParameterType: "string",
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.CreateParameter(context.Background(), "hosts", 5, &ParameterRequest{
		Name: "ntp_server", Value: "pool.ntp.org", ParameterType: "string",
	})
	require.NoError(t, err)
	assert.Equal(t, 42, result.ID)
	assert.Equal(t, `"pool.ntp.org"`, string(result.Value))
}

func TestReadParameter(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/hostgroups/6/parameters/42", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(Parameter{
			Base:  Base{ID: 42, Name: "ntp_server"},
			Value: json.RawMessage(`true`),
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.ReadParameter(context.Background(), "hostgroups", 6, 42)
	require.NoError(t, err)
	assert.Equal(t, "true", string(result.Value))
}

func TestUpdateParameter(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/domains/1/parameters/9", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)
		require.NoError(t, json.NewEncoder(w).Encode(Parameter{Base: Base{ID: 9}}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	_, err := client.UpdateParameter(context.Background(), "domains", 1, 9, &ParameterRequest{})
	require.NoError(t, err)
}

func TestDeleteParameter(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/subnets/3/parameters/7", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	err := client.DeleteParameter(context.Background(), "subnets", 3, 7)
	require.NoError(t, err)
}

func TestFindParameterByName(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/organizations/2/parameters", r.URL.Path)
		assert.Equal(t, `name="ntp_server"`, r.URL.Query().Get("search"))
		require.NoError(t, json.NewEncoder(w).Encode(QueryResponse{
			Results: []json.RawMessage{[]byte(`{"id":42,"name":"ntp_server"}`)},
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.FindParameterByName(context.Background(), "organizations", 2, "ntp_server")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 42, result.ID)
}

func TestFindParameterByName_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(QueryResponse{Results: []json.RawMessage{}}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	result, err := client.FindParameterByName(context.Background(), "hosts", 1, "missing")
	require.NoError(t, err)
	assert.Nil(t, result)
}
