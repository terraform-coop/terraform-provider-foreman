package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HanseMerkur/terraform-provider-utils/log"
)

const (
	DefaultTemplateEndpointPrefix = "/operatingsystems/%d/os_default_templates"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------
// The ForemanDefaultTemplate API model represents the parameter name. DefaultTemplates serve as an
// identification string that defines autonomy, authority, or control for
// a portion of a network.
type ForemanDefaultTemplate struct {
	// Inherits the base object's attributes
	ForemanObject

	OperatingSystemId      int `json:"operatingsystem_id"`
	ProvisioningTemplateId int `json:"provisioning_template_id"`
	TemplateKindId         int `json:"template_kind_id"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateDefaultTemplate creates a new ForemanDefaultTemplate with the attributes of the supplied
// ForemanDefaultTemplate reference and returns the created ForemanDefaultTemplate reference.
// The returned reference will have its ID and other API default values set by
// this function.
func (c *Client) CreateDefaultTemplate(ctx context.Context, d *ForemanDefaultTemplate) (*ForemanDefaultTemplate, error) {
	log.Tracef("foreman/api/parameter.go#Create")

	reqEndpoint := fmt.Sprintf(DefaultTemplateEndpointPrefix, d.OperatingSystemId)

	// All parameters are send individually. Yeay for that
	var createdDefaultTemplate ForemanDefaultTemplate
	wrapped, _ := c.wrapParameters("os_default_template", d)
	parameterJSONBytes, jsonEncErr := json.Marshal(wrapped)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("parameterJSONBytes: [%s]", parameterJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(parameterJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	sendErr := c.SendAndParse(req, &createdDefaultTemplate)
	if sendErr != nil {
		return nil, sendErr
	}
	log.Debugf("createdDefaultTemplate: [%+v]", createdDefaultTemplate)

	return &createdDefaultTemplate, nil
}

// ReadDefaultTemplate reads the attributes of a ForemanDefaultTemplate identified by the
// supplied ID and returns a ForemanDefaultTemplate reference.
func (c *Client) ReadDefaultTemplate(ctx context.Context, d *ForemanDefaultTemplate, id int) (*ForemanDefaultTemplate, error) {
	log.Tracef("foreman/api/parameter.go#Read")

	reqEndpoint := fmt.Sprintf(DefaultTemplateEndpointPrefix+"/%d", d.OperatingSystemId, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readDefaultTemplate ForemanDefaultTemplate
	sendErr := c.SendAndParse(req, &readDefaultTemplate)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readDefaultTemplate: [%+v]", readDefaultTemplate)

	return &readDefaultTemplate, nil
}

// UpdateDefaultTemplate deletes all parameters for the subject resource and re-creates them
// as we look at them differently on either side this is the safest way to reach sync
func (c *Client) UpdateDefaultTemplate(ctx context.Context, d *ForemanDefaultTemplate, id int) (*ForemanDefaultTemplate, error) {
	log.Tracef("foreman/api/parameter.go#Update")

	reqEndpoint := fmt.Sprintf(DefaultTemplateEndpointPrefix+"/%d", d.OperatingSystemId, id)
	wrapped, _ := c.wrapParameters("os_default_template", d)
	parameterJSONBytes, jsonEncErr := json.Marshal(wrapped)

	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("parameterJSONBytes: [%s]", parameterJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(parameterJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedDefaultTemplate ForemanDefaultTemplate
	sendErr := c.SendAndParse(req, &updatedDefaultTemplate)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedDefaultTemplate: [%+v]", updatedDefaultTemplate)

	return &updatedDefaultTemplate, nil
}

// DeleteDefaultTemplate deletes the ForemanDefaultTemplates for the given resource
func (c *Client) DeleteDefaultTemplate(ctx context.Context, d *ForemanDefaultTemplate, id int) error {
	log.Tracef("foreman/api/parameter.go#Delete")

	reqEndpoint := fmt.Sprintf(DefaultTemplateEndpointPrefix+"/%d", d.OperatingSystemId, id)

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

// QueryDefaultTemplate queries for a ForemanDefaultTemplate based on the attributes of the
// supplied ForemanDefaultTemplate reference and returns a QueryResponse struct
// containing query/response metadata and the matching parameters.
func (c *Client) QueryDefaultTemplate(ctx context.Context, d *ForemanDefaultTemplate) (QueryResponse, error) {
	log.Tracef("foreman/api/parameter.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", DefaultTemplateEndpointPrefix)
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

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanDefaultTemplate for
	// the results
	results := []ForemanDefaultTemplate{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanDefaultTemplate to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}
