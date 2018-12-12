package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	EnvironmentEndpointPrefix = "environments"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanEnvironment API model represents a puppet environment
type ForemanEnvironment struct {
	// Inherits the base object's attributes
	ForemanObject
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateEnvironment creates a new ForemanEnvironment with the attributes of
// the supplied ForemanEnvironment reference and returns the created
// ForemanEnvironment reference.  The returned reference will have its ID and
// other API default values set by this function.
func (c *Client) CreateEnvironment(e *ForemanEnvironment) (*ForemanEnvironment, error) {
	log.Tracef("foreman/api/environment.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", EnvironmentEndpointPrefix)

	environmentJSONBytes, jsonEncErr := json.Marshal(e)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("environmentJSONBytes: [%s]", environmentJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(environmentJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdEnvironment ForemanEnvironment
	sendErr := c.SendAndParse(req, &createdEnvironment)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdEnvironment: [%+v]", createdEnvironment)

	return &createdEnvironment, nil
}

// ReadEnvironment reads the attributes of a ForemanEnvironment identified by
// the supplied ID and returns a ForemanEnvironment reference.
func (c *Client) ReadEnvironment(id int) (*ForemanEnvironment, error) {
	log.Tracef("foreman/api/environment.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", EnvironmentEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readEnvironment ForemanEnvironment
	sendErr := c.SendAndParse(req, &readEnvironment)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readEnvironment: [%+v]", readEnvironment)

	return &readEnvironment, nil
}

// UpdateEnvironment updates a ForemanEnvironment's attributes.  The
// environment with the ID of the supplied ForemanEnvironment will be updated.
// A new ForemanEnvironment reference is returned with the attributes from the
// result of the update operation.
func (c *Client) UpdateEnvironment(e *ForemanEnvironment) (*ForemanEnvironment, error) {
	log.Tracef("foreman/api/environment.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", EnvironmentEndpointPrefix, e.Id)

	environmentJSONBytes, jsonEncErr := json.Marshal(e)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("environmentJSONBytes: [%s]", environmentJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(environmentJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedEnvironment ForemanEnvironment
	sendErr := c.SendAndParse(req, &updatedEnvironment)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedEnvironment: [%+v]", updatedEnvironment)

	return &updatedEnvironment, nil
}

// DeleteEnvironment deletes the ForemanEnvironment identified by the supplied
// ID
func (c *Client) DeleteEnvironment(id int) error {
	log.Tracef("foreman/api/environment.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", EnvironmentEndpointPrefix, id)

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

// QueryEnvironment queries for a ForemanEnvironment based on the attributes of
// the supplied ForemanEnvironment reference and returns a QueryResponse struct
// containing query/response metadata and the matching environments.
func (c *Client) QueryEnvironment(e *ForemanEnvironment) (QueryResponse, error) {
	log.Tracef("foreman/api/environment.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", EnvironmentEndpointPrefix)
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
	name := `"` + e.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanEnvironment for
	// the results
	results := []ForemanEnvironment{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanEnvironment to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}
