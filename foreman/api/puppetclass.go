package api

import (
	"encoding/json"
	"net/http"

	"github.com/HanseMerkur/terraform-provider-utils/log"
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

// ReadComputeProfile reads the attributes of a ForemanComputeProfile identified by
// the supplied ID and returns a ForemanComputeProfile reference.
func (c *Client) ReadPuppetClass(id int) (*ForemanPuppetClass, error) {
	log.Tracef("foreman/api/puppetclass.go#Read")

	//reqEndpoint := fmt.Sprintf("/%s/%d", PuppetClassEndpointPrefix, id)

	req, reqErr := c.NewRequest(
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

	log.Debugf("readCPuppetClass: [%+v]", readPuppetClass)

	return &readPuppetClass, nil
}

// -----------------------------------------------------------------------------
// Query Implementation
// -----------------------------------------------------------------------------

// QueryComputeProfile queries for a ForemanComputeProfile based on the attributes
// of the supplied ForemanComputeProfile reference and returns a QueryResponse
// struct containing query/response metadata and the matching template kinds
func (c *Client) QueryPuppetClass(t *ForemanPuppetClass) (QueryResponse, error) {
	log.Tracef("foreman/api/puppetclass.go#Search")

	// The Puppet module search API has a different response format to normal. Results
	// are returned in a map instead of an array, with the class name as the key
	queryResponse := QueryResponsePuppet{}
	realQueryResponse := QueryResponse{}

	req, reqErr := c.NewRequest(
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

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanPuppetClass for
	// the results
	results := []ForemanPuppetClass{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results[t.Name])
	if jsonEncErr != nil {
		return QueryResponse{}, jsonEncErr
	}

	log.Debugf("ReMarshall: [%+v]", resultsBytes)

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
