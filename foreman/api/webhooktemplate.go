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
	WebhookTemplateEndpointPrefix = "webhook_templates"
)

// The ForemanWebhookTemplate API model represents a webhook template.
// Webhook templates are scripts used to build the payload of a
// webhook request.
type ForemanWebhookTemplate struct {
	ForemanObject
	Template              string `json:"template"`
	Snippet               bool   `json:"snippet"`
	AuditComment          string `json:"audit_comment"`
	Locked                bool   `json:"locked"`
	Default               bool   `json:"default"`
	Description           string `json:"description"`
	LocationIds           []int  `json:"location_ids,omitempty"`
	OrganizationIds       []int  `json:"organization_ids,omitempty"`
	DefaultLocationId     int    `json:"location_id,omitempty"`
	DefaultOrganizationId int    `json:"organization_id,omitempty"`
}

type ForemanWebhookTemplateResponse struct {
	ForemanObject
	Template      string           `json:"template"`
	Snippet       bool             `json:"snippet"`
	AuditComment  string           `json:"audit_comment"`
	Locked        bool             `json:"locked"`
	Default       bool             `json:"default"`
	Description   string           `json:"description"`
	Locations     []EntityResponse `json:"locations,omitempty"`
	Organizations []EntityResponse `json:"organizations,omitempty"`
}

// CreateWebhookTemplate creates a new ForemanWebhookTemplate with
// the attributes of the supplied ForemanWebhookTemplate reference and
// returns the created ForemanWebhookTemplate reference.  The returned
// reference will have its ID and other API default values set by this
// function.
func (c *Client) CreateWebhookTemplate(ctx context.Context, t *ForemanWebhookTemplate) (*ForemanWebhookTemplate, error) {
	log.Tracef("foreman/api/webhooktemplate.go#Create")

	if t.DefaultLocationId == 0 {
		t.DefaultLocationId = c.clientConfig.LocationID
	}

	if t.DefaultOrganizationId == 0 {
		t.DefaultOrganizationId = c.clientConfig.OrganizationID
	}

	reqEndpoint := fmt.Sprintf("/%s", WebhookTemplateEndpointPrefix)

	tJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("webhook_template", t)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("templateJSONBytes: [%s]", tJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(tJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdTemplate ForemanWebhookTemplate
	sendErr := c.SendAndParse(req, &createdTemplate)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdTemplate: [%+v]", createdTemplate)

	return &createdTemplate, nil
}

// ReadWebhookTemplate reads the attributes of a
// ForemanWebhookTemplate identified by the supplied ID and returns a
// ForemanWebhookTemplate reference.
func (c *Client) ReadWebhookTemplate(ctx context.Context, id int) (*ForemanWebhookTemplateResponse, error) {
	log.Tracef("foreman/api/webhooktemplate.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", WebhookTemplateEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readTemplate ForemanWebhookTemplateResponse
	sendErr := c.SendAndParse(req, &readTemplate)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readTemplate: [%+v]", readTemplate)

	return &readTemplate, nil
}

// UpdateWebhookTemplate updates a ForemanWebhookTemplate's
// attributes.  The template with the ID of the supplied
// ForemanWebhookTemplate reference is returned with the attributes from
// ForemanWebhookTemplate will be updated. A new
// the result of the update operation.
func (c *Client) UpdateWebhookTemplate(ctx context.Context, t *ForemanWebhookTemplate) (*ForemanWebhookTemplate, error) {
	log.Tracef("foreman/api/webhooktemplate.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", WebhookTemplateEndpointPrefix, t.Id)
	tJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("webhook_template", t)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("templateJSONBytes: [%s]", tJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(tJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedTemplate ForemanWebhookTemplate
	sendErr := c.SendAndParse(req, &updatedTemplate)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedTemplate: [%+v]", updatedTemplate)

	return &updatedTemplate, nil
}

// DeleteWebhookTemplate deletes the ForemanWebhookTemplate
// identified by the supplied ID
func (c *Client) DeleteWebhookTemplate(ctx context.Context, id int) error {
	log.Tracef("foreman/api/webhooktemplate.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", WebhookTemplateEndpointPrefix, id)

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

// QueryWebhookTemplate queries for a ForemanWebhookTemplate based on
// the attributes of the supplied ForemanWebhookTemplate reference and
// returns a QueryResponse struct containing query/response metadata and the
// matching templates.
func (c *Client) QueryWebhookTemplate(ctx context.Context, t *ForemanWebhookTemplate) (QueryResponse, error) {
	log.Tracef("foreman/api/webhooktemplate.go#Query")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", WebhookTemplateEndpointPrefix)
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
	name := "\"" + t.Name + "\""
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanWebhookTemplate for
	// the results
	results := []ForemanWebhookTemplate{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanWebhookTemplate to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}
