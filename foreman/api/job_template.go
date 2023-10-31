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
	JobTemplateEndpointPrefix string = "job_templates"
)

type ForemanJobTemplate struct {
	ForemanObject

	Description       string      `json:"description"`
	DescriptionFormat string      `json:"description_format"`
	Template          string      `json:"template"`
	Locked            bool        `json:"locked"`
	JobCategory       string      `json:"job_category"`
	ProviderType      string      `json:"provider_type"`
	Snippet           bool        `json:"snippet"`
	TemplateInputs    []string    `json:"template_inputs"`
	EffectiveUser     interface{} `json:"effective_user"`

	Locations     []interface{} `json:"locations"`
	Organizations []interface{} `json:"organizations"`
}

/// CRUD

func (c *Client) CreateJobTemplate(ctx context.Context, jtObj *ForemanJobTemplate) (*ForemanJobTemplate, error) {
	TraceFunctionCall()

	const endpoint = "/" + JobTemplateEndpointPrefix

	wrapped, err := c.WrapJSONWithTaxonomy("job_template", jtObj)
	if err != nil {
		return nil, err
	}

	log.Debugf("jtObj: %+v", jtObj)

	req, err := c.NewRequestWithContext(
		ctx, http.MethodPost, endpoint, bytes.NewBuffer(wrapped),
	)
	if err != nil {
		return nil, err
	}

	var created ForemanJobTemplate
	err = c.SendAndParse(req, &created)
	if err != nil {
		return nil, err
	}

	log.Debugf("%+v", created)

	return &created, nil
}

func (c *Client) QueryJobTemplate(ctx context.Context, jt *ForemanJobTemplate) (QueryResponse, error) {
	TraceFunctionCall()

	qresp := QueryResponse{}
	const endpoint = "/" + JobTemplateEndpointPrefix

	req, err := c.NewRequestWithContext(
		ctx, http.MethodGet, endpoint, nil,
	)
	if err != nil {
		return qresp, err
	}

	reqQuery := req.URL.Query()
	name := `"` + jt.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	err = c.SendAndParse(req, &qresp)
	if err != nil {
		return qresp, err
	}

	results := []ForemanJobTemplate{}
	resultsBytes, err := json.Marshal(qresp.Results)
	if err != nil {
		return qresp, err
	}

	err = json.Unmarshal(resultsBytes, &results)
	if err != nil {
		return qresp, err
	}

	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	qresp.Results = iArr

	return qresp, nil
}

func (c *Client) ReadJobTemplate(ctx context.Context, id int) (*ForemanJobTemplate, error) {
	TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", JobTemplateEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readJT ForemanJobTemplate
	sendErr := c.SendAndParse(req, &readJT)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("ReadJobTemplate jt: [%+v]", readJT)

	return &readJT, nil
}

func (c *Client) UpdateJobTemplate(ctx context.Context, jtObj *ForemanJobTemplate) (*ForemanJobTemplate, error) {
	endpoint := fmt.Sprintf("/%s/%d", JobTemplateEndpointPrefix, jtObj.Id)

	wrappedJT, err := c.WrapJSONWithTaxonomy("job_template", jtObj)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequestWithContext(
		ctx, http.MethodPut, endpoint, bytes.NewBuffer(wrappedJT),
	)
	if err != nil {
		return nil, err
	}

	var updatedJT ForemanJobTemplate
	err = c.SendAndParse(req, &updatedJT)
	if err != nil {
		return nil, err
	}

	return &updatedJT, nil
}

func (c *Client) DeleteJobTemplate(ctx context.Context, id int) error {
	TraceFunctionCall()

	endpoint := fmt.Sprintf("/%s/%d", JobTemplateEndpointPrefix, id)
	req, err := c.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.SendAndParse(req, nil)
}
