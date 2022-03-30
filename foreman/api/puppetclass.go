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

// ReadPuppetClass reads the attributes of a ForemanPuppetClass identified by
// the supplied ID and returns a ForemanPuppetClass reference.
func (c *Client) ReadPuppetClass(id int) (*ForemanPuppetClass, error) {
	log.Tracef("foreman/api/puppetclass.go#Read")

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

// QueryPuppetClass queries for a ForemanPuppetClass based on the attributes
// of the supplied ForemanPuppetClass reference and returns a QueryResponse
// struct containing query/response metadata
// The Puppet module search API has a different response format to normal. Results
// are returned in a map instead of an array, with the class name as the key.
// To work around this the results field is unmarshalled and then remarshalled
// into an array to normalise it
func (c *Client) QueryPuppetClass(t *ForemanPuppetClass) (QueryResponse, error) {
	log.Tracef("foreman/api/puppetclass.go#Search")

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

	// Get the first (and only) key from the results map
	// We could just use the search name that's passed in, but then
	// the unit test will fail as it passes in empty string for name
	resultKey := make([]string, len(queryResponse.Results))
	i := 0
	for k := range queryResponse.Results {
		resultKey[i] = k
		i++
	}

	// Results will be Unmarshaled into a []map[string]interface{}
	// Encode back to JSON, then Unmarshal into []ForemanPuppetClass for
	// the results
	results := []ForemanPuppetClass{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results[resultKey[0]])
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
