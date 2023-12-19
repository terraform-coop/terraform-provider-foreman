package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"

	"github.com/HanseMerkur/terraform-provider-utils/log"
)

const (
	SmartClassParameterEndpointPrefix      = "puppet/smarts_class_paramaters"
	SmartClassParameterQueryEndpointPrefix = "puppet/puppetclasses/%d/smart_class_parameters"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanSmartClassParameter API model represents a smart class parameter
type ForemanSmartClassParameter struct {
	// Inherits the base object's attributes
	ForemanObject

	// Smart class parameter name
	Parameter string `json:"parameter"`
	// ID of the owning puppet class
	PuppetClassId int `json:"puppetclass_id"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// ReadSmartClassParamter reads the attributes of a ForemanSmartClassParameter identified by
// the supplied ID and returns a ForemanSmartClassParameter reference.
func (c *Client) ReadSmartClassParameter(ctx context.Context, id int) (*ForemanSmartClassParameter, error) {
	utils.TraceFunctionCall()

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		SmartClassParameterEndpointPrefix,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readSmartClassParameter ForemanSmartClassParameter
	sendErr := c.SendAndParse(req, &readSmartClassParameter)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readSmartClassParameter: [%+v]", readSmartClassParameter)

	return &readSmartClassParameter, nil
}

// -----------------------------------------------------------------------------
// Query Implementation
// -----------------------------------------------------------------------------

// QuerySmartClassParameter queries for a ForemanSmartClassParameter based on the attributes
// of the supplied ForemanSmartClassParameter reference and returns a QueryResponse
// struct containing query/response metadata
func (c *Client) QuerySmartClassParameter(ctx context.Context, t *ForemanSmartClassParameter) (QueryResponse, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf(SmartClassParameterQueryEndpointPrefix, t.PuppetClassId)

	queryResponse := QueryResponse{}

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return QueryResponse{}, reqErr
	}

	// dynamically build the query based on the attributes
	reqQuery := req.URL.Query()
	param := `"` + t.Parameter + `"`
	reqQuery.Set("search", "parameter="+param)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return QueryResponse{}, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	// Encode back to JSON, then Unmarshal into []ForemanPuppetClass for
	// the results
	results := []ForemanSmartClassParameter{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return QueryResponse{}, jsonEncErr
	}

	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return QueryResponse{}, jsonDecErr
	}
	// convert the search results from []ForemanSmartParameterClass to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}

	queryResponse.Results = iArr

	return queryResponse, nil
}
