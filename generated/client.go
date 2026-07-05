package generated

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
	ForemanAPIVersion       = "2"
	ForemanAPIURLPrefix     = "/api"
	ForemanKatelloURLPrefix = "/katello/api"
	ForemanTasksURLPrefix   = "/foreman_tasks/api"
	ForemanPuppetURLPrefix  = "/foreman_puppet/api"

	defaultRequestTimeout = 60 * time.Second
)

type ForemanClient struct {
	serverURL   url.URL
	httpClient  *http.Client
	credentials ClientCredentials
	config      ClientConfig
}

type ClientCredentials struct {
	Username string
	Password string
}

type ClientConfig struct {
	TLSInsecure    bool
	NegotiateAuth  bool
	OrganizationID int
	LocationID     int
}

func NewClient(serverURL url.URL, creds ClientCredentials, cfg ClientConfig) *ForemanClient {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.TLSInsecure,
	}
	client := &http.Client{Timeout: defaultRequestTimeout}
	if cfg.NegotiateAuth {
		client.Transport = &spnego.Transport{Transport: http.Transport{TLSClientConfig: tlsCfg}}
	} else {
		client.Transport = &http.Transport{TLSClientConfig: tlsCfg}
	}
	return &ForemanClient{
		serverURL:   serverURL,
		httpClient:  client,
		credentials: creds,
		config:      cfg,
	}
}

func (c *ForemanClient) newRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
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
		reqURL.Path = ForemanKatelloURLPrefix + strings.TrimPrefix(ep, "katello")
	case strings.HasPrefix(ep, "/katello/api"):
		reqURL.Path = ep
	case strings.HasPrefix(ep, "puppet/"):
		reqURL.Path = ForemanPuppetURLPrefix + strings.TrimPrefix(ep, "puppet")
	case strings.HasPrefix(ep, "foreman_tasks"):
		reqURL.Path = ep
	default:
		if strings.HasPrefix(ep, "/") {
			reqURL.Path = ForemanAPIURLPrefix + ep
		} else {
			reqURL.Path = ForemanAPIURLPrefix + "/" + ep
		}
	}
	reqURL.RawQuery = rawQuery

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), reqURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("User-Agent", "terraform-provider-foreman")
	req.Header.Add("Accept", "application/json,version="+ForemanAPIVersion)
	req.Header.Add("Content-Type", "application/json")
	if !c.config.NegotiateAuth {
		req.SetBasicAuth(c.credentials.Username, c.credentials.Password)
	}
	return req, nil
}

func (c *ForemanClient) send(req *http.Request) (int, []byte, error) {
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

func (c *ForemanClient) do(ctx context.Context, method, endpoint string, reqBody, respObj interface{}) error {
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
		var task ForemanTask
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
func (c *ForemanClient) addTaxonomy(reqBody interface{}) interface{} {
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

func (c *ForemanClient) Get(ctx context.Context, endpoint string, respObj interface{}) error {
	return c.do(ctx, http.MethodGet, endpoint, nil, respObj)
}

func (c *ForemanClient) Post(ctx context.Context, endpoint, wrapperKey string, reqBody, respObj interface{}) error {
	wrapped := c.wrapRequestBody(wrapperKey, c.addTaxonomy(reqBody))
	return c.do(ctx, http.MethodPost, endpoint, wrapped, respObj)
}

func (c *ForemanClient) Put(ctx context.Context, endpoint, wrapperKey string, reqBody, respObj interface{}) error {
	wrapped := c.wrapRequestBody(wrapperKey, c.addTaxonomy(reqBody))
	return c.do(ctx, http.MethodPut, endpoint, wrapped, respObj)
}

func (c *ForemanClient) Delete(ctx context.Context, endpoint string) error {
	err := c.do(ctx, http.MethodDelete, endpoint, nil, nil)
	if err != nil && IsNotFoundError(err) {
		return nil // Already deleted is not an error
	}
	return err
}

func (c *ForemanClient) wrapRequestBody(wrapperKey string, reqBody interface{}) interface{} {
	if wrapperKey == "" {
		return reqBody
	}
	return map[string]interface{}{wrapperKey: reqBody}
}

type HTTPError struct {
	Endpoint   string
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error: endpoint=%s status=%d body=%s", e.Endpoint, e.StatusCode, e.Body)
}

// IsNotFoundError returns true if the error is a 404 (resource not found).
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == http.StatusNotFound
	}
	return false
}

func (c *ForemanClient) waitForKatelloTask(ctx context.Context, taskID int) (*ForemanTask, error) {
	endpoint := fmt.Sprintf("/foreman_tasks/api/tasks/%d", taskID)
	for i := 0; i < 10; i++ {
		var task ForemanTask
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
