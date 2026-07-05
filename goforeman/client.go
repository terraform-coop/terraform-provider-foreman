package goforeman

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dpotapov/go-spnego"
)

const (
	APIVersion       = "2"
	APIURLPrefix     = "/api"
	KatelloURLPrefix = "/katello/api"
	TasksURLPrefix   = "/foreman_tasks/api"
	PuppetURLPrefix  = "/foreman_puppet/api"

	defaultRequestTimeout = 60 * time.Second
)

type Client struct {
	serverURL   url.URL
	httpClient  *http.Client
	credentials clientCredentials
	config      clientConfig
	timeout     time.Duration
}

type clientCredentials struct {
	Username string
	Password string
}

type clientConfig struct {
	TLSInsecure    bool
	NegotiateAuth  bool
	OrganizationID int
	LocationID     int
}

// Option configures a Client at construction time.
type Option func(*Client)

// WithBasicAuth authenticates every request with HTTP basic auth.
func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		c.credentials = clientCredentials{Username: username, Password: password}
	}
}

// WithNegotiateAuth authenticates via the HTTP negotiate (SPNEGO/Kerberos)
// mechanism instead of basic auth.
func WithNegotiateAuth() Option {
	return func(c *Client) { c.config.NegotiateAuth = true }
}

// WithTLSInsecure skips TLS certificate verification.
func WithTLSInsecure() Option {
	return func(c *Client) { c.config.TLSInsecure = true }
}

// WithTaxonomy scopes every request to the given organization and location:
// the IDs are injected into create/update bodies (nested inside the
// resource's own wrapped hash, the only placement Foreman honors) and, for
// org-scoped Katello endpoints, into the URL. Pass 0 to leave a dimension
// unscoped.
func WithTaxonomy(organizationID, locationID int) Option {
	return func(c *Client) {
		c.config.OrganizationID = organizationID
		c.config.LocationID = locationID
	}
}

// WithTimeout overrides the per-request timeout (default 60s). Long-running
// synchronous Foreman calls (large host creates, slow Katello operations
// that aren't wrapped in an async task) may need more.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// WithHTTPClient supplies a fully custom *http.Client, bypassing the
// transport the other options would construct (their TLS/negotiate/timeout
// settings then no longer apply - configure the supplied client yourself).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

func NewClient(serverURL url.URL, opts ...Option) *Client {
	c := &Client{
		serverURL: serverURL,
		timeout:   defaultRequestTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: c.config.TLSInsecure,
		}
		c.httpClient = &http.Client{Timeout: c.timeout}
		if c.config.NegotiateAuth {
			c.httpClient.Transport = &spnego.Transport{Transport: http.Transport{TLSClientConfig: tlsCfg}}
		} else {
			c.httpClient.Transport = &http.Transport{TLSClientConfig: tlsCfg}
		}
	}
	return c
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	reqURL := c.serverURL
	ep := endpoint

	// Every Query<Resource>/search-by-name endpoint is built as
	// "path?search=...", but url.URL treats "?" found inside .Path as a
	// literal character to percent-encode (%3F), not a query separator -
	// assigning the whole thing to reqURL.Path below would send the "?"
	// and everything after it as part of the request path, never as a
	// real query string, and Foreman's router 404s on it. Split it out
	// and assign to RawQuery (already in encoded form from the caller)
	// before the path even gets built.
	var rawQuery string
	if i := strings.IndexByte(ep, '?'); i >= 0 {
		ep, rawQuery = ep[:i], ep[i+1:]
	}

	switch {
	// "katello/", "puppet/" (with the trailing slash) rather than a bare
	// prefix match: a bare "puppet" match previously caught the core
	// Foreman "puppetclasses" endpoint (no such namespacing intended) and
	// misrouted it to the Puppet plugin's own URL prefix entirely -
	// confirmed against a real server, every puppetclasses lookup 404'd.
	case strings.HasPrefix(ep, "katello/"):
		reqURL.Path = KatelloURLPrefix + strings.TrimPrefix(ep, "katello")
	case strings.HasPrefix(ep, "/katello/api"):
		reqURL.Path = ep
	case strings.HasPrefix(ep, "puppet/"):
		reqURL.Path = PuppetURLPrefix + strings.TrimPrefix(ep, "puppet")
	case strings.HasPrefix(ep, "foreman_tasks"):
		reqURL.Path = ep
	default:
		if strings.HasPrefix(ep, "/") {
			reqURL.Path = APIURLPrefix + ep
		} else {
			reqURL.Path = APIURLPrefix + "/" + ep
		}
	}
	reqURL.RawQuery = rawQuery

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), reqURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("User-Agent", "terraform-provider-foreman")
	req.Header.Add("Accept", "application/json,version="+APIVersion)
	req.Header.Add("Content-Type", "application/json")
	if !c.config.NegotiateAuth {
		req.SetBasicAuth(c.credentials.Username, c.credentials.Password)
	}
	return req, nil
}

