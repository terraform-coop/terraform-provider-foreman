package generated

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadForemanSetting(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/settings/append_domain_name_for_hosts", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(ForemanSetting{
			ID: "append_domain_name_for_hosts", Value: json.RawMessage("true"),
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	result, err := client.ReadForemanSetting(context.Background(), "append_domain_name_for_hosts")
	require.NoError(t, err)
	assert.Equal(t, "append_domain_name_for_hosts", result.ID)
	assert.Equal(t, "true", string(result.Value))
}

func TestUpdateForemanSetting(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/settings/http_proxy", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)
		var body map[string]map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "http://proxy.example.com", body["setting"]["value"])
		require.NoError(t, json.NewEncoder(w).Encode(ForemanSetting{
			ID: "http_proxy", Value: json.RawMessage(`"http://proxy.example.com"`),
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	result, err := client.UpdateForemanSetting(context.Background(), "http_proxy", &ForemanSettingRequest{Value: "http://proxy.example.com"})
	require.NoError(t, err)
	assert.Equal(t, `"http://proxy.example.com"`, string(result.Value))
}

func TestQueryForemanSetting(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/settings", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(QueryResponse{
			Results: []json.RawMessage{[]byte(`{"id":"http_proxy","name":"http_proxy","value":"x"}`)},
		}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})
	result, err := client.QueryForemanSetting(context.Background(), "http_proxy")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "http_proxy", result.ID)
}
