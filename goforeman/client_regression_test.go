package goforeman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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
		WithTaxonomy(5, 10),
	)

	// Post should add taxonomy to wrapped body
	var resp Domain
	err := client.Post(context.Background(), "domains", "domain",
		&DomainRequest{Name: "test"}, &resp)
	require.NoError(t, err)

	// Foreman only honors organization_id/location_id nested inside the
	// resource's own wrapped hash, not as siblings of it - see issue #179.
	domain, ok := receivedBody["domain"].(map[string]interface{})
	require.True(t, ok, "expected 'domain' wrapper key in request body")
	assert.Equal(t, float64(5), domain["organization_id"])
	assert.Equal(t, float64(10), domain["location_id"])
	assert.NotContains(t, receivedBody, "organization_id", "organization_id must not be a sibling of the wrapped hash")
	assert.NotContains(t, receivedBody, "location_id", "location_id must not be a sibling of the wrapped hash")
}

func TestClient_TaxonomyWrapping_NoWrapperKey(t *testing.T) {
	t.Parallel()

	var receivedBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedBody))
		w.WriteHeader(200)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{"id": "test"}))
	}))
	defer srv.Close()

	client := NewClient(
		parseURL(srv.URL),
		WithTaxonomy(5, 10),
	)

	// Some endpoints (e.g. autosign) take no wrapper key at all - taxonomy
	// must still land in the one and only body map, not get dropped.
	var resp map[string]interface{}
	err := client.Post(context.Background(), "test-endpoint", "", map[string]string{"id": "test"}, &resp)
	require.NoError(t, err)
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

	client := NewClient(parseURL(srv.URL))

	var resp Domain
	err := client.Post(context.Background(), "domains", "domain",
		&DomainRequest{Name: "test"}, &resp)
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

	client2 := NewClient(parseURL(srv2.URL))
	err = client2.Post(context.Background(), "/katello/api/products", "",
		&KatelloProductRequest{Name: "test"}, &resp)
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

	client := NewClient(parseURL(srv.URL))

	// Get should return HTTPError with IsNotFound=true
	err := client.Get(context.Background(), "domains/999", nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound), "expected ErrNotFound match for 404")

	// Delete with 404 should NOT be an error (already deleted)
	err = client.Delete(context.Background(), "domains/999")
	assert.NoError(t, err, "404 on delete should not be an error")
}

func TestErrNotFound_ErrorsIs(t *testing.T) {
	t.Parallel()

	assert.False(t, errors.Is(nil, ErrNotFound))
	assert.False(t, errors.Is(ErrNotFound, assert.AnError)) // unrelated errors never match
	assert.True(t, errors.Is(&HTTPError{StatusCode: 404}, ErrNotFound))
	assert.False(t, errors.Is(&HTTPError{StatusCode: 500}, ErrNotFound))
	// wrapped errors still match
	assert.True(t, errors.Is(fmt.Errorf("reading host: %w", &HTTPError{StatusCode: 404}), ErrNotFound))
}

func TestNewClient_Options(t *testing.T) {
	t.Parallel()

	u := url.URL{Scheme: "https", Host: "foreman.example.com"}

	t.Run("defaults", func(t *testing.T) {
		c := NewClient(u)
		assert.Equal(t, defaultRequestTimeout, c.httpClient.Timeout)
		assert.False(t, c.config.TLSInsecure)
	})

	t.Run("all options", func(t *testing.T) {
		c := NewClient(u,
			WithBasicAuth("admin", "secret"),
			WithTLSInsecure(),
			WithTaxonomy(3, 7),
			WithTimeout(5*time.Minute),
		)
		assert.Equal(t, "admin", c.credentials.Username)
		assert.True(t, c.config.TLSInsecure)
		assert.Equal(t, 3, c.config.OrganizationID)
		assert.Equal(t, 7, c.config.LocationID)
		assert.Equal(t, 5*time.Minute, c.httpClient.Timeout)
	})

	t.Run("custom http client bypasses transport construction", func(t *testing.T) {
		hc := &http.Client{Timeout: time.Second}
		c := NewClient(u, WithHTTPClient(hc))
		assert.Same(t, hc, c.httpClient)
	})
}

func TestListAll_Pagination(t *testing.T) {
	t.Parallel()

	// Two pages of results; assert both are collected and the loop stops.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "100", r.URL.Query().Get("per_page"))
		var results []json.RawMessage
		switch r.URL.Query().Get("page") {
		case "1":
			for i := 0; i < 100; i++ {
				results = append(results, json.RawMessage(fmt.Sprintf(`{"id":%d}`, i)))
			}
		case "2":
			results = []json.RawMessage{json.RawMessage(`{"id":100}`)}
		default:
			t.Errorf("unexpected page %q requested", r.URL.Query().Get("page"))
		}
		require.NoError(t, json.NewEncoder(w).Encode(QueryResponse{Subtotal: 101, Results: results}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	domains, err := client.ListDomains(context.Background())
	require.NoError(t, err)
	assert.Len(t, domains, 101)
	assert.Equal(t, 100, domains[100].ID)
}

func TestListAll_EmptyPageGuard(t *testing.T) {
	t.Parallel()

	// A server reporting an inconsistent (too large) Subtotal with an empty
	// results page must terminate, not loop forever.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(QueryResponse{Subtotal: 9999, Results: nil}))
	}))
	defer srv.Close()

	client := NewClient(parseURL(srv.URL))
	domains, err := client.ListDomains(context.Background())
	require.NoError(t, err)
	assert.Empty(t, domains)
}
