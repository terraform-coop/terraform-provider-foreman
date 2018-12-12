package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	SmartProxyEndpointPrefix = "smart_proxies"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanSmartProxy API model representing a proxy server.  Smart proxies
// provide an API for a higher-level orchestration tool.  Foreman supports
// the following smart proxies:
//   1. DHCP - ISC DHCP & MS DHCP servers
//   2. DNS - bind & MS DNS servers
//   3. Puppet >= 0.24.x
//   4. Puppet CA
//   5. Realm - manage host registration to a realm (eg: FreeIPA)
//   6. Templates - Proxy template requests from hosts in isolated networks
//   7. TFTP
type ForemanSmartProxy struct {
	// Inherits the base object's attributes
	ForemanObject

	// Uniform resource locator of the proxy (ie: https://server:8008)
	URL string `json:"url"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateSmartProxy creates a new ForemanSmartProxy with the attributes of the
// supplied ForemanSmartProxy reference and returns the created
// ForemanSmartProxy reference.  The returned reference will have its ID and
// other API default values set by this function.
func (c *Client) CreateSmartProxy(s *ForemanSmartProxy) (*ForemanSmartProxy, error) {
	log.Tracef("foreman/api/smartproxy.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", SmartProxyEndpointPrefix)

	sJSONBytes, jsonEncErr := json.Marshal(s)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("smartproxyJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdSmartProxy ForemanSmartProxy
	sendErr := c.SendAndParse(req, &createdSmartProxy)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdSmartProxy: [%+v]", createdSmartProxy)

	return &createdSmartProxy, nil
}

// ReadSmartProxy reads the attributes of a ForemanSmartProxy identified by the
// supplied ID and returns a ForemanSmartProxy reference.
func (c *Client) ReadSmartProxy(id int) (*ForemanSmartProxy, error) {
	log.Tracef("foreman/api/smartproxy.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", SmartProxyEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readSmartProxy ForemanSmartProxy
	sendErr := c.SendAndParse(req, &readSmartProxy)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readSmartProxy: [%+v]", readSmartProxy)

	return &readSmartProxy, nil
}

// UpdateSmartProxy updates a ForemanSmartProxy's attributes.  The smart proxy
// with the ID of the supplied ForemanSmartProxy will be updated. A new
// ForemanSmartProxy reference is returned with the attributes from the result
// of the update operation.
func (c *Client) UpdateSmartProxy(s *ForemanSmartProxy) (*ForemanSmartProxy, error) {
	log.Tracef("foreman/api/smartproxy.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", SmartProxyEndpointPrefix, s.Id)

	sJSONBytes, jsonEncErr := json.Marshal(s)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("smartproxyJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedSmartProxy ForemanSmartProxy
	sendErr := c.SendAndParse(req, &updatedSmartProxy)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedSmartProxy: [%+v]", updatedSmartProxy)

	return &updatedSmartProxy, nil
}

// DeleteSmartProxy deletes the ForemanSmartProxy identified by the supplied ID
func (c *Client) DeleteSmartProxy(id int) error {
	log.Tracef("foreman/api/smartproxy.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", SmartProxyEndpointPrefix, id)

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

// QuerySmartProxy queries for a ForemanSmartProxy based on the attributes of
// the supplied ForemanSmartProxy reference and returns a QueryResponse struct
// containing query/response metadata and the matching smart proxy.
func (c *Client) QuerySmartProxy(s *ForemanSmartProxy) (QueryResponse, error) {
	log.Tracef("foreman/api/smartproxy.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", SmartProxyEndpointPrefix)
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
	// Encode back to JSON, then Unmarshal into []ForemanSmartProxy for
	// the results
	results := []ForemanSmartProxy{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanSmartProxy to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}
