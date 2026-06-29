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
	cleanhttp "github.com/hashicorp/go-cleanhttp"
)

const (
	ForemanAPIVersion       = "2"
	ForemanAPIURLPrefix     = "/api"
	ForemanKatelloURLPrefix = "/katello/api"
	ForemanTasksURLPrefix   = "/foreman_tasks/api"
	ForemanPuppetURLPrefix  = "/foreman_puppet/api"
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
	cleanClient := cleanhttp.DefaultClient()
	if cfg.NegotiateAuth {
		transCfg := &spnego.Transport{}
		transCfg.TLSClientConfig = tlsCfg
		cleanClient.Transport = transCfg
	} else {
		transCfg := &http.Transport{}
		transCfg.TLSClientConfig = tlsCfg
		cleanClient.Transport = transCfg
	}
	return &ForemanClient{
		serverURL:   serverURL,
		httpClient:  cleanClient,
		credentials: creds,
		config:      cfg,
	}
}

func (c *ForemanClient) newRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	reqURL := c.serverURL
	ep := endpoint

	switch {
	case strings.HasPrefix(ep, "katello"):
		reqURL.Path = ForemanKatelloURLPrefix + strings.TrimPrefix(ep, "katello")
	case strings.HasPrefix(ep, "/katello/api"):
		reqURL.Path = ep
	case strings.HasPrefix(ep, "puppet"):
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

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), reqURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("User-Agent", "terraform-provider-foreman")
	req.Header.Add("Accept", "application/json,version="+ForemanAPIVersion)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(c.credentials.Username, c.credentials.Password)
	return req, nil
}

func (c *ForemanClient) send(req *http.Request) (int, []byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return -1, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}
	return resp.StatusCode, body, nil
}

func (c *ForemanClient) do(ctx context.Context, method, endpoint string, reqBody, respObj interface{}) error {
	var bodyReader io.Reader
	if reqBody != nil {
		payload, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
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
			return fmt.Errorf("failed to parse async task: %w", err)
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

func (c *ForemanClient) addTaxonomy(body interface{}) interface{} {
	m, ok := body.(map[string]interface{})
	if !ok {
		return body
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
	wrapped := c.wrapRequestBody(wrapperKey, reqBody)
	taxonomy := c.addTaxonomy(wrapped)
	return c.do(ctx, http.MethodPost, endpoint, taxonomy, respObj)
}

func (c *ForemanClient) Put(ctx context.Context, endpoint, wrapperKey string, reqBody, respObj interface{}) error {
	wrapped := c.wrapRequestBody(wrapperKey, reqBody)
	taxonomy := c.addTaxonomy(wrapped)
	return c.do(ctx, http.MethodPut, endpoint, taxonomy, respObj)
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

func (e *HTTPError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsNotFoundError returns true if the error is a 404 (resource not found).
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.IsNotFound()
	}
	return false
}

func (c *ForemanClient) waitForKatelloTask(ctx context.Context, taskID int) (*ForemanTask, error) {
	endpoint := fmt.Sprintf("/foreman_tasks/api/tasks/%d", taskID)
	for i := 0; i < 10; i++ {
		var task ForemanTask
		if err := c.Get(ctx, endpoint, &task); err != nil {
			return nil, fmt.Errorf("failed to poll task %d: %w", taskID, err)
		}
		if !task.Pending {
			return &task, nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return nil, fmt.Errorf("task %d did not complete within timeout", taskID)
}
