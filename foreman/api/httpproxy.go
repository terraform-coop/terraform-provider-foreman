package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"
)

const (
	HTTPProxyEndpointPrefix = "http_proxies"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanHTTPProxy API model representing a http proxy. Defining HTTP Proxies
// that exist on your network allows you to perform various actions through those proxies.
type ForemanHTTPProxy struct {
	// Inherits the base object's attributes
	ForemanObject

	// Uniform resource locator of the proxy (ie: https://server:8008)
	URL string `json:"url"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateHTTPProxy creates a new ForemanHTTPProxy with the attributes of the
// supplied ForemanHTTPProxy reference and returns the created
// ForemanHTTPProxy reference.  The returned reference will have its ID and
// other API default values set by this function.
func (c *Client) CreateHTTPProxy(ctx context.Context, s *ForemanHTTPProxy) (*ForemanHTTPProxy, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s", HTTPProxyEndpointPrefix)

	sJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("http_proxy", s)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	utils.Debugf("HTTPProxyJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdHTTPProxy ForemanHTTPProxy
	sendErr := c.SendAndParse(req, &createdHTTPProxy)
	if sendErr != nil {
		return nil, sendErr
	}

	utils.Debugf("createdHTTPProxy: [%+v]", createdHTTPProxy)

	return &createdHTTPProxy, nil
}

// ReadHTTPProxy reads the attributes of a ForemanHTTPProxy identified by the
// supplied ID and returns a ForemanHTTPProxy reference.
func (c *Client) ReadHTTPProxy(ctx context.Context, id int) (*ForemanHTTPProxy, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", HTTPProxyEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readHTTPProxy ForemanHTTPProxy
	sendErr := c.SendAndParse(req, &readHTTPProxy)
	if sendErr != nil {
		return nil, sendErr
	}

	utils.Debugf("readHTTPProxy: [%+v]", readHTTPProxy)

	return &readHTTPProxy, nil
}

// UpdateHTTPProxy updates a ForemanHTTPProxy's attributes.  The smart proxy
// with the ID of the supplied ForemanHTTPProxy will be updated. A new
// ForemanHTTPProxy reference is returned with the attributes from the result
// of the update operation.
func (c *Client) UpdateHTTPProxy(ctx context.Context, s *ForemanHTTPProxy) (*ForemanHTTPProxy, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", HTTPProxyEndpointPrefix, s.Id)

	sJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("http_proxy", s)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	utils.Debugf("HTTPProxyJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedHTTPProxy ForemanHTTPProxy
	sendErr := c.SendAndParse(req, &updatedHTTPProxy)
	if sendErr != nil {
		return nil, sendErr
	}

	utils.Debugf("updatedHTTPProxy: [%+v]", updatedHTTPProxy)

	return &updatedHTTPProxy, nil
}

// DeleteHTTPProxy deletes the ForemanHTTPProxy identified by the supplied ID
func (c *Client) DeleteHTTPProxy(ctx context.Context, id int) error {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", HTTPProxyEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return reqErr
	}

	return c.SendAndParse(req, nil)
}

// -----------------------------------------------------------------------------
// Query Implementation
// -----------------------------------------------------------------------------

// QueryHTTPProxy queries for a ForemanHTTPProxy based on the attributes of
// the supplied ForemanHTTPProxy reference and returns a QueryResponse struct
// containing query/response metadata and the matching smart proxy.
func (c *Client) QueryHTTPProxy(ctx context.Context, s *ForemanHTTPProxy) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", HTTPProxyEndpointPrefix)
	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return queryResponse, reqErr
	}

	// dynamically build the query based on the attributes
	reqQuery := req.URL.Query()
	name := `"` + s.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	utils.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanHTTPProxy for
	// the results
	results := []ForemanHTTPProxy{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanHTTPProxy to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}
