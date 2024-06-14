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

	// Map of DomainParameters
	DomainParameters []ForemanKVParameter `json:"domain_parameters_attributes,omitempty"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateDomain creates a new ForemanDomain with the attributes of the supplied
// ForemanDomain reference and returns the created ForemanDomain reference.
// The returned reference will have its ID and other API default values set by
// this function.
func (c *Client) CreateDomain(ctx context.Context, d *ForemanDomain) (*ForemanDomain, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s", DomainEndpointPrefix)

	domainJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("domain", d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	utils.Debugf("domainJSONBytes: [%s]", domainJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
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

	utils.Debugf("createdDomain: [%+v]", createdDomain)

	return &createdDomain, nil
}

// ReadDomain reads the attributes of a ForemanDomain identified by the
// supplied ID and returns a ForemanDomain reference.
func (c *Client) ReadDomain(ctx context.Context, id int) (*ForemanDomain, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", DomainEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
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

	utils.Debugf("readDomain: [%+v]", readDomain)

	return &readDomain, nil
}

// UpdateDomain updates a ForemanDomain's attributes.  The domain with the ID
// of the supplied ForemanDomain will be updated. A new ForemanDomain reference
// is returned with the attributes from the result of the update operation.
func (c *Client) UpdateDomain(ctx context.Context, d *ForemanDomain, id int) (*ForemanDomain, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", DomainEndpointPrefix, id)

	domainJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("domain", d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	utils.Debugf("domainJSONBytes: [%s]", domainJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
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

	utils.Debugf("updatedDomain: [%+v]", updatedDomain)

	return &updatedDomain, nil
}

// DeleteDomain deletes the ForemanDomain identified by the supplied ID
func (c *Client) DeleteDomain(ctx context.Context, id int) error {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", DomainEndpointPrefix, id)

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

// QueryDomain queries for a ForemanDomain based on the attributes of the
// supplied ForemanDomain reference and returns a QueryResponse struct
// containing query/response metadata and the matching domains.
func (c *Client) QueryDomain(ctx context.Context, d *ForemanDomain) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", DomainEndpointPrefix)
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
	name := `"` + d.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	utils.Debugf("queryResponse: [%+v]", queryResponse)

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
