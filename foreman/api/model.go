package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"

	"github.com/HanseMerkur/terraform-provider-utils/log"
)

const (
	ModelEndpointPrefix = "models"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanModel API model represents a specific vendor hardware model
type ForemanModel struct {
	// Inherits the base object's attributes
	ForemanObject

	// Additional information about this hardware model
	Info string `json:"info"`
	// Name or class of the hardware vendor
	VendorClass string `json:"vendor_class"`
	// Name of the specific hardware model
	HardwareModel string `json:"hardware_model"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateModel creates a new ForemanModel with the attributes of the supplied
// ForemanModel reference and returns the created ForemanModel reference.  The
// returned reference will have its ID and other API default values set by this
// function.
func (c *Client) CreateModel(ctx context.Context, m *ForemanModel) (*ForemanModel, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s", ModelEndpointPrefix)

	mJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("model", m)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("modelJSONBytes: [%s]", mJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(mJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdModel ForemanModel
	sendErr := c.SendAndParse(req, &createdModel)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdModel: [%+v]", createdModel)

	return &createdModel, nil
}

// ReadModel reads the attributes of a ForemanModel identified by the supplied
// ID and returns a ForemanModel reference.
func (c *Client) ReadModel(ctx context.Context, id int) (*ForemanModel, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", ModelEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readModel ForemanModel
	sendErr := c.SendAndParse(req, &readModel)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readModel: [%+v]", readModel)

	return &readModel, nil
}

// UpdateModel updates a ForemanModel's attributes.  The model with the ID of
// the supplied ForemanModel will be updated. A new ForemanModel reference is
// returned with the attributes from the result of the update operation.
func (c *Client) UpdateModel(ctx context.Context, m *ForemanModel) (*ForemanModel, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", ModelEndpointPrefix, m.Id)

	mJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("model", m)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("modelJSONBytes: [%s]", mJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(mJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedModel ForemanModel
	sendErr := c.SendAndParse(req, &updatedModel)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedModel: [%+v]", updatedModel)

	return &updatedModel, nil
}

// DeleteModel deletes the ForemanModel identified by the supplied ID
func (c *Client) DeleteModel(ctx context.Context, id int) error {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", ModelEndpointPrefix, id)

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

// QueryModel queries for a ForemanModel based on the attributes of the
// supplied ForemanModel reference and returns a QueryResponse struct
// containing query/response metadata and the matching model.
func (c *Client) QueryModel(ctx context.Context, m *ForemanModel) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", ModelEndpointPrefix)
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
	name := `"` + m.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanModel for
	// the results
	results := []ForemanModel{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanModel to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}
