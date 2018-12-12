package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	DomainEndpointPrefix = "domains"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanDomain API model represents the domain name. Domains serve as an
// identification string that defines autonomy, authority, or control for
// a portion of a network.
type ForemanDomain struct {
	// Inherits the base object's attributes
	ForemanObject

	// Fully qualified domain name
	Fullname string `json:"fullname"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateDomain creates a new ForemanDomain with the attributes of the supplied
// ForemanDomain reference and returns the created ForemanDomain reference.
// The returned reference will have its ID and other API default values set by
// this function.
func (c *Client) CreateDomain(d *ForemanDomain) (*ForemanDomain, error) {
	log.Tracef("foreman/api/domain.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", DomainEndpointPrefix)

	domainJSONBytes, jsonEncErr := json.Marshal(d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("domainJSONBytes: [%s]", domainJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(domainJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdDomain ForemanDomain
	sendErr := c.SendAndParse(req, &createdDomain)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdDomain: [%+v]", createdDomain)

	return &createdDomain, nil
}

// ReadDomain reads the attributes of a ForemanDomain identified by the
// supplied ID and returns a ForemanDomain reference.
func (c *Client) ReadDomain(id int) (*ForemanDomain, error) {
	log.Tracef("foreman/api/domain.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", DomainEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readDomain ForemanDomain
	sendErr := c.SendAndParse(req, &readDomain)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readDomain: [%+v]", readDomain)

	return &readDomain, nil
}

// UpdateDomain updates a ForemanDomain's attributes.  The domain with the ID
// of the supplied ForemanDomain will be updated. A new ForemanDomain reference
// is returned with the attributes from the result of the update operation.
func (c *Client) UpdateDomain(d *ForemanDomain) (*ForemanDomain, error) {
	log.Tracef("foreman/api/domain.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", DomainEndpointPrefix, d.Id)

	domainJSONBytes, jsonEncErr := json.Marshal(d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("domainJSONBytes: [%s]", domainJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(domainJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedDomain ForemanDomain
	sendErr := c.SendAndParse(req, &updatedDomain)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedDomain: [%+v]", updatedDomain)

	return &updatedDomain, nil
}

// DeleteDomain deletes the ForemanDomain identified by the supplied ID
func (c *Client) DeleteDomain(id int) error {
	log.Tracef("foreman/api/domain.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", DomainEndpointPrefix, id)

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

// QueryDomain queries for a ForemanDomain based on the attributes of the
// supplied ForemanDomain reference and returns a QueryResponse struct
// containing query/response metadata and the matching domains.
func (c *Client) QueryDomain(d *ForemanDomain) (QueryResponse, error) {
	log.Tracef("foreman/api/domain.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", DomainEndpointPrefix)
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
	name := `"` + d.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanDomain for
	// the results
	results := []ForemanDomain{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanDomain to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}
