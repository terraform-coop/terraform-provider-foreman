package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"
)

const (
	TemplateKindEndpointPrefix = "template_kinds"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanTemplateKind API model represents a category of provisioning
// template. Examples include:
//  1. PXELinux
//  2. provision
//  3. PXEGrub
//  4. ZTP
type ForemanTemplateKind struct {
	// Inherits the base object's attributes
	ForemanObject
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// ReadTemplateKind reads the attributes of a ForemanTemplateKind identified by
// the supplied ID and returns a ForemanTemplateKind reference.
func (c *Client) ReadTemplateKind(ctx context.Context, id int) (*ForemanTemplateKind, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", TemplateKindEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readTemplateKind ForemanTemplateKind
	sendErr := c.SendAndParse(req, &readTemplateKind)
	if sendErr != nil {
		return nil, sendErr
	}

	utils.Debugf("readTemplateKind: [%+v]", readTemplateKind)

	return &readTemplateKind, nil
}

// -----------------------------------------------------------------------------
// Query Implementation
// -----------------------------------------------------------------------------

// QueryTemplateKind queries for a ForemanTemplateKind based on the attributes
// of the supplied ForemanTemplateKind reference and returns a QueryResponse
// struct containing query/response metadata and the matching template kinds
func (c *Client) QueryTemplateKind(ctx context.Context, t *ForemanTemplateKind) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", TemplateKindEndpointPrefix)
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
	name := `"` + t.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	utils.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanTemplateKind for
	// the results
	results := []ForemanTemplateKind{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanTemplateKind to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}
