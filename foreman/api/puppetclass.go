package api

import (
	"context"
	"encoding/json"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"
	"strings"
)

const (
	PuppetClassEndpointPrefix = "puppet/puppetclasses"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanPuppetClass API model represents a Puppet class
type ForemanPuppetClass struct {
	// Inherits the base object's attributes
	ForemanObject
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// ReadPuppetClass reads the attributes of a ForemanPuppetClass identified by
// the supplied ID and returns a ForemanPuppetClass reference.
func (c *Client) ReadPuppetClass(ctx context.Context, id int) (*ForemanPuppetClass, error) {
	utils.TraceFunctionCall()

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		PuppetClassEndpointPrefix,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readPuppetClass ForemanPuppetClass
	sendErr := c.SendAndParse(req, &readPuppetClass)
	if sendErr != nil {
		return nil, sendErr
	}

	utils.Debugf("readCPuppetClass: [%+v]", readPuppetClass)

	return &readPuppetClass, nil
}

// -----------------------------------------------------------------------------
// Query Implementation
// -----------------------------------------------------------------------------

// QueryPuppetClass queries for a ForemanPuppetClass based on the attributes
// of the supplied ForemanPuppetClass reference and returns a QueryResponse
// struct containing query/response metadata
// The Puppet module search API has a different response format to normal. Results
// are returned in a map instead of an array, with the class name as the key.
// To work around this the results field is unmarshalled and then remarshalled
// into an array to normalise it
func (c *Client) QueryPuppetClass(ctx context.Context, t *ForemanPuppetClass) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponsePuppet{}
	realQueryResponse := QueryResponse{}

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		PuppetClassEndpointPrefix,
		nil,
	)
	if reqErr != nil {
		return QueryResponse{}, reqErr
	}

	// dynamically build the query based on the attributes
	reqQuery := req.URL.Query()
	name := `"` + t.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return QueryResponse{}, sendErr
	}

	utils.Debugf("queryResponse: [%+v]", queryResponse)

	nestedIndex := strings.Index(t.Name, ":")
	var indexName string
	if nestedIndex > 0 {
		indexName = string(t.Name[0:nestedIndex])
	} else {
		indexName = t.Name
	}

	// Results will be Unmarshaled into a []map[string]interface{}
	// Encode back to JSON, then Unmarshal into []ForemanPuppetClass for
	// the results
	results := []ForemanPuppetClass{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results[indexName])
	if jsonEncErr != nil {
		return QueryResponse{}, jsonEncErr
	}

	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return QueryResponse{}, jsonDecErr
	}
	// convert the search results from []ForemanPuppetClass to []interface
	// and set the search results on the query
	iArr := make([]interface{}, 1)
	for idx, val := range results {
		iArr[idx] = val
	}

	realQueryResponse.Subtotal = queryResponse.Subtotal
	realQueryResponse.Results = iArr

	return realQueryResponse, nil
}
