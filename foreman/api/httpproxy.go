package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HanseMerkur/terraform-provider-utils/log"
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
func (c *Client) CreateHTTPProxy(s *ForemanHTTPProxy) (*ForemanHTTPProxy, error) {
	log.Tracef("foreman/api/httpproxy.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", HTTPProxyEndpointPrefix)

	sJSONBytes, jsonEncErr := c.WrapJSON("http_proxy", s)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("HTTPProxyJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequest(
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

	log.Debugf("createdHTTPProxy: [%+v]", createdHTTPProxy)

	return &createdHTTPProxy, nil
}

// ReadHTTPProxy reads the attributes of a ForemanHTTPProxy identified by the
// supplied ID and returns a ForemanHTTPProxy reference.
func (c *Client) ReadHTTPProxy(id int) (*ForemanHTTPProxy, error) {
	log.Tracef("foreman/api/HTTPProxy.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", HTTPProxyEndpointPrefix, id)

	req, reqErr := c.NewRequest(
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

	log.Debugf("readHTTPProxy: [%+v]", readHTTPProxy)

	return &readHTTPProxy, nil
}

// UpdateHTTPProxy updates a ForemanHTTPProxy's attributes.  The smart proxy
// with the ID of the supplied ForemanHTTPProxy will be updated. A new
// ForemanHTTPProxy reference is returned with the attributes from the result
// of the update operation.
func (c *Client) UpdateHTTPProxy(s *ForemanHTTPProxy) (*ForemanHTTPProxy, error) {
	log.Tracef("foreman/api/HTTPProxy.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", HTTPProxyEndpointPrefix, s.Id)

	sJSONBytes, jsonEncErr := c.WrapJSON("http_proxy", s)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("HTTPProxyJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequest(
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

	log.Debugf("updatedHTTPProxy: [%+v]", updatedHTTPProxy)

	return &updatedHTTPProxy, nil
}

// DeleteHTTPProxy deletes the ForemanHTTPProxy identified by the supplied ID
func (c *Client) DeleteHTTPProxy(id int) error {
	log.Tracef("foreman/api/HTTPProxy.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", HTTPProxyEndpointPrefix, id)

	req, reqErr := c.NewRequest(
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
func (c *Client) QueryHTTPProxy(s *ForemanHTTPProxy) (QueryResponse, error) {
	log.Tracef("foreman/api/HTTPProxy.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", HTTPProxyEndpointPrefix)
	req, reqErr := c.NewRequest(
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

	log.Debugf("queryResponse: [%+v]", queryResponse)

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
