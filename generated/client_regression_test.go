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

func TestClient_TaxonomyWrapping(t *testing.T) {
	t.Parallel()

	var receivedBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedBody))
		w.WriteHeader(200)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{"id": 1, "name": "test"}))
	}))
	defer srv.Close()

	client := NewClient(
		parseURL(srv.URL),
		ClientCredentials{},
		ClientConfig{OrganizationID: 5, LocationID: 10},
	)

	// Post should add taxonomy to wrapped body
	var resp ForemanDomain
	err := client.Post(context.Background(), "domains", "domain",
		&ForemanDomainRequest{Name: "test"}, &resp)
	require.NoError(t, err)

	// Check taxonomy was added to outer map (alongside wrapper key)
	assert.Equal(t, float64(5), receivedBody["organization_id"])
	assert.Equal(t, float64(10), receivedBody["location_id"])
}

func TestClient_WrapperKey(t *testing.T) {
	t.Parallel()

	var receivedBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedBody))
		w.WriteHeader(200)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{"id": 1, "name": "test"}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})

	var resp ForemanDomain
	err := client.Post(context.Background(), "domains", "domain",
		&ForemanDomainRequest{Name: "test"}, &resp)
	require.NoError(t, err)

	// Body should be wrapped in {"domain": {...}}
	_, hasDomain := receivedBody["domain"]
	assert.True(t, hasDomain, "expected 'domain' wrapper key in request body")

	// Test with empty wrapper key (Katello pattern)
	receivedBody2 := map[string]interface{}{}
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedBody2))
		w.WriteHeader(200)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{"id": 1, "name": "test"}))
	}))
	defer srv2.Close()

	client2 := NewClient(parseURL(srv2.URL), ClientCredentials{}, ClientConfig{})
	err = client2.Post(context.Background(), "/katello/api/products", "",
		&ForemanKatelloProductRequest{Name: "test"}, &resp)
	require.NoError(t, err)

	// Body should NOT be wrapped when wrapperKey is empty
	_, hasDomain2 := receivedBody2["domain"]
	assert.False(t, hasDomain2, "expected no wrapper key when empty")
	assert.Equal(t, "test", receivedBody2["name"])
}

func TestClient_404Handling(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, err := w.Write([]byte(`{"error":{"message":"Resource not found"}}`))
		require.NoError(t, err)
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL), ClientCredentials{}, ClientConfig{})

	// Get should return HTTPError with IsNotFound=true
	err := client.Get(context.Background(), "domains/999", nil)
	require.Error(t, err)
	assert.True(t, IsNotFoundError(err), "expected IsNotFoundError for 404")

	// Delete with 404 should NOT be an error (already deleted)
	err = client.Delete(context.Background(), "domains/999")
	assert.NoError(t, err, "404 on delete should not be an error")
}

func TestClient_IsNotFoundError(t *testing.T) {
	t.Parallel()

	assert.False(t, IsNotFoundError(nil))
	assert.False(t, IsNotFoundError(assert.AnError))

	httpErr := &HTTPError{StatusCode: 404}
	assert.True(t, IsNotFoundError(httpErr))

	httpErr2 := &HTTPError{StatusCode: 500}
	assert.False(t, IsNotFoundError(httpErr2))
}