func (c *Client) send(req *http.Request) (int, []byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return -1, nil, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("reading response: %w", err)
	}
	return resp.StatusCode, body, nil
}

func (c *Client) do(ctx context.Context, method, endpoint string, reqBody, respObj interface{}) error {
	var bodyReader io.Reader
	if reqBody != nil {
		payload, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshaling request: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := c.newRequest(ctx, method, endpoint, bodyReader)
	if err != nil {
		return err
	}

	statusCode, respBody, err := c.send(req)
	if err != nil {
		return err
	}

	if statusCode == 202 {
		var task Task
		if err := json.Unmarshal(respBody, &task); err != nil {
			return fmt.Errorf("parsing async task: %w", err)
		}
		if task.Pending {
			finished, err := c.waitForKatelloTask(ctx, task.ID)
			if err != nil {
				return err
			}
			if finished.Result != "success" {
				return fmt.Errorf("async task failed: %s (result: %s)", finished.Label, finished.Result)
			}
		}
		// For 202, the response body may be the created/updated resource
		if respObj != nil && len(respBody) > 0 {
			return json.Unmarshal(respBody, respObj)
		}
		return nil
	}

	if statusCode < 200 || statusCode > 299 {
		return &HTTPError{
			Endpoint:   req.URL.String(),
			StatusCode: statusCode,
			Body:       string(respBody),
		}
	}

	if respObj != nil {
		return json.Unmarshal(respBody, respObj)
	}
	return nil
}

// addTaxonomy injects organization_id/location_id directly into the
// resource's own request body (reqBody - typically a *FooRequest struct
// pointer, round-tripped through JSON to merge as a plain map). Foreman
// only honors these fields as part of the resource's own permitted params
// (e.g. {"host": {"organization_id": 1, ...}}); the same fields placed as
// siblings of the wrapped hash (e.g. {"host": {...}, "organization_id": 1})
// are silently ignored server-side and creation fails with "Organization
// can't be blank" - confirmed against a real Foreman server, see
// https://github.com/terraform-coop/terraform-provider-foreman/issues/179.
func (c *Client) addTaxonomy(reqBody interface{}) interface{} {
	if c.config.OrganizationID <= 0 && c.config.LocationID <= 0 {
		return reqBody
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return reqBody
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return reqBody
	}
	if c.config.OrganizationID > 0 {
		m["organization_id"] = c.config.OrganizationID
	}
	if c.config.LocationID > 0 {
		m["location_id"] = c.config.LocationID
	}
	return m
}

func (c *Client) Get(ctx context.Context, endpoint string, respObj interface{}) error {
	return c.do(ctx, http.MethodGet, endpoint, nil, respObj)
}

func (c *Client) Post(ctx context.Context, endpoint, wrapperKey string, reqBody, respObj interface{}) error {
	wrapped := c.wrapRequestBody(wrapperKey, c.addTaxonomy(reqBody))
	return c.do(ctx, http.MethodPost, endpoint, wrapped, respObj)
}

func (c *Client) Put(ctx context.Context, endpoint, wrapperKey string, reqBody, respObj interface{}) error {
	wrapped := c.wrapRequestBody(wrapperKey, c.addTaxonomy(reqBody))
	return c.do(ctx, http.MethodPut, endpoint, wrapped, respObj)
}

func (c *Client) Delete(ctx context.Context, endpoint string) error {
	err := c.do(ctx, http.MethodDelete, endpoint, nil, nil)
	if err != nil && errors.Is(err, ErrNotFound) {
		return nil // Already deleted is not an error
	}
	return err
}

func (c *Client) wrapRequestBody(wrapperKey string, reqBody interface{}) interface{} {
	if wrapperKey == "" {
		return reqBody
	}
	return map[string]interface{}{wrapperKey: reqBody}
}

// ErrNotFound is matched by errors.Is for any HTTPError with a 404 status,
// so callers can distinguish "the resource is gone" from real failures:
//
//	if errors.Is(err, goforeman.ErrNotFound) { ... }
var ErrNotFound = errors.New("not found")

type HTTPError struct {
	Endpoint   string
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error: endpoint=%s status=%d body=%s", e.Endpoint, e.StatusCode, e.Body)
}

// Is makes errors.Is(err, ErrNotFound) true for 404 responses.
func (e *HTTPError) Is(target error) bool {
	return target == ErrNotFound && e.StatusCode == http.StatusNotFound
}

func (c *Client) waitForKatelloTask(ctx context.Context, taskID int) (*Task, error) {
	endpoint := fmt.Sprintf("/foreman_tasks/api/tasks/%d", taskID)
	for i := 0; i < 10; i++ {
		var task Task
		if err := c.Get(ctx, endpoint, &task); err != nil {
			return nil, fmt.Errorf("polling task %d: %w", taskID, err)
		}
		if !task.Pending {
			return &task, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("task %d polling cancelled: %w", taskID, ctx.Err())
		case <-time.After(time.Duration(i+1) * time.Second):
		}
	}
	return nil, fmt.Errorf("task %d did not complete within timeout", taskID)
}
